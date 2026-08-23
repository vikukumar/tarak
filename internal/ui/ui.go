package ui

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed dist/*
var embeddedDist embed.FS

// Handler returns an http.Handler that serves the embedded Next.js dashboard SPA and all its assets.
func Handler() http.Handler {
	distFS, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<!DOCTYPE html><html><body><h1>TARAK Cluster Dashboard</h1><p>Embedded Next.js UI Active</p></body></html>"))
		})
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := r.URL.Path
		cleanPath := strings.TrimPrefix(reqPath, "/")
		strippedPath := strings.TrimPrefix(cleanPath, "dashboard/")
		strippedPath = strings.TrimPrefix(strippedPath, "/")

		// Helper to serve file with correct Content-Type
		serveDirect := func(fsPath string) bool {
			if fsPath == "" {
				return false
			}
			f, err := distFS.Open(fsPath)
			if err != nil {
				return false
			}
			_ = f.Close()

			ext := strings.ToLower(filepath.Ext(fsPath))
			switch ext {
			case ".js":
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			case ".css":
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			case ".png":
				w.Header().Set("Content-Type", "image/png")
			case ".jpg", ".jpeg":
				w.Header().Set("Content-Type", "image/jpeg")
			case ".ico":
				w.Header().Set("Content-Type", "image/x-icon")
			case ".svg":
				w.Header().Set("Content-Type", "image/svg+xml")
			case ".json", ".webmanifest":
				w.Header().Set("Content-Type", "application/json")
			case ".woff2":
				w.Header().Set("Content-Type", "font/woff2")
			case ".woff":
				w.Header().Set("Content-Type", "font/woff")
			default:
				if cType := mime.TypeByExtension(ext); cType != "" {
					w.Header().Set("Content-Type", cType)
				}
			}

			w.Header().Set("Access-Control-Allow-Origin", "*")
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/" + fsPath
			fileServer.ServeHTTP(w, r2)
			return true
		}

		// 1. Direct match and asset resolution
		candidatePaths := []string{
			cleanPath,
			strippedPath,
			"assets/" + cleanPath,
			"assets/" + strippedPath,
			strings.TrimPrefix(cleanPath, "assets/"),
			strings.TrimPrefix(strippedPath, "assets/"),
		}

		for _, p := range candidatePaths {
			if p != "" && serveDirect(p) {
				return
			}
		}

		// 2. For asset files (.js, .css, .jpg, .png, .svg, .woff2, .ico, .map, .webmanifest), DO NOT return HTML fallback
		ext := strings.ToLower(filepath.Ext(cleanPath))
		if ext == ".js" || ext == ".css" || ext == ".map" || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".svg" || ext == ".woff2" || ext == ".woff" || ext == ".ico" || ext == ".webmanifest" {
			http.NotFound(w, r)
			return
		}

		// 3. HTML Page route resolution: check /index.html and .html variants
		trimmedClean := strings.Trim(cleanPath, "/")
		trimmedStripped := strings.Trim(strippedPath, "/")

		htmlCandidates := []string{
			trimmedClean + "/index.html",
			trimmedClean + ".html",
			trimmedStripped + "/index.html",
			trimmedStripped + ".html",
			"dashboard/" + trimmedClean + "/index.html",
			"dashboard/" + trimmedClean + ".html",
			"dashboard/" + trimmedStripped + "/index.html",
			"dashboard/" + trimmedStripped + ".html",
		}

		for _, cand := range htmlCandidates {
			if cand == "/index.html" || cand == ".html" {
				continue
			}
			if content, err := fs.ReadFile(distFS, cand); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(content)
				return
			}
		}

		// 4. SPA Fallback: serve dashboard/index.html or index.html
		if strings.HasPrefix(reqPath, "/dashboard") {
			if content, err := fs.ReadFile(distFS, "dashboard/index.html"); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(content)
				return
			}
		}

		if content, err := fs.ReadFile(distFS, "index.html"); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)
			return
		}

		http.NotFound(w, r)
	})
}
