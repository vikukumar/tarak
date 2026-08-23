"use client";

import React, { useState, useEffect } from "react";
import { Network, RefreshCw, Plus, FileCode, Trash2, Edit3, Shield, Globe } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function MeshPassthroughPage() {
  const [policies, setPolicies] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [policyName, setPolicyName] = useState("");
  const [endpointMatch, setEndpointMatch] = useState("*.googleapis.com");
  const [selectedItem, setSelectedItem] = useState<any | null>(null);
  const [itemToEdit, setItemToEdit] = useState<any | null>(null);
  const [itemToDelete, setItemToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchPolicies = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/mesh.tarak.io/v1/meshes/default/passthrough-policies");
      setPolicies(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchPolicies();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!policyName.trim()) return;

    await tarakFetch("/apis/mesh.tarak.io/v1/meshes/default/passthrough-policies", {
      method: "POST",
      body: JSON.stringify({
        name: policyName.trim().toLowerCase(),
        match: endpointMatch.trim(),
        tlsOrigination: "Auto",
        status: "Enabled",
      }),
    });
    setIsCreateOpen(false);
    setPolicyName("");
    fetchPolicies();
  };

  const confirmDelete = async () => {
    if (!itemToDelete) return;
    setIsDeleting(true);
    try {
      const name = itemToDelete.name || itemToDelete.metadata?.name;
      await tarakFetch(`/apis/mesh.tarak.io/v1/meshes/default/passthrough-policies/${name}`, {
        method: "DELETE",
      });
      setItemToDelete(null);
      fetchPolicies();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Egress Policy",
      sortable: true,
      render: (p) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
            <Globe size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{p.name || p.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              target: {p.match || "*"}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "status",
      header: "Egress Status",
      render: () => <Badge variant="emerald" dot>Allowed (TLS 1.3)</Badge>,
    },
    {
      key: "tls",
      header: "TLS Origination",
      render: (p) => (
        <Badge variant="purple">{p.tlsOrigination || "Auto-Encrypt"}</Badge>
      ),
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
            title="Modify Policy"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedItem(p)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Policy"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setItemToDelete(p)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Policy"
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
            <Network size={24} className="text-purple-400" />
            <span>Egress Passthrough Policies</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Configure whether outbound mesh requests are permitted to external endpoints or restricted
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button variant="secondary" size="sm" onClick={fetchPolicies} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Button size="sm" onClick={() => setIsCreateOpen(true)}>
            <Plus size={14} />
            <span>Add Policy</span>
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="p-5 space-y-3 border-cyan-500/20 bg-gradient-to-b from-cyan-950/20 to-slate-950/40">
          <div className="flex items-center justify-between">
            <h3 className="font-bold text-white text-sm">Cluster Egress Mode</h3>
            <Badge variant="emerald" dot>Passthrough Enabled</Badge>
          </div>
          <p className="text-xs text-slate-300 leading-relaxed">
            Workloads in the default mesh are permitted to communicate with external endpoints and third-party APIs with zero-trust validation.
          </p>
        </Card>

        <Card className="p-5 space-y-3 border-purple-500/20 bg-gradient-to-b from-purple-950/20 to-slate-950/40">
          <div className="flex items-center justify-between">
            <h3 className="font-bold text-white text-sm">Egress TLS Encryption</h3>
            <Badge variant="purple">Auto-Origination</Badge>
          </div>
          <p className="text-xs text-slate-300 leading-relaxed">
            Outbound traffic to external databases, SaaS APIs, and payment gateways is automatically upgraded to TLS 1.3.
          </p>
        </Card>
      </div>

      <DataTable
        columns={columns}
        data={
          policies.length > 0
            ? policies
            : [
                {
                  name: "default-egress-passthrough",
                  match: "* (All external endpoints)",
                  tlsOrigination: "Auto (TLS 1.3)",
                },
              ]
        }
        searchKey="name"
        onRowClick={(p) => setSelectedItem(p)}
      />

      {/* Create Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Add Egress Passthrough Rule"
        maxWidth="sm"
      >
        <form onSubmit={handleCreate} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Policy Name</label>
            <input
              type="text"
              value={policyName}
              onChange={(e) => setPolicyName(e.target.value)}
              placeholder="e.g. allow-openai-api"
              required
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Target Domain / IP Match</label>
            <input
              type="text"
              value={endpointMatch}
              onChange={(e) => setEndpointMatch(e.target.value)}
              placeholder="e.g. api.openai.com or *.amazonaws.com"
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="secondary" onClick={() => setIsCreateOpen(false)}>
              Cancel
            </Button>
            <Button type="submit">Save Egress Rule</Button>
          </div>
        </form>
      </Modal>

      {/* Detail Drawer */}
      <ResourceDetailDrawer
        isOpen={!!selectedItem}
        onClose={() => setSelectedItem(null)}
        resourceType="PassthroughPolicy"
        resourceName={selectedItem?.name || selectedItem?.metadata?.name || ""}
        rawResource={selectedItem}
        onActionComplete={fetchPolicies}
      />

      {/* Edit Modal */}
      <EditResourceModal
        isOpen={!!itemToEdit}
        onClose={() => setItemToEdit(null)}
        resourceType="PassthroughPolicy"
        resourceName={itemToEdit?.name || itemToEdit?.metadata?.name || ""}
        rawResource={itemToEdit}
        onSaved={fetchPolicies}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmModal
        isOpen={!!itemToDelete}
        onClose={() => setItemToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete Passthrough Policy"
        message={`Are you sure you want to delete policy "${itemToDelete?.name || itemToDelete?.metadata?.name}"? External requests to matching endpoints may be blocked.`}
        confirmText="Delete Policy"
        isLoading={isDeleting}
      />
    </div>
  );
}
