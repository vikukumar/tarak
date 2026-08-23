"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Shield, RefreshCw, Plus, FileCode, Trash2, Edit3, Network, ArrowRight } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { ConfirmModal } from "@/components/ui/ConfirmModal";
import { EditResourceModal } from "@/components/modals/EditResourceModal";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function NetworkPoliciesPage() {
  const { selectedNamespace } = useCluster();
  const [policies, setPolicies] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedPolicy, setSelectedPolicy] = useState<any | null>(null);
  const [policyToEdit, setPolicyToEdit] = useState<any | null>(null);
  const [policyToDelete, setPolicyToDelete] = useState<any | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const fetchPolicies = async () => {
    setIsLoading(true);
    try {
      const url =
        selectedNamespace === "_all"
          ? "/apis/networking.k8s.io/v1/networkpolicies"
          : `/apis/networking.k8s.io/v1/namespaces/${selectedNamespace}/networkpolicies`;
      const res = await tarakFetch(url);
      setPolicies(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchPolicies();
  }, [selectedNamespace]);

  const confirmDelete = async () => {
    if (!policyToDelete) return;
    setIsDeleting(true);
    try {
      const ns =
        policyToDelete.metadata?.namespace ||
        (selectedNamespace === "_all" ? "default" : selectedNamespace);
      const name = policyToDelete.metadata?.name;
      await tarakFetch(`/apis/networking.k8s.io/v1/namespaces/${ns}/networkpolicies/${name}`, {
        method: "DELETE",
      });
      setPolicyToDelete(null);
      fetchPolicies();
    } finally {
      setIsDeleting(false);
    }
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Policy Name",
      sortable: true,
      render: (pol) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
            <Shield size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{pol.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {pol.metadata?.namespace || selectedNamespace}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "podSelector",
      header: "Target Pods",
      render: (pol) => {
        const matchLabels = pol.spec?.podSelector?.matchLabels || {};
        const entries = Object.entries(matchLabels);
        if (entries.length === 0) {
          return <Badge variant="slate">All Pods in NS</Badge>;
        }
        return (
          <div className="flex flex-wrap gap-1">
            {entries.map(([k, v]) => (
              <span
                key={k}
                className="px-2 py-0.5 rounded bg-slate-900 border border-white/10 font-mono text-[11px] text-cyan-300"
              >
                {k}={String(v)}
              </span>
            ))}
          </div>
        );
      },
    },
    {
      key: "policyTypes",
      header: "Types",
      render: (pol) => {
        const types = pol.spec?.policyTypes || ["Ingress"];
        return (
          <div className="flex gap-1">
            {types.map((t: string) => (
              <Badge key={t} variant={t === "Ingress" ? "cyan" : "indigo"}>
                {t}
              </Badge>
            ))}
          </div>
        );
      },
    },
    {
      key: "age",
      header: "Age",
      render: (pol) => (
        <span className="text-slate-400 text-xs">
          {formatAge(pol.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (pol) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setPolicyToEdit(pol)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 transition-colors"
            title="Modify Policy"
          >
            <Edit3 size={14} />
          </button>
          <button
            onClick={() => setSelectedPolicy(pol)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Policy"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => setPolicyToDelete(pol)}
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
            <Network size={24} className="text-emerald-400" />
            <span>Network Policies & Firewall</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            L3/L4 ingress and egress packet filtering and pod traffic isolation in{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchPolicies}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Create Network Policy</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={policies}
        searchKey="name"
        searchPlaceholder="Filter policies by name..."
        emptyMessage={`No network policies defined in ${
          selectedNamespace === "_all" ? "cluster" : selectedNamespace + " namespace"
        }`}
        onRowClick={(pol) => setSelectedPolicy(pol)}
      />

      <ResourceDetailDrawer
        isOpen={!!selectedPolicy}
        onClose={() => setSelectedPolicy(null)}
        resourceType="NetworkPolicy"
        resourceName={selectedPolicy?.metadata?.name || ""}
        namespace={selectedPolicy?.metadata?.namespace || selectedNamespace}
        rawResource={selectedPolicy}
        onActionComplete={fetchPolicies}
      />

      <EditResourceModal
        isOpen={!!policyToEdit}
        onClose={() => setPolicyToEdit(null)}
        resourceType="NetworkPolicy"
        resourceName={policyToEdit?.metadata?.name || ""}
        namespace={policyToEdit?.metadata?.namespace || selectedNamespace}
        rawResource={policyToEdit}
        onSaved={fetchPolicies}
      />

      <ConfirmModal
        isOpen={!!policyToDelete}
        onClose={() => setPolicyToDelete(null)}
        onConfirm={confirmDelete}
        title="Delete Network Policy"
        message={`Are you sure you want to delete policy "${policyToDelete?.metadata?.name}"? Traffic restrictions governed by this rule will be lifted.`}
        confirmText="Delete Policy"
        isLoading={isDeleting}
      />
    </div>
  );
}
