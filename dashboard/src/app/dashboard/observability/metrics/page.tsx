"use client";

import React from "react";
import { Cpu, HardDrive, Zap, Activity } from "lucide-react";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";

export default function MetricsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-xl font-bold text-white flex items-center gap-2">
          <Cpu size={22} className="text-cyan-400" />
          <span>Cluster Metrics & Hardware Telemetry</span>
        </h1>
        <p className="text-xs text-slate-400 mt-1">
          Prometheus compatible runtime resource utilization and kernel telemetry
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card className="p-5 space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-semibold uppercase tracking-wider">
              CPU Utilization
            </span>
            <Badge variant="cyan">4% Allocated</Badge>
          </div>
          <div className="text-3xl font-extrabold text-white">0.32 / 8 Cores</div>
          <div className="w-full bg-slate-900 rounded-full h-2 overflow-hidden border border-white/10">
            <div className="bg-gradient-to-r from-cyan-500 to-indigo-500 h-full w-[4%]" />
          </div>
        </Card>

        <Card className="p-5 space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-semibold uppercase tracking-wider">
              Memory Utilization
            </span>
            <Badge variant="emerald">12% Allocated</Badge>
          </div>
          <div className="text-3xl font-extrabold text-white">1.92 / 16 GB</div>
          <div className="w-full bg-slate-900 rounded-full h-2 overflow-hidden border border-white/10">
            <div className="bg-gradient-to-r from-emerald-500 to-cyan-500 h-full w-[12%]" />
          </div>
        </Card>

        <Card className="p-5 space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-semibold uppercase tracking-wider">
              Goroutines & Handlers
            </span>
            <Badge variant="indigo">Ultra-Low Overhead</Badge>
          </div>
          <div className="text-3xl font-extrabold text-white">38 Active</div>
          <p className="text-[11px] text-slate-400">Zero JVM/V8 background bloat, pure Go binary</p>
        </Card>
      </div>
    </div>
  );
}
