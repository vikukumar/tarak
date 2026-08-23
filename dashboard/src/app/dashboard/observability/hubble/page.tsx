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
}

const generateMockFlows = (): NetworkFlow[] => {
  const now = new Date();
  const formatTime = (d: Date) => {
    const pad = (n: number) => n.toString().padStart(2, "0");
    const ms = d.getMilliseconds().toString().padStart(3, "0");
    return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}.${ms}`;
  };

  return [
    {
      id: "flow-101",
      time: formatTime(now),
      isoTime: now.toISOString(),
      source: "ingress-controller-7bf8",
      dest: "storefront-web-59d4",
      sourceNamespace: "tarak-system",
      destNamespace: "production",
      protocol: "HTTP/2 (GET /api/v1/products)",
      port: 8080,
      verdict: "FORWARDED",
      latency: "0.42ms",
      bytes: "1.4 KB",
    },
    {
      id: "flow-102",
      time: formatTime(new Date(now.getTime() - 250)),
      isoTime: new Date(now.getTime() - 250).toISOString(),
      source: "storefront-web-59d4",
      dest: "payments-api-91ec",
      sourceNamespace: "production",
      destNamespace: "finance",
      protocol: "gRPC (OrderService.Process)",
      port: 9000,
      verdict: "FORWARDED",
      latency: "0.85ms",
      bytes: "850 B",
    },
    {
      id: "flow-103",
      time: formatTime(new Date(now.getTime() - 600)),
      isoTime: new Date(now.getTime() - 600).toISOString(),
      source: "unknown-scanner (198.51.100.24)",
      dest: "cluster-apiserver:6443",
      sourceNamespace: "external",
      destNamespace: "tarak-system",
      protocol: "TCP SYN",
      port: 6443,
      verdict: "DROPPED",
      latency: "0.01ms",
      bytes: "64 B",
    },
    {
      id: "flow-104",
      time: formatTime(new Date(now.getTime() - 1100)),
      isoTime: new Date(now.getTime() - 1100).toISOString(),
      source: "storefront-web-59d4",
      dest: "coredns-resolver:53",
      sourceNamespace: "production",
      destNamespace: "kube-system",
      protocol: "UDP (DNS query: payments.finance.svc)",
      port: 53,
      verdict: "FORWARDED",
      latency: "0.12ms",
      bytes: "128 B",
    },
    {
      id: "flow-105",
      time: formatTime(new Date(now.getTime() - 1600)),
      isoTime: new Date(now.getTime() - 1600).toISOString(),
      source: "payments-api-91ec",
      dest: "payments-db-0",
      sourceNamespace: "finance",
      destNamespace: "finance",
      protocol: "PostgreSQL (mTLS Encrypted)",
      port: 5432,
      verdict: "FORWARDED",
      latency: "0.31ms",
      bytes: "4.2 KB",
    },
  ];
};

export default function HubblePage() {
  const [flows, setFlows] = useState<NetworkFlow[]>(generateMockFlows());
  const [isLive, setIsLive] = useState(true);
  const [search, setSearch] = useState("");
  const [selectedVerdict, setSelectedVerdict] = useState<string>("All");
  const [selectedFlow, setSelectedFlow] = useState<NetworkFlow | null>(flows[0]);

  useEffect(() => {
    if (!isLive) return;
    const interval = setInterval(() => {
      setFlows(generateMockFlows());
    }, 2000);
    return () => clearInterval(interval);
  }, [isLive]);

  const filteredFlows = flows.filter(
    (f) =>
      (selectedVerdict === "All" || f.verdict === selectedVerdict) &&
      (f.source.toLowerCase().includes(search.toLowerCase()) ||
        f.dest.toLowerCase().includes(search.toLowerCase()) ||
        f.protocol.toLowerCase().includes(search.toLowerCase()))
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
              <Activity size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Hubble Network Flow Visualizer <span className="text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 via-cyan-300 to-purple-400">(eBPF & L7 Traffic)</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Real-time visual service dependency map, container traffic streams, and microsecond packet inspection.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setIsLive(!isLive)}
            className={`text-xs ${isLive ? "text-emerald-400 border-emerald-500/40" : "text-slate-400"}`}
          >
            {isLive ? <Pause size={14} className="mr-1" /> : <Play size={14} className="mr-1" />}
            {isLive ? "Live Stream (Active)" : "Paused"}
          </Button>
          <Button size="sm" onClick={() => setFlows(generateMockFlows())} className="bg-slate-800 text-white">
            <RefreshCw size={14} className="mr-1" /> Poll Now
          </Button>
        </div>
      </div>

      {/* Visual Hubble Graph Overview */}
      <div className="p-6 rounded-2xl bg-gradient-to-b from-[#070c18] to-[#04060c] border border-white/10 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <span className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5 font-mono">
            <Radio size={14} className="text-emerald-400 animate-pulse" /> Live Container Service Mesh Topology
          </span>
          <span className="text-xs text-emerald-400 font-mono">Flow rate: ~2.4k pkts/sec</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 pt-2">
          <div className="p-4 rounded-xl bg-slate-900/60 border border-cyan-500/30 text-center space-y-2 relative overflow-hidden group">
            <Globe className="text-cyan-400 mx-auto" size={24} />
            <div className="text-xs font-bold text-white">Ingress Proxy</div>
            <span className="text-[10px] text-cyan-300 font-mono block">10.244.0.1</span>
            <div className="text-[10px] text-slate-400">Public HTTP Gateway</div>
          </div>

          <div className="p-4 rounded-xl bg-slate-900/60 border border-purple-500/30 text-center space-y-2 relative overflow-hidden group">
            <Box className="text-purple-400 mx-auto" size={24} />
            <div className="text-xs font-bold text-white">Storefront Web</div>
            <span className="text-[10px] text-purple-300 font-mono block">10.244.0.2</span>
            <div className="text-[10px] text-slate-400">production namespace</div>
          </div>

          <div className="p-4 rounded-xl bg-slate-900/60 border border-emerald-500/30 text-center space-y-2 relative overflow-hidden group">
            <Server className="text-emerald-400 mx-auto" size={24} />
            <div className="text-xs font-bold text-white">Payments API</div>
            <span className="text-[10px] text-emerald-300 font-mono block">10.244.1.5</span>
            <div className="text-[10px] text-slate-400">finance namespace</div>
          </div>

          <div className="p-4 rounded-xl bg-slate-900/60 border border-amber-500/30 text-center space-y-2 relative overflow-hidden group">
            <Database className="text-amber-400 mx-auto" size={24} />
            <div className="text-xs font-bold text-white">Payments DB</div>
            <span className="text-[10px] text-amber-300 font-mono block">10.244.1.8</span>
            <div className="text-[10px] text-slate-400">PostgreSQL (StatefulSet)</div>
          </div>
        </div>
      </div>

      {/* Filter and Table */}
      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row items-center gap-3">
          <div className="relative flex-1 w-full">
            <Search className="absolute left-3 top-2.5 text-slate-500" size={16} />
            <input
              type="text"
              placeholder="Search by pod, protocol, IP, or namespace..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-9 pr-4 py-2 bg-slate-950/80 border border-white/10 rounded-xl text-xs sm:text-sm text-white placeholder-slate-500 focus:outline-none focus:border-emerald-500/50 font-mono"
            />
          </div>

          <div className="flex items-center gap-1.5">
            {["All", "FORWARDED", "DROPPED"].map((v) => (
              <button
                key={v}
                onClick={() => setSelectedVerdict(v)}
                className={`px-3 py-1.5 rounded-xl text-xs font-bold font-mono transition-all border ${
                  selectedVerdict === v
                    ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40"
                    : "bg-slate-900/60 text-slate-400 border-white/10 hover:text-white"
                }`}
              >
                {v}
              </button>
            ))}
          </div>
        </div>

        {/* Flows Table */}
        <div className="overflow-x-auto rounded-2xl border border-white/10 bg-slate-900/60 shadow-xl">
          <table className="w-full text-left text-xs border-collapse font-mono">
            <thead>
              <tr className="bg-slate-950/90 text-slate-300 uppercase tracking-wider font-bold border-b border-white/10">
                <th className="p-3.5">Timestamp (Precise)</th>
                <th className="p-3.5">Source Pod</th>
                <th className="p-3.5">Destination Pod</th>
                <th className="p-3.5">L7 Protocol / Command</th>
                <th className="p-3.5">Port</th>
                <th className="p-3.5">Verdict</th>
                <th className="p-3.5 text-right">Latency</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {filteredFlows.map((flow) => (
                <tr
                  key={flow.id}
                  onClick={() => setSelectedFlow(flow)}
                  className="hover:bg-white/[0.03] transition-colors cursor-pointer"
                >
                  <td className="p-3.5 text-slate-400 font-bold whitespace-nowrap">{flow.time}</td>
                  <td className="p-3.5 text-cyan-300 font-bold whitespace-nowrap">{flow.source}</td>
                  <td className="p-3.5 text-purple-300 font-bold whitespace-nowrap">{flow.dest}</td>
                  <td className="p-3.5 text-slate-300 max-w-xs truncate">{flow.protocol}</td>
                  <td className="p-3.5 text-slate-400 font-bold">{flow.port}</td>
                  <td className="p-3.5">
                    <Badge variant={flow.verdict === "FORWARDED" ? "emerald" : "rose"}>
                      {flow.verdict}
                    </Badge>
                  </td>
                  <td className="p-3.5 text-right text-emerald-400 font-bold whitespace-nowrap">{flow.latency}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Selected Flow Inspector */}
      {selectedFlow && (
        <div className="p-5 rounded-2xl bg-slate-900/80 border border-white/10 space-y-3 shadow-2xl">
          <div className="flex items-center justify-between border-b border-white/10 pb-2">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-300 font-mono flex items-center gap-2">
              <Zap size={14} className="text-cyan-400" /> Deep Flow Packet Inspector: {selectedFlow.id}
            </span>
            <span className="text-xs text-slate-400 font-mono">{selectedFlow.isoTime}</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs font-mono">
            <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 space-y-1">
              <span className="text-[10px] text-slate-400 uppercase font-sans">Source Endpoint</span>
              <div className="text-cyan-300 font-bold">{selectedFlow.source}</div>
              <div className="text-slate-400">Namespace: {selectedFlow.sourceNamespace}</div>
            </div>

            <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 space-y-1">
              <span className="text-[10px] text-slate-400 uppercase font-sans">Destination Endpoint</span>
              <div className="text-purple-300 font-bold">{selectedFlow.dest}</div>
              <div className="text-slate-400">Namespace: {selectedFlow.destNamespace}</div>
            </div>

            <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 space-y-1">
              <span className="text-[10px] text-slate-400 uppercase font-sans">Telemetry & Payload</span>
              <div className="text-emerald-300 font-bold">Latency: {selectedFlow.latency}</div>
              <div className="text-slate-400">Transferred: {selectedFlow.bytes}</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
