"use client";

import React, { useState, useEffect } from "react";
import { Globe, Plus, RefreshCw } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function NamespacesPage() {
  const [namespaces, setNamespaces] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newNsName, setNewNsName] = useState("");
  const [error, setError] = useState<string | null>(null);

  const fetchNamespaces = async () => {
    setIsLoading(true);
    const res = await tarakFetch("/api/v1/namespaces");
    setNamespaces(res.data?.items || []);
    setIsLoading(false);
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
      fetchNamespaces();
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Namespace",
      sortable: true,
      render: (ns) => (
        <div className="flex items-center gap-2">
          <Globe size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{ns.metadata?.name}</span>
        </div>
      ),
    },
    {
      key: "status",
      header: "Status",
      render: () => <Badge variant="emerald" dot>Active</Badge>,
    },
    {
      key: "age",
      header: "Age",
      render: (ns) => formatAge(ns.metadata?.creationTimestamp),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Globe size={22} className="text-cyan-400" />
            <span>Namespaces</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Isolated multi-tenant logical clusters for workload partitioning
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={fetchNamespaces} isLoading={isLoading}>
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
        emptyMessage="No namespaces found"
      />

      {/* Create Namespace Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="Create New Namespace"
        maxWidth="sm"
      >
        <form onSubmit={handleCreate} className="space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-xs text-rose-300">
              {error}
            </div>
          )}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Namespace Name</label>
            <input
              type="text"
              value={newNsName}
              onChange={(e) => setNewNsName(e.target.value)}
              placeholder="e.g. production"
              required
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <Button type="submit" className="w-full">
            Create Namespace
          </Button>
        </form>
      </Modal>
    </div>
  );
}
