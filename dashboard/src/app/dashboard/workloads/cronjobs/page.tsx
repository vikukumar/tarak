"use client";

import React, { useState, useEffect } from "react";
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
import { tarakFetch } from "@/lib/api";

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

export default function CronJobsPage() {
  const [cronJobs, setCronJobs] = useState<CronJobItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isRunning, setIsRunning] = useState<string | null>(null);

  const fetchCronJobs = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/batch/v1/cronjobs");
      const items = res.data?.items || [];
      const mapped: CronJobItem[] = items.map((raw: any) => {
        const spec = raw.spec || {};
        const status = raw.status || {};
        const template = spec.jobTemplate?.spec?.template?.spec || {};
        const img = template.containers?.[0]?.image || "container:latest";
        return {
          name: raw.metadata?.name || "cronjob",
          namespace: raw.metadata?.namespace || "default",
          schedule: spec.schedule || "* * * * *",
          humanSchedule: `Schedule: ${spec.schedule || "* * * * *"}`,
          suspend: !!spec.suspend,
          active: status.active?.length || 0,
          lastSchedule: status.lastScheduleTime ? new Date(status.lastScheduleTime).toLocaleTimeString() : "Never",
          concurrencyPolicy: (spec.concurrencyPolicy as "Allow" | "Forbid" | "Replace") || "Allow",
          image: img,
        };
      });
      setCronJobs(mapped);
    } catch {
      setCronJobs([]);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchCronJobs();
  }, []);

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
            <span className="p-2 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-400">
              <Clock size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              CronJobs <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 via-indigo-300 to-cyan-400">& Scheduled Workloads</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Automated recurring batch tasks, maintenance scripts, and backup orchestrations.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={fetchCronJobs}>
            <RefreshCw size={14} className={`mr-1.5 ${isLoading ? "animate-spin" : ""}`} /> Refresh
          </Button>
          <Button size="sm" className="bg-gradient-to-r from-purple-600 to-cyan-600 text-white shadow-lg shadow-purple-950/40">
            <Plus size={14} className="mr-1.5" /> Create CronJob
          </Button>
        </div>
      </div>

      {/* Filter / Search Bar */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <Search size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Filter CronJobs by name, namespace, or cron expression..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-2 rounded-xl bg-slate-900/80 border border-white/10 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-purple-500/50"
          />
        </div>
      </div>

      {/* CronJobs Grid */}
      {filtered.length === 0 ? (
        <div className="p-12 rounded-2xl bg-slate-900/40 border border-white/10 text-center space-y-3">
          <Clock size={36} className="text-slate-600 mx-auto" />
          <h3 className="text-sm font-bold text-white">No CronJobs Configured</h3>
          <p className="text-xs text-slate-400 max-w-md mx-auto">
            Schedule automated jobs to run periodically using standard crontab syntax.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((cj) => (
            <div
              key={`${cj.namespace}/${cj.name}`}
              className="p-5 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-4 hover:border-white/20 transition-all flex flex-col justify-between"
            >
              <div className="space-y-3">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <span className="font-bold text-white text-sm font-mono">{cj.name}</span>
                    <p className="text-xs text-slate-400 font-mono mt-0.5">ns: {cj.namespace}</p>
                  </div>
                  <Badge variant={cj.suspend ? "amber" : "emerald"}>
                    {cj.suspend ? "Suspended" : "Active"}
                  </Badge>
                </div>

                <div className="p-3 rounded-xl bg-slate-950/60 border border-white/5 space-y-1">
                  <div className="flex items-center justify-between text-xs font-mono">
                    <span className="text-slate-400 flex items-center gap-1">
                      <Calendar size={12} className="text-purple-400" /> Expression
                    </span>
                    <span className="text-cyan-300 font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
                      {cj.schedule}
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-400 font-sans">{cj.humanSchedule}</p>
                </div>

                <div className="space-y-1 text-xs text-slate-400 font-mono">
                  <div className="flex justify-between">
                    <span>Active Pods:</span>
                    <span className="text-white font-bold">{cj.active}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Concurrency:</span>
                    <span className="text-purple-300">{cj.concurrencyPolicy}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Last Run:</span>
                    <span className="text-slate-300">{cj.lastSchedule}</span>
                  </div>
                  <div className="flex justify-between truncate">
                    <span>Image:</span>
                    <span className="text-cyan-400 truncate max-w-[140px]">{cj.image}</span>
                  </div>
                </div>
              </div>

              <div className="flex items-center justify-between pt-2 border-t border-white/5">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => triggerRunNow(cj.name)}
                  disabled={isRunning === cj.name}
                  className="text-xs text-purple-300 border-purple-500/30"
                >
                  <Play size={12} className="mr-1" />
                  {isRunning === cj.name ? "Triggering..." : "Run Now"}
                </Button>

                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => toggleSuspend(cj.name)}
                  className={`text-xs ${cj.suspend ? "text-emerald-400" : "text-amber-400"}`}
                >
                  {cj.suspend ? "Resume" : "Suspend"}
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
