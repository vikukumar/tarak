"use client";

import React, { useState, useEffect } from "react";
import { Cloud, Shield, Globe, ExternalLink, RefreshCw, Plus, CheckCircle2 } from "lucide-react";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { tarakFetch } from "@/lib/api";

export default function TunnelsPage() {
  const [tunnels, setTunnels] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const fetchTunnels = async () => {
    setIsLoading(true);
    const res = await tarakFetch("/apis/networking.tarak.io/v1/tunnels");
    setTunnels(res.data?.items || []);
    setIsLoading(false);
  };

  useEffect(() => {
    fetchTunnels();
  }, []);

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Tunnel Name",
      sortable: true,
      render: (t) => (
        <div className="flex items-center gap-2">
          <Cloud size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{t.name}</span>
        </div>
      ),
    },
    {
      key: "type",
      header: "Provider",
      render: (t) => (
        <Badge variant={t.type === "cloudflare" ? "cyan" : "indigo"}>
          {t.type?.toUpperCase()}
        </Badge>
      ),
    },
    {
      key: "status",
      header: "Status",
      render: (t) => (
        <Badge variant="emerald" dot>
          {t.status || "Active"}
        </Badge>
      ),
    },
    {
      key: "publicUrl",
      header: "Public URL",
      render: (t) => (
        <a
          href={t.publicUrl}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1 text-cyan-400 hover:underline font-mono text-xs"
        >
          <span>{t.publicUrl}</span>
          <ExternalLink size={10} />
        </a>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Cloud size={22} className="text-cyan-400" />
            <span>Cloudflare & Tailscale Integrations</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Zero-trust edge tunnel ingress and mesh VPN connectivity managed from UI
          </p>
        </div>

        <Button variant="secondary" size="sm" onClick={fetchTunnels} isLoading={isLoading}>
          <RefreshCw size={14} />
          <span>Refresh</span>
        </Button>
      </div>

      {/* Integration Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="p-5 space-y-3 border-cyan-500/30">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <Cloud size={20} className="text-cyan-400" />
              <h3 className="font-bold text-white text-sm">Cloudflare Tunnel (cloudflared)</h3>
            </div>
            <Badge variant="cyan">Native Ingress</Badge>
          </div>
          <p className="text-xs text-slate-300">
            Publish cluster workloads to your custom domain with automatic Cloudflare DDoS protection,
            WAF, and global SSL certificates.
          </p>
          <div className="flex items-center gap-2 text-[11px] font-mono text-cyan-300">
            <span>ingressClassName: <strong>tarak-cloudflare</strong></span>
          </div>
        </Card>

        <Card className="p-5 space-y-3 border-indigo-500/30">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <Shield size={20} className="text-indigo-400" />
              <h3 className="font-bold text-white text-sm">Tailscale Mesh Network</h3>
            </div>
            <Badge variant="indigo">MagicDNS</Badge>
          </div>
          <p className="text-xs text-slate-300">
            Peer-to-peer wireguard overlay network connecting your local cluster seamlessly with
            remote developers, edge nodes, and cloud VPCs.
          </p>
          <div className="flex items-center gap-2 text-[11px] font-mono text-indigo-300">
            <span>ingressClassName: <strong>tarak-tailscale</strong></span>
          </div>
        </Card>
      </div>

      <DataTable
        columns={columns}
        data={tunnels}
        searchKey="name"
        emptyMessage="No active tunnels found"
      />
    </div>
  );
}
