import React from "react";
import { cn } from "@/lib/utils";

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: "cyan" | "emerald" | "amber" | "rose" | "indigo" | "muted";
  dot?: boolean;
}

export const Badge: React.FC<BadgeProps> = ({
  className,
  variant = "cyan",
  dot = false,
  children,
  ...props
}) => {
  const variantStyles = {
    cyan: "bg-cyan-500/10 text-cyan-400 border-cyan-500/30",
    emerald: "bg-emerald-500/10 text-emerald-400 border-emerald-500/30",
    amber: "bg-amber-500/10 text-amber-400 border-amber-500/30",
    rose: "bg-rose-500/10 text-rose-400 border-rose-500/30",
    indigo: "bg-indigo-500/10 text-indigo-400 border-indigo-500/30",
    muted: "bg-slate-800/60 text-slate-300 border-white/10",
  };

  const dotStyles = {
    cyan: "bg-cyan-400 shadow-[0_0_8px_#00f0ff]",
    emerald: "bg-emerald-400 shadow-[0_0_8px_#10b981]",
    amber: "bg-amber-400 shadow-[0_0_8px_#f59e0b]",
    rose: "bg-rose-400 shadow-[0_0_8px_#f43f5e]",
    indigo: "bg-indigo-400 shadow-[0_0_8px_#6366f1]",
    muted: "bg-slate-400",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold border tracking-wide",
        variantStyles[variant],
        className
      )}
      {...props}
    >
      {dot && <span className={cn("w-1.5 h-1.5 rounded-full", dotStyles[variant])} />}
      {children}
    </span>
  );
};
