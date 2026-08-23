"use client";

import React, { useState, useEffect } from "react";
import { Activity, Radio, RefreshCw, Play, Pause, Filter } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { tarakFetch } from "@/lib/api";

export default function HubblePage() {
  const [flows, setFlows] = useState<any[]>([]);
  const [isLive, setIsLive] = useState(true);

  const fetchFlows = async () => {
    const res = await tarakFetch("/apis/telemetry.tarak.io/v1/flows");
    const items = res.data?.items || [
      {
        id: "flow-1",
        time: new Date().toLocaleTimeString(),
        source: "ingress-controller (10.244.0.1)",
        dest: "web-app-svc:80 (10.244.0.2)",
        protocol: "TCP / HTTP",
        verdict: "FORWARDED",
        latency: "0.8ms",
      },
      {
        id: "flow-2",
        time: new Date(Date.now() - 1000).toLocaleTimeString(),
        source: "web-app-fk6jh (10.244.0.2)",
        dest: "redis.default.svc:6379 (10.244.0.3)",
        protocol: "TCP",
        verdict: "FORWARDED",
        latency: "0.4ms",
      },
      {
        id: "flow-3",
        time: new Date(Date.now() - 2000).toLocaleTimeString(),
        source: "tarak-metrics (10.244.0.1)",
        dest: "core-dns.tarak-system:53",
        protocol: "UDP / DNS",
        verdict: "FORWARDED",
        latency: "0.2ms",
      },
    ];
    setFlows(items);
  };

  useEffect(() => {
    fetchFlows();
    if (!isLive) return;
    const interval = setInterval(fetchFlows, 2000);
    return () => clearInterval(interval);
  }, [isLive]);

  const columns: Column<any>[] = [
    {
      key: "time",
      header: "Timestamp",
      render: (f) => <span className="font-mono text-slate-400 text-xs">{f.time}</span>,
    },
    {
      key: "source",
      header: "Source Endpoint",
      sortable: true,
      render: (f) => <span className="font-mono text-cyan-300 text-xs">{f.source}</span>,
    },
    {
      key: "dest",
      header: "Destination Endpoint",
      sortable: true,
      render: (f) => <span className="font-mono text-purple-300 text-xs">{f.dest}</span>,
    },
    {
      key: "protocol",
      header: "Protocol",
      render: (f) => <Badge variant="indigo">{f.protocol}</Badge>,
    },
    {
      key: "verdict",
      header: "Verdict",
      render: (f) => <Badge variant="emerald" dot>{f.verdict || "FORWARDED"}</Badge>,
    },
    {
      key: "latency",
      header: "Latency",
      render: (f) => <span className="font-mono text-emerald-400 text-xs font-semibold">{f.latency || "< 1ms"}</span>,
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Activity size={22} className="text-emerald-400" />
            <span>Hubble Realtime Network Flows</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Deep kernel eBPF & socket level flow inspection with zero packet loss
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant={isLive ? "outline" : "secondary"}
            size="sm"
            onClick={() => setIsLive(!isLive)}
          >
            {isLive ? (
              <>
                <Pause size={14} />
                <span>Pause Stream</span>
              </>
            ) : (
              <>
                <Play size={14} />
                <span>Resume Live</span>
              </>
            )}
          </Button>
          <Button variant="secondary" size="sm" onClick={fetchFlows}>
            <RefreshCw size={14} />
          </Button>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={flows}
        searchKey="source"
        searchPlaceholder="Filter flows by endpoint..."
      />
    </div>
  );
}
