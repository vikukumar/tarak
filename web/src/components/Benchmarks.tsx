import React from 'react';
import { Gauge, Cpu, CheckCircle2, XCircle } from 'lucide-react';

export const Benchmarks: React.FC = () => {
  return (
    <section style={{ maxWidth: 1000, margin: '4rem auto 0 auto', padding: '0 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '2.5rem' }}>
        <div style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: '0.4rem',
          color: 'var(--accent-cyan)',
          fontSize: '0.85rem',
          fontWeight: 700,
          textTransform: 'uppercase',
          letterSpacing: '1px',
          marginBottom: '0.5rem'
        }}>
          <Gauge size={16} />
          <span>Performance & Footprint</span>
        </div>
        <h2 style={{ fontSize: 'clamp(1.8rem, 3.5vw, 2.6rem)', fontWeight: 800, color: '#fff' }}>
          TARAK vs <span className="text-gradient">K3s & Kubernetes</span>
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1rem', marginTop: '0.5rem' }}>
          Tested under identical conditions on 4-Core x86_64, 8GB RAM Linux & Windows hosts:
        </p>
      </div>

      <div className="glass-card" style={{ padding: '0', overflow: 'hidden' }}>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', minWidth: 620 }}>
            <thead>
              <tr style={{ background: 'rgba(15, 23, 42, 0.9)', borderBottom: '1px solid var(--border-glass)' }}>
                <th style={{ padding: '1rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.82rem', textTransform: 'uppercase' }}>Metric</th>
                <th style={{ padding: '1rem 1.25rem', color: 'var(--accent-cyan)', fontSize: '0.88rem', fontWeight: 700 }}>⚡ TARAK</th>
                <th style={{ padding: '1rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.82rem', textTransform: 'uppercase' }}>K3s</th>
                <th style={{ padding: '1rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.82rem', textTransform: 'uppercase' }}>Kubernetes (K8s)</th>
              </tr>
            </thead>
            <tbody>
              <tr style={{ background: 'rgba(0, 240, 255, 0.03)', borderBottom: '1px solid rgba(255, 255, 255, 0.05)' }}>
                <td style={{ padding: '1.1rem 1.25rem', fontWeight: 600, color: '#fff' }}>Idle RAM Usage</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--accent-cyan)', fontWeight: 700 }}>~22 MB</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>~600 MB</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>~1.8 GB</td>
              </tr>
              <tr style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)' }}>
                <td style={{ padding: '1.1rem 1.25rem', fontWeight: 600, color: '#fff' }}>Binary Size</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--accent-cyan)', fontWeight: 700 }}>~18 MB (Self-Contained)</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>~75 MB</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>~400 MB+</td>
              </tr>
              <tr style={{ background: 'rgba(0, 240, 255, 0.03)', borderBottom: '1px solid rgba(255, 255, 255, 0.05)' }}>
                <td style={{ padding: '1.1rem 1.25rem', fontWeight: 600, color: '#fff' }}>Cold Boot Time</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--accent-cyan)', fontWeight: 700 }}>&lt; 180 ms</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>~15 - 30 seconds</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>~60 - 120 seconds</td>
              </tr>
              <tr style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)' }}>
                <td style={{ padding: '1.1rem 1.25rem', fontWeight: 600, color: '#fff' }}>External Daemons Needed</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--accent-emerald)', fontWeight: 700, display: 'flex', alignItems: 'center', gap: 6 }}>
                  <CheckCircle2 size={16} /> None (Native TCR)
                </td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>containerd, runc</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>etcd, containerd, runc, CNI</td>
              </tr>
              <tr style={{ background: 'rgba(0, 240, 255, 0.03)', borderBottom: '1px solid rgba(255, 255, 255, 0.05)' }}>
                <td style={{ padding: '1.1rem 1.25rem', fontWeight: 600, color: '#fff' }}>Inbuilt Cloudflare & Tailscale</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--accent-emerald)', fontWeight: 700, display: 'flex', alignItems: 'center', gap: 6 }}>
                  <CheckCircle2 size={16} /> Inbuilt Native
                </td>
                <td style={{ padding: '1.1rem 1.25rem', color: '#f87171', display: 'flex', alignItems: 'center', gap: 6 }}>
                  <XCircle size={16} /> Extra plugins / helm
                </td>
                <td style={{ padding: '1.1rem 1.25rem', color: '#f87171', display: 'flex', alignItems: 'center', gap: 6 }}>
                  <XCircle size={16} /> Complex ingress setup
                </td>
              </tr>
              <tr>
                <td style={{ padding: '1.1rem 1.25rem', fontWeight: 600, color: '#fff' }}>Platform OS Support</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--accent-cyan)', fontWeight: 700 }}>
                  Linux, macOS, Windows (AMD64 & ARM64)
                </td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>Linux Only (WSL for Windows)</td>
                <td style={{ padding: '1.1rem 1.25rem', color: 'var(--text-secondary)' }}>Linux Only (WSL for Windows)</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
};
