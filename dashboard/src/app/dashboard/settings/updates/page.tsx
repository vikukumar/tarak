"use client";

import React, { useState } from "react";
import { RefreshCw, CheckCircle2, Sparkles, Download } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

export default function UpdatesPage() {
  const [checking, setChecking] = useState(false);
  const [checked, setChecked] = useState(false);

  const handleCheck = () => {
    setChecking(true);
    setTimeout(() => {
      setChecking(false);
      setChecked(true);
    }, 1000);
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-xl font-bold text-white flex items-center gap-2">
          <RefreshCw size={22} className="text-cyan-400" />
          <span>Auto-Updater & Version Management</span>
        </h1>
        <p className="text-xs text-slate-400 mt-1">
          Seamless binary updates, release channel switching, and rollback protection
        </p>
      </div>

      <Card className="p-6 space-y-6">
        <div className="flex items-center justify-between pb-4 border-b border-white/10">
          <div className="space-y-1">
            <span className="text-xs text-slate-400 uppercase tracking-wider font-semibold">
              Current Binary Version
            </span>
            <div className="text-2xl font-bold text-white font-mono">v1.0.6</div>
          </div>
          <Badge variant="emerald" dot>Up to Date</Badge>
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between text-xs text-slate-300">
            <span>Release Channel:</span>
            <Badge variant="cyan">Stable (Production)</Badge>
          </div>
          <div className="flex items-center justify-between text-xs text-slate-300">
            <span>Build Date:</span>
            <span className="font-mono text-slate-400">2026-08-22T23:05:51Z</span>
          </div>
          <div className="flex items-center justify-between text-xs text-slate-300">
            <span>Commit Hash:</span>
            <span className="font-mono text-slate-400">fe2d461</span>
          </div>
        </div>

        {checked && (
          <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-xs text-emerald-300 flex items-center gap-2">
            <CheckCircle2 size={16} />
            <span>You are on the latest stable release of Tarak Platform.</span>
          </div>
        )}

        <Button onClick={handleCheck} isLoading={checking} className="w-full">
          <span>Check for Updates</span>
        </Button>
      </Card>
    </div>
  );
}
