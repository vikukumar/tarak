"use client";

import React, { useState, useEffect } from "react";
import { Lock, Shield, RefreshCw, Plus, FileCode, Trash2, Edit3 } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { Modal } from "@/components/ui/Modal";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function MeshPermissionsPage() {
  const [items, setItems] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [permName, setPermName] = useState("");
  const [sourceService, setSourceService] = useState("*");
  const [destService, setDestService] = useState("*");
  const [selectedItem, setSelectedItem] = useState<any | null>(null);
  const [itemToEdit, setItemToEdit] = useState<any | null>(null);
  const [itemToDelete, setItemToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchPermissions = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/mesh.tarak.io/v1/meshes/default/traffic-permissions");
      setItems(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchPermissions();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!permName.trim()) return;

    await tarakFetch("/apis/mesh.tarak.io/v1/meshes/default/traffic-permissions", {
      method: "POST",
      body: JSON.stringify({
        name: permName.trim().toLowerCase(),
        sources: [sourceService.trim()],
        destinations: [destService.trim()],
        action: "ALLOW",
      }),
    });
    setIsCreateOpen(false);
    setPermName("");
    fetchPermissions();
  };

  const confirmDelete = async () => {
    if (!itemToDelete) return;
    setIsDeleting(true);
    try {
      const name = itemToDelete.name || itemToDelete.metadata?.name;
      await tarakFetch(`/apis/mesh.tarak.io/v1/meshes/default/traffic-permissions/${name}`, {
        method: "DELETE",
      });
      setItemToDelete(null);
      fetchPermissions();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Permission Policy",
      sortable: true,
      render: (p) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-purple-500/15 border border-purple-500/30 flex items-center justify-center text-purple-400">
            <Lock size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{p.name || p.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              mesh: {p.mesh || "default"}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "sources",
      header: "Allowed Sources",
      render: (p) => (
        <Badge variant="cyan">{p.sources?.join(", ") || "service: *"}</Badge>
      ),
    },
    {
      key: "destinations",
      header: "Allowed Destinations",
      render: (p) => (
        <Badge variant="purple">{p.destinations?.join(", ") || "service: *"}</Badge>
      ),
    },
    {
      key: "action",
      header: "Enforcement",
      render: () => <Badge variant="emerald">ALLOW (mTLS Verified)</Badge>,
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
          <button
            onClick={() => setItemToEdit(p)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Modify Permission"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedItem(p)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Permission"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setItemToDelete(p)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Permission"
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
            <Lock size={24} className="text-purple-400" />
            <span>mTLS Traffic Permissions</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Zero-Trust access control rules regulating service-to-service communication inside the mesh
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button variant="secondary" size="sm" onClick={fetchPermissions} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Button size="sm" onClick={() => setIsCreateOpen(true)}>
            <Plus size={14} />
            <span>Add Permission</span>
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={items.length > 0 ? items : [{ name: "default-mesh-allow-all", sources: ["*"], destinations: ["*"], mesh: "default" }]}
        searchKey="name"
        onRowClick={(p) => setSelectedItem(p)}
      />

      {/* Create Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Add Traffic Permission Policy"
        maxWidth="sm"
      >
        <form onSubmit={handleCreate} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Policy Name</label>
            <input
              type="text"
              value={permName}
              onChange={(e) => setPermName(e.target.value)}
              placeholder="e.g. allow-frontend-to-backend"
              required
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Source Workload Selector</label>
            <input
              type="text"
              value={sourceService}
              onChange={(e) => setSourceService(e.target.value)}
              placeholder="service: frontend or *"
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Destination Service Selector</label>
            <input
              type="text"
              value={destService}
              onChange={(e) => setDestService(e.target.value)}
              placeholder="service: backend-api or *"
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setIsCreateOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">Save Policy</Button>
          </div>
        </form>
      </Modal>

      {/* Resource Detail Drawer */}
      <ResourceDetailDrawer
        isOpen={!!selectedItem}
        onClose={() => setSelectedItem(null)}
        resourceType="TrafficPermission"
        resourceName={selectedItem?.name || selectedItem?.metadata?.name || ""}
        rawResource={selectedItem}
        onActionComplete={fetchPermissions}
      />

      {/* Edit Resource Modal */}
      <EditResourceModal
        isOpen={!!itemToEdit}
        onClose={() => setItemToEdit(null)}
        resourceType="TrafficPermission"
        resourceName={itemToEdit?.name || itemToEdit?.metadata?.name || ""}
        rawResource={itemToEdit}
        onSaved={fetchPermissions}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={!!itemToDelete}
        onClose={() => setItemToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete Traffic Permission"
        message={`Are you sure you want to delete permission rule "${itemToDelete?.name || itemToDelete?.metadata?.name}"? Workloads governed by this rule will be affected.`}
        confirmText="Delete Permission"
        isLoading={isDeleting}
      />
    </div>
  );
}
