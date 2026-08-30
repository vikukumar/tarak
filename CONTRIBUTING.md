# Contributing to Tarak

First off, thanks for taking the time to contribute! Tarak is a community-driven project and your help is critical to making it the best native container runtime out there.

---

## How Can I Contribute?

### 🐛 Reporting Bugs

Bugs are tracked as [GitHub Issues](https://github.com/vikukumar/tarak/issues). When filing a bug:

- Use a **clear and descriptive title**
- Describe **exact steps to reproduce** the problem
- Include the **Tarak version** (`tarakctl version`) and **OS/platform**
- Attach relevant **server logs** (run `tarak server` output) and **CLI output**
- Describe the **expected vs actual behavior**

### 💡 Suggesting Enhancements

Enhancement suggestions are also tracked as GitHub Issues.

- Use a clear and descriptive title
- Provide a step-by-step description of the suggested enhancement
- Explain why this would benefit most Tarak users

### 🔀 Pull Requests

- Fork the repository and create your branch from `main`
- Fill in the PR description template
- Write or update tests for your changes
- Ensure `go build ./...` and `go test ./...` pass before submitting
- Include screenshots or terminal output whenever possible

---

## Development Setup

### Prerequisites

- **Go 1.22+** — [download](https://go.dev/dl/)
- Git

### Clone & Build

```bash
git clone https://github.com/vikukumar/tarak.git
cd tarak

# Build all binaries
go build -o bin/tarak.exe   ./cmd/tarak     # All-in-one server + CLI
go build -o bin/tarakctl.exe ./cmd/tarakctl  # Standard CLI (tarakctl)
go build -o bin/taraktl.exe  ./cmd/taraktl  # Alias CLI (taraktl)

# Or build everything at once
go build ./...
```

### Project Layout

```
tarak/
├── api/               # Kubernetes-compatible API type definitions
│   ├── apps/v1/       # Deployment, ReplicaSet, StatefulSet, DaemonSet types
│   └── core/v1/       # Pod, Service, ConfigMap, Secret, Namespace types
├── cmd/               # Binary entrypoints (tarak, tarakctl, taraktl)
├── internal/
│   ├── controller/    # Deployment, Pod, Service controller manager (reconcile loops)
│   ├── loadbalancer/  # MetalLB-compatible LoadBalancer IP assignment
│   ├── network/       # DNS resolver, NodePort TCP proxy, CNI bridge
│   ├── policy/        # TarakSecurityPolicy admission engine
│   ├── runtime/       # Container runtime engine + TCR (Tarak Container Runtime)
│   │   ├── exec_shell.go    # Interactive exec shell REPL
│   │   ├── runtime.go       # RunPodContainers, GetLogs (real tail-f), DialPod
│   │   └── tcr/             # TCR: image pull, process lifecycle, nativebridge
│   ├── server/        # Kubernetes-compatible API HTTP server
│   ├── statestore/    # BoltDB-backed state store (pods, services, deployments, etc.)
│   └── version/       # Version metadata
├── pkg/
│   ├── api/           # Admission webhooks, validation
│   └── cli/           # tarakctl/taraktl CLI (all kubectl-compatible commands)
├── docs/              # GitHub Pages documentation site
├── examples/          # Example YAML manifests and Go SDK usage
├── app-sample.yml     # Complete sample app with Deployment, Services, Ingress
└── Makefile           # Build, test, lint targets
```

### Running Tests

```bash
go test ./...
```

### Linting

```bash
# Install golangci-lint if not installed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

golangci-lint run
```

---

## Key Architecture Decisions

### Controller Reconcile Loop
The controller manager reconciles state every **10 seconds** (not 1s) and **skips pods that are already `Running` and healthy** — verified by calling `GetContainerInfo` per container. This avoids the noisy hot-loop that re-pulls images every second.

### Real `exec -it` Interactive Shell
`ExecuteContainerShell` detects interactive shell invocations (`sh`, `bash`, `ash` with no `-c` argument) and:
1. Tries to run the **host shell binary** with `Dir=rootfs` and container environment — giving real bash inside the container filesystem
2. Falls back to a **pure-Go REPL loop** that reads stdin line-by-line and dispatches each command through `executeShellString`

### Real `logs -f` Streaming
`GetLogs` with `follow=true` uses a **200ms polling loop** that opens the log file, seeks to the previously-read offset, copies new bytes to the HTTP response writer (calling `Flush()` after each write), and closes the file — equivalent to `tail -f` without requiring `inotify` or `fsnotify`.

### kubeconfig Auth Priority
`newClient()` resolves credentials in this order:
1. Static bearer `token:`
2. `tokenFile:` (projected service account tokens)
3. `exec:` credential provider plugin (EKS, GKE, AKS)
4. `client-certificate-data:` / `client-certificate:` + key
5. `username:` / `password:` basic auth

---

## Thank You!

Every contribution matters — whether it's a bug report, a typo fix, a new command, or a major feature. Thank you for helping make Tarak better!
