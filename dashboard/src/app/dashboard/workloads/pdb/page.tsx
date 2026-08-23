"use client";

import React, { useState, useEffect } from "react";
import {
  ShieldAlert,
  ShieldCheck,
  Plus,
  RefreshCw,
  Search,
  CheckCircle2,
  AlertTriangle,
  Layers,
  Box,
  Trash2,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { tarakFetch } from "@/lib/api";

interface PDBItem {
  name: string;
  namespace: string;
  minAvailable: string;
  maxUnavailable: string;
  allowedDisruptions: number;
  currentHealthy: number;
  desiredHealthy: number;
  totalPods: number;
  selector: string;
}

export default function PDBPage() {
  const [pdbs, setPdbs] = useState<PDBItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");

  const fetchPDBs = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/policy/v1/poddisruptionbudgets");
      const items = res.data?.items || [];
      const mapped: PDBItem[] = items.map((raw: any) => {
        const spec = raw.spec || {};
        const status = raw.status || {};
        let sel = "*";
        if (spec.selector?.matchLabels) {
          sel = Object.entries(spec.selector.matchLabels)
            .map(([k, v]) => `${k}=${v}`)
            .join(",");
        }
        return {
          name: raw.metadata?.name || "pdb",
          namespace: raw.metadata?.namespace || "default",
          minAvailable: spec.minAvailable !== undefined ? String(spec.minAvailable) : "N/A",
          maxUnavailable: spec.maxUnavailable !== undefined ? String(spec.maxUnavailable) : "N/A",
          allowedDisruptions: status.disruptionsAllowed ?? 1,
          currentHealthy: status.currentHealthy ?? 0,
          desiredHealthy: status.desiredHealthy ?? 0,
          totalPods: status.expectedPods ?? 0,
          selector: sel,
        };
      });
      setPdbs(mapped);
    } catch {
      setPdbs([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchPDBs();
  }, []);

  const filtered = pdbs.filter(
    (p) =>
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.namespace.toLowerCase().includes(search.toLowerCase()) ||
      p.selector.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-400">
              <ShieldAlert size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Pod Disruption Budgets <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 via-indigo-300 to-cyan-400">(PDB)</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Ensure high availability during voluntary node drains, kernel upgrades, and cluster maintenance.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={fetchPDBs}>
            <RefreshCw size={14} className={`mr-1.5 ${isLoading ? "animate-spin" : ""}`} /> Refresh
          </Button>
          <Button size="sm" className="bg-gradient-to-r from-purple-600 to-cyan-600 text-white shadow-lg shadow-purple-950/40">
            <Plus size={14} className="mr-1.5" /> Create PDB
          </Button>
        </div>
      </div>

      {/* Filter / Search Bar */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Filter PDBs by name, namespace, or selector..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-2 rounded-xl bg-slate-900/80 border border-white/10 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-purple-500/50"
          />
        </div>
      </div>

      {/* Grid */}
      {filtered.length === 0 ? (
        <div className="p-12 rounded-2xl bg-slate-900/40 border border-white/10 text-center space-y-3">
          <ShieldCheck size={36} className="text-slate-600 mx-auto" />
          <h3 className="text-sm font-bold text-white">No PodDisruptionBudgets Configured</h3>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Define minimum available or maximum unavailable constraints to prevent simultaneous pod disruptions during cluster operations.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filtered.map((pdb) => (
            <div
              key={`${pdb.namespace}/${pdb.name}`}
              className="p-5 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-4 hover:border-white/20 transition-all"
            >
              <div className="flex items-start justify-between gap-2">
                <div className="space-y-1">
                  <span className="font-bold text-white text-sm font-mono">{pdb.name}</span>
                  <div className="text-xs text-slate-400 font-mono">
                    Namespace: <span className="text-cyan-300">{pdb.namespace}</span> • Selector: <span className="text-purple-300">{pdb.selector}</span>
                  </div>
                </div>
                <Badge variant={pdb.allowedDisruptions > 0 ? "emerald" : "rose"}>
                  {pdb.allowedDisruptions} Allowed Disruptions
                </Badge>
              </div>

              <div className="grid grid-cols-3 gap-2 font-mono text-xs text-center">
                <div className="p-3 rounded-xl bg-[#04060c] border border-white/10">
                  <span className="text-[10px] text-slate-400 block font-sans uppercase">Min Available</span>
                  <span className="text-cyan-300 font-bold">{pdb.minAvailable}</span>
                </div>
                <div className="p-3 rounded-xl bg-[#04060c] border border-white/10">
                  <span className="text-[10px] text-slate-400 block font-sans uppercase">Max Unavailable</span>
                  <span className="text-purple-300 font-bold">{pdb.maxUnavailable}</span>
                </div>
                <div className="p-3 rounded-xl bg-[#04060c] border border-white/10">
                  <span className="text-[10px] text-slate-400 block font-sans uppercase">Healthy Pods</span>
                  <span className="text-emerald-300 font-bold">{pdb.currentHealthy} / {pdb.totalPods}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
