import React, { useState, useEffect } from "react";
import { Header } from "./components/Header";
import { ClusterCanvas } from "./components/ClusterCanvas";
import { HomePage } from "./pages/Home";
import { ArchitecturePage } from "./pages/Architecture";
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

      {/* Header */}
      <Header activeTab={activeTab} setActiveTab={handleNavigate} />

      {/* Main Content Area */}
      <main className="max-w-6xl mx-auto px-4 sm:px-6 py-10 relative z-10">
        {activeTab === "home" && <HomePage onNavigate={handleNavigate} />}
        {activeTab === "architecture" && <ArchitecturePage />}
        {activeTab === "releases" && <ReleasesPage />}
        {activeTab !== "home" && activeTab !== "architecture" && activeTab !== "releases" && (
          <div className="p-8 rounded-2xl bg-slate-900/60 border border-white/10 text-center space-y-3">
            <h2 className="text-2xl font-bold text-white capitalize">{activeTab.replace("-", " ")}</h2>
            <p className="text-slate-400 text-sm">
              Explore interactive modules, documentation, and live tutorials for {activeTab}.
            </p>
            <button
              onClick={() => handleNavigate("home")}
              className="mt-4 px-4 py-2 rounded-xl bg-cyan-500/20 text-cyan-300 border border-cyan-500/30 text-xs font-bold"
            >
              ← Back to Home
            </button>
          </div>
        )}
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
