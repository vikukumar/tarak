"use client";

import React, { useState } from "react";
import {
  Search,
  Bell,
  RefreshCw,
  Sun,
  Moon,
  Sparkles,
  Shield,
  LogOut,
  User,
  Key,
  Layers,
  CheckCircle2,
  AlertTriangle,
  Radio,
  ExternalLink,
} from "lucide-react";
import { useClusterState } from "@/hooks/useClusterState";
import { useAuth } from "@/hooks/useAuth";
import { cn } from "@/lib/utils";

interface TopbarProps {
  onOpenMobileMenu?: () => void;
}

export const Topbar: React.FC<TopbarProps> = () => {
  const { namespaces, selectedNamespace, setSelectedNamespace, refresh, isLoading } =
    useClusterState();
  const { user, logout } = useAuth();

  const [showNotifications, setShowNotifications] = useState(false);
  const [showProfileMenu, setShowProfileMenu] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const notifications = [
    {
      id: "1",
      title: "Zero-Trust mTLS Active",
      message: "Cluster mesh certificates auto-rotated successfully",
      type: "success",
      time: "2m ago",
    },
    {
      id: "2",
      title: "Cloudflare Tunnel Connected",
      message: "Public tunnel routed to app.vikshro.in",
      type: "info",
      time: "10m ago",
    },
    {
      id: "3",
      title: "MetalLB IP Pool",
      message: "192.168.1.240 assigned to web-app-lb",
      type: "info",
      time: "25m ago",
    },
  ];

  const handleManualRefresh = () => {
    setIsRefreshing(true);
    refresh();
    setTimeout(() => setIsRefreshing(false), 600);
  };

  return (
    <header className="h-16 border-b border-white/10 bg-[#070c18]/80 backdrop-blur-xl sticky top-0 z-20 px-4 md:px-6 flex items-center justify-between gap-4">
      {/* Left: Cluster Status & Dynamic Namespace Filter */}
      <div className="flex items-center gap-3 md:gap-5">
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-xs font-semibold text-emerald-400">
          <span className="w-2 h-2 rounded-full bg-emerald-400 shadow-[0_0_8px_#10b981] animate-pulse" />
          <span className="hidden sm:inline">tarak-cluster-prod</span>
        </div>

        {/* Dynamic Namespace Selector */}
        <div className="flex items-center gap-2 bg-slate-900/60 border border-white/10 px-3 py-1.5 rounded-lg text-xs">
          <Layers size={14} className="text-slate-400" />
          <select
            value={selectedNamespace}
            onChange={(e) => setSelectedNamespace(e.target.value)}
            className="bg-transparent text-white outline-none cursor-pointer font-medium"
          >
            {namespaces.map((ns) => (
              <option key={ns} value={ns} className="bg-slate-900 text-white">
                ns: {ns}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Right Controls */}
      <div className="flex items-center gap-2 sm:gap-3">
        {/* Sync Button */}
        <button
          onClick={handleManualRefresh}
          className="p-2 rounded-lg text-slate-300 hover:text-white bg-slate-900/60 hover:bg-white/10 border border-white/10 transition-colors"
          title="Refresh cluster state"
        >
          <RefreshCw size={16} className={cn(isRefreshing ? "animate-spin text-cyan-400" : "")} />
        </button>

        {/* Notifications */}
        <div className="relative">
          <button
            onClick={() => setShowNotifications(!showNotifications)}
            className="p-2 rounded-lg text-slate-300 hover:text-white bg-slate-900/60 hover:bg-white/10 border border-white/10 transition-colors relative"
            title="Cluster Notifications"
          >
            <Bell size={16} />
            <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-cyan-400 shadow-[0_0_6px_#00f0ff]" />
          </button>

          {showNotifications && (
            <div className="absolute right-0 mt-2 w-80 glass-panel rounded-xl p-4 shadow-2xl z-50 border border-white/10 space-y-3">
              <div className="flex items-center justify-between pb-2 border-b border-white/10">
                <span className="text-xs font-bold text-white uppercase tracking-wider">
                  Cluster Alerts
                </span>
                <span className="text-[10px] text-cyan-400 bg-cyan-500/10 px-1.5 py-0.5 rounded">
                  3 New
                </span>
              </div>
              <div className="space-y-2 max-h-64 overflow-y-auto pr-1">
                {notifications.map((n) => (
                  <div key={n.id} className="p-2 rounded-lg bg-white/5 border border-white/5 text-xs">
                    <div className="font-semibold text-slate-200">{n.title}</div>
                    <div className="text-[11px] text-slate-400 mt-0.5">{n.message}</div>
                    <div className="text-[9px] text-slate-500 mt-1">{n.time}</div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* User Profile & PAT Dropdown */}
        <div className="relative">
          <button
            onClick={() => setShowProfileMenu(!showProfileMenu)}
            className="flex items-center gap-2 p-1.5 pr-3 rounded-lg bg-slate-900/60 hover:bg-white/10 border border-white/10 transition-colors"
          >
            <div className="w-7 h-7 rounded-full bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center text-xs font-bold text-slate-950">
              {user?.username?.charAt(0).toUpperCase() || "A"}
            </div>
            <div className="hidden md:flex flex-col text-left">
              <span className="text-xs font-semibold text-white leading-none">
                {user?.username || "Super-Admin"}
              </span>
              <span className="text-[10px] text-emerald-400 font-mono mt-0.5">cluster-admin</span>
            </div>
          </button>

          {showProfileMenu && (
            <div className="absolute right-0 mt-2 w-56 glass-panel rounded-xl p-2 shadow-2xl z-50 border border-white/10 space-y-1">
              <div className="px-3 py-2 border-b border-white/10">
                <div className="text-xs font-bold text-white">{user?.username || "Super Admin"}</div>
                <div className="text-[11px] text-slate-400 font-mono">admin@tarak.io</div>
              </div>

              <a
                href="/dashboard/settings/pat"
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs text-slate-300 hover:text-white hover:bg-white/10 transition-colors"
              >
                <Key size={14} className="text-cyan-400" />
                <span>Personal Access Tokens</span>
              </a>

              <a
                href="/dashboard/settings/profile"
                className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs text-slate-300 hover:text-white hover:bg-white/10 transition-colors"
              >
                <User size={14} className="text-indigo-400" />
                <span>Account Profile</span>
              </a>

              <button
                onClick={logout}
                className="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-xs text-rose-400 hover:bg-rose-500/10 transition-colors text-left"
              >
                <LogOut size={14} />
                <span>Sign Out</span>
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};
