"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Sidebar } from "@/components/navigation/Sidebar";
import { Topbar } from "@/components/navigation/Topbar";
import { MobileBottomNav } from "@/components/navigation/MobileBottomNav";
import { useAuth } from "@/hooks/useAuth";
import { useTarakWebSocket } from "@/hooks/useTarakWebSocket";

export default function DashboardShellLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { isAuthenticated, loading } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [liveToast, setLiveToast] = useState<string | null>(null);

  // Connect to live WebSocket event stream
  useTarakWebSocket((evt) => {
    if (evt.type === "POD_UPDATED" || evt.type === "TELEMETRY_PULSE") {
      // live updates
    }
  });

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      router.replace("/login");
    }
  }, [isAuthenticated, loading, router]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#070c18]">
        <div className="flex flex-col items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-cyan-500 to-indigo-600 animate-pulse" />
          <span className="text-xs text-slate-400">Authenticating with Tarak Control Plane...</span>
        </div>
      </div>
    );
  }

  return (
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
  );
}
