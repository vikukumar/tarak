import React from "react";
import { Mermaid } from "../components/Mermaid";
import { Layers, Shield, Cpu, Network, Database, Zap, Activity } from "lucide-react";

export const ArchitecturePage: React.FC = () => {
  const controlPlaneChart = `
graph TD
    subgraph ClientLayer["CLI, UI & SDK Access"]
        CLI["tarakctl / taraktl CLI"]
        UI["Embedded Next.js 16 UI"]
        SDK["Go SDK pkg/client"]
    end

    subgraph ControlPlane["tarakd — Control Plane Core"]
        API["HTTPS REST API Server (:6443)"]
        AUTH["ECDSA P-256 PKI / Auth Middleware"]
        SCHED["Deterministic Pod Scheduler"]
        CTRL["Workload Controllers (Apps/Batch)"]
        BOLT[("Embedded BoltDB Engine (ACID Store)")]
        TUNNEL_SRV["Inbuilt WireGuard & Mesh Gateway"]
    end

    subgraph WorkerNodes["Worker Node Infrastructure (taraks)"]
        subgraph Node1["Worker Node 01"]
            KUB1["taraks Node Agent"]
            TCR1["TCR Native Runtime Sandbox"]
            PODS1["Multi-Container Pod Workloads"]
            NET1["Virtual Pod Bridge (10.244.1.0/24)"]
        end

        subgraph Node2["Worker Node 02"]
            KUB2["taraks Node Agent"]
            TCR2["TCR Native Runtime Sandbox"]
            PODS2["Multi-Container Pod Workloads"]
            NET2["Virtual Pod Bridge (10.244.2.0/24)"]
        end
    end

    CLI -->|mTLS HTTPS| API
    UI -->|WebSocket / REST| API
    SDK -->|REST API| API

    API --> AUTH
    AUTH --> SCHED
    SCHED --> CTRL
    CTRL --> BOLT

    API <-->|Mutual TLS 1.3 Heartbeat| KUB1
    API <-->|Mutual TLS 1.3 Heartbeat| KUB2

    KUB1 --> TCR1
    TCR1 --> PODS1
    PODS1 --- NET1

    KUB2 --> TCR2
    TCR2 --> PODS2
    PODS2 --- NET2

    NET1 <-->|MetalLB VIP & Encrypted Mesh| NET2
  `;

  const runtimeChart = `
sequenceDiagram
    autonumber
    participant CLI as tarakctl / API Server
    participant Engine as TCR Runtime Engine
    participant Reg as OCI Registry (Docker Hub / GHCR)
    participant Layer as Layer Cache & Rootfs Unpacker
    participant Sandbox as Process Sandbox (cgroups/Win HCS)
    participant Logs as Real Log Stream (/stdout.log)

    CLI->>Engine: RunPod(Spec: Multi-Containers, InitContainers)
    Engine->>Reg: Pull Manifest & Layer Blobs
    Reg-->>Engine: Stream Compressed Tar Layers
    Engine->>Layer: Unpack OCI Layers to rootfs/
    Engine->>Sandbox: Spawn Process (Entrypoint + Cmd + Envs)
    Note over Sandbox: Isolated working directory, Port forwarders, and Env vars
    Sandbox->>Logs: Stream Process stdout/stderr
    Engine-->>CLI: Pod Status: Running (PID, IP, Ready)
  `;

  const meshChart = `
flowchart LR
    subgraph IngressLayer["Edge Gateways & Ingress"]
        CF["Cloudflare Tunnel (Zero-Trust HTTPS)"]
        TS["Tailscale Private WireGuard Mesh"]
        MLB["MetalLB Bare-Metal VIP Announcer (:80/:443)"]
    end

    subgraph ServiceMesh["Tarak Zero-Trust Service Mesh"]
        GW["Tarak Ingress Gateway (SNI / Path Router)"]
        
        subgraph PolicyEngine["Mesh Policy Enforcement"]
            TP["TrafficPermissions (mTLS Verification)"]
            EP["Egress Passthrough (TLS 1.3 Auto-Origination)"]
            TR["TrafficRoutes (Canary 90/10 Split)"]
        end

        subgraph Workloads["Mesh Workloads"]
            V1["Microservice v1 (90% Stable)"]
            V2["Microservice v2 (10% Canary)"]
            EXT["External APIs (Stripe / OpenAI / AWS)"]
        end
    end

    CF --> GW
    TS --> GW
    MLB --> GW

    GW --> PolicyEngine
    PolicyEngine -->|90% Traffic Split| V1
    PolicyEngine -->|10% Canary Split| V2
    V1 -->|Auto TLS 1.3 Origination| EXT
  `;

  const telemetryChart = `
flowchart TD
    subgraph KernelSampling["Kernel-Level Resource Sampling"]
        WIN["Windows Kernel32 (GetSystemTimes + GlobalMemoryStatusEx)"]
        LNX["Linux Kernel (/proc/stat + /proc/meminfo)"]
        CGROUP["Cgroups v2 / Windows HCS Job Objects"]
    end

    subgraph Collector["Tarak Telemetry Engine"]
        SAMPLER["SampleSystemMetrics() Live Collector"]
        PROM["Prometheus /metrics Exporter"]
        K8S_METRICS["/apis/metrics.k8s.io/v1beta1 (Node & Pod Telemetry)"]
    end

    subgraph Consumers["Observability Visualizers"]
        DASH["Live Next.js 16 Dashboard (/dashboard/observability/metrics)"]
        GRAFANA["Prometheus / Grafana Scrapers"]
        HUBBLE["Hubble L4/L7 Network Flow Analyzer"]
    end

    WIN --> SAMPLER
    LNX --> SAMPLER
    CGROUP --> SAMPLER

    SAMPLER --> PROM
    SAMPLER --> K8S_METRICS

    PROM --> GRAFANA
    K8S_METRICS --> DASH
    SAMPLER --> HUBBLE
  `;

  return (
    <div className="space-y-10 animate-fade-in">
      {/* Title */}
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-bold uppercase tracking-wider">
          System Internals & Topology
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          Cluster & Runtime <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-indigo-400 to-purple-400">Architecture</span>
        </h1>
        <p className="text-slate-400 max-w-2xl mx-auto text-sm sm:text-base leading-relaxed">
          Tarak is engineered from first principles in Go as a zero-dependency, ultra-low overhead container orchestration platform.
        </p>
      </div>

      {/* 1. Control Plane & Multi-Node Cluster */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <span className="text-xs font-mono text-cyan-400 uppercase tracking-widest font-bold">Topology 01</span>
            <h2 className="text-xl font-bold text-white">Control Plane & Multi-Node Cluster Topology</h2>
          </div>
          <span className="px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/30 text-purple-300 text-xs font-semibold">
            Zero-etcd • Embedded BoltDB
          </span>
        </div>
        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Separates control plane into a high-throughput API server (<code className="text-cyan-300">tarakd</code>) and worker agents (<code className="text-cyan-300">taraks</code>), or runs unified inside a single binary (<code className="text-cyan-300">tarak</code>).
        </p>

        <Mermaid chart={controlPlaneChart} id="chart-control-plane" />

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2">
          <div className="p-4 rounded-xl bg-slate-950/70 border border-white/5 space-y-1">
            <strong className="text-white text-xs font-bold block">Deterministic Scheduler</strong>
            <p className="text-slate-400 text-xs leading-relaxed">Sub-5ms placement algorithm evaluating memory pressure, CPU allocation, and port bindings.</p>
          </div>
          <div className="p-4 rounded-xl bg-slate-950/70 border border-white/5 space-y-1">
            <strong className="text-white text-xs font-bold block">Embedded BoltDB Engine</strong>
            <p className="text-slate-400 text-xs leading-relaxed">Single-file ACID transactional database with monotonic revision locks, eliminating etcd.</p>
          </div>
          <div className="p-4 rounded-xl bg-slate-950/70 border border-white/5 space-y-1">
            <strong className="text-white text-xs font-bold block">Zero-Trust WireGuard Mesh</strong>
            <p className="text-slate-400 text-xs leading-relaxed">Automatic point-to-point WireGuard and Tailscale overlay for cross-cloud hybrid clusters.</p>
          </div>
        </div>
      </div>

      {/* 2. TCR Container Runtime Engine */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <span className="text-xs font-mono text-emerald-400 uppercase tracking-widest font-bold">Topology 02</span>
            <h2 className="text-xl font-bold text-white">Tarak Container Runtime (TCR) Engine</h2>
          </div>
          <span className="px-3 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs font-semibold">
            Native Process Isolation
          </span>
        </div>
        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          TCR pulls OCI images directly from registries, unpacks filesystem layers, and spawns isolated process sandboxes using native OS primitives with dedicated stdout/stderr log streaming.
        </p>

        <Mermaid chart={runtimeChart} id="chart-runtime" />
      </div>

      {/* 3. Service Mesh & Ingress Gateway */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <span className="text-xs font-mono text-purple-400 uppercase tracking-widest font-bold">Topology 03</span>
            <h2 className="text-xl font-bold text-white">Universal Service Mesh & Ingress Gateway</h2>
          </div>
          <span className="px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-xs font-semibold">
            mTLS Zero-Trust • Canary Routing
          </span>
        </div>
        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Built-in Service Mesh plane providing mTLS encryption, TrafficPermissions policy enforcement, Egress Passthrough rules, and weighted Canary traffic splitting.
        </p>

        <Mermaid chart={meshChart} id="chart-mesh" />
      </div>

      {/* 4. Real-Time Hardware Telemetry Pipeline */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <span className="text-xs font-mono text-cyan-400 uppercase tracking-widest font-bold">Topology 04</span>
            <h2 className="text-xl font-bold text-white">Real-Time Hardware Telemetry Pipeline</h2>
          </div>
          <span className="px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/30 text-purple-300 text-xs font-semibold">
            Prometheus • Hubble Flows
          </span>
        </div>
        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Physical kernel telemetry sampled directly via OS syscalls (<code className="text-cyan-300">GlobalMemoryStatusEx</code>, <code className="text-cyan-300">GetSystemTimes</code>, and <code className="text-cyan-300">/proc/stat</code>) streamed into Prometheus metrics and dashboard visualizers.
        </p>

        <Mermaid chart={telemetryChart} id="chart-telemetry" />
      </div>
    </div>
  );
};
