import React, { useState, useEffect } from 'react';
import { Activity, ShieldCheck, AlertTriangle, ArrowRight, Filter, Zap, Globe, RefreshCw } from 'lucide-react';

interface Props {
  namespace: string;
}

interface NetworkFlow {
  id: string;
  timestamp: string;
  srcPod: string;
  srcNS: string;
  srcIP: string;
  dstPod: string;
  dstNS: string;
  dstIP: string;
  dstPort: number;
  protocol: string;
  verdict: string;
  statusCode: number;
  latencyMs: number;
  bytes: number;
  summary: string;
}

export const HubbleVisualizer: React.FC<Props> = ({ namespace }) => {
  const [protocolFilter, setProtocolFilter] = useState<string>('ALL');
  const [flows, setFlows] = useState<NetworkFlow[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const fetchFlows = async () => {
    setIsLoading(true);
    try {
      const res = await fetch('/apis/telemetry.tarak.io/v1/flows');
      if (res.ok) {
        const data = await res.json();
        setFlows(data.items || []);
      }
    } catch {
      // Fallback gracefully
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchFlows();
    const interval = setInterval(fetchFlows, 3000);
    return () => clearInterval(interval);
  }, []);

  const filteredFlows = flows.filter(f => {
    if (protocolFilter !== 'ALL' && f.protocol !== protocolFilter) return false;
    return true;
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Hubble Flow Top Stats */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1rem' }}>
        <div className="glass-card" style={{ padding: '1.25rem' }}>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem', marginBottom: '0.35rem' }}>TOTAL NETWORK FLOWS</div>
          <div style={{ fontSize: '1.8rem', fontWeight: 800, color: 'var(--accent-cyan)' }}>{flows.length}</div>
          <div style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>Live telemetry buffer</div>
        </div>

        <div className="glass-card" style={{ padding: '1.25rem' }}>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem', marginBottom: '0.35rem' }}>FORWARDED PACKETS</div>
          <div style={{ fontSize: '1.8rem', fontWeight: 800, color: 'var(--accent-green)' }}>
            {flows.filter(f => f.verdict === 'FORWARDED').length}
          </div>
          <div style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>mTLS authenticated</div>
        </div>

        <div className="glass-card" style={{ padding: '1.25rem' }}>
          <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem', marginBottom: '0.35rem' }}>ZERO-TRUST BLOCKED</div>
          <div style={{ fontSize: '1.8rem', fontWeight: 800, color: 'var(--accent-pink)' }}>
            {flows.filter(f => f.verdict === 'DROPPED').length}
          </div>
          <div style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>Unauthorized lateral attempts</div>
        </div>
      </div>

      {/* Realtime Flows Table */}
      <div className="glass-card" style={{ padding: '1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '1rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <Activity size={20} color="var(--accent-cyan)" />
            <h3 style={{ color: '#fff', fontSize: '1.1rem' }}>Hubble Realtime Traffic Flows</h3>
            <span style={{
              background: 'rgba(57, 255, 20, 0.15)',
              color: 'var(--accent-green)',
              padding: '2px 8px',
              borderRadius: 4,
              fontSize: '0.75rem',
              fontWeight: 700
            }}>
              POLLING LIVE (3s)
            </span>
          </div>

          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <button onClick={fetchFlows} className="btn-secondary" style={{ padding: '0.45rem 0.75rem', fontSize: '0.85rem' }}>
              <RefreshCw size={14} className={isLoading ? 'spin' : ''} />
            </button>
            <div style={{ display: 'flex', gap: '0.3rem' }}>
              {['ALL', 'HTTP', 'TCP'].map(p => (
                <button
                  key={p}
                  onClick={() => setProtocolFilter(p)}
                  style={{
                    background: protocolFilter === p ? 'rgba(0, 240, 255, 0.15)' : 'transparent',
                    color: protocolFilter === p ? 'var(--accent-cyan)' : 'var(--text-muted)',
                    border: protocolFilter === p ? '1px solid var(--accent-cyan)' : '1px solid var(--border-glass)',
                    borderRadius: 6,
                    padding: '0.35rem 0.75rem',
                    fontSize: '0.8rem',
                    cursor: 'pointer'
                  }}
                >
                  {p}
                </button>
              ))}
            </div>
          </div>
        </div>

        {filteredFlows.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
            <Activity size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
            <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>No network flows recorded yet</p>
            <p style={{ fontSize: '0.82rem', marginTop: '0.25rem' }}>Inbound API requests and workload network packets will automatically appear here in realtime.</p>
          </div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table className="data-table" style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ color: 'var(--text-muted)', fontSize: '0.8rem', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.75rem' }}>TIME</th>
                  <th style={{ padding: '0.75rem' }}>SOURCE</th>
                  <th style={{ padding: '0.75rem' }}>DESTINATION</th>
                  <th style={{ padding: '0.75rem' }}>PROTO / PORT</th>
                  <th style={{ padding: '0.75rem' }}>VERDICT</th>
                  <th style={{ padding: '0.75rem' }}>SUMMARY</th>
                </tr>
              </thead>
              <tbody>
                {filteredFlows.map((flow, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)', fontSize: '0.88rem' }}>
                    <td style={{ padding: '0.75rem', color: 'var(--text-muted)', fontSize: '0.8rem' }}>
                      {new Date(flow.timestamp).toLocaleTimeString()}
                    </td>
                    <td style={{ padding: '0.75rem' }}>
                      <span style={{ color: 'var(--accent-cyan)', fontWeight: 600 }}>{flow.srcPod || flow.srcIP}</span>
                      <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem', marginLeft: 4 }}>({flow.srcNS})</span>
                    </td>
                    <td style={{ padding: '0.75rem' }}>
                      <span style={{ color: '#fff', fontWeight: 600 }}>{flow.dstPod || flow.dstIP}</span>
                      <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem', marginLeft: 4 }}>({flow.dstNS})</span>
                    </td>
                    <td style={{ padding: '0.75rem', color: 'var(--accent-purple)' }}>{flow.protocol} / {flow.dstPort}</td>
                    <td style={{ padding: '0.75rem' }}>
                      <span style={{
                        background: flow.verdict === 'FORWARDED' ? 'rgba(57, 255, 20, 0.15)' : 'rgba(255, 0, 85, 0.15)',
                        color: flow.verdict === 'FORWARDED' ? 'var(--accent-green)' : 'var(--accent-pink)',
                        border: `1px solid ${flow.verdict === 'FORWARDED' ? 'rgba(57, 255, 20, 0.3)' : 'rgba(255, 0, 85, 0.3)'}`,
                        padding: '2px 8px',
                        borderRadius: 4,
                        fontSize: '0.75rem',
                        fontWeight: 700
                      }}>
                        {flow.verdict}
                      </span>
                    </td>
                    <td style={{ padding: '0.75rem', fontFamily: 'var(--font-mono)', fontSize: '0.82rem', color: 'var(--text-secondary)' }}>
                      {flow.summary} ({flow.latencyMs.toFixed(1)}ms)
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};
