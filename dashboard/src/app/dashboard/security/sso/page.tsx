"use client";

import React, { useState } from "react";
import { Key, Shield, CheckCircle2, Plus, Github } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

export default function SsoPage() {
  const providers = [
    {
      name: "GitHub Enterprise OAuth",
      type: "OAuth 2.0",
      status: "Configured",
      clientId: "tarak-github-app-client-id",
      issuer: "https://github.com",
    },
    {
      name: "Corporate Okta OIDC",
      type: "OIDC / OpenID Connect",
      status: "Active",
      clientId: "0oa83kdf83kd...",
      issuer: "https://auth.company.okta.com",
    },
    {
      name: "Google Cloud Identity",
      type: "OAuth 2.0",
      status: "Available",
      clientId: "apps.googleusercontent.com",
      issuer: "https://accounts.google.com",
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2">
            <Key size={22} className="text-indigo-400" />
            <span>Single Sign-On (SSO) & Identity Providers</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Enterprise authentication connectors, SAML 2.0, and OAuth/OIDC providers
          </p>
        </div>

        <Button size="sm">
          <Plus size={14} />
          <span>Add Identity Provider</span>
        </Button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {providers.map((p) => (
          <Card key={p.name} interactive className="p-5 space-y-3">
            <div className="flex items-center justify-between">
              <span className="font-bold text-white text-sm">{p.name}</span>
              <Badge variant="emerald" dot>
                {p.status}
              </Badge>
            </div>
            <div className="space-y-1 text-xs text-slate-400">
              <div>Type: <span className="text-cyan-300 font-mono">{p.type}</span></div>
              <div>Issuer: <span className="text-slate-300 font-mono">{p.issuer}</span></div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
