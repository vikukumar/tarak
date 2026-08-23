import React, { useState } from "react";
import { Copy, Check, Cloud, Shield, Globe, ArrowRight, Zap, Network, Radio } from "lucide-react";

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
          Auto-Discovery & Edge Ingress
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          Inbuilt <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-indigo-300 to-purple-400">Tunnels, CNI & Mesh</span>
        </h1>
        <p className="text-slate-400 max-w-xl mx-auto text-sm sm:text-base">
          Zero-config host auto-detection for Cloudflare Tunnels & Tailscale WireGuard mesh, plus high-performance native CNI networking.
        </p>
      </div>

      {/* Auto-Detection Highlight Banner */}
      <div className="p-6 rounded-2xl bg-gradient-to-r from-purple-950/40 via-slate-900/80 to-cyan-950/40 border border-purple-500/30 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg sm:text-xl font-bold text-white flex items-center gap-2">
            <Zap className="text-amber-400" size={20} />
            <span>⚡ Host Auto-Detection & Automated Cluster Sync</span>
          </h2>
          <span className="px-2.5 py-0.5 rounded-full bg-emerald-500/15 border border-emerald-500/30 text-emerald-300 text-xs font-semibold">
            Zero-Config Enabled
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Tarak automatically inspects your host machine for existing <code className="text-cyan-300 font-mono">cloudflared</code> or <code className="text-purple-300 font-mono">tailscale</code> binaries, daemon sockets, CGNAT <code className="text-purple-300 font-mono">100.64.0.0/10</code> interfaces, and environment tokens. When detected:
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2">
          <div className="p-4 rounded-xl bg-[#04060c]/80 border border-white/10 space-y-2">
            <div className="text-cyan-300 font-bold text-xs uppercase flex items-center gap-1.5">
              <Cloud size={14} /> Cloudflare Auto-Register
            </div>
            <p className="text-xs text-slate-400">
              Instantly activates public HTTP ingress through Cloudflare Quick/Named Tunnels with automatic SSL without touching router ports.
            </p>
          </div>

          <div className="p-4 rounded-xl bg-[#04060c]/80 border border-white/10 space-y-2">
            <div className="text-purple-300 font-bold text-xs uppercase flex items-center gap-1.5">
              <Shield size={14} /> Tailscale WireGuard Auto-Mesh
            </div>
            <p className="text-xs text-slate-400">
              Joins private Tailnet, registers MagicDNS, and propagates secure encrypted node-to-node tunnels to newly registered cluster nodes.
            </p>
          </div>
        </div>
      </div>

      {/* Inbuilt CNI Section */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg sm:text-xl font-bold text-white flex items-center gap-2">
            <Network className="text-cyan-400" size={20} />
            <span>1. Inbuilt Container Network Interface (CNI)</span>
          </h2>
          <span className="px-2.5 py-0.5 rounded-full bg-cyan-500/15 border border-cyan-500/30 text-cyan-300 text-xs font-semibold">
            Native IPAM & Virtual Bridge
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Tarak includes a zero-dependency CNI engine providing deterministic IP address management, pod-to-pod routing, CoreDNS resolution, and service proxies:
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 font-mono text-xs">
          <div className="p-3 rounded-lg bg-[#04060c] border border-white/10">
            <span className="text-slate-400 block text-[10px] uppercase font-sans">Pod CIDR</span>
            <span className="text-cyan-300 font-bold">10.244.0.0/16</span>
            <span className="text-slate-400 text-[11px] block mt-1">/24 per worker node</span>
          </div>
          <div className="p-3 rounded-lg bg-[#04060c] border border-white/10">
            <span className="text-slate-400 block text-[10px] uppercase font-sans">Service CIDR</span>
            <span className="text-purple-300 font-bold">10.96.0.0/12</span>
            <span className="text-slate-400 text-[11px] block mt-1">ClusterIP VIP range</span>
          </div>
          <div className="p-3 rounded-lg bg-[#04060c] border border-white/10">
            <span className="text-slate-400 block text-[10px] uppercase font-sans">CoreDNS Server</span>
            <span className="text-emerald-300 font-bold">10.96.0.10</span>
            <span className="text-slate-400 text-[11px] block mt-1">Internal service discovery</span>
          </div>
        </div>
      </div>

      {/* Cloudflare Tunnels */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg sm:text-xl font-bold text-white flex items-center gap-2">
            <Cloud className="text-cyan-400" size={20} />
            <span>2. Cloudflare Ingress Tunnels</span>
          </h2>
          <span className="px-2.5 py-0.5 rounded-full bg-cyan-500/15 border border-cyan-500/30 text-cyan-300 text-xs font-semibold">
            Public Zero-Port Forwarding
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Expose your cluster endpoints to the public internet instantly without port forwarding, dynamic DNS, or public static IPs:
        </p>

        <div className="relative rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400 font-mono">
            <span>Cloudflare Ingress Commands</span>
            <button
              onClick={() => copyCode("# Quick ephemeral tunnel (instant free trycloudflare URL):\ntarak server --cloudflare-tunnel\n\n# Named persistent tunnel (custom domain):\ntarak server --cloudflare-tunnel --cloudflare-token <YOUR_TOKEN>", 1)}
              className="flex items-center gap-1 text-slate-300 hover:text-cyan-300"
            >
              {copiedIndex === 1 ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
              <span>{copiedIndex === 1 ? "Copied" : "Copy"}</span>
            </button>
          </div>
          <pre className="p-4 font-mono text-xs text-cyan-300 overflow-x-auto leading-relaxed whitespace-pre">
{`# Quick ephemeral tunnel (instant free trycloudflare URL):
tarak server --cloudflare-tunnel

# Named persistent tunnel (custom domain):
tarak server --cloudflare-tunnel --cloudflare-token <YOUR_TOKEN>`}
          </pre>
        </div>
      </div>

      {/* Tailscale Private Mesh */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg sm:text-xl font-bold text-white flex items-center gap-2">
            <Shield className="text-purple-400" size={20} />
            <span>3. Tailscale Private WireGuard Mesh</span>
          </h2>
          <span className="px-2.5 py-0.5 rounded-full bg-purple-500/15 border border-purple-500/30 text-purple-300 text-xs font-semibold">
            Zero-Trust MagicDNS
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Create an end-to-end encrypted mesh interconnecting distributed cloud VMs, home labs, and edge nodes on your private Tailnet:
        </p>

        <div className="relative rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400 font-mono">
            <span>Tailscale Mesh Activation</span>
            <button
              onClick={() => copyCode("# Start control plane with Tailscale mesh:\ntarak server --tailscale --tailscale-authkey <tskey-auth-...>\n\n# Or configure worker agent to join through Tailscale IP:\ntaraks --server https://100.x.y.z:6443 --token <cluster-token>", 2)}
              className="flex items-center gap-1 text-slate-300 hover:text-cyan-300"
            >
              {copiedIndex === 2 ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
              <span>{copiedIndex === 2 ? "Copied" : "Copy"}</span>
            </button>
          </div>
          <pre className="p-4 font-mono text-xs text-purple-300 overflow-x-auto leading-relaxed whitespace-pre">
{`# Start control plane with Tailscale mesh:
tarak server --tailscale --tailscale-authkey <tskey-auth-...>

# Or configure worker agent to join through Tailscale IP:
taraks --server https://100.x.y.z:6443 --token <cluster-token>`}
          </pre>
        </div>
      </div>
    </div>
  );
};
