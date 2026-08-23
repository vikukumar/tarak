"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  Sparkles,
  Box,
  Server,
  Globe,
  Radio,
  Cpu,
  Activity,
  Layers,
  Terminal,
  ArrowUpRight,
  ShieldCheck,
  Zap,
  Network,
  Cloud,
  CheckCircle2,
  RefreshCw,
} from "lucide-react";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";

export default function ClusterOverviewPage() {
  const { selectedNamespace, clusterInfo, refresh, isLoading } = useClusterState();
  const [pods, setPods] = useState<any[]>([]);
  const [nodes, setNodes] = useState<any[]>([]);

  useEffect(() => {
    async function loadData() {
      const [pRes, nRes] = await Promise.all([
        tarakFetch(`/api/v1/namespaces/${selectedNamespace}/pods`),
        tarakFetch("/api/v1/nodes"),
      ]);
      setPods(pRes.data?.items || []);
      setNodes(nRes.data?.items || []);
    }
    loadData();
  }, [selectedNamespace]);

  return (
    <div className="space-y-6">
      {/* Hero Cluster Status Banner */}
      <div className="relative rounded-2xl p-6 md:p-8 bg-gradient-to-r from-cyan-950/40 via-slate-900/60 to-indigo-950/40 border border-cyan-500/20 overflow-hidden shadow-2xl">
        <div className="absolute top-0 right-0 w-96 h-96 bg-cyan-500/10 rounded-full blur-3xl pointer-events-none" />

        <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 shadow-[0_0_10px_#10b981] animate-pulse" />
              <span className="text-xs font-mono text-cyan-400 font-semibold tracking-wider uppercase">
                Zero-Trust Control Plane Active
              </span>
            </div>
            <h1 className="text-2xl md:text-3xl font-extrabold text-white tracking-tight">
              Tarak Orchestration Cluster
            </h1>
            <p className="text-sm text-slate-300 max-w-xl">
              High-performance container orchestration engine running native TCR isolation, 
              inbuilt MetalLB load balancing, and universal mTLS service mesh.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <Link href="/dashboard/devtools/terminal">
              <Button variant="outline" size="md">
                <Terminal size={16} />
                <span>Web Terminal</span>
              </Button>
            </Link>
            <Link href="/dashboard/devtools/manifests">
              <Button size="md">
                <Zap size={16} />
                <span>Deploy YAML</span>
              </Button>
            </Link>
          </div>
        </div>
      </div>

      {/* KPI Metrics Cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <Card interactive className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Nodes Online</span>
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <Cpu size={18} />
            </div>
          </div>
          <div className="text-2xl font-bold text-white tracking-tight">
            {clusterInfo.nodesCount}
          </div>
          <div className="flex items-center gap-1.5 text-[11px] text-emerald-400">
            <CheckCircle2 size={12} />
            <span>Ready & Schedulable</span>
          </div>
        </Card>

        <Card interactive className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Active Pods</span>
            <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
              <Box size={18} />
            </div>
          </div>
          <div className="text-2xl font-bold text-white tracking-tight">
            {clusterInfo.podsCount}
          </div>
          <div className="text-[11px] text-slate-400">
            in namespace <span className="text-cyan-400 font-mono font-semibold">{selectedNamespace}</span>
          </div>
        </Card>

        <Card interactive className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Services & LB</span>
            <div className="p-2 rounded-lg bg-indigo-500/10 text-indigo-400">
              <Server size={18} />
            </div>
          </div>
          <div className="text-2xl font-bold text-white tracking-tight">
            {clusterInfo.servicesCount}
          </div>
          <div className="text-[11px] text-indigo-300">
            MetalLB: <span className="font-mono">192.168.1.240</span>
          </div>
        </Card>

        <Card interactive className="space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">mTLS Multi-Mesh</span>
            <div className="p-2 rounded-lg bg-purple-500/10 text-purple-400">
              <Radio size={18} />
            </div>
          </div>
          <div className="text-2xl font-bold text-white tracking-tight">Active</div>
          <div className="text-[11px] text-purple-300">Strict Zero-Trust Mode</div>
        </Card>
      </div>

      {/* ArgoCD Style Live Topology Graph */}
      <Card className="space-y-4">
        <CardHeader>
          <CardTitle>
            <Network size={18} className="text-cyan-400" />
            <span>Live Workload Topology (ArgoCD Sync)</span>
          </CardTitle>
          <div className="flex items-center gap-2">
            <Badge variant="emerald" dot>Synced</Badge>
            <Badge variant="cyan">GitOps</Badge>
          </div>
        </CardHeader>

        <div className="p-6 rounded-xl bg-slate-950/60 border border-white/5 flex flex-col md:flex-row items-center justify-between gap-6 overflow-x-auto">
          {/* Node 1: Ingress */}
          <div className="flex flex-col items-center p-4 rounded-xl bg-slate-900/80 border border-cyan-500/30 w-44 text-center shadow-lg">
            <Globe size={24} className="text-cyan-400 mb-2" />
            <span className="text-xs font-bold text-white">Ingress Controller</span>
            <span className="text-[10px] text-slate-400 font-mono mt-1">*.local / public</span>
          </div>

          <div className="text-slate-600 font-bold hidden md:block">──►</div>

          {/* Node 2: Service / LB */}
          <div className="flex flex-col items-center p-4 rounded-xl bg-slate-900/80 border border-indigo-500/30 w-44 text-center shadow-lg">
            <Server size={24} className="text-indigo-400 mb-2" />
            <span className="text-xs font-bold text-white">LoadBalancer Service</span>
            <span className="text-[10px] text-slate-400 font-mono mt-1">MetalLB auto-ip</span>
          </div>

          <div className="text-slate-600 font-bold hidden md:block">──►</div>

          {/* Node 3: Mesh Proxy */}
          <div className="flex flex-col items-center p-4 rounded-xl bg-slate-900/80 border border-purple-500/30 w-44 text-center shadow-lg">
            <Radio size={24} className="text-purple-400 mb-2" />
            <span className="text-xs font-bold text-white">mTLS Sidecar Mesh</span>
            <span className="text-[10px] text-slate-400 font-mono mt-1">Zero-Trust Policy</span>
          </div>

          <div className="text-slate-600 font-bold hidden md:block">──►</div>

          {/* Node 4: Running Pods */}
          <div className="flex flex-col items-center p-4 rounded-xl bg-slate-900/80 border border-emerald-500/30 w-44 text-center shadow-lg">
            <Box size={24} className="text-emerald-400 mb-2" />
            <span className="text-xs font-bold text-white">Application Pods</span>
            <span className="text-[10px] text-emerald-400 font-mono mt-1">{pods.length} Running</span>
          </div>
        </div>
      </Card>

      {/* Quick Launchers */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Link href="/dashboard/networking/tunnels" className="group">
          <Card interactive className="p-5 h-full flex items-start gap-4">
            <div className="p-3 rounded-xl bg-cyan-500/10 text-cyan-400 group-hover:scale-110 transition-transform">
              <Cloud size={22} />
            </div>
            <div className="space-y-1">
              <h4 className="text-sm font-bold text-white flex items-center gap-1.5">
                Cloudflare Tunnels
                <ArrowUpRight size={14} className="opacity-60" />
              </h4>
              <p className="text-xs text-slate-400">
                Expose local workloads to public HTTPS domain without open ports.
              </p>
            </div>
          </Card>
        </Link>

        <Link href="/dashboard/mesh/overview" className="group">
          <Card interactive className="p-5 h-full flex items-start gap-4">
            <div className="p-3 rounded-xl bg-purple-500/10 text-purple-400 group-hover:scale-110 transition-transform">
              <Radio size={22} />
            </div>
            <div className="space-y-1">
              <h4 className="text-sm font-bold text-white flex items-center gap-1.5">
                Universal Multi-Mesh
                <ArrowUpRight size={14} className="opacity-60" />
              </h4>
              <p className="text-xs text-slate-400">
                Manage isolated service meshes, mTLS permissions, and traffic split.
              </p>
            </div>
          </Card>
        </Link>

        <Link href="/dashboard/observability/hubble" className="group">
          <Card interactive className="p-5 h-full flex items-start gap-4">
            <div className="p-3 rounded-xl bg-emerald-500/10 text-emerald-400 group-hover:scale-110 transition-transform">
              <Activity size={22} />
            </div>
            <div className="space-y-1">
              <h4 className="text-sm font-bold text-white flex items-center gap-1.5">
                Hubble Network Visualizer
                <ArrowUpRight size={14} className="opacity-60" />
              </h4>
              <p className="text-xs text-slate-400">
                Live stream TCP/UDP flows, latency telemetry, and DNS queries.
              </p>
            </div>
          </Card>
        </Link>
      </div>
    </div>
  );
}
