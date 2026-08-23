"use client";

import React, { useState, useEffect } from "react";
import { Activity, RefreshCw, AlertTriangle, Info, CheckCircle2, XCircle, Search, Filter } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function EventsPage() {
  const { selectedNamespace } = useCluster();
  const [events, setEvents] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [filterType, setFilterType] = useState<"all" | "Normal" | "Warning">("all");

  const fetchEvents = async () => {
    setIsLoading(true);
    try {
      const url =
        selectedNamespace === "_all"
          ? "/api/v1/events"
          : `/api/v1/namespaces/${selectedNamespace}/events`;
      const res = await tarakFetch(url);
      setEvents(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchEvents();
    const interval = setInterval(fetchEvents, 5000);
    return () => clearInterval(interval);
  }, [selectedNamespace]);

  const filteredEvents = events.filter((e) => {
    if (filterType === "all") return true;
    return (e.type || "Normal") === filterType;
  });

  const columns: Column<any>[] = [
    {
      key: "type",
      header: "Type",
      render: (ev) => {
        const isWarn = ev.type === "Warning";
        return (
          <Badge variant={isWarn ? "rose" : "emerald"} dot>
            {ev.type || "Normal"}
          </Badge>
        );
      },
    },
    {
      key: "reason",
      header: "Reason",
      sortable: true,
      render: (ev) => (
        <span className="font-bold font-mono text-xs text-white">
          {ev.reason || "Scheduled"}
        </span>
      ),
    },
    {
      key: "object",
      header: "Involved Object",
      render: (ev) => {
        const obj = ev.involvedObject || {};
        return (
          <div>
            <span className="font-semibold text-cyan-300 block">
              {obj.kind || "Pod"} / {obj.name || ev.metadata?.name}
            </span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {obj.namespace || ev.metadata?.namespace || "default"}
            </span>
          </div>
        );
      },
    },
    {
      key: "message",
      header: "Message",
      render: (ev) => (
        <span className="text-slate-300 text-xs line-clamp-2 max-w-md font-sans">
          {ev.message || "Reconciled resource successfully"}
        </span>
      ),
    },
    {
      key: "source",
      header: "Source Component",
      render: (ev) => (
        <span className="px-2 py-0.5 rounded bg-slate-900 border border-white/10 font-mono text-[11px] text-indigo-300">
          {ev.source?.component || "tarak-controller"}
        </span>
      ),
    },
    {
      key: "age",
      header: "Time",
      render: (ev) => (
        <span className="text-slate-400 text-xs font-mono">
          {formatAge(ev.lastTimestamp || ev.metadata?.creationTimestamp)}
        </span>
      ),
    },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <Activity size={24} className="text-cyan-400" />
            <span>Diagnostic Events & Audit</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Real-time cluster lifecycle audit log and state changes in{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchEvents}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Sync Events</span>
          </Button>
        </div>
      </div>

      {/* Filter Chips */}
      <div className="flex items-center gap-2">
        <button
          onClick={() => setFilterType("all")}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all ${
            filterType === "all"
              ? "bg-cyan-500/20 border-cyan-500/40 text-cyan-300"
              : "bg-slate-900/50 border-white/10 text-slate-400 hover:text-white"
          }`}
        >
          All Events ({events.length})
        </button>
        <button
          onClick={() => setFilterType("Normal")}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all ${
            filterType === "Normal"
              ? "bg-emerald-500/20 border-emerald-500/40 text-emerald-300"
              : "bg-slate-900/50 border-white/10 text-slate-400 hover:text-white"
          }`}
        >
          Normal ({events.filter((e) => (e.type || "Normal") === "Normal").length})
        </button>
        <button
          onClick={() => setFilterType("Warning")}
          className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all ${
            filterType === "Warning"
              ? "bg-rose-500/20 border-rose-500/40 text-rose-300"
              : "bg-slate-900/50 border-white/10 text-slate-400 hover:text-white"
          }`}
        >
          Warnings ({events.filter((e) => e.type === "Warning").length})
        </button>
      </div>

      <DataTable
        columns={columns}
        data={filteredEvents}
        searchKey="message"
        searchPlaceholder="Search event messages..."
        emptyMessage="No diagnostic events captured in this interval."
      />
    </div>
  );
}
