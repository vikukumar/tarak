// Package client provides a high-performance, typed Go client SDK for Tarak clusters.
//
// Example usage:
//
//	c, err := client.NewClientFromKubeconfig("~/.tarak/config")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	pods, err := c.Pods("default").List(context.Background())
//	for _, p := range pods {
//	    fmt.Printf("Pod: %s (Phase: %s)\n", p.Name, p.Phase)
//	}
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Client is the primary entry point for interacting with the Tarak API server.
type Client struct {
	serverURL  string
	httpClient *http.Client
	token      string
}

// ClientOptions configures the client connection parameters.
type ClientOptions struct {
	ServerURL  string
	CAData     []byte
	CertData   []byte
	KeyData    []byte
	Token      string
	Insecure   bool
	Timeout    time.Duration
}

// NewClient creates a new Tarak API client from programmatic options.
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.ServerURL == "" {
		opts.ServerURL = "https://127.0.0.1:6443"
	}
	opts.ServerURL = strings.TrimRight(opts.ServerURL, "/")

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if opts.Insecure {
		tlsConfig.InsecureSkipVerify = true
	} else if len(opts.CAData) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(opts.CAData) {
			return nil, fmt.Errorf("failed to parse CA certificate PEM")
		}
		tlsConfig.RootCAs = pool
	}

	if len(opts.CertData) > 0 && len(opts.KeyData) > 0 {
		cert, err := tls.X509KeyPair(opts.CertData, opts.KeyData)
		if err != nil {
			return nil, fmt.Errorf("failed to load client keypair: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}

	return &Client{
		serverURL:  opts.ServerURL,
		httpClient: httpClient,
		token:      opts.Token,
	}, nil
}

// NewClientFromKubeconfig creates a Tarak client by parsing a kubeconfig file.
func NewClientFromKubeconfig(kubeconfigPath string) (*Client, error) {
	if kubeconfigPath == "" {
		home, _ := os.UserHomeDir()
		kubeconfigPath = filepath.Join(home, ".tarak", "config")
	}

	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig %q: %w", kubeconfigPath, err)
	}

	var cfg struct {
		Clusters []struct {
			Name    string `yaml:"name"`
			Cluster struct {
				Server                   string `yaml:"server"`
				CertificateAuthorityData []byte `yaml:"certificate-authority-data"`
				InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify"`
			} `yaml:"cluster"`
		} `yaml:"clusters"`
		Users []struct {
			Name string `yaml:"name"`
			User struct {
				ClientCertificateData []byte `yaml:"client-certificate-data"`
				ClientKeyData         []byte `yaml:"client-key-data"`
				Token                 string `yaml:"token"`
			} `yaml:"user"`
		} `yaml:"users"`
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse kubeconfig yaml: %w", err)
	}

	var serverURL string
	var caData, certData, keyData []byte
	var insecure bool
	var token string

	if len(cfg.Clusters) > 0 {
		serverURL = cfg.Clusters[0].Cluster.Server
		caData = cfg.Clusters[0].Cluster.CertificateAuthorityData
		insecure = cfg.Clusters[0].Cluster.InsecureSkipTLSVerify
	}
	if len(cfg.Users) > 0 {
		certData = cfg.Users[0].User.ClientCertificateData
		keyData = cfg.Users[0].User.ClientKeyData
		token = cfg.Users[0].User.Token
	}

	return NewClient(ClientOptions{
		ServerURL: serverURL,
		CAData:    caData,
		CertData:  certData,
		KeyData:   keyData,
		Insecure:  insecure,
		Token:     token,
	})
}

// ServerURL returns the connected API server base URL.
func (c *Client) ServerURL() string {
	return c.serverURL
}

// ─── Generic HTTP helper methods ─────────────────────────────────────────────

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.serverURL+path, nil)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed (%d): %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *Client) post(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed (%d): %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *Client) delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.serverURL+path, nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API delete failed (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}
