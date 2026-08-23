"use client";

import React, { useState, useEffect } from "react";
import { Workflow, RefreshCw, Plus } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { tarakFetch } from "@/lib/api";

export default function MeshRoutesPage() {
  const [routes, setRoutes] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchRoutes = async () => {
    setIsLoading(true);
    const res = await tarakFetch("/apis/mesh.tarak.io/v1/routes");
    setRoutes(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchRoutes();
  }, []);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Route Name",
      sortable: true,
      render: (r) => (
        <div className="flex items-center gap-2">
          <Workflow size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{r.name}</span>
        </div>
      ),
    },
    {
      key: "service",
      header: "Target Service",
      render: (r) => <span className="font-mono text-cyan-300">{r.service}</span>,
    },
    {
      key: "split",
      header: "Canary Traffic Split",
      render: (r) => {
        const dests = r.destinations || [];
        return (
          <div className="flex gap-2">
            {dests.map((d: any, idx: number) => (
              <Badge key={idx} variant={idx === 0 ? "emerald" : "cyan"}>
                {d.subset || "v1"}: {d.weight}%
              </Badge>
            ))}
          </div>
        );
      },
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Workflow size={22} className="text-cyan-400" />
            <span>Traffic Routing & Canary Splits</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Dynamic weighted traffic splitting, canary rollouts, and blue-green releases in the mesh
          </p>
        </div>

        <Button variant="secondary" size="sm" onClick={fetchRoutes} isLoading={isLoading}>
          <RefreshCw size={14} />
          <span>Refresh</span>
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={routes}
        searchKey="name"
        emptyMessage="No custom mesh traffic routes defined"
      />
    </div>
  );
}
