// taraks — The dedicated high-performance Tarak Node Agent & Worker Daemon.
//
// taraks runs on worker nodes to manage native container isolation (TCR),
// execute pod workloads, manage loopback networking, and report node health metrics.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/vikukumar/tarak/internal/runtime"
	"github.com/vikukumar/tarak/internal/runtime/tcr"
	"github.com/vikukumar/tarak/internal/version"
)

func main() {
	// Reexec handler for TCR container init and exec
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

	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "taraks",
		Short: "taraks — Tarak Native Node Runtime & Worker Agent",
		Long: `taraks is the standalone worker node daemon for the Tarak platform.
It provides zero-dependency native container execution, virtual pod bridges,
and high-frequency hardware metrics sampling without Docker or WSL.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newAgentCommand())
	root.AddCommand(newVersionCommand())
	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("taraks %s\n", version.String())
		},
	}
}

func newAgentCommand() *cobra.Command {
	var (
		dataDir   string
		serverURL string
		logLevel  string
		nodeName  string
	)

	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Start the Tarak node worker agent",
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

			log.Info("starting taraks node worker agent",
				zap.String("version", version.Version),
				zap.String("commit", version.Commit),
				zap.String("node", nodeName),
				zap.String("dataDir", dataDir),
			)

			// Initialize native container engine
			engine := runtime.NewEngine(dataDir, log)
			ver := engine.GetRuntimeVersion()

			log.Info("TCR engine initialized",
				zap.String("runtime", ver.RuntimeName),
				zap.String("mode", ver.EngineMode),
				zap.String("os", ver.OS),
				zap.String("arch", ver.Arch),
			)

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Info("shutting down taraks agent...")
					return nil
				case <-ticker.C:
					sys := runtime.SampleSystemMetrics()
					log.Debug("node metrics sampled",
						zap.Int("cpuMillicores", sys.CPUMillicores),
						zap.Float64("cpuPercent", sys.CPUPercent),
						zap.Uint64("memoryUsed", sys.UsedMemoryBytes),
					)
				}
			}
		},
	}

	home, _ := os.UserHomeDir()
	defaultDir := filepath.Join(home, ".tarak", "data")

	f := cmd.Flags()
	f.StringVar(&dataDir, "data-dir", defaultDir, "Root directory for persistent container images and state")
	f.StringVar(&serverURL, "server", "https://127.0.0.1:6443", "Tarak control-plane API server URL")
	f.StringVar(&nodeName, "node-name", "", "Override node name (defaults to hostname)")
	f.StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, error")

	return cmd
}
