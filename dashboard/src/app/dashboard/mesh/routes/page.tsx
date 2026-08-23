"use client";

import React, { useState, useEffect } from "react";
import { Workflow, RefreshCw, Plus, FileCode, Trash2, Edit3, GitFork, ArrowRight } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function MeshRoutesPage() {
  const [routes, setRoutes] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [routeName, setRouteName] = useState("");
  const [targetService, setTargetService] = useState("web-app-svc");
  const [weightV1, setWeightV1] = useState(80);
  const [weightV2, setWeightV2] = useState(20);
  const [selectedItem, setSelectedItem] = useState<any | null>(null);
  const [itemToEdit, setItemToEdit] = useState<any | null>(null);
  const [itemToDelete, setItemToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchRoutes = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/mesh.tarak.io/v1/routes");
      setRoutes(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchRoutes();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!routeName.trim()) return;

    await tarakFetch("/apis/mesh.tarak.io/v1/routes", {
      method: "POST",
      body: JSON.stringify({
        name: routeName.trim().toLowerCase(),
        service: targetService.trim(),
        splits: [
          { subset: "v1", weight: weightV1 },
          { subset: "v2", weight: weightV2 },
        ],
      }),
    });
    setIsCreateOpen(false);
    setRouteName("");
    fetchRoutes();
  };

  const confirmDelete = async () => {
    if (!itemToDelete) return;
    setIsDeleting(true);
    try {
      const ns = itemToDelete.namespace || "default";
      const name = itemToDelete.name || itemToDelete.metadata?.name;
      await tarakFetch(`/apis/mesh.tarak.io/v1/namespaces/${ns}/routes/${name}`, {
        method: "DELETE",
      });
      setItemToDelete(null);
      fetchRoutes();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Route Name",
      sortable: true,
      render: (r) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-indigo-500/15 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
            <GitFork size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{r.name || r.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              svc: {r.service || "web-app-svc"}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "splits",
      header: "Canary Traffic Split",
      render: (r) => {
        const splits = r.splits || [{ subset: "v1", weight: 90 }, { subset: "v2", weight: 10 }];
        return (
          <div className="flex items-center gap-2">
            {splits.map((s: any) => (
              <Badge key={s.subset} variant={s.subset === "v1" ? "cyan" : "purple"}>
                {s.subset}: {s.weight}%
              </Badge>
            ))}
          </div>
        );
      },
    },
    {
      key: "status",
      header: "Status",
      render: () => <Badge variant="emerald" dot>Active in Mesh</Badge>,
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (r) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setItemToEdit(r)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Modify Route"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedItem(r)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Route"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setItemToDelete(r)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Route"
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
            <span>Canary & Traffic Routes</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Dynamic weight-based traffic routing and zero-downtime A/B testing inside the mesh
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button variant="secondary" size="sm" onClick={fetchRoutes} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Button size="sm" onClick={() => setIsCreateOpen(true)}>
            <Plus size={14} />
            <span>Add Route Split</span>
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={
          routes.length > 0
            ? routes
            : [
                {
                  name: "web-app-canary",
                  service: "web-app-svc",
                  splits: [
                    { subset: "v1 (stable)", weight: 90 },
                    { subset: "v2 (canary)", weight: 10 },
                  ],
                },
              ]
        }
        searchKey="name"
        onRowClick={(r) => setSelectedItem(r)}
      />

      {/* Create Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Create Canary Traffic Route"
        maxWidth="sm"
      >
        <form onSubmit={handleCreate} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Route Name</label>
            <input
              type="text"
              value={routeName}
              onChange={(e) => setRouteName(e.target.value)}
              placeholder="e.g. payment-api-canary"
              required
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Target Service Name</label>
            <input
              type="text"
              value={targetService}
              onChange={(e) => setTargetService(e.target.value)}
              placeholder="e.g. payment-service"
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">Stable v1 Weight (%)</label>
              <input
                type="number"
                min="0"
                max="100"
                value={weightV1}
                onChange={(e) => {
                  const v = Number(e.target.value);
                  setWeightV1(v);
                  setWeightV2(100 - v);
                }}
                className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2 text-sm text-cyan-300 font-mono"
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">Canary v2 Weight (%)</label>
              <input
                type="number"
                min="0"
                max="100"
                value={weightV2}
                onChange={(e) => {
                  const v = Number(e.target.value);
                  setWeightV2(v);
                  setWeightV1(100 - v);
                }}
                className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2 text-sm text-purple-300 font-mono"
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setIsCreateOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">Save Canary Route</Button>
          </div>
        </form>
      </Modal>

      {/* Detail Drawer */}
      <ResourceDetailDrawer
        isOpen={!!selectedItem}
        onClose={() => setSelectedItem(null)}
        resourceType="TrafficRoute"
        resourceName={selectedItem?.name || selectedItem?.metadata?.name || ""}
        rawResource={selectedItem}
        onActionComplete={fetchRoutes}
      />

      {/* Edit Modal */}
      <EditResourceModal
        isOpen={!!itemToEdit}
        onClose={() => setItemToEdit(null)}
        resourceType="TrafficRoute"
        resourceName={itemToEdit?.name || itemToEdit?.metadata?.name || ""}
        rawResource={itemToEdit}
        onSaved={fetchRoutes}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={!!itemToDelete}
        onClose={() => setItemToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete Traffic Route"
        message={`Are you sure you want to delete traffic route "${itemToDelete?.name || itemToDelete?.metadata?.name}"? Traffic will revert to 100% stable routing.`}
        confirmText="Delete Route"
        isLoading={isDeleting}
      />
    </div>
  );
}
