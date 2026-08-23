"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, RefreshCw, Plus, FileCode, Trash2, Edit3, Shield, Cpu } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function DaemonSetsPage() {
  const { selectedNamespace } = useCluster();
  const [daemonSets, setDaemonSets] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedDS, setSelectedDS] = useState<any | null>(null);
  const [dsToEdit, setDsToEdit] = useState<any | null>(null);
  const [dsToDelete, setDsToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchDaemonSets = async () => {
    setIsLoading(true);
    try {
      const url =
        selectedNamespace === "_all"
          ? "/apis/apps/v1/daemonsets"
          : `/apis/apps/v1/namespaces/${selectedNamespace}/daemonsets`;
      const res = await tarakFetch(url);
      setDaemonSets(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchDaemonSets();
  }, [selectedNamespace]);

  const confirmDelete = async () => {
    if (!dsToDelete) return;
    setIsDeleting(true);
    try {
      const ns =
        dsToDelete.metadata?.namespace ||
        (selectedNamespace === "_all" ? "default" : selectedNamespace);
      const name = dsToDelete.metadata?.name;
      await tarakFetch(`/apis/apps/v1/namespaces/${ns}/daemonsets/${name}`, {
        method: "DELETE",
      });
      setDsToDelete(null);
      fetchDaemonSets();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "DaemonSet Name",
      sortable: true,
      render: (ds) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
            <Layers size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{ds.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {ds.metadata?.namespace || selectedNamespace}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "desired",
      header: "Desired Nodes",
      render: (ds) => (
        <span className="font-mono text-xs font-semibold text-emerald-400">
          {ds.status?.desiredNumberScheduled || 1} Nodes Active
        </span>
      ),
    },
    {
      key: "current",
      header: "Current",
      render: (ds) => (
        <Badge variant="emerald">
          {ds.status?.currentNumberScheduled || 1} / {ds.status?.desiredNumberScheduled || 1}
        </Badge>
      ),
    },
    {
      key: "image",
      header: "Container Image",
      render: (ds) => (
        <span className="text-cyan-300 font-mono text-xs">
          {ds.spec?.template?.spec?.containers?.[0]?.image || "tarak-agent:latest"}
        </span>
      ),
    },
    {
      key: "age",
      header: "Age",
      render: (ds) => (
        <span className="text-slate-400 text-xs">
          {formatAge(ds.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (ds) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setDsToEdit(ds)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Modify DaemonSet"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedDS(ds)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect DaemonSet"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setDsToDelete(ds)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete DaemonSet"
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
            <Layers size={24} className="text-cyan-400" />
            <span>DaemonSets</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            System background pods running across every cluster host node in{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchDaemonSets}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Create DaemonSet</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={daemonSets}
        searchKey="name"
        searchPlaceholder="Filter daemonsets by name..."
        emptyMessage={`No daemonsets deployed in ${
          selectedNamespace === "_all" ? "cluster" : selectedNamespace + " namespace"
        }`}
        onRowClick={(ds) => setSelectedDS(ds)}
      />

      <ResourceDetailDrawer
        isOpen={!!selectedDS}
        onClose={() => setSelectedDS(null)}
        resourceType="DaemonSet"
        resourceName={selectedDS?.metadata?.name || ""}
        namespace={selectedDS?.metadata?.namespace || selectedNamespace}
        rawResource={selectedDS}
        onActionComplete={fetchDaemonSets}
      />

      <EditResourceModal
        isOpen={!!dsToEdit}
        onClose={() => setDsToEdit(null)}
        resourceType="DaemonSet"
        resourceName={dsToEdit?.metadata?.name || ""}
        namespace={dsToEdit?.metadata?.namespace || selectedNamespace}
        rawResource={dsToEdit}
        onSaved={fetchDaemonSets}
      />

      <ConfirmModal
        isOpen={!!dsToDelete}
        onClose={() => setDsToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete DaemonSet"
        message={`Are you sure you want to delete DaemonSet "${dsToDelete?.metadata?.name}"? Its active pods on all nodes will be terminated.`}
        confirmText="Delete DaemonSet"
        isLoading={isDeleting}
      />
    </div>
  );
}
