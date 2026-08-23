"use client";

import React, { useState } from "react";
import {
  Shield,
  ShieldAlert,
  ShieldCheck,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Sliders,
  Plus,
  Search,
  RefreshCw,
  Layers,
  ArrowRight,
  FileCode,
  Lock,
  Zap,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

interface PolicyRule {
  name: string;
  description: string;
  category: "Security" | "Reliability" | "BestPractices";
  severity: "High" | "Medium" | "Low";
  mode: "Enforce" | "Audit";
  enabled: boolean;
  matchKinds: string[];
}

interface Violation {
  id: string;
  policyName: string;
  resource: string;
  namespace: string;
  kind: string;
  message: string;
  severity: "High" | "Medium" | "Low";
  mode: "Enforce" | "Audit";
  time: string;
  remediation: string;
}

const initialRules: PolicyRule[] = [
  {
    name: "disallow-privileged-containers",
    description: "Privileged containers can easily escalate host root access and are strictly forbidden.",
    category: "Security",
    severity: "High",
    mode: "Enforce",
    enabled: true,
    matchKinds: ["Pod", "Deployment", "DaemonSet"],
  },
  {
    name: "require-run-as-non-root",
    description: "Containers must execute as a non-root user (UID > 0) to adhere to Pod Security Standards.",
    category: "Security",
    severity: "High",
    mode: "Enforce",
    enabled: true,
    matchKinds: ["Pod", "Deployment"],
  },
  {
    name: "require-resource-limits",
    description: "All containers must define explicit CPU and memory request/limit bounds for fair scheduling.",
    category: "Reliability",
    severity: "Medium",
    mode: "Audit",
    enabled: true,
    matchKinds: ["Pod", "Deployment", "StatefulSet"],
  },
  {
    name: "disallow-host-namespaces",
    description: "Sharing host PID, IPC, or Network namespaces breaks node isolation boundaries.",
    category: "Security",
    severity: "High",
    mode: "Enforce",
    enabled: true,
    matchKinds: ["Pod"],
  },
  {
    name: "require-read-only-rootfs",
    description: "An immutable read-only root filesystem prevents runtime malware persistence inside containers.",
    category: "BestPractices",
    severity: "Medium",
    mode: "Audit",
    enabled: true,
    matchKinds: ["Pod", "Deployment"],
  },
  {
    name: "disallow-default-namespace",
    description: "Workloads should be deployed into dedicated team namespaces rather than default.",
    category: "BestPractices",
    severity: "Low",
    mode: "Audit",
    enabled: true,
    matchKinds: ["Pod", "Deployment", "Service"],
  },
];

const initialViolations: Violation[] = [
  {
    id: "violation-101",
    policyName: "disallow-default-namespace",
    resource: "pod/default/frontend-proxy-7f89",
    namespace: "default",
    kind: "Pod",
    message: "Workload deployed in 'default' namespace violates namespace governance rule.",
    severity: "Low",
    mode: "Audit",
    time: "12 mins ago",
    remediation: "Migrate pod manifest namespace to 'production' or 'staging'.",
  },
  {
    id: "violation-102",
    policyName: "require-resource-limits",
    resource: "deployment/kube-system/metrics-agent",
    namespace: "kube-system",
    kind: "Deployment",
    message: "Container 'collector' does not define resources.limits.cpu.",
    severity: "Medium",
    mode: "Audit",
    time: "34 mins ago",
    remediation: "Set resources.limits.cpu to '500m' and memory to '256Mi'.",
  },
];

export default function PolicyDashboardPage() {
  const [rules, setRules] = useState<PolicyRule[]>(initialRules);
  const [violations, setViolations] = useState<Violation[]>(initialViolations);
  const [search, setSearch] = useState("");
  const [activeTab, setActiveTab] = useState<"rules" | "violations">("rules");

  const toggleRuleMode = (name: string) => {
    setRules((prev) =>
      prev.map((r) =>
        r.name === name
          ? { ...r, mode: r.mode === "Enforce" ? "Audit" : "Enforce" }
          : r
      )
    );
  };

  const toggleRuleEnabled = (name: string) => {
    setRules((prev) =>
      prev.map((r) => (r.name === name ? { ...r, enabled: !r.enabled } : r))
    );
  };

  const filteredRules = rules.filter(
    (r) =>
      r.name.toLowerCase().includes(search.toLowerCase()) ||
      r.category.toLowerCase().includes(search.toLowerCase()) ||
      r.description.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-purple-500/10 border border-purple-500/30 text-purple-400">
              <ShieldCheck size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Kyverno Policy Engine <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 via-indigo-300 to-cyan-400">& Compliance Analysis</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Real-time admission validation, Pod Security Standards enforcement, and automated audit violation reports.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button variant="outline" size="sm" onClick={() => setRules([...initialRules])}>
            <RefreshCw size={14} className="mr-1.5" /> Re-scan Cluster
          </Button>
          <Button size="sm" className="bg-gradient-to-r from-purple-600 to-cyan-600 text-white shadow-lg shadow-purple-900/30">
            <Plus size={14} className="mr-1.5" /> Create ClusterPolicy
          </Button>
        </div>
      </div>

      {/* Compliance Metrics */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div className="p-4 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-1">
          <span className="text-xs text-slate-400 font-medium">Total Rules Active</span>
          <div className="text-2xl font-black text-white">{rules.filter((r) => r.enabled).length}</div>
          <span className="text-[11px] text-purple-400">Kyverno Compatible</span>
        </div>

        <div className="p-4 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-1">
          <span className="text-xs text-slate-400 font-medium">Enforce Mode (Blocking)</span>
          <div className="text-2xl font-black text-rose-400">
            {rules.filter((r) => r.enabled && r.mode === "Enforce").length}
          </div>
          <span className="text-[11px] text-rose-500/80">Strict Admission</span>
        </div>

        <div className="p-4 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-1">
          <span className="text-xs text-slate-400 font-medium">Audit Mode (Monitoring)</span>
          <div className="text-2xl font-black text-cyan-400">
            {rules.filter((r) => r.enabled && r.mode === "Audit").length}
          </div>
          <span className="text-[11px] text-cyan-500/80">Non-blocking logs</span>
        </div>

        <div className="p-4 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-1">
          <span className="text-xs text-slate-400 font-medium">Active Policy Violations</span>
          <div className="text-2xl font-black text-amber-400">{violations.length}</div>
          <span className="text-[11px] text-amber-500/80">Requires Attention</span>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex items-center gap-2 border-b border-white/10 pb-2">
        <button
          onClick={() => setActiveTab("rules")}
          className={`px-4 py-2 rounded-xl text-xs sm:text-sm font-bold transition-all ${
            activeTab === "rules"
              ? "bg-purple-600/20 text-purple-300 border border-purple-500/30"
              : "text-slate-400 hover:text-white"
          }`}
        >
          Active Policies & Security Rules ({rules.length})
        </button>
        <button
          onClick={() => setActiveTab("violations")}
          className={`px-4 py-2 rounded-xl text-xs sm:text-sm font-bold transition-all flex items-center gap-1.5 ${
            activeTab === "violations"
              ? "bg-amber-600/20 text-amber-300 border border-amber-500/30"
              : "text-slate-400 hover:text-white"
          }`}
        >
          <AlertTriangle size={14} className="text-amber-400" />
          Violation Analysis Reports ({violations.length})
        </button>
      </div>

      {/* Tab 1: Rules Table */}
      {activeTab === "rules" && (
        <div className="space-y-4">
          <div className="relative">
            <Search className="absolute left-3 top-2.5 text-slate-500" size={16} />
            <input
              type="text"
              placeholder="Filter policies by name, category, or description..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-9 pr-4 py-2 bg-slate-950/80 border border-white/10 rounded-xl text-xs sm:text-sm text-white placeholder-slate-500 focus:outline-none focus:border-purple-500/50"
            />
          </div>

          <div className="overflow-x-auto rounded-2xl border border-white/10 bg-slate-900/60 shadow-xl">
            <table className="w-full text-left text-xs border-collapse font-mono">
              <thead>
                <tr className="bg-slate-950/90 text-slate-300 uppercase tracking-wider font-bold border-b border-white/10 font-sans">
                  <th className="p-3.5">Policy Name</th>
                  <th className="p-3.5">Category</th>
                  <th className="p-3.5">Severity</th>
                  <th className="p-3.5">Scope</th>
                  <th className="p-3.5">Enforcement Mode</th>
                  <th className="p-3.5 text-right">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/5 font-sans">
                {filteredRules.map((rule, idx) => (
                  <tr key={idx} className="hover:bg-white/[0.02] transition-colors">
                    <td className="p-3.5">
                      <div className="font-bold text-white font-mono text-xs">{rule.name}</div>
                      <div className="text-[11px] text-slate-400 mt-0.5 max-w-md">{rule.description}</div>
                    </td>
                    <td className="p-3.5">
                      <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-slate-800 text-slate-300 border border-white/10">
                        {rule.category}
                      </span>
                    </td>
                    <td className="p-3.5">
                      <Badge variant={rule.severity === "High" ? "rose" : rule.severity === "Medium" ? "amber" : "cyan"}>
                        {rule.severity}
                      </Badge>
                    </td>
                    <td className="p-3.5 font-mono text-[11px] text-slate-400">
                      {rule.matchKinds.join(", ")}
                    </td>
                    <td className="p-3.5">
                      <button
                        onClick={() => toggleRuleMode(rule.name)}
                        className={`px-2.5 py-1 rounded-lg text-xs font-bold font-mono transition-all border ${
                          rule.mode === "Enforce"
                            ? "bg-rose-500/20 text-rose-300 border-rose-500/40 hover:bg-rose-500/30"
                            : "bg-cyan-500/20 text-cyan-300 border-cyan-500/40 hover:bg-cyan-500/30"
                        }`}
                      >
                        {rule.mode} (Click to switch)
                      </button>
                    </td>
                    <td className="p-3.5 text-right">
                      <button
                        onClick={() => toggleRuleEnabled(rule.name)}
                        className={`px-2 py-1 rounded text-xs font-bold transition-all ${
                          rule.enabled
                            ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                            : "bg-slate-800 text-slate-500 border border-white/10"
                        }`}
                      >
                        {rule.enabled ? "Active" : "Disabled"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Tab 2: Violations Stream */}
      {activeTab === "violations" && (
        <div className="space-y-4">
          {violations.map((violation) => (
            <div
              key={violation.id}
              className="p-5 rounded-2xl bg-slate-900/80 border border-amber-500/30 shadow-xl space-y-3"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <span className="p-1.5 rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/30">
                    <AlertTriangle size={16} />
                  </span>
                  <span className="font-bold text-white text-sm font-mono">{violation.resource}</span>
                  <Badge variant="amber">{violation.severity}</Badge>
                </div>
                <span className="text-xs text-slate-400 font-mono">{violation.time}</span>
              </div>

              <div className="p-3 rounded-xl bg-[#04060c] border border-white/10 text-xs text-slate-300 font-mono">
                {violation.message}
              </div>

              <div className="flex flex-wrap items-center justify-between gap-2 text-xs pt-1">
                <div className="text-emerald-400 flex items-center gap-1.5 font-sans">
                  <Zap size={13} />
                  <span><strong>Recommended Remediation:</strong> {violation.remediation}</span>
                </div>
                <Button size="sm" variant="outline" className="text-xs text-cyan-300 border-cyan-500/30">
                  Auto-Fix Manifest
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
