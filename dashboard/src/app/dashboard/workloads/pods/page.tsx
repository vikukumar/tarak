"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  Box,
  Terminal,
  Trash2,
  Code,
  RefreshCw,
  Plus,
  FileCode,
  Radio,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function PodsPage() {
  const { selectedNamespace } = useCluster();
  const [pods, setPods] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedPod, setSelectedPod] = useState<any | null>(null);

  const fetchPods = async () => {
    setIsLoading(true);
    try {
      const url =
        selectedNamespace === "_all"
          ? "/api/v1/pods"
          : `/api/v1/namespaces/${selectedNamespace}/pods`;
      const res = await tarakFetch(url);
      setPods(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchPods();
  }, [selectedNamespace]);

  const handleDelete = async (pod: any) => {
    const ns = pod.metadata?.namespace || selectedNamespace;
    const name = pod.metadata?.name;
    if (!confirm(`Delete pod "${name}" in namespace "${ns}"?`)) return;
    await tarakFetch(`/api/v1/namespaces/${ns}/pods/${name}`, {
      method: "DELETE",
    });
    fetchPods();
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Pod Name",
      sortable: true,
      render: (p) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
            <Box size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{p.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {p.metadata?.namespace || selectedNamespace}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "status",
      header: "Status",
      sortable: true,
      render: (p) => {
        const phase = p.status?.phase || "Pending";
        const isRunning = phase === "Running";
        return (
          <Badge variant={isRunning ? "emerald" : "amber"} dot>
            {phase}
          </Badge>
        );
      },
    },
    {
      key: "ip",
      header: "Pod IP",
      render: (p) => (
        <span className="font-mono text-cyan-300">
          {p.status?.podIP || "<none>"}
        </span>
      ),
    },
    {
      key: "node",
      header: "Node",
      render: (p) => (
        <span className="font-mono text-slate-300">
          {p.spec?.nodeName || "<none>"}
        </span>
      ),
    },
    {
      key: "restarts",
      header: "Restarts",
      render: (p) => {
        const restarts = p.status?.containerStatuses?.[0]?.restartCount ?? 0;
        return <span className="font-mono font-semibold">{restarts}</span>;
      },
    },
    {
      key: "age",
      header: "Age",
      render: (p) => (
        <span className="text-slate-400">
          {formatAge(p.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (p) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <Link
            href={`/dashboard/devtools/terminal?pod=${p.metadata?.name}&namespace=${
              p.metadata?.namespace || selectedNamespace
            }`}
          >
            <button
              className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
              title="Exec Web Terminal"
            >
              <Terminal size={14} />
            </button>
          </Link>
          <button
            onClick={() => setSelectedPod(p)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Details & Logs"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => handleDelete(p)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Pod"
          >
            <Trash2 size={14} />
          </button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <Box size={24} className="text-cyan-400" />
            <span>Pods</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Running container workloads in scope{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchPods}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Deploy Manifest</span>
            </Button>
          </Link>
        </div>
      </div>

      {/* Interactive Data Table */}
      <DataTable
        columns={columns}
        data={pods}
        searchKey="name"
        searchPlaceholder="Filter pods by name..."
        emptyMessage={`No pods found in ${
          selectedNamespace === "_all" ? "cluster" : selectedNamespace + " namespace"
        }`}
        onRowClick={(pod) => setSelectedPod(pod)}
      />

      {/* Slide-Over Detail Drawer */}
      <ResourceDetailDrawer
        isOpen={!!selectedPod}
        onClose={() => setSelectedPod(null)}
        resourceType="Pod"
        resourceName={selectedPod?.metadata?.name || ""}
        namespace={selectedPod?.metadata?.namespace || selectedNamespace}
        rawResource={selectedPod}
        onActionComplete={fetchPods}
      />
    </div>
  );
}
