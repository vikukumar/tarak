"use client";

import React, { useState } from "react";
import {
  GitBranch,
  RefreshCw,
  Plus,
  Search,
  CheckCircle2,
  AlertTriangle,
  Clock,
  ExternalLink,
  Layers,
  ArrowRight,
  RotateCcw,
  FileCode,
  Shield,
  Box,
  Server,
  Workflow,
  Globe,
  Database,
  Sparkles,
  Zap,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

interface AppResource {
  kind: string;
  name: string;
  namespace: string;
  status: "Synced" | "OutOfSync";
  health: "Healthy" | "Progressing" | "Degraded";
}

interface GitOpsApp {
  name: string;
  namespace: string;
  repoURL: string;
  targetRevision: string;
  path: string;
  destServer: string;
  destNamespace: string;
  syncStatus: "Synced" | "OutOfSync";
  healthStatus: "Healthy" | "Progressing" | "Degraded";
  autoSync: boolean;
  lastSynced: string;
  resources: AppResource[];
}

const initialApps: GitOpsApp[] = [
  {
    name: "ecommerce-storefront",
    namespace: "tarak-cd",
    repoURL: "https://github.com/vikukumar/tarak-examples",
    targetRevision: "main",
    path: "manifests/apps/storefront",
    destServer: "in-cluster (https://127.0.0.1:6443)",
    destNamespace: "production",
    syncStatus: "Synced",
    healthStatus: "Healthy",
    autoSync: true,
    lastSynced: "4 mins ago",
    resources: [
      { kind: "Deployment", name: "storefront-web", namespace: "production", status: "Synced", health: "Healthy" },
      { kind: "Service", name: "storefront-svc", namespace: "production", status: "Synced", health: "Healthy" },
      { kind: "Ingress", name: "storefront-ing", namespace: "production", status: "Synced", health: "Healthy" },
      { kind: "ConfigMap", name: "storefront-env", namespace: "production", status: "Synced", health: "Healthy" },
    ],
  },
  {
    name: "payments-gateway",
    namespace: "tarak-cd",
    repoURL: "https://github.com/vikukumar/tarak-examples",
    targetRevision: "v2.1.0",
    path: "manifests/services/payments",
    destServer: "in-cluster (https://127.0.0.1:6443)",
    destNamespace: "finance",
    syncStatus: "Synced",
    healthStatus: "Healthy",
    autoSync: false,
    lastSynced: "18 mins ago",
    resources: [
      { kind: "StatefulSet", name: "payments-db", namespace: "finance", status: "Synced", health: "Healthy" },
      { kind: "Deployment", name: "payments-api", namespace: "finance", status: "Synced", health: "Healthy" },
      { kind: "Service", name: "payments-clusterip", namespace: "finance", status: "Synced", health: "Healthy" },
    ],
  },
  {
    name: "ai-inference-engine",
    namespace: "tarak-cd",
    repoURL: "https://github.com/vikukumar/tarak-examples",
    targetRevision: "feat/tensorrt-v3",
    path: "manifests/ai/inference",
    destServer: "in-cluster (https://127.0.0.1:6443)",
    destNamespace: "ai-workloads",
    syncStatus: "OutOfSync",
    healthStatus: "Progressing",
    autoSync: true,
    lastSynced: "1 hour ago",
    resources: [
      { kind: "Deployment", name: "tensor-runtime", namespace: "ai-workloads", status: "OutOfSync", health: "Progressing" },
      { kind: "Service", name: "tensor-grpc", namespace: "ai-workloads", status: "Synced", health: "Healthy" },
    ],
  },
];

export default function ContinuousDeliveryPage() {
  const [apps, setApps] = useState<GitOpsApp[]>(initialApps);
  const [selectedApp, setSelectedApp] = useState<GitOpsApp | null>(apps[0]);
  const [search, setSearch] = useState("");
  const [isSyncing, setIsSyncing] = useState<string | null>(null);
  const [showDiff, setShowDiff] = useState(false);

  const handleSync = (name: string) => {
    setIsSyncing(name);
    setTimeout(() => {
      setApps((prev) =>
        prev.map((a) =>
          a.name === name
            ? {
                ...a,
                syncStatus: "Synced",
                healthStatus: "Healthy",
                lastSynced: "Just now",
                resources: a.resources.map((r) => ({ ...r, status: "Synced", health: "Healthy" })),
              }
            : a
        )
      );
      if (selectedApp?.name === name) {
        setSelectedApp((prev) =>
          prev
            ? {
                ...prev,
                syncStatus: "Synced",
                healthStatus: "Healthy",
                lastSynced: "Just now",
                resources: prev.resources.map((r) => ({ ...r, status: "Synced", health: "Healthy" })),
              }
            : null
        );
      }
      setIsSyncing(null);
    }, 1200);
  };

  const filteredApps = apps.filter(
    (a) =>
      a.name.toLowerCase().includes(search.toLowerCase()) ||
      a.destNamespace.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-400">
              <GitBranch size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Continuous Delivery <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 via-indigo-300 to-cyan-400">(ArgoCD GitOps)</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Declarative GitOps application management with automated synchronization, visual resource tree, and live drift reconciliation.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={() => setApps([...initialApps])}>
            <RefreshCw size={14} className="mr-1.5" /> Refresh
          </Button>
          <Button size="sm" className="bg-gradient-to-r from-purple-600 to-cyan-600 hover:from-purple-500 hover:to-cyan-500 text-white shadow-lg shadow-purple-900/30">
            <Plus size={14} className="mr-1.5" /> New Application
          </Button>
        </div>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="p-4 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-1">
          <span className="text-xs text-slate-400 font-medium">Total Applications</span>
          <div className="text-2xl font-black text-white">{apps.length}</div>
          <span className="text-[11px] text-purple-400">GitOps Managed</span>
        </div>

        <div className="p-4 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-1">
          <span className="text-xs text-slate-400 font-medium">Synced Status</span>
          <div className="text-2xl font-black text-emerald-400">
            {apps.filter((a) => a.syncStatus === "Synced").length}
          </div>
          <span className="text-[11px] text-emerald-500/80">In sync with Git</span>
        </div>

        <div className="p-4 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-1">
          <span className="text-xs text-slate-400 font-medium">Healthy Workloads</span>
          <div className="text-2xl font-black text-cyan-400">
            {apps.filter((a) => a.healthStatus === "Healthy").length}
          </div>
          <span className="text-[11px] text-cyan-500/80">Pods & Probes Ready</span>
        </div>

        <div className="p-4 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-1">
          <span className="text-xs text-slate-400 font-medium">Out of Sync / Drift</span>
          <div className="text-2xl font-black text-amber-400">
            {apps.filter((a) => a.syncStatus === "OutOfSync").length}
          </div>
          <span className="text-[11px] text-amber-500/80">Pending Deployment</span>
        </div>
      </div>

      {/* Main Apps Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Apps List (Left Col) */}
        <div className="lg:col-span-5 space-y-4">
          <div className="relative">
            <Search className="absolute left-3 top-2.5 text-slate-500" size={16} />
            <input
              type="text"
              placeholder="Search applications or namespaces..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-9 pr-4 py-2 bg-slate-950/80 border border-white/10 rounded-xl text-xs sm:text-sm text-white placeholder-slate-500 focus:outline-none focus:border-purple-500/50"
            />
          </div>

          <div className="space-y-3">
            {filteredApps.map((app) => {
              const isSelected = selectedApp?.name === app.name;
              return (
                <div
                  key={app.name}
                  onClick={() => setSelectedApp(app)}
                  className={`p-4 rounded-2xl border transition-all cursor-pointer ${
                    isSelected
                      ? "bg-purple-950/30 border-purple-500/50 shadow-xl shadow-purple-950/40 ring-1 ring-purple-500/30"
                      : "bg-slate-900/50 border-white/10 hover:border-white/20 hover:bg-slate-900/80"
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-white text-sm">{app.name}</span>
                        {app.autoSync && (
                          <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-cyan-500/10 text-cyan-300 border border-cyan-500/30">
                            Auto-Sync
                          </span>
                        )}
                      </div>
                      <div className="text-xs text-slate-400 font-mono">
                        Target: <span className="text-slate-300">{app.destNamespace}</span> • Rev: <span className="text-purple-300">{app.targetRevision}</span>
                      </div>
                    </div>

                    <div className="flex flex-col items-end gap-1.5">
                      <Badge variant={app.syncStatus === "Synced" ? "emerald" : "amber"}>
                        {app.syncStatus}
                      </Badge>
                      <Badge variant={app.healthStatus === "Healthy" ? "cyan" : "amber"}>
                        {app.healthStatus}
                      </Badge>
                    </div>
                  </div>

                  <div className="flex items-center justify-between mt-3 pt-3 border-t border-white/5 text-[11px] text-slate-400">
                    <span className="flex items-center gap-1">
                      <Clock size={12} /> {app.lastSynced}
                    </span>
                    <span className="font-mono text-purple-400 flex items-center gap-1">
                      {app.resources.length} objects <ArrowRight size={12} />
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Selected App Detail & Tree View (Right Col) */}
        <div className="lg:col-span-7 space-y-4">
          {selectedApp ? (
            <div className="p-6 rounded-2xl bg-slate-900/80 border border-white/10 shadow-2xl space-y-6">
              {/* App Detail Header */}
              <div className="flex flex-wrap items-center justify-between gap-3 pb-4 border-b border-white/10">
                <div className="space-y-1">
                  <div className="flex items-center gap-2.5">
                    <h2 className="text-xl font-bold text-white">{selectedApp.name}</h2>
                    <Badge variant={selectedApp.syncStatus === "Synced" ? "emerald" : "amber"}>
                      {selectedApp.syncStatus}
                    </Badge>
                    <Badge variant={selectedApp.healthStatus === "Healthy" ? "cyan" : "amber"}>
                      {selectedApp.healthStatus}
                    </Badge>
                  </div>
                  <p className="text-xs text-slate-400 font-mono">{selectedApp.repoURL} ({selectedApp.path})</p>
                </div>

                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => setShowDiff(true)}
                    className="text-xs"
                  >
                    <FileCode size={13} className="mr-1 text-cyan-400" /> Diff View
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => handleSync(selectedApp.name)}
                    disabled={isSyncing === selectedApp.name}
                    className="bg-gradient-to-r from-emerald-600 to-cyan-600 text-white text-xs shadow-lg shadow-emerald-950/40"
                  >
                    <RefreshCw
                      size={13}
                      className={`mr-1 ${isSyncing === selectedApp.name ? "animate-spin" : ""}`}
                    />
                    {isSyncing === selectedApp.name ? "Syncing..." : "Sync App"}
                  </Button>
                </div>
              </div>

              {/* Resource Topology Tree (ArgoCD Style) */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                    <Layers size={14} className="text-purple-400" /> Live Resource Topology Tree
                  </span>
                  <span className="text-xs text-slate-400 font-mono">
                    Namespace: <span className="text-cyan-300 font-bold">{selectedApp.destNamespace}</span>
                  </span>
                </div>

                <div className="p-4 rounded-xl bg-[#04060c] border border-white/10 space-y-3">
                  {selectedApp.resources.map((res, rIdx) => (
                    <div
                      key={rIdx}
                      className="p-3 rounded-lg bg-slate-900/60 border border-white/10 flex items-center justify-between gap-3 hover:border-white/20 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <div className="p-2 rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-300">
                          {res.kind === "Deployment" && <Workflow size={16} />}
                          {res.kind === "Service" && <Server size={16} />}
                          {res.kind === "Ingress" && <Globe size={16} />}
                          {res.kind === "ConfigMap" && <Database size={16} />}
                          {res.kind === "StatefulSet" && <Database size={16} />}
                        </div>
                        <div>
                          <div className="text-xs font-bold text-white flex items-center gap-2">
                            <span>{res.name}</span>
                            <span className="text-[10px] text-slate-400 font-mono">({res.kind})</span>
                          </div>
                          <div className="text-[11px] text-slate-400 font-mono">
                            Status: <span className="text-emerald-400">{res.status}</span> • Health: <span className="text-cyan-400">{res.health}</span>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <Badge variant={res.health === "Healthy" ? "emerald" : "amber"}>
                          {res.health}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Git Repository Source Info */}
              <div className="p-4 rounded-xl bg-slate-950/60 border border-white/10 space-y-2 text-xs">
                <div className="text-slate-400 font-bold uppercase tracking-wider text-[10px]">
                  GitOps Repository Metadata
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-slate-300 font-mono">
                  <div>Repository: <span className="text-white">{selectedApp.repoURL}</span></div>
                  <div>Target Revision: <span className="text-purple-300">{selectedApp.targetRevision}</span></div>
                  <div>Cluster API: <span className="text-cyan-300">{selectedApp.destServer}</span></div>
                  <div>Sync Mode: <span className="text-emerald-300">{selectedApp.autoSync ? "Automated" : "Manual"}</span></div>
                </div>
              </div>
            </div>
          ) : (
            <div className="p-12 text-center text-slate-500 rounded-2xl border border-white/10 bg-slate-900/30">
              Select an application to view live topology and Git sync state.
            </div>
          )}
        </div>
      </div>

      {/* Diff Viewer Modal */}
      {showDiff && selectedApp && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in">
          <div className="bg-[#0b1120] border border-white/10 rounded-2xl max-w-3xl w-full p-6 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between border-b border-white/10 pb-3">
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                <FileCode size={18} className="text-cyan-400" />
                <span>Live Git-to-Cluster Diff: {selectedApp.name}</span>
              </h3>
              <Button size="sm" variant="ghost" onClick={() => setShowDiff(false)}>
                ✕
              </Button>
            </div>

            <pre className="p-4 rounded-xl bg-[#04060c] border border-white/10 font-mono text-xs overflow-x-auto text-slate-300 leading-relaxed">
{`--- Live Cluster State (production/storefront-web)
+++ Desired Git State (main:manifests/apps/storefront)
@@ -14,7 +14,7 @@
     spec:
       containers:
       - name: web
-        image: storefront:v1.0.5
+        image: storefront:v1.0.6
         resources:
           limits:
             cpu: "1000m"`}
            </pre>

            <div className="flex justify-end gap-2 pt-2">
              <Button variant="outline" size="sm" onClick={() => setShowDiff(false)}>
                Close
              </Button>
              <Button
                size="sm"
                onClick={() => {
                  handleSync(selectedApp.name);
                  setShowDiff(false);
                }}
                className="bg-emerald-600 hover:bg-emerald-500 text-white"
              >
                Accept & Reconcile Drift
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
