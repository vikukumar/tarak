"use client";

import React, { useState } from "react";
import { Key, Plus, Trash2, Copy, Check } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { DataTable, Column } from "@/components/ui/DataTable";
import { Modal } from "@/components/ui/Modal";

export default function PatTokensPage() {
  const [tokens, setTokens] = useState<any[]>([
    {
      id: "1",
      name: "super-admin-master-pat",
      prefix: "tarak_pat_7a8b9c...",
      scope: "cluster-admin (Full Access)",
      created: "2026-08-22",
      expires: "Never",
    },
    {
      id: "2",
      name: "ci-cd-github-actions",
      prefix: "tarak_pat_3e4f5a...",
      scope: "deployment-writer",
      created: "2026-08-23",
      expires: "90 Days",
    },
  ]);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newTokenName, setNewTokenName] = useState("");
  const [generatedSecret, setGeneratedSecret] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleGenerate = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTokenName.trim()) return;

    const secret = `tarak_pat_${Math.random().toString(36).substring(2, 15)}_${Math.random().toString(36).substring(2, 15)}`;
    setGeneratedSecret(secret);
    setTokens((prev) => [
      ...prev,
      {
        id: String(Date.now()),
        name: newTokenName.trim(),
        prefix: secret.substring(0, 16) + "...",
        scope: "cluster-admin",
        created: new Date().toISOString().split("T")[0],
        expires: "Never",
      },
    ]);
  };

  const columns: Column<any>[] = [
    {
      key: "name",
      header: "Token Description",
      sortable: true,
      render: (t) => (
        <div className="flex items-center gap-2">
          <Key size={16} className="text-cyan-400" />
          <span className="font-semibold text-white">{t.name}</span>
        </div>
      ),
    },
    {
      key: "prefix",
      header: "Token Key",
      render: (t) => <span className="font-mono text-xs text-slate-300">{t.prefix}</span>,
    },
    {
      key: "scope",
      header: "RBAC Scope",
      render: (t) => <Badge variant="indigo">{t.scope}</Badge>,
    },
    {
      key: "expires",
      header: "Expiration",
      render: (t) => <Badge variant="emerald">{t.expires}</Badge>,
    },
    {
      key: "actions",
      header: "Actions",
      className: "text-right",
      render: (t) => (
        <button
          onClick={() => setTokens((prev) => prev.filter((it) => it.id !== t.id))}
          className="p-1.5 rounded-lg bg-white/5 hover:bg-rose-500/20 text-rose-400 transition-colors"
          title="Revoke Token"
        >
          <Trash2 size={14} />
        </button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Key size={22} className="text-cyan-400" />
            <span>Personal Access Tokens (PAT)</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Manage programmatic bearer tokens for remote CLI access (`tarakctl login`) and automation CI/CD
          </p>
        </div>

        <Button size="sm" onClick={() => setIsModalOpen(true)}>
          <Plus size={14} />
          <span>Generate New Token</span>
        </Button>
      </div>

      <DataTable columns={columns} data={tokens} searchKey="name" />

      {/* Generate Token Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setGeneratedSecret(null);
          setNewTokenName("");
        }}
        title="Generate Personal Access Token"
        maxWidth="md"
      >
        {generatedSecret ? (
          <div className="space-y-4">
            <p className="text-xs text-emerald-300">
              Make sure to copy your personal access token now. You won&apos;t be able to see it again!
            </p>
            <div className="flex items-center gap-2 p-3 rounded-xl bg-slate-950 border border-cyan-500/30 text-xs font-mono text-cyan-300 break-all">
              <span className="flex-1">{generatedSecret}</span>
              <button
                onClick={() => {
                  navigator.clipboard.writeText(generatedSecret);
                  setCopied(true);
                  setTimeout(() => setCopied(false), 2000);
                }}
                className="p-1.5 rounded-lg bg-white/10 hover:bg-white/20 text-white transition-colors"
              >
                {copied ? <Check size={14} className="text-emerald-400" /> : <Copy size={14} />}
              </button>
            </div>
            <Button onClick={() => setIsModalOpen(false)} className="w-full">
              Done
            </Button>
          </div>
        ) : (
          <form onSubmit={handleGenerate} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">Token Description</label>
              <input
                type="text"
                value={newTokenName}
                onChange={(e) => setNewTokenName(e.target.value)}
                placeholder="e.g. laptop-tarakctl-access"
                required
                className="w-full bg-slate-900 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-white focus:outline-none focus:border-cyan-400 font-mono"
              />
            </div>
            <Button type="submit" className="w-full">
              Generate Token
            </Button>
          </form>
        )}
      </Modal>
    </div>
  );
}
