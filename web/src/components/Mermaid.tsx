import React, { useEffect, useRef, useState } from "react";
import mermaid from "mermaid";
import { Check, Copy } from "lucide-react";

mermaid.initialize({
  startOnLoad: false,
  theme: "dark",
  themeVariables: {
    darkMode: true,
    background: "#04060c",
    primaryColor: "#00f0ff",
    primaryTextColor: "#f8fafc",
    primaryBorderColor: "#00f0ff",
    lineColor: "#38bdf8",
    secondaryColor: "#a855f7",
    tertiaryColor: "#10b981",
    fontFamily: "JetBrains Mono, monospace",
  },
  flowchart: {
    curve: "basis",
    htmlLabels: true,
  },
});

interface MermaidProps {
  chart: string;
  id?: string;
}

export const Mermaid: React.FC<MermaidProps> = ({ chart, id = "mermaid-chart" }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [svgContent, setSvgContent] = useState<string>("");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const renderChart = async () => {
      try {
        const uniqueId = `${id}-${Math.random().toString(36).substring(2, 9)}`;
        const { svg } = await mermaid.render(uniqueId, chart.trim());
        if (isMounted) {
          setSvgContent(svg);
        }
      } catch (err) {
        console.error("Mermaid render error:", err);
      }
    };
    renderChart();
    return () => {
      isMounted = false;
    };
  }, [chart, id]);

  const handleCopy = () => {
    navigator.clipboard.writeText(chart);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative group my-4 rounded-xl border border-white/10 bg-[#04060c]/90 p-4 shadow-2xl overflow-x-auto">
      <button
        onClick={handleCopy}
        className="absolute top-3 right-3 p-1.5 rounded-lg bg-slate-900/80 border border-white/10 text-slate-400 hover:text-cyan-300 hover:border-cyan-500/30 transition-all opacity-0 group-hover:opacity-100 text-xs flex items-center gap-1"
        title="Copy Mermaid Code"
      >
        {copied ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
        <span>{copied ? "Copied" : "Source"}</span>
      </button>

      <div
        ref={containerRef}
        dangerouslySetInnerHTML={{ __html: svgContent }}
        className="flex justify-center min-w-max"
      />
    </div>
  );
};
