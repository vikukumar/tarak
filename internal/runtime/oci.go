package runtime

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// OCIImageConfig represents the execution configuration of an image
type OCIImageConfig struct {
	Env        []string `json:"Env"`
	Entrypoint []string `json:"Entrypoint"`
	Cmd        []string `json:"Cmd"`
	WorkingDir string   `json:"WorkingDir"`
}

type ociManifest struct {
	Config struct {
		Digest string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		Digest string `json:"digest"`
	} `json:"layers"`
}

type ociImageJSON struct {
	Config OCIImageConfig `json:"config"`
}

// pullAndExtractOCI natively pulls an image from Docker Hub and extracts its filesystem.
func (e *Engine) pullAndExtractOCI(ctx context.Context, image string, targetRootfs string) (string, OCIImageConfig, error) {
	// 1. Parse Image Name
	repo := image
	tag := "latest"
	if strings.Contains(image, ":") {
		parts := strings.SplitN(image, ":", 2)
		repo = parts[0]
		tag = parts[1]
	}
	if !strings.Contains(repo, "/") {
		repo = "library/" + repo
	}

	e.log.Info("Starting native OCI pull", zap.String("repo", repo), zap.String("tag", tag))

	// 2. Fetch Auth Token
	tokenURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", OCIImageConfig{}, fmt.Errorf("fetch auth token: %w", err)
	}
	defer resp.Body.Close()
	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", OCIImageConfig{}, fmt.Errorf("decode token: %w", err)
	}

	// 3. Fetch Manifest
	manifestURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", repo, tag)
	req, _ = http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", OCIImageConfig{}, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", OCIImageConfig{}, fmt.Errorf("manifest request failed: %s", resp.Status)
	}

	var rawManifest map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawManifest); err != nil {
		return "", OCIImageConfig{}, fmt.Errorf("decode manifest: %w", err)
	}

	// Handle OCI Image Index / Manifest List
	if mediaType, _ := rawManifest["mediaType"].(string); strings.Contains(mediaType, "manifest.list") || strings.Contains(mediaType, "index") {
		manifests, _ := rawManifest["manifests"].([]interface{})
		var targetDigest string
		for _, m := range manifests {
			mMap, _ := m.(map[string]interface{})
			platform, _ := mMap["platform"].(map[string]interface{})
			if platform != nil && platform["architecture"] == "amd64" && platform["os"] == "linux" {
				targetDigest, _ = mMap["digest"].(string)
				break
			}
		}
		if targetDigest == "" && len(manifests) > 0 {
			mMap, _ := manifests[0].(map[string]interface{})
			targetDigest, _ = mMap["digest"].(string)
		}
		
		if targetDigest == "" {
			return "", OCIImageConfig{}, fmt.Errorf("could not find compatible amd64/linux manifest in list")
		}

		manifestURL = fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", repo, targetDigest)
		req, _ = http.NewRequestWithContext(ctx, "GET", manifestURL, nil)
		req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
		req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")
		resp2, err := http.DefaultClient.Do(req)
		if err != nil || resp2.StatusCode != 200 {
			return "", OCIImageConfig{}, fmt.Errorf("fetch target manifest failed")
		}
		defer resp2.Body.Close()
		
		rawManifest = make(map[string]interface{})
		if err := json.NewDecoder(resp2.Body).Decode(&rawManifest); err != nil {
			return "", OCIImageConfig{}, fmt.Errorf("decode target manifest: %w", err)
		}
	}

	var manifest ociManifest
	rawBytes, _ := json.Marshal(rawManifest)
	if err := json.Unmarshal(rawBytes, &manifest); err != nil {
		return "", OCIImageConfig{}, fmt.Errorf("parse manifest: %w", err)
	}

	// 4. Fetch Config Blob
	configURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/blobs/%s", repo, manifest.Config.Digest)
	req, _ = http.NewRequestWithContext(ctx, "GET", configURL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", OCIImageConfig{}, fmt.Errorf("fetch config blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", OCIImageConfig{}, fmt.Errorf("config blob request failed: %s - %s", resp.Status, string(body))
	}

	var config ociImageJSON
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return "", OCIImageConfig{}, fmt.Errorf("decode config: %w", err)
	}

	// 5. Download and Extract Layers sequentially
	if err := os.MkdirAll(targetRootfs, 0755); err != nil {
		return "", OCIImageConfig{}, fmt.Errorf("create rootfs dir: %w", err)
	}

	for i, layer := range manifest.Layers {
		e.log.Info("Extracting layer", zap.Int("index", i), zap.String("digest", layer.Digest[:15]))
		
		blobURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/blobs/%s", repo, layer.Digest)
		req, _ = http.NewRequestWithContext(ctx, "GET", blobURL, nil)
		req.Header.Set("Authorization", "Bearer "+tokenResp.Token)
		bResp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", OCIImageConfig{}, fmt.Errorf("download layer %s: %w", layer.Digest, err)
		}
		if bResp.StatusCode != 200 {
			bResp.Body.Close()
			return "", OCIImageConfig{}, fmt.Errorf("download layer %s failed: %s", layer.Digest, bResp.Status)
		}
		
		if err := extractTarGz(bResp.Body, targetRootfs); err != nil {
			bResp.Body.Close()
			return "", OCIImageConfig{}, fmt.Errorf("extract layer %s: %w", layer.Digest, err)
		}
		bResp.Body.Close()
	}

	e.log.Info("Finished native OCI pull", zap.String("digest", manifest.Config.Digest))
	return manifest.Config.Digest, config.Config, nil
}

func extractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		
		// Basic prevention against path traversal
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) && target != filepath.Clean(dest) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				continue
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				continue
			}
			f.Close()
		case tar.TypeSymlink:
			// Ignore symlink errors
			_ = os.Symlink(header.Linkname, target)
		}
	}
	return nil
}
