import React, { useState, useEffect } from "react";
import { Header } from "./components/Header";
import { ClusterCanvas } from "./components/ClusterCanvas";
import { HomePage } from "./pages/Home";
import { GettingStartedPage } from "./pages/GettingStarted";
import { MultiNodePage } from "./pages/MultiNode";
import { TunnelsPage } from "./pages/Tunnels";
import { ArchitecturePage } from "./pages/Architecture";
import { ApiReferencePage } from "./pages/ApiReference";
import { CliReferencePage } from "./pages/CliReference";
import { ReleasesPage } from "./pages/Releases";

export default function App() {
  const [activeTab, setActiveTab] = useState<string>("home");

  // Sync hash routing
  useEffect(() => {
    const handleHashChange = () => {
      const hash = window.location.hash.replace("#", "") || "home";
      setActiveTab(hash);
    };

    handleHashChange();
    window.addEventListener("hashchange", handleHashChange);
    return () => window.removeEventListener("hashchange", handleHashChange);
  }, []);

  const handleNavigate = (tab: string) => {
    setActiveTab(tab);
    window.location.hash = tab;
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  return (
    <div className="min-h-screen bg-[#030712] text-slate-100 font-sans relative selection:bg-cyan-500/30 selection:text-cyan-200">
      {/* Background Simulation */}
      <ClusterCanvas />

      {/* Header with Horizontal Logo */}
      <Header activeTab={activeTab} setActiveTab={handleNavigate} />

      {/* Main Content Area */}
      <main className="max-w-6xl mx-auto px-4 sm:px-6 py-10 relative z-10">
        {activeTab === "home" && <HomePage onNavigate={handleNavigate} />}
        {activeTab === "getting-started" && <GettingStartedPage onNavigate={handleNavigate} />}
        {activeTab === "multi-node" && <MultiNodePage onNavigate={handleNavigate} />}
        {activeTab === "tunnels" && <TunnelsPage />}
        {activeTab === "architecture" && <ArchitecturePage />}
        {activeTab === "api-reference" && <ApiReferencePage />}
        {activeTab === "cli-reference" && <CliReferencePage />}
        {activeTab === "releases" && <ReleasesPage />}
      </main>

      {/* Footer */}
      <footer className="border-t border-white/10 bg-[#050810]/85 backdrop-blur-xl py-8 text-center text-xs text-slate-400 relative z-10">
        <p>© 2026 TARAK Container Orchestration Platform. Open source under MIT License.</p>
        <p className="mt-2 text-[11px] text-slate-400">
          Created with ❤️ by{" "}
          <a
            href="https://github.com/vikukumar"
            target="_blank"
            rel="noreferrer"
            className="text-cyan-400 hover:underline font-semibold"
          >
            Vikash Kumar (@vikukumar)
          </a>{" "}
          • Made in India 🇮🇳
        </p>
      </footer>
    </div>
  );
}
