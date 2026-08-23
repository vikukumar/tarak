"use client";

import React, { useState, useEffect, useRef } from "react";
import { Terminal as TerminalIcon, Send, RefreshCw, Layers, Box, Trash2, Shield, User, CornerDownLeft } from "lucide-react";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { useCluster } from "@/context/ClusterContext";
import { tarakFetch } from "@/lib/api";

export default function WebTerminalPage() {
  const { selectedNamespace } = useCluster();
  const [pods, setPods] = useState<any[]>([]);
  const [selectedPod, setSelectedPod] = useState<string>("");
  const [selectedContainer, setSelectedContainer] = useState<string>("");
  const [execUser, setExecUser] = useState<"root" | "non-root">("root");
  const [history, setHistory] = useState<Array<{ type: "cmd" | "out" | "err" | "info"; text: string }>>([
    { type: "info", text: "Connected to Tarak Container Runtime Engine (TCR) POSIX Shell." },
    { type: "info", text: "Session authenticated. Type standard Linux commands (e.g. ls -la, ps aux, pwd, env, uname -a, cat /etc/hosts) and press Enter." },
  ]);
  const [inputLine, setInputLine] = useState("");
  const [commandHistory, setCommandHistory] = useState<string[]>([]);
  const [historyIndex, setHistoryIndex] = useState<number>(-1);
  const [isExecuting, setIsExecuting] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Load pods list
  useEffect(() => {
    async function loadPods() {
      const url =
        selectedNamespace === "_all"
          ? "/api/v1/pods"
          : `/api/v1/namespaces/${selectedNamespace}/pods`;
      const res = await tarakFetch(url);
      const items = res.data?.items || [];
      setPods(items);
      if (items.length > 0) {
        setSelectedPod(items[0].metadata?.name);
        const containers = items[0].spec?.containers || [];
        if (containers.length > 0) {
          setSelectedContainer(containers[0].name);
        }
      } else {
        setSelectedPod("");
        setSelectedContainer("");
      }
    }
    loadPods();
  }, [selectedNamespace]);

  useEffect(() => {
    const currentPod = pods.find((p) => p.metadata?.name === selectedPod);
    if (currentPod) {
      const containers = currentPod.spec?.containers || [];
      if (containers.length > 0) {
        setSelectedContainer(containers[0].name);
      }
    }
  }, [selectedPod, pods]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [history]);

  const promptSymbol = execUser === "root" ? "#" : "$";
  const userName = execUser === "root" ? "root" : "appuser";
  const promptText = `${userName}@${selectedPod || "pod"}:/${promptSymbol}`;

  const runCommand = async (cmdToRun: string) => {
    if (!cmdToRun.trim() || !selectedPod) return;
    const cmd = cmdToRun.trim();

    if (cmd === "clear" || cmd === "cls") {
      setHistory([
        { type: "info", text: `Terminal session cleared. Active: ${userName}@${selectedPod} [${selectedContainer || "main"}]` },
      ]);
      return;
    }

    setHistory((prev) => [...prev, { type: "cmd", text: `${promptText} ${cmd}` }]);
    setCommandHistory((prev) => [...prev, cmd]);
    setHistoryIndex(-1);
    setIsExecuting(true);

    try {
      const currentPod = pods.find((p) => p.metadata?.name === selectedPod);
      const ns = currentPod?.metadata?.namespace || (selectedNamespace === "_all" ? "default" : selectedNamespace);
      const containerParam = selectedContainer ? `?container=${encodeURIComponent(selectedContainer)}` : "";

      const res = await tarakFetch(
        `/api/v1/namespaces/${ns}/pods/${selectedPod}/exec${containerParam}`,
        {
          method: "POST",
          body: JSON.stringify({
            command: cmd.split(" "),
            container: selectedContainer,
            user: userName,
            tty: true,
          }),
        }
      );

      const data = res.data || {};
      if (data.stdout) {
        setHistory((prev) => [...prev, { type: "out", text: data.stdout }]);
      }
      if (data.stderr) {
        setHistory((prev) => [...prev, { type: "err", text: data.stderr }]);
      }
      if (!data.stdout && !data.stderr) {
        setHistory((prev) => [
          ...prev,
          { type: "info", text: `Process exited with status ${data.exitCode ?? 0}` },
        ]);
      }
    } catch (err: any) {
      setHistory((prev) => [
        ...prev,
        { type: "err", text: `Exec Error: ${err.message || "Failed to execute command in container"}` },
      ]);
    } finally {
      setIsExecuting(false);
      setTimeout(() => inputRef.current?.focus(), 50);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "ArrowUp") {
      e.preventDefault();
      if (commandHistory.length === 0) return;
      const nextIdx = historyIndex === -1 ? commandHistory.length - 1 : Math.max(0, historyIndex - 1);
      setHistoryIndex(nextIdx);
      setInputLine(commandHistory[nextIdx]);
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      if (historyIndex === -1) return;
      const nextIdx = historyIndex + 1;
      if (nextIdx >= commandHistory.length) {
        setHistoryIndex(-1);
        setInputLine("");
      } else {
        setHistoryIndex(nextIdx);
        setInputLine(commandHistory[nextIdx]);
      }
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputLine.trim()) return;
    const cmd = inputLine;
    setInputLine("");
    runCommand(cmd);
  };

  const currentPodObj = pods.find((p) => p.metadata?.name === selectedPod);
  const containerList = currentPodObj?.spec?.containers || [];

  return (
    <div className="space-y-6 animate-fade-in max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-white flex items-center gap-2.5">
            <TerminalIcon size={24} className="text-cyan-400" />
            <span>Interactive Container Web Terminal</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Direct pseudo-terminal execution and live POSIX shell inside running container processes.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {/* Pod Selector */}
          {pods.length > 0 && (
            <div className="flex items-center gap-2 bg-slate-900 border border-white/10 px-3 py-1.5 rounded-lg text-xs">
              <Box size={14} className="text-cyan-400" />
              <select
                value={selectedPod}
                onChange={(e) => setSelectedPod(e.target.value)}
                className="bg-transparent text-white outline-none cursor-pointer font-mono font-bold"
              >
                {pods.map((p) => (
                  <option
                    key={p.metadata?.name}
                    value={p.metadata?.name}
                    className="bg-slate-950 text-white"
                  >
                    {p.metadata?.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Container Selector */}
          {containerList.length > 0 && (
            <div className="flex items-center gap-2 bg-slate-900 border border-white/10 px-3 py-1.5 rounded-lg text-xs">
              <Layers size={14} className="text-purple-400" />
              <select
                value={selectedContainer}
                onChange={(e) => setSelectedContainer(e.target.value)}
                className="bg-transparent text-purple-300 outline-none cursor-pointer font-mono font-semibold"
              >
                {containerList.map((c: any) => (
                  <option key={c.name} value={c.name} className="bg-slate-950 text-white">
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* User Mode (Root / Non-Root) Toggle */}
          <div className="flex items-center gap-1 bg-slate-900 border border-white/10 p-1 rounded-lg text-xs font-mono">
            <button
              onClick={() => setExecUser("root")}
              className={`px-2 py-0.5 rounded transition-all ${
                execUser === "root"
                  ? "bg-rose-500/20 text-rose-300 border border-rose-500/30 font-bold"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              root (uid 0)
            </button>
            <button
              onClick={() => setExecUser("non-root")}
              className={`px-2 py-0.5 rounded transition-all ${
                execUser === "non-root"
                  ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 font-bold"
                  : "text-slate-400 hover:text-white"
              }`}
            >
              appuser (uid 1000)
            </button>
          </div>

          <Button
            variant="secondary"
            size="sm"
            onClick={() =>
              setHistory([
                { type: "info", text: "Terminal cleared." },
                { type: "info", text: `Active container: ${selectedContainer || "default"} on pod ${selectedPod}` },
              ])
            }
            title="Clear terminal"
          >
            <Trash2 size={14} />
          </Button>
        </div>
      </div>

      {/* Shortcut Badges */}
      <div className="flex flex-wrap items-center gap-2 text-xs">
        <span className="text-slate-400 font-mono">Quick POSIX commands:</span>
        {["ls -la", "pwd", "ps aux", "uptime", "uname -a", "whoami", "id", "cat /etc/hosts", "cat /etc/os-release", "env"].map((sc) => (
          <button
            key={sc}
            onClick={() => runCommand(sc)}
            className="px-2.5 py-1 rounded bg-slate-900/80 hover:bg-cyan-500/20 border border-white/10 hover:border-cyan-500/30 text-cyan-300 font-mono text-[11px] transition-colors"
          >
            {sc}
          </button>
        ))}
      </div>

      {/* Terminal View Container */}
      <Card className="p-0 border-white/10 overflow-hidden shadow-2xl flex flex-col h-[540px]">
        {/* Terminal Header */}
        <div className="p-3 bg-slate-950 border-b border-white/10 flex items-center justify-between text-xs text-slate-400 font-mono">
          <div className="flex items-center gap-2">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500/80" />
            <span className="w-2.5 h-2.5 rounded-full bg-amber-500/80" />
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-500/80" />
            <span className="ml-2 text-slate-300 font-bold">{promptText}</span>
            {selectedContainer && (
              <Badge variant="purple" className="text-[10px] py-0 px-2 font-mono">
                container: {selectedContainer}
              </Badge>
            )}
            <Badge variant={execUser === "root" ? "rose" : "emerald"} className="text-[10px] py-0 px-2 font-mono">
              user: {userName}
            </Badge>
          </div>
          <span className="text-cyan-400 font-semibold text-[11px]">TCR POSIX Shell Active</span>
        </div>

        {/* Output Stream */}
        <div
          className="flex-1 p-4 bg-[#050914] text-xs font-mono overflow-y-auto space-y-1.5 leading-relaxed select-text"
          onClick={() => inputRef.current?.focus()}
        >
          {history.map((h, i) => (
            <div key={i} className="whitespace-pre-wrap break-all">
              {h.type === "cmd" && (
                <span className="text-white font-bold">{h.text}</span>
              )}
              {h.type === "out" && (
                <span className="text-cyan-300">{h.text}</span>
              )}
              {h.type === "err" && (
                <span className="text-rose-400">{h.text}</span>
              )}
              {h.type === "info" && (
                <span className="text-slate-500 italic">{h.text}</span>
              )}
            </div>
          ))}
          <div ref={bottomRef} />
        </div>

        {/* Input Prompt Form */}
        <form
          onSubmit={handleSubmit}
          className="p-3 bg-slate-950/90 border-t border-white/10 flex items-center gap-3 font-mono text-xs"
        >
          <span className="text-cyan-400 font-bold flex-shrink-0">
            {promptText}
          </span>
          <input
            ref={inputRef}
            type="text"
            value={inputLine}
            onChange={(e) => setInputLine(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={isExecuting || !selectedPod}
            placeholder={
              !selectedPod
                ? "No pod available in selected namespace..."
                : isExecuting
                ? "Executing..."
                : "Type container command here (e.g. ls -la, ps aux, cat /etc/hosts)..."
            }
            className="flex-1 bg-transparent text-white placeholder:text-slate-600 outline-none"
            autoFocus
          />
          <Button
            type="submit"
            size="sm"
            disabled={isExecuting || !inputLine.trim() || !selectedPod}
            className="bg-cyan-600 hover:bg-cyan-500 text-white px-3 py-1"
          >
            <Send size={12} className="mr-1" /> Run
          </Button>
        </form>
      </Card>
    </div>
  );
}
