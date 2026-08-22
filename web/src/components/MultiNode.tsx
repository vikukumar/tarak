import React from 'react';
import { Network, Server, HardDrive, ShieldCheck, ArrowRight } from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

export const MultiNode: React.FC<Props> = ({ onToast }) => {
  const copy = (txt: string) => {
    navigator.clipboard.writeText(txt);
    onToast('Copied to clipboard!');
  };

  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: '2rem 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <span className="badge badge-purple" style={{ marginBottom: '1rem' }}>
          <Network size={14} /> Clustering
        </span>
        <h1 style={{ fontSize: 'clamp(2.2rem, 4vw, 3rem)', fontWeight: 800, color: '#fff' }}>
          Multi-Node <span className="text-gradient">Clustering Architecture</span>
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.1rem', marginTop: '0.5rem' }}>
          Scale across bare-metal, VPS, Raspberry Pi, and edge machines with 1 Control Plane + N Workers.
        </p>
      </div>

      {/* Visual Cluster Diagram */}
      <div className="glass-card" style={{ padding: '2.5rem', marginBottom: '2.5rem', textAlign: 'center' }}>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))',
          gap: '1.5rem',
          alignItems: 'center'
        }}>
          {/* Master Box */}
          <div style={{
            background: 'rgba(0, 240, 255, 0.05)',
            border: '1px solid rgba(0, 240, 255, 0.3)',
            borderRadius: 14,
            padding: '1.5rem'
          }}>
            <Server size={32} color="var(--accent-cyan)" style={{ marginBottom: '0.75rem' }} />
            <h4 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '0.3rem' }}>Control Plane (Master)</h4>
            <code style={{ fontSize: '0.82rem' }}>tarakd</code>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.5rem' }}>
              BoltDB State Store, Scheduler, Ingress Controller & CA
            </p>
          </div>

          <div style={{ color: 'var(--accent-purple)', fontWeight: 700, fontSize: '1.5rem' }}>
            ⟵ mTLS Heartbeat ⟶
          </div>

          {/* Workers Box */}
          <div style={{
            background: 'rgba(168, 85, 247, 0.05)',
            border: '1px solid rgba(168, 85, 247, 0.3)',
            borderRadius: 14,
            padding: '1.5rem'
          }}>
            <HardDrive size={32} color="var(--accent-purple)" style={{ marginBottom: '0.75rem' }} />
            <h4 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '0.3rem' }}>Worker Nodes (N)</h4>
            <code style={{ fontSize: '0.82rem' }}>taraks</code>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.82rem', marginTop: '0.5rem' }}>
              Native TCR Runtime, Process Isolation & Pod Sandboxing
            </p>
          </div>
        </div>
      </div>

      {/* Setup Steps */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
        <div className="glass-card" style={{ padding: '2rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.25rem', marginBottom: '0.75rem' }}>
            1. Start Master Server (<code>tarakd</code>)
          </h3>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
            Run the dedicated control plane on your master host:
          </p>
          <div className="code-box" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>tarakd --listen-addr 0.0.0.0:8443 --token my-cluster-secret-token</span>
            <button onClick={() => copy('tarakd --listen-addr 0.0.0.0:8443 --token my-cluster-secret-token')} style={{ background: 'transparent', border: 'none', color: 'var(--accent-cyan)', cursor: 'pointer', fontWeight: 600 }}>Copy</button>
          </div>
        </div>

        <div className="glass-card" style={{ padding: '2rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.25rem', marginBottom: '0.75rem' }}>
            2. Join Worker Nodes (<code>taraks</code>)
          </h3>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
            On every worker node, point the daemon to the control plane:
          </p>
          <div className="code-box" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>taraks --server https://192.168.1.50:8443 --token my-cluster-secret-token --node-name worker-01</span>
            <button onClick={() => copy('taraks --server https://192.168.1.50:8443 --token my-cluster-secret-token --node-name worker-01')} style={{ background: 'transparent', border: 'none', color: 'var(--accent-cyan)', cursor: 'pointer', fontWeight: 600 }}>Copy</button>
          </div>
        </div>

        <div className="glass-card" style={{ padding: '2rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.25rem', marginBottom: '0.75rem' }}>
            3. Verify Nodes
          </h3>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
            Inspect cluster health and node heartbeats:
          </p>
          <pre className="code-box">{`tarakctl get nodes

# Output:
# NAME        STATUS   ROLES          AGE   VERSION
# master-01   Ready    controlplane   15m   v1.0.6
# worker-01   Ready    worker         3m    v1.0.6
# worker-02   Ready    worker         1m    v1.0.6`}</pre>
        </div>
      </div>
    </div>
  );
};
