"use client";

import React, { useState, useRef, useEffect } from "react";
import {
  Bell,
  RefreshCw,
  User,
  Shield,
  Layers,
  ChevronDown,
  Check,
  Plus,
  Globe,
  Search,
  Key,
  Settings,
  LogOut,
  Sparkles,
} from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { useCluster } from "@/context/ClusterContext";
import { cn } from "@/lib/utils";

export const Topbar: React.FC = () => {
  const { user, logout } = useAuth();
  const {
    namespaces,
    selectedNamespace,
    setSelectedNamespace,
    refresh,
    createNamespace,
  } = useCluster();

  const [showNotifications, setShowNotifications] = useState(false);
  const [showProfile, setShowProfile] = useState(false);
  const [showNamespaceDropdown, setShowNamespaceDropdown] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [newNamespaceName, setNewNamespaceName] = useState("");
  const [isCreatingNs, setIsCreatingNs] = useState(false);
  const [nsSearch, setNsSearch] = useState("");

  const dropdownRef = useRef<HTMLDivElement>(null);
  const profileRef = useRef<HTMLDivElement>(null);
  const notifRef = useRef<HTMLDivElement>(null);

  // Close dropdowns when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setShowNamespaceDropdown(false);
      }
      if (
        profileRef.current &&
        !profileRef.current.contains(event.target as Node)
      ) {
        setShowProfile(false);
      }
      if (
        notifRef.current &&
        !notifRef.current.contains(event.target as Node)
      ) {
        setShowNotifications(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleManualRefresh = async () => {
    setIsRefreshing(true);
    await refresh();
    setTimeout(() => setIsRefreshing(false), 500);
  };

  const handleCreateNamespace = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newNamespaceName.trim()) return;
    setIsCreatingNs(true);
    try {
      const ok = await createNamespace(newNamespaceName.trim());
      if (ok) {
        setNewNamespaceName("");
        setShowNamespaceDropdown(false);
      }
    } finally {
      setIsCreatingNs(false);
    }
  };

  const filteredNamespaces = namespaces.filter((ns) =>
    ns.toLowerCase().includes(nsSearch.toLowerCase())
  );

  return (
    <header className="h-16 border-b border-white/10 bg-[#070c18]/90 backdrop-blur-xl sticky top-0 z-30 px-4 md:px-6 flex items-center justify-between gap-4">
      {/* Left: Branding & Modern Namespace Selector */}
      <div className="flex items-center gap-3 md:gap-4">
        {/* Mobile Header Logo */}
        <a href="/dashboard/" className="md:hidden flex items-center">
          <img
            src="/assets/tarak_logo_horizontal.png"
            alt="TARAK Platform"
            className="h-7 object-contain drop-shadow-[0_0_12px_rgba(0,240,255,0.4)]"
          />
        </a>

        {/* Cluster Status Pill */}
        <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-xs font-semibold text-emerald-400">
          <span className="w-2 h-2 rounded-full bg-emerald-400 shadow-[0_0_8px_#10b981] animate-pulse" />
          <span>tarak-cluster-prod</span>
        </div>

        {/* Modern Glassmorphism Namespace Dropdown */}
        <div className="relative" ref={dropdownRef}>
          <button
            onClick={() => setShowNamespaceDropdown(!showNamespaceDropdown)}
            className="flex items-center gap-2.5 px-3 py-1.5 rounded-xl bg-slate-900/80 hover:bg-slate-800/90 border border-white/15 text-xs text-white transition-all shadow-sm group hover:border-cyan-500/40"
          >
            <div className="w-5 h-5 rounded-md bg-cyan-500/20 text-cyan-400 flex items-center justify-center">
              <Globe size={13} />
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-[11px] text-slate-400 font-medium hidden xs:inline">
                Namespace:
              </span>
              <span className="font-bold text-cyan-300 font-mono tracking-tight">
                {selectedNamespace === "_all" ? "All Namespaces" : selectedNamespace}
              </span>
            </div>
            <ChevronDown
              size={14}
              className={cn(
                "text-slate-400 transition-transform duration-200",
                showNamespaceDropdown ? "rotate-180 text-cyan-400" : ""
              )}
            />
          </button>

          {/* Floating Dropdown Popover */}
          {showNamespaceDropdown && (
            <div className="absolute left-0 mt-2 w-72 glass-panel rounded-xl p-3 shadow-2xl z-50 border border-white/15 space-y-3 animate-fade-in">
              {/* Search Filter */}
              <div className="relative">
                <Search
                  size={14}
                  className="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400"
                />
                <input
                  type="text"
                  value={nsSearch}
                  onChange={(e) => setNsSearch(e.target.value)}
                  placeholder="Filter namespaces..."
                  className="w-full bg-slate-950/80 border border-white/10 rounded-lg pl-8 pr-3 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-400"
                />
              </div>

              {/* Namespaces List */}
              <div className="max-h-48 overflow-y-auto space-y-1 pr-1 font-mono text-xs">
                {/* All Namespaces option */}
                <button
                  onClick={() => {
                    setSelectedNamespace("_all");
                    setShowNamespaceDropdown(false);
                  }}
                  className={cn(
                    "w-full flex items-center justify-between px-3 py-2 rounded-lg transition-colors text-left",
                    selectedNamespace === "_all"
                      ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/30"
                      : "hover:bg-white/5 text-slate-300"
                  )}
                >
                  <div className="flex items-center gap-2">
                    <Sparkles size={14} className="text-cyan-400" />
                    <span className="font-sans font-semibold">All Namespaces</span>
                  </div>
                  {selectedNamespace === "_all" && <Check size={14} className="text-cyan-400" />}
                </button>

                {filteredNamespaces.map((ns) => (
                  <button
                    key={ns}
                    onClick={() => {
                      setSelectedNamespace(ns);
                      setShowNamespaceDropdown(false);
                    }}
                    className={cn(
                      "w-full flex items-center justify-between px-3 py-2 rounded-lg transition-colors text-left",
                      selectedNamespace === ns
                        ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/30"
                        : "hover:bg-white/5 text-slate-300"
                    )}
                  >
                    <div className="flex items-center gap-2">
                      <div className="w-2 h-2 rounded-full bg-cyan-400 shadow-[0_0_6px_rgba(0,240,255,0.6)]" />
                      <span className="font-bold">{ns}</span>
                    </div>
                    {selectedNamespace === ns && <Check size={14} className="text-cyan-400" />}
                  </button>
                ))}
              </div>

              {/* Quick Create Namespace */}
              <form
                onSubmit={handleCreateNamespace}
                className="pt-2 border-t border-white/10 flex items-center gap-2"
              >
                <input
                  type="text"
                  value={newNamespaceName}
                  onChange={(e) => setNewNamespaceName(e.target.value)}
                  placeholder="New namespace name..."
                  className="flex-1 bg-slate-950/80 border border-white/10 rounded-lg px-2.5 py-1 text-[11px] text-white placeholder-slate-500 focus:outline-none focus:border-cyan-400"
                />
                <button
                  type="submit"
                  disabled={isCreatingNs || !newNamespaceName.trim()}
                  className="px-2.5 py-1 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-bold text-xs disabled:opacity-30 transition-colors flex items-center gap-1"
                >
                  <Plus size={13} />
                  <span>Add</span>
                </button>
              </form>
            </div>
          )}
        </div>
      </div>

      {/* Right Controls */}
      <div className="flex items-center gap-2 sm:gap-3">
        {/* Sync Button */}
        <button
          onClick={handleManualRefresh}
          className="p-2 rounded-xl text-slate-300 hover:text-white bg-slate-900/80 hover:bg-white/10 border border-white/10 transition-colors"
          title="Refresh cluster state"
        >
          <RefreshCw
            size={16}
            className={cn(isRefreshing ? "animate-spin text-cyan-400" : "")}
          />
        </button>

        {/* Notifications */}
        <div className="relative" ref={notifRef}>
          <button
            onClick={() => setShowNotifications(!showNotifications)}
            className="p-2 rounded-xl text-slate-300 hover:text-white bg-slate-900/80 hover:bg-white/10 border border-white/10 transition-colors relative"
            title="Cluster Notifications"
          >
            <Bell size={16} />
            <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-cyan-400 shadow-[0_0_6px_#00f0ff]" />
          </button>

          {showNotifications && (
            <div className="absolute right-0 mt-2 w-80 glass-panel rounded-xl p-4 shadow-2xl z-50 border border-white/15 space-y-3 animate-fade-in">
              <div className="flex items-center justify-between pb-2 border-b border-white/10">
                <span className="text-xs font-bold text-white uppercase tracking-wider">
                  Cluster Alerts
                </span>
                <span className="text-[10px] text-cyan-400 bg-cyan-500/10 px-1.5 py-0.5 rounded font-mono">
                  Live
                </span>
              </div>
              <div className="space-y-2 text-xs">
                <div className="p-2.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 space-y-0.5">
                  <div className="font-semibold flex items-center gap-1.5">
                    <Shield size={13} />
                    <span>mTLS Mesh Active</span>
                  </div>
                  <p className="text-[11px] text-slate-400">
                    Zero-Trust sidecar proxies configured for namespace {selectedNamespace}.
                  </p>
                </div>
                <div className="p-2.5 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-300 space-y-0.5">
                  <div className="font-semibold">Control Plane Ready</div>
                  <p className="text-[11px] text-slate-400">
                    All API controllers healthy on node vikshro_msm.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Profile / Account Dropdown */}
        <div className="relative" ref={profileRef}>
          <button
            onClick={() => setShowProfile(!showProfile)}
            className="flex items-center gap-2 pl-2 pr-3 py-1.5 rounded-xl bg-slate-900/80 hover:bg-white/10 border border-white/10 transition-colors"
          >
            <div className="w-7 h-7 rounded-lg bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center text-slate-950 font-bold text-xs shadow-md">
              {user?.username ? user.username[0].toUpperCase() : "A"}
            </div>
            <span className="text-xs font-medium text-slate-200 hidden md:inline truncate max-w-[100px]">
              {user?.username || "Admin"}
            </span>
          </button>

          {showProfile && (
            <div className="absolute right-0 mt-2 w-56 glass-panel rounded-xl p-3 shadow-2xl z-50 border border-white/15 space-y-2 animate-fade-in text-xs">
              <div className="px-2 py-1.5 border-b border-white/10">
                <span className="font-bold text-white block">
                  {user?.username || "Administrator"}
                </span>
                <span className="text-[10px] text-slate-400 font-mono">
                  Role: Super-Admin (Master)
                </span>
              </div>
              <a
                href="/dashboard/settings/profile/"
                className="flex items-center gap-2 px-2 py-1.5 rounded-lg text-slate-300 hover:text-white hover:bg-white/5 transition-colors"
              >
                <Settings size={14} />
                <span>Profile Settings</span>
              </a>
              <a
                href="/dashboard/settings/pat/"
                className="flex items-center gap-2 px-2 py-1.5 rounded-lg text-slate-300 hover:text-white hover:bg-white/5 transition-colors"
              >
                <Key size={14} />
                <span>Access Tokens</span>
              </a>
              <button
                onClick={logout}
                className="w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-rose-400 hover:bg-rose-500/10 transition-colors text-left"
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
