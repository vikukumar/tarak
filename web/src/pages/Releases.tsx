import React, { useEffect, useState } from "react";
import { Download, ExternalLink, Tag, Sparkles } from "lucide-react";
import { MarkdownViewer } from "../components/MarkdownViewer";

export const ReleasesPage: React.FC = () => {
  const [releases, setReleases] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadReleases() {
      try {
        const res = await fetch("/data/releases.json");
        if (res.ok) {
          const data = await res.json();
          setReleases(data);
        }
      } catch (err) {
        console.error("Failed to load local releases.json:", err);
      } finally {
        setLoading(false);
      }
    }
    loadReleases();
  }, []);

  return (
    <div className="space-y-8 animate-fade-in">
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs font-bold uppercase tracking-wider">
          Continuous Delivery
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          Releases & <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-purple-400">Changelog</span>
        </h1>
        <p className="text-slate-400 max-w-xl mx-auto text-sm">
          Track published versions, feature highlights, and multi-platform binary builds for Linux, Windows, and macOS.
        </p>
      </div>

      <div className="space-y-6 max-w-4xl mx-auto">
        {releases.map((rel) => (
          <div
            key={rel.version}
            className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4"
          >
            <div className="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-white/10">
              <div className="flex items-center gap-2.5">
                <span className="text-lg font-bold text-white">⚡ {rel.name}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className={`px-2.5 py-0.5 rounded-full text-xs font-bold ${
                  rel.isLatest ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/40" : "bg-purple-500/20 text-purple-300 border border-purple-500/30"
                }`}>
                  {rel.tag}
                </span>
                <span className="text-xs text-slate-400 font-mono">{rel.date}</span>
              </div>
            </div>

            <div className="space-y-2">
              <h4 className="text-xs font-bold uppercase tracking-wider text-cyan-300">
                🚀 Release Highlights & Features
              </h4>
              <ul className="list-disc ml-5 space-y-1 text-xs text-slate-300">
                {rel.highlights?.map((h: string, idx: number) => (
                  <li key={idx}>{h}</li>
                ))}
              </ul>
            </div>

            <div className="space-y-2 pt-2">
              <h4 className="text-xs font-bold uppercase tracking-wider text-purple-300">
                📦 Multi-Platform Binaries
              </h4>
              <div className="flex flex-wrap gap-2">
                {rel.binaries?.map((b: string) => (
                  <span key={b} className="px-2.5 py-1 rounded-lg bg-slate-950 border border-white/10 text-xs font-mono text-slate-200">
                    {b}
                  </span>
                ))}
              </div>
            </div>

            <div className="pt-2">
              <a
                href={rel.downloadUrl || `https://github.com/vikukumar/tarak/releases/tag/${rel.tag}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-cyan-500/15 hover:bg-cyan-500/25 border border-cyan-500/30 text-cyan-300 text-xs font-bold transition-all"
              >
                <Download size={14} />
                <span>Download {rel.tag} Assets on GitHub</span>
                <ExternalLink size={12} />
              </a>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
