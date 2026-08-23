"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, RefreshCw, Plus, FileCode, Trash2, Edit3, Code, Shield, Box } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function CRDsPage() {
  const [crds, setCrds] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [crdGroup, setCrdGroup] = useState("apps.tarak.io");
  const [crdKind, setCrdKind] = useState("MicroService");
  const [crdPlural, setCrdPlural] = useState("microservices");
  const [crdScope, setCrdScope] = useState("Namespaced");
  const [selectedCrd, setSelectedCrd] = useState<any | null>(null);
  const [crdToEdit, setCrdToEdit] = useState<any | null>(null);
  const [crdToDelete, setCrdToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchCrds = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/apiextensions.k8s.io/v1/customresourcedefinitions");
      setCrds(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchCrds();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!crdKind.trim() || !crdGroup.trim()) return;

    const crdName = `${crdPlural.toLowerCase()}.${crdGroup.toLowerCase()}`;
    await tarakFetch("/apis/apiextensions.k8s.io/v1/customresourcedefinitions", {
      method: "POST",
      body: JSON.stringify({
        apiVersion: "apiextensions.k8s.io/v1",
        kind: "CustomResourceDefinition",
        metadata: {
          name: crdName,
        },
        spec: {
          group: crdGroup.trim().toLowerCase(),
          names: {
            kind: crdKind.trim(),
            plural: crdPlural.trim().toLowerCase(),
            singular: crdKind.trim().toLowerCase(),
          },
          scope: crdScope,
          versions: [
            {
              name: "v1",
              served: true,
              storage: true,
              schema: {
                openAPIV3Schema: {
                  type: "object",
                  properties: {
                    spec: { type: "object" },
                    status: { type: "object" },
                  },
                },
              },
            },
          ],
        },
      }),
    });
    setIsCreateOpen(false);
    fetchCrds();
  };

  const confirmDelete = async () => {
    if (!crdToDelete) return;
    setIsDeleting(true);
    try {
      const name = crdToDelete.metadata?.name;
      await tarakFetch(`/apis/apiextensions.k8s.io/v1/customresourcedefinitions/${name}`, {
        method: "DELETE",
      });
      setCrdToDelete(null);
      fetchCrds();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Custom Resource (Kind)",
      sortable: true,
      render: (crd) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-amber-500/15 border border-amber-500/30 flex items-center justify-center text-amber-400">
            <Code size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">
              {crd.spec?.names?.kind || crd.metadata?.name}
            </span>
            <span className="text-[10px] text-slate-400 font-mono">
              {crd.metadata?.name}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "group",
      header: "API Group",
      render: (crd) => (
        <span className="font-mono text-xs text-cyan-300">
          {crd.spec?.group || "apiextensions.k8s.io"}
        </span>
      ),
    },
    {
      key: "scope",
      header: "Scope",
      render: (crd) => (
        <Badge variant={crd.spec?.scope === "Cluster" ? "purple" : "cyan"}>
          {crd.spec?.scope || "Namespaced"}
        </Badge>
      ),
    },
    {
      key: "version",
      header: "Versions",
      render: (crd) => {
        const versions = crd.spec?.versions || [{ name: "v1" }];
        return (
          <div className="flex gap-1">
            {versions.map((v: any) => (
              <span key={v.name} className="px-2 py-0.5 rounded bg-slate-900 border border-white/10 font-mono text-[11px] text-emerald-300">
                {v.name}
              </span>
            ))}
          </div>
        );
      },
    },
    {
      key: "age",
      header: "Age",
      render: (crd) => (
        <span className="text-slate-400 text-xs">
          {formatAge(crd.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (crd) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setCrdToEdit(crd)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Modify CRD Schema"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedCrd(crd)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect CRD"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setCrdToDelete(crd)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete CRD"
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
            <Code size={24} className="text-amber-400" />
            <span>Custom Resource Definitions (CRDs)</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Dynamic schema extensions and custom Kubernetes API resource types registered in the cluster
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button variant="secondary" size="sm" onClick={fetchCrds} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Button size="sm" onClick={() => setIsCreateOpen(true)}>
            <Plus size={14} />
            <span>Create CRD</span>
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={crds}
        searchKey="name"
        searchPlaceholder="Filter CRDs by name..."
        emptyMessage="No CustomResourceDefinitions defined yet."
        onRowClick={(crd) => setSelectedCrd(crd)}
      />

      {/* Create CRD Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Register CustomResourceDefinition"
        maxWidth="md"
      >
        <form onSubmit={handleCreate} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">Resource Kind (Singular)</label>
              <input
                type="text"
                value={crdKind}
                onChange={(e) => {
                  setCrdKind(e.target.value);
                  setCrdPlural(e.target.value.toLowerCase() + "s");
                }}
                placeholder="e.g. MicroService"
                required
                className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">Plural Name</label>
              <input
                type="text"
                value={crdPlural}
                onChange={(e) => setCrdPlural(e.target.value)}
                placeholder="e.g. microservices"
                required
                className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">API Group</label>
              <input
                type="text"
                value={crdGroup}
                onChange={(e) => setCrdGroup(e.target.value)}
                placeholder="e.g. apps.tarak.io"
                required
                className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">Scope</label>
              <select
                value={crdScope}
                onChange={(e) => setCrdScope(e.target.value)}
                className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400"
              >
                <option value="Namespaced">Namespaced</option>
                <option value="Cluster">Cluster-Scoped</option>
              </select>
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setIsCreateOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">Register CRD</Button>
          </div>
        </form>
      </Modal>

      {/* Detail Drawer */}
      <ResourceDetailDrawer
        isOpen={!!selectedCrd}
        onClose={() => setSelectedCrd(null)}
        resourceType="CustomResourceDefinition"
        resourceName={selectedCrd?.metadata?.name || ""}
        rawResource={selectedCrd}
        onActionComplete={fetchCrds}
      />

      {/* Edit Modal */}
      <EditResourceModal
        isOpen={!!crdToEdit}
        onClose={() => setCrdToEdit(null)}
        resourceType="CustomResourceDefinition"
        resourceName={crdToEdit?.metadata?.name || ""}
        rawResource={crdToEdit}
        onSaved={fetchCrds}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={!!crdToDelete}
        onClose={() => setCrdToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete CustomResourceDefinition"
        message={`Are you sure you want to delete CRD "${crdToDelete?.metadata?.name}"? All custom resource instances of this schema will be removed.`}
        confirmText="Delete CRD"
        isLoading={isDeleting}
      />
    </div>
  );
}
