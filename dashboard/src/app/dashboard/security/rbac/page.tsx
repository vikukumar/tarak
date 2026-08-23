"use client";

import React, { useState, useEffect } from "react";
import { Shield, Lock, RefreshCw, CheckCircle2, XCircle } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { tarakFetch } from "@/lib/api";

export default function RbacPage() {
  const [roles, setRoles] = useState<any[]>([
    {
      role: "cluster-admin",
      scope: "Cluster-Wide",
      verbs: "get, list, watch, create, update, delete, exec",
      resources: "*",
      binding: "admin (Super-Admin)",
    },
    {
      role: "admin",
      scope: "Namespace (default)",
      verbs: "get, list, watch, create, update, delete",
      resources: "pods, deployments, services, ingresses",
      binding: "developers-group",
    },
    {
      role: "view",
      scope: "Namespace (*)",
      verbs: "get, list, watch",
      resources: "pods, services, logs",
      binding: "monitoring-sa",
    },
  ]);

  const columns: Column<any>[] = [
    {
      key: "role",
      header: "Role Name",
      sortable: true,
      render: (r) => (
        <div className="flex items-center gap-2">
          <Shield size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{r.role}</span>
        </div>
      ),
    },
    {
      key: "scope",
      header: "Binding Scope",
      render: (r) => <Badge variant="indigo">{r.scope}</Badge>,
    },
    {
      key: "resources",
      header: "Allowed Resources",
      render: (r) => <span className="font-mono text-cyan-300 text-xs">{r.resources}</span>,
    },
    {
      key: "verbs",
      header: "Allowed Verbs",
      render: (r) => <span className="font-mono text-emerald-400 text-xs">{r.verbs}</span>,
    },
    {
      key: "binding",
      header: "Subjects / Users",
      render: (r) => <span className="text-slate-300 font-medium text-xs">{r.binding}</span>,
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Lock size={22} className="text-cyan-400" />
            <span>RBAC Permission Matrix</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Role-Based Access Control rules, subjects, and namespace permission boundaries
          </p>
        </div>

        <Button variant="secondary" size="sm">
          <RefreshCw size={14} />
          <span>Refresh</span>
        </Button>
      </div>

      <DataTable columns={columns} data={roles} searchKey="role" />
    </div>
  );
}
