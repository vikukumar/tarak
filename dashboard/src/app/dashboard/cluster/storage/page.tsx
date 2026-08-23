"use client";

import React, { useState, useEffect } from "react";
import { Database, RefreshCw, Plus } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function StoragePage() {
  const { selectedNamespace } = useClusterState();
  const [pvcs, setPvcs] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchPvcs = async () => {
    setIsLoading(true);
    const res = await tarakFetch(`/api/v1/namespaces/${selectedNamespace}/persistentvolumeclaims`);
    setPvcs(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchPvcs();
  }, [selectedNamespace]);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "PVC Name",
      sortable: true,
      render: (p) => (
        <div className="flex items-center gap-2">
          <Database size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{p.metadata?.name}</span>
        </div>
      ),
    },
    {
      key: "status",
      header: "Status",
      render: () => <Badge variant="emerald" dot>Bound</Badge>,
    },
    {
      key: "capacity",
      header: "Capacity",
      render: (p) => p.spec?.resources?.requests?.storage || "10Gi",
    },
    {
      key: "age",
      header: "Age",
      render: (p) => formatAge(p.metadata?.creationTimestamp),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Database size={22} className="text-cyan-400" />
            <span>Persistent Volume Claims (PVC)</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Dynamic storage volume claims in namespace <span className="text-cyan-400 font-mono">{selectedNamespace}</span>
          </p>
        </div>

        <Button variant="secondary" size="sm" onClick={fetchPvcs} isLoading={isLoading}>
          <RefreshCw size={14} />
          <span>Refresh</span>
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={pvcs}
        searchKey="name"
        emptyMessage={`No PVC storage claims found in ${selectedNamespace} namespace`}
      />
    </div>
  );
}
