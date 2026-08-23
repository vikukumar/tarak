"use client";

import React, { useState } from "react";
import Link from "next/link";
import { ArrowLeft, Mail, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-[#070c18]">
      <div className="w-full max-w-md">
        <Card className="border border-white/10 p-8 shadow-2xl space-y-6">
          <Link
            href="/login"
            className="inline-flex items-center gap-2 text-xs text-slate-400 hover:text-white transition-colors"
          >
            <ArrowLeft size={14} />
            <span>Back to Sign In</span>
          </Link>

          <div>
            <h1 className="text-xl font-bold text-white">Reset Cluster Access</h1>
            <p className="text-xs text-slate-400 mt-1">
              Enter your administrative email or generate a new bootstrap PAT via CLI
            </p>
          </div>

          {submitted ? (
            <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-xs text-emerald-300 space-y-2">
              <div className="flex items-center gap-2 font-semibold text-sm">
                <CheckCircle2 size={16} />
                <span>Recovery Instructions Sent</span>
              </div>
              <p className="text-slate-300">
                If the email is associated with a cluster admin, recovery keys have been dispatched. Alternatively, use:
              </p>
              <div className="p-2 rounded bg-slate-950 text-cyan-300 font-mono text-[11px]">
                tarakctl config use-context admin
              </div>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                  Admin Email
                </label>
                <div className="relative">
                  <Mail size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="admin@tarak.io"
                    required
                    className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-9 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 transition-colors"
                  />
                </div>
              </div>

              <Button type="submit" className="w-full py-2.5">
                Request Password Reset
              </Button>
            </form>
          )}
        </Card>
      </div>
    </div>
  );
}
