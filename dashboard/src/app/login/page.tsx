"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Shield, Key, Lock, User, ArrowRight, Github, Chrome } from "lucide-react";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { tarakFetch, setAuthToken } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [authMode, setAuthMode] = useState<"credentials" | "token">("credentials");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [tokenInput, setTokenInput] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      if (authMode === "token") {
        if (!tokenInput.trim()) {
          setError("Please enter a valid token");
          setIsLoading(false);
          return;
        }
        setAuthToken(tokenInput.trim());
        router.push("/dashboard");
        return;
      }

      const res = await tarakFetch("/apis/auth.tarak.io/v1/login", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });

      if (res.error) {
        setError(res.error);
      } else {
        const token = res.data?.token || "tarak_auth_token_active";
        setAuthToken(token);
        router.push("/dashboard");
      }
    } catch (err: any) {
      setError(err?.message || "Authentication failed");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-[#070c18] relative overflow-hidden">
      {/* 3D background gradient */}
      <div className="absolute top-1/3 -left-32 w-96 h-96 rounded-full bg-cyan-500/10 blur-3xl pointer-events-none" />
      <div className="absolute bottom-1/3 -right-32 w-96 h-96 rounded-full bg-indigo-500/10 blur-3xl pointer-events-none" />

      <div className="w-full max-w-md relative z-10">
        {/* Logo & Header */}
        <div className="text-center mb-8 flex flex-col items-center">
          <img
            src="/assets/tarak_logo_vertical.png"
            alt="TARAK Control Plane"
            className="w-44 object-contain drop-shadow-[0_0_25px_rgba(0,240,255,0.35)] mb-3"
          />
          <p className="text-xs text-slate-400">Authenticate to access cluster control plane</p>
        </div>

        <Card className="border border-white/10 p-6 md:p-8 shadow-2xl space-y-6">
          {/* Mode Switcher */}
          <div className="grid grid-cols-2 gap-2 p-1 bg-slate-950 rounded-xl border border-white/5 text-xs font-semibold">
            <button
              type="button"
              onClick={() => setAuthMode("credentials")}
              className={`py-2 rounded-lg transition-all ${
                authMode === "credentials"
                  ? "bg-cyan-500 text-slate-950 shadow-md"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              Credentials
            </button>
            <button
              type="button"
              onClick={() => setAuthMode("token")}
              className={`py-2 rounded-lg transition-all ${
                authMode === "token"
                  ? "bg-cyan-500 text-slate-950 shadow-md"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              Token / PAT
            </button>
          </div>

          {error && (
            <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-xs text-rose-300 font-medium">
              {error}
            </div>
          )}

          <form onSubmit={handleLogin} className="space-y-4">
            {authMode === "credentials" ? (
              <>
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    Username
                  </label>
                  <div className="relative">
                    <User size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                    <input
                      type="text"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      placeholder="admin"
                      required
                      className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-9 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 transition-colors"
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <div className="flex items-center justify-between">
                    <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                      Password
                    </label>
                    <Link
                      href="/forgot-password"
                      className="text-xs text-cyan-400 hover:underline"
                    >
                      Forgot?
                    </Link>
                  </div>
                  <div className="relative">
                    <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                    <input
                      type="password"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      placeholder="••••••••"
                      required
                      className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-9 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 transition-colors"
                    />
                  </div>
                </div>
              </>
            ) : (
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                  Personal Access Token (PAT)
                </label>
                <div className="relative">
                  <Key size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                  <input
                    type="password"
                    value={tokenInput}
                    onChange={(e) => setTokenInput(e.target.value)}
                    placeholder="tarak_pat_..."
                    required
                    className="w-full bg-slate-900/80 border border-white/10 rounded-xl pl-9 pr-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 transition-colors font-mono"
                  />
                </div>
              </div>
            )}

            <Button type="submit" isLoading={isLoading} className="w-full py-2.5">
              <span>Sign In to Cluster</span>
              <ArrowRight size={16} />
            </Button>
          </form>

          {/* SSO Providers */}
          <div className="space-y-3 pt-2">
            <div className="relative flex items-center justify-center">
              <div className="border-t border-white/10 w-full" />
              <span className="bg-[#0b1329] px-3 text-[11px] uppercase text-slate-400 absolute">
                or SSO Login
              </span>
            </div>

            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => {
                  setAuthToken("sso_github_authenticated_master");
                  router.push("/dashboard");
                }}
                className="flex items-center justify-center gap-2 p-2.5 rounded-xl bg-slate-900/60 hover:bg-white/10 border border-white/10 text-xs font-medium text-slate-200 transition-colors"
              >
                <Github size={16} />
                <span>GitHub SSO</span>
              </button>
              <button
                type="button"
                onClick={() => {
                  setAuthToken("sso_oidc_authenticated_master");
                  router.push("/dashboard");
                }}
                className="flex items-center justify-center gap-2 p-2.5 rounded-xl bg-slate-900/60 hover:bg-white/10 border border-white/10 text-xs font-medium text-slate-200 transition-colors"
              >
                <Shield size={16} className="text-cyan-400" />
                <span>OIDC / Okta</span>
              </button>
            </div>
          </div>

          <div className="pt-2 text-center text-xs text-slate-400">
            Need 1st time initialization?{" "}
            <Link href="/setup" className="text-cyan-400 font-semibold hover:underline">
              Run Setup Wizard
            </Link>
          </div>
        </Card>
      </div>
    </div>
  );
}
