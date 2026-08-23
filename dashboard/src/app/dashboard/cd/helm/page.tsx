"use client";

import React, { useState, useEffect } from "react";
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
import { tarakFetch } from "@/lib/api";

interface HelmRelease {
  name: string;
  namespace: string;
  chart: string;
  appVersion: string;
  revision: number;
  status: "deployed" | "failed" | "pending-upgrade";
  updated: string;
}

export default function HelmReleasesPage() {
  const [releases, setReleases] = useState<HelmRelease[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [selectedRelease, setSelectedRelease] = useState<HelmRelease | null>(null);

  const fetchReleases = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/helm.tarak.io/v1/releases");
      const items = res.data?.items || [];
      const mapped: HelmRelease[] = items.map((raw: any) => {
        const spec = raw.spec || {};
        const status = raw.status || {};
        return {
          name: raw.metadata?.name || "helm-release",
          namespace: raw.metadata?.namespace || "default",
          chart: spec.chart || "custom-chart",
          appVersion: spec.appVersion || "v1.0.0",
          revision: status.version || 1,
          status: (status.status as "deployed" | "failed" | "pending-upgrade") || "deployed",
          updated: status.lastUpdated ? new Date(status.lastUpdated).toLocaleDateString() : "Recently",
        };
      });
      setReleases(mapped);
      if (mapped.length > 0 && !selectedRelease) {
        setSelectedRelease(mapped[0]);
      }
    } catch {
      setReleases([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchReleases();
  }, []);

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
          <Button variant="outline" size="sm" onClick={fetchReleases}>
            <RefreshCw size={14} className={`mr-1.5 ${isLoading ? "animate-spin" : ""}`} /> Refresh
          </Button>
          <Button size="sm" className="bg-gradient-to-r from-cyan-600 to-purple-600 text-white shadow-lg shadow-cyan-950/40">
            <Plus size={14} className="mr-1.5" /> Install Helm Chart
          </Button>
        </div>
      </div>

      {/* Search */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Filter releases by name, namespace, or chart..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-2 rounded-xl bg-slate-900/80 border border-white/10 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500/50"
          />
        </div>
      </div>

      {/* Grid */}
      {filtered.length === 0 ? (
        <div className="p-12 rounded-2xl bg-slate-900/40 border border-white/10 text-center space-y-3">
          <Package size={36} className="text-slate-600 mx-auto" />
          <h3 className="text-sm font-bold text-white">No Helm Releases Found</h3>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Install applications from Helm OCI or HTTPS repositories directly into your Tarak cluster.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {filtered.map((rel) => (
            <div
              key={`${rel.namespace}/${rel.name}`}
              onClick={() => setSelectedRelease(rel)}
              className={`p-5 rounded-2xl bg-slate-900/70 border transition-all cursor-pointer space-y-3 ${
                selectedRelease?.name === rel.name && selectedRelease?.namespace === rel.namespace
                  ? "border-cyan-500/50 shadow-cyan-950/30 shadow-lg"
                  : "border-white/10 hover:border-white/20"
              }`}
            >
              <div className="flex items-start justify-between gap-2">
                <div>
                  <span className="font-bold text-white text-sm font-mono">{rel.name}</span>
                  <p className="text-xs text-slate-400 font-mono mt-0.5">ns: {rel.namespace}</p>
                </div>
                <Badge variant={rel.status === "deployed" ? "emerald" : "rose"}>
                  {rel.status}
                </Badge>
              </div>

              <div className="p-3 rounded-xl bg-slate-950/60 border border-white/5 space-y-1 font-mono text-xs">
                <div className="flex justify-between text-slate-400">
                  <span>Chart:</span>
                  <span className="text-cyan-300 font-bold truncate max-w-[130px]">{rel.chart}</span>
                </div>
                <div className="flex justify-between text-slate-400">
                  <span>App Version:</span>
                  <span className="text-white">{rel.appVersion}</span>
                </div>
                <div className="flex justify-between text-slate-400">
                  <span>Revision:</span>
                  <span className="text-purple-300">rev #{rel.revision}</span>
                </div>
                <div className="flex justify-between text-slate-400">
                  <span>Updated:</span>
                  <span className="text-slate-300">{rel.updated}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
