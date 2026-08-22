<p align="center">
  <img src="assets/tarak_logo.jpg" alt="Tarak Logo" width="600" />
</p>

# Tarak

**Tarak** is a lightning-fast, lightweight, and native container orchestration platform that works everywhere—without requiring Docker Desktop, WSL, or heavy VMs. Compatible with Kubernetes APIs, but reimagined for pure speed, minimal memory usage, and zero external dependencies.

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8.svg?style=flat-square&logo=go)](https://go.dev/)
[![Platform](https://img.shields.io/badge/platform-windows%20%7C%20linux%20%7C%20macos-lightgrey?style=flat-square)](https://vikukumar.github.io/tarak/)
[![Architecture](https://img.shields.io/badge/arch-amd64%20%7C%20arm64-blueviolet?style=flat-square)](https://vikukumar.github.io/tarak/)
[![Release](https://img.shields.io/github/v/tag/vikukumar/tarak?style=flat-square&color=success&label=release)](https://github.com/vikukumar/tarak/releases)
[![Docs](https://img.shields.io/badge/docs-online-blue?style=flat-square)](https://vikukumar.github.io/tarak/)
[![Go Reference](https://pkg.go.dev/badge/github.com/vikukumar/tarak.svg)](https://pkg.go.dev/github.com/vikukumar/tarak/pkg/client)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)

---

## ⚡ Quick Installation

### Linux & macOS (One-Line Install)
```bash
curl -fsSL https://raw.githubusercontent.com/vikukumar/tarak/main/install.sh | bash
```

### Windows (PowerShell One-Line Install)
```powershell
irm https://raw.githubusercontent.com/vikukumar/tarak/main/install.ps1 | iex
```

### Go SDK Package (for developers)
```bash
go get -u github.com/vikukumar/tarak/pkg/client@latest
```

---

## 🚀 Why Tarak vs Kubernetes & K3s?

| Feature / Metric | ⚡ Tarak | ⛵ K3s | ☸️ Standard Kubernetes (K8s) |
|---|---|---|---|
| **Idle RAM Footprint** | **< 25 MB** | ~500 MB | ~1.2 GB+ |
| **Cold Start Time** | **< 180 ms** | 10 – 20 sec | 45 – 90 sec |
| **Windows Native Isolation** | **Native (Zero WSL / No Hyper-V)** | Requires WSL2 | Requires Docker/WSL2 |
| **External Dependencies** | **0 (Pure Go Standalone)** | containerd, runc | containerd/CRI-O, etcd, CNI |
| **Kubernetes API Compatibility** | **100% Core K8s API Compliant** | 100% Compliant | Standard |
| **Cross-Platform Support** | Windows / Linux / macOS (ARM64 & AMD64) | Linux only (WSL on Win) | Linux only (WSL on Win) |

---

## 📦 Included Binaries in Release

Each release includes ready-to-run binaries for all platforms and architectures:

* ⚡ **`tarak`**: The **All-in-One** unified runtime (Server + Agent + Init + CLI in one single executable).
* 🖥️ **`tarakd`**: The dedicated standalone API server & control-plane daemon.
* 🤖 **`taraks`**: The dedicated lightweight worker node runtime agent.
* 🎮 **`tarakctl`** (alias **`taraktl`**): The high-performance Kubernetes-compatible CLI control tool.

---

## 🛠️ Quick Start

### 1. Start the Server
Start the Tarak API Server and native node runtime locally:
```bash
tarak server
```

### 2. Deploy an App
Using `tarakctl`, deploy a sample workload. It automatically extracts and runs real OCI containers:
```bash
tarakctl apply -f app-sample.yml
tarakctl get pods -n demo -o wide
```

### 3. Access your Workload
```bash
tarakctl port-forward pod/web-app-pod 8080:80 -n demo
```
Visit `http://localhost:8080` in your browser!

---

## 🔌 Reusable Go Client SDK Example

You can embed Tarak orchestration directly into your Go applications using `pkg/client`:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/vikukumar/tarak/pkg/client"
)

func main() {
    c, err := client.NewClientFromKubeconfig("")
    if err != nil {
        log.Fatal(err)
    }

    pods, err := c.Pods("default").List(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    for _, p := range pods {
        fmt.Printf("Pod: %s (Phase: %s, IP: %s)\n", p.Name, p.Phase, p.IP)
    }
}
```

---

## 📖 Full Documentation & GitHub Pages

Visit the live interactive documentation portal:
👉 **[https://vikukumar.github.io/tarak/](https://vikukumar.github.io/tarak/)**

* 📚 [Getting Started Guide](https://vikukumar.github.io/tarak/getting-started.html)
* 🏛️ [System Architecture](https://vikukumar.github.io/tarak/architecture.html)
* 💻 [CLI Reference](https://vikukumar.github.io/tarak/cli-reference.html)
* 📦 [Releases & Changelog Explorer](https://vikukumar.github.io/tarak/releases.html)

---

## 🤝 Contributing
Contributions make the open source community such an amazing place to learn, inspire, and create. Please read our [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## 🛡️ Security
If you discover a security vulnerability within Tarak, please refer to [SECURITY.md](SECURITY.md).

## 📄 License
Distributed under the MIT License. See [LICENSE](LICENSE) for more information.
