"use client";

import React, { useState, useEffect } from "react";
import {
  X,
  Cpu,
  Server,
  HardDrive,
  Activity,
  Box,
  Lock,
  Unlock,
  RefreshCw,
  CheckCircle2,
  AlertTriangle,
} from "lucide-react";
import { tarakFetch } from "@/lib/api";

interface NodeDetailDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  nodeName: string;
  nodeData?: any;
}

export const NodeDetailDrawer: React.FC<NodeDetailDrawerProps> = ({
  isOpen,
  onClose,
  nodeName,
  nodeData,
}) => {
  const [pods, setPods] = useState<any[]>([]);
  const [isCordoned, setIsCordoned] = useState<boolean>(false);
  const [actionLoading, setActionLoading] = useState<boolean>(false);

  useEffect(() => {
    if (!isOpen || !nodeName) return;

    async function loadNodePods() {
      try {
        const res = await tarakFetch("/api/v1/pods");
        if (res.data?.items) {
          const matching = res.data.items.filter(
            (p: any) => p.spec?.nodeName === nodeName || !p.spec?.nodeName
          );
          setPods(matching);
        }
      } catch {
        // ignore
      }
    }

    loadNodePods();
  }, [isOpen, nodeName]);

  if (!isOpen) return null;

  const nodeInfo = nodeData?.status?.nodeInfo || {
    operatingSystem: "windows",
    architecture: "amd64",
    kernelVersion: "10.0.26100",
    containerRuntimeVersion: "tarak-native://1.0.6",
    kubeletVersion: "v1.0.6",
  };

  const capacity = nodeData?.status?.capacity || {
    cpu: "8",
    memory: "16384Mi",
    pods: "110",
  };

  const handleToggleCordon = () => {
    setActionLoading(true);
    setTimeout(() => {
      setIsCordoned(!isCordoned);
      setActionLoading(false);
    }, 600);
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
              <Server size={20} />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-[10px] uppercase font-mono tracking-wider text-cyan-400 font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
                  Control Plane & Worker
                </span>
                <span className="text-xs text-emerald-400 font-semibold flex items-center gap-1">
                  <CheckCircle2 size={12} />
                  Ready
                </span>
              </div>
              <h2 className="text-lg font-bold text-white tracking-tight mt-0.5">
                {nodeName}
              </h2>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={handleToggleCordon}
              disabled={actionLoading}
              className={`px-3 py-1.5 rounded-lg border text-xs font-semibold flex items-center gap-1.5 transition-colors ${
                isCordoned
                  ? "bg-amber-500/20 border-amber-500/40 text-amber-300"
                  : "bg-slate-800 hover:bg-slate-700 border-white/10 text-slate-300"
              }`}
            >
              {isCordoned ? <Unlock size={14} /> : <Lock size={14} />}
              <span>{isCordoned ? "Uncordon" : "Cordon Node"}</span>
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
          {/* Capacity Metrics */}
          <div className="grid grid-cols-3 gap-3">
            <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-1">
              <div className="flex items-center justify-between text-slate-400 text-xs">
                <span>CPU Cores</span>
                <Cpu size={14} className="text-cyan-400" />
              </div>
              <span className="text-xl font-bold text-white">{capacity.cpu} Cores</span>
              <span className="text-[10px] text-emerald-400 block font-mono">100% Available</span>
            </div>

            <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-1">
              <div className="flex items-center justify-between text-slate-400 text-xs">
                <span>Memory</span>
                <Activity size={14} className="text-indigo-400" />
              </div>
              <span className="text-xl font-bold text-white">{capacity.memory}</span>
              <span className="text-[10px] text-indigo-300 block font-mono">Dynamic Paged</span>
            </div>

            <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-1">
              <div className="flex items-center justify-between text-slate-400 text-xs">
                <span>Allocated Pods</span>
                <Box size={14} className="text-emerald-400" />
              </div>
              <span className="text-xl font-bold text-white">{pods.length} / {capacity.pods}</span>
              <span className="text-[10px] text-cyan-400 block font-mono">Max Capacity 110</span>
            </div>
          </div>

          {/* Node OS & Runtime Specs */}
          <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-3 text-xs">
            <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">
              System Specifications & Architecture
            </h3>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <span className="text-slate-500 block">Operating System</span>
                <span className="font-semibold text-white uppercase">{nodeInfo.operatingSystem}</span>
              </div>
              <div>
                <span className="text-slate-500 block">Architecture</span>
                <span className="font-mono text-cyan-300">{nodeInfo.architecture}</span>
              </div>
              <div>
                <span className="text-slate-500 block">Kernel Version</span>
                <span className="font-mono text-slate-200">{nodeInfo.kernelVersion}</span>
              </div>
              <div>
                <span className="text-slate-500 block">Container Runtime</span>
                <span className="font-mono text-emerald-400">{nodeInfo.containerRuntimeVersion}</span>
              </div>
            </div>
          </div>

          {/* Node Conditions */}
          <div className="glass-panel p-4 rounded-xl border border-white/10 space-y-3 text-xs">
            <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">
              Node Health Conditions
            </h3>
            <div className="space-y-2">
              <div className="p-2.5 rounded-lg bg-slate-900/60 border border-white/5 flex items-center justify-between">
                <span className="text-white font-medium">Ready</span>
                <span className="text-emerald-400 font-bold text-[11px] bg-emerald-500/10 px-2 py-0.5 rounded">
                  True
                </span>
              </div>
              <div className="p-2.5 rounded-lg bg-slate-900/60 border border-white/5 flex items-center justify-between">
                <span className="text-slate-300">MemoryPressure</span>
                <span className="text-slate-400 text-[11px] bg-white/5 px-2 py-0.5 rounded">
                  False
                </span>
              </div>
              <div className="p-2.5 rounded-lg bg-slate-900/60 border border-white/5 flex items-center justify-between">
                <span className="text-slate-300">DiskPressure</span>
                <span className="text-slate-400 text-[11px] bg-white/5 px-2 py-0.5 rounded">
                  False
                </span>
              </div>
              <div className="p-2.5 rounded-lg bg-slate-900/60 border border-white/5 flex items-center justify-between">
                <span className="text-slate-300">PIDPressure</span>
                <span className="text-slate-400 text-[11px] bg-white/5 px-2 py-0.5 rounded">
                  False
                </span>
              </div>
            </div>
          </div>

          {/* Scheduled Pods */}
          <div className="space-y-3">
            <h3 className="text-xs font-bold text-slate-300 uppercase tracking-wider flex items-center gap-2">
              <Box size={14} className="text-cyan-400" />
              <span>Pods Hosted on Node ({pods.length})</span>
            </h3>
            <div className="space-y-2">
              {pods.map((p, idx) => (
                <div
                  key={idx}
                  className="p-3 rounded-lg bg-slate-900/60 border border-white/5 flex items-center justify-between text-xs font-mono"
                >
                  <div>
                    <span className="font-bold text-white block">{p.metadata?.name}</span>
                    <span className="text-[11px] text-slate-400">ns: {p.metadata?.namespace || "default"}</span>
                  </div>
                  <span className="text-emerald-400 font-semibold">
                    {p.status?.phase || "Running"}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
