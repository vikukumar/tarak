import React, { useState } from "react";
import { Terminal, Copy, Check, Server, Shield, Cpu, Layers, HardDrive, Play, Zap } from "lucide-react";

export const CliReferencePage: React.FC = () => {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const copyCode = (code: string, idx: number) => {
    navigator.clipboard.writeText(code);
    setCopiedIndex(idx);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  const commandGroups = [
    {
      group: "1. tarak init — Cluster Bootstrap & PKI Generator",
      desc: "Generates the ECDSA P-256 root CA, signs server/client certificates, provisions state directories, and emits an admin kubeconfig.",
      syntax: "tarak init [flags]",
      example: "tarak init --data-dir=/var/lib/tarak --bind-address=0.0.0.0:6443 --tls-san=cluster.internal,10.0.0.1 --kubeconfig=/etc/tarak/admin.kubeconfig",
      flags: [
        { flag: "--data-dir <path>", default: "~/.tarak/data", desc: "Directory to store cluster BoltDB state, PKI keys, and layer cache" },
        { flag: "--bind-address <host:port>", default: "127.0.0.1:6443", desc: "API server advertise address and port" },
        { flag: "--tls-san <list>", default: "localhost,127.0.0.1", desc: "Comma-separated Subject Alternative Names (SANs) for the TLS certificate" },
        { flag: "--kubeconfig <path>", default: "~/.tarak/config", desc: "Output file path for the cluster administrator kubeconfig" },
      ],
    },
    {
      group: "2. tarak server / tarakd — Production Control Plane Server",
      desc: "Starts the high-performance API server, declarative controllers, BoltDB transactional store, and ingress proxy.",
      syntax: "tarak server [flags]  OR  tarakd server [flags]",
      example: "tarak server --bind-address=0.0.0.0:6443 --ingress-http-addr=0.0.0.0:80 --cloudflare-tunnel --cpu-limit=8 --memory-limit=16Gi",
      flags: [
        { flag: "--bind-address <host:port>", default: "0.0.0.0:6443", desc: "Host address and port for the Kubernetes HTTPS API server" },
        { flag: "--ingress-http-addr <host:port>", default: "0.0.0.0:8080", desc: "Host address and port for the HTTP Ingress reverse proxy" },
        { flag: "--data-dir <path>", default: "~/.tarak/data", desc: "Persistent state directory for BoltDB database" },
        { flag: "--cloudflare-tunnel", default: "false", desc: "Instantly create a zero-config Cloudflare public ingress tunnel" },
        { flag: "--cloudflare-token <token>", default: "none", desc: "Cloudflare Named Tunnel credentials token for custom domain binding" },
        { flag: "--tailscale", default: "false", desc: "Enable Tailscale zero-trust private WireGuard mesh networking" },
        { flag: "--tailscale-authkey <key>", default: "none", desc: "Tailscale pre-authenticated node join key" },
        { flag: "--tls-san <list>", default: "localhost,127.0.0.1", desc: "Subject Alternative Names (IPs/domains) for TLS handshake" },
        { flag: "--cpu-limit <cores>", default: "100% host", desc: "Maximum CPU cores allocated to scheduled workloads" },
        { flag: "--memory-limit <bytes>", default: "100% host", desc: "Maximum RAM allocated to scheduled workloads (e.g. 16Gi)" },
        { flag: "--gpu-limit <count>", default: "all host GPUs", desc: "GPU devices limit allocated for AI/ML container execution" },
        { flag: "--log-level <level>", default: "info", desc: "Logging verbosity: debug | info | warn | error" },
        { flag: "--shutdown-timeout <duration>", default: "30s", desc: "Graceful drain deadline during process termination" },
        { flag: "--insecure", default: "false", desc: "Allow unauthenticated dev access (development only)" },
      ],
    },
    {
      group: "3. tarak agent / taraks — Worker Node Runtime Agent",
      desc: "Starts the local node execution agent, TCR container sandbox, and virtual pod bridge CNI.",
      syntax: "tarak agent [flags]  OR  taraks [flags]",
      example: "taraks --server=https://192.168.1.10:6443 --token=my-secret-token --node-name=worker-node-01 --data-dir=/var/lib/tarak",
      flags: [
        { flag: "--server <url>", default: "https://127.0.0.1:6443", desc: "Master control plane API server endpoint URL" },
        { flag: "--token <token>", default: "none", desc: "Cluster mutual TLS join authentication token" },
        { flag: "--node-name <string>", default: "hostname", desc: "Unique worker node identifier in the cluster" },
        { flag: "--data-dir <path>", default: "~/.tarak/data", desc: "Local container rootfs layers and overlay storage directory" },
        { flag: "--log-level <level>", default: "info", desc: "Logging verbosity: debug | info | warn | error" },
      ],
    },
    {
      group: "4. tarakctl / taraktl — Cluster Management CLI",
      desc: "CLI tool for declarative resource management, workload inspection, live log streaming, and container terminal exec.",
      syntax: "tarakctl <command> [options]",
      example: "tarakctl get pods -A -o wide",
      flags: [
        { flag: "get <resource>", default: "po, svc, no...", desc: "List active cluster objects (pods, nodes, services, deployments, crds)" },
        { flag: "apply -f <manifest.yaml>", default: "—", desc: "Create or update resources declaratively from YAML or JSON manifests" },
        { flag: "delete <kind> <name>", default: "—", desc: "Delete a resource from cluster state" },
        { flag: "logs <pod> [-f] [-c <container>]", default: "—", desc: "Stream real stdout and stderr logs for active pod containers" },
        { flag: "exec -it <pod> [-c <container>] -- <cmd>", default: "—", desc: "Spawn interactive terminal session inside a container sandbox" },
        { flag: "run <name> --image=<img policy>", default: "—", desc: "Imperatively launch a new containerized pod" },
        { flag: "tunnel list", default: "—", desc: "Display active Cloudflare and Tailscale tunnel ingress endpoints" },
        { flag: "version", default: "—", desc: "Display client and server version information" },
      ],
    },
  ];

  return (
    <div className="space-y-12 animate-fade-in max-w-5xl mx-auto">
      {/* Title */}
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/30 text-purple-300 text-xs font-bold uppercase tracking-wider">
          Complete CLI & Flag Manual
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          TARAK <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 via-indigo-300 to-cyan-400">Command Suite</span>
        </h1>
        <p className="text-slate-400 max-w-2xl mx-auto text-sm sm:text-base leading-relaxed">
          Exhaustive documentation for every command, flag, default value, and execution option in <code className="text-cyan-300 font-mono">tarak</code>, <code className="text-cyan-300 font-mono">tarakd</code>, <code className="text-purple-300 font-mono">taraks</code>, and <code className="text-purple-300 font-mono">tarakctl</code>.
        </p>
      </div>

      {/* Linux Background Daemon Guide */}
      <div className="p-6 rounded-2xl bg-gradient-to-r from-cyan-950/40 to-slate-900/80 border border-cyan-500/30 shadow-2xl space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-xl font-bold text-white flex items-center gap-2">
            <Server className="text-cyan-400" size={20} />
            <span>Linux Background Daemon & Systemd Service Guide</span>
          </h2>
          <span className="px-2.5 py-0.5 rounded-full bg-emerald-500/15 border border-emerald-500/30 text-emerald-300 text-xs font-semibold">
            Production 24/7 Setup
          </span>
        </div>

        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
          Run Tarak as an automatic systemd background daemon on Ubuntu, Debian, RHEL, or Rocky Linux:
        </p>

        <div className="relative rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
          <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400 font-mono">
            <span>/etc/systemd/system/tarak.service</span>
            <button
              onClick={() => copyCode(`sudo tee /etc/systemd/system/tarak.service << 'EOF'
[Unit]
Description=Tarak Container Orchestrator Control Plane
Documentation=https://tarak.vikshro.in/
After=network-online.target
Wants=network-online.target

[Service]
Type=exec
ExecStart=/usr/local/bin/tarak server --data-dir=/var/lib/tarak --bind-address=0.0.0.0:6443 --log-level=info
Restart=always
RestartSec=5s
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
KillMode=process

[Install]
WantedBy=multi-user.target
EOF

# Reload & start daemon:
sudo systemctl daemon-reload
sudo systemctl enable --now tarak
sudo systemctl status tarak`, 999)}
              className="flex items-center gap-1 text-slate-300 hover:text-cyan-300"
            >
              {copiedIndex === 999 ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
              <span>{copiedIndex === 999 ? "Copied" : "Copy"}</span>
            </button>
          </div>
          <pre className="p-4 font-mono text-xs text-cyan-300 overflow-x-auto leading-relaxed whitespace-pre">
{`sudo tee /etc/systemd/system/tarak.service << 'EOF'
[Unit]
Description=Tarak Container Orchestrator Control Plane
Documentation=https://tarak.vikshro.in/
After=network-online.target
Wants=network-online.target

[Service]
Type=exec
ExecStart=/usr/local/bin/tarak server --data-dir=/var/lib/tarak --bind-address=0.0.0.0:6443 --log-level=info
Restart=always
RestartSec=5s
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
KillMode=process

[Install]
WantedBy=multi-user.target
EOF

# Reload & start daemon:
sudo systemctl daemon-reload
sudo systemctl enable --now tarak
sudo systemctl status tarak`}
          </pre>
        </div>
      </div>

      {/* Command Groups */}
      <div className="space-y-8">
        {commandGroups.map((cg, idx) => (
          <div key={idx} className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-5">
            <div className="space-y-1">
              <h2 className="text-lg sm:text-xl font-bold text-white">{cg.group}</h2>
              <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">{cg.desc}</p>
            </div>

            {/* Example syntax box */}
            <div className="relative rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
              <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400 font-mono">
                <span>Syntax & Example</span>
                <button
                  onClick={() => copyCode(cg.example, idx)}
                  className="flex items-center gap-1 text-slate-300 hover:text-cyan-300"
                >
                  {copiedIndex === idx ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
                  <span>{copiedIndex === idx ? "Copied" : "Copy"}</span>
                </button>
              </div>
              <pre className="p-4 font-mono text-xs text-purple-300 overflow-x-auto leading-relaxed">
                {cg.example}
              </pre>
            </div>

            {/* Flags Table */}
            <div className="overflow-x-auto rounded-xl border border-white/10">
              <table className="w-full text-left text-xs border-collapse font-mono">
                <thead>
                  <tr className="bg-slate-950/90 text-slate-300 uppercase tracking-wider font-bold border-b border-white/10">
                    <th className="p-3">Flag Option</th>
                    <th className="p-3">Default</th>
                    <th className="p-3 font-sans">Description</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/5">
                  {cg.flags.map((fl, fIdx) => (
                    <tr key={fIdx} className="hover:bg-white/[0.02] transition-colors">
                      <td className="p-3 text-cyan-300 font-bold whitespace-nowrap">{fl.flag}</td>
                      <td className="p-3 text-slate-400 whitespace-nowrap">{fl.default}</td>
                      <td className="p-3 text-slate-300 font-sans text-xs">{fl.desc}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
