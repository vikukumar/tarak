"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Workflow, RefreshCw, Plus, Trash2, Edit3, Code, Play } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function DeploymentsPage() {
  const { selectedNamespace } = useClusterState();
  const [deployments, setDeployments] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedYaml, setSelectedYaml] = useState<string | null>(null);
  const [scaleDeployment, setScaleDeployment] = useState<any | null>(null);
  const [scaleCount, setScaleCount] = useState(1);

  const fetchDeployments = async () => {
    setIsLoading(true);
    const res = await tarakFetch(`/apis/apps/v1/namespaces/${selectedNamespace}/deployments`);
    setDeployments(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchDeployments();
  }, [selectedNamespace]);

  const handleScale = async () => {
    if (!scaleDeployment) return;
    const name = scaleDeployment.metadata?.name;
    const updated = {
      ...scaleDeployment,
      spec: {
        ...scaleDeployment.spec,
        replicas: scaleCount,
      },
    };
    await tarakFetch(`/apis/apps/v1/namespaces/${selectedNamespace}/deployments/${name}`, {
      method: "PUT",
      body: JSON.stringify(updated),
    });
    setScaleDeployment(null);
    fetchDeployments();
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Delete deployment ${name}?`)) return;
    await tarakFetch(`/apis/apps/v1/namespaces/${selectedNamespace}/deployments/${name}`, {
      method: "DELETE",
    });
    fetchDeployments();
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Deployment Name",
      sortable: true,
      render: (d) => (
        <div className="flex items-center gap-2">
          <Workflow size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{d.metadata?.name}</span>
        </div>
      ),
    },
    {
      key: "ready",
      header: "Pods Ready",
      render: (d) => {
        const ready = d.status?.readyReplicas || 0;
        const total = d.spec?.replicas || 1;
        const isReady = ready === total && total > 0;
        return (
          <Badge variant={isReady ? "emerald" : "amber"} dot>
            {ready}/{total}
          </Badge>
        );
      },
    },
    {
      key: "image",
      header: "Container Image",
      render: (d) => (
        <span className="text-cyan-300 font-mono">
          {d.spec?.template?.spec?.containers?.[0]?.image || "nginx:alpine"}
        </span>
      ),
    },
    {
      key: "age",
      header: "Age",
      render: (d) => formatAge(d.metadata?.creationTimestamp),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (d) => (
        <div className="flex items-center justify-end gap-1.5" onClick={(e) => e.stopPropagation()}>
          <button
            onClick={() => {
              setScaleDeployment(d);
              setScaleCount(d.spec?.replicas || 1);
            }}
            className="p-1.5 rounded-lg bg-white/5 hover:bg-cyan-500/20 text-cyan-400 transition-colors"
            title="Scale Replicas"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedYaml(JSON.stringify(d, null, 2))}
            className="p-1.5 rounded-lg bg-white/5 hover:bg-indigo-500/20 text-indigo-400 transition-colors"
            title="Inspect Definition"
          >
            <Code size={14} />
          </button>
          <button
            onClick={() => handleDelete(d.metadata?.name)}
            className="p-1.5 rounded-lg bg-white/5 hover:bg-rose-500/20 text-rose-400 transition-colors"
            title="Delete Deployment"
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
            <Workflow size={22} className="text-cyan-400" />
            <span>Deployments</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Declarative workload rollouts in namespace <span className="text-cyan-400 font-mono">{selectedNamespace}</span>
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={fetchDeployments} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Deploy New</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={deployments}
        searchKey="name"
        searchPlaceholder="Filter deployments..."
        emptyMessage={`No deployments found in ${selectedNamespace} namespace`}
      />

      {/* Scale Modal */}
      <Modal
        isOpen={!!scaleDeployment}
        onClose={() => setScaleDeployment(null)}
        title={`Scale ${scaleDeployment?.metadata?.name}`}
        maxWidth="sm"
      >
        <div className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Desired Replicas</label>
            <input
              type="number"
              min={0}
              max={100}
              value={scaleCount}
              onChange={(e) => setScaleCount(parseInt(e.target.value) || 0)}
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <Button onClick={handleScale} className="w-full">
            Apply Scale
          </Button>
        </div>
      </Modal>

      {/* Definition Modal */}
      <Modal
        isOpen={!!selectedYaml}
        onClose={() => setSelectedYaml(null)}
        title="Deployment Definition"
        maxWidth="2xl"
      >
        <pre className="p-4 rounded-xl bg-slate-950 border border-white/10 text-xs font-mono text-cyan-300 overflow-x-auto">
          {selectedYaml}
        </pre>
      </Modal>
    </div>
  );
}
