"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Globe, RefreshCw, Plus, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function IngressPage() {
  const { selectedNamespace } = useClusterState();
  const [ingresses, setIngresses] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchIngresses = async () => {
    setIsLoading(true);
    const res = await tarakFetch(`/apis/networking.k8s.io/v1/namespaces/${selectedNamespace}/ingresses`);
    setIngresses(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchIngresses();
  }, [selectedNamespace]);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Ingress Name",
      sortable: true,
      render: (ing) => (
        <div className="flex items-center gap-2">
          <Globe size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{ing.metadata?.name}</span>
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
        if (rules.length === 0) return <span className="text-slate-500">*</span>;
        return (
          <div className="flex flex-wrap gap-1.5">
            {rules.map((r: any, i: number) => (
              <a
                key={i}
                href={`http://${r.host || "localhost:8080"}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-cyan-500/10 text-cyan-300 font-mono text-[11px] hover:bg-cyan-500/20"
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
      render: (ing) => formatAge(ing.metadata?.creationTimestamp),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Globe size={22} className="text-cyan-400" />
            <span>Ingress Routes & Domains</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            HTTP/HTTPS virtual host routing, automatic SSL/TLS, and reverse proxies in{" "}
            <span className="text-cyan-400 font-mono">{selectedNamespace}</span>
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={fetchIngresses} isLoading={isLoading}>
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
        emptyMessage={`No ingress routes defined in ${selectedNamespace} namespace`}
      />
    </div>
  );
}
