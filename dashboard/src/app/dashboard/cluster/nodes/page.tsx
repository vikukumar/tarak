"use client";

import React, { useState, useEffect } from "react";
import { Cpu, Server, CheckCircle2, RefreshCw, Layers } from "lucide-react";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function NodesPage() {
  const [nodes, setNodes] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchNodes = async () => {
    setIsLoading(true);
    const res = await tarakFetch("/api/v1/nodes");
    setNodes(res.data?.items || []);
    setIsLoading(false);
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
        <div className="flex items-center gap-2">
          <Server size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{n.metadata?.name}</span>
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
      key: "version",
      header: "Kubelet Version",
      render: (n) => n.status?.nodeInfo?.kubeletVersion || "v1.30.0-tarak",
    },
    {
      key: "runtime",
      header: "Container Runtime",
      render: (n) => (
        <span className="text-cyan-300 font-mono text-[11px]">
          {n.status?.nodeInfo?.containerRuntimeVersion || "tarak://v1.30.0"}
        </span>
      ),
    },
    {
      key: "os",
      header: "OS Image",
      render: (n) => n.status?.nodeInfo?.osImage || "Windows / Linux Native",
    },
    {
      key: "age",
      header: "Age",
      render: (n) => formatAge(n.metadata?.creationTimestamp),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Cpu size={22} className="text-cyan-400" />
            <span>Cluster Nodes</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Registered physical and virtual compute instances in Tarak cluster
          </p>
        </div>

        <Button variant="secondary" size="sm" onClick={fetchNodes} isLoading={isLoading}>
          <RefreshCw size={14} />
          <span>Refresh</span>
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={nodes}
        searchKey="name"
        emptyMessage="No nodes registered in cluster"
      />
    </div>
  );
}
