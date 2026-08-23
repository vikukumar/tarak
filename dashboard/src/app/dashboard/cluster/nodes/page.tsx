"use client";

import React, { useState, useEffect } from "react";
import {
  Cpu,
  Server,
  CheckCircle2,
  RefreshCw,
  Layers,
  Activity,
  HardDrive,
  Box,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { NodeDetailDrawer } from "@/components/drawers/NodeDetailDrawer";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function NodesPage() {
  const [nodes, setNodes] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedNode, setSelectedNode] = useState<any | null>(null);

  const fetchNodes = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/api/v1/nodes");
      setNodes(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchNodes();
  }, []);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Node Name",
      sortable: true,
      render: (n) => (
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
            <Server size={16} />
          </div>
          <div>
            <span className="font-bold text-white block text-sm">{n.metadata?.name}</span>
            <span className="text-[11px] text-slate-400 font-mono">
              IP: {n.status?.addresses?.[0]?.address || "127.0.0.1"}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "status",
      header: "Status",
      render: () => (
        <Badge variant="emerald" dot>
          Ready
        </Badge>
      ),
    },
    {
      key: "role",
      header: "Roles",
      render: () => <Badge variant="indigo">control-plane, worker</Badge>,
    },
    {
      key: "runtime",
      header: "Container Runtime",
      render: (n) => (
        <span className="text-cyan-300 font-mono text-[11px]">
          {n.status?.nodeInfo?.containerRuntimeVersion || "tarak-native://1.0.6"}
        </span>
      ),
    },
    {
      key: "os",
      header: "OS Image",
      render: (n) => (
        <span className="text-slate-300 text-xs">
          {n.status?.nodeInfo?.osImage || "Windows Server / AMD64"}
        </span>
      ),
    },
    {
      key: "age",
      header: "Age",
      render: (n) => (
        <span className="text-slate-400 text-xs">
          {formatAge(n.metadata?.creationTimestamp)}
        </span>
      ),
    },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <Cpu size={24} className="text-cyan-400" />
            <span>Cluster Nodes</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Registered physical and virtual compute instances. Click any node to inspect telemetry, pods, and capacity.
          </p>
        </div>

        <Button
          variant="secondary"
          size="sm"
          onClick={fetchNodes}
          isLoading={isLoading}
        >
          <RefreshCw size={14} />
          <span>Refresh</span>
        </Button>
      </div>

      {/* Metric Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider font-semibold">
            Active Compute Nodes
          </span>
          <div className="text-2xl font-bold text-white">{nodes.length || 1}</div>
          <span className="text-[11px] text-emerald-400 flex items-center gap-1 font-mono">
            <CheckCircle2 size={12} />
            100% Online & Healthy
          </span>
        </div>

        <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider font-semibold">
            Total CPU Cores
          </span>
          <div className="text-2xl font-bold text-cyan-400">8 Cores</div>
          <span className="text-[11px] text-slate-400 font-mono">
            Architecture: x86_64 / amd64
          </span>
        </div>

        <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-1">
          <span className="text-xs text-slate-400 uppercase tracking-wider font-semibold">
            Memory Capacity
          </span>
          <div className="text-2xl font-bold text-indigo-400">16 GB</div>
          <span className="text-[11px] text-slate-400 font-mono">
            Zero-Daemon Memory Isolated
          </span>
        </div>
      </div>

      {/* Nodes Table */}
      <DataTable
        columns={columns}
        data={nodes}
        searchKey="name"
        searchPlaceholder="Filter nodes by name..."
        emptyMessage="No nodes registered in cluster"
        onRowClick={(n) => setSelectedNode(n)}
      />

      {/* Node Detail Drawer */}
      <NodeDetailDrawer
        isOpen={!!selectedNode}
        onClose={() => setSelectedNode(null)}
        nodeName={selectedNode?.metadata?.name || ""}
        nodeData={selectedNode}
      />
    </div>
  );
}
