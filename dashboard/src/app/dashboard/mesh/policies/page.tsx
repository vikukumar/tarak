"use client";

import React, { useState, useEffect } from "react";
import {
  Radio,
  Sliders,
  Plus,
  Search,
  RefreshCw,
  Lock,
  Workflow,
  Zap,
  Activity,
  Shield,
  Clock,
  CheckCircle2,
  Trash2,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

interface MeshPolicy {
  id: string;
  name: string;
  type: "RateLimit" | "CircuitBreaker" | "FaultInjection" | "HealthCheck" | "mTLS";
  mesh: string;
  targetRef: string;
  rules: string;
  enabled: boolean;
}

import { tarakFetch } from "@/lib/api";

export default function MeshPoliciesPage() {
  const [policies, setPolicies] = useState<MeshPolicy[]>([]);
  const [search, setSearch] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [selectedType, setSelectedType] = useState<string>("All");

  const fetchPolicies = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/mesh.tarak.io/v1/proxypatches");
      const items = res.data?.items || [];
      setPolicies(items);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchPolicies();
  }, []);

  const togglePolicy = (id: string) => {
    setPolicies((prev) =>
      prev.map((p) => (p.id === id ? { ...p, enabled: !p.enabled } : p))
    );
  };

  const deletePolicy = (id: string) => {
    setPolicies((prev) => prev.filter((p) => p.id !== id));
  };

  const filtered = policies.filter(
    (p) =>
      (selectedType === "All" || p.type === selectedType) &&
      (p.name.toLowerCase().includes(search.toLowerCase()) ||
        p.targetRef?.toLowerCase().includes(search.toLowerCase()))
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Sliders size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Service Mesh Policies <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-indigo-300 to-purple-400">(Kuma / Kong Mesh)</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Fine-grained L4/L7 traffic policies including Rate Limiting, Circuit Breakers, Fault Injection, and Active Health Checks.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={fetchPolicies}>
            <RefreshCw size={14} className={`mr-1.5 ${isLoading ? "animate-spin" : ""}`} /> Refresh
          </Button>
          <Button size="sm" className="bg-gradient-to-r from-cyan-600 to-purple-600 text-white shadow-lg shadow-cyan-950/40">
            <Plus size={14} className="mr-1.5" /> Create Mesh Policy
          </Button>
        </div>
      </div>

      {/* Type Filter Buttons */}
      <div className="flex flex-wrap items-center gap-2">
        {["All", "RateLimit", "CircuitBreaker", "FaultInjection", "HealthCheck"].map((type) => (
          <button
            key={type}
            onClick={() => setSelectedType(type)}
            className={`px-3 py-1.5 rounded-xl text-xs font-bold font-mono transition-all border ${
              selectedType === type
                ? "bg-cyan-500/20 text-cyan-300 border-cyan-500/40 shadow-lg shadow-cyan-950/40"
                : "bg-slate-900/60 text-slate-400 border-white/10 hover:text-white"
            }`}
          >
            {type}
          </button>
        ))}
      </div>

      {/* Policies Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {filtered.map((pol) => (
          <div
            key={pol.id}
            className="p-5 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-4 hover:border-white/20 transition-all"
          >
            <div className="flex items-start justify-between gap-2">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-white text-sm font-mono">{pol.name}</span>
                  <Badge variant={pol.type === "RateLimit" ? "cyan" : pol.type === "CircuitBreaker" ? "rose" : "purple"}>
                    {pol.type}
                  </Badge>
                </div>
                <div className="text-xs text-slate-400 font-mono">
                  Mesh: <span className="text-purple-300">{pol.mesh}</span> • Target: <span className="text-cyan-300">{pol.targetRef}</span>
                </div>
              </div>

              <button
                onClick={() => togglePolicy(pol.id)}
                className={`px-2.5 py-1 rounded-lg text-xs font-bold font-mono border transition-all ${
                  pol.enabled
                    ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/30"
                    : "bg-slate-800 text-slate-500 border-white/10"
                }`}
              >
                {pol.enabled ? "Active" : "Disabled"}
              </button>
            </div>

            <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 text-xs font-mono text-slate-300">
              {pol.rules}
            </div>

            <div className="flex items-center justify-between text-xs pt-2 border-t border-white/5">
              <span className="text-slate-400 flex items-center gap-1">
                <Lock size={12} className="text-cyan-400" /> mTLS Enforced
              </span>
              <Button size="sm" variant="ghost" onClick={() => deletePolicy(pol.id)} className="text-rose-400 hover:text-rose-300 text-xs h-7 px-2">
                <Trash2 size={13} className="mr-1" /> Delete
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
