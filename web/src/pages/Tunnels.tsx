import React, { useState } from "react";
import { Copy, Check, Cloud, Shield, Globe, ArrowRight } from "lucide-react";

export const TunnelsPage: React.FC = () => {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const copyCode = (code: string, idx: number) => {
    navigator.clipboard.writeText(code);
    setCopiedIndex(idx);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  return (
    <div className="space-y-10 animate-fade-in max-w-4xl mx-auto">
      {/* Title */}
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs font-bold uppercase tracking-wider">
          Edge Ingress & Private Mesh
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          Inbuilt <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-indigo-300 to-purple-400">Tunnels & Ingress</span>
        </h1>
        <p className="text-slate-400 max-w-xl mx-auto text-sm sm:text-base">
          Zero-config public HTTPS endpoints via Cloudflare Tunnels and secure private cluster interconnects via Tailscale MagicDNS.
        </p>
      </div>

      {/* Cloudflare Tunnels */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg sm:text-xl font-bold text-white flex items-center gap-2">
            <Cloud className="text-cyan-400" size={20} />
            <span>1. Inbuilt Cloudflare Tunnels</span>
          </h2>
          <span className="px-2.5 py-0.5 rounded-full bg-cyan-500/15 border border-cyan-500/30 text-cyan-300 text-xs font-semibold">
            Public Zero-Port Forwarding
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Expose your cluster endpoints to the public internet instantly without port forwarding, dynamic DNS, or public static IPs:
        </p>

        <div className="relative rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400">
            <span className="font-mono">Cloudflare Ingress Commands</span>
            <button
              onClick={() => copyCode("# Quick ephemeral tunnel (instant free trycloudflare URL):\ntarak --cloudflare-tunnel\n\n# Named persistent tunnel (custom domain):\ntarak --cloudflare-tunnel --cloudflare-token <YOUR_CLOUDFLARE_TUNNEL_TOKEN>", 1)}
              className="flex items-center gap-1 text-slate-300 hover:text-cyan-300"
            >
              {copiedIndex === 1 ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
              <span>{copiedIndex === 1 ? "Copied" : "Copy"}</span>
            </button>
          </div>
          <pre className="p-4 font-mono text-xs text-cyan-300 overflow-x-auto leading-relaxed whitespace-pre">
{`# Quick ephemeral tunnel (instant free trycloudflare URL):
tarak --cloudflare-tunnel

# Named persistent tunnel (custom domain):
tarak --cloudflare-tunnel --cloudflare-token <YOUR_CLOUDFLARE_TUNNEL_TOKEN>`}
          </pre>
        </div>
      </div>

      {/* Tailscale Private Mesh */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg sm:text-xl font-bold text-white flex items-center gap-2">
            <Shield className="text-purple-400" size={20} />
            <span>2. Tailscale Zero-Trust Private Mesh</span>
          </h2>
          <span className="px-2.5 py-0.5 rounded-full bg-purple-500/15 border border-purple-500/30 text-purple-300 text-xs font-semibold">
            Private WireGuard
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Securely join your cluster node directly into your private Tailscale Tailnet mesh:
        </p>

        <div className="relative rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400">
            <span className="font-mono">Tailscale Mesh Command</span>
            <button
              onClick={() => copyCode("tarak --tailscale --tailscale-authkey <tskey-auth-...>", 2)}
              className="flex items-center gap-1 text-slate-300 hover:text-cyan-300"
            >
              {copiedIndex === 2 ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
              <span>{copiedIndex === 2 ? "Copied" : "Copy"}</span>
            </button>
          </div>
          <pre className="p-4 font-mono text-xs text-purple-300 overflow-x-auto leading-relaxed">
            tarak --tailscale --tailscale-authkey &lt;tskey-auth-...&gt;
          </pre>
        </div>
      </div>

      {/* Ingress Resource Example */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <h2 className="text-lg sm:text-xl font-bold text-white flex items-center gap-2">
          <Globe className="text-emerald-400" size={20} />
          <span>3. Declarative Ingress Resource</span>
        </h2>

        <pre className="p-4 rounded-xl bg-[#04060c] border border-white/10 font-mono text-xs text-slate-300 overflow-x-auto leading-relaxed">
{`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-ingress
  namespace: default
spec:
  ingressClassName: tarak-cloudflare
  rules:
  - host: app.vikshro.in
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-service
            port:
              number: 80`}
        </pre>
      </div>
    </div>
  );
};
