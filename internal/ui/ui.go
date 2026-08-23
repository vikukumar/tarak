package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var embeddedDist embed.FS

// Handler returns an http.Handler that serves the embedded Next.js dashboard SPA.
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

		// Check exact file in embedded filesystem
		if cleanPath != "" {
			if f, err := distFS.Open(cleanPath); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}

			// Check directory index (e.g. dashboard/workloads/pods/index.html)
			dirIndex := strings.TrimSuffix(cleanPath, "/") + "/index.html"
			if f, err := distFS.Open(dirIndex); err == nil {
				_ = f.Close()
				if content, err := fs.ReadFile(distFS, dirIndex); err == nil {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(content)
					return
				}
			}

			// Check cleanPath + .html
			htmlPath := cleanPath + ".html"
			if f, err := distFS.Open(htmlPath); err == nil {
				_ = f.Close()
				if content, err := fs.ReadFile(distFS, htmlPath); err == nil {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(content)
					return
				}
			}
		}

		// Fallback to dashboard/index.html or index.html for SPA routes
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

		fileServer.ServeHTTP(w, r)
	})
}
