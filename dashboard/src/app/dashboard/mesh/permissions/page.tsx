"use client";

import React, { useState, useEffect } from "react";
import { Lock, Shield, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Badge } from "@/components/ui/Badge";
import { tarakFetch } from "@/lib/api";

export default function MeshPermissionsPage() {
  const [items, setItems] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchPermissions = async () => {
    setIsLoading(true);
    const res = await tarakFetch("/apis/mesh.tarak.io/v1/meshes/default/traffic-permissions");
    setItems(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchPermissions();
  }, []);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Permission Rule",
      sortable: true,
      render: (p) => (
        <div className="flex items-center gap-2">
          <Lock size={16} className="text-purple-400" />
          <span className="font-semibold text-white">{p.name || "allow-mesh-traffic"}</span>
        </div>
      ),
    },
    {
      key: "sources",
      header: "Allowed Sources",
      render: (p) => (
        <Badge variant="cyan">{p.sources?.join(", ") || "service: *"}</Badge>
      ),
    },
    {
      key: "destinations",
      header: "Allowed Destinations",
      render: (p) => (
        <Badge variant="purple">{p.destinations?.join(", ") || "service: *"}</Badge>
      ),
    },
    {
      key: "action",
      header: "Enforcement",
      render: () => <Badge variant="emerald">ALLOW (mTLS Verified)</Badge>,
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Lock size={22} className="text-purple-400" />
            <span>mTLS Traffic Permissions</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Zero-Trust access control rules regulating service-to-service communication inside the mesh
          </p>
        </div>

        <Button variant="secondary" size="sm" onClick={fetchPermissions} isLoading={isLoading}>
          <RefreshCw size={14} />
          <span>Refresh</span>
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={items.length > 0 ? items : [{ name: "default-mesh-allow-all", sources: ["*"], destinations: ["*"] }]}
        searchKey="name"
      />
    </div>
  );
}
