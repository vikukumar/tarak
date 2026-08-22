// taraks — The dedicated high-performance Tarak Node Agent & Worker Daemon.
//
// taraks runs on worker nodes to manage native container isolation (TCR),
// execute pod workloads, manage loopback networking, and report node health metrics.
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	tarakruntime "github.com/vikukumar/tarak/internal/runtime"
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
		insecure  bool
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
			nodeName = strings.ToLower(nodeName)

			log.Info("starting taraks node worker agent",
				zap.String("version", version.Version),
				zap.String("commit", version.Commit),
				zap.String("node", nodeName),
				zap.String("server", serverURL),
				zap.String("dataDir", dataDir),
			)

			// Initialize native container engine
			engine := tarakruntime.NewEngine(dataDir, log)
			ver := engine.GetRuntimeVersion()

			log.Info("TCR engine initialized",
				zap.String("runtime", ver.RuntimeName),
				zap.String("mode", ver.EngineMode),
				zap.String("os", ver.OS),
				zap.String("arch", ver.Arch),
			)

			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec
				},
				Timeout: 10 * time.Second,
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			// Register worker node with the control plane
			if err := registerWorkerNode(ctx, client, serverURL, nodeName, log); err != nil {
				log.Warn("initial node registration notice (retrying on heartbeat)", zap.Error(err))
			}

			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Info("shutting down taraks agent...")
					return nil
				case <-ticker.C:
					sys := tarakruntime.SampleSystemMetrics()
					log.Debug("node metrics sampled",
						zap.Int("cpuMillicores", int(sys.CPUMillicores)),
						zap.Float64("cpuPercent", sys.CPUPercent),
						zap.Uint64("memoryUsed", sys.UsedMemoryBytes),
					)
					// Send heartbeat to control plane
					_ = sendNodeHeartbeat(ctx, client, serverURL, nodeName, sys)
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
	f.BoolVar(&insecure, "insecure", true, "Allow insecure TLS connection to API server")

	return cmd
}

func registerWorkerNode(ctx context.Context, client *http.Client, serverURL, nodeName string, log *zap.Logger) error {
	nodeObj := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Node",
		"metadata": map[string]interface{}{
			"name": nodeName,
			"labels": map[string]string{
				"tarak.io/hostname":                nodeName,
				"kubernetes.io/hostname":           nodeName,
				"node-role.kubernetes.io/worker":   "true",
				"tarak.io/role":                    "worker",
				"node.kubernetes.io/instance-type": "tarak.worker",
				"kubernetes.io/os":                 runtime.GOOS,
				"kubernetes.io/arch":               runtime.GOARCH,
			},
		},
		"status": map[string]interface{}{
			"phase": "Running",
			"conditions": []map[string]interface{}{
				{
					"type":    "Ready",
					"status":  "True",
					"reason":  "TarakWorkerReady",
					"message": "taraks worker agent active and healthy",
				},
			},
			"nodeInfo": map[string]interface{}{
				"kubeletVersion":          "v" + version.Version + "-tarak",
				"osImage":                 fmt.Sprintf("Tarak Native (%s/%s)", runtime.GOOS, runtime.GOARCH),
				"architecture":            runtime.GOARCH,
				"operatingSystem":         runtime.GOOS,
				"containerRuntimeVersion": "tcr://" + version.Version,
			},
		},
	}

	data, err := json.Marshal(nodeObj)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(serverURL, "/")+"/api/v1/nodes", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict {
		log.Info("worker node successfully registered with control plane", zap.String("node", nodeName))
		return nil
	}

	return fmt.Errorf("node registration returned status: %d", resp.StatusCode)
}

func sendNodeHeartbeat(ctx context.Context, client *http.Client, serverURL, nodeName string, sys tarakruntime.SystemMetrics) error {
	heartbeat := map[string]interface{}{
		"status": map[string]interface{}{
			"phase": "Running",
			"conditions": []map[string]interface{}{
				{
					"type":               "Ready",
					"status":             "True",
					"lastHeartbeatTime": time.Now().UTC().Format(time.RFC3339),
					"reason":             "TarakWorkerHeartbeat",
				},
			},
			"capacity": map[string]interface{}{
				"cpu":    fmt.Sprintf("%dm", sys.CPUMillicores),
				"memory": fmt.Sprintf("%dKi", sys.TotalMemoryBytes/1024),
			},
			"allocatable": map[string]interface{}{
				"cpu":    fmt.Sprintf("%dm", sys.CPUMillicores),
				"memory": fmt.Sprintf("%dKi", sys.FreeMemoryBytes/1024),
			},
		},
	}

	data, _ := json.Marshal(heartbeat)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/api/v1/nodes/%s/status", strings.TrimRight(serverURL, "/"), nodeName), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
