// Package main provides the all-in-one tarak binary with cluster bootstrap (init)
// and server run commands.
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/vikukumar/tarak/internal/network"
	"github.com/vikukumar/tarak/internal/runtime"
	"github.com/vikukumar/tarak/internal/runtime/tcr"
	"github.com/vikukumar/tarak/internal/server"
	"github.com/vikukumar/tarak/internal/version"
	"github.com/vikukumar/tarak/pkg/security"
)

func main() {
	// ── TCR Reexec Handlers ───────────────────────────────────────────────────
	// The tarak binary is re-executed as part of the container runtime lifecycle.
	// These special commands are checked BEFORE cobra processes any flags.
	//
	// __tcr_init__: Container init process — sets up namespaces, mounts, chroots,
	//               then exec's the actual container process (nginx, node, etc.)
	// __tcr_exec__: Exec process — enters an existing container's namespaces
	//               and runs a command inside it.
	//
	// These are platform-specific: on Linux they perform real namespace operations;
	// on Windows/macOS they run process-isolated commands.
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case tcr.InitArg:
			if err := tcr.RunContainerInit(); err != nil {
				fmt.Fprintln(os.Stderr, "tcr-init:", err)
				os.Exit(1)
			}
			os.Exit(0)
		case tcr.ExecArg:
			if err := tcr.RunContainerExec(); err != nil {
				fmt.Fprintln(os.Stderr, "tcr-exec:", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	rootCmd := &cobra.Command{
		Use:   "tarak",
		Short: "Tarak — Native Kubernetes-Compatible Container Orchestration Platform",
		Long: `Tarak is an independent, production-grade container orchestration platform
built from first principles with full Kubernetes API compatibility.`,
	}

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newServerCmd())
	rootCmd.AddCommand(newAgentCmd())
	rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// ─── tarak init ───────────────────────────────────────────────────────────────

type initOptions struct {
	dataDir     string
	bindAddress string
	sans        []string
	kubeconfig  string
}

func newInitCmd() *cobra.Command {
	opts := initOptions{
		dataDir:     defaultDataDir(),
		bindAddress: "127.0.0.1:6443",
		sans:        []string{"localhost", "127.0.0.1"},
		kubeconfig:  defaultKubeconfigPath(),
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a new Tarak cluster (PKI, state store, and admin kubeconfig)",
		Long: `Initialize generates the root Certificate Authority (CA), signs API server
and cluster administrator certificates, initializes the state directory,
and writes an admin kubeconfig file ready for use with tarakctl or kubectl.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(opts)
		},
	}

	cmd.Flags().StringVar(&opts.dataDir, "data-dir", opts.dataDir, "Directory to store cluster state and certificates")
	cmd.Flags().StringVar(&opts.bindAddress, "bind-address", opts.bindAddress, "API server advertise host:port")
	cmd.Flags().StringSliceVar(&opts.sans, "tls-san", opts.sans, "Subject Alternative Names (SANs) for the API server certificate")
	cmd.Flags().StringVar(&opts.kubeconfig, "kubeconfig", opts.kubeconfig, "Path to output administrator kubeconfig file")

	return cmd
}

func runInit(opts initOptions) error {
	fmt.Println("================================================================")
	fmt.Println("               Tarak Cluster Initialization                     ")
	fmt.Println("================================================================")
	fmt.Println()

	pkiDir := filepath.Join(opts.dataDir, "pki")
	if err := os.MkdirAll(pkiDir, 0700); err != nil {
		return fmt.Errorf("create pki directory %q: %w", pkiDir, err)
	}

	// 1. Generate Root CA
	fmt.Printf("[1/5] Generating Root Certificate Authority (P-256 ECDSA)...\n")
	ca, err := security.GenerateCA()
	if err != nil {
		return fmt.Errorf("generate CA: %w", err)
	}

	// 2. Sign Server Cert
	fmt.Printf("[2/5] Signing API Server certificate with SANs: %v...\n", opts.sans)
	serverCert, err := ca.SignServerCert(security.ServerCertOptions{
		CommonName: "tarak-apiserver",
		SANs:       opts.sans,
		ValidFor:   365 * 24 * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("sign server cert: %w", err)
	}

	// 3. Sign Admin Client Cert
	fmt.Printf("[3/5] Signing Cluster Admin client certificate (O=system:masters)...\n")
	adminCert, err := ca.SignClientCert(security.ClientCertOptions{
		CommonName:    "tarak-admin",
		Organizations: []string{"system:masters"},
		ValidFor:      365 * 24 * time.Hour,
	})
	if err != nil {
		return fmt.Errorf("sign admin cert: %w", err)
	}

	// Write PKI files to disk
	if err := security.WritePKI(pkiDir, ca, serverCert, adminCert); err != nil {
		return fmt.Errorf("write pki files: %w", err)
	}
	fmt.Printf("      -> Saved PKI files to: %s\n", pkiDir)

	// 4. Generate Kubeconfig
	fmt.Printf("[4/5] Generating Administrator Kubeconfig at: %s...\n", opts.kubeconfig)
	if err := writeKubeconfig(opts.kubeconfig, opts.bindAddress, ca.CertPEM, adminCert.CertPEM, adminCert.KeyPEM); err != nil {
		return fmt.Errorf("write kubeconfig: %w", err)
	}

	// 5. Detect & Display Container Runtime Environment
	fmt.Printf("[5/5] Detecting Container Runtime Environment...\n")
	rtReport := runtime.ProbeHostRuntimes(nil)
	fmt.Printf("      -> Active Engine: %s (%s)\n", rtReport.Name, rtReport.Version)
	fmt.Printf("      -> Mode: %s\n", rtReport.Description)

	fmt.Println()
	fmt.Println("Initialization complete!")
	fmt.Println()
	fmt.Println("To start the Tarak control plane server:")
	fmt.Printf("    tarak server --data-dir %q --bind-address %q\n", opts.dataDir, opts.bindAddress)
	fmt.Println()
	fmt.Println("To interact with your cluster:")
	fmt.Printf("    tarakctl get namespaces\n")
	fmt.Println()
	return nil
}

func writeKubeconfig(path, serverAddr string, caPEM, certPEM, keyPEM []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	serverURL := serverAddr
	if !filepath.IsAbs(serverAddr) && len(serverAddr) > 0 && serverAddr[0] != 'h' {
		serverURL = "https://" + serverAddr
	}

	config := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Config",
		"clusters": []map[string]interface{}{
			{
				"name": "tarak-cluster",
				"cluster": map[string]interface{}{
					"server":                     serverURL,
					"certificate-authority-data": base64.StdEncoding.EncodeToString(caPEM),
				},
			},
		},
		"contexts": []map[string]interface{}{
			{
				"name": "tarak-admin@tarak-cluster",
				"context": map[string]interface{}{
					"cluster":   "tarak-cluster",
					"user":      "tarak-admin",
					"namespace": "default",
				},
			},
		},
		"current-context": "tarak-admin@tarak-cluster",
		"users": []map[string]interface{}{
			{
				"name": "tarak-admin",
				"user": map[string]interface{}{
					"client-certificate-data": base64.StdEncoding.EncodeToString(certPEM),
					"client-key-data":         base64.StdEncoding.EncodeToString(keyPEM),
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, yamlData, 0600)
}

// ─── tarak server ─────────────────────────────────────────────────────────────

func newServerCmd() *cobra.Command {
	var (
		dataDir          string
		bindAddress      string
		ingressHTTPAddr  string
		cloudflareTunnel bool
		cloudflareToken  string
		tailscale        bool
		tailscaleAuthKey string
		sans             []string
		allowInsecure    bool
		logLevel         string
		shutdownTimeout  time.Duration
		cpuLimit         string
		memoryLimit      string
		gpuLimit         string
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the Tarak API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			var zapCfg zap.Config
			if logLevel == "debug" {
				zapCfg = zap.NewDevelopmentConfig()
			} else {
				zapCfg = zap.NewProductionConfig()
			}
			logger, err := zapCfg.Build()
			if err != nil {
				return fmt.Errorf("build logger: %w", err)
			}
			defer logger.Sync() //nolint:errcheck

			srv, err := server.New(server.Config{
				BindAddress:        bindAddress,
				IngressHTTPAddress: ingressHTTPAddr,
				CloudflareTunnel:   cloudflareTunnel,
				CloudflareToken:    cloudflareToken,
				Tailscale:          tailscale,
				TailscaleAuthKey:   tailscaleAuthKey,
				DataDir:            dataDir,
				AllowInsecureAuth:  allowInsecure,
				SANs:               sans,
				CPULimit:           cpuLimit,
				MemoryLimit:        memoryLimit,
				GPULimit:           gpuLimit,
				Log:                logger,
				ShutdownTimeout:    shutdownTimeout,
			})
			if err != nil {
				return fmt.Errorf("initialize server: %w", err)
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			return srv.Run(ctx)
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", defaultDataDir(), "Data directory")
	cmd.Flags().StringVar(&bindAddress, "bind-address", "0.0.0.0:6443", "Bind address:port")
	cmd.Flags().StringVar(&ingressHTTPAddr, "ingress-http-addr", "0.0.0.0:8080", "Ingress HTTP reverse proxy address:port")
	cmd.Flags().BoolVar(&cloudflareTunnel, "cloudflare-tunnel", false, "Enable built-in Cloudflare tunneling")
	cmd.Flags().StringVar(&cloudflareToken, "cloudflare-token", "", "Cloudflare Named Tunnel token")
	cmd.Flags().BoolVar(&tailscale, "tailscale", false, "Enable Tailscale mesh networking")
	cmd.Flags().StringVar(&tailscaleAuthKey, "tailscale-authkey", "", "Tailscale authentication key")
	cmd.Flags().StringSliceVar(&sans, "tls-san", []string{"localhost", "127.0.0.1"}, "TLS SANs")
	cmd.Flags().BoolVar(&allowInsecure, "insecure", false, "Allow unauthenticated access (dev only)")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level: debug|info|warn|error")
	cmd.Flags().DurationVar(&shutdownTimeout, "shutdown-timeout", 30*time.Second, "Graceful shutdown deadline")
	cmd.Flags().StringVar(&cpuLimit, "cpu-limit", "", "CPU cores limit for node allocation (defaults to 100% host CPUs)")
	cmd.Flags().StringVar(&memoryLimit, "memory-limit", "", "Memory limit for node allocation (e.g. 16Gi, defaults to 100% host RAM)")
	cmd.Flags().StringVar(&gpuLimit, "gpu-limit", "", "GPU limit for node allocation (defaults to all host GPUs)")

	return cmd
}

// ─── tarak agent ────────────────────────────────────────────────────────────
func newAgentCmd() *cobra.Command {
	var (
		dataDir   string
		serverURL string
		logLevel  string
		nodeName  string
	)

	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Start the Tarak node worker runtime agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			var log *zap.Logger
			var err error
			if logLevel == "debug" {
				log, err = zap.NewDevelopment()
			} else {
				log, err = zap.NewProduction()
			}
			if err != nil {
				return fmt.Errorf("create logger: %w", err)
			}
			defer log.Sync() //nolint:errcheck

			if nodeName == "" {
				h, _ := os.Hostname()
				nodeName = h
			}

			log.Info("starting tarak worker agent",
				zap.String("version", version.Version),
				zap.String("node", nodeName),
				zap.String("dataDir", dataDir),
				zap.String("server", serverURL),
			)

			// ── Subsystems: TCR Engine, CNI, Micro-CoreDNS, Micro-Kubelet ──
			_ = os.MkdirAll(dataDir, 0755)
			_ = runtime.NewEngine(dataDir, log)
			_ = network.NewInbuiltCNI(network.CNIConfig{}, log)
			dnsServer := network.NewMicroCoreDNS("0.0.0.0", 5353, "cluster.local", log)

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			_ = dnsServer.Start(ctx)
			log.Info("tarak node worker active and listening for pod schedules",
				zap.String("node", nodeName),
				zap.String("runtime", "tcr-native"),
			)

			<-ctx.Done()
			log.Info("shutting down tarak agent...")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "data-dir", defaultDataDir(), "Data directory for containers")
	cmd.Flags().StringVar(&serverURL, "server", "https://127.0.0.1:6443", "Tarak server URL")
	cmd.Flags().StringVar(&nodeName, "node-name", "", "Node name override")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level: debug|info|warn|error")

	return cmd
}

// ─── tarak version ────────────────────────────────────────────────────────────

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print Tarak version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Tarak %s\n", version.String())
		},
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./data"
	}
	return filepath.Join(home, ".tarak", "data")
}

func defaultKubeconfigPath() string {
	if env := os.Getenv("KUBECONFIG"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "./kubeconfig.yaml"
	}
	return filepath.Join(home, ".tarak", "config")
}
