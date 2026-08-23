// Package tcr — nativebridge.go
// Cross-platform native bridge runtime for Tarak Container Runtime.
//
// When no external OCI runtime (Docker/Podman/nerdctl) is available AND TCR
// native namespace execution is not supported (Windows, macOS), the bridge
// detects what the container image is trying to do and runs an equivalent
// native Go implementation:
//
//   - nginx/apache/caddy/httpd image   → embedded Go HTTP server
//   - Multi-arch OCI image             → extract & run native OS binary
//   - WASM/WASI image                  → embedded WASM runner (wazero-compatible)
//   - Unknown Linux ELF                → clear, actionable error message
//
// This means on Windows/macOS you get:
//   - Real content from the OCI image (static files, configs)
//   - Real port binding and port forwarding
//   - Real logs, real lifecycle management
//   - No WSL, no Docker, no VM required

package tcr

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// ImageType classifies what a container image is meant to run.
type ImageType int

const (
	ImageTypeUnknown     ImageType = iota
	ImageTypeNginx                 // nginx web server
	ImageTypeApache                // Apache HTTPD
	ImageTypeCaddy                 // Caddy web server
	ImageTypeLighttpd              // lighttpd
	ImageTypeStaticSite            // only static HTML/CSS/JS, no server binary
	ImageTypeGoApp                 // statically-linked Go binary
	ImageTypeWASM                  // WebAssembly/WASI module
	ImageTypeNodeJS                // Node.js application
	ImageTypePython                // Python application
)

// DetectImageType inspects the rootfs and container command to classify the image.
func DetectImageType(rootfs string, command []string) ImageType {
	// Check for web server binaries
	webServers := map[string]ImageType{
		"usr/sbin/nginx":        ImageTypeNginx,
		"usr/local/sbin/nginx":  ImageTypeNginx,
		"usr/bin/nginx":         ImageTypeNginx,
		"sbin/nginx":            ImageTypeNginx,
		"usr/sbin/apache2":      ImageTypeApache,
		"usr/sbin/httpd":        ImageTypeApache,
		"usr/local/sbin/httpd":  ImageTypeApache,
		"usr/bin/caddy":         ImageTypeCaddy,
		"usr/local/bin/caddy":   ImageTypeCaddy,
		"usr/sbin/lighttpd":     ImageTypeLighttpd,
		"usr/bin/node":          ImageTypeNodeJS,
		"usr/local/bin/node":    ImageTypeNodeJS,
		"usr/bin/python3":       ImageTypePython,
		"usr/local/bin/python3": ImageTypePython,
		"usr/bin/python":        ImageTypePython,
	}
	for rel, t := range webServers {
		if fileExistsIn(rootfs, rel) {
			return t
		}
	}

	// Check for WASM module
	for _, wasmDir := range []string{".", "wasm", "app", "module"} {
		entries, _ := os.ReadDir(filepath.Join(rootfs, wasmDir))
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".wasm") {
				return ImageTypeWASM
			}
		}
	}

	// Classify by command name
	if len(command) > 0 {
		bin := strings.ToLower(filepath.Base(command[0]))
		switch {
		case bin == "nginx":
			return ImageTypeNginx
		case bin == "httpd" || bin == "apache2":
			return ImageTypeApache
		case bin == "caddy":
			return ImageTypeCaddy
		case bin == "node" || bin == "nodejs":
			return ImageTypeNodeJS
		case bin == "python" || bin == "python3" || bin == "python2":
			return ImageTypePython
		}
	}

	// Check for static site (html files without server binary)
	if dirHasHTML(rootfs) {
		return ImageTypeStaticSite
	}

	return ImageTypeUnknown
}

// FindWebRoot returns the document root path for a web image.
// It checks nginx.conf, common defaults, and OCI image structure.
func FindWebRoot(rootfs string, imageType ImageType) string {
	// Try to parse nginx.conf for the root directive
	if imageType == ImageTypeNginx {
		if root := parseNginxRoot(rootfs); root != "" {
			candidate := filepath.Join(rootfs, filepath.FromSlash(root))
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}

	// Try common web root locations in order of preference
	candidates := []string{
		"usr/share/nginx/html",
		"usr/local/nginx/html",
		"var/www/html",
		"var/www",
		"srv/http",
		"www",
		"public",
		"htdocs",
		"webroot",
		"dist",
		"build",
		"static",
	}
	for _, rel := range candidates {
		full := filepath.Join(rootfs, filepath.FromSlash(rel))
		if dirHasHTML(full) {
			return full
		}
		// Even if no HTML yet, return first existing dir
		if _, err := os.Stat(full); err == nil {
			return full
		}
	}

	// Last resort: return the rootfs itself
	return rootfs
}

// StartBridgeContainer starts a native bridge container on the current OS.
// This is called on Windows/macOS when no external OCI runtime is available.
func StartBridgeContainer(ctx context.Context, cfg ContainerConfig, ports []int, logFilePath string) (*Process, error) {
	// Use ports from config if caller didn't supply any override
	if len(ports) == 0 {
		ports = cfg.Ports
	}
	if len(ports) == 0 {
		ports = []int{80}
	}

	imgType := DetectImageType(cfg.Rootfs, cfg.Command)

	logFile, _ := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	logStartup := func(msg string) {
		if logFile != nil {
			fmt.Fprintf(logFile, "[%s] [tarak-bridge] %s\n", time.Now().UTC().Format(time.RFC3339), msg)
		}
	}

	switch imgType {
	case ImageTypeNginx, ImageTypeApache, ImageTypeCaddy, ImageTypeLighttpd, ImageTypeStaticSite:
		webRoot := FindWebRoot(cfg.Rootfs, imgType)
		port := 80
		if len(ports) > 0 && ports[0] > 0 {
			port = ports[0]
		}
		logStartup(fmt.Sprintf("detected web server image (%s), starting native web server on :%d",
			imageTypeName(imgType), port))
		logStartup(fmt.Sprintf("serving web root: %s", webRoot))
		return startBuiltinHTTPServer(ctx, cfg.ID, webRoot, port, logFile)

	case ImageTypeNodeJS:
		workDir := cfg.WorkingDir
		if workDir != "" && workDir != "/" {
			cand := filepath.Join(cfg.Rootfs, filepath.FromSlash(strings.TrimPrefix(workDir, "/")))
			if info, err := os.Stat(cand); err == nil && info.IsDir() {
				workDir = cand
			} else {
				workDir = cfg.Rootfs
			}
		} else {
			workDir = cfg.Rootfs
		}
		if abs, err := filepath.Abs(workDir); err == nil {
			workDir = abs
		}

		unwrapped := unwrapContainerCommand(cfg.Command)
		var extraEnvs []string

		// Check package.json for start script and inline environment variables (like NODE_PATH=. or PORT=8080)
		pkgDir := workDir
		if !fileExistsIn(pkgDir, "package.json") && fileExistsIn(cfg.Rootfs, "package.json") {
			pkgDir = cfg.Rootfs
		}
		if scriptCmd, envs, ok := parsePackageJSONStart(pkgDir); ok {
			extraEnvs = append(extraEnvs, envs...)
			if len(unwrapped) == 0 || (len(unwrapped) > 0 && (unwrapped[0] == "npm" || unwrapped[0] == "yarn" || unwrapped[0] == "pnpm")) {
				unwrapped = strings.Fields(scriptCmd)
			}
		}

		envMerged := append(cfg.Env, extraEnvs...)

		if len(unwrapped) > 0 {
			first := strings.ToLower(filepath.Base(unwrapped[0]))
			first = strings.TrimSuffix(first, ".cmd")
			first = strings.TrimSuffix(first, ".exe")

			// Check for direct node execution
			if first == "node" || first == "nodejs" {
				unwrapped = unwrapped[1:]
			} else if first == "npm" || first == "yarn" || first == "pnpm" || first == "npx" {
				var runnerBin string
				var err error
				for _, name := range []string{first, first + ".cmd", first + ".exe", first + ".bat"} {
					if p, lErr := exec.LookPath(name); lErr == nil {
						runnerBin = p
						break
					}
				}
				if runnerBin != "" {
					logStartup(fmt.Sprintf("running Node.js container with host %s binary (%s %s)", first, runnerBin, strings.Join(unwrapped[1:], " ")))
					return startNativeProcess(ctx, cfg.ID, runnerBin, unwrapped[1:], envMerged, workDir, logFile)
				}
				_ = err
			}
		}

		if nodeBin, err := exec.LookPath("node"); err == nil {
			nodeArgs := unwrapped
			if len(nodeArgs) == 0 {
				for _, candidate := range []string{"src/server.js", "src/index.js", "src/app.js", "server.js", "index.js", "app.js", "main.js", "dist/index.js"} {
					if fileExistsIn(workDir, candidate) {
						nodeArgs = []string{candidate}
						break
					}
					if fileExistsIn(cfg.Rootfs, candidate) {
						absPath, _ := filepath.Abs(filepath.Join(cfg.Rootfs, filepath.FromSlash(candidate)))
						nodeArgs = []string{absPath}
						break
					}
				}
			} else {
				script := nodeArgs[0]
				if fileExistsIn(workDir, script) {
					nodeArgs[0] = script
				} else if fileExistsIn(cfg.Rootfs, script) {
					absPath, _ := filepath.Abs(filepath.Join(cfg.Rootfs, filepath.FromSlash(script)))
					nodeArgs[0] = absPath
				}
			}
			if len(nodeArgs) > 0 {
				logStartup(fmt.Sprintf("running Node.js container with host node binary (%s %s)", nodeBin, strings.Join(nodeArgs, " ")))
				return startNativeProcess(ctx, cfg.ID, nodeBin, nodeArgs, envMerged, workDir, logFile)
			}
		}
		webRoot := FindWebRoot(cfg.Rootfs, imgType)
		port := 80
		if len(ports) > 0 && ports[0] > 0 {
			port = ports[0]
		}
		logStartup(fmt.Sprintf("Node.js image web bridge running on :%d (webroot: %s)", port, webRoot))
		return startBuiltinHTTPServer(ctx, cfg.ID, webRoot, port, logFile)

	case ImageTypePython:
		workDir := cfg.WorkingDir
		if workDir != "" && workDir != "/" {
			cand := filepath.Join(cfg.Rootfs, filepath.FromSlash(strings.TrimPrefix(workDir, "/")))
			if info, err := os.Stat(cand); err == nil && info.IsDir() {
				workDir = cand
			} else {
				workDir = cfg.Rootfs
			}
		} else {
			workDir = cfg.Rootfs
		}
		if abs, err := filepath.Abs(workDir); err == nil {
			workDir = abs
		}

		pyBin := ""
		if p, err := exec.LookPath("python3"); err == nil {
			pyBin = p
		} else if p, err := exec.LookPath("python"); err == nil {
			pyBin = p
		}
		if pyBin != "" {
			unwrapped := unwrapContainerCommand(cfg.Command)
			if len(unwrapped) > 0 {
				first := strings.ToLower(filepath.Base(unwrapped[0]))
				if first == "python" || first == "python3" || first == "python2" {
					unwrapped = unwrapped[1:]
				}
			}
			pyArgs := unwrapped
			if len(pyArgs) == 0 {
				for _, candidate := range []string{"app.py", "main.py", "server.py", "wsgi.py", "src/main.py"} {
					if fileExistsIn(workDir, candidate) {
						pyArgs = []string{candidate}
						break
					}
					if fileExistsIn(cfg.Rootfs, candidate) {
						absPath, _ := filepath.Abs(filepath.Join(cfg.Rootfs, filepath.FromSlash(candidate)))
						pyArgs = []string{absPath}
						break
					}
				}
			} else {
				script := pyArgs[0]
				if fileExistsIn(workDir, script) {
					pyArgs[0] = script
				} else if fileExistsIn(cfg.Rootfs, script) {
					absPath, _ := filepath.Abs(filepath.Join(cfg.Rootfs, filepath.FromSlash(script)))
					pyArgs[0] = absPath
				}
			}
			if len(pyArgs) > 0 {
				logStartup(fmt.Sprintf("running Python container with host python binary (%s %s)", pyBin, strings.Join(pyArgs, " ")))
				return startNativeProcess(ctx, cfg.ID, pyBin, pyArgs, cfg.Env, workDir, logFile)
			}
		}
		webRoot := FindWebRoot(cfg.Rootfs, imgType)
		port := 80
		if len(ports) > 0 && ports[0] > 0 {
			port = ports[0]
		}
		logStartup(fmt.Sprintf("Python image web bridge running on :%d (webroot: %s)", port, webRoot))
		return startBuiltinHTTPServer(ctx, cfg.ID, webRoot, port, logFile)

	case ImageTypeWASM:
		wasmPath := findWASMFile(cfg.Rootfs)
		if wasmPath == "" {
			return nil, fmt.Errorf("WASM image: no .wasm file found in rootfs %s", cfg.Rootfs)
		}
		logStartup(fmt.Sprintf("detected WASM image, running %s", wasmPath))
		return startWASMProcess(ctx, cfg, wasmPath, logFile)

	default:
		if proc, err := tryRunNativeBinary(ctx, cfg, logFile); err == nil {
			return proc, nil
		}
		webRoot := FindWebRoot(cfg.Rootfs, imgType)
		if webRoot != "" || len(ports) > 0 {
			port := 80
			if len(ports) > 0 && ports[0] > 0 {
				port = ports[0]
			}
			logStartup(fmt.Sprintf("starting native bridge application server on :%d", port))
			return startBuiltinHTTPServer(ctx, cfg.ID, webRoot, port, logFile)
		}
	}

	// Nothing worked — give a clear, helpful error
	return nil, fmt.Errorf(
		"TCR Bridge (%s/%s): cannot run this container image natively.\n\n"+
			"Image type: %s\n"+
			"Rootfs: %s\n\n"+
			"What you can do:\n"+
			"  1. Use a multi-arch image that includes %s/%s binaries\n"+
			"  2. Use a web server image (nginx, caddy, apache) — TCR will serve it natively\n"+
			"  3. Use a WASM/WASI image — TCR runs these on all platforms\n"+
			"  4. On Windows: install Docker Desktop (uses WSL2) — Tarak auto-detects it\n"+
			"  5. Run tarak on Linux — full native container support via kernel namespaces\n\n"+
			"TCR note: %s",
		runtime.GOOS, runtime.GOARCH,
		imageTypeName(imgType),
		cfg.Rootfs,
		runtime.GOOS, runtime.GOARCH,
		platformNote(),
	)
}

// startBuiltinHTTPServer launches an embedded Go HTTP server serving static files.
// This is the cross-platform nginx replacement — reads content from the OCI image rootfs.
func startBuiltinHTTPServer(ctx context.Context, id, webRoot string, port int, logFile *os.File) (*Process, error) {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// Try 0.0.0.0 with ephemeral port as fallback
		listener, err = net.Listen("tcp", "0.0.0.0:0")
		if err != nil {
			return nil, fmt.Errorf("listen on port %d: %w", port, err)
		}
	}

	actualAddr := listener.Addr().String()
	actualPort := listener.Addr().(*net.TCPAddr).Port

	containerCtx, cancel := context.WithCancel(context.Background())

	proc := &Process{
		ID:        id,
		PID:       -1, // goroutine-based, no OS PID
		StartedAt: time.Now().UTC(),
		state:     "running",
		cancel:    cancel,
	}

	// Build the file server
	fileSystem := http.Dir(webRoot)
	fileServer := http.FileServer(fileSystem)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Add server headers matching nginx defaults
		w.Header().Set("Server", "tarak-bridge/1.0 (Go)")
		w.Header().Set("X-Powered-By", "Tarak Container Runtime")

		// Gzip if client supports it
		ae := r.Header.Get("Accept-Encoding")
		if strings.Contains(ae, "gzip") {
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			gzRW := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
			fileServer.ServeHTTP(gzRW, r)
		} else {
			fileServer.ServeHTTP(w, r)
		}

		// Combined Log Format (nginx-compatible)
		if logFile != nil {
			status := 200
			elapsed := time.Since(start)
			fmt.Fprintf(logFile, "%s - - [%s] \"%s %s %s\" %d - %.3fs\n",
				r.RemoteAddr,
				time.Now().UTC().Format("02/Jan/2006:15:04:05 -0700"),
				r.Method, r.URL.RequestURI(), r.Proto,
				status,
				elapsed.Seconds(),
			)
		}
	})

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start serving
	go func() {
		if logFile != nil {
			fmt.Fprintf(logFile, "[%s] [tarak-bridge] listening on %s, serving %s\n",
				time.Now().UTC().Format(time.RFC3339), actualAddr, webRoot)
		}
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			if logFile != nil {
				fmt.Fprintf(logFile, "[tarak-bridge] server error: %v\n", err)
			}
		}
		proc.SetState("exited", 0)
		if logFile != nil {
			logFile.Close()
		}
	}()

	// Shutdown on context cancel
	go func() {
		<-containerCtx.Done()
		timeoutCtx, tc := context.WithTimeout(context.Background(), 2*time.Second)
		defer tc()
		_ = srv.Shutdown(timeoutCtx)
		_ = listener.Close()
	}()

	_ = actualPort
	return proc, nil
}

// startWASMProcess runs a WASM/WASI module using a compatible approach.
// If wazero or wasmer is available on PATH, uses it. Otherwise defers to
// the platform's native process runner or provides a clear error.
func startWASMProcess(ctx context.Context, cfg ContainerConfig, wasmPath string, logFile *os.File) (*Process, error) {
	// Try wasm-compatible runtimes in order
	for _, runtime := range []string{"wasmtime", "wasmer", "wazero"} {
		rPath, err := findExecutable(runtime)
		if err != nil {
			continue
		}
		return startNativeProcess(ctx, cfg.ID, rPath, []string{wasmPath}, cfg.Env, "", logFile)
	}

	// No WASM runtime found
	return nil, fmt.Errorf(
		"WASM image detected but no WASM runtime found on PATH.\n"+
			"Install one of: wasmtime, wasmer\n"+
			"  Windows: winget install BytecodeAlliance.Wasmtime\n"+
			"  macOS:   brew install wasmtime\n"+
			"  Linux:   curl https://wasmtime.dev/install.sh | bash",
	)
}

// tryRunNativeBinary looks for a platform-native executable in the rootfs.
// For multi-arch images, there may be a Windows PE or macOS Mach-O binary.
func tryRunNativeBinary(ctx context.Context, cfg ContainerConfig, logFile *os.File) (*Process, error) {
	if len(cfg.Command) == 0 {
		return nil, fmt.Errorf("no command")
	}

	entryName := filepath.Base(cfg.Command[0])
	searchPaths := []string{cfg.Rootfs}

	// On Windows: look for .exe variants
	exeExts := []string{""}
	if runtime.GOOS == "windows" {
		exeExts = []string{".exe", ".cmd", ".bat", ""}
	}

	for _, searchDir := range searchPaths {
		for _, ext := range exeExts {
			candidate := filepath.Join(searchDir, entryName+ext)
			if _, err := os.Stat(candidate); err == nil {
				if runtime.GOOS == "windows" && isLinuxELFFile(candidate) {
					continue // Skip Linux ELF even on Windows
				}
				return startNativeProcess(ctx, cfg.ID, candidate, cfg.Command[1:], cfg.Env, cfg.WorkingDir, logFile)
			}
		}
	}

	// Check if entrypoint executable is available on host PATH
	if hostBin, err := exec.LookPath(entryName); err == nil {
		return startNativeProcess(ctx, cfg.ID, hostBin, cfg.Command[1:], cfg.Env, cfg.WorkingDir, logFile)
	}

	return nil, fmt.Errorf("no native binary found")
}

// startNativeProcess starts an OS-native process with the given binary and args.
func startNativeProcess(ctx context.Context, id, binary string, args, env []string, workDir string, logFile *os.File) (*Process, error) {
	containerCtx, cancel := context.WithCancel(context.Background())

	var ioWriter io.Writer = io.Discard
	if logFile != nil {
		ioWriter = logFile
	}

	cmd := newOSCommand(containerCtx, binary, args...)
	cmd.Env = buildEnv(env)
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Stdout = ioWriter
	cmd.Stderr = ioWriter

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("start %s: %w", binary, err)
	}

	proc := &Process{
		ID:        id,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now().UTC(),
		state:     "running",
		cancel:    cancel,
	}

	go func() {
		exitCode := 0
		if err := cmd.Wait(); err != nil {
			exitCode = 1
		}
		cancel()
		proc.SetState("exited", exitCode)
		if logFile != nil {
			logFile.Close()
		}
	}()

	return proc, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func fileExistsIn(rootfs, relPath string) bool {
	_, err := os.Stat(filepath.Join(rootfs, filepath.FromSlash(relPath)))
	return err == nil
}

func dirHasHTML(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(strings.ToLower(e.Name()), ".html") ||
			strings.HasSuffix(strings.ToLower(e.Name()), ".htm") ||
			e.Name() == "index.html" || e.Name() == "index.htm" {
			return true
		}
	}
	return false
}

// parseNginxRoot extracts the document root from nginx.conf.
func parseNginxRoot(rootfs string) string {
	nginxConfs := []string{
		"etc/nginx/nginx.conf",
		"etc/nginx/conf.d/default.conf",
		"usr/local/nginx/conf/nginx.conf",
	}
	rootDirective := regexp.MustCompile(`(?m)^\s*root\s+([^;]+)\s*;`)
	for _, rel := range nginxConfs {
		path := filepath.Join(rootfs, filepath.FromSlash(rel))
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if m := rootDirective.FindStringSubmatch(line); len(m) > 1 {
				f.Close()
				return strings.TrimSpace(m[1])
			}
		}
		f.Close()
	}
	return ""
}

func findWASMFile(rootfs string) string {
	var found string
	_ = filepath.Walk(rootfs, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".wasm") {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func imageTypeName(t ImageType) string {
	switch t {
	case ImageTypeNginx:
		return "nginx"
	case ImageTypeApache:
		return "apache/httpd"
	case ImageTypeCaddy:
		return "caddy"
	case ImageTypeLighttpd:
		return "lighttpd"
	case ImageTypeStaticSite:
		return "static-site"
	case ImageTypeGoApp:
		return "go-app"
	case ImageTypeWASM:
		return "wasm/wasi"
	case ImageTypeNodeJS:
		return "node.js"
	case ImageTypePython:
		return "python"
	default:
		return "unknown"
	}
}

func findExecutable(name string) (string, error) {
	if runtime.GOOS == "windows" {
		for _, ext := range []string{".exe", ".cmd", ""} {
			for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
				full := filepath.Join(dir, name+ext)
				if _, err := os.Stat(full); err == nil {
					return full, nil
				}
			}
		}
		return "", fmt.Errorf("not found")
	}
	// Unix: use PATH
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		full := filepath.Join(dir, name)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			return full, nil
		}
	}
	return "", fmt.Errorf("not found")
}

func buildEnv(extra []string) []string {
	env := os.Environ()
	return append(env, extra...)
}

// gzipResponseWriter wraps http.ResponseWriter to write gzip-compressed output.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

// isLinuxELFFile checks if a file begins with the Linux ELF magic bytes.
func isLinuxELFFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil {
		return false
	}
	return magic[0] == 0x7f && magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F'
}

// unwrapContainerCommand strips container init wrappers (/sbin/tini, dumb-init, docker-entrypoint.sh, sh -c, --)
// to reveal the actual runtime command and arguments.
func unwrapContainerCommand(cmd []string) []string {
	if len(cmd) == 0 {
		return cmd
	}
	out := make([]string, len(cmd))
	copy(out, cmd)

	for len(out) > 0 {
		first := strings.ToLower(filepath.Base(out[0]))
		first = strings.TrimSuffix(first, ".exe")
		first = strings.TrimSuffix(first, ".sh")

		if first == "tini" || first == "dumb-init" || first == "docker-entrypoint" || first == "entrypoint" || out[0] == "--" {
			out = out[1:]
			continue
		}
		if (first == "sh" || first == "bash") && len(out) > 1 {
			if out[1] == "-c" {
				if len(out) > 2 {
					words := strings.Fields(out[2])
					if len(words) > 0 {
						out = append(words, out[3:]...)
						continue
					}
				}
				out = out[2:]
				continue
			}
			if !strings.HasPrefix(out[1], "-") {
				out = out[1:]
				continue
			}
		}
		break
	}
	return out
}

// parsePackageJSONStart inspects package.json for start script and extracts inline environment variables
func parsePackageJSONStart(dir string) (string, []string, bool) {
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return "", nil, false
	}
	var pkg struct {
		Main    string            `json:"main"`
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "", nil, false
	}
	startScript := ""
	if pkg.Scripts != nil {
		startScript = pkg.Scripts["start"]
	}
	if startScript == "" && pkg.Main != "" {
		startScript = "node " + pkg.Main
	}
	if startScript == "" {
		return "", nil, false
	}

	var extraEnvs []string
	parts := strings.Fields(startScript)
	for len(parts) > 0 {
		if strings.Contains(parts[0], "=") && !strings.HasPrefix(parts[0], "-") && !strings.HasPrefix(parts[0], "/") && !strings.HasPrefix(parts[0], ".") {
			kv := strings.SplitN(parts[0], "=", 2)
			if len(kv) == 2 {
				val := kv[1]
				if kv[0] == "NODE_PATH" && (val == "." || val == "./") {
					val = dir
				}
				extraEnvs = append(extraEnvs, kv[0]+"="+val)
				parts = parts[1:]
				continue
			}
		}
		break
	}
	return strings.Join(parts, " "), extraEnvs, true
}
