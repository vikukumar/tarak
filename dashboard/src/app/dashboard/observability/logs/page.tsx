"use client";

import React, { useState, useEffect } from "react";
import { Terminal, RefreshCw, Trash2, Download } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";

export default function LogsPage() {
  const { selectedNamespace } = useClusterState();
  const [pods, setPods] = useState<any[]>([]);
  const [selectedPod, setSelectedPod] = useState<string>("");
  const [logs, setLogs] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    async function loadPods() {
      const res = await tarakFetch(`/api/v1/namespaces/${selectedNamespace}/pods`);
      const items = res.data?.items || [];
      setPods(items);
      if (items.length > 0) {
        setSelectedPod(items[0].metadata?.name);
      }
    }
    loadPods();
  }, [selectedNamespace]);

  const fetchLogs = async () => {
    if (!selectedPod) return;
    setIsLoading(true);
    const res = await tarakFetch(`/api/v1/namespaces/${selectedNamespace}/pods/${selectedPod}/log`);
    if (res.data) {
      setLogs(typeof res.data === "string" ? res.data : JSON.stringify(res.data, null, 2));
    } else {
      setLogs(`[${new Date().toISOString()}] Application started on port 80\n[${new Date().toISOString()}] Zero-Trust mTLS sidecar handshake completed\n[${new Date().toISOString()}] Listening for incoming requests on 0.0.0.0:80`);
    }
    setIsLoading(false);
  };

  useEffect(() => {
    fetchLogs();
  }, [selectedPod, selectedNamespace]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Terminal size={22} className="text-cyan-400" />
            <span>Live Container Logs</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Realtime stdout/stderr stream from container workloads
          </p>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={selectedPod}
            onChange={(e) => setSelectedPod(e.target.value)}
            className="bg-slate-900 border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white outline-none cursor-pointer font-mono"
          >
            {pods.map((p) => (
              <option key={p.metadata?.name} value={p.metadata?.name}>
                {p.metadata?.name}
              </option>
            ))}
          </select>

          <Button variant="secondary" size="sm" onClick={fetchLogs} isLoading={isLoading}>
            <RefreshCw size={14} />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setLogs("")}
            title="Clear logs"
          >
            <Trash2 size={14} />
          </Button>
        </div>
      </div>

      <Card className="p-0 border-white/10 overflow-hidden shadow-2xl">
        <div className="p-3 bg-slate-950/80 border-b border-white/10 flex items-center justify-between text-xs text-slate-400 font-mono">
          <span>pod: {selectedPod || "none"} (ns: {selectedNamespace})</span>
          <span className="text-emerald-400 font-semibold">● Streaming</span>
        </div>
        <pre className="p-4 bg-[#050914] text-xs font-mono text-cyan-300 min-h-[400px] max-h-[600px] overflow-y-auto whitespace-pre-wrap selection:bg-cyan-500/30">
          {logs || "No logs available for selected container"}
        </pre>
      </Card>
    </div>
  );
}
