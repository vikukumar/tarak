import React, { useState } from "react";
import { Copy, Check, ArrowRight, Shield, Zap, Cpu, Network, Database, Terminal } from "lucide-react";

interface HomeProps {
  onNavigate: (tab: string) => void;
}

export const HomePage: React.FC<HomeProps> = ({ onNavigate }) => {
  const [activeTab, setActiveTab] = useState<"bash" | "powershell" | "go">("bash");
  const [copied, setCopied] = useState(false);

  const installCommands = {
    bash: "curl -fsSL https://tarak.vikshro.in/install.sh | bash",
    powershell: "irm https://tarak.vikshro.in/install.ps1 | iex",
    go: "go get -u github.com/vikukumar/tarak/pkg/client@latest",
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(installCommands[activeTab]);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="space-y-16 animate-fade-in">
      {/* Hero Section */}
      <section className="text-center space-y-6 pt-4 pb-2">
        <div className="flex flex-wrap items-center justify-center gap-2">
          <span className="px-3.5 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs font-bold">
            ⚡ v1.0.6 Production Ready
          </span>
          <span className="px-3.5 py-1 rounded-full bg-purple-500/10 border border-purple-500/30 text-purple-300 text-xs font-bold">
            Zero-Dependency Pure Go
          </span>
          <span className="px-3.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-xs font-bold">
            OCI Native Sandbox
          </span>
        </div>

        <h1 className="text-4xl sm:text-6xl font-black text-white tracking-tight leading-tight">
          Ultra-Lightweight <br />
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-indigo-300 to-purple-400">
            Kubernetes Alternative
          </span>
        </h1>

        <p className="text-slate-300 max-w-2xl mx-auto text-sm sm:text-base leading-relaxed">
          A pure Go container orchestrator that runs anywhere with zero external dependencies.
          <strong className="text-white block mt-1 font-semibold">
            10x faster & 20x lighter than K8s and K3s with built-in Cloudflare Tunnels and Tailscale Mesh.
          </strong>
        </p>

        {/* Installer Box */}
        <div className="max-w-xl mx-auto rounded-2xl bg-[#04060c]/90 border border-white/10 shadow-2xl overflow-hidden text-left">
          <div className="flex items-center justify-between p-3 bg-slate-950 border-b border-white/10">
            <div className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-full bg-rose-500/80" />
              <span className="w-2.5 h-2.5 rounded-full bg-amber-500/80" />
              <span className="w-2.5 h-2.5 rounded-full bg-emerald-500/80" />
            </div>

            <div className="flex items-center gap-1">
              {(["bash", "powershell", "go"] as const).map((type) => (
                <button
                  key={type}
                  onClick={() => setActiveTab(type)}
                  className={`px-2.5 py-1 rounded-md text-[11px] font-bold font-mono transition-all ${
                    activeTab === type
                      ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/30"
                      : "text-slate-400 hover:text-white"
                  }`}
                >
                  {type === "bash" ? "Linux / macOS" : type === "powershell" ? "Windows PS" : "Go SDK"}
                </button>
              ))}
            </div>
          </div>

          <div className="p-4 flex items-center justify-between gap-3">
            <code className="text-xs sm:text-sm font-mono text-cyan-300 break-all select-all">
              {installCommands[activeTab]}
            </code>
            <button
              onClick={handleCopy}
              className="p-2 rounded-lg bg-white/5 hover:bg-white/10 border border-white/15 text-slate-300 hover:text-white transition-all flex-shrink-0"
              title="Copy to clipboard"
            >
              {copied ? <Check size={16} className="text-emerald-400" /> : <Copy size={16} />}
            </button>
          </div>
        </div>

        {/* Buttons */}
        <div className="flex flex-wrap items-center justify-center gap-3 pt-2">
          <button
            onClick={() => onNavigate("getting-started")}
            className="px-6 py-3 rounded-xl bg-gradient-to-r from-cyan-400 to-indigo-500 text-slate-950 font-bold text-sm shadow-[0_0_25px_rgba(0,240,255,0.35)] hover:shadow-[0_0_35px_rgba(0,240,255,0.55)] transition-all flex items-center gap-2"
          >
            <span>⚡ Get Started in 30 Seconds</span>
            <ArrowRight size={15} />
          </button>
          <button
            onClick={() => onNavigate("architecture")}
            className="px-6 py-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/15 text-white font-semibold text-sm transition-all"
          >
            📐 Explore Architecture
          </button>
        </div>
      </section>

      {/* Feature Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-5">
        <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 space-y-3">
          <div className="w-10 h-10 rounded-xl bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
            <Zap size={20} />
          </div>
          <h3 className="text-base font-bold text-white">Sub-180ms Cold Start</h3>
          <p className="text-xs text-slate-400 leading-relaxed">
            Zero JVM or heavy container runtime overhead. Instant bootstrap from a single binary with sub-22MB idle RAM.
          </p>
        </div>

        <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 space-y-3">
          <div className="w-10 h-10 rounded-xl bg-purple-500/15 border border-purple-500/30 flex items-center justify-center text-purple-400">
            <Network size={20} />
          </div>
          <h3 className="text-base font-bold text-white">Zero-Config Mesh</h3>
          <p className="text-xs text-slate-400 leading-relaxed">
            Expose workloads globally via Cloudflare Tunnels and connect hybrid clusters with encrypted Tailscale WireGuard mesh.
          </p>
        </div>

        <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 space-y-3">
          <div className="w-10 h-10 rounded-xl bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
            <Shield size={20} />
          </div>
          <h3 className="text-base font-bold text-white">Automatic mTLS Security</h3>
          <p className="text-xs text-slate-400 leading-relaxed">
            Inbuilt ECDSA P-256 Certificate Authority automatically encrypts and validates all node-to-node traffic.
          </p>
        </div>
      </div>
    </div>
  );
};
