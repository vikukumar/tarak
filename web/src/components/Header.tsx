import React, { useState } from "react";
import { Star, Menu, X } from "lucide-react";

interface HeaderProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

export const Header: React.FC<HeaderProps> = ({ activeTab, setActiveTab }) => {
  const [mobileOpen, setMobileOpen] = useState(false);

  const navItems = [
    { id: "home", label: "Home" },
    { id: "getting-started", label: "Getting Started" },
    { id: "multi-node", label: "Multi-Node" },
    { id: "tunnels", label: "Tunnels & Ingress" },
    { id: "architecture", label: "Architecture" },
    { id: "api-reference", label: "API Reference" },
    { id: "cli-reference", label: "CLI Reference" },
    { id: "releases", label: "Releases" },
  ];

  return (
    <header className="sticky top-0 z-50 bg-[#050810]/85 backdrop-blur-xl border-b border-white/10 px-6 py-3.5 flex items-center justify-between transition-all">
      {/* Brand Horizontal Logo */}
      <a
        href="#home"
        onClick={() => setActiveTab("home")}
        className="flex items-center no-underline z-50 flex-shrink-0"
      >
        <img
          src="/assets/tarak_logo_horizontal.png"
          alt="TARAK"
          className="h-11 w-auto min-w-[150px] max-w-[220px] object-contain drop-shadow-[0_0_16px_rgba(0,240,255,0.4)] hover:scale-105 transition-transform"
        />
      </a>

      {/* Desktop Navigation */}
      <nav className="hidden xl:flex items-center gap-1.5">
        {navItems.map((item) => (
          <button
            key={item.id}
            onClick={() => setActiveTab(item.id)}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-all whitespace-nowrap ${
              activeTab === item.id
                ? "bg-cyan-500/10 border border-cyan-500/30 text-cyan-400 font-semibold"
                : "text-slate-400 hover:text-white hover:bg-white/5 border border-transparent"
            }`}
          >
            {item.label}
          </button>
        ))}

        <a
          href="https://github.com/vikukumar/tarak"
          target="_blank"
          rel="noreferrer"
          className="ml-2 px-4 py-1.5 rounded-xl bg-white/5 hover:bg-white/10 border border-white/15 text-white text-sm font-semibold flex items-center gap-2 hover:shadow-[0_0_20px_rgba(0,240,255,0.3)] hover:border-cyan-500/40 transition-all"
        >
          <Star size={14} className="text-amber-400 fill-amber-400" />
          <span>Star on GitHub</span>
        </a>
      </nav>

      {/* Mobile Toggle */}
      <button
        onClick={() => setMobileOpen(!mobileOpen)}
        className="xl:hidden p-2 rounded-lg bg-white/5 border border-white/10 text-white"
        aria-label="Toggle Menu"
      >
        {mobileOpen ? <X size={20} /> : <Menu size={20} />}
      </button>

      {/* Mobile Drawer */}
      {mobileOpen && (
        <div className="xl:hidden fixed inset-0 top-[60px] bg-[#050810]/95 backdrop-blur-2xl z-40 p-6 flex flex-col gap-3">
          {navItems.map((item) => (
            <button
              key={item.id}
              onClick={() => {
                setActiveTab(item.id);
                setMobileOpen(false);
              }}
              className={`px-4 py-3 rounded-xl text-left text-base font-semibold ${
                activeTab === item.id
                  ? "bg-cyan-500/15 border border-cyan-500/30 text-cyan-400"
                  : "text-slate-300 hover:bg-white/5"
              }`}
            >
              {item.label}
            </button>
          ))}
          <a
            href="https://github.com/vikukumar/tarak"
            target="_blank"
            rel="noreferrer"
            className="mt-4 px-4 py-3 rounded-xl bg-cyan-500/20 border border-cyan-500/40 text-cyan-300 font-bold flex items-center justify-center gap-2"
          >
            <Star size={16} className="text-amber-400 fill-amber-400" />
            <span>Star on GitHub</span>
          </a>
        </div>
      )}
    </header>
  );
};
