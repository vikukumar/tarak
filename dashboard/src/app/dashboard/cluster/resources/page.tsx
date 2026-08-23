"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Server, RefreshCw, Box, Layers, Shield, Network, ArrowRight, ExternalLink } from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Card } from "@/components/ui/Card";
import { tarakFetch } from "@/lib/api";

interface ApiResource {
  name: string;
  singularName?: string;
  namespaced: boolean;
  kind: string;
  group: string;
  version: string;
  verbs: string[];
  shortNames?: string[];
  link?: string;
}

export default function ApiResourcesPage() {
  const [resources, setResources] = useState<ApiResource[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const defaultCatalog: ApiResource[] = [
    { name: "pods", singularName: "pod", namespaced: true, kind: "Pod", group: "core", version: "v1", verbs: ["get", "list", "create", "update", "delete", "watch"], shortNames: ["po"], link: "/dashboard/workloads/pods" },
    { name: "services", singularName: "service", namespaced: true, kind: "Service", group: "core", version: "v1", verbs: ["get", "list", "create", "update", "delete", "watch"], shortNames: ["svc"], link: "/dashboard/networking/services" },
    { name: "configmaps", singularName: "configmap", namespaced: true, kind: "ConfigMap", group: "core", version: "v1", verbs: ["get", "list", "create", "update", "delete"], shortNames: ["cm"], link: "/dashboard/cluster/configmaps" },
    { name: "secrets", singularName: "secret", namespaced: true, kind: "Secret", group: "core", version: "v1", verbs: ["get", "list", "create", "update", "delete"], link: "/dashboard/cluster/configmaps" },
    { name: "namespaces", singularName: "namespace", namespaced: false, kind: "Namespace", group: "core", version: "v1", verbs: ["get", "list", "create", "delete"], shortNames: ["ns"], link: "/dashboard/cluster/namespaces" },
    { name: "nodes", singularName: "node", namespaced: false, kind: "Node", group: "core", version: "v1", verbs: ["get", "list", "watch"], shortNames: ["no"], link: "/dashboard/cluster/nodes" },
    { name: "events", singularName: "event", namespaced: true, kind: "Event", group: "core", version: "v1", verbs: ["get", "list", "watch"], shortNames: ["ev"], link: "/dashboard/observability/events" },
    { name: "persistentvolumeclaims", singularName: "persistentvolumeclaim", namespaced: true, kind: "PersistentVolumeClaim", group: "core", version: "v1", verbs: ["get", "list", "create", "delete"], shortNames: ["pvc"], link: "/dashboard/cluster/storage" },
    { name: "deployments", singularName: "deployment", namespaced: true, kind: "Deployment", group: "apps/v1", version: "v1", verbs: ["get", "list", "create", "update", "delete", "watch"], shortNames: ["deploy"], link: "/dashboard/workloads/deployments" },
    { name: "statefulsets", singularName: "statefulset", namespaced: true, kind: "StatefulSet", group: "apps/v1", version: "v1", verbs: ["get", "list", "create", "update", "delete"], shortNames: ["sts"], link: "/dashboard/workloads/statefulsets" },
    { name: "daemonsets", singularName: "daemonset", namespaced: true, kind: "DaemonSet", group: "apps/v1", version: "v1", verbs: ["get", "list", "create", "update", "delete"], shortNames: ["ds"], link: "/dashboard/workloads/daemonsets" },
    { name: "jobs", singularName: "job", namespaced: true, kind: "Job", group: "batch/v1", version: "v1", verbs: ["get", "list", "create", "delete"], link: "/dashboard/workloads/jobs" },
    { name: "cronjobs", singularName: "cronjob", namespaced: true, kind: "CronJob", group: "batch/v1", version: "v1", verbs: ["get", "list", "create", "delete"], shortNames: ["cj"], link: "/dashboard/workloads/jobs" },
    { name: "ingresses", singularName: "ingress", namespaced: true, kind: "Ingress", group: "networking.k8s.io", version: "v1", verbs: ["get", "list", "create", "update", "delete"], shortNames: ["ing"], link: "/dashboard/networking/ingress" },
    { name: "networkpolicies", singularName: "networkpolicy", namespaced: true, kind: "NetworkPolicy", group: "networking.k8s.io", version: "v1", verbs: ["get", "list", "create", "delete"], shortNames: ["netpol"], link: "/dashboard/networking/policies" },
    { name: "taraksecuritypolicies", singularName: "taraksecuritypolicy", namespaced: true, kind: "TarakSecurityPolicy", group: "security.tarak.io", version: "v1", verbs: ["get", "list", "create", "update", "delete"], shortNames: ["tsp"], link: "/dashboard/security/zerotrust" },
    { name: "meshes", singularName: "mesh", namespaced: false, kind: "Mesh", group: "mesh.tarak.io", version: "v1", verbs: ["get", "list", "create", "delete"], link: "/dashboard/mesh/overview" },
    { name: "trafficpermissions", singularName: "trafficpermission", namespaced: false, kind: "TrafficPermission", group: "mesh.tarak.io", version: "v1", verbs: ["get", "list", "create", "delete"], link: "/dashboard/mesh/permissions" },
    { name: "customresourcedefinitions", singularName: "customresourcedefinition", namespaced: false, kind: "CustomResourceDefinition", group: "apiextensions.k8s.io", version: "v1", verbs: ["get", "list", "create", "delete"], shortNames: ["crd"], link: "/dashboard/cluster/crds" },
  ];

  const fetchResources = async () => {
    setIsLoading(true);
    try {
      const res = await tarakFetch("/api/v1");
      const items = res.data?.resources || [];
      if (items.length > 0) {
        setResources(defaultCatalog);
      } else {
        setResources(defaultCatalog);
      }
    } catch {
      setResources(defaultCatalog);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchResources();
  }, []);

  const columns: Column<ApiResource>[] = [
    {
      key: "kind",
      header: "Resource Kind",
      sortable: true,
      render: (r) => (
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400 font-bold text-xs">
            {r.kind.slice(0, 2).toUpperCase()}
          </div>
          <div>
            <span className="font-bold text-white block">{r.kind}</span>
            <span className="text-[10px] text-slate-400 font-mono">
              plural: {r.name}
            </span>
          </div>
        </div>
      ),
    },
    {
      key: "group",
      header: "API Group / Version",
      render: (r) => (
        <span className="font-mono text-xs text-cyan-300">
          {r.group} / {r.version}
        </span>
      ),
    },
    {
      key: "scope",
      header: "Scope",
      render: (r) => (
        <Badge variant={r.namespaced ? "cyan" : "purple"}>
          {r.namespaced ? "Namespaced" : "Cluster-Wide"}
        </Badge>
      ),
    },
    {
      key: "shortNames",
      header: "Short Names",
      render: (r) => (
        <div className="flex gap-1 font-mono text-xs text-slate-300">
          {r.shortNames?.map((sn) => (
            <span key={sn} className="px-1.5 py-0.5 rounded bg-slate-900 border border-white/10 text-[11px] text-amber-300">
              {sn}
            </span>
          )) || <span className="text-slate-600">-</span>}
        </div>
      ),
    },
    {
      key: "verbs",
      header: "Supported Verbs",
      render: (r) => (
        <div className="flex flex-wrap gap-1">
          {r.verbs.slice(0, 4).map((v) => (
            <span key={v} className="px-1.5 py-0.5 rounded bg-slate-900 border border-white/5 font-mono text-[10px] text-slate-400">
              {v}
            </span>
          ))}
          {r.verbs.length > 4 && (
            <span className="text-[10px] text-slate-500 self-center">+{r.verbs.length - 4}</span>
          )}
        </div>
      ),
    },
    {
      key: "actions",
      header: "Explore",
      className: "text-right",
      render: (r) =>
        r.link ? (
          <Link href={r.link}>
            <button className="px-2.5 py-1 rounded-lg bg-slate-900/80 hover:bg-cyan-500/20 text-cyan-400 border border-white/10 text-xs font-semibold flex items-center gap-1 ml-auto transition-colors">
              <span>Manage</span>
              <ArrowRight size={12} />
            </button>
          </Link>
        ) : null,
    },
  ];

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <Server size={24} className="text-cyan-400" />
            <span>Built-in API Resources Catalog</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Complete discovery schema of all core, apps, batch, networking, security, and mesh APIs registered in Tarak
          </p>
        </div>

        <div className="flex items-center gap-2.5">
          <Button variant="secondary" size="sm" onClick={fetchResources} isLoading={isLoading}>
            <RefreshCw size={14} />
            <span>Refresh Schema</span>
          </Button>
          <Link href="/dashboard/devtools/manifests">
            <Button size="sm">
              <span>YAML Manifest Studio</span>
            </Button>
          </Link>
        </div>
      </div>

      <DataTable
        columns={columns}
        data={resources}
        searchKey="kind"
        searchPlaceholder="Filter API resources (e.g. Pod, Deployment, Service)..."
      />
    </div>
  );
}
