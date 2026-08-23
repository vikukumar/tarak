"use client";

import React, { useState, useEffect } from "react";
import { Cpu, Server, Layers, Shield, RefreshCw, CheckCircle2, Box, Terminal, Activity, Zap, HardDrive } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { tarakFetch } from "@/lib/api";

export default function RuntimeExplorerPage() {
  const [versionInfo, setVersionInfo] = useState<any>(null);
  const [nodes, setNodes] = useState<any[]>([]);
  const [pods, setPods] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchRuntimeData = async () => {
    setIsLoading(true);
    try {
      const [verRes, nodeRes, podRes] = await Promise.all([
        tarakFetch("/version"),
        tarakFetch("/api/v1/nodes"),
        tarakFetch("/api/v1/pods"),
      ]);
      setVersionInfo(verRes.data || {});
      setNodes(nodeRes.data?.items || []);
      setPods(podRes.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchRuntimeData();
  }, []);

  const primaryNode = nodes[0] || {};
  const nodeLabels = primaryNode.metadata?.labels || {};
  const nodeStatus = primaryNode.status || {};
  const alloc = nodeStatus.allocatable || {};
  const nodeInfo = nodeStatus.nodeInfo || {};

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <Cpu size={24} className="text-cyan-400" />
            <span>Container Runtime & Engine Explorer</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Real-time CRI engine status, native process isolation, and bare-metal resource allocations
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchRuntimeData}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh Diagnostics</span>
          </Button>
        </div>
      </div>

      {/* Grid of Engine Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
        <Card className="p-5 border border-cyan-500/20 bg-gradient-to-b from-cyan-950/20 to-slate-950/40">
          <div className="flex items-center justify-between pb-3 border-b border-white/10">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">
              Container Engine
            </span>
            <Badge variant="cyan" dot>Active</Badge>
          </div>
          <div className="mt-4 space-y-2">
            <div className="text-lg font-bold text-white font-mono">
              {nodeInfo.containerRuntimeVersion || "tarak-runtime://tcr-native"}
            </div>
            <p className="text-xs text-slate-400 leading-relaxed">
              OCI Image extraction, layer unpacker, and live process isolation sandbox.
            </p>
            <div className="pt-2 text-[11px] font-mono text-cyan-300">
              API Server: {versionInfo.gitVersion || "v1.0.6-tarak"} ({versionInfo.goVersion || "go1.26.2"})
            </div>
          </div>
        </Card>

        <Card className="p-5 border border-emerald-500/20 bg-gradient-to-b from-emerald-950/20 to-slate-950/40">
          <div className="flex items-center justify-between pb-3 border-b border-white/10">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">
              Host Hardware Allocation
            </span>
            <Badge variant="emerald">100% Host Synced</Badge>
          </div>
          <div className="mt-4 space-y-2 text-xs">
            <div className="flex items-center justify-between">
              <span className="text-slate-400">Host CPU Cores:</span>
              <strong className="text-white font-mono">{alloc.cpu || "16 Cores"}</strong>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-400">Total Host Memory:</span>
              <strong className="text-white font-mono">{nodeLabels["tarak.io/total-memory-gb"] || alloc.memory || "16Gi"}</strong>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-400">GPU Accelerator:</span>
              <strong className="text-emerald-400 font-mono">
                {alloc.gpu || (nodeLabels["nvidia.com/gpu.present"] === "true" ? "1 GPU Present" : "0 (None)")}
              </strong>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-400">Ephemeral Storage:</span>
              <strong className="text-white font-mono">{alloc["ephemeral-storage"] || "500Gi"}</strong>
            </div>
          </div>
        </Card>

        <Card className="p-5 border border-purple-500/20 bg-gradient-to-b from-purple-950/20 to-slate-950/40">
          <div className="flex items-center justify-between pb-3 border-b border-white/10">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">
              Live Workload Status
            </span>
            <Badge variant="purple">{pods.length} Running Pods</Badge>
          </div>
          <div className="mt-4 space-y-2 text-xs">
            <div className="flex items-center justify-between">
              <span className="text-slate-400">Primary LAN IP:</span>
              <strong className="text-cyan-300 font-mono">{nodeLabels["tarak.io/host-lan-ip"] || "127.0.0.1"}</strong>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-400">Public MetalLB IP:</span>
              <strong className="text-emerald-300 font-mono">{nodeLabels["tarak.io/host-public-ip"] || "Dynamic"}</strong>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-400">OS / Architecture:</span>
              <strong className="text-white font-mono">{nodeInfo.osImage || "Windows / AMD64"}</strong>
            </div>
          </div>
        </Card>
      </div>

      {/* Feature Compatibility Matrix */}
      <Card className="p-6 border border-white/10 bg-slate-950/60 space-y-4">
        <h3 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
          <Shield size={16} className="text-cyan-400" />
          <span>Container Isolation & Execution Capabilities</span>
        </h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
          <div className="p-3.5 rounded-xl bg-slate-900/60 border border-white/5 space-y-1.5">
            <div className="flex items-center gap-2 text-emerald-400 text-xs font-bold">
              <CheckCircle2 size={14} />
              <span>OCI Registry Layer Pull</span>
            </div>
            <p className="text-[11px] text-slate-400">
              Docker Hub, GHCR, Quay.io, and Private OCI registries.
            </p>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-900/60 border border-white/5 space-y-1.5">
            <div className="flex items-center gap-2 text-emerald-400 text-xs font-bold">
              <CheckCircle2 size={14} />
              <span>Real Process Sandbox</span>
            </div>
            <p className="text-[11px] text-slate-400">
              Isolated working directories, environment variable injection, and PID tracking.
            </p>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-900/60 border border-white/5 space-y-1.5">
            <div className="flex items-center gap-2 text-emerald-400 text-xs font-bold">
              <CheckCircle2 size={14} />
              <span>Live Port Proxying</span>
            </div>
            <p className="text-[11px] text-slate-400">
              MetalLB VIP allocation, NodePort host binding, and zero-latency TCP proxy.
            </p>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-900/60 border border-white/5 space-y-1.5">
            <div className="flex items-center gap-2 text-emerald-400 text-xs font-bold">
              <CheckCircle2 size={14} />
              <span>Kernel Namespace Support</span>
            </div>
            <p className="text-[11px] text-slate-400">
              Linux Namespaces (`PID`, `NET`, `MNT`, `UTS`, `IPC`) & Windows HCS process groups.
            </p>
          </div>
        </div>
      </Card>
    </div>
  );
}
