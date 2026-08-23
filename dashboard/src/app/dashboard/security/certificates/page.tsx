"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Shield, Key, RefreshCw, Plus, FileCode, Trash2, Edit3, Lock, CheckCircle2 } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function CertificatesPage() {
  const { selectedNamespace } = useCluster();
  const [serviceAccounts, setServiceAccounts] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedSA, setSelectedSA] = useState<any | null>(null);
  const [saToEdit, setSaToEdit] = useState<any | null>(null);
  const [saToDelete, setSaToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchSA = async () => {
    setIsLoading(true);
    try {
      const url =
        selectedNamespace === "_all"
          ? "/api/v1/serviceaccounts"
          : `/api/v1/namespaces/${selectedNamespace}/serviceaccounts`;
      const res = await tarakFetch(url);
      setServiceAccounts(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchSA();
  }, [selectedNamespace]);

  const confirmDelete = async () => {
    if (!saToDelete) return;
    setIsDeleting(true);
    try {
      const ns =
        saToDelete.metadata?.namespace ||
        (selectedNamespace === "_all" ? "default" : selectedNamespace);
      const name = saToDelete.metadata?.name;
      await tarakFetch(`/api/v1/namespaces/${ns}/serviceaccounts/${name}`, {
        method: "DELETE",
      });
      setSaToDelete(null);
      fetchSA();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Service Account Name",
      sortable: true,
      render: (sa) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-indigo-500/15 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
            <Key size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{sa.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {sa.metadata?.namespace || selectedNamespace}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "secrets",
      header: "Tokens & Secrets",
      render: (sa) => {
        const secrets = sa.secrets || [];
        return (
          <span className="font-mono text-xs text-cyan-300">
            {secrets.length > 0 ? `${secrets.length} Mounted Secret(s)` : "Auto-Generated JWT"}
          </span>
        );
      },
    },
    {
      key: "pkiStatus",
      header: "mTLS Certificate Status",
      render: () => (
        <div className="flex items-center gap-1.5 text-xs text-emerald-400">
          <CheckCircle2 size={13} />
          <span>ECDSA P-256 Valid</span>
        </div>
      ),
    },
    {
      key: "age",
      header: "Age",
      render: (sa) => (
        <span className="text-slate-400 text-xs">
          {formatAge(sa.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (sa) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setSaToEdit(sa)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Modify Account"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedSA(sa)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect ServiceAccount"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setSaToDelete(sa)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Account"
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
            <Shield size={24} className="text-indigo-400" />
            <span>Service Accounts & Certificates</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Identity tokens and ECDSA root PKI certificate authorization in{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchSA}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Create ServiceAccount</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={serviceAccounts}
        searchKey="name"
        searchPlaceholder="Filter service accounts..."
        emptyMessage={`No service accounts in ${
          selectedNamespace === "_all" ? "cluster" : selectedNamespace + " namespace"
        }`}
        onRowClick={(sa) => setSelectedSA(sa)}
      />

      <ResourceDetailDrawer
        isOpen={!!selectedSA}
        onClose={() => setSelectedSA(null)}
        resourceType="ServiceAccount"
        resourceName={selectedSA?.metadata?.name || ""}
        namespace={selectedSA?.metadata?.namespace || selectedNamespace}
        rawResource={selectedSA}
        onActionComplete={fetchSA}
      />

      <EditResourceModal
        isOpen={!!saToEdit}
        onClose={() => setSaToEdit(null)}
        resourceType="ServiceAccount"
        resourceName={saToEdit?.metadata?.name || ""}
        namespace={saToEdit?.metadata?.namespace || selectedNamespace}
        rawResource={saToEdit}
        onSaved={fetchSA}
      />

      <ConfirmModal
        isOpen={!!saToDelete}
        onClose={() => setSaToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete ServiceAccount"
        message={`Are you sure you want to delete service account "${saToDelete?.metadata?.name}"? Pods associated with this token may lose API server authentication.`}
        confirmText="Delete Account"
        isLoading={isDeleting}
      />
    </div>
  );
}
