// Tarak — the main entry point.
//
// This binary provides the tarakd server command.
// The tarakctl client CLI is in cmd/tarakctl/main.go.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/server"
	"github.com/vikukumar/tarak/internal/version"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "tarakd",
		Short: "Tarak — cloud-native container orchestration",
		Long: `Tarak is a production-grade, lightweight, high-performance container
orchestration platform designed as an alternative to Kubernetes.

Start the API server:

  tarakd server --data-dir /var/lib/tarak

`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newServerCommand())
	root.AddCommand(newVersionCommand())
	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("tarakd %s\n", version.String())
		},
	}
}

func newServerCommand() *cobra.Command {
	var cfg server.Config
	var logLevel string
	var insecure bool

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the Tarak API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build logger.
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

			cfg.Log = log
			cfg.AllowInsecureAuth = insecure

			srv, err := server.New(cfg)
			if err != nil {
				return fmt.Errorf("create server: %w", err)
			}

			// Handle OS signals for graceful shutdown.
			ctx, stop := signal.NotifyContext(context.Background(),
				syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			log.Info("starting tarakd",
				zap.String("version", version.Version),
				zap.String("commit", version.Commit),
			)

			return srv.Run(ctx)
		},
	}

	f := cmd.Flags()
	f.StringVar(&cfg.BindAddress, "bind-address", "0.0.0.0:6443", "Address and port to listen on")
	f.StringVar(&cfg.IngressHTTPAddress, "ingress-http-addr", "0.0.0.0:8080", "Address and port for Ingress HTTP reverse proxy")
	f.BoolVar(&cfg.CloudflareTunnel, "cloudflare-tunnel", false, "Enable built-in Cloudflare tunnel")
	f.StringVar(&cfg.CloudflareToken, "cloudflare-token", "", "Optional Cloudflare Named Tunnel token")
	f.BoolVar(&cfg.Tailscale, "tailscale", false, "Enable Tailscale mesh networking")
	f.StringVar(&cfg.TailscaleAuthKey, "tailscale-authkey", "", "Tailscale authentication key")
	f.StringVar(&cfg.DataDir, "data-dir", "/var/lib/tarak", "Root directory for persistent data")
	f.StringVar(&cfg.PKIDir, "pki-dir", "", "PKI directory (default: <data-dir>/pki)")
	f.StringVar(&cfg.StateStorePath, "state-store", "", "BoltDB state store path (default: <data-dir>/state.db)")
	f.StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, error")
	f.BoolVar(&insecure, "insecure", false, "Allow unauthenticated requests (dev/test only)")
	f.StringArrayVar(&cfg.SANs, "tls-san", nil, "Additional SANs for the API server TLS certificate")

	return cmd
}
