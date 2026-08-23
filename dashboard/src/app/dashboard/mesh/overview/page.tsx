"use client";

import React, { useState, useEffect } from "react";
import { Radio, Shield, Network, RefreshCw, Plus, Lock, CheckCircle2 } from "lucide-react";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";
import { tarakFetch } from "@/lib/api";
import { formatAge } from "@/lib/utils";

export default function MeshOverviewPage() {
  const [meshes, setMeshes] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newMeshName, setNewMeshName] = useState("");
  const [mtlsMode, setMtlsMode] = useState("Strict");

  const fetchMeshes = async () => {
    setIsLoading(true);
    const res = await tarakFetch("/apis/mesh.tarak.io/v1/meshes");
    setMeshes(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchMeshes();
  }, []);

  const handleCreateMesh = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMeshName.trim()) return;

    await tarakFetch("/apis/mesh.tarak.io/v1/meshes", {
      method: "POST",
      body: JSON.stringify({
        name: newMeshName.trim().toLowerCase(),
        mtls: {
          mode: mtlsMode,
          trustDomain: `${newMeshName.trim().toLowerCase()}.tarak.mesh`,
        },
      }),
    });
    setIsCreateOpen(false);
    setNewMeshName("");
    fetchMeshes();
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Mesh Tenant",
      sortable: true,
      render: (m) => (
        <div className="flex items-center gap-2">
          <Radio size={16} className="text-purple-400" />
          <span className="font-semibold text-white">{m.name || m.metadata?.name}</span>
        </div>
      ),
    },
    {
      key: "status",
      header: "Phase",
      render: () => <Badge variant="emerald" dot>Active</Badge>,
    },
    {
      key: "mtls",
      header: "mTLS Zero-Trust",
      render: (m) => (
        <Badge variant={m.mtls?.mode === "Strict" ? "purple" : "cyan"}>
          {m.mtls?.mode || "Strict"} Mode
        </Badge>
      ),
    },
    {
      key: "domain",
      header: "Virtual DNS Trust Domain",
      render: (m) => (
        <span className="font-mono text-xs text-cyan-300">
          {m.mtls?.trustDomain || `${m.name || "default"}.tarak.mesh`}
        </span>
      ),
    },
    {
      key: "passthrough",
      header: "Egress Policy",
      render: () => <Badge variant="emerald">Passthrough Enabled</Badge>,
    },
    {
      key: "age",
      header: "Age",
      render: (m) => formatAge(m.createdAt || m.metadata?.creationTimestamp),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Radio size={22} className="text-purple-400" />
            <span>Multi-Mesh Control Plane</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Universal Service Mesh (Kuma / Kong-Mesh equivalent) with zero-trust mTLS and automatic sidecar injection
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="secondary" size="sm" onClick={fetchMeshes} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </Button>
          <Button size="sm" onClick={() => setIsCreateOpen(true)}>
            <Plus size={14} />
            <span>Create Mesh</span>
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={meshes}
        searchKey="name"
        emptyMessage="No meshes defined in cluster"
      />

      {/* Create Mesh Modal */}
      <Modal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        title="Create Isolated Service Mesh"
        maxWidth="sm"
      >
        <form onSubmit={handleCreateMesh} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Mesh Name</label>
            <input
              type="text"
              value={newMeshName}
              onChange={(e) => setNewMeshName(e.target.value)}
              placeholder="e.g. finance-mesh"
              required
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
            />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">mTLS Mode</label>
            <select
              value={mtlsMode}
              onChange={(e) => setMtlsMode(e.target.value)}
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono cursor-pointer"
            >
              <option value="Strict">Strict (Enforce mTLS on all pods)</option>
              <option value="Permissive">Permissive (mTLS + Plaintext)</option>
            </select>
          </div>
          <Button type="submit" className="w-full">
            Create Mesh
          </Button>
        </form>
      </Modal>
    </div>
  );
}
