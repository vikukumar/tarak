import React from "react";
import { Network, Code, CheckCircle, Server } from "lucide-react";

export const ApiReferencePage: React.FC = () => {
  const endpoints = [
    { method: "GET", path: "/api/v1/namespaces/{ns}/pods", desc: "List or watch container pods with label selector and field filters" },
    { method: "POST", path: "/api/v1/namespaces/{ns}/pods", desc: "Create, schedule, and initialize a new container pod" },
    { method: "GET", path: "/api/v1/namespaces/{ns}/pods/{name}", desc: "Get detailed pod runtime status, container states, and IP" },
    { method: "DELETE", path: "/api/v1/namespaces/{ns}/pods/{name}", desc: "Gracefully terminate and delete a container pod" },
    { method: "GET", path: "/api/v1/namespaces/{ns}/pods/{name}/log", desc: "Stream stdout/stderr logs for a specific container" },
    { method: "POST", path: "/api/v1/namespaces/{ns}/pods/{name}/exec", desc: "Interactive WebSocket pseudo-terminal container execution" },
    { method: "GET", path: "/api/v1/namespaces/{ns}/services", desc: "List all ClusterIP, NodePort, and LoadBalancer services" },
    { method: "POST", path: "/api/v1/namespaces/{ns}/services", desc: "Create or register a new internal service endpoint" },
    { method: "GET", path: "/apis/apps/v1/namespaces/{ns}/deployments", desc: "List active deployments and replica set controllers" },
    { method: "POST", path: "/apis/apps/v1/namespaces/{ns}/deployments", desc: "Create or update declarative deployment specs" },
    { method: "GET", path: "/apis/mesh.tarak.io/v1/trafficpermissions", desc: "List active zero-trust mTLS service mesh policies" },
    { method: "GET", path: "/apis/mesh.tarak.io/v1/trafficroutes", desc: "List canary and weighted traffic split rules" },
    { method: "GET", path: "/apis/metrics.k8s.io/v1beta1/nodes", desc: "Kernel-level physical CPU, memory, and disk telemetry" },
    { method: "/metrics", path: "/metrics", desc: "Standard Prometheus scraper metrics endpoint" },
    { method: "GET", path: "/healthz & /livez", desc: "Liveness and readiness health probes for load balancers" },
  ];

  return (
    <div className="space-y-10 animate-fade-in max-w-5xl mx-auto">
      {/* Title */}
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs font-bold uppercase tracking-wider">
          REST API & Schema Manual
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          Kubernetes-Compatible <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-indigo-400">API Reference</span>
        </h1>
        <p className="text-slate-400 max-w-xl mx-auto text-sm sm:text-base">
          Full REST OpenAPI 3.0 specification for Tarak control plane, core resources, CRDs, and live telemetry endpoints.
        </p>
      </div>

      {/* Endpoints Table */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <h2 className="text-lg font-bold text-white">Cluster Endpoints Directory</h2>
        
        <div className="overflow-x-auto rounded-xl border border-white/10">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-slate-950/80 text-slate-300 uppercase tracking-wider font-bold border-b border-white/10">
                <th className="p-3">Method</th>
                <th className="p-3">Endpoint Path</th>
                <th className="p-3">Description</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 font-mono">
              {endpoints.map((ep, idx) => (
                <tr key={idx} className="hover:bg-white/[0.02] transition-colors">
                  <td className="p-3">
                    <span className={`px-2 py-0.5 rounded font-bold text-[11px] ${
                      ep.method === "GET" ? "bg-emerald-500/20 text-emerald-300" :
                      ep.method === "POST" ? "bg-cyan-500/20 text-cyan-300" :
                      ep.method === "DELETE" ? "bg-rose-500/20 text-rose-300" :
                      "bg-purple-500/20 text-purple-300"
                    }`}>
                      {ep.method}
                    </span>
                  </td>
                  <td className="p-3 text-cyan-300 font-semibold">{ep.path}</td>
                  <td className="p-3 text-slate-300 font-sans text-xs">{ep.desc}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
