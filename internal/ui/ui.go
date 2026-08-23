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
			if ext == ".js" {
				w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			} else if ext == ".css" {
				w.Header().Set("Content-Type", "text/css; charset=utf-8")
			} else if cType := mime.TypeByExtension(ext); cType != "" {
				w.Header().Set("Content-Type", cType)
			}

			r2 := r.Clone(r.Context())
			r2.URL.Path = "/" + fsPath
			fileServer.ServeHTTP(w, r2)
			return true
		}

		// 1. Direct match for cleanPath (e.g. _next/static/..., assets/..., index.html)
		if serveDirect(cleanPath) {
			return
		}

		// 2. Direct match for strippedPath (e.g. /dashboard/_next/static/... -> _next/static/...)
		if strippedPath != cleanPath && serveDirect(strippedPath) {
			return
		}

		// 3. For asset files (.js, .css, .jpg, .png, .svg, .woff2, .ico, .map), DO NOT return HTML fallback
		ext := strings.ToLower(filepath.Ext(cleanPath))
		if ext == ".js" || ext == ".css" || ext == ".map" || ext == ".png" || ext == ".jpg" || ext == ".svg" || ext == ".woff2" || ext == ".woff" || ext == ".ico" {
			http.NotFound(w, r)
			return
		}

		// 4. HTML Page route resolution: check /index.html and .html variants
		candidates := []string{
			cleanPath + "/index.html",
			cleanPath + ".html",
			strippedPath + "/index.html",
			strippedPath + ".html",
		}

		for _, cand := range candidates {
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

		// 5. SPA Fallback: serve dashboard/index.html or index.html
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
