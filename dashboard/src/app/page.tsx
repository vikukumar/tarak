"use client";

import { useEffect } from "react";
import { getAuthToken } from "@/lib/api";

export default function RootPage() {
  useEffect(() => {
    const token = getAuthToken();
    if (token) {
      window.location.href = "/dashboard/";
    } else {
      window.location.href = "/dashboard/";
    }
  }, []);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#070c18]">
      <div className="flex flex-col items-center gap-4">
        <div className="w-12 h-12 rounded-2xl bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center shadow-[0_0_25px_rgba(0,240,255,0.5)] animate-pulse">
          <span className="font-bold text-slate-950 text-2xl">T</span>
        </div>
        <span className="text-sm font-medium text-slate-400">Connecting to Tarak Control Plane...</span>
      </div>
    </div>
  );
}
