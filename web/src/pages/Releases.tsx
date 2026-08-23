import React, { useEffect, useState } from "react";
import { Download, ExternalLink, Tag, Sparkles, CheckCircle2, Package, Calendar, User } from "lucide-react";
import defaultReleases from "../../public/data/releases.json";

export const ReleasesPage: React.FC = () => {
  const [releases, setReleases] = useState<any[]>(defaultReleases);
  const [loading, setLoading] = useState(false);
  const [activeTag, setActiveTag] = useState<string>("all");

  useEffect(() => {
    async function fetchGitHubReleases() {
      try {
        setLoading(true);
        const res = await fetch("https://api.github.com/repos/vikukumar/tarak/releases");
        if (res.ok) {
          const ghList: any[] = await res.json();
          if (Array.isArray(ghList) && ghList.length > 0) {
            const mapped = ghList.map((gh, idx) => {
              // Find matching local release for enriched highlights if available
              const localMatch = defaultReleases.find((r) => r.tag === gh.tag_name || r.version === gh.tag_name?.replace(/^v/, ""));
              
              const binaries = gh.assets?.map((a: any) => a.name) || localMatch?.binaries || [
                "tarak (Unified All-in-One Engine)",
                "tarakctl / taraktl (CLI)",
                "tarakd (Control Plane API Server)",
                "taraks (Worker Node Agent)"
              ];

              return {
                version: (gh.tag_name || "").replace(/^v/, ""),
                tag: gh.tag_name,
                name: gh.name || `Tarak ${gh.tag_name}`,
                date: gh.published_at ? gh.published_at.substring(0, 10) : localMatch?.date || "2026-08-23",
                isLatest: idx === 0,
                status: idx === 0 ? "Production Ready" : "Stable",
                body: gh.body || "",
                highlights: localMatch?.highlights || [
                  "Multi-platform production binary build for Linux, Windows, and macOS (AMD64 & ARM64)",
                  "Built-in TCR OCI Sandbox with process isolation",
                  "Zero-Trust mTLS Security & Cloudflare/Tailscale Tunnel Integration",
                  "Full Kubernetes API & Helm chart compatibility",
                ],
                binaries: binaries,
                platforms: localMatch?.platforms || [
                  "linux/amd64",
                  "linux/arm64",
                  "windows/amd64",
                  "windows/arm64",
                  "darwin/amd64",
                  "darwin/arm64"
                ],
                downloadUrl: gh.html_url || `https://github.com/vikukumar/tarak/releases/tag/${gh.tag_name}`,
                assets: gh.assets || [],
              };
            });
            setReleases(mapped);
          }
        }
      } catch (err) {
        console.warn("Using bundled releases data:", err);
      } finally {
        setLoading(false);
      }
    }
    fetchGitHubReleases();
  }, []);

  return (
    <div className="space-y-10 animate-fade-in">
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs font-bold uppercase tracking-wider">
          Continuous Delivery & Releases
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          Releases & <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-purple-400">Changelog</span>
        </h1>
        <p className="text-slate-400 max-w-xl mx-auto text-xs sm:text-sm">
          Track official releases, feature highlights, and multi-platform binary distributions for Linux, Windows, and macOS.
        </p>
      </div>

      <div className="space-y-6 max-w-4xl mx-auto">
        {releases.map((rel) => (
          <div
            key={rel.version || rel.tag}
            className={`p-6 sm:p-7 rounded-2xl border transition-all ${
              rel.isLatest
                ? "bg-slate-900/80 border-cyan-500/40 shadow-[0_0_30px_rgba(0,240,255,0.1)]"
                : "bg-slate-900/50 border-white/10"
            } space-y-5`}
          >
            <div className="flex flex-wrap items-center justify-between gap-3 pb-4 border-b border-white/10">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-xl bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400 font-bold">
                  ⚡
                </div>
                <div>
                  <h3 className="text-lg font-bold text-white">{rel.name}</h3>
                  <div className="flex items-center gap-2 text-xs text-slate-400">
                    <span className="flex items-center gap-1 font-mono">
                      <Calendar size={12} /> {rel.date}
                    </span>
                    <span>•</span>
                    <span className="text-emerald-400 font-semibold">{rel.status || "Stable"}</span>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <span className={`px-3 py-1 rounded-full text-xs font-bold font-mono ${
                  rel.isLatest
                    ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-[0_0_12px_rgba(0,240,255,0.3)]"
                    : "bg-purple-500/20 text-purple-300 border border-purple-500/30"
                }`}>
                  {rel.tag}
                </span>
              </div>
            </div>

            {/* Highlights */}
            <div className="space-y-2">
              <h4 className="text-xs font-bold uppercase tracking-wider text-cyan-300 flex items-center gap-1.5">
                <Sparkles size={13} />
                <span>Release Highlights</span>
              </h4>
              <ul className="space-y-1.5 text-xs text-slate-300 pl-1">
                {rel.highlights?.map((h: string, idx: number) => (
                  <li key={idx} className="flex items-start gap-2">
                    <CheckCircle2 size={14} className="text-cyan-400 flex-shrink-0 mt-0.5" />
                    <span>{h}</span>
                  </li>
                ))}
              </ul>
            </div>

            {/* Multi-Platform Binary Matrix */}
            <div className="space-y-2 pt-1">
              <h4 className="text-xs font-bold uppercase tracking-wider text-purple-300 flex items-center gap-1.5">
                <Package size={13} />
                <span>Included Binaries & Architectures</span>
              </h4>
              <div className="flex flex-wrap gap-2">
                {rel.binaries?.map((b: string, i: number) => (
                  <span key={i} className="px-2.5 py-1 rounded-lg bg-slate-950/80 border border-white/10 text-xs font-mono text-slate-200">
                    {b}
                  </span>
                ))}
              </div>
            </div>

            {/* Direct GitHub Download Links */}
            <div className="pt-2 flex flex-wrap items-center gap-3">
              <a
                href={rel.downloadUrl || `https://github.com/vikukumar/tarak/releases/tag/${rel.tag}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-cyan-500/20 hover:bg-cyan-500/30 border border-cyan-500/40 text-cyan-300 text-xs font-bold transition-all shadow-[0_0_15px_rgba(0,240,255,0.2)]"
              >
                <Download size={14} />
                <span>Download {rel.tag} Release Package</span>
                <ExternalLink size={12} />
              </a>
              <a
                href={`https://github.com/vikukumar/tarak/tree/${rel.tag}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1.5 px-3.5 py-2.5 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-300 text-xs font-medium transition-all"
              >
                <span>View Git Source</span>
                <ExternalLink size={11} />
              </a>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
