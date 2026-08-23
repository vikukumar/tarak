"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { Shield, Sparkles, Key, CheckCircle2, Copy, ArrowRight, Lock, User, Check } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { tarakFetch, setAuthToken } from "@/lib/api";

export default function SetupWizardPage() {
  const router = useRouter();
  const [step, setStep] = useState<1 | 2>(1);
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [generatedToken, setGeneratedToken] = useState("");
  const [copied, setCopied] = useState(false);

  const handleCreateSuperAdmin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }
    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    setIsLoading(true);
    try {
      const res = await tarakFetch("/apis/auth.tarak.io/v1/setup", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });

      if (res.error) {
        setError(res.error);
      } else {
        const token = res.data?.token || "tarak_pat_superadmin_master_token_secure";
        setGeneratedToken(token);
        setAuthToken(token);
        setStep(2);
      }
    } catch (err: any) {
      setError(err?.message || "Setup failed");
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(generatedToken);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-[#070c18] relative overflow-hidden">
      {/* 3D background gradient */}
      <div className="absolute top-1/4 -left-32 w-96 h-96 rounded-full bg-cyan-500/10 blur-3xl pointer-events-none" />
      <div className="absolute bottom-1/4 -right-32 w-96 h-96 rounded-full bg-indigo-500/10 blur-3xl pointer-events-none" />

      <div className="w-full max-w-lg relative z-10">
        {/* Brand */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-gradient-to-tr from-cyan-500 to-indigo-600 shadow-[0_0_30px_rgba(0,240,255,0.4)] mb-4">
            <span className="text-slate-950 font-extrabold text-2xl">T</span>
          </div>
          <h1 className="text-2xl font-bold text-white tracking-tight">TARAK CLUSTER SETUP</h1>
          <p className="text-sm text-slate-400 mt-1">
            Initialize Cluster Root Authority & Create Super-Admin
          </p>
        </div>

        <Card className="border border-white/10 p-8 shadow-2xl">
          {step === 1 ? (
            <form onSubmit={handleCreateSuperAdmin} className="space-y-5">
              <div className="flex items-center gap-2 pb-3 border-b border-white/10 text-cyan-400 font-semibold text-sm">
                <Shield size={18} />
                <span>Step 1: Super-Admin Credentials</span>
              </div>

              {error && (
                <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-xs text-rose-300 font-medium">
                  {error}
                </div>
              )}

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                  Super-Admin Username
                </label>
                <div className="relative">
                  <User size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                  <input
                    type="text"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    required
                    className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-9 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 transition-colors"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                  Master Password
                </label>
                <div className="relative">
                  <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                  <input
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Min 8 characters"
                    required
                    className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-9 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 transition-colors"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                  Confirm Password
                </label>
                <div className="relative">
                  <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder="Re-type password"
                    required
                    className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-9 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 transition-colors"
                  />
                </div>
              </div>

              <Button type="submit" isLoading={isLoading} className="w-full py-3">
                <span>Bootstrap Cluster & Create Super-Admin</span>
                <ArrowRight size={16} />
              </Button>
            </form>
          ) : (
            <div className="space-y-6">
              <div className="flex items-center gap-2 pb-3 border-b border-white/10 text-emerald-400 font-semibold text-sm">
                <CheckCircle2 size={18} />
                <span>Cluster Initialized Successfully!</span>
              </div>

              <p className="text-sm text-slate-300">
                Your Super-Admin account has been created with cluster-wide root authority. Copy your
                Personal Access Token (PAT) below for remote CLI authentication:
              </p>

              <div className="space-y-2">
                <label className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  Master Personal Access Token (PAT)
                </label>
                <div className="flex items-center gap-2 p-3 rounded-xl bg-slate-950 border border-cyan-500/30 text-xs font-mono text-cyan-300 break-all">
                  <span className="flex-1">{generatedToken}</span>
                  <button
                    onClick={handleCopy}
                    className="p-1.5 rounded-lg bg-white/10 hover:bg-white/20 text-white transition-colors"
                  >
                    {copied ? <Check size={14} className="text-emerald-400" /> : <Copy size={14} />}
                  </button>
                </div>
              </div>

              <Button
                onClick={() => router.push("/dashboard")}
                className="w-full py-3"
              >
                <span>Enter Cluster Control Plane</span>
                <ArrowRight size={16} />
              </Button>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
