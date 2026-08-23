"use client";

import React, { useState, useEffect } from "react";
import {
  Activity,
  Radio,
  RefreshCw,
  Play,
  Pause,
  Filter,
  Search,
  ArrowRight,
  Shield,
  Zap,
  Globe,
  Database,
  Server,
  Box,
  Layers,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { tarakFetch } from "@/lib/api";

interface NetworkFlow {
  id: string;
  time: string;
  isoTime: string;
  source: string;
  dest: string;
  sourceNamespace: string;
  destNamespace: string;
  protocol: string;
  port: number;
  verdict: "FORWARDED" | "DROPPED" | "AUDIT";
  latency: string;
  bytes: string;
  summary?: string;
}

export default function HubblePage() {
  const [flows, setFlows] = useState<NetworkFlow[]>([]);
  const [isLive, setIsLive] = useState(true);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [selectedVerdict, setSelectedVerdict] = useState<string>("All");
  const [selectedFlow, setSelectedFlow] = useState<NetworkFlow | null>(null);

  const fetchFlows = async () => {
    try {
      const res = await tarakFetch("/apis/telemetry.tarak.io/v1/flows");
      const items = res.data?.items || [];
      const mapped: NetworkFlow[] = items.map((f: any) => {
        const d = f.timestamp ? new Date(f.timestamp) : new Date();
        const pad = (n: number) => n.toString().padStart(2, "0");
        const ms = d.getMilliseconds().toString().padStart(3, "0");
        const timeStr = `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}.${ms}`;
        const b = f.bytes || 0;
        const bStr = b > 1024 ? `${(b / 1024).toFixed(1)} KB` : `${b} B`;

        return {
          id: f.id || `flow-${Date.now()}`,
          time: timeStr,
          isoTime: d.toISOString(),
          source: f.srcPod || f.srcIp || "ingress",
          dest: f.dstPod || f.dstIp || "cluster",
          sourceNamespace: f.srcNs || "default",
          destNamespace: f.dstNs || "default",
          protocol: f.protocol || "TCP",
          port: f.dstPort || 80,
          verdict: (f.verdict as "FORWARDED" | "DROPPED" | "AUDIT") || "FORWARDED",
          latency: f.latencyMs ? `${f.latencyMs.toFixed(2)}ms` : "0.10ms",
          bytes: bStr,
          summary: f.summary || "",
        };
      });
      setFlows(mapped);
      if (mapped.length > 0 && !selectedFlow) {
        setSelectedFlow(mapped[0]);
      }
    } catch {
      setFlows([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchFlows();
  }, []);

  useEffect(() => {
    if (!isLive) return;
    const interval = setInterval(fetchFlows, 2500);
    return () => clearInterval(interval);
  }, [isLive]);

  const filteredFlows = flows.filter(
    (f) =>
      (selectedVerdict === "All" || f.verdict === selectedVerdict) &&
      (f.source.toLowerCase().includes(search.toLowerCase()) ||
        f.dest.toLowerCase().includes(search.toLowerCase()) ||
        f.protocol.toLowerCase().includes(search.toLowerCase()) ||
        f.sourceNamespace.toLowerCase().includes(search.toLowerCase()) ||
        f.destNamespace.toLowerCase().includes(search.toLowerCase()))
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-400">
              <Radio size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Hubble Network Flows <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 via-indigo-300 to-cyan-400">& Telemetry</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Real-time eBPF and CNI packet stream, L3/L4/L7 flow monitoring, and zero-trust policy enforcement verdicts.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            size="sm"
            variant="outline"
            onClick={() => setIsLive(!isLive)}
            className={`border ${isLive ? "border-emerald-500/30 text-emerald-400 bg-emerald-500/10" : "border-slate-700 text-slate-400"}`}
          >
            {isLive ? <Pause size={14} className="mr-1.5" /> : <Play size={14} className="mr-1.5" />}
            {isLive ? "Live Stream Active" : "Stream Paused"}
          </Button>

          <Button size="sm" variant="outline" onClick={fetchFlows}>
            <RefreshCw size={14} className={`mr-1.5 ${isLoading ? "animate-spin" : ""}`} /> Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Filter by pod, IP, namespace, or protocol..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-2 rounded-xl bg-slate-900/80 border border-white/10 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-purple-500/50"
          />
        </div>

        <div className="flex items-center gap-1.5 bg-slate-900/80 p-1 rounded-xl border border-white/10 text-xs font-mono">
          {["All", "FORWARDED", "DROPPED", "AUDIT"].map((v) => (
            <button
              key={v}
              onClick={() => setSelectedVerdict(v)}
              className={`px-3 py-1 rounded-lg transition-all ${
                selectedVerdict === v
                  ? "bg-purple-600 text-white font-bold shadow-md shadow-purple-900/30"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              {v}
            </button>
          ))}
        </div>
      </div>

      {/* Flows Table */}
      {filteredFlows.length === 0 ? (
        <div className="p-12 rounded-2xl bg-slate-900/40 border border-white/10 text-center space-y-3 font-mono">
          <Radio size={36} className="text-slate-600 mx-auto" />
          <h3 className="text-sm font-bold text-white">No Live Network Flows Recorded</h3>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Network flows captured by the native CNI driver and proxy will appear here dynamically in real time.
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-white/10 bg-slate-900/70 shadow-xl">
          <table className="w-full text-left text-xs border-collapse font-mono">
            <thead>
              <tr className="bg-slate-950/90 text-slate-300 uppercase tracking-wider font-bold border-b border-white/10 font-sans">
                <th className="p-3.5">Timestamp</th>
                <th className="p-3.5">Source</th>
                <th className="p-3.5">Destination</th>
                <th className="p-3.5">Protocol / Port</th>
                <th className="p-3.5">Verdict</th>
                <th className="p-3.5">Latency</th>
                <th className="p-3.5">Bytes</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {filteredFlows.map((flow) => (
                <tr
                  key={flow.id}
                  onClick={() => setSelectedFlow(flow)}
                  className={`hover:bg-white/[0.03] transition-colors cursor-pointer ${
                    selectedFlow?.id === flow.id ? "bg-purple-500/10" : ""
                  }`}
                >
                  <td className="p-3.5 text-slate-400 whitespace-nowrap">{flow.time}</td>
                  <td className="p-3.5">
                    <span className="text-purple-300 font-bold block">{flow.source}</span>
                    <span className="text-[10px] text-slate-500">ns: {flow.sourceNamespace}</span>
                  </td>
                  <td className="p-3.5">
                    <span className="text-cyan-300 font-bold block">{flow.dest}</span>
                    <span className="text-[10px] text-slate-500">ns: {flow.destNamespace}</span>
                  </td>
                  <td className="p-3.5">
                    <span className="text-white font-bold block">{flow.protocol}</span>
                    <span className="text-[10px] text-slate-400">Port: {flow.port}</span>
                  </td>
                  <td className="p-3.5 whitespace-nowrap">
                    <Badge
                      variant={
                        flow.verdict === "FORWARDED"
                          ? "emerald"
                          : flow.verdict === "DROPPED"
                          ? "rose"
                          : "amber"
                      }
                      dot
                    >
                      {flow.verdict}
                    </Badge>
                  </td>
                  <td className="p-3.5 text-emerald-400 whitespace-nowrap">{flow.latency}</td>
                  <td className="p-3.5 text-slate-400 whitespace-nowrap">{flow.bytes}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
