"use client";

import React, { useState, useEffect } from "react";
import {
  X,
  Globe,
  Layers,
  Box,
  Workflow,
  Server,
  Network,
  ArrowRight,
  Shield,
  Trash2,
} from "lucide-react";
import { tarakFetch } from "@/lib/api";
import { useCluster } from "@/context/ClusterContext";

interface NamespaceDetailDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  namespaceName: string;
  onDeleteNamespace?: () => void;
}

export const NamespaceDetailDrawer: React.FC<NamespaceDetailDrawerProps> = ({
  isOpen,
  onClose,
  namespaceName,
  onDeleteNamespace,
}) => {
  const { setSelectedNamespace } = useCluster();
  const [pods, setPods] = useState<any[]>([]);
  const [deployments, setDeployments] = useState<any[]>([]);
  const [services, setServices] = useState<any[]>([]);
  const [ingresses, setIngresses] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!isOpen || !namespaceName) return;

    async function loadNamespaceResources() {
      setLoading(true);
      try {
        const [podsRes, depRes, svcRes, ingRes] = await Promise.all([
          tarakFetch(`/api/v1/namespaces/${namespaceName}/pods`),
          tarakFetch(`/apis/apps/v1/namespaces/${namespaceName}/deployments`),
          tarakFetch(`/api/v1/namespaces/${namespaceName}/services`),
          tarakFetch(`/apis/networking.k8s.io/v1/namespaces/${namespaceName}/ingresses`),
        ]);
        setPods(podsRes.data?.items || []);
        setDeployments(depRes.data?.items || []);
        setServices(svcRes.data?.items || []);
        setIngresses(ingRes.data?.items || []);
      } finally {
        setLoading(false);
      }
    }

    loadNamespaceResources();
  }, [isOpen, namespaceName]);

  if (!isOpen) return null;

  const handleSwitchToNamespace = (url: string) => {
    setSelectedNamespace(namespaceName);
    window.location.href = url;
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end animate-fade-in">
      <div
        className="fixed inset-0 bg-black/70 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />

      <div className="relative w-full max-w-2xl bg-[#0b1329] border-l border-white/15 h-full flex flex-col shadow-2xl z-10 text-slate-100">
        {/* Header */}
        <div className="p-5 border-b border-white/10 flex items-center justify-between gap-4 bg-slate-950/60">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-cyan-500/20 to-indigo-600/20 border border-cyan-500/30 flex items-center justify-center text-cyan-400 font-bold">
              <Globe size={20} />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-[10px] uppercase font-mono tracking-wider text-cyan-400 font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
                  Namespace Scope
                </span>
                <span className="text-xs text-emerald-400 font-semibold">Active</span>
              </div>
              <h2 className="text-lg font-bold text-white tracking-tight mt-0.5">
                {namespaceName}
              </h2>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => {
                setSelectedNamespace(namespaceName);
                onClose();
              }}
              className="px-3 py-1.5 rounded-lg bg-cyan-500/20 hover:bg-cyan-500/30 text-cyan-300 font-bold text-xs border border-cyan-500/30 flex items-center gap-1.5 transition-colors"
            >
              <span>Set as Active Scope</span>
            </button>
            <button
              onClick={onClose}
              className="p-2 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
            >
              <X size={18} />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-5 space-y-6">
          {/* Summary Stats */}
          <div className="grid grid-cols-4 gap-3">
            <div className="glass-panel p-3 rounded-xl border border-white/10 text-center">
              <span className="text-[10px] uppercase tracking-wider text-slate-400 font-bold block">Pods</span>
              <span className="text-lg font-bold text-cyan-400">{pods.length}</span>
            </div>
            <div className="glass-panel p-3 rounded-xl border border-white/10 text-center">
              <span className="text-[10px] uppercase tracking-wider text-slate-400 font-bold block">Deployments</span>
              <span className="text-lg font-bold text-indigo-400">{deployments.length}</span>
            </div>
            <div className="glass-panel p-3 rounded-xl border border-white/10 text-center">
              <span className="text-[10px] uppercase tracking-wider text-slate-400 font-bold block">Services</span>
              <span className="text-lg font-bold text-emerald-400">{services.length}</span>
            </div>
            <div className="glass-panel p-3 rounded-xl border border-white/10 text-center">
              <span className="text-[10px] uppercase tracking-wider text-slate-400 font-bold block">Ingress</span>
              <span className="text-lg font-bold text-amber-400">{ingresses.length}</span>
            </div>
          </div>

          {/* Pods list */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
                <Box size={14} className="text-cyan-400" />
                <span>Running Pods ({pods.length})</span>
              </h3>
              <button
                onClick={() => handleSwitchToNamespace("/dashboard/workloads/pods/")}
                className="text-[11px] text-cyan-400 hover:underline flex items-center gap-1"
              >
                <span>View all pods</span>
                <ArrowRight size={12} />
              </button>
            </div>
            <div className="space-y-2">
              {pods.length > 0 ? (
                pods.map((p, idx) => (
                  <div
                    key={idx}
                    className="p-3 rounded-lg bg-slate-900/60 border border-white/5 flex items-center justify-between text-xs font-mono"
                  >
                    <div>
                      <span className="font-bold text-white block">{p.metadata?.name}</span>
                      <span className="text-[11px] text-slate-400">Node: {p.spec?.nodeName || "local"}</span>
                    </div>
                    <span className="text-[10px] font-bold px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                      {p.status?.phase || "Running"}
                    </span>
                  </div>
                ))
              ) : (
                <div className="text-xs text-slate-500 p-3 text-center border border-dashed border-white/10 rounded-lg">
                  No pods in this namespace
                </div>
              )}
            </div>
          </div>

          {/* Deployments list */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
                <Workflow size={14} className="text-indigo-400" />
                <span>Deployments ({deployments.length})</span>
              </h3>
              <button
                onClick={() => handleSwitchToNamespace("/dashboard/workloads/deployments/")}
                className="text-[11px] text-indigo-400 hover:underline flex items-center gap-1"
              >
                <span>View deployments</span>
                <ArrowRight size={12} />
              </button>
            </div>
            <div className="space-y-2">
              {deployments.length > 0 ? (
                deployments.map((d, idx) => (
                  <div
                    key={idx}
                    className="p-3 rounded-lg bg-slate-900/60 border border-white/5 flex items-center justify-between text-xs font-mono"
                  >
                    <div>
                      <span className="font-bold text-white block">{d.metadata?.name}</span>
                      <span className="text-[11px] text-slate-400">Replicas: {d.spec?.replicas || 1}</span>
                    </div>
                    <span className="text-cyan-400 text-xs font-bold">
                      {d.status?.readyReplicas || 0}/{d.spec?.replicas || 1} Ready
                    </span>
                  </div>
                ))
              ) : (
                <div className="text-xs text-slate-500 p-3 text-center border border-dashed border-white/10 rounded-lg">
                  No deployments in this namespace
                </div>
              )}
            </div>
          </div>

          {/* Services & Ingress */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
                <Server size={14} className="text-emerald-400" />
                <span>Services & Networking ({services.length})</span>
              </h3>
              <button
                onClick={() => handleSwitchToNamespace("/dashboard/networking/services/")}
                className="text-[11px] text-emerald-400 hover:underline flex items-center gap-1"
              >
                <span>View services</span>
                <ArrowRight size={12} />
              </button>
            </div>
            <div className="space-y-2">
              {services.length > 0 ? (
                services.map((s, idx) => (
                  <div
                    key={idx}
                    className="p-3 rounded-lg bg-slate-900/60 border border-white/5 flex items-center justify-between text-xs font-mono"
                  >
                    <div>
                      <span className="font-bold text-white block">{s.metadata?.name}</span>
                      <span className="text-[11px] text-slate-400">Type: {s.spec?.type || "ClusterIP"}</span>
                    </div>
                    <span className="text-emerald-400 text-xs">
                      Port: {s.spec?.ports?.[0]?.port || 80}
                    </span>
                  </div>
                ))
              ) : (
                <div className="text-xs text-slate-500 p-3 text-center border border-dashed border-white/10 rounded-lg">
                  No services in this namespace
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
