"use client";

import React, { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Sparkles,
  Layers,
  Radio,
  Terminal,
  Menu,
  X,
  ChevronRight,
  Shield,
  Server,
  Network,
  Activity,
  Settings,
} from "lucide-react";
import { navigationConfig } from "./Sidebar";
import { cn } from "@/lib/utils";

export const MobileBottomNav: React.FC = () => {
  const pathname = usePathname();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const mainTabs = [
    { label: "Overview", href: "/dashboard", icon: Sparkles },
    { label: "Workloads", href: "/dashboard/workloads/pods", icon: Layers },
    { label: "Mesh", href: "/dashboard/mesh/overview", icon: Radio },
    { label: "Terminal", href: "/dashboard/devtools/terminal", icon: Terminal },
  ];

  return (
    <>
      {/* Phone App Bottom Navigation Bar */}
      <div className="md:hidden fixed bottom-0 left-0 right-0 h-16 bg-[#070c18]/95 backdrop-blur-2xl border-t border-white/10 z-40 px-3 flex items-center justify-around">
        {mainTabs.map((tab) => {
          const Icon = tab.icon;
          const isActive = pathname === tab.href;
          return (
            <Link
              key={tab.label}
              href={tab.href}
              className={cn(
                "flex flex-col items-center justify-center flex-1 py-1 transition-colors",
                isActive ? "text-cyan-400 font-semibold" : "text-slate-400 hover:text-white"
              )}
            >
              <Icon size={18} className={cn(isActive ? "text-cyan-400" : "")} />
              <span className="text-[10px] mt-1">{tab.label}</span>
            </Link>
          );
        })}

        {/* More Drawer Button */}
        <button
          onClick={() => setDrawerOpen(true)}
          className={cn(
            "flex flex-col items-center justify-center flex-1 py-1 transition-colors",
            drawerOpen ? "text-cyan-400 font-semibold" : "text-slate-400 hover:text-white"
          )}
        >
          <Menu size={18} />
          <span className="text-[10px] mt-1">Menu</span>
        </button>
      </div>

      {/* Slide-up Mobile Drawer */}
      {drawerOpen && (
        <div className="md:hidden fixed inset-0 z-50 flex flex-col justify-end">
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black/80 backdrop-blur-sm transition-opacity"
            onClick={() => setDrawerOpen(false)}
          />

          {/* Drawer Content */}
          <div className="relative bg-[#0b1329] border-t border-white/15 rounded-t-3xl p-5 max-h-[80vh] overflow-y-auto z-10 space-y-4 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-white/10">
              <div className="flex items-center gap-2">
                <div className="w-7 h-7 rounded-lg bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center font-bold text-slate-950 text-xs">
                  T
                </div>
                <span className="font-bold text-white text-sm">Cluster Navigation</span>
              </div>
              <button
                onClick={() => setDrawerOpen(false)}
                className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-white/10"
              >
                <X size={18} />
              </button>
            </div>

            {/* Hierarchical menu list */}
            <div className="space-y-3">
              {navigationConfig.map((item) => {
                const Icon = item.icon;
                if (!item.children) {
                  return (
                    <Link
                      key={item.title}
                      href={item.href || "#"}
                      onClick={() => setDrawerOpen(false)}
                      className="flex items-center justify-between p-3 rounded-xl bg-white/5 border border-white/5 text-sm font-medium text-white"
                    >
                      <div className="flex items-center gap-3">
                        <Icon size={18} className="text-cyan-400" />
                        <span>{item.title}</span>
                      </div>
                      <ChevronRight size={16} className="text-slate-400" />
                    </Link>
                  );
                }

                return (
                  <div key={item.title} className="space-y-1.5">
                    <div className="text-xs font-bold text-slate-400 uppercase tracking-wider px-1 pt-2">
                      {item.title}
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      {item.children.map((sub) => {
                        const SubIcon = sub.icon;
                        return (
                          <Link
                            key={sub.title}
                            href={sub.href || "#"}
                            onClick={() => setDrawerOpen(false)}
                            className="flex items-center gap-2 p-2.5 rounded-xl bg-white/5 border border-white/5 text-xs text-slate-200 hover:bg-white/10"
                          >
                            <SubIcon size={14} className="text-cyan-400 flex-shrink-0" />
                            <span className="truncate">{sub.title}</span>
                          </Link>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}
    </>
  );
};
