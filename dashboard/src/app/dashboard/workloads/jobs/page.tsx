"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Zap, RefreshCw, Plus } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function JobsPage() {
  const { selectedNamespace } = useClusterState();
  const [items, setItems] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchItems = async () => {
    setIsLoading(true);
    const res = await tarakFetch(`/apis/batch/v1/namespaces/${selectedNamespace}/jobs`);
    setItems(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchItems();
  }, [selectedNamespace]);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Job Name",
      sortable: true,
      render: (j) => (
        <div className="flex items-center gap-2">
          <Zap size={16} className="text-amber-400" />
          <span className="font-semibold text-white">{j.metadata?.name}</span>
        </div>
      ),
    },
    {
      key: "completions",
      header: "Completions",
      render: (j) => <Badge variant="emerald">{j.status?.succeeded || 0}/1</Badge>,
    },
    {
      key: "age",
      header: "Age",
      render: (j) => formatAge(j.metadata?.creationTimestamp),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Zap size={22} className="text-amber-400" />
            <span>Jobs & CronJobs</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Run-to-completion batch tasks and scheduled crons in{" "}
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
              <span>Create Job</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={items}
        searchKey="name"
        emptyMessage={`No batch jobs found in ${selectedNamespace} namespace`}
      />
    </div>
  );
}
