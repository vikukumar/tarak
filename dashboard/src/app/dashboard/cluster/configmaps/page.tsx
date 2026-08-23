"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { FileText, Key, RefreshCw, Plus, FileCode, Trash2, Edit3, Lock, Database } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge, cn } from "@/lib/utils";

export default function ConfigMapsPage() {
  const { selectedNamespace } = useCluster();
  const [activeTab, setActiveTab] = useState<"configmaps" | "secrets">("configmaps");
  const [items, setItems] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedItem, setSelectedItem] = useState<any | null>(null);
  const [itemToEdit, setItemToEdit] = useState<any | null>(null);
  const [itemToDelete, setItemToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchItems = async () => {
    setIsLoading(true);
    try {
      const resourcePath = activeTab === "configmaps" ? "configmaps" : "secrets";
      const url =
        selectedNamespace === "_all"
          ? `/api/v1/${resourcePath}`
          : `/api/v1/namespaces/${selectedNamespace}/${resourcePath}`;
      const res = await tarakFetch(url);
      setItems(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchItems();
  }, [selectedNamespace, activeTab]);

  const confirmDelete = async () => {
    if (!itemToDelete) return;
    setIsDeleting(true);
    try {
      const resourcePath = activeTab === "configmaps" ? "configmaps" : "secrets";
      const ns =
        itemToDelete.metadata?.namespace ||
        (selectedNamespace === "_all" ? "default" : selectedNamespace);
      const name = itemToDelete.metadata?.name;
      await tarakFetch(`/api/v1/namespaces/${ns}/${resourcePath}/${name}`, {
        method: "DELETE",
      });
      setItemToDelete(null);
      fetchItems();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Name",
      sortable: true,
      render: (item) => (
        <div className="flex items-center gap-2.5">
          <div
            className={cn(
              "w-7 h-7 rounded-lg border flex items-center justify-center",
              activeTab === "configmaps"
                ? "bg-amber-500/15 border-amber-500/30 text-amber-400"
                : "bg-purple-500/15 border-purple-500/30 text-purple-400"
            )}
          >
            {activeTab === "configmaps" ? <FileText size={15} /> : <Key size={15} />}
          </div>
          <div>
            <span className="font-bold text-white block">{item.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {item.metadata?.namespace || selectedNamespace}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "dataKeys",
      header: "Keys / Entries",
      render: (item) => {
        const keys = Object.keys(item.data || item.stringData || {});
        return (
          <div className="flex flex-wrap gap-1">
            {keys.length > 0 ? (
              keys.slice(0, 3).map((k) => (
                <span
                  key={k}
                  className="px-2 py-0.5 rounded bg-slate-900 border border-white/10 font-mono text-[11px] text-cyan-300"
                >
                  {k}
                </span>
              ))
            ) : (
              <span className="text-slate-500 font-mono text-xs">0 keys</span>
            )}
            {keys.length > 3 && (
              <span className="text-[10px] text-slate-400 self-center">
                +{keys.length - 3} more
              </span>
            )}
          </div>
        );
      },
    },
    {
      key: "type",
      header: "Type",
      render: (item) => (
        <Badge variant={activeTab === "configmaps" ? "amber" : "purple"}>
          {item.type || (activeTab === "configmaps" ? "ConfigMap" : "Opaque")}
        </Badge>
      ),
    },
    {
      key: "age",
      header: "Age",
      render: (item) => (
        <span className="text-slate-400 text-xs">
          {formatAge(item.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (item) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setItemToEdit(item)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Modify"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedItem(item)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setItemToDelete(item)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete"
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
            <Database size={24} className="text-amber-400" />
            <span>ConfigMaps & Secrets</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Application configuration key-value data and encrypted credential store in{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchItems}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Create {activeTab === "configmaps" ? "ConfigMap" : "Secret"}</span>
            </Button>
          </Link>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-white/10 pb-3">
        <button
          onClick={() => setActiveTab("configmaps")}
          className={cn(
            "px-4 py-2 rounded-xl text-xs font-semibold flex items-center gap-2 transition-all",
            activeTab === "configmaps"
              ? "bg-amber-500/15 border border-amber-500/30 text-amber-300 shadow-md"
              : "bg-slate-900/40 text-slate-400 hover:text-white border border-white/5"
          )}
        >
          <FileText size={14} />
          <span>ConfigMaps</span>
        </button>
        <button
          onClick={() => setActiveTab("secrets")}
          className={cn(
            "px-4 py-2 rounded-xl text-xs font-semibold flex items-center gap-2 transition-all",
            activeTab === "secrets"
              ? "bg-purple-500/15 border border-purple-500/30 text-purple-300 shadow-md"
              : "bg-slate-900/40 text-slate-400 hover:text-white border border-white/5"
          )}
        >
          <Lock size={14} />
          <span>Encrypted Secrets</span>
        </button>
      </div>

      <DataTable
        columns={columns}
        data={items}
        searchKey="name"
        searchPlaceholder={`Filter ${activeTab} by name...`}
        emptyMessage={`No ${activeTab} found in ${
          selectedNamespace === "_all" ? "cluster" : selectedNamespace + " namespace"
        }`}
        onRowClick={(item) => setSelectedItem(item)}
      />

      <ResourceDetailDrawer
        isOpen={!!selectedItem}
        onClose={() => setSelectedItem(null)}
        resourceType={activeTab === "configmaps" ? "ConfigMap" : "Secret"}
        resourceName={selectedItem?.metadata?.name || ""}
        namespace={selectedItem?.metadata?.namespace || selectedNamespace}
        rawResource={selectedItem}
        onActionComplete={fetchItems}
      />

      <EditResourceModal
        isOpen={!!itemToEdit}
        onClose={() => setItemToEdit(null)}
        resourceType={activeTab === "configmaps" ? "ConfigMap" : "Secret"}
        resourceName={itemToEdit?.metadata?.name || ""}
        namespace={itemToEdit?.metadata?.namespace || selectedNamespace}
        rawResource={itemToEdit}
        onSaved={fetchItems}
      />

      <ConfirmModal
        isOpen={!!itemToDelete}
        onClose={() => setItemToDelete(null)}
        onConfirm={confirmDelete}
        title={`Delete ${activeTab === "configmaps" ? "ConfigMap" : "Secret"}`}
        message={`Are you sure you want to delete "${itemToDelete?.metadata?.name}"? Any mounted pod volumes using this key will lose access.`}
        confirmText="Delete Resource"
        isLoading={isDeleting}
      />
    </div>
  );
}
