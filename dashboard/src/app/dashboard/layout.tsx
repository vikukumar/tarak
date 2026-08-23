"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { Sidebar } from "@/components/navigation/Sidebar";
import { Topbar } from "@/components/navigation/Topbar";
import { MobileBottomNav } from "@/components/navigation/MobileBottomNav";
import { useAuth } from "@/hooks/useAuth";
import { useTarakWebSocket } from "@/hooks/useTarakWebSocket";
import { Shield, ArrowRight } from "lucide-react";

import { ClusterProvider } from "@/context/ClusterContext";

export default function DashboardShellLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, loading, needsSetup } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [liveToast, setLiveToast] = useState<string | null>(null);

  // Connect to live WebSocket event stream
  useTarakWebSocket((evt) => {
    if (evt.type === "POD_UPDATED" || evt.type === "TELEMETRY_PULSE") {
      // live updates
    }
  });

  useEffect(() => {
    if (!loading) {
      if (needsSetup) {
        window.location.href = "/setup/";
      }
    }
  }, [loading, needsSetup]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#070c18]">
        <div className="flex flex-col items-center gap-4">
          <img
            src="/assets/icon.png"
            alt="TARAK"
            className="w-12 h-12 object-contain animate-pulse drop-shadow-[0_0_20px_rgba(0,240,255,0.5)]"
          />
          <span className="text-xs text-slate-400 font-medium">Loading Tarak Control Plane...</span>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#070c18] p-4">
        <div className="max-w-md w-full glass-panel rounded-2xl p-8 border border-white/10 text-center space-y-6 shadow-2xl">
          <img
            src="/assets/tarak_logo_vertical.png"
            alt="TARAK Platform"
            className="w-44 object-contain mx-auto drop-shadow-[0_0_25px_rgba(0,240,255,0.35)]"
          />
          <div className="space-y-1.5">
            <h2 className="text-xl font-bold text-white tracking-tight">Authentication Required</h2>
            <p className="text-xs text-slate-400">
              Your session has expired or master credentials are required to manage cluster workloads.
            </p>
          </div>
          <div className="flex flex-col gap-2.5">
            <a
              href="/login/"
              className="w-full py-2.5 rounded-xl bg-gradient-to-r from-cyan-500 to-indigo-600 text-slate-950 font-bold text-xs hover:opacity-90 transition-opacity flex items-center justify-center gap-2 shadow-lg shadow-cyan-500/20"
            >
              <span>Sign In to Cluster</span>
              <ArrowRight size={14} />
            </a>
            <a
              href="/setup/"
              className="w-full py-2.5 rounded-xl bg-slate-900 hover:bg-white/10 text-slate-300 font-semibold text-xs border border-white/10 transition-colors"
            >
              1st Time Super-Admin Setup
            </a>
          </div>
        </div>
      </div>
    );
  }

  return (
    <ClusterProvider>
      <div className="flex min-h-screen bg-[#070c18] text-slate-100">
        {/* Collapsible Sidebar */}
        <Sidebar
          collapsed={collapsed}
          onToggleCollapse={() => setCollapsed(!collapsed)}
        />

        {/* Main Content Area */}
        <div className="flex-1 flex flex-col min-w-0 pb-20 md:pb-6">
          <Topbar />
          <main className="flex-1 p-4 md:p-6 max-w-7xl w-full mx-auto space-y-6">
            {children}
          </main>
        </div>

        {/* Mobile App Dock */}
        <MobileBottomNav />

        {/* Live Toast */}
        {liveToast && (
          <div className="fixed bottom-20 right-6 z-50 p-3 rounded-xl bg-slate-900/95 border border-cyan-500/40 text-xs text-white shadow-2xl animate-fade-in">
            {liveToast}
          </div>
        )}
      </div>
    </ClusterProvider>
  );
}
