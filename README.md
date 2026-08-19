<p align="center">
  <img src="assets/tarak_logo.jpg" alt="Tarak Logo" width="600" />
</p>

# Tarak

**Tarak** is a lightning-fast, lightweight, and native container orchestration platform that works everywhere—without requiring Docker Desktop, WSL, or heavy VMs. Compatible with Kubernetes APIs, but reimagined for pure speed and simplicity.

![Go](https://img.shields.io/badge/go-1.22+-00ADD8.svg?style=flat-square&logo=go)
![Platform](https://img.shields.io/badge/platform-windows%20%7C%20linux%20%7C%20macos-lightgrey?style=flat-square)
![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)

## 🚀 Features

* **Native OCI Image Extraction**: Directly pulls, extracts, and caches container layers from Docker Hub natively on your host OS. No daemon required!
* **Zero Dependencies**: Drop Docker Desktop, Podman, and WSL. Tarak bridges networking natively using Virtual Pod Bridge technology.
* **Kubernetes API Compatibility**: Fully compatible with standard Kubernetes deployments, pods, services, and namespaces.
* **Single Binary Deployment**: Both the control plane and node agent are bundled into one lightning-fast `tarak` executable.
* **Tarakctl CLI**: Comes with a drop-in replacement for `kubectl`, built specifically for native performance.

## 🛠️ Quick Start

### 1. Start the Server
Start the Tarak API Server and native node runtime locally:
```bash
tarak server
```

### 2. Deploy an App
Using `tarakctl`, deploy a sample workload. It automatically extracts and runs OCI containers:
```bash
tarakctl apply -f app-sample.yml
tarakctl get pods -o wide
```

### 3. Access your Workload
```bash
tarakctl port-forward pod/web-app-pod 8080:80
```
Visit `http://localhost:8080` in your browser!

## 📦 Architecture

Tarak features a tightly-coupled Controller Manager and native Container Engine. The `TCR` (Tarak Container Runtime) directly translates OCI images into isolated native processes mapped to specific loopback namespaces.

## 🤝 Contributing
Contributions make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**. Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## 🛡️ Security
If you discover a security vulnerability within Tarak, please refer to [SECURITY.md](SECURITY.md) for reporting guidelines.

## 📄 License
Distributed under the MIT License. See [LICENSE](LICENSE) for more information.
