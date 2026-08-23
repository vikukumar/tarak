"use client";

import React, { useState, useEffect, useRef } from "react";
import { Terminal, RefreshCw, Trash2, Download, Box, Search, Layers, Play, Pause } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";

export default function LogsPage() {
  const { selectedNamespace } = useCluster();
  const [pods, setPods] = useState<any[]>([]);
  const [selectedPod, setSelectedPod] = useState<string>("");
  const [selectedContainer, setSelectedContainer] = useState<string>("");
  const [logs, setLogs] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);
  const [isFollowing, setIsFollowing] = useState(true);
  const [searchFilter, setSearchFilter] = useState("");
  const [tailLines, setTailLines] = useState(100);
  const logBottomRef = useRef<HTMLDivElement>(null);

  // Load pods list
  useEffect(() => {
    async function loadPods() {
      const url =
        selectedNamespace === "_all"
          ? "/api/v1/pods"
          : `/api/v1/namespaces/${selectedNamespace}/pods`;
      const res = await tarakFetch(url);
      const items = res.data?.items || [];
      setPods(items);
      if (items.length > 0) {
        setSelectedPod(items[0].metadata?.name);
        const containers = items[0].spec?.containers || [];
        if (containers.length > 0) {
          setSelectedContainer(containers[0].name);
        }
      } else {
        setSelectedPod("");
        setSelectedContainer("");
        setLogs("");
      }
    }
    loadPods();
  }, [selectedNamespace]);

  // When selected pod changes, update container list
  useEffect(() => {
    const currentPod = pods.find((p) => p.metadata?.name === selectedPod);
    if (currentPod) {
      const containers = currentPod.spec?.containers || [];
      if (containers.length > 0) {
        setSelectedContainer(containers[0].name);
      }
    }
  }, [selectedPod, pods]);

  const fetchLogs = async () => {
    if (!selectedPod) return;
    setIsLoading(true);
    try {
      const currentPod = pods.find((p) => p.metadata?.name === selectedPod);
      const ns = currentPod?.metadata?.namespace || (selectedNamespace === "_all" ? "default" : selectedNamespace);
      const containerParam = selectedContainer ? `&container=${encodeURIComponent(selectedContainer)}` : "";
      const res = await tarakFetch(
        `/api/v1/namespaces/${ns}/pods/${selectedPod}/log?tailLines=${tailLines}${containerParam}`
      );
      if (res.data) {
        setLogs(
          typeof res.data === "string"
            ? res.data
            : JSON.stringify(res.data, null, 2)
        );
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, [selectedPod, selectedContainer, tailLines]);

  // Auto-follow polling
  useEffect(() => {
    if (!isFollowing || !selectedPod) return;
    const interval = setInterval(fetchLogs, 2500);
    return () => clearInterval(interval);
  }, [isFollowing, selectedPod, selectedContainer, tailLines]);

  useEffect(() => {
    if (isFollowing) {
      logBottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [logs, isFollowing]);

  const currentPodObj = pods.find((p) => p.metadata?.name === selectedPod);
  const containerList = currentPodObj?.spec?.containers || [];
  const initContainers = currentPodObj?.spec?.initContainers || [];

  const filteredLogs = logs
    .split("\n")
    .filter((line) => !searchFilter || line.toLowerCase().includes(searchFilter.toLowerCase()))
    .join("\n");

  const handleDownload = () => {
    const blob = new Blob([logs], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${selectedPod || "pod"}-${selectedContainer || "container"}-logs.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <Terminal size={24} className="text-cyan-400" />
            <span>Live Container Logs</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Realtime stdout/stderr stream from container workloads in{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {/* Pod Selector */}
          {pods.length > 0 && (
            <div className="flex items-center gap-2 bg-slate-900 border border-white/10 px-3 py-1.5 rounded-lg text-xs">
              <Box size={14} className="text-cyan-400" />
              <select
                value={selectedPod}
                onChange={(e) => setSelectedPod(e.target.value)}
                className="bg-transparent text-white outline-none cursor-pointer font-mono font-bold"
              >
                {pods.map((p) => (
                  <option
                    key={p.metadata?.name}
                    value={p.metadata?.name}
                    className="bg-slate-950 text-white"
                  >
                    {p.metadata?.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Container Selector (for multi-container pods) */}
          {(containerList.length > 1 || initContainers.length > 0) && (
            <div className="flex items-center gap-2 bg-slate-900 border border-white/10 px-3 py-1.5 rounded-lg text-xs">
              <Layers size={14} className="text-purple-400" />
              <select
                value={selectedContainer}
                onChange={(e) => setSelectedContainer(e.target.value)}
                className="bg-transparent text-purple-300 outline-none cursor-pointer font-mono font-semibold"
              >
                {containerList.map((c: any) => (
                  <option key={c.name} value={c.name} className="bg-slate-950 text-white">
                    app: {c.name}
                  </option>
                ))}
                {initContainers.map((c: any) => (
                  <option key={c.name} value={c.name} className="bg-slate-950 text-amber-300">
                    init: {c.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Follow Toggle */}
          <button
            onClick={() => setIsFollowing(!isFollowing)}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold border flex items-center gap-1.5 transition-all ${
              isFollowing
                ? "bg-emerald-500/20 border-emerald-500/40 text-emerald-300"
                : "bg-slate-900 border-white/10 text-slate-400"
            }`}
          >
            {isFollowing ? <Play size={12} /> : <Pause size={12} />}
            <span>{isFollowing ? "Following (2.5s)" : "Paused"}</span>
          </button>

          <Button
            variant="secondary"
            size="sm"
            onClick={fetchLogs}
            isLoading={isLoading}
            title="Refresh logs"
          >
            <RefreshCw size={14} />
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={handleDownload}
            title="Download log file"
          >
            <Download size={14} />
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => setLogs("")}
            title="Clear view"
          >
            <Trash2 size={14} />
          </Button>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="flex items-center gap-3">
        <div className="flex-1 relative">
          <Search size={14} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            value={searchFilter}
            onChange={(e) => setSearchFilter(e.target.value)}
            placeholder="Search log lines..."
            className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-9 pr-4 py-2 text-xs text-white placeholder:text-slate-500 focus:outline-none focus:border-cyan-400"
          />
        </div>
        <select
          value={tailLines}
          onChange={(e) => setTailLines(Number(e.target.value))}
          className="bg-slate-900 border border-white/10 rounded-xl px-3 py-2 text-xs text-slate-300 outline-none"
        >
          <option value={50}>Last 50 lines</option>
          <option value={100}>Last 100 lines</option>
          <option value={200}>Last 200 lines</option>
          <option value={500}>Last 500 lines</option>
        </select>
      </div>

      {/* Terminal View */}
      <Card className="p-0 border-white/10 overflow-hidden shadow-2xl flex flex-col h-[520px]">
        <div className="p-3 bg-slate-950 border-b border-white/10 flex items-center justify-between text-xs text-slate-400 font-mono">
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500/80" />
            <span className="w-2.5 h-2.5 rounded-full bg-amber-500/80" />
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500/80" />
            <span className="ml-2 text-slate-300 font-bold">
              {selectedPod || "No pod selected"}
            </span>
            {selectedContainer && (
              <Badge variant="purple" className="text-[10px] py-0 px-2">
                {selectedContainer}
              </Badge>
            )}
          </div>
          <span className="text-cyan-400 text-[11px]">stdout / stderr active</span>
        </div>

        <div className="flex-1 p-4 bg-[#050914] text-xs font-mono text-cyan-300 overflow-y-auto whitespace-pre-wrap leading-relaxed select-text">
          {filteredLogs || (
            <span className="text-slate-500 italic">
              {isLoading ? "Fetching stream..." : "No logs available for this container."}
            </span>
          )}
          <div ref={logBottomRef} />
        </div>
      </Card>
    </div>
  );
}
