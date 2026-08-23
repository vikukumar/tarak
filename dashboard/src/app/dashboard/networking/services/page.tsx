"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Server, RefreshCw, Plus } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function ServicesPage() {
  const { selectedNamespace } = useClusterState();
  const [services, setServices] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchServices = async () => {
    setIsLoading(true);
    const res = await tarakFetch(`/api/v1/namespaces/${selectedNamespace}/services`);
    setServices(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchServices();
  }, [selectedNamespace]);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Service Name",
      sortable: true,
      render: (svc) => (
        <div className="flex items-center gap-2">
          <Server size={16} className="text-indigo-400" />
          <span className="font-semibold text-white">{svc.metadata?.name}</span>
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
      render: (svc) => formatAge(svc.metadata?.creationTimestamp),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Server size={22} className="text-indigo-400" />
            <span>Services & MetalLB</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Internal service endpoints and automated bare-metal IP assignments in{" "}
            <span className="text-cyan-400 font-mono">{selectedNamespace}</span>
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={fetchServices} isLoading={isLoading}>
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
        emptyMessage={`No services deployed in ${selectedNamespace} namespace`}
      />
    </div>
  );
}
