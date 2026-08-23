"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Globe, RefreshCw, Plus, ExternalLink, FileCode, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function IngressPage() {
  const { selectedNamespace } = useCluster();
  const [ingresses, setIngresses] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedIngress, setSelectedIngress] = useState<any | null>(null);

  const fetchIngresses = async () => {
    setIsLoading(true);
    try {
      const url =
        selectedNamespace === "_all"
          ? "/apis/networking.k8s.io/v1/ingresses"
          : `/apis/networking.k8s.io/v1/namespaces/${selectedNamespace}/ingresses`;
      const res = await tarakFetch(url);
      setIngresses(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchIngresses();
  }, [selectedNamespace]);

  const handleDelete = async (ing: any) => {
    const ns = ing.metadata?.namespace || selectedNamespace;
    const name = ing.metadata?.name;
    if (!confirm(`Delete ingress "${name}" in namespace "${ns}"?`)) return;
    await tarakFetch(`/apis/networking.k8s.io/v1/namespaces/${ns}/ingresses/${name}`, {
      method: "DELETE",
    });
    fetchIngresses();
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Ingress Name",
      sortable: true,
      render: (ing) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
            <Globe size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{ing.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {ing.metadata?.namespace || selectedNamespace}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "class",
      header: "IngressClass",
      render: (ing) => (
        <Badge variant="indigo">{ing.spec?.ingressClassName || "tarak"}</Badge>
      ),
    },
    {
      key: "hosts",
      header: "Virtual Hosts & Routes",
      render: (ing) => {
        const rules = ing.spec?.rules || [];
        if (rules.length === 0) return <span className="text-slate-500 font-mono">*</span>;
        return (
          <div className="flex flex-wrap gap-1.5" onClick={(e) => e.stopPropagation()}>
            {rules.map((r: any, i: number) => (
              <a
                key={i}
                href={`http://${r.host || "localhost:8080"}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-cyan-500/10 text-cyan-300 font-mono text-[11px] hover:bg-cyan-500/20 border border-cyan-500/20"
              >
                <span>{r.host || "*"}</span>
                <ExternalLink size={10} />
              </a>
            ))}
          </div>
        );
      },
    },
    {
      key: "age",
      header: "Age",
      render: (ing) => (
        <span className="text-slate-400 text-xs">
          {formatAge(ing.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (ing) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setSelectedIngress(ing)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Ingress"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => handleDelete(ing)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Ingress"
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
            <Globe size={24} className="text-cyan-400" />
            <span>Ingress Routes & Domains</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            HTTP/HTTPS virtual host routing and reverse proxies in scope{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchIngresses}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Create Ingress</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={ingresses}
        searchKey="name"
        searchPlaceholder="Filter ingress routes..."
        emptyMessage={`No ingress routes defined in ${
          selectedNamespace === "_all" ? "cluster" : selectedNamespace + " namespace"
        }`}
        onRowClick={(ing) => setSelectedIngress(ing)}
      />

      <ResourceDetailDrawer
        isOpen={!!selectedIngress}
        onClose={() => setSelectedIngress(null)}
        resourceType="Ingress"
        resourceName={selectedIngress?.metadata?.name || ""}
        namespace={selectedIngress?.metadata?.namespace || selectedNamespace}
        rawResource={selectedIngress}
        onActionComplete={fetchIngresses}
      />
    </div>
  );
}
