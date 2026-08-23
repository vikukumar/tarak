"use client";

import React, { useState } from "react";
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

const initialHPA: HPAItem[] = [
  {
    name: "storefront-autoscale",
    namespace: "production",
    targetRef: "Deployment/storefront-web",
    minReplicas: 2,
    maxReplicas: 10,
    currentReplicas: 3,
    cpuTarget: 75,
    cpuCurrent: 42,
    memoryTarget: 80,
    memoryCurrent: 61,
    lastScale: "22 mins ago (2 -> 3)",
  },
  {
    name: "payments-api-autoscale",
    namespace: "finance",
    targetRef: "Deployment/payments-api",
    minReplicas: 3,
    maxReplicas: 15,
    currentReplicas: 5,
    cpuTarget: 60,
    cpuCurrent: 68,
    memoryTarget: 70,
    memoryCurrent: 55,
    lastScale: "5 mins ago (4 -> 5)",
  },
];

export default function HPAPage() {
  const [hpas, setHpas] = useState<HPAItem[]>(initialHPA);
  const [search, setSearch] = useState("");

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
          <Button size="sm" className="bg-gradient-to-r from-emerald-600 to-cyan-600 text-white shadow-lg shadow-emerald-950/40">
            <Plus size={14} className="mr-1.5" /> Create HPA Rule
          </Button>
        </div>
      </div>

      {/* HPA Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {filtered.map((hpa) => (
          <div
            key={hpa.name}
            className="p-5 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-4 hover:border-white/20 transition-all"
          >
            <div className="flex items-start justify-between gap-2">
              <div className="space-y-1">
                <span className="font-bold text-white text-sm font-mono">{hpa.name}</span>
                <div className="text-xs text-slate-400 font-mono">
                  Target: <span className="text-cyan-300">{hpa.targetRef}</span> ({hpa.namespace})
                </div>
              </div>
              <Badge variant="emerald" dot>
                {hpa.currentReplicas} / {hpa.maxReplicas} Replicas
              </Badge>
            </div>

            {/* Metrics Gauges */}
            <div className="grid grid-cols-2 gap-3 pt-2">
              <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 space-y-1.5 font-mono text-xs">
                <div className="flex items-center justify-between text-slate-400">
                  <span className="flex items-center gap-1"><Cpu size={12} className="text-cyan-400" /> CPU Usage</span>
                  <span className="text-cyan-300 font-bold">{hpa.cpuCurrent}%</span>
                </div>
                <div className="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden">
                  <div
                    className={`h-full ${hpa.cpuCurrent > hpa.cpuTarget ? "bg-amber-400" : "bg-cyan-400"}`}
                    style={{ width: `${Math.min(100, hpa.cpuCurrent)}%` }}
                  />
                </div>
                <div className="text-[10px] text-slate-400">Target Threshold: {hpa.cpuTarget}%</div>
              </div>

              <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 space-y-1.5 font-mono text-xs">
                <div className="flex items-center justify-between text-slate-400">
                  <span className="flex items-center gap-1"><Database size={12} className="text-purple-400" /> Memory Usage</span>
                  <span className="text-purple-300 font-bold">{hpa.memoryCurrent}%</span>
                </div>
                <div className="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-purple-400"
                    style={{ width: `${Math.min(100, hpa.memoryCurrent)}%` }}
                  />
                </div>
                <div className="text-[10px] text-slate-400">Target Threshold: {hpa.memoryTarget}%</div>
              </div>
            </div>

            <div className="flex items-center justify-between pt-2 border-t border-white/5 text-xs text-slate-400 font-mono">
              <span>Bounds: {hpa.minReplicas} min — {hpa.maxReplicas} max</span>
              <span className="text-emerald-400">{hpa.lastScale}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
