import React from "react";
import { cn } from "@/lib/utils";

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  interactive?: boolean;
}

export const Card: React.FC<CardProps> = ({
  className,
  interactive = false,
  children,
  ...props
}) => {
  return (
    <div
      className={cn(
        interactive ? "glass-panel-interactive" : "glass-panel",
        "rounded-xl p-5 text-card-foreground",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
};

export const CardHeader: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}) => (
  <div className={cn("flex items-center justify-between pb-3 mb-3 border-b border-white/5", className)} {...props}>
    {children}
  </div>
);

export const CardTitle: React.FC<React.HTMLAttributes<HTMLHeadingElement>> = ({
  className,
  children,
  ...props
}) => (
  <h3 className={cn("text-base font-semibold text-white tracking-wide flex items-center gap-2", className)} {...props}>
    {children}
  </h3>
);
