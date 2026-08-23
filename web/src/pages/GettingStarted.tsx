import React, { useState } from "react";
import { Copy, Check, Terminal, Play, Server, Layers, Code, ArrowRight } from "lucide-react";

export const GettingStartedPage: React.FC<{ onNavigate: (tab: string) => void }> = ({ onNavigate }) => {
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const copyCode = (code: string, idx: number) => {
    navigator.clipboard.writeText(code);
    setCopiedIndex(idx);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  const steps = [
    {
      title: "1. One-Line Installation",
      tag: "All Platforms",
      desc: "Install all 5 core binaries (tarak, tarakd, taraks, tarakctl, taraktl) with pre-configured PATH:",
      code: `# Linux & macOS (Universal Bash Installer):
curl -fsSL https://tarak.vikshro.in/install.sh | bash

# Windows 10/11 / Server 2022 (PowerShell):
irm https://tarak.vikshro.in/install.ps1 | iex

# Homebrew (macOS / Linux):
brew install vikukumar/tap/tarak

# Go Install (Source build):
go install github.com/vikukumar/tarak/cmd/tarak@latest`,
    },
    {
      title: "2. Start All-in-One Engine or Initialize Cluster",
      tag: "Sub-180ms Cold Start",
      desc: "Start both the control plane server and local worker runtime in a single standalone binary:",
      code: `# Start all-in-one local single-node cluster (:6443 / :8443):
tarak

# Or run in background as a managed system daemon:
tarak server --data-dir=/var/lib/tarak --log-level=info`,
    },
    {
      title: "3. Deploy Your First Containerized Workload",
      tag: "Declarative YAML",
      desc: "Deploy a high-performance Nginx web service with automated port binding:",
      code: `# Imperative Run:
tarakctl run my-web --image=nginx:alpine --port=80

# Or Declarative YAML Manifest (app.yaml):
cat <<EOF | tarakctl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web-app
  template:
    metadata:
      labels:
        app: web-app
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: web-service
  namespace: default
spec:
  type: ClusterIP
  selector:
    app: web-app
  ports:
  - port: 80
    targetPort: 80
EOF`,
    },
    {
      title: "4. Verify Status, Stream Logs & Interactive Exec",
      tag: "Live Observability",
      desc: "Inspect running pods, stream real stdout logs, and spawn interactive terminal shells:",
      code: `# Inspect active pods & services:
tarakctl get pods -o wide
tarakctl get services

# Stream live container logs:
tarakctl logs -f deployment/web-app

# Interactive shell execution into active container:
tarakctl exec -it pod/web-app-7b9494696-abcd -- /bin/sh`,
    },
    {
      title: "5. Integrate Go Client SDK",
      tag: "pkg/client",
      desc: "Programmatically manage your cluster from any Go microservice with zero external dependencies:",
      code: `go get -u github.com/vikukumar/tarak/pkg/client@latest

package main

import (
    "context"
    "fmt"
    "log"
    "github.com/vikukumar/tarak/pkg/client"
)

func main() {
    c, err := client.NewClient("https://127.0.0.1:8443", client.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }

    pods, err := c.Pods("default").List(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Connected to Tarak! Active pods: %d\\n", len(pods.Items))
}`,
    },
  ];

  return (
    <div className="space-y-10 animate-fade-in max-w-4xl mx-auto">
      {/* Title */}
      <div className="text-center space-y-3">
        <span className="inline-block px-3 py-1 rounded-full bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs font-bold uppercase tracking-wider">
          Quickstart Manual
        </span>
        <h1 className="text-3xl sm:text-5xl font-extrabold text-white tracking-tight">
          Getting Started in <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-indigo-400">30 Seconds</span>
        </h1>
        <p className="text-slate-400 max-w-xl mx-auto text-sm sm:text-base">
          From zero to a fully operational, multi-container cluster in seconds with zero prerequisites.
        </p>
      </div>

      {/* Steps List */}
      <div className="space-y-6">
        {steps.map((step, idx) => (
          <div key={idx} className="p-6 rounded-2xl bg-slate-900/60 border border-white/10 shadow-2xl space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2 className="text-lg sm:text-xl font-bold text-white flex items-center gap-2">
                <span>{step.title}</span>
              </h2>
              <span className="px-2.5 py-0.5 rounded-full bg-cyan-500/15 border border-cyan-500/30 text-cyan-300 text-xs font-semibold">
                {step.tag}
              </span>
            </div>

            <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">{step.desc}</p>

            <div className="relative group rounded-xl bg-[#04060c] border border-white/10 overflow-hidden">
              <div className="flex items-center justify-between px-3 py-2 bg-slate-950/80 border-b border-white/10 text-xs text-slate-400">
                <span className="font-mono">Terminal Command / Manifest</span>
                <button
                  onClick={() => copyCode(step.code, idx)}
                  className="flex items-center gap-1 text-slate-300 hover:text-cyan-300 transition-colors"
                >
                  {copiedIndex === idx ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
                  <span>{copiedIndex === idx ? "Copied" : "Copy"}</span>
                </button>
              </div>
              <pre className="p-4 font-mono text-xs text-cyan-300 overflow-x-auto leading-relaxed whitespace-pre">
                {step.code}
              </pre>
            </div>
          </div>
        ))}
      </div>

      {/* Next Steps CTA */}
      <div className="p-6 rounded-2xl bg-gradient-to-r from-cyan-950/40 to-purple-950/40 border border-cyan-500/30 flex flex-wrap items-center justify-between gap-4">
        <div>
          <h3 className="text-base font-bold text-white">Ready to scale across multiple servers?</h3>
          <p className="text-xs text-slate-400 mt-1">Learn how to configure multi-node clusters with WireGuard encrypted mesh.</p>
        </div>
        <button
          onClick={() => onNavigate("multi-node")}
          className="px-5 py-2.5 rounded-xl bg-cyan-400 hover:bg-cyan-300 text-slate-950 font-bold text-xs flex items-center gap-2 transition-all shadow-[0_0_20px_rgba(0,240,255,0.3)]"
        >
          <span>Multi-Node Guide</span>
          <ArrowRight size={14} />
        </button>
      </div>
    </div>
  );
};
