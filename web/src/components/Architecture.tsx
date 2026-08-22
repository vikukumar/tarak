import React from 'react';
import { Cpu, Database, Shield, Layers, Zap } from 'lucide-react';

export const Architecture: React.FC = () => {
  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: '2rem 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <span className="badge badge-emerald" style={{ marginBottom: '1rem' }}>
          <Cpu size={14} /> System Internals
        </span>
        <h1 style={{ fontSize: 'clamp(2.2rem, 4vw, 3rem)', fontWeight: 800, color: '#fff' }}>
          Internal <span className="text-gradient">Architecture Deep Dive</span>
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.1rem', marginTop: '0.5rem' }}>
          Engineered in pure Go with zero external daemons for maximum speed, security, and portability.
        </p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.5rem', marginBottom: '2.5rem' }}>
        <div className="glass-card" style={{ padding: '2rem' }}>
          <Zap size={28} color="var(--accent-cyan)" style={{ marginBottom: '1rem' }} />
          <h3 style={{ color: '#fff', fontSize: '1.2rem', marginBottom: '0.5rem' }}>TCR Container Runtime</h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', lineHeight: 1.7 }}>
            Tarak Container Runtime (TCR) directly leverages OS system calls (Linux namespaces/cgroups, macOS sandbox, Windows process isolation) to spawn and manage pods in sub-milliseconds without containerd or Docker.
          </p>
        </div>

        <div className="glass-card" style={{ padding: '2rem' }}>
          <Database size={28} color="var(--accent-purple)" style={{ marginBottom: '1rem' }} />
          <h3 style={{ color: '#fff', fontSize: '1.2rem', marginBottom: '0.5rem' }}>Embedded BoltDB Storage</h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', lineHeight: 1.7 }}>
            Replaces heavyweight etcd clusters with an embedded, transactional B+ tree engine that guarantees ACID semantics, monotonic revision counter, and sub-microsecond state lookups.
          </p>
        </div>

        <div className="glass-card" style={{ padding: '2rem' }}>
          <Shield size={28} color="var(--accent-emerald)" style={{ marginBottom: '1rem' }} />
          <h3 style={{ color: '#fff', fontSize: '1.2rem', marginBottom: '0.5rem' }}>Zero-Trust mTLS PKI</h3>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', lineHeight: 1.7 }}>
            Automated internal Certificate Authority generates Mutual TLS (mTLS) certificates on the fly, Ed25519 cluster tokens, and AES-256-GCM encryption for stored secrets.
          </p>
        </div>
      </div>
    </div>
  );
};
