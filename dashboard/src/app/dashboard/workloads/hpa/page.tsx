"use client";

import React, { useState, useEffect } from "react";
import {
  TrendingUp,
  Cpu,
  Database,
  Plus,
  RefreshCw,
  Search,
  Activity,
  Layers,
  ArrowUpRight,
  Shield,
  Zap,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { tarakFetch } from "@/lib/api";

interface HPAItem {
  name: string;
  namespace: string;
  targetRef: string;
  minReplicas: number;
  maxReplicas: number;
  currentReplicas: number;
  cpuTarget: number;
  cpuCurrent: number;
  memoryTarget: number;
  memoryCurrent: number;
  lastScale: string;
}

export default function HPAPage() {
  const [hpas, setHpas] = useState<HPAItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");

  const fetchHPAs = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/autoscaling/v2/horizontalpodautoscalers");
      const items = res.data?.items || [];
      const mapped: HPAItem[] = items.map((raw: any) => {
        const spec = raw.spec || {};
        const status = raw.status || {};
        const targetRef = spec.scaleTargetRef
          ? `${spec.scaleTargetRef.kind || "Deployment"}/${spec.scaleTargetRef.name || ""}`
          : "Workload";
        return {
          name: raw.metadata?.name || "hpa",
          namespace: raw.metadata?.namespace || "default",
          targetRef,
          minReplicas: spec.minReplicas || 1,
          maxReplicas: spec.maxReplicas || 10,
          currentReplicas: status.currentReplicas || spec.minReplicas || 1,
          cpuTarget: 80,
          cpuCurrent: status.currentCPUUtilizationPercentage || 0,
          memoryTarget: 80,
          memoryCurrent: 0,
          lastScale: status.lastScaleTime ? new Date(status.lastScaleTime).toLocaleTimeString() : "Never",
        };
      });
      setHpas(mapped);
    } catch {
      setHpas([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchHPAs();
  }, []);

  const filtered = hpas.filter(
    (h) =>
      h.name.toLowerCase().includes(search.toLowerCase()) ||
      h.namespace.toLowerCase().includes(search.toLowerCase()) ||
      h.targetRef.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
              <TrendingUp size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Horizontal Pod Autoscaling <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 via-cyan-300 to-purple-400">(HPA Metrics)</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Dynamic replica scaling based on CPU, memory, and custom hardware metrics thresholds.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={fetchHPAs}>
            <RefreshCw size={14} className={`mr-1.5 ${isLoading ? "animate-spin" : ""}`} /> Refresh
          </Button>
          <Button size="sm" className="bg-gradient-to-r from-emerald-600 to-cyan-600 text-white shadow-lg shadow-emerald-950/40">
            <Plus size={14} className="mr-1.5" /> Create HPA Rule
          </Button>
        </div>
      </div>

      {/* Filter / Search Bar */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Filter HPAs by name, namespace, or target workload..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-2 rounded-xl bg-slate-900/80 border border-white/10 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-emerald-500/50"
          />
        </div>
      </div>

      {/* HPA Content */}
      {filtered.length === 0 ? (
        <div className="p-12 rounded-2xl bg-slate-900/40 border border-white/10 text-center space-y-3">
          <TrendingUp size={36} className="text-slate-600 mx-auto" />
          <h3 className="text-sm font-bold text-white">No HorizontalPodAutoscalers Configured</h3>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Scale your Deployments or StatefulSets automatically based on CPU and memory consumption. Create an HPA rule to get started.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filtered.map((h) => (
            <div
              key={`${h.namespace}/${h.name}`}
              className="p-5 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-4 hover:border-white/20 transition-all"
            >
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-white text-base">{h.name}</span>
                    <Badge variant="purple">{h.namespace}</Badge>
                  </div>
                  <p className="text-xs text-slate-400 font-mono mt-0.5">Target: {h.targetRef}</p>
                </div>
                <Badge variant="emerald">
                  {h.currentReplicas} / {h.maxReplicas} Replicas
                </Badge>
              </div>

              {/* Gauges */}
              <div className="grid grid-cols-2 gap-3 pt-2">
                <div className="p-3 rounded-xl bg-slate-950/60 border border-white/5 space-y-1">
                  <div className="flex items-center justify-between text-[11px] text-slate-400">
                    <span className="flex items-center gap-1 font-mono">
                      <Cpu size={12} className="text-emerald-400" /> CPU Utilization
                    </span>
                    <span className="font-bold text-white">{h.cpuCurrent}% / {h.cpuTarget}%</span>
                  </div>
                  <div className="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all ${
                        h.cpuCurrent >= h.cpuTarget ? "bg-amber-400" : "bg-emerald-400"
                      }`}
                      style={{ width: `${Math.min(100, (h.cpuCurrent / h.cpuTarget) * 100)}%` }}
                    />
                  </div>
                </div>

                <div className="p-3 rounded-xl bg-slate-950/60 border border-white/5 space-y-1">
                  <div className="flex items-center justify-between text-[11px] text-slate-400">
                    <span className="flex items-center gap-1 font-mono">
                      <Database size={12} className="text-cyan-400" /> Memory Utilization
                    </span>
                    <span className="font-bold text-white">{h.memoryCurrent}% / {h.memoryTarget}%</span>
                  </div>
                  <div className="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
                    <div
                      className="h-full bg-cyan-400 rounded-full transition-all"
                      style={{ width: `${Math.min(100, (h.memoryCurrent / h.memoryTarget) * 100)}%` }}
                    />
                  </div>
                </div>
              </div>

              <div className="flex items-center justify-between text-[11px] text-slate-400 pt-2 border-t border-white/5 font-mono">
                <span>Min: {h.minReplicas} | Max: {h.maxReplicas}</span>
                <span>Last Scale: {h.lastScale}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
