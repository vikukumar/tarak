"use client";

import React from "react";
import { Network, RefreshCw, Shield } from "lucide-react";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

export default function MeshPassthroughPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-white flex items-center gap-2">
          <Network size={22} className="text-purple-400" />
          <span>Egress Passthrough Policies</span>
        </h1>
        <p className="text-xs text-slate-400 mt-1">
          Configure whether outbound mesh requests are permitted to external endpoints or restricted
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="p-6 space-y-4 border-cyan-500/30">
          <div className="flex items-center justify-between">
            <h3 className="font-bold text-white text-base">Cluster Egress Status</h3>
            <Badge variant="emerald" dot>
              Passthrough Enabled
            </Badge>
          </div>
          <p className="text-xs text-slate-300">
            Workloads in the default mesh are currently allowed to communicate with external internet
            and third-party APIs over TLS origination.
          </p>
          <div className="p-3 rounded-lg bg-slate-950 text-xs font-mono text-cyan-300 space-y-1">
            <div>mesh.tarak.io/passthrough: "true"</div>
            <div>mesh.tarak.io/tls-origination: "auto"</div>
          </div>
        </Card>

        <Card className="p-6 space-y-4 border-purple-500/30">
          <div className="flex items-center justify-between">
            <h3 className="font-bold text-white text-base">Egress TLS Encryption</h3>
            <Badge variant="purple">Auto-Encrypt</Badge>
          </div>
          <p className="text-xs text-slate-300">
            Outbound traffic to external databases, payment gateways, and SaaS endpoints is automatically
            upgraded to TLS 1.3 at the sidecar proxy layer.
          </p>
        </Card>
      </div>
    </div>
  );
}
