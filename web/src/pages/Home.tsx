import React, { useState, useEffect } from "react";
import {
  Copy,
  Check,
  ArrowRight,
  Shield,
  Zap,
  Cpu,
  Network,
  Database,
  Terminal,
  Layers,
  Sparkles,
  ExternalLink,
  Download,
  Server,
  Activity,
  CheckCircle2,
  XCircle,
  Clock,
  HardDrive
} from "lucide-react";
import defaultReleases from "../../public/data/releases.json";

interface HomeProps {
  onNavigate: (tab: string) => void;
}

export const HomePage: React.FC<HomeProps> = ({ onNavigate }) => {
  const [activeTab, setActiveTab] = useState<"bash" | "powershell" | "go" | "helm">("bash");
  const [copied, setCopied] = useState(false);
  const [latestRelease, setLatestRelease] = useState(defaultReleases[0] || {
    tag: "v1.0.6",
    version: "1.0.6",
    name: "Tarak v1.0.6",
    date: "2026-08-23",
  });

  useEffect(() => {
    async function fetchLatest() {
      try {
        const res = await fetch("https://api.github.com/repos/vikukumar/tarak/releases/latest");
        if (res.ok) {
          const gh = await res.json();
          setLatestRelease({
            tag: gh.tag_name || "v1.0.6",
            version: (gh.tag_name || "v1.0.6").replace(/^v/, ""),
            name: gh.name || `Tarak ${gh.tag_name}`,
            date: gh.published_at ? gh.published_at.substring(0, 10) : "2026-08-23",
          });
        }
      } catch {
        // Fallback to bundled releases
      }
    }
    fetchLatest();
  }, []);

  const installCommands = {
    bash: "curl -fsSL https://tarak.vikshro.in/install.sh | bash",
    powershell: "irm https://tarak.vikshro.in/install.ps1 | iex",
    go: "go install github.com/vikukumar/tarak/cmd/tarak@latest",
    helm: "tarakctl helm install my-app ./chart",
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(installCommands[activeTab]);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const comparisons = [
    {
      feature: "Memory Overhead (Idle RAM)",
      tarak: "~22 - 30 MB",
      k8s: "2,048 - 4,096 MB (2-4 GB+)",
      k3s: "512 - 1,024 MB",
      docker: "500 - 1,200 MB (Desktop/Swarm)",
      nomad: "180 - 350 MB",
      advantage: "95% less RAM than K8s",
    },
    {
      feature: "Cold Startup Time",
      tarak: "< 100 ms",
      k8s: "45 - 90 seconds",
      k3s: "15 - 30 seconds",
      docker: "5 - 10 seconds",
      nomad: "3 - 8 seconds",
      advantage: "Instant sub-second bootstrap",
    },
    {
      feature: "Binary Architecture",
      tarak: "Single ~30MB Go Binary",
      k8s: "10+ separate daemons & binaries",
      k3s: "Packaged distribution (~120MB)",
      docker: "Multi-daemon (dockerd, containerd, runc)",
      nomad: "Single binary (~80MB)",
      advantage: "100% self-contained",
    },
    {
      feature: "Bare Metal Windows & macOS",
      tarak: "Native (Zero WSL / Zero VM)",
      k8s: "Requires Linux VM / WSL2",
      k3s: "Requires Linux VM / WSL2",
      docker: "Heavy Hyper-V / WSL2 VM",
      nomad: "Native (raw exec only)",
      advantage: "Zero Hypervisor / VM needed",
    },
    {
      feature: "Built-in Cloudflare & Tailscale Tunnels",
      tarak: "Native 1-Click Zero-Trust Ingress",
      k8s: "Requires Cloud LB + External Ingress + Tunnels",
      k3s: "Requires Traefik + Manual Tunnels",
      docker: "Manual Port Forwarding / Ngrok",
      nomad: "Manual Consul / Ingress Setup",
      advantage: "Global public URL in seconds",
    },
    {
      feature: "Built-in Service Mesh & Canary Routing",
      tarak: "Inbuilt mTLS + ProxyPatch + Routes",
      k8s: "Heavyweight Istio / Linkerd / Kuma (500MB+)",
      k3s: "Requires external Istio/Linkerd",
      docker: "None (Basic overlay network)",
      nomad: "Requires Consul Connect",
      advantage: "Zero-cost microservice mesh",
    },
    {
      feature: "Zero-Trust Kyverno Policy Engine",
      tarak: "Inbuilt validation & violation reports",
      k8s: "Requires Kyverno / OPA Gatekeeper Operator",
      k3s: "Requires Kyverno Operator",
      docker: "None",
      nomad: "Sentinel (Enterprise only)",
      advantage: "Built-in compliance & audit",
    },
    {
      feature: "State & Storage Engine",
      tarak: "Embedded ACID bbolt KV Store",
      k8s: "Heavy distributed etcd cluster",
      k3s: "Embedded SQLite / Kine",
      docker: "Local dockerd state",
      nomad: "Raft consensus store",
      advantage: "Zero-maintenance instant storage",
    },
    {
      feature: "Built-in Real-Time Dashboard UI",
      tarak: "Built-in Next.js 16 Reactive Web UI",
      k8s: "Separate Dashboard Helm + Token setup",
      k3s: "None (Requires external tools)",
      docker: "Docker Desktop / Portainer",
      nomad: "Built-in UI",
      advantage: "49+ unified workload screens",
    },
    {
      feature: "Kubernetes Manifest Compatibility",
      tarak: "Full standard Pods, Deployments, Services, PDB, HPA",
      k8s: "Native",
      k3s: "Native",
      docker: "Docker Compose only",
      nomad: "HCL Job specs only",
      advantage: "Use standard kubectl & Helm",
    },
  ];

  return (
    <div className="space-y-20 animate-fade-in">
      {/* ─── Hero Section ──────────────────────────────────────────────────────── */}
      <section className="text-center space-y-7 pt-4 pb-4">
        {/* Badges */}
        <div className="flex flex-wrap items-center justify-center gap-2">
          <a
            href="#releases"
            onClick={() => onNavigate("releases")}
            className="inline-flex items-center gap-1.5 px-3.5 py-1 rounded-full bg-cyan-500/10 hover:bg-cyan-500/20 border border-cyan-500/30 text-cyan-300 text-xs font-bold transition-all shadow-[0_0_15px_rgba(0,240,255,0.2)]"
          >
            <Sparkles size={13} className="text-cyan-400 animate-pulse" />
            <span>⚡ {latestRelease.tag} Production Ready</span>
          </a>
          <span className="px-3.5 py-1 rounded-full bg-purple-500/10 border border-purple-500/30 text-purple-300 text-xs font-bold">
            Zero-Dependency Pure Go
          </span>
          <span className="px-3.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-xs font-bold">
            OCI Native Sandbox
          </span>
          <span className="px-3.5 py-1 rounded-full bg-amber-500/10 border border-amber-500/30 text-amber-300 text-xs font-bold">
            Native Windows, Linux & macOS
          </span>
        </div>

        {/* Title */}
        <h1 className="text-4xl sm:text-6xl md:text-7xl font-black text-white tracking-tight leading-tight sm:leading-none">
          Ultra-Lightweight <br />
          <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-indigo-300 to-purple-400">
            Kubernetes Alternative
          </span>
        </h1>

        {/* Description */}
        <p className="text-slate-300 max-w-3xl mx-auto text-sm sm:text-lg leading-relaxed">
          TARAK is a next-generation pure Go container orchestrator engineered for extreme performance.
          <span className="text-white block mt-1 font-semibold">
            10x faster & 20x lighter than K8s and K3s with built-in Cloudflare Named Tunnels, Tailscale WireGuard Mesh, and zero external daemons.
          </span>
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
              {(["bash", "powershell", "go", "helm"] as const).map((type) => (
                <button
                  key={type}
                  onClick={() => setActiveTab(type)}
                  className={`px-2.5 py-1 rounded-md text-[11px] font-bold font-mono transition-all ${
                    activeTab === type
                      ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/30 shadow-[0_0_10px_rgba(0,240,255,0.25)]"
                      : "text-slate-400 hover:text-white"
                  }`}
                >
                  {type === "bash" ? "Linux / macOS" : type === "powershell" ? "Windows PS" : type === "go" ? "Go SDK" : "Helm"}
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

        {/* CTA Buttons */}
        <div className="flex flex-wrap items-center justify-center gap-3 pt-2">
          <button
            onClick={() => onNavigate("getting-started")}
            className="px-6 py-3 rounded-xl bg-gradient-to-r from-cyan-400 to-indigo-500 text-slate-950 font-bold text-sm shadow-[0_0_25px_rgba(0,240,255,0.35)] hover:shadow-[0_0_35px_rgba(0,240,255,0.55)] hover:scale-105 transition-all flex items-center gap-2"
          >
            <span>⚡ Quickstart in 30 Seconds</span>
            <ArrowRight size={16} />
          </button>
          <button
            onClick={() => onNavigate("architecture")}
            className="px-6 py-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/15 text-white font-semibold text-sm hover:border-cyan-500/40 transition-all flex items-center gap-2"
          >
            <Layers size={16} className="text-cyan-400" />
            <span>Architecture & Design</span>
          </button>
          <a
            href={`https://github.com/vikukumar/tarak/releases/tag/${latestRelease.tag}`}
            target="_blank"
            rel="noreferrer"
            className="px-5 py-3 rounded-xl bg-slate-900/80 hover:bg-slate-800 border border-white/10 text-slate-300 hover:text-white font-medium text-sm transition-all flex items-center gap-2"
          >
            <Download size={15} className="text-purple-400" />
            <span>Download {latestRelease.tag}</span>
            <ExternalLink size={12} />
          </a>
        </div>
      </section>

      {/* ─── Metric Highlights ─────────────────────────────────────────────────── */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-2xl bg-[#080c1a]/70 border border-white/10 backdrop-blur-xl space-y-1">
          <div className="flex items-center justify-between text-cyan-400">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Memory Footprint</span>
            <Zap size={18} />
          </div>
          <p className="text-3xl font-black text-white">~22 MB</p>
          <p className="text-[11px] text-slate-400 font-medium">95% less RAM than Kubernetes</p>
        </div>

        <div className="p-5 rounded-2xl bg-[#080c1a]/70 border border-white/10 backdrop-blur-xl space-y-1">
          <div className="flex items-center justify-between text-purple-400">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Cold Start</span>
            <Clock size={18} />
          </div>
          <p className="text-3xl font-black text-white">&lt; 100 ms</p>
          <p className="text-[11px] text-slate-400 font-medium">Instant cluster bootstrap</p>
        </div>

        <div className="p-5 rounded-2xl bg-[#080c1a]/70 border border-white/10 backdrop-blur-xl space-y-1">
          <div className="flex items-center justify-between text-emerald-400">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">External Daemons</span>
            <CheckCircle2 size={18} />
          </div>
          <p className="text-3xl font-black text-white">0 (Zero)</p>
          <p className="text-[11px] text-slate-400 font-medium">No Docker, WSL, or VM required</p>
        </div>

        <div className="p-5 rounded-2xl bg-[#080c1a]/70 border border-white/10 backdrop-blur-xl space-y-1">
          <div className="flex items-center justify-between text-amber-400">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Binary Size</span>
            <HardDrive size={18} />
          </div>
          <p className="text-3xl font-black text-white">~30 MB</p>
          <p className="text-[11px] text-slate-400 font-medium">Self-contained Go static binary</p>
        </div>
      </div>

      {/* ─── Detailed Comparison Matrix ────────────────────────────────────────── */}
      <section className="space-y-6">
        <div className="text-center space-y-3">
          <span className="px-3.5 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs font-bold uppercase tracking-wider">
            In-Depth Benchmarks & Comparison
          </span>
          <h2 className="text-3xl sm:text-4xl font-extrabold text-white tracking-tight">
            How TARAK Compares to Alternatives
          </h2>
          <p className="text-slate-400 max-w-2xl mx-auto text-xs sm:text-sm">
            Architectural comparison between TARAK, full Kubernetes (K8s), K3s, Docker / Docker Swarm, and HashiCorp Nomad.
          </p>
        </div>

        <div className="overflow-x-auto rounded-2xl border border-white/10 bg-[#050814]/80 backdrop-blur-2xl shadow-2xl">
          <table className="w-full text-left text-xs sm:text-sm border-collapse min-w-[700px]">
            <thead>
              <tr className="border-b border-white/10 bg-white/[0.03] text-slate-300">
                <th className="p-4 font-bold text-white">Capability & Metric</th>
                <th className="p-4 font-black text-cyan-300 bg-cyan-500/10 border-x border-cyan-500/20">
                  ⚡ TARAK
                </th>
                <th className="p-4 font-semibold text-slate-300">Kubernetes (K8s)</th>
                <th className="p-4 font-semibold text-slate-300">K3s</th>
                <th className="p-4 font-semibold text-slate-300">Docker / Swarm</th>
                <th className="p-4 font-semibold text-slate-300">Nomad</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 text-slate-300">
              {comparisons.map((row, idx) => (
                <tr key={idx} className="hover:bg-white/[0.02] transition-colors">
                  <td className="p-4 font-medium text-white">
                    <div>{row.feature}</div>
                    <span className="text-[10px] text-cyan-400 font-semibold">{row.advantage}</span>
                  </td>
                  <td className="p-4 font-bold text-cyan-300 bg-cyan-500/[0.04] border-x border-cyan-500/20 font-mono text-xs">
                    {row.tarak}
                  </td>
                  <td className="p-4 text-slate-400 font-mono text-xs">{row.k8s}</td>
                  <td className="p-4 text-slate-400 font-mono text-xs">{row.k3s}</td>
                  <td className="p-4 text-slate-400 font-mono text-xs">{row.docker}</td>
                  <td className="p-4 text-slate-400 font-mono text-xs">{row.nomad}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {/* ─── Architectural Pillars ─────────────────────────────────────────────── */}
      <section className="space-y-8">
        <div className="text-center space-y-3">
          <span className="px-3.5 py-1 rounded-full bg-purple-500/10 border border-purple-500/30 text-purple-300 text-xs font-bold uppercase tracking-wider">
            Engineered For Simplicity & Power
          </span>
          <h2 className="text-3xl sm:text-4xl font-extrabold text-white tracking-tight">
            Six Pillars of TARAK Architecture
          </h2>
          <p className="text-slate-400 max-w-xl mx-auto text-xs sm:text-sm">
            Everything you need for modern edge, dev, and production microservices built directly into a single runtime.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {/* Card 1 */}
          <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 hover:border-cyan-500/30 transition-all space-y-3">
            <div className="w-10 h-10 rounded-xl bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
              <Zap size={20} />
            </div>
            <h3 className="text-base font-bold text-white">1. TCR Native Container Engine</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Extracts OCI image layers and executes native processes with realistic POSIX sandboxing. Runs on Windows, Linux, and macOS without Docker Desktop, WSL2, or Linux VMs.
            </p>
          </div>

          {/* Card 2 */}
          <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 hover:border-purple-500/30 transition-all space-y-3">
            <div className="w-10 h-10 rounded-xl bg-purple-500/15 border border-purple-500/30 flex items-center justify-center text-purple-400">
              <Network size={20} />
            </div>
            <h3 className="text-base font-bold text-white">2. Virtual Pod Bridge & Hubble Telemetry</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Inbuilt Layer 2/3 virtual bridge driver (<code className="text-cyan-300 font-mono">10.244.0.0/16</code>) with MetalLB virtual IP pools, ClusterIP routing, and live Hubble telemetry stream.
            </p>
          </div>

          {/* Card 3 */}
          <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 hover:border-emerald-500/30 transition-all space-y-3">
            <div className="w-10 h-10 rounded-xl bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
              <Shield size={20} />
            </div>
            <h3 className="text-base font-bold text-white">3. Zero-Trust Security & Kyverno Engine</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Automatic P-256 ECDSA mTLS Certificate Authority, RBAC authorization matrix, and Kyverno-compatible policy evaluation with live compliance reports.
            </p>
          </div>

          {/* Card 4 */}
          <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 hover:border-amber-500/30 transition-all space-y-3">
            <div className="w-10 h-10 rounded-xl bg-amber-500/15 border border-amber-500/30 flex items-center justify-center text-amber-400">
              <Layers size={20} />
            </div>
            <h3 className="text-base font-bold text-white">4. Cloudflare Tunnels & Tailscale Mesh</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Expose workloads to the public internet securely with 1-click Cloudflare Named Tunnels and create cross-cloud hybrid clusters over encrypted Tailscale WireGuard mesh.
            </p>
          </div>

          {/* Card 5 */}
          <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 hover:border-sky-500/30 transition-all space-y-3">
            <div className="w-10 h-10 rounded-xl bg-sky-500/15 border border-sky-500/30 flex items-center justify-center text-sky-400">
              <Database size={20} />
            </div>
            <h3 className="text-base font-bold text-white">5. GitOps CD & Helm Management</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Continuous delivery sync engine with auto-reconciliation (ArgoCD equivalent) and native Helm release manager supporting chart rollouts, history, and status.
            </p>
          </div>

          {/* Card 6 */}
          <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 hover:border-rose-500/30 transition-all space-y-3">
            <div className="w-10 h-10 rounded-xl bg-rose-500/15 border border-rose-500/30 flex items-center justify-center text-rose-400">
              <Terminal size={20} />
            </div>
            <h3 className="text-base font-bold text-white">6. Interactive Terminal & 49+ UI Screens</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              Embedded Next.js 16 Web Dashboard with interactive multi-container Web Terminal, live log streams, CRD schema inspector, and cluster topology visualizer.
            </p>
          </div>
        </div>
      </section>

      {/* ─── Bottom CTA Banner ─────────────────────────────────────────────────── */}
      <section className="p-8 sm:p-12 rounded-3xl bg-gradient-to-br from-cyan-950/40 via-slate-950 to-purple-950/40 border border-cyan-500/30 shadow-[0_0_50px_rgba(0,240,255,0.15)] text-center space-y-5">
        <h2 className="text-2xl sm:text-4xl font-extrabold text-white">
          Ready to run Kubernetes workloads with 95% less overhead?
        </h2>
        <p className="text-slate-300 max-w-xl mx-auto text-xs sm:text-sm">
          Download the latest {latestRelease.tag} binary or install in seconds with a single shell command.
        </p>
        <div className="flex flex-wrap items-center justify-center gap-3 pt-2">
          <button
            onClick={() => onNavigate("getting-started")}
            className="px-6 py-3 rounded-xl bg-cyan-400 text-slate-950 font-bold text-sm shadow-[0_0_20px_rgba(0,240,255,0.4)] hover:bg-cyan-300 transition-all flex items-center gap-2"
          >
            <span>⚡ Deploy Your First App</span>
            <ArrowRight size={15} />
          </button>
          <a
            href="https://github.com/vikukumar/tarak"
            target="_blank"
            rel="noreferrer"
            className="px-6 py-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/15 text-white font-semibold text-sm transition-all"
          >
            ⭐ View on GitHub
          </a>
        </div>
      </section>
    </div>
  );
};
