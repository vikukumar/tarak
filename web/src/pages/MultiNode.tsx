import React, { useState } from "react";
import { Copy, Check, Server, Shield, Network, ArrowRight } from "lucide-react";
import { Mermaid } from "../components/Mermaid";

export const MultiNodePage: React.FC<{ onNavigate: (tab: string) => void }> = ({ onNavigate }) => {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const copyCode = (code: string, idx: number) => {
    navigator.clipboard.writeText(code);
    setCopiedIndex(idx);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  const multiNodeTopology = `
graph LR
    subgraph ControlPlaneNode["Leader Node: 192.168.1.10"]
        tarakd["tarakd (:6443 / :8443)"]
        boltdb[("Embedded BoltDB")]
        ca["Inbuilt P-256 CA"]
    end

    subgraph Worker1["Worker Node 01: 192.168.1.20"]
        taraks1["taraks Agent"]
        tcr1["TCR Runtime"]
        sub1["Pod Subnet: 10.244.1.0/24"]
    end

    subgraph Worker2["Worker Node 02: 192.168.1.21"]
        taraks2["taraks Agent"]
        tcr2["TCR Runtime"]
        sub2["Pod Subnet: 10.244.2.0/24"]
    end

    tarakd <-->|mTLS 1.3 Heartbeat| taraks1
    tarakd <-->|mTLS 1.3 Heartbeat| taraks2

    sub1 <-->|WireGuard Encrypted Tunnel| sub2
  `;

  return (
    <div className="space-y-10 animate-fade-in max-w-4xl mx-auto">
      {/* Title */}
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/30 text-purple-300 text-xs font-bold uppercase tracking-wider">
          Enterprise Clustering
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          Multi-Node <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-cyan-400">Clustering</span>
        </h1>
        <p className="text-slate-400 max-w-xl mx-auto text-sm sm:text-base">
          Scale your container fleet across bare-metal servers, cloud VMs, and edge gateways with automated WireGuard mesh and mutual TLS.
        </p>
      </div>

      {/* Cluster Diagram */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <h2 className="text-lg font-bold text-white">Cluster Interconnect Architecture</h2>
        <Mermaid chart={multiNodeTopology} id="chart-multinode" />
      </div>

      {/* Step 1: Start tarakd Control Plane */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h3 className="text-base sm:text-lg font-bold text-white">
            1. Start the Control Plane Server (<code className="text-cyan-300">tarakd</code>)
          </h3>
          <span className="px-2.5 py-0.5 rounded-full bg-cyan-500/15 border border-cyan-500/30 text-cyan-300 text-xs font-semibold">
            Master Node
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          On your primary master node (e.g. <code className="text-cyan-300">192.168.1.10</code>), launch the control plane daemon with your cluster join token:
        </p>

        <div className="relative rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400">
            <span className="font-mono">Master Node Bash</span>
            <button
              onClick={() => copyCode("tarakd --listen-addr=0.0.0.0:6443 --token=my-secret-cluster-token --node-ip=192.168.1.10", 1)}
              className="flex items-center gap-1 text-slate-300 hover:text-cyan-300"
            >
              {copiedIndex === 1 ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
              <span>{copiedIndex === 1 ? "Copied" : "Copy"}</span>
            </button>
          </div>
          <pre className="p-4 font-mono text-xs text-cyan-300 overflow-x-auto leading-relaxed">
            tarakd --listen-addr=0.0.0.0:6443 --token=my-secret-cluster-token --node-ip=192.168.1.10
          </pre>
        </div>
      </div>

      {/* Step 2: Join Worker Nodes */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h3 className="text-base sm:text-lg font-bold text-white">
            2. Join Worker Nodes (<code className="text-purple-300">taraks</code>)
          </h3>
          <span className="px-2.5 py-0.5 rounded-full bg-purple-500/15 border border-purple-500/30 text-purple-300 text-xs font-semibold">
            Worker Nodes
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          On each worker machine or VM, execute the worker agent pointing to the master API endpoint:
        </p>

        <div className="relative rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400">
            <span className="font-mono">Worker Node Bash</span>
            <button
              onClick={() => copyCode("taraks --server=https://192.168.1.10:6443 --token=my-secret-cluster-token --node-name=worker-01", 2)}
              className="flex items-center gap-1 text-slate-300 hover:text-cyan-300"
            >
              {copiedIndex === 2 ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
              <span>{copiedIndex === 2 ? "Copied" : "Copy"}</span>
            </button>
          </div>
          <pre className="p-4 font-mono text-xs text-purple-300 overflow-x-auto leading-relaxed">
            taraks --server=https://192.168.1.10:6443 --token=my-secret-cluster-token --node-name=worker-01
          </pre>
        </div>
      </div>

      {/* Step 3: Verify Nodes */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <h3 className="text-base sm:text-lg font-bold text-white">3. Verify Cluster Nodes</h3>
        <p className="text-xs sm:text-sm text-slate-300">Run <code className="text-cyan-300">tarakctl get nodes</code> to inspect cluster health:</p>

        <pre className="p-4 rounded-xl bg-[#04060c] border border-white/10 font-mono text-xs text-slate-300 overflow-x-auto leading-relaxed">
NAME             STATUS   ROLES          AGE   VERSION   INTERNAL-IP
master-node      Ready    controlplane   10m   v1.0.6    192.168.1.10
worker-01        Ready    worker         2m    v1.0.6    192.168.1.20
worker-02        Ready    worker         1m    v1.0.6    192.168.1.21
        </pre>
      </div>
    </div>
  );
};
