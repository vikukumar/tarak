"use client";

import React, { useState, useEffect } from "react";
import { Radio, Shield, Network, RefreshCw, Plus, Lock, FileCode, Trash2, Edit3 } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function MeshOverviewPage() {
  const [meshes, setMeshes] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newMeshName, setNewMeshName] = useState("");
  const [mtlsMode, setMtlsMode] = useState("Strict");
  const [selectedMesh, setSelectedMesh] = useState<any | null>(null);
  const [meshToEdit, setMeshToEdit] = useState<any | null>(null);
  const [meshToDelete, setMeshToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchMeshes = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/mesh.tarak.io/v1/meshes");
      setMeshes(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMeshes();
  }, []);

  const handleCreateMesh = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMeshName.trim()) return;

    await tarakFetch("/apis/mesh.tarak.io/v1/meshes", {
      method: "POST",
      body: JSON.stringify({
        name: newMeshName.trim().toLowerCase(),
        mtls: {
          mode: mtlsMode,
          trustDomain: `${newMeshName.trim().toLowerCase()}.tarak.mesh`,
        },
      }),
    });
    setIsCreateOpen(false);
    setNewMeshName("");
    fetchMeshes();
  };

  const confirmDelete = async () => {
    if (!meshToDelete) return;
    setIsDeleting(true);
    try {
      const name = meshToDelete.name || meshToDelete.metadata?.name;
      await tarakFetch(`/apis/mesh.tarak.io/v1/meshes/${name}`, {
        method: "DELETE",
      });
      setMeshToDelete(null);
      fetchMeshes();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Mesh Tenant",
      sortable: true,
      render: (m) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-purple-500/15 border border-purple-500/30 flex items-center justify-center text-purple-400">
            <Radio size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{m.name || m.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              domain: {m.mtls?.trustDomain || `${m.name || "default"}.tarak.mesh`}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "status",
      header: "Phase",
      render: () => <Badge variant="emerald" dot>Active</Badge>,
    },
    {
      key: "mtls",
      header: "mTLS Zero-Trust",
      render: (m) => (
        <Badge variant={m.mtls?.mode === "Strict" ? "purple" : "cyan"}>
          {m.mtls?.mode || "Strict"} Mode
        </Badge>
      ),
    },
    {
      key: "passthrough",
      header: "Egress Policy",
      render: () => <Badge variant="emerald">Passthrough Allowed</Badge>,
    },
    {
      key: "age",
      header: "Age",
      render: (m) => (
        <span className="text-slate-400 text-xs">
          {formatAge(m.createdAt || m.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (m) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setMeshToEdit(m)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Modify Mesh"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedMesh(m)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Mesh"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setMeshToDelete(m)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Mesh"
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
            <Radio size={24} className="text-purple-400" />
            <span>Multi-Mesh Control Plane</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Universal Service Mesh control plane with zero-trust mTLS encryption and traffic routing
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button variant="secondary" size="sm" onClick={fetchMeshes} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Button size="sm" onClick={() => setIsCreateOpen(true)}>
            <Plus size={14} />
            <span>Create Mesh</span>
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={meshes}
        searchKey="name"
        emptyMessage="No meshes defined in cluster"
        onRowClick={(m) => setSelectedMesh(m)}
      />

      {/* Create Mesh Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Create Isolated Service Mesh"
        maxWidth="sm"
      >
        <form onSubmit={handleCreateMesh} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Mesh Tenant Name</label>
            <input
              type="text"
              value={newMeshName}
              onChange={(e) => setNewMeshName(e.target.value)}
              placeholder="e.g. production-mesh"
              required
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">mTLS Security Mode</label>
            <select
              value={mtlsMode}
              onChange={(e) => setMtlsMode(e.target.value)}
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400"
            >
              <option value="Strict">Strict (Enforce ECDSA mTLS Certificate)</option>
              <option value="Permissive">Permissive (Accept Plaintext + mTLS)</option>
              <option value="Disabled">Disabled</option>
            </select>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setIsCreateOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">Create Mesh</Button>
          </div>
        </form>
      </Modal>

      {/* Resource Detail Drawer */}
      <ResourceDetailDrawer
        isOpen={!!selectedMesh}
        onClose={() => setSelectedMesh(null)}
        resourceType="Mesh"
        resourceName={selectedMesh?.name || selectedMesh?.metadata?.name || ""}
        rawResource={selectedMesh}
        onActionComplete={fetchMeshes}
      />

      {/* Edit Resource Modal */}
      <EditResourceModal
        isOpen={!!meshToEdit}
        onClose={() => setMeshToEdit(null)}
        resourceType="Mesh"
        resourceName={meshToEdit?.name || meshToEdit?.metadata?.name || ""}
        rawResource={meshToEdit}
        onSaved={fetchMeshes}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={!!meshToDelete}
        onClose={() => setMeshToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete Mesh Tenant"
        message={`Are you sure you want to delete mesh "${meshToDelete?.name || meshToDelete?.metadata?.name}"? Connected workloads will lose mesh routing policies.`}
        confirmText="Delete Mesh"
        isLoading={isDeleting}
      />
    </div>
  );
}
