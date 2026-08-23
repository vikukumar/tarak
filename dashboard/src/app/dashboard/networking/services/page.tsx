"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Server, RefreshCw, Plus, FileCode, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { ResourceDetailDrawer } from "@/components/drawers/ResourceDetailDrawer";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function ServicesPage() {
  const { selectedNamespace } = useCluster();
  const [services, setServices] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedService, setSelectedService] = useState<any | null>(null);

  const fetchServices = async () => {
    setIsLoading(true);
    try {
      const url =
        selectedNamespace === "_all"
          ? "/api/v1/services"
          : `/api/v1/namespaces/${selectedNamespace}/services`;
      const res = await tarakFetch(url);
      setServices(res.data?.items || []);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchServices();
  }, [selectedNamespace]);

  const handleDelete = async (svc: any) => {
    const ns = svc.metadata?.namespace || selectedNamespace;
    const name = svc.metadata?.name;
    if (!confirm(`Delete service "${name}" in namespace "${ns}"?`)) return;
    await tarakFetch(`/api/v1/namespaces/${ns}/services/${name}`, {
      method: "DELETE",
    });
    fetchServices();
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Service Name",
      sortable: true,
      render: (svc) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-indigo-500/15 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
            <Server size={15} />
          </div>
          <div>
            <span className="font-bold text-white block">{svc.metadata?.name}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              ns: {svc.metadata?.namespace || selectedNamespace}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "type",
      header: "Type",
      render: (svc) => {
        const t = svc.spec?.type || "ClusterIP";
        const isLB = t === "LoadBalancer";
        return <Badge variant={isLB ? "emerald" : "indigo"}>{t}</Badge>;
      },
    },
    {
      key: "clusterIP",
      header: "Cluster IP",
      render: (svc) => (
        <span className="font-mono text-xs text-slate-300">
          {svc.spec?.clusterIP || "<none>"}
        </span>
      ),
    },
    {
      key: "externalIP",
      header: "External IP (MetalLB)",
      render: (svc) => {
        const lbIngress = svc.status?.loadBalancer?.ingress?.[0]?.ip;
        if (lbIngress) {
          return (
            <Badge variant="emerald" dot>
              {lbIngress}
            </Badge>
          );
        }
        return <span className="text-slate-500 font-mono text-xs">&lt;none&gt;</span>;
      },
    },
    {
      key: "ports",
      header: "Port(s)",
      render: (svc) => {
        const ports = svc.spec?.ports || [];
        return (
          <span className="font-mono text-xs text-cyan-300">
            {ports.map((p: any) => `${p.port}/${p.protocol || "TCP"}`).join(", ") || "-"}
          </span>
        );
      },
    },
    {
      key: "age",
      header: "Age",
      render: (svc) => (
        <span className="text-slate-400 text-xs">
          {formatAge(svc.metadata?.creationTimestamp)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (svc) => (
        <div
          className="flex items-center justify-end gap-1.5"
          onClick={(e) => e.stopPropagation()}
        >
          <button
            onClick={() => setSelectedService(svc)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-indigo-500/20 text-indigo-400 border border-white/10 transition-colors"
            title="Inspect Service"
          >
            <FileCode size={14} />
          </button>
          <button
            onClick={() => handleDelete(svc)}
            className="p-1.5 rounded-lg bg-slate-900/80 hover:bg-rose-500/20 text-rose-400 border border-white/10 transition-colors"
            title="Delete Service"
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
            <Server size={24} className="text-indigo-400" />
            <span>Services & MetalLB</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Internal service endpoints and automated bare-metal IP assignments in scope{" "}
            <span className="text-cyan-300 font-mono font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
              {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
            </span>
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchServices}
            isLoading={isLoading}
          >
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Expose Service</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={services}
        searchKey="name"
        searchPlaceholder="Filter services by name..."
        emptyMessage={`No services deployed in ${
          selectedNamespace === "_all" ? "cluster" : selectedNamespace + " namespace"
        }`}
        onRowClick={(svc) => setSelectedService(svc)}
      />

      <ResourceDetailDrawer
        isOpen={!!selectedService}
        onClose={() => setSelectedService(null)}
        resourceType="Service"
        resourceName={selectedService?.metadata?.name || ""}
        namespace={selectedService?.metadata?.namespace || selectedNamespace}
        rawResource={selectedService}
        onActionComplete={fetchServices}
      />
    </div>
  );
}
