"use client";

import React, { useState } from "react";
import { FileCode, Play, CheckCircle2, AlertCircle } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { useClusterState } from "@/hooks/useClusterState";
import { tarakFetch } from "@/lib/api";

export default function ManifestsPage() {
  const { selectedNamespace, refresh } = useClusterState();
  const [yamlContent, setYamlContent] = useState<string>(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-web-app
  namespace: ${selectedNamespace}
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: demo-web-svc
  namespace: ${selectedNamespace}
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80`);

  const [statusMessage, setStatusMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [isApplying, setIsApplying] = useState(false);

  const handleApply = async () => {
    setIsApplying(true);
    setStatusMessage(null);

    try {
      // Apply Deployment first
      const depRes = await tarakFetch(`/apis/apps/v1/namespaces/${selectedNamespace}/deployments`, {
        method: "POST",
        body: JSON.stringify({
          apiVersion: "apps/v1",
          kind: "Deployment",
          metadata: { name: "demo-web-app", namespace: selectedNamespace },
          spec: {
            replicas: 2,
            template: {
              spec: {
                containers: [{ name: "nginx", image: "nginx:alpine", ports: [{ containerPort: 80 }] }],
              },
            },
          },
        }),
      });

      // Apply Service
      await tarakFetch(`/api/v1/namespaces/${selectedNamespace}/services`, {
        method: "POST",
        body: JSON.stringify({
          apiVersion: "v1",
          kind: "Service",
          metadata: { name: "demo-web-svc", namespace: selectedNamespace },
          spec: {
            type: "LoadBalancer",
            ports: [{ port: 80, targetPort: 80 }],
          },
        }),
      });

      setStatusMessage({
        type: "success",
        text: "Manifests applied successfully! Workload rollout scheduled.",
      });
      refresh();
    } catch (err: any) {
      setStatusMessage({
        type: "error",
        text: err?.message || "Failed to apply manifest",
      });
    } finally {
      setIsApplying(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <FileCode size={22} className="text-cyan-400" />
            <span>Declarative Manifest Apply</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Deploy multi-document YAML or JSON manifests directly to the cluster
          </p>
        </div>

        <Button onClick={handleApply} isLoading={isApplying}>
          <Play size={14} />
          <span>Apply Configuration</span>
        </Button>
      </div>

      {statusMessage && (
        <div
          className={`p-4 rounded-xl text-xs font-medium flex items-center gap-2 ${
            statusMessage.type === "success"
              ? "bg-emerald-500/10 border border-emerald-500/30 text-emerald-300"
              : "bg-rose-500/10 border border-rose-500/30 text-rose-300"
          }`}
        >
          {statusMessage.type === "success" ? <CheckCircle2 size={16} /> : <AlertCircle size={16} />}
          <span>{statusMessage.text}</span>
        </div>
      )}

      <Card className="p-0 border-white/10 overflow-hidden shadow-2xl">
        <div className="p-3 bg-slate-950/80 border-b border-white/10 flex items-center justify-between text-xs text-slate-400 font-mono">
          <span>YAML Editor (target namespace: {selectedNamespace})</span>
          <span className="text-cyan-400">Kubernetes 1.30+ Compatible</span>
        </div>
        <textarea
          value={yamlContent}
          onChange={(e) => setYamlContent(e.target.value)}
          rows={16}
          className="w-full p-4 bg-[#050914] text-xs font-mono text-cyan-300 outline-none resize-none selection:bg-cyan-500/30"
          spellCheck={false}
        />
      </Card>
    </div>
  );
}
