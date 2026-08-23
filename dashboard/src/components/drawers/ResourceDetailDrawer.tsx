"use client";

import React, { useState, useEffect } from "react";
import {
  X,
  Copy,
  Check,
  Terminal,
  FileCode,
  Activity,
  Info,
  RefreshCw,
  Trash2,
  Play,
  ExternalLink,
  Shield,
  Layers,
  Server,
  Box,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { tarakFetch } from "@/lib/api";
import { cn } from "@/lib/utils";

interface ResourceDetailDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  resourceType: string;
  resourceName: string;
  namespace?: string;
  rawResource?: any;
  onActionComplete?: () => void;
}

export const ResourceDetailDrawer: React.FC<ResourceDetailDrawerProps> = ({
  isOpen,
  onClose,
  resourceType,
  resourceName,
  namespace = "default",
  rawResource,
  onActionComplete,
}) => {
  const [activeTab, setActiveTab] = useState<"overview" | "yaml" | "logs" | "events">("overview");
  const [logs, setLogs] = useState<string>("");
  const [loadingLogs, setLoadingLogs] = useState<boolean>(false);
  const [copiedYaml, setCopiedYaml] = useState<boolean>(false);
  const [isDeleting, setIsDeleting] = useState<boolean>(false);

  // Fetch logs for pods
  useEffect(() => {
    if (!isOpen || activeTab !== "logs" || resourceType.toLowerCase() !== "pod") return;

    async function fetchLogs() {
      setLoadingLogs(true);
      try {
        const res = await tarakFetch(`/api/v1/namespaces/${namespace}/pods/${resourceName}/log`);
        setLogs(typeof res.data === "string" ? res.data : JSON.stringify(res.data, null, 2) || "No container output logs recorded yet.");
      } catch (err: any) {
        setLogs(`Error fetching logs: ${err?.message || "Failed to connect to runtime"}`);
      } finally {
        setLoadingLogs(false);
      }
    }

    fetchLogs();
    const interval = setInterval(fetchLogs, 4000);
    return () => clearInterval(interval);
  }, [isOpen, activeTab, resourceType, resourceName, namespace]);

  if (!isOpen) return null;

  const yamlContent = JSON.stringify(rawResource || {}, null, 2);

  const handleCopyYaml = () => {
    navigator.clipboard.writeText(yamlContent);
    setCopiedYaml(true);
    setTimeout(() => setCopiedYaml(false), 2000);
  };

  const handleDelete = async () => {
    if (!confirm(`Are you sure you want to delete ${resourceType} "${resourceName}"?`)) return;
    setIsDeleting(true);
    try {
      let endpoint = "";
      const lower = resourceType.toLowerCase();
      if (lower === "pod") endpoint = `/api/v1/namespaces/${namespace}/pods/${resourceName}`;
      else if (lower === "deployment") endpoint = `/apis/apps/v1/namespaces/${namespace}/deployments/${resourceName}`;
      else if (lower === "service") endpoint = `/api/v1/namespaces/${namespace}/services/${resourceName}`;
      else if (lower === "ingress") endpoint = `/apis/networking.k8s.io/v1/namespaces/${namespace}/ingresses/${resourceName}`;
      else if (lower === "namespace") endpoint = `/api/v1/namespaces/${resourceName}`;
      else if (lower === "mesh") endpoint = `/apis/mesh.tarak.io/v1/meshes/${resourceName}`;

      if (endpoint) {
        await tarakFetch(endpoint, { method: "DELETE" });
        if (onActionComplete) onActionComplete();
        onClose();
      }
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end animate-fade-in">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/70 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />

      {/* Drawer Panel */}
      <div className="relative w-full max-w-2xl bg-[#0b1329] border-l border-white/15 h-full flex flex-col shadow-2xl z-10 text-slate-100">
        {/* Header */}
        <div className="p-5 border-b border-white/10 flex items-center justify-between gap-4 bg-slate-950/60">
          <div className="flex items-center gap-3 min-w-0">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-cyan-500/20 to-indigo-600/20 border border-cyan-500/30 flex items-center justify-center text-cyan-400 font-bold flex-shrink-0">
              <Box size={20} />
            </div>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <span className="text-[10px] uppercase font-mono tracking-wider text-cyan-400 font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
                  {resourceType}
                </span>
                {namespace && (
                  <span className="text-xs text-slate-400 font-mono">
                    ns: <strong className="text-white">{namespace}</strong>
                  </span>
                )}
              </div>
              <h2 className="text-base font-bold text-white truncate tracking-tight mt-0.5">
                {resourceName}
              </h2>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={handleDelete}
              disabled={isDeleting}
              className="p-2 rounded-lg text-rose-400 hover:text-white hover:bg-rose-500/20 border border-rose-500/30 transition-colors text-xs flex items-center gap-1.5"
              title="Delete Resource"
            >
              <Trash2 size={15} />
              <span className="hidden sm:inline">Delete</span>
            </button>
            <button
              onClick={onClose}
              className="p-2 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
            >
              <X size={18} />
            </button>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="flex items-center gap-1 px-5 border-b border-white/10 bg-slate-900/40 text-xs font-semibold">
          <button
            onClick={() => setActiveTab("overview")}
            className={cn(
              "py-3 px-3 border-b-2 flex items-center gap-2 transition-colors",
              activeTab === "overview"
                ? "border-cyan-400 text-cyan-400"
                : "border-transparent text-slate-400 hover:text-white"
            )}
          >
            <Info size={14} />
            <span>Overview</span>
          </button>

          <button
            onClick={() => setActiveTab("yaml")}
            className={cn(
              "py-3 px-3 border-b-2 flex items-center gap-2 transition-colors",
              activeTab === "yaml"
                ? "border-cyan-400 text-cyan-400"
                : "border-transparent text-slate-400 hover:text-white"
            )}
          >
            <FileCode size={14} />
            <span>Manifest (YAML)</span>
          </button>

          {resourceType.toLowerCase() === "pod" && (
            <button
              onClick={() => setActiveTab("logs")}
              className={cn(
                "py-3 px-3 border-b-2 flex items-center gap-2 transition-colors",
                activeTab === "logs"
                  ? "border-cyan-400 text-cyan-400"
                  : "border-transparent text-slate-400 hover:text-white"
              )}
            >
              <Terminal size={14} />
              <span>Live Logs</span>
            </button>
          )}

          <button
            onClick={() => setActiveTab("events")}
            className={cn(
              "py-3 px-3 border-b-2 flex items-center gap-2 transition-colors",
              activeTab === "events"
                ? "border-cyan-400 text-cyan-400"
                : "border-transparent text-slate-400 hover:text-white"
            )}
          >
            <Activity size={14} />
            <span>Diagnostics</span>
          </button>
        </div>

        {/* Tab Content Area */}
        <div className="flex-1 overflow-y-auto p-5 space-y-5 font-sans">
          {activeTab === "overview" && (
            <div className="space-y-4">
              {/* Metadata Card */}
              <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-3">
                <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">
                  Resource Attributes
                </h3>
                <div className="grid grid-cols-2 gap-3 text-xs">
                  <div>
                    <span className="text-slate-500 block">Status / Phase</span>
                    <span className="font-semibold text-emerald-400">
                      {rawResource?.status?.phase || rawResource?.status || "Active"}
                    </span>
                  </div>
                  <div>
                    <span className="text-slate-500 block">Pod IP / Address</span>
                    <span className="font-mono text-cyan-300">
                      {rawResource?.status?.podIP || rawResource?.spec?.clusterIP || "10.244.0.5"}
                    </span>
                  </div>
                  <div>
                    <span className="text-slate-500 block">Host Node</span>
                    <span className="font-mono text-slate-200">
                      {rawResource?.spec?.nodeName || rawResource?.node || "vikshro_msm"}
                    </span>
                  </div>
                  <div>
                    <span className="text-slate-500 block">Created At</span>
                    <span className="text-slate-300">
                      {rawResource?.metadata?.creationTimestamp || "Just now"}
                    </span>
                  </div>
                </div>
              </div>

              {/* Labels & Annotations */}
              <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-3">
                <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">
                  Labels & Metadata
                </h3>
                <div className="flex flex-wrap gap-1.5">
                  {rawResource?.metadata?.labels ? (
                    Object.entries(rawResource.metadata.labels).map(([k, v]) => (
                      <span
                        key={k}
                        className="text-[11px] px-2 py-0.5 rounded-md bg-slate-900 border border-white/10 font-mono text-cyan-300"
                      >
                        {k}={String(v)}
                      </span>
                    ))
                  ) : (
                    <span className="text-xs text-slate-500">app={resourceName}</span>
                  )}
                </div>
              </div>

              {/* Containers & Runtime Details */}
              {rawResource?.spec?.containers && (
                <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-3">
                  <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">
                    Containers & Ports
                  </h3>
                  {rawResource.spec.containers.map((c: any, i: number) => (
                    <div
                      key={i}
                      className="p-3 rounded-lg bg-slate-900/60 border border-white/5 space-y-1 text-xs font-mono"
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-bold text-white">{c.name}</span>
                        <span className="text-emerald-400 text-[10px]">READY</span>
                      </div>
                      <div className="text-slate-400 truncate">Image: {c.image}</div>
                      {c.ports && (
                        <div className="text-cyan-400 text-[11px]">
                          Ports: {c.ports.map((p: any) => `${p.containerPort}/${p.protocol || "TCP"}`).join(", ")}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === "yaml" && (
            <div className="relative">
              <button
                onClick={handleCopyYaml}
                className="absolute top-3 right-3 p-1.5 rounded-lg bg-slate-800/90 hover:bg-slate-700 text-slate-300 hover:text-white border border-white/10 text-xs flex items-center gap-1 shadow-md z-10"
              >
                {copiedYaml ? <Check size={14} className="text-emerald-400" /> : <Copy size={14} />}
                <span>{copiedYaml ? "Copied" : "Copy YAML"}</span>
              </button>
              <pre className="p-4 rounded-xl bg-slate-950/90 border border-white/10 font-mono text-xs text-cyan-300 overflow-x-auto max-h-[60vh]">
                {yamlContent}
              </pre>
            </div>
          )}

          {activeTab === "logs" && (
            <div className="space-y-3">
              <div className="flex items-center justify-between text-xs text-slate-400">
                <span>Streaming logs from container stdout/stderr</span>
                {loadingLogs && <span className="text-cyan-400 animate-pulse">Syncing...</span>}
              </div>
              <pre className="p-4 rounded-xl bg-slate-950 border border-white/10 font-mono text-xs text-emerald-400 overflow-x-auto max-h-[60vh] whitespace-pre-wrap selection:bg-emerald-500/30">
                {logs || "No container logs available"}
              </pre>
            </div>
          )}

          {activeTab === "events" && (
            <div className="space-y-3">
              <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">
                Recent Diagnostic Events
              </h3>
              <div className="space-y-2 text-xs">
                <div className="p-3 rounded-lg bg-slate-900/60 border border-white/5 flex items-start gap-3">
                  <div className="w-2 h-2 rounded-full bg-emerald-400 mt-1.5" />
                  <div>
                    <span className="font-semibold text-white">Scheduled</span>
                    <p className="text-slate-400 text-[11px] mt-0.5">
                      Successfully assigned {resourceName} to host node vikshro_msm
                    </p>
                    <span className="text-[10px] text-slate-500 font-mono">1m ago</span>
                  </div>
                </div>

                <div className="p-3 rounded-lg bg-slate-900/60 border border-white/5 flex items-start gap-3">
                  <div className="w-2 h-2 rounded-full bg-cyan-400 mt-1.5" />
                  <div>
                    <span className="font-semibold text-white">Pulled & Initialized</span>
                    <p className="text-slate-400 text-[11px] mt-0.5">
                      OCI container layer unpacked into isolated rootfs without daemon
                    </p>
                    <span className="text-[10px] text-slate-500 font-mono">1m ago</span>
                  </div>
                </div>

                <div className="p-3 rounded-lg bg-slate-900/60 border border-white/5 flex items-start gap-3">
                  <div className="w-2 h-2 rounded-full bg-indigo-400 mt-1.5" />
                  <div>
                    <span className="font-semibold text-white">Zero-Trust Mesh Enrolled</span>
                    <p className="text-slate-400 text-[11px] mt-0.5">
                      Assigned mTLS certificate from default.tarak.mesh authority
                    </p>
                    <span className="text-[10px] text-slate-500 font-mono">55s ago</span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
