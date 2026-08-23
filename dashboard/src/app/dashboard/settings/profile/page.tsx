"use client";

import React, { useState } from "react";
import { User, Shield, Key, Mail, CheckCircle2 } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { useAuth } from "@/hooks/useAuth";

export default function ProfilePage() {
  const { user } = useAuth();
  const [saved, setSaved] = useState(false);

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    setSaved(true);
    setTimeout(() => setSaved(false), 2500);
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-xl font-bold text-white flex items-center gap-2">
          <User size={22} className="text-indigo-400" />
          <span>User Account & Security Profile</span>
        </h1>
        <p className="text-xs text-slate-400 mt-1">
          Manage administrative credentials and role assignment
        </p>
      </div>

      {saved && (
        <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-xs text-emerald-300 flex items-center gap-2">
          <CheckCircle2 size={16} />
          <span>Profile preferences saved</span>
        </div>
      )}

      <Card className="p-6 space-y-6">
        <div className="flex items-center gap-4 pb-6 border-b border-white/10">
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center font-bold text-slate-950 text-2xl shadow-[0_0_20px_rgba(0,240,255,0.3)]">
            {user?.username?.charAt(0).toUpperCase() || "A"}
          </div>
          <div className="space-y-1">
            <h3 className="font-bold text-white text-base">{user?.username || "Super-Admin"}</h3>
            <div className="flex items-center gap-2">
              <Badge variant="emerald" dot>Active Session</Badge>
              <Badge variant="indigo">cluster-admin</Badge>
            </div>
          </div>
        </div>

        <form onSubmit={handleSave} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Username</label>
            <input
              type="text"
              defaultValue={user?.username || "admin"}
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2 text-sm text-white focus:outline-none focus:border-cyan-400"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Email Address</label>
            <input
              type="email"
              defaultValue="admin@tarak.io"
              className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2 text-sm text-white focus:outline-none focus:border-cyan-400"
            />
          </div>

          <Button type="submit">Save Changes</Button>
        </form>
      </Card>
    </div>
  );
}
