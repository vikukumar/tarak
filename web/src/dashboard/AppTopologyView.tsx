import React, { useState } from 'react';
import { 
  GitPullRequest, 
  RefreshCw, 
  Play, 
  RotateCcw, 
  Sliders, 
  Trash2, 
  CheckCircle2, 
  AlertCircle, 
  Clock, 
  Box, 
  Network, 
  Globe, 
  ExternalLink,
  ChevronRight
} from 'lucide-react';

interface Props {
  namespace: string;
  onToast: (msg: string) => void;
}

interface AppResource {
  kind: string;
  name: string;
  status: 'Healthy' | 'Progressing' | 'Degraded' | 'Synced';
  info: string;
}

export const AppTopologyView: React.FC<Props> = ({ namespace, onToast }) => {
  const [selectedApp, setSelectedApp] = useState<string>('production-storefront');
  const [replicas, setReplicas] = useState<number>(3);
  const [isSyncing, setIsSyncing] = useState<boolean>(false);

  const apps = [
    {
      id: 'production-storefront',
      name: 'production-storefront',
      repo: 'https://github.com/vikukumar/tarak-sample-app',
      targetRev: 'main (a4f91e)',
      syncStatus: 'Synced',
      healthStatus: 'Healthy',
      ingressHost: 'store.vikshro.in',
      resources: [
        { kind: 'Application', name: 'production-storefront', status: 'Healthy', info: 'Git Revision: a4f91e' },
        { kind: 'Deployment', name: 'storefront-web', status: 'Healthy', info: `Desired: ${replicas} / Ready: ${replicas}` },
        { kind: 'ReplicaSet', name: 'storefront-web-6d9b7c', status: 'Healthy', info: `${replicas} Pods Active` },
        { kind: 'Pod', name: 'storefront-web-6d9b7c-7d2x1', status: 'Healthy', info: '10.244.0.12 (Node: worker-01)' },
        { kind: 'Pod', name: 'storefront-web-6d9b7c-8m4k9', status: 'Healthy', info: '10.244.0.14 (Node: worker-02)' },
        { kind: 'Pod', name: 'storefront-web-6d9b7c-1p0z3', status: 'Healthy', info: '10.244.0.19 (Node: worker-01)' },
        { kind: 'Service', name: 'storefront-svc', status: 'Healthy', info: 'ClusterIP: 10.96.12.8:80' },
        { kind: 'Ingress', name: 'storefront-cf-ingress', status: 'Healthy', info: 'https://store.vikshro.in (Cloudflare Tunnel)' }
      ]
    },
    {
      id: 'user-auth-api',
      name: 'user-auth-api',
      repo: 'https://github.com/vikukumar/tarak-auth-service',
      targetRev: 'v2.1.0',
      syncStatus: 'Synced',
      healthStatus: 'Healthy',
      ingressHost: 'auth.vikshro.in',
      resources: [
        { kind: 'Application', name: 'user-auth-api', status: 'Healthy', info: 'Git Revision: v2.1.0' },
        { kind: 'Deployment', name: 'auth-service', status: 'Healthy', info: 'Desired: 2 / Ready: 2' },
        { kind: 'ReplicaSet', name: 'auth-service-58f79', status: 'Healthy', info: '2 Pods Active' },
        { kind: 'Pod', name: 'auth-service-58f79-22a1', status: 'Healthy', info: '10.244.0.21 (Node: worker-01)' },
        { kind: 'Pod', name: 'auth-service-58f79-99b4', status: 'Healthy', info: '10.244.0.22 (Node: worker-02)' },
        { kind: 'Service', name: 'auth-svc', status: 'Healthy', info: 'ClusterIP: 10.96.44.19:50051' },
        { kind: 'Ingress', name: 'auth-tailscale-ingress', status: 'Healthy', info: 'https://auth.ts.net (Tailscale Mesh)' }
      ]
    }
  ];

  const currentApp = apps.find(a => a.id === selectedApp) || apps[0];

  const handleSync = () => {
    setIsSyncing(true);
    setTimeout(() => {
      setIsSyncing(false);
      onToast(`Application ${currentApp.name} synced with Git repository!`);
    }, 800);
  };

  const handleRestart = () => {
    onToast(`Initiated rolling restart for ${currentApp.name}`);
  };

  const handleScale = (delta: number) => {
    const next = Math.max(1, replicas + delta);
    setReplicas(next);
    onToast(`Scaled deployment ${currentApp.name} to ${next} replicas`);
  };

  const getStatusBadge = (s: string) => {
    switch (s) {
      case 'Healthy':
      case 'Synced':
        return <span className="badge badge-emerald"><CheckCircle2 size={12} /> {s}</span>;
      case 'Progressing':
        return <span className="badge badge-cyan"><Clock size={12} /> {s}</span>;
      default:
        return <span className="badge badge-purple"><AlertCircle size={12} /> {s}</span>;
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* App Header & GitOps Sync Controls */}
      <div className="glass-card" style={{ padding: '1.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.4rem' }}>
            <GitPullRequest size={22} color="var(--accent-cyan)" />
            <h2 style={{ fontSize: '1.4rem', fontWeight: 800, color: '#fff' }}>{currentApp.name}</h2>
            {getStatusBadge(currentApp.healthStatus)}
            {getStatusBadge(currentApp.syncStatus)}
          </div>
          <div style={{ display: 'flex', gap: '1rem', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
            <span>Repo: <b>{currentApp.repo}</b></span>
            <span>Revision: <code style={{ color: 'var(--accent-cyan)' }}>{currentApp.targetRev}</code></span>
          </div>
        </div>

        {/* Action Buttons (ArgoCD Toolbar) */}
        <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
          <button onClick={handleSync} className="btn-primary" style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }}>
            <RefreshCw size={14} className={isSyncing ? 'spin-icon' : ''} />
            <span>{isSyncing ? 'Syncing...' : 'Sync Git'}</span>
          </button>
          <button onClick={handleRestart} className="btn-secondary" style={{ padding: '0.5rem 0.9rem', fontSize: '0.85rem' }}>
            <RotateCcw size={14} />
            <span>Restart</span>
          </button>
          <button onClick={() => handleScale(1)} className="btn-secondary" style={{ padding: '0.5rem 0.9rem', fontSize: '0.85rem' }}>
            <Sliders size={14} />
            <span>Scale (+1)</span>
          </button>
        </div>
      </div>

      {/* Visual Dependency Graph (ArgoCD Tree Flow) */}
      <div className="glass-card" style={{ padding: '2rem' }}>
        <h3 style={{ color: '#fff', fontSize: '1.15rem', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Network size={18} color="var(--accent-cyan)" />
          <span>Application Resource Tree (ArgoCD Visual Topology)</span>
        </h3>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: '1.25rem' }}>
          {currentApp.resources.map((res, idx) => (
            <div key={idx} style={{
              background: 'rgba(15, 23, 42, 0.7)',
              border: '1px solid var(--border-glass)',
              borderRadius: 12,
              padding: '1.25rem',
              position: 'relative',
              boxShadow: '0 4px 20px rgba(0, 0, 0, 0.2)'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                <span style={{
                  background: 'rgba(0, 240, 255, 0.1)',
                  color: 'var(--accent-cyan)',
                  border: '1px solid rgba(0, 240, 255, 0.3)',
                  padding: '2px 8px',
                  borderRadius: 6,
                  fontSize: '0.75rem',
                  fontWeight: 700
                }}>
                  {res.kind}
                </span>
                {getStatusBadge(res.status)}
              </div>

              <h4 style={{ color: '#fff', fontSize: '0.98rem', marginBottom: '0.4rem', wordBreak: 'break-all' }}>{res.name}</h4>
              <p style={{ color: 'var(--text-secondary)', fontSize: '0.82rem', fontFamily: 'var(--font-mono)' }}>{res.info}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
