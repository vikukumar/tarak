"use client";

import React, { useState } from "react";
import {
  Layers,
  Package,
  Plus,
  RefreshCw,
  Search,
  CheckCircle2,
  AlertTriangle,
  RotateCcw,
  FileCode,
  Globe,
  Database,
  Shield,
  Zap,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

interface HelmRelease {
  name: string;
  namespace: string;
  chart: string;
  appVersion: string;
  revision: number;
  status: "deployed" | "failed" | "pending-upgrade";
  updated: string;
}

const initialReleases: HelmRelease[] = [
  {
    name: "cert-manager",
    namespace: "cert-manager",
    chart: "cert-manager-v1.14.4",
    appVersion: "v1.14.4",
    revision: 2,
    status: "deployed",
    updated: "2 days ago",
  },
  {
    name: "prometheus-stack",
    namespace: "monitoring",
    chart: "kube-prometheus-stack-58.2.0",
    appVersion: "v0.73.1",
    revision: 4,
    status: "deployed",
    updated: "5 hours ago",
  },
  {
    name: "redis-ha-cluster",
    namespace: "cache",
    chart: "redis-ha-4.24.1",
    appVersion: "7.2.4",
    revision: 1,
    status: "deployed",
    updated: "1 day ago",
  },
];

export default function HelmReleasesPage() {
  const [releases, setReleases] = useState<HelmRelease[]>(initialReleases);
  const [search, setSearch] = useState("");
  const [selectedRelease, setSelectedRelease] = useState<HelmRelease | null>(releases[0]);

  const filtered = releases.filter(
    (r) =>
      r.name.toLowerCase().includes(search.toLowerCase()) ||
      r.namespace.toLowerCase().includes(search.toLowerCase()) ||
      r.chart.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Package size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Helm Package Releases <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-indigo-300 to-purple-400">& Chart Catalog</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Native Helm 3 chart lifecycle management, values override editor, and revision rollbacks.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button size="sm" className="bg-gradient-to-r from-cyan-600 to-purple-600 text-white shadow-lg shadow-cyan-950/40">
            <Plus size={14} className="mr-1.5" /> Install Helm Chart
          </Button>
        </div>
      </div>

      {/* Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {filtered.map((rel) => (
          <div
            key={rel.name}
            className="p-5 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-4 hover:border-white/20 transition-all cursor-pointer"
            onClick={() => setSelectedRelease(rel)}
          >
            <div className="flex items-start justify-between gap-2">
              <div className="space-y-1">
                <span className="font-bold text-white text-sm font-mono">{rel.name}</span>
                <div className="text-xs text-slate-400 font-mono">
                  Namespace: <span className="text-cyan-300">{rel.namespace}</span>
                </div>
              </div>
              <Badge variant={rel.status === "deployed" ? "emerald" : "rose"}>
                {rel.status}
              </Badge>
            </div>

            <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 font-mono text-xs space-y-1 text-slate-300">
              <div>Chart: <span className="text-purple-300">{rel.chart}</span></div>
              <div>App Version: <span className="text-cyan-300">{rel.appVersion}</span></div>
            </div>

            <div className="flex items-center justify-between text-xs text-slate-400 font-mono pt-2 border-t border-white/5">
              <span>Revision #{rel.revision}</span>
              <span className="text-slate-400">{rel.updated}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
