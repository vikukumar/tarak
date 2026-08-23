"use client";

import React, { useState, useEffect } from "react";
import { Globe, Plus, RefreshCw, Layers, CheckCircle2, Box } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { NamespaceDetailDrawer } from "@/components/drawers/NamespaceDetailDrawer";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function NamespacesPage() {
  const { selectedNamespace, setSelectedNamespace, refresh } = useCluster();
  const [namespaces, setNamespaces] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newNsName, setNewNsName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [selectedNsForDrawer, setSelectedNsForDrawer] = useState<string | null>(null);

  const fetchNamespaces = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/api/v1/namespaces");
      setNamespaces(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchNamespaces();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newNsName.trim()) return;
    setError(null);

    const nsObj = {
      apiVersion: "v1",
      kind: "Namespace",
      metadata: {
        name: newNsName.trim().toLowerCase(),
      },
    };

    const res = await tarakFetch("/api/v1/namespaces", {
      method: "POST",
      body: JSON.stringify(nsObj),
    });

    if (res.error) {
      setError(res.error);
    } else {
      setIsModalOpen(false);
      setNewNsName("");
      await fetchNamespaces();
      await refresh();
      setSelectedNamespace(newNsName.trim().toLowerCase());
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Namespace",
      sortable: true,
      render: (ns) => {
        const name = ns.metadata?.name;
        const isCurrent = selectedNamespace === name;
        return (
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
              <Globe size={16} />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-bold text-white text-sm">{name}</span>
                {isCurrent && (
                  <span className="text-[10px] uppercase font-bold text-cyan-400 bg-cyan-500/10 px-1.5 py-0.5 rounded border border-cyan-500/20">
                    Active Scope
                  </span>
                )}
              </div>
              <span className="text-[11px] text-slate-400 font-mono">
                uid: {ns.metadata?.uid?.slice(0, 12) || "system-core"}
              </span>
            </div>
          </div>
        );
      },
    },
    {
      key: "status",
      header: "Status",
      render: () => <Badge variant="emerald" dot>Active</Badge>,
    },
    {
      key: "age",
      header: "Age",
      render: (ns) => (
        <span className="text-slate-400">
          {formatAge(ns.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "action",
      header: "Quick Scope",
      className: "text-right",
      render: (ns) => {
        const name = ns.metadata?.name;
        const isCurrent = selectedNamespace === name;
        return (
          <div className="flex items-center justify-end" onClick={(e) => e.stopPropagation()}>
            <button
              onClick={() => setSelectedNamespace(name)}
              className={`px-3 py-1 rounded-lg text-xs font-semibold border transition-all ${
                isCurrent
                  ? "bg-cyan-500/20 border-cyan-500/40 text-cyan-300 shadow-sm"
                  : "bg-slate-900/80 hover:bg-slate-800 border-white/10 text-slate-300 hover:text-white"
              }`}
            >
              {isCurrent ? "Active" : "Switch To"}
            </button>
          </div>
        );
      },
    },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <Globe size={24} className="text-cyan-400" />
            <span>Namespaces</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Isolated multi-tenant logical clusters. Click any row to inspect workloads in that namespace.
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchNamespaces}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Button size="sm" onClick={() => setIsModalOpen(true)}>
            <Plus size={14} />
            <span>Create Namespace</span>
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={namespaces}
        searchKey="name"
        searchPlaceholder="Filter namespaces..."
        emptyMessage="No namespaces found in cluster"
        onRowClick={(ns) => setSelectedNsForDrawer(ns.metadata?.name)}
      />

      {/* Create Namespace Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Create Logical Namespace"
        maxWidth="sm"
      >
        <form onSubmit={handleCreate} className="space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-xs text-rose-300">
              {error}
            </div>
          )}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">
              Namespace Identifier
            </label>
            <input
              type="text"
              value={newNsName}
              onChange={(e) => setNewNsName(e.target.value)}
              placeholder="e.g. production, staging, microservices"
              required
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <Button type="submit" className="w-full">
            Provision Namespace
          </Button>
        </form>
      </Modal>

      {/* Namespace Inspector Drawer */}
      <NamespaceDetailDrawer
        isOpen={!!selectedNsForDrawer}
        onClose={() => setSelectedNsForDrawer(null)}
        namespaceName={selectedNsForDrawer || ""}
      />
    </div>
  );
}
