"use client";

import React, { useState, useEffect } from "react";
import { X, Save, RefreshCw, AlertCircle, CheckCircle2, FileCode, Sliders } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { tarakFetch } from "@/lib/api";
import { cn } from "@/lib/utils";

interface EditResourceModalProps {
  isOpen: boolean;
  onClose: () => void;
  resourceType: string;
  resourceName: string;
  namespace?: string;
  rawResource: any;
  onSaved?: () => void;
}

export const EditResourceModal: React.FC<EditResourceModalProps> = ({
  isOpen,
  onClose,
  resourceType,
  resourceName,
  namespace = "default",
  rawResource,
  onSaved,
}) => {
  const [activeTab, setActiveTab] = useState<"visual" | "yaml">("visual");
  const [yamlText, setYamlText] = useState<string>("");
  const [replicas, setReplicas] = useState<number>(1);
  const [image, setImage] = useState<string>("");
  const [svcType, setSvcType] = useState<string>("ClusterIP");
  const [isSaving, setIsSaving] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  useEffect(() => {
    if (rawResource) {
      setYamlText(JSON.stringify(rawResource, null, 2));
      const rep = rawResource.spec?.replicas;
      if (typeof rep === "number") {
        setReplicas(rep);
      }
      const containers = rawResource.spec?.template?.spec?.containers || rawResource.spec?.containers;
      if (containers && containers[0]?.image) {
        setImage(containers[0].image);
      }
      if (rawResource.spec?.type) {
        setSvcType(rawResource.spec.type);
      }
      setErrorMsg(null);
      setSuccessMsg(null);
    }
  }, [rawResource, isOpen]);

  if (!isOpen || !rawResource) return null;

  const getEndpoint = () => {
    const lower = resourceType.toLowerCase();
    const effectiveNs = namespace && namespace !== "_all" ? namespace : rawResource.metadata?.namespace || "default";

    if (lower === "pod" || lower === "pods") return `/api/v1/namespaces/${effectiveNs}/pods/${resourceName}`;
    if (lower === "deployment" || lower === "deployments") return `/apis/apps/v1/namespaces/${effectiveNs}/deployments/${resourceName}`;
    if (lower === "statefulset" || lower === "statefulsets") return `/apis/apps/v1/namespaces/${effectiveNs}/statefulsets/${resourceName}`;
    if (lower === "daemonset" || lower === "daemonsets") return `/apis/apps/v1/namespaces/${effectiveNs}/daemonsets/${resourceName}`;
    if (lower === "service" || lower === "services") return `/api/v1/namespaces/${effectiveNs}/services/${resourceName}`;
    if (lower === "ingress" || lower === "ingresses") return `/apis/networking.k8s.io/v1/namespaces/${effectiveNs}/ingresses/${resourceName}`;
    if (lower === "configmap" || lower === "configmaps") return `/api/v1/namespaces/${effectiveNs}/configmaps/${resourceName}`;
    if (lower === "secret" || lower === "secrets") return `/api/v1/namespaces/${effectiveNs}/secrets/${resourceName}`;
    if (lower === "namespace" || lower === "namespaces") return `/api/v1/namespaces/${resourceName}`;
    return `/api/v1/namespaces/${effectiveNs}/${lower}/${resourceName}`;
  };

  const handleSave = async () => {
    setIsSaving(true);
    setErrorMsg(null);
    setSuccessMsg(null);

    try {
      let payload: any;
      if (activeTab === "yaml") {
        try {
          payload = JSON.parse(yamlText);
        } catch (e: any) {
          throw new Error("Invalid JSON format: " + e.message);
        }
      } else {
        // Clone raw object and apply visual changes
        payload = JSON.parse(JSON.stringify(rawResource));
        if (payload.spec) {
          if (typeof payload.spec.replicas === "number") {
            payload.spec.replicas = Number(replicas);
          }
          if (image) {
            if (payload.spec.template?.spec?.containers?.[0]) {
              payload.spec.template.spec.containers[0].image = image;
            } else if (payload.spec.containers?.[0]) {
              payload.spec.containers[0].image = image;
            }
          }
          if (payload.spec.type && svcType) {
            payload.spec.type = svcType;
          }
        }
      }

      const endpoint = getEndpoint();
      await tarakFetch(endpoint, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      setSuccessMsg("Resource updated successfully!");
      if (onSaved) onSaved();
      setTimeout(() => {
        onClose();
      }, 1000);
    } catch (err: any) {
      setErrorMsg(err.message || "Failed to update resource");
    } finally {
      setIsSaving(false);
    }
  };

  const isWorkload =
    resourceType.toLowerCase().includes("deployment") ||
    resourceType.toLowerCase().includes("statefulset");
  const isService = resourceType.toLowerCase().includes("service");

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div
        className="fixed inset-0 bg-black/80 backdrop-blur-md transition-opacity"
        onClick={onClose}
      />

      <div className="relative w-full max-w-3xl glass-panel bg-[#090f1e] rounded-2xl border border-white/15 p-6 shadow-2xl z-10 max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-white/10">
          <div>
            <div className="flex items-center gap-2">
              <span className="text-[10px] uppercase font-mono tracking-wider text-cyan-400 font-bold bg-cyan-500/10 px-2 py-0.5 rounded border border-cyan-500/20">
                {resourceType}
              </span>
              <span className="text-xs text-slate-400 font-mono">
                ns: <strong className="text-white">{namespace}</strong>
              </span>
            </div>
            <h3 className="text-lg font-bold text-white tracking-wide mt-1">
              Edit {resourceName}
            </h3>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* Tab Toggle */}
        <div className="flex items-center gap-2 pt-4 pb-2 text-xs font-semibold">
          <button
            onClick={() => setActiveTab("visual")}
            className={cn(
              "py-1.5 px-3 rounded-lg flex items-center gap-1.5 border transition-all",
              activeTab === "visual"
                ? "bg-cyan-500/15 border-cyan-500/30 text-cyan-400 shadow-sm"
                : "bg-slate-900/50 border-white/5 text-slate-400 hover:text-white"
            )}
          >
            <Sliders size={14} />
            <span>Quick Controls</span>
          </button>
          <button
            onClick={() => setActiveTab("yaml")}
            className={cn(
              "py-1.5 px-3 rounded-lg flex items-center gap-1.5 border transition-all",
              activeTab === "yaml"
                ? "bg-cyan-500/15 border-cyan-500/30 text-cyan-400 shadow-sm"
                : "bg-slate-900/50 border-white/5 text-slate-400 hover:text-white"
            )}
          >
            <FileCode size={14} />
            <span>Raw Manifest (JSON / YAML)</span>
          </button>
        </div>

        {/* Alerts */}
        {errorMsg && (
          <div className="my-2 p-3 rounded-xl bg-rose-500/15 border border-rose-500/30 flex items-center gap-2 text-xs text-rose-400">
            <AlertCircle size={16} className="flex-shrink-0" />
            <span>{errorMsg}</span>
          </div>
        )}
        {successMsg && (
          <div className="my-2 p-3 rounded-xl bg-emerald-500/15 border border-emerald-500/30 flex items-center gap-2 text-xs text-emerald-400">
            <CheckCircle2 size={16} className="flex-shrink-0" />
            <span>{successMsg}</span>
          </div>
        )}

        {/* Form Body */}
        <div className="flex-1 overflow-y-auto py-3 space-y-4 font-sans text-xs">
          {activeTab === "visual" ? (
            <div className="space-y-4">
              {isWorkload && (
                <div className="p-4 rounded-xl bg-slate-950/60 border border-white/10 space-y-3">
                  <label className="block font-bold text-slate-300">
                    Desired Replicas (Scale Workload)
                  </label>
                  <div className="flex items-center gap-4">
                    <input
                      type="range"
                      min="0"
                      max="10"
                      value={replicas}
                      onChange={(e) => setReplicas(Number(e.target.value))}
                      className="w-full h-2 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-cyan-500"
                    />
                    <span className="w-12 text-center font-mono font-bold text-sm text-cyan-300 px-2 py-1 bg-slate-900 rounded border border-white/10">
                      {replicas}
                    </span>
                  </div>
                </div>
              )}

              {isService && (
                <div className="p-4 rounded-xl bg-slate-950/60 border border-white/10 space-y-3">
                  <label className="block font-bold text-slate-300">
                    Service Type
                  </label>
                  <select
                    value={svcType}
                    onChange={(e) => setSvcType(e.target.value)}
                    className="w-full bg-slate-900 border border-white/15 rounded-lg px-3 py-2 text-white font-mono"
                  >
                    <option value="ClusterIP">ClusterIP (Internal Only)</option>
                    <option value="NodePort">NodePort (Host Port Binding)</option>
                    <option value="LoadBalancer">LoadBalancer (MetalLB & VIP Allocation)</option>
                  </select>
                </div>
              )}

              <div className="p-4 rounded-xl bg-slate-950/60 border border-white/10 space-y-2">
                <label className="block font-bold text-slate-300">
                  Container Image
                </label>
                <input
                  type="text"
                  value={image}
                  onChange={(e) => setImage(e.target.value)}
                  placeholder="e.g. nginx:alpine, redis:7-alpine, python:3.11"
                  className="w-full bg-slate-900 border border-white/15 rounded-lg px-3 py-2 text-white font-mono"
                />
              </div>
            </div>
          ) : (
            <div className="h-full">
              <textarea
                value={yamlText}
                onChange={(e) => setYamlText(e.target.value)}
                rows={16}
                className="w-full h-full min-h-[280px] bg-slate-950 border border-white/10 rounded-xl p-3 font-mono text-xs text-cyan-300 focus:outline-none focus:border-cyan-500"
                spellCheck={false}
              />
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 pt-4 border-t border-white/10">
          <Button variant="secondary" size="sm" onClick={onClose} disabled={isSaving}>
            Cancel
          </Button>
          <Button size="sm" onClick={handleSave} isLoading={isSaving}>
            <Save size={14} />
            <span>Apply Changes</span>
          </Button>
        </div>
      </div>
    </div>
  );
};
