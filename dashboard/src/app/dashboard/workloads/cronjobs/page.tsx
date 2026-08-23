"use client";

import React, { useState } from "react";
import {
  Clock,
  Play,
  Pause,
  RefreshCw,
  Plus,
  Search,
  Calendar,
  CheckCircle2,
  Trash2,
  FileCode,
  Zap,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

interface CronJobItem {
  name: string;
  namespace: string;
  schedule: string;
  humanSchedule: string;
  suspend: boolean;
  active: number;
  lastSchedule: string;
  concurrencyPolicy: "Allow" | "Forbid" | "Replace";
  image: string;
}

const initialCronJobs: CronJobItem[] = [
  {
    name: "db-nightly-backup",
    namespace: "production",
    schedule: "0 2 * * *",
    humanSchedule: "Every day at 02:00 UTC",
    suspend: false,
    active: 0,
    lastSchedule: "14 hours ago",
    concurrencyPolicy: "Forbid",
    image: "backup-tool:v2.1",
  },
  {
    name: "analytics-aggregator",
    namespace: "analytics",
    schedule: "*/15 * * * *",
    humanSchedule: "Every 15 minutes",
    suspend: false,
    active: 1,
    lastSchedule: "8 mins ago",
    concurrencyPolicy: "Allow",
    image: "spark-driver:latest",
  },
  {
    name: "cache-pruner",
    namespace: "default",
    schedule: "0 */6 * * *",
    humanSchedule: "Every 6 hours",
    suspend: true,
    active: 0,
    lastSchedule: "1 day ago",
    concurrencyPolicy: "Replace",
    image: "redis-pruner:1.0",
  },
];

export default function CronJobsPage() {
  const [cronJobs, setCronJobs] = useState<CronJobItem[]>(initialCronJobs);
  const [search, setSearch] = useState("");
  const [isRunning, setIsRunning] = useState<string | null>(null);

  const toggleSuspend = (name: string) => {
    setCronJobs((prev) =>
      prev.map((c) => (c.name === name ? { ...c, suspend: !c.suspend } : c))
    );
  };

  const triggerRunNow = (name: string) => {
    setIsRunning(name);
    setTimeout(() => {
      setCronJobs((prev) =>
        prev.map((c) => (c.name === name ? { ...c, active: c.active + 1, lastSchedule: "Just now" } : c))
      );
      setIsRunning(null);
    }, 1000);
  };

  const filtered = cronJobs.filter(
    (c) =>
      c.name.toLowerCase().includes(search.toLowerCase()) ||
      c.namespace.toLowerCase().includes(search.toLowerCase()) ||
      c.schedule.includes(search)
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-400">
              <Clock size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Scheduled CronJobs <span className="text-transparent bg-clip-text bg-gradient-to-r from-amber-400 via-orange-300 to-rose-400">& Automated Batch Tasks</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Automated recurring job scheduling with concurrency protection, failure backoff, and execution history.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button size="sm" className="bg-gradient-to-r from-amber-600 to-rose-600 text-white shadow-lg shadow-amber-950/40">
            <Plus size={14} className="mr-1.5" /> Create CronJob
          </Button>
        </div>
      </div>

      {/* Grid of CronJobs */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {filtered.map((cron) => (
          <div
            key={cron.name}
            className="p-5 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-4 hover:border-white/20 transition-all"
          >
            <div className="flex items-start justify-between gap-2">
              <div className="space-y-1">
                <span className="font-bold text-white text-sm font-mono">{cron.name}</span>
                <div className="text-xs text-slate-400 font-mono">
                  Namespace: <span className="text-cyan-300">{cron.namespace}</span>
                </div>
              </div>
              <Badge variant={cron.suspend ? "slate" : "emerald"}>
                {cron.suspend ? "Suspended" : "Active"}
              </Badge>
            </div>

            <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 space-y-1 font-mono text-xs">
              <div className="flex items-center justify-between text-amber-300 font-bold">
                <span>{cron.schedule}</span>
                <span className="text-[10px] text-slate-400 font-sans">({cron.concurrencyPolicy})</span>
              </div>
              <div className="text-[11px] text-slate-400 font-sans">{cron.humanSchedule}</div>
            </div>

            <div className="grid grid-cols-2 gap-2 text-[11px] font-mono text-slate-300">
              <div>Running: <span className="text-cyan-400 font-bold">{cron.active}</span></div>
              <div>Last: <span className="text-slate-400">{cron.lastSchedule}</span></div>
            </div>

            <div className="flex items-center justify-between pt-2 border-t border-white/5">
              <Button
                size="sm"
                variant="outline"
                onClick={() => triggerRunNow(cron.name)}
                disabled={isRunning === cron.name}
                className="text-xs text-amber-300 border-amber-500/30"
              >
                <Zap size={12} className="mr-1" />
                {isRunning === cron.name ? "Launching..." : "Trigger Now"}
              </Button>
              <button
                onClick={() => toggleSuspend(cron.name)}
                className="text-xs text-slate-400 hover:text-white font-mono"
              >
                {cron.suspend ? "Resume" : "Suspend"}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
