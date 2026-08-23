"use client";

import React, { useState, useEffect } from "react";
import { Shield, Lock, CheckCircle2, RefreshCw } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { tarakFetch } from "@/lib/api";

export default function ZeroTrustPage() {
  const [policies, setPolicies] = useState<any[]>([
    {
      name: "strict-zero-trust",
      privileged: false,
      readOnlyRootFs: true,
      encryptionAtRest: true,
      networkIsolation: true,
      phase: "Enforced",
    },
    {
      name: "baseline-developer-policy",
      privileged: false,
      readOnlyRootFs: false,
      encryptionAtRest: true,
      networkIsolation: true,
      phase: "Enforced",
    },
  ]);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Policy Name",
      sortable: true,
      render: (p) => (
        <div className="flex items-center gap-2">
          <Shield size={16} className="text-emerald-400" />
          <span className="font-semibold text-white">{p.name}</span>
        </div>
      ),
    },
    {
      key: "privileged",
      header: "Privileged Allowed",
      render: (p) => (
        <Badge variant={p.privileged ? "rose" : "emerald"}>
          {p.privileged ? "ALLOWED" : "DENIED"}
        </Badge>
      ),
    },
    {
      key: "readOnlyRootFs",
      header: "Read-Only RootFS",
      render: (p) => (
        <Badge variant={p.readOnlyRootFs ? "emerald" : "amber"}>
          {p.readOnlyRootFs ? "ENFORCED" : "OPTIONAL"}
        </Badge>
      ),
    },
    {
      key: "encryptionAtRest",
      header: "Memory/Disk Encryption",
      render: () => <Badge variant="emerald">AES-256 GCM</Badge>,
    },
    {
      key: "phase",
      header: "Status",
      render: () => <Badge variant="emerald" dot>Enforced</Badge>,
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Shield size={22} className="text-emerald-400" />
            <span>TarakSecurityPolicy (Zero-Trust Security)</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Native container isolation, syscall filters, and memory encryption policies
          </p>
        </div>

        <Button variant="secondary" size="sm">
          <RefreshCw size={14} />
          <span>Refresh</span>
        </Button>
      </div>

      <DataTable columns={columns} data={policies} searchKey="name" />
    </div>
  );
}
