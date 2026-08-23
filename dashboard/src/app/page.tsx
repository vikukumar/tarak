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
        <img
          src="/assets/icon.png"
          alt="TARAK"
          className="w-14 h-14 object-contain animate-pulse drop-shadow-[0_0_25px_rgba(0,240,255,0.6)]"
        />
        <span className="text-sm font-medium text-slate-400">Connecting to Tarak Control Plane...</span>
      </div>
    </div>
  );
}
