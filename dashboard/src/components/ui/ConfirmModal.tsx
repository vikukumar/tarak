"use client";

import React from "react";
import { AlertTriangle, Trash2, X } from "lucide-react";
import { Button } from "@/components/ui/Button";

interface ConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: "danger" | "warning" | "primary";
  isLoading?: boolean;
}

export const ConfirmModal: React.FC<ConfirmModalProps> = ({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = "Delete Resource",
  cancelText = "Cancel",
  variant = "danger",
  isLoading = false,
}) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 animate-fade-in">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/75 backdrop-blur-sm transition-opacity"
        onClick={!isLoading ? onClose : undefined}
      />

      {/* Modal Card */}
      <div className="relative w-full max-w-md bg-[#0b1329] border border-white/15 rounded-2xl shadow-2xl overflow-hidden z-10 text-slate-100 animate-slide-in">
        <div className="p-6 space-y-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-center gap-3">
              <div
                className={`w-11 h-11 rounded-xl flex items-center justify-center font-bold border ${
                  variant === "danger"
                    ? "bg-rose-500/15 border-rose-500/30 text-rose-400 shadow-[0_0_15px_rgba(244,63,94,0.2)]"
                    : "bg-amber-500/15 border-amber-500/30 text-amber-400 shadow-[0_0_15px_rgba(245,158,11,0.2)]"
                }`}
              >
                {variant === "danger" ? (
                  <Trash2 size={20} />
                ) : (
                  <AlertTriangle size={20} />
                )}
              </div>
              <div>
                <h3 className="text-base font-bold text-white tracking-tight">
                  {title}
                </h3>
                <span className="text-[11px] font-mono text-slate-400">
                  Confirmation Required
                </span>
              </div>
            </div>

            <button
              onClick={onClose}
              disabled={isLoading}
              className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition-colors"
            >
              <X size={16} />
            </button>
          </div>

          <p className="text-xs text-slate-300 leading-relaxed bg-slate-950/50 p-3.5 rounded-xl border border-white/5 font-mono">
            {message}
          </p>

          <div className="flex items-center justify-end gap-2.5 pt-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={onClose}
              disabled={isLoading}
            >
              {cancelText}
            </Button>
            <Button
              variant={variant === "danger" ? "danger" : "primary"}
              size="sm"
              onClick={onConfirm}
              isLoading={isLoading}
            >
              {confirmText}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
};
