package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var embeddedDist embed.FS

// Handler returns an http.Handler that serves the embedded React dashboard SPA.
func Handler() http.Handler {
	distFS, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<!DOCTYPE html><html><body><h1>TARAK Cluster Dashboard</h1><p>Embedded UI Active</p></body></html>"))
		})
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := r.URL.Path

		// Strip /dashboard prefix if present
		cleanPath := strings.TrimPrefix(reqPath, "/dashboard")
		cleanPath = strings.TrimPrefix(cleanPath, "/")
		if cleanPath == "" || cleanPath == "index.html" {
			if content, err := fs.ReadFile(distFS, "index.html"); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(content)
				return
			}
		}

		// If the specific file exists in the embedded FS, serve it
		if f, err := distFS.Open(cleanPath); err == nil {
			_ = f.Close()
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/" + cleanPath
			fileServer.ServeHTTP(w, r2)
			return
		}

		// Fallback to index.html for Single Page Application (SPA) client-side routing
		if content, err := fs.ReadFile(distFS, "index.html"); err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
