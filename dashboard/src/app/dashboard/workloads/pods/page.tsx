"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Box, Terminal, Trash2, Code, RefreshCw, Plus, CheckCircle2, AlertCircle } from "lucide-react";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function PodsPage() {
  const { selectedNamespace } = useClusterState();
  const [pods, setPods] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedYaml, setSelectedYaml] = useState<string | null>(null);

  const fetchPods = async () => {
    setIsLoading(true);
    const res = await tarakFetch(`/api/v1/namespaces/${selectedNamespace}/pods`);
    setPods(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchPods();
  }, [selectedNamespace]);

  const handleDelete = async (name: string) => {
    if (!confirm(`Delete pod ${name}?`)) return;
    await tarakFetch(`/api/v1/namespaces/${selectedNamespace}/pods/${name}`, {
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
        <div className="flex items-center gap-2">
          <Box size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{p.metadata?.name}</span>
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
      render: (p) => p.status?.podIP || "<none>",
    },
    {
      key: "node",
      header: "Node",
      render: (p) => p.spec?.nodeName || "<none>",
    },
    {
      key: "restarts",
      header: "Restarts",
      render: (p) => {
        const restarts = p.status?.containerStatuses?.[0]?.restartCount ?? 0;
        return <span>{restarts}</span>;
      },
    },
    {
      key: "age",
      header: "Age",
      render: (p) => formatAge(p.metadata?.creationTimestamp),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (p) => (
        <div className="flex items-center justify-end gap-1.5" onClick={(e) => e.stopPropagation()}>
          <Link href={`/dashboard/devtools/terminal?pod=${p.metadata?.name}&namespace=${selectedNamespace}`}>
            <button className="p-1.5 rounded-lg bg-white/5 hover:bg-cyan-500/20 text-cyan-400 transition-colors" title="Exec Terminal">
              <Terminal size={14} />
            </button>
          </Link>
          <button
            onClick={() => setSelectedYaml(JSON.stringify(p, null, 2))}
            className="p-1.5 rounded-lg bg-white/5 hover:bg-indigo-500/20 text-indigo-400 transition-colors"
            title="Inspect JSON/YAML"
          >
            <Code size={14} />
          </button>
          <button
            onClick={() => handleDelete(p.metadata?.name)}
            className="p-1.5 rounded-lg bg-white/5 hover:bg-rose-500/20 text-rose-400 transition-colors"
            title="Delete Pod"
          >
            <Trash2 size={14} />
          </button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Box size={22} className="text-cyan-400" />
            <span>Pods</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Running container instances in namespace <span className="text-cyan-400 font-mono">{selectedNamespace}</span>
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={fetchPods} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Create Pod</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={pods}
        searchKey="name"
        searchPlaceholder="Filter pods by name..."
        emptyMessage={`No pods running in ${selectedNamespace} namespace`}
      />

      {/* Manifest Modal */}
      <Modal
        isOpen={!!selectedYaml}
        onClose={() => setSelectedYaml(null)}
        title="Resource Definition"
        maxWidth="2xl"
      >
        <pre className="p-4 rounded-xl bg-slate-950 border border-white/10 text-xs font-mono text-cyan-300 overflow-x-auto">
          {selectedYaml}
        </pre>
      </Modal>
    </div>
  );
}
