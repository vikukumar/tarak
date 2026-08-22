import React, { useState } from 'react';
import { Activity, ShieldCheck, AlertTriangle, ArrowRight, Filter, Zap, Globe } from 'lucide-react';

interface Props {
  namespace: string;
}

export const HubbleVisualizer: React.FC<Props> = ({ namespace }) => {
  const [protocolFilter, setProtocolFilter] = useState<string>('ALL');

  const flows = [
    {
      id: 'flow-1',
      time: 'Just now',
      src: 'storefront-ingress',
      srcNs: 'default',
      dst: 'storefront-web',
      dstNs: 'default',
      proto: 'HTTP',
      port: 80,
      verdict: 'FORWARDED',
      code: 200,
      latency: '0.8ms',
      summary: 'GET /api/v1/products (200 OK)'
    },
    {
      id: 'flow-2',
      time: '1s ago',
      src: 'storefront-web',
      srcNs: 'default',
      dst: 'auth-service',
      dstNs: 'default',
      proto: 'TCP',
      port: 50051,
      verdict: 'FORWARDED',
      code: 200,
      latency: '1.2ms',
      summary: 'gRPC SessionValidation'
    },
    {
      id: 'flow-3',
      time: '2s ago',
      src: 'auth-service',
      srcNs: 'default',
      dst: 'db-primary-0',
      dstNs: 'tarak-system',
      proto: 'TCP',
      port: 5432,
      verdict: 'FORWARDED',
      code: 200,
      latency: '0.4ms',
      summary: 'PostgreSQL TLS Connection'
    },
    {
      id: 'flow-4',
      time: '4s ago',
      src: 'unknown-crawler',
      srcNs: 'tarak-public',
      dst: 'db-primary-0',
      dstNs: 'tarak-system',
      proto: 'TCP',
      port: 5432,
      verdict: 'DROPPED',
      code: 403,
      latency: '0.1ms',
      summary: 'Zero-Trust NetworkPolicy Blocked'
    }
  ];

  const filtered = flows.filter(f => protocolFilter === 'ALL' || f.proto === protocolFilter);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Metrics Banner */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem' }}>
        <div className="glass-card" style={{ padding: '1.25rem' }}>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase' }}>Traffic Throughput</span>
          <h3 style={{ color: 'var(--accent-cyan)', fontSize: '1.6rem', marginTop: '0.2rem' }}>4.8k req/s</h3>
        </div>
        <div className="glass-card" style={{ padding: '1.25rem' }}>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase' }}>P95 Latency</span>
          <h3 style={{ color: 'var(--accent-emerald)', fontSize: '1.6rem', marginTop: '0.2rem' }}>0.94 ms</h3>
        </div>
        <div className="glass-card" style={{ padding: '1.25rem' }}>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase' }}>Zero-Trust Dropped</span>
          <h3 style={{ color: '#fb7185', fontSize: '1.6rem', marginTop: '0.2rem' }}>14 pkts</h3>
        </div>
      </div>

      {/* Network Topology Visual Flow Chart */}
      <div className="glass-card" style={{ padding: '2rem', textAlign: 'center' }}>
        <h3 style={{ color: '#fff', fontSize: '1.15rem', marginBottom: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 6 }}>
          <Activity size={18} color="var(--accent-cyan)" /> Live Hubble Service-to-Service Graph
        </h3>

        <div style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '1.5rem'
        }}>
          {/* Node Ingress */}
          <div style={{ background: 'rgba(0, 240, 255, 0.08)', border: '1px solid var(--accent-cyan)', borderRadius: 10, padding: '1rem 1.5rem' }}>
            <Globe size={24} color="var(--accent-cyan)" style={{ marginBottom: 4 }} />
            <h5 style={{ color: '#fff', margin: 0 }}>Cloudflare Ingress</h5>
            <small style={{ color: 'var(--text-muted)' }}>store.vikshro.in</small>
          </div>

          <div style={{ color: 'var(--accent-cyan)', fontWeight: 700 }}>⟶ 0.8ms ⟶</div>

          {/* Node Web App */}
          <div style={{ background: 'rgba(168, 85, 247, 0.08)', border: '1px solid var(--accent-purple)', borderRadius: 10, padding: '1rem 1.5rem' }}>
            <Zap size={24} color="var(--accent-purple)" style={{ marginBottom: 4 }} />
            <h5 style={{ color: '#fff', margin: 0 }}>storefront-web</h5>
            <small style={{ color: 'var(--text-muted)' }}>3 Replicas</small>
          </div>

          <div style={{ color: 'var(--accent-purple)', fontWeight: 700 }}>⟶ 1.2ms ⟶</div>

          {/* Node Auth Service */}
          <div style={{ background: 'rgba(16, 185, 129, 0.08)', border: '1px solid var(--accent-emerald)', borderRadius: 10, padding: '1rem 1.5rem' }}>
            <ShieldCheck size={24} color="var(--accent-emerald)" style={{ marginBottom: 4 }} />
            <h5 style={{ color: '#fff', margin: 0 }}>auth-service</h5>
            <small style={{ color: 'var(--text-muted)' }}>Tailscale Mesh</small>
          </div>
        </div>
      </div>

      {/* Real-time Flows Table */}
      <div className="glass-card" style={{ padding: 0, overflow: 'hidden' }}>
        <div style={{ padding: '0.9rem 1.25rem', background: 'rgba(15, 23, 42, 0.8)', borderBottom: '1px solid var(--border-glass)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.75rem' }}>
          <h4 style={{ color: '#fff', fontSize: '0.95rem', margin: 0 }}>Recent Network Flow Events</h4>
          <div style={{ display: 'flex', gap: '0.35rem' }}>
            {['ALL', 'HTTP', 'TCP', 'UDP'].map(p => (
              <button
                key={p}
                onClick={() => setProtocolFilter(p)}
                style={{
                  background: protocolFilter === p ? 'rgba(0, 240, 255, 0.15)' : 'transparent',
                  color: protocolFilter === p ? 'var(--accent-cyan)' : 'var(--text-muted)',
                  border: protocolFilter === p ? '1px solid rgba(0, 240, 255, 0.4)' : '1px solid transparent',
                  borderRadius: 4,
                  padding: '2px 8px',
                  fontSize: '0.75rem',
                  fontWeight: 600,
                  cursor: 'pointer'
                }}
              >
                {p}
              </button>
            ))}
          </div>
        </div>

        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', minWidth: 700 }}>
            <thead>
              <tr style={{ background: 'rgba(10, 15, 30, 0.9)', borderBottom: '1px solid var(--border-glass)' }}>
                <th style={{ padding: '0.75rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.78rem' }}>TIME</th>
                <th style={{ padding: '0.75rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.78rem' }}>SOURCE</th>
                <th style={{ padding: '0.75rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.78rem' }}>DESTINATION</th>
                <th style={{ padding: '0.75rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.78rem' }}>PROTO</th>
                <th style={{ padding: '0.75rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.78rem' }}>VERDICT</th>
                <th style={{ padding: '0.75rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.78rem' }}>LATENCY</th>
                <th style={{ padding: '0.75rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.78rem' }}>SUMMARY</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((f, idx) => (
                <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.04)' }}>
                  <td style={{ padding: '0.85rem 1.25rem', color: 'var(--text-muted)', fontSize: '0.82rem' }}>{f.time}</td>
                  <td style={{ padding: '0.85rem 1.25rem', color: '#fff', fontFamily: 'var(--font-mono)', fontSize: '0.85rem' }}>{f.src}</td>
                  <td style={{ padding: '0.85rem 1.25rem', color: 'var(--accent-cyan)', fontFamily: 'var(--font-mono)', fontSize: '0.85rem' }}>{f.dst}:{f.port}</td>
                  <td style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.82rem' }}>{f.proto}</td>
                  <td style={{ padding: '0.85rem 1.25rem' }}>
                    {f.verdict === 'FORWARDED' ? (
                      <span className="badge badge-emerald">Forwarded</span>
                    ) : (
                      <span className="badge badge-purple" style={{ background: 'rgba(244, 63, 94, 0.15)', color: '#fb7185', border: '1px solid rgba(244, 63, 94, 0.4)' }}>
                        Dropped
                      </span>
                    )}
                  </td>
                  <td style={{ padding: '0.85rem 1.25rem', color: 'var(--accent-emerald)', fontSize: '0.82rem', fontFamily: 'var(--font-mono)' }}>{f.latency}</td>
                  <td style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.82rem' }}>{f.summary}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
