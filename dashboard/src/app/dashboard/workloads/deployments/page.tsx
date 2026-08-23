"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import {
  Workflow,
  RefreshCw,
  Plus,
  Trash2,
  Edit3,
  FileCode,
  Layers,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function DeploymentsPage() {
  const { selectedNamespace } = useCluster();
  const [deployments, setDeployments] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedDeployment, setSelectedDeployment] = useState<any | null>(null);
  const [scaleDeployment, setScaleDeployment] = useState<any | null>(null);
  const [scaleCount, setScaleCount] = useState(1);
  const [depToDelete, setDepToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchDeployments = async () => {
    setIsLoading(true);
    try {
      const url =
        selectedNamespace === "_all"
          ? "/apis/apps/v1/deployments"
          : `/apis/apps/v1/namespaces/${selectedNamespace}/deployments`;
      const res = await tarakFetch(url);
      setDeployments(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchDeployments();
  }, [selectedNamespace]);

  const handleScale = async () => {
    if (!scaleDeployment) return;
    const ns = scaleDeployment.metadata?.namespace || selectedNamespace;
    const name = scaleDeployment.metadata?.name;
    const updated = {
      ...scaleDeployment,
      spec: {
        ...scaleDeployment.spec,
        replicas: scaleCount,
      },
    };
    await tarakFetch(`/apis/apps/v1/namespaces/${ns}/deployments/${name}`, {
      method: "PUT",
      body: JSON.stringify(updated),
    });
    setScaleDeployment(null);
    fetchDeployments();
  };

  const confirmDelete = async () => {
    if (!depToDelete) return;
    setIsDeleting(true);
    try {
      const ns = depToDelete.metadata?.namespace || selectedNamespace;
      const name = depToDelete.metadata?.name;
      await tarakFetch(`/apis/apps/v1/namespaces/${ns}/deployments/${name}`, {
        method: "DELETE",
      });
      setDepToDelete(null);
      fetchDeployments();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Deployment Name",
      sortable: true,
      render: (d) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-indigo-500/15 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
            <Workflow size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{d.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {d.metadata?.namespace || selectedNamespace}
            </span>
          </div>
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
      render: (d) => (
        <span className="text-slate-400">
          {formatAge(d.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (d) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => {
              setScaleDeployment(d);
              setScaleCount(d.spec?.replicas || 1);
            }}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Scale Replicas"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedDeployment(d)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Definition"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setDepToDelete(d)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Deployment"
          >
            <Trash2 size={14} />
          </button>
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <Workflow size={24} className="text-indigo-400" />
            <span>Deployments</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Declarative workload rollouts in scope{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchDeployments}
            isLoading={isLoading}
          >
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
        searchPlaceholder="Filter deployments by name..."
        emptyMessage={`No deployments found in ${
          selectedNamespace === "_all" ? "cluster" : selectedNamespace + " namespace"
        }`}
        onRowClick={(d) => setSelectedDeployment(d)}
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
            <label className="text-xs font-semibold text-slate-300">
              Desired Replicas
            </label>
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

      {/* Slide-Over Detail Drawer */}
      <ResourceDetailDrawer
        isOpen={!!selectedDeployment}
        onClose={() => setSelectedDeployment(null)}
        resourceType="Deployment"
        resourceName={selectedDeployment?.metadata?.name || ""}
        namespace={selectedDeployment?.metadata?.namespace || selectedNamespace}
        rawResource={selectedDeployment}
        onActionComplete={fetchDeployments}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={!!depToDelete}
        onClose={() => setDepToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete Deployment"
        message={`Are you sure you want to delete deployment "${depToDelete?.metadata?.name}" from namespace "${depToDelete?.metadata?.namespace || selectedNamespace}"? All child pods and running containers will be descaled and terminated.`}
        confirmText="Delete Deployment"
        isLoading={isDeleting}
      />
    </div>
  );
}
