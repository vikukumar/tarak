import React from "react";
import { Terminal, Code, Layers } from "lucide-react";

export const CliReferencePage: React.FC = () => {
  const cliCommands = [
    { command: "tarak", aliases: "tarak server", desc: "Launch all-in-one local control plane and worker container runtime" },
    { command: "tarakd", aliases: "tarak-master", desc: "Dedicated production control-plane API server daemon" },
    { command: "taraks", aliases: "tarak-node", desc: "Dedicated high-performance worker agent daemon" },
    { command: "tarakctl get pods", aliases: "tarakctl get po", desc: "List all container pods with IP, status, and node location" },
    { command: "tarakctl get nodes", aliases: "tarakctl get no", desc: "List all control plane and worker cluster nodes" },
    { command: "tarakctl get services", aliases: "tarakctl get svc", desc: "List all active internal and external services" },
    { command: "tarakctl apply -f <file>", aliases: "tarakctl apply", desc: "Apply declarative multi-resource YAML or JSON manifests" },
    { command: "tarakctl delete <kind> <name>", aliases: "tarakctl del", desc: "Delete a resource from cluster state" },
    { command: "tarakctl logs <pod> [-f]", aliases: "tarakctl log", desc: "Stream stdout and stderr logs for active pod containers" },
    { command: "tarakctl exec -it <pod> -- <cmd>", aliases: "tarakctl exec", desc: "Spawn interactive terminal session inside a container sandbox" },
    { command: "tarakctl tunnel list", aliases: "tarakctl tun ls", desc: "List live Cloudflare and Tailscale tunnel endpoints" },
    { command: "tarakctl version", aliases: "tarakctl v", desc: "Display client and server version information" },
  ];

  return (
    <div className="space-y-10 animate-fade-in max-w-5xl mx-auto">
      {/* Title */}
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/30 text-purple-300 text-xs font-bold uppercase tracking-wider">
          Command-Line Interface Manual
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          CLI <span className="text-transparent bg-clip-text bg-gradient-to-r from-purple-400 to-cyan-400">Reference Manual</span>
        </h1>
        <p className="text-slate-400 max-w-xl mx-auto text-sm sm:text-base">
          Complete flag documentation and usage patterns for <code className="text-cyan-300 font-mono">tarak</code> and <code className="text-purple-300 font-mono">tarakctl</code>.
        </p>
      </div>

      {/* Commands Table */}
      <div className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
        <h2 className="text-lg font-bold text-white">CLI Command Directory</h2>

        <div className="overflow-x-auto rounded-xl border border-white/10">
          <table className="w-full text-left text-xs border-collapse">
            <thead>
              <tr className="bg-slate-950/80 text-slate-300 uppercase tracking-wider font-bold border-b border-white/10">
                <th className="p-3">Command</th>
                <th className="p-3">Aliases</th>
                <th className="p-3">Description</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 font-mono">
              {cliCommands.map((cmd, idx) => (
                <tr key={idx} className="hover:bg-white/[0.02] transition-colors">
                  <td className="p-3 text-cyan-300 font-bold">{cmd.command}</td>
                  <td className="p-3 text-slate-400">{cmd.aliases}</td>
                  <td className="p-3 text-slate-300 font-sans text-xs">{cmd.desc}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
