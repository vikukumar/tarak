"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Database, RefreshCw, Plus } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function StatefulSetsPage() {
  const { selectedNamespace } = useClusterState();
  const [items, setItems] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchItems = async () => {
    setIsLoading(true);
    const res = await tarakFetch(`/apis/apps/v1/namespaces/${selectedNamespace}/statefulsets`);
    setItems(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchItems();
  }, [selectedNamespace]);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "StatefulSet Name",
      sortable: true,
      render: (s) => (
        <div className="flex items-center gap-2">
          <Database size={16} className="text-indigo-400" />
          <span className="font-semibold text-white">{s.metadata?.name}</span>
        </div>
      ),
    },
    {
      key: "replicas",
      header: "Desired Replicas",
      render: (s) => <Badge variant="indigo">{s.spec?.replicas || 1} Replicas</Badge>,
    },
    {
      key: "age",
      header: "Age",
      render: (s) => formatAge(s.metadata?.creationTimestamp),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Database size={22} className="text-indigo-400" />
            <span>StatefulSets</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Stateful workload sets with ordered persistent storage in{" "}
            <span className="text-cyan-400 font-mono">{selectedNamespace}</span>
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={fetchItems} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <Plus size={14} />
              <span>Create StatefulSet</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={items}
        searchKey="name"
        emptyMessage={`No statefulsets found in ${selectedNamespace} namespace`}
      />
    </div>
  );
}
