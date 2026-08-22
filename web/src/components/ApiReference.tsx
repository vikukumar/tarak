import React, { useState } from 'react';
import { Code2, Search } from 'lucide-react';

export const ApiReference: React.FC = () => {
  const [search, setSearch] = useState('');
  const [selectedMethod, setSelectedMethod] = useState<string>('ALL');

  const endpoints = [
    { method: 'GET', path: '/api/v1/namespaces/{ns}/pods', desc: 'List or watch pods within a namespace' },
    { method: 'POST', path: '/api/v1/namespaces/{ns}/pods', desc: 'Create and schedule a new pod' },
    { method: 'GET', path: '/api/v1/namespaces/{ns}/pods/{name}', desc: 'Get specific pod specification and status' },
    { method: 'DELETE', path: '/api/v1/namespaces/{ns}/pods/{name}', desc: 'Delete a pod with graceful teardown' },
    { method: 'GET', path: '/api/v1/nodes', desc: 'List all registered control plane and worker nodes' },
    { method: 'POST', path: '/api/v1/nodes', desc: 'Register a new worker node with the control plane' },
    { method: 'GET', path: '/apis/apps/v1/namespaces/{ns}/deployments', desc: 'List all deployments' },
    { method: 'POST', path: '/apis/apps/v1/namespaces/{ns}/deployments', desc: 'Create a deployment replica controller' },
    { method: 'GET', path: '/apis/networking.k8s.io/v1/ingresses', desc: 'List active ingress rules and routing table' },
    { method: 'GET', path: '/apis/networking.tarak.io/v1/tunnels', desc: 'Inspect live Cloudflare & Tailscale tunnel status' },
    { method: 'GET', path: '/healthz', desc: 'Liveness probe endpoint' },
    { method: 'GET', path: '/livez', desc: 'Readiness probe endpoint' }
  ];

  const filtered = endpoints.filter(ep => {
    const matchMethod = selectedMethod === 'ALL' || ep.method === selectedMethod;
    const matchSearch = ep.path.toLowerCase().includes(search.toLowerCase()) || ep.desc.toLowerCase().includes(search.toLowerCase());
    return matchMethod && matchSearch;
  });

  const getMethodBadge = (m: string) => {
    switch (m) {
      case 'GET': return { bg: 'rgba(59, 130, 246, 0.15)', color: '#60a5fa', border: 'rgba(59, 130, 246, 0.4)' };
      case 'POST': return { bg: 'rgba(16, 185, 129, 0.15)', color: '#34d399', border: 'rgba(16, 185, 129, 0.4)' };
      case 'PUT': return { bg: 'rgba(245, 158, 11, 0.15)', color: '#fbbf24', border: 'rgba(245, 158, 11, 0.4)' };
      case 'DELETE': return { bg: 'rgba(244, 63, 94, 0.15)', color: '#fb7185', border: 'rgba(244, 63, 94, 0.4)' };
      default: return { bg: 'rgba(255, 255, 255, 0.1)', color: '#fff', border: 'rgba(255, 255, 255, 0.2)' };
    }
  };

  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: '2rem 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <span className="badge badge-cyan" style={{ marginBottom: '1rem' }}>
          <Code2 size={14} /> REST API Reference
        </span>
        <h1 style={{ fontSize: 'clamp(2.2rem, 4vw, 3rem)', fontWeight: 800, color: '#fff' }}>
          Kubernetes-Compatible <span className="text-gradient">REST Endpoints</span>
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.1rem', marginTop: '0.5rem' }}>
          Discover and interact directly with Tarak's HTTP/JSON APIs.
        </p>
      </div>

      {/* Search & Filter Bar */}
      <div className="glass-card" style={{ padding: '1rem 1.5rem', marginBottom: '2rem', display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flex: 1, minWidth: 200 }}>
          <Search size={18} color="var(--text-muted)" />
          <input
            type="text"
            placeholder="Search API endpoints (e.g. /pods, /tunnels)..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            style={{
              background: 'transparent',
              border: 'none',
              color: '#fff',
              outline: 'none',
              width: '100%',
              fontSize: '0.95rem'
            }}
          />
        </div>

        <div style={{ display: 'flex', gap: '0.4rem' }}>
          {['ALL', 'GET', 'POST', 'DELETE'].map(m => (
            <button
              key={m}
              onClick={() => setSelectedMethod(m)}
              style={{
                background: selectedMethod === m ? 'rgba(0, 240, 255, 0.15)' : 'transparent',
                color: selectedMethod === m ? 'var(--accent-cyan)' : 'var(--text-muted)',
                border: selectedMethod === m ? '1px solid rgba(0, 240, 255, 0.3)' : '1px solid transparent',
                borderRadius: 6,
                padding: '0.3rem 0.65rem',
                fontSize: '0.8rem',
                fontWeight: 600,
                cursor: 'pointer'
              }}
            >
              {m}
            </button>
          ))}
        </div>
      </div>

      {/* Endpoints List */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        {filtered.map((ep, idx) => {
          const badgeStyle = getMethodBadge(ep.method);
          return (
            <div key={idx} className="glass-card" style={{
              padding: '1.2rem 1.5rem',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              flexWrap: 'wrap',
              gap: '0.75rem'
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
                <span style={{
                  background: badgeStyle.bg,
                  color: badgeStyle.color,
                  border: `1px solid ${badgeStyle.border}`,
                  padding: '0.25rem 0.65rem',
                  borderRadius: 6,
                  fontFamily: 'var(--font-mono)',
                  fontWeight: 700,
                  fontSize: '0.8rem'
                }}>
                  {ep.method}
                </span>
                <code style={{ fontSize: '0.95rem', color: '#f8fafc' }}>{ep.path}</code>
              </div>
              <span style={{ color: 'var(--text-secondary)', fontSize: '0.88rem' }}>{ep.desc}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
