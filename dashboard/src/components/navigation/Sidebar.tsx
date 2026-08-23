"use client";

import React, { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Layers,
  Box,
  Server,
  Network,
  Globe,
  Radio,
  Activity,
  Shield,
  Terminal,
  FileCode,
  Settings,
  Key,
  ChevronDown,
  ChevronRight,
  Sparkles,
  RefreshCw,
  Cpu,
  Database,
  Lock,
  Workflow,
  PanelLeftClose,
  PanelLeftOpen,
  Cloud,
  Zap,
} from "lucide-react";
import { cn } from "@/lib/utils";

export interface NavItem {
  title: string;
  href?: string;
  icon: React.ElementType;
  badge?: string;
  children?: NavItem[];
}

export const navigationConfig: NavItem[] = [
  {
    title: "Cluster Overview",
    href: "/dashboard",
    icon: Sparkles,
  },
  {
    title: "Compute & Workloads",
    icon: Layers,
    children: [
      { title: "Pods", href: "/dashboard/workloads/pods", icon: Box },
      { title: "Deployments", href: "/dashboard/workloads/deployments", icon: Workflow },
      { title: "StatefulSets", href: "/dashboard/workloads/statefulsets", icon: Database },
      { title: "Jobs & CronJobs", href: "/dashboard/workloads/jobs", icon: Zap },
    ],
  },
  {
    title: "Cluster & Nodes",
    icon: Server,
    children: [
      { title: "Nodes", href: "/dashboard/cluster/nodes", icon: Cpu },
      { title: "Namespaces", href: "/dashboard/cluster/namespaces", icon: Globe },
      { title: "Storage (PVC / SC)", href: "/dashboard/cluster/storage", icon: Database },
    ],
  },
  {
    title: "Networking & Ingress",
    icon: Network,
    children: [
      { title: "Ingress & Domains", href: "/dashboard/networking/ingress", icon: Globe },
      { title: "Services & MetalLB", href: "/dashboard/networking/services", icon: Server },
      { title: "Cloudflare & Tailscale", href: "/dashboard/networking/tunnels", icon: Cloud, badge: "Tunnel" },
    ],
  },
  {
    title: "Service Mesh (Kuma)",
    icon: Radio,
    badge: "mTLS",
    children: [
      { title: "Mesh Overview", href: "/dashboard/mesh/overview", icon: Radio },
      { title: "Traffic Permissions", href: "/dashboard/mesh/permissions", icon: Lock },
      { title: "Passthrough Policies", href: "/dashboard/mesh/passthrough", icon: Network },
      { title: "Canary & Routes", href: "/dashboard/mesh/routes", icon: Workflow },
    ],
  },
  {
    title: "Observability",
    icon: Activity,
    children: [
      { title: "Hubble Network Flows", href: "/dashboard/observability/hubble", icon: Activity, badge: "Live" },
      { title: "Cluster Metrics", href: "/dashboard/observability/metrics", icon: Cpu },
      { title: "Container Logs", href: "/dashboard/observability/logs", icon: Terminal },
    ],
  },
  {
    title: "Security & Zero-Trust",
    icon: Shield,
    children: [
      { title: "RBAC Matrix", href: "/dashboard/security/rbac", icon: Lock },
      { title: "Tarak Security Policies", href: "/dashboard/security/zerotrust", icon: Shield },
      { title: "SSO & Identity", href: "/dashboard/security/sso", icon: Key },
    ],
  },
  {
    title: "Developer Tools",
    icon: Terminal,
    children: [
      { title: "Web Terminal", href: "/dashboard/devtools/terminal", icon: Terminal },
      { title: "YAML Manifest Apply", href: "/dashboard/devtools/manifests", icon: FileCode },
    ],
  },
  {
    title: "Administration",
    icon: Settings,
    children: [
      { title: "Personal Access Tokens", href: "/dashboard/settings/pat", icon: Key },
      { title: "Profile & Avatar", href: "/dashboard/settings/profile", icon: Settings },
      { title: "Auto-Updater", href: "/dashboard/settings/updates", icon: RefreshCw },
    ],
  },
];

interface SidebarProps {
  collapsed: boolean;
  onToggleCollapse: () => void;
  onItemClick?: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  collapsed,
  onToggleCollapse,
  onItemClick,
}) => {
  const pathname = usePathname();
  const [openMenus, setOpenMenus] = useState<Record<string, boolean>>({
    "Compute & Workloads": true,
    "Networking & Ingress": true,
    "Service Mesh (Kuma)": true,
  });

  const toggleMenu = (title: string) => {
    setOpenMenus((prev) => ({ ...prev, [title]: !prev[title] }));
  };

  return (
    <aside
      className={cn(
        "hidden md:flex flex-col h-screen sticky top-0 bg-[#070c18]/90 backdrop-blur-2xl border-r border-white/10 transition-all duration-300 z-30",
        collapsed ? "w-20" : "w-72"
      )}
    >
      {/* Brand Header */}
      <div className="h-16 flex items-center justify-between px-5 border-b border-white/10">
        <div className="flex items-center gap-3 overflow-hidden">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center shadow-[0_0_15px_rgba(0,240,255,0.4)] flex-shrink-0">
            <span className="font-bold text-slate-950 text-lg">T</span>
          </div>
          {!collapsed && (
            <div className="flex flex-col">
              <span className="font-bold text-white text-base tracking-wider leading-none">
                TARAK
              </span>
              <span className="text-[10px] text-cyan-400 font-mono tracking-widest mt-1">
                CONTROL PLANE
              </span>
            </div>
          )}
        </div>
        <button
          onClick={onToggleCollapse}
          className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
          title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
          {collapsed ? <PanelLeftOpen size={18} /> : <PanelLeftClose size={18} />}
        </button>
      </div>

      {/* Navigation Items */}
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5 custom-scrollbar">
        {navigationConfig.map((item) => {
          const Icon = item.icon;
          const hasChildren = item.children && item.children.length > 0;
          const isOpen = openMenus[item.title];
          const isDirectActive = item.href === pathname;
          const isChildActive = item.children?.some((c) => c.href === pathname);

          if (!hasChildren && item.href) {
            return (
              <Link
                key={item.title}
                href={item.href}
                onClick={onItemClick}
                className={cn(
                  "flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-sm font-medium transition-all group",
                  isDirectActive
                    ? "bg-cyan-500/15 text-cyan-400 border border-cyan-500/30 shadow-[0_0_15px_rgba(0,240,255,0.15)]"
                    : "text-slate-300 hover:text-white hover:bg-white/5 border border-transparent"
                )}
                title={collapsed ? item.title : undefined}
              >
                <Icon size={18} className={cn("flex-shrink-0", isDirectActive ? "text-cyan-400" : "text-slate-400 group-hover:text-white")} />
                {!collapsed && (
                  <div className="flex items-center justify-between flex-1 overflow-hidden">
                    <span className="truncate">{item.title}</span>
                    {item.badge && (
                      <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-cyan-500/20 text-cyan-400 border border-cyan-500/30 font-semibold">
                        {item.badge}
                      </span>
                    )}
                  </div>
                )}
              </Link>
            );
          }

          return (
            <div key={item.title} className="space-y-1">
              <button
                onClick={() => toggleMenu(item.title)}
                className={cn(
                  "w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-sm font-medium transition-all group",
                  isChildActive
                    ? "text-white bg-white/5"
                    : "text-slate-400 hover:text-white hover:bg-white/5"
                )}
                title={collapsed ? item.title : undefined}
              >
                <Icon size={18} className={cn("flex-shrink-0", isChildActive ? "text-cyan-400" : "text-slate-400 group-hover:text-white")} />
                {!collapsed && (
                  <>
                    <span className="flex-1 text-left truncate">{item.title}</span>
                    {item.badge && (
                      <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 font-semibold mr-1">
                        {item.badge}
                      </span>
                    )}
                    {isOpen ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                  </>
                )}
              </button>

              {/* Submenu */}
              {!collapsed && isOpen && item.children && (
                <div className="pl-6 pr-1 py-1 space-y-1 border-l border-white/10 ml-5">
                  {item.children.map((sub) => {
                    const SubIcon = sub.icon;
                    const isSubActive = sub.href === pathname;
                    return (
                      <Link
                        key={sub.title}
                        href={sub.href || "#"}
                        onClick={onItemClick}
                        className={cn(
                          "flex items-center justify-between px-3 py-1.5 rounded-lg text-xs font-medium transition-all",
                          isSubActive
                            ? "bg-cyan-500/20 text-cyan-400 font-semibold"
                            : "text-slate-400 hover:text-white hover:bg-white/5"
                        )}
                      >
                        <div className="flex items-center gap-2.5">
                          <SubIcon size={14} className={isSubActive ? "text-cyan-400" : "text-slate-400"} />
                          <span>{sub.title}</span>
                        </div>
                        {sub.badge && (
                          <span className="text-[9px] px-1 rounded bg-slate-800 text-slate-300 font-mono">
                            {sub.badge}
                          </span>
                        )}
                      </Link>
                    );
                  })}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Footer Status */}
      {!collapsed && (
        <div className="p-4 border-t border-white/10 text-xs text-slate-400 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-emerald-400 shadow-[0_0_8px_#10b981]" />
            <span>v1.0.6 (Stable)</span>
          </div>
          <span className="text-[11px] text-slate-500">Made in India ❤️</span>
        </div>
      )}
    </aside>
  );
};
