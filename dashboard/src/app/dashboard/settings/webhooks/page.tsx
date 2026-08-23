"use client";

import React, { useState } from "react";
import {
  Zap,
  Plus,
  Search,
  RefreshCw,
  Clock,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  ExternalLink,
  Send,
  Trash2,
  FileCode,
  Shield,
} from "lucide-react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";

interface WebhookEndpoint {
  id: string;
  name: string;
  url: string;
  type: string;
  events: string[];
  enabled: boolean;
}

interface DeliveryRecord {
  id: string;
  webhookName: string;
  event: string;
  statusCode: number;
  success: boolean;
  latencyMs: number;
  time: string;
  payload: string;
}

const initialWebhooks: WebhookEndpoint[] = [
  {
    id: "wh-1",
    name: "Slack Incident Alerts",
    url: "https://hooks.slack.com/services/T00/B00/XXXXX",
    type: "EventNotification",
    events: ["pod.crash", "node.notready", "policy.violation"],
    enabled: true,
  },
  {
    id: "wh-2",
    name: "PagerDuty Critical Dispatcher",
    url: "https://events.pagerduty.com/v2/enqueue",
    type: "EventNotification",
    events: ["node.notready", "mesh.partition"],
    enabled: true,
  },
];

const initialDeliveries: DeliveryRecord[] = [
  {
    id: "del-101",
    webhookName: "Slack Incident Alerts",
    event: "policy.violation",
    statusCode: 200,
    success: true,
    latencyMs: 124.5,
    time: "8 mins ago",
    payload: `{"event":"policy.violation","resource":"pod/default/frontend","rule":"disallow-default-namespace"}`,
  },
  {
    id: "del-102",
    webhookName: "PagerDuty Critical Dispatcher",
    event: "hpa.scale",
    statusCode: 202,
    success: true,
    latencyMs: 88.2,
    time: "24 mins ago",
    payload: `{"event":"hpa.scale","resource":"deployment/production/storefront","fromReplicas":2,"toReplicas":5}`,
  },
];

export default function WebhooksPage() {
  const [webhooks, setWebhooks] = useState<WebhookEndpoint[]>(initialWebhooks);
  const [deliveries, setDeliveries] = useState<DeliveryRecord[]>(initialDeliveries);
  const [isTriggering, setIsTriggering] = useState<string | null>(null);

  const testTrigger = (id: string, name: string) => {
    setIsTriggering(id);
    setTimeout(() => {
      const newDel: DeliveryRecord = {
        id: `del-${Date.now()}`,
        webhookName: name,
        event: "manual.ping",
        statusCode: 200,
        success: true,
        latencyMs: 95.4,
        time: "Just now",
        payload: `{"event":"manual.ping","timestamp":"${new Date().toISOString()}","source":"dashboard-ui"}`,
      };
      setDeliveries([newDel, ...deliveries]);
      setIsTriggering(null);
    }, 800);
  };

  const toggleWebhook = (id: string) => {
    setWebhooks((prev) =>
      prev.map((w) => (w.id === id ? { ...w, enabled: !w.enabled } : w))
    );
  };

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-400">
              <Zap size={22} />
            </span>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">
              Webhook Subscriptions <span className="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-indigo-300 to-purple-400">& Delivery Stream</span>
            </h1>
          </div>
          <p className="text-xs sm:text-sm text-slate-400 mt-1">
            Real-time event notification webhooks for Slack, Discord, PagerDuty, and custom admission dispatchers.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Button size="sm" className="bg-gradient-to-r from-cyan-600 to-purple-600 text-white shadow-lg shadow-cyan-950/40">
            <Plus size={14} className="mr-1.5" /> Register Webhook
          </Button>
        </div>
      </div>

      {/* Webhook Endpoints Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {webhooks.map((wh) => (
          <div
            key={wh.id}
            className="p-5 rounded-2xl bg-slate-900/70 border border-white/10 shadow-xl space-y-4 hover:border-white/20 transition-all"
          >
            <div className="flex items-start justify-between gap-2">
              <div className="space-y-1">
                <span className="font-bold text-white text-sm">{wh.name}</span>
                <p className="text-xs text-cyan-300 font-mono truncate max-w-sm">{wh.url}</p>
              </div>
              <button
                onClick={() => toggleWebhook(wh.id)}
                className={`px-2.5 py-1 rounded-lg text-xs font-bold font-mono border transition-all ${
                  wh.enabled
                    ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/30"
                    : "bg-slate-800 text-slate-500 border-white/10"
                }`}
              >
                {wh.enabled ? "Active" : "Disabled"}
              </button>
            </div>

            <div className="flex flex-wrap items-center gap-1.5">
              <span className="text-[11px] text-slate-400 font-sans">Events:</span>
              {wh.events.map((e, idx) => (
                <span key={idx} className="px-2 py-0.5 rounded text-[10px] font-mono bg-purple-500/10 text-purple-300 border border-purple-500/20">
                  {e}
                </span>
              ))}
            </div>

            <div className="flex items-center justify-between pt-2 border-t border-white/5">
              <Button
                size="sm"
                variant="outline"
                onClick={() => testTrigger(wh.id, wh.name)}
                disabled={isTriggering === wh.id}
                className="text-xs text-cyan-300 border-cyan-500/30"
              >
                <Send size={12} className="mr-1" />
                {isTriggering === wh.id ? "Firing Ping..." : "Test Trigger"}
              </Button>
              <Button size="sm" variant="ghost" className="text-rose-400 hover:text-rose-300 text-xs">
                <Trash2 size={13} className="mr-1" /> Remove
              </Button>
            </div>
          </div>
        ))}
      </div>

      {/* Deliveries History */}
      <div className="space-y-3">
        <h2 className="text-sm font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5 font-mono">
          <Clock size={14} className="text-purple-400" /> Recent Webhook Executions & Audit Logs
        </h2>

        <div className="overflow-x-auto rounded-2xl border border-white/10 bg-slate-900/60 shadow-xl">
          <table className="w-full text-left text-xs border-collapse font-mono">
            <thead>
              <tr className="bg-slate-950/90 text-slate-300 uppercase tracking-wider font-bold border-b border-white/10 font-sans">
                <th className="p-3.5">Timestamp</th>
                <th className="p-3.5">Endpoint</th>
                <th className="p-3.5">Event Type</th>
                <th className="p-3.5">Status Code</th>
                <th className="p-3.5">Latency</th>
                <th className="p-3.5">Payload Preview</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {deliveries.map((del) => (
                <tr key={del.id} className="hover:bg-white/[0.02] transition-colors">
                  <td className="p-3.5 text-slate-400 whitespace-nowrap">{del.time}</td>
                  <td className="p-3.5 text-white font-bold whitespace-nowrap">{del.webhookName}</td>
                  <td className="p-3.5 text-purple-300 font-bold whitespace-nowrap">{del.event}</td>
                  <td className="p-3.5 whitespace-nowrap">
                    <Badge variant={del.success ? "emerald" : "rose"} dot>
                      {del.statusCode} {del.success ? "OK" : "Error"}
                    </Badge>
                  </td>
                  <td className="p-3.5 text-emerald-400 whitespace-nowrap">{del.latencyMs}ms</td>
                  <td className="p-3.5 text-slate-400 truncate max-w-xs">{del.payload}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
