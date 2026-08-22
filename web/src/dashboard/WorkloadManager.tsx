import React, { useState } from 'react';
import { 
  Box, 
  Layers, 
  Server, 
  Globe, 
  Plus, 
  Trash2, 
  Edit3, 
  CheckCircle2, 
  Play, 
  Check, 
  Copy,
  Code
} from 'lucide-react';

interface Props {
  namespace: string;
  onToast: (msg: string) => void;
}

export const WorkloadManager: React.FC<Props> = ({ namespace, onToast }) => {
  const [activeKind, setActiveKind] = useState<'Pods' | 'Deployments' | 'Services' | 'Ingresses' | 'Nodes'>('Pods');
  const [showYamlModal, setShowYamlModal] = useState<boolean>(false);
  const [yamlContent, setYamlContent] = useState<string>(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample-nginx
  namespace: default
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80`);

  const [pods, setPods] = useState([
    { name: 'storefront-web-6d9b7c-7d2x1', ready: '1/1', status: 'Running', restarts: 0, age: '2h', ip: '10.244.0.12', node: 'worker-01' },
    { name: 'storefront-web-6d9b7c-8m4k9', ready: '1/1', status: 'Running', restarts: 0, age: '2h', ip: '10.244.0.14', node: 'worker-02' },
    { name: 'auth-service-58f79-22a1', ready: '1/1', status: 'Running', restarts: 0, age: '5h', ip: '10.244.0.21', node: 'worker-01' },
    { name: 'db-primary-0', ready: '1/1', status: 'Running', restarts: 0, age: '1d', ip: '10.244.0.5', node: 'master-01' }
  ]);

  const [deployments, setDeployments] = useState([
    { name: 'storefront-web', ready: '3/3', upToDate: 3, available: 3, age: '2h', image: 'nginx:alpine' },
    { name: 'auth-service', ready: '2/2', upToDate: 2, available: 2, age: '5h', image: 'golang:1.24-alpine' }
  ]);

  const handleDeletePod = (name: string) => {
    setPods(pods.filter(p => p.name !== name));
    onToast(`Pod ${name} deleted`);
  };

  const handleApplyYaml = () => {
    setShowYamlModal(false);
    onToast('Resource manifest applied successfully!');
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Action Bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap' }}>
          {(['Pods', 'Deployments', 'Services', 'Ingresses', 'Nodes'] as const).map(k => (
            <button
              key={k}
              onClick={() => setActiveKind(k)}
              style={{
                background: activeKind === k ? 'rgba(0, 240, 255, 0.15)' : 'rgba(255, 255, 255, 0.04)',
                color: activeKind === k ? 'var(--accent-cyan)' : 'var(--text-secondary)',
                border: activeKind === k ? '1px solid rgba(0, 240, 255, 0.4)' : '1px solid var(--border-glass)',
                padding: '0.45rem 0.9rem',
                borderRadius: 8,
                fontSize: '0.85rem',
                fontWeight: 600,
                cursor: 'pointer'
              }}
            >
              {k}
            </button>
          ))}
        </div>

        <button onClick={() => setShowYamlModal(true)} className="btn-primary" style={{ padding: '0.45rem 1rem', fontSize: '0.85rem' }}>
          <Plus size={16} />
          <span>Apply YAML</span>
        </button>
      </div>

      {/* Resource Table */}
      <div className="glass-card" style={{ padding: 0, overflow: 'hidden' }}>
        <div style={{ overflowX: 'auto' }}>
          {activeKind === 'Pods' && (
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', minWidth: 650 }}>
              <thead>
                <tr style={{ background: 'rgba(15, 23, 42, 0.85)', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>NAME</th>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>READY</th>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>STATUS</th>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>IP</th>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>NODE</th>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>ACTIONS</th>
                </tr>
              </thead>
              <tbody>
                {pods.map((p, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.04)' }}>
                    <td style={{ padding: '0.9rem 1.25rem', color: '#fff', fontWeight: 600, fontFamily: 'var(--font-mono)', fontSize: '0.88rem' }}>
                      {p.name}
                    </td>
                    <td style={{ padding: '0.9rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>{p.ready}</td>
                    <td style={{ padding: '0.9rem 1.25rem' }}>
                      <span className="badge badge-emerald"><CheckCircle2 size={12} /> {p.status}</span>
                    </td>
                    <td style={{ padding: '0.9rem 1.25rem', color: 'var(--accent-cyan)', fontFamily: 'var(--font-mono)', fontSize: '0.85rem' }}>{p.ip}</td>
                    <td style={{ padding: '0.9rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>{p.node}</td>
                    <td style={{ padding: '0.9rem 1.25rem' }}>
                      <button
                        onClick={() => handleDeletePod(p.name)}
                        style={{
                          background: 'rgba(244, 63, 94, 0.1)',
                          border: '1px solid rgba(244, 63, 94, 0.3)',
                          color: '#fb7185',
                          borderRadius: 6,
                          padding: '0.3rem 0.6rem',
                          cursor: 'pointer',
                          display: 'flex',
                          alignItems: 'center',
                          gap: 4,
                          fontSize: '0.78rem'
                        }}
                      >
                        <Trash2 size={12} />
                        <span>Delete</span>
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}

          {activeKind === 'Deployments' && (
            <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', minWidth: 650 }}>
              <thead>
                <tr style={{ background: 'rgba(15, 23, 42, 0.85)', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>NAME</th>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>READY</th>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>IMAGE</th>
                  <th style={{ padding: '0.85rem 1.25rem', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>AGE</th>
                </tr>
              </thead>
              <tbody>
                {deployments.map((d, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.04)' }}>
                    <td style={{ padding: '0.9rem 1.25rem', color: '#fff', fontWeight: 600, fontFamily: 'var(--font-mono)' }}>{d.name}</td>
                    <td style={{ padding: '0.9rem 1.25rem', color: 'var(--accent-cyan)' }}>{d.ready}</td>
                    <td style={{ padding: '0.9rem 1.25rem', color: 'var(--text-secondary)', fontFamily: 'var(--font-mono)' }}>{d.image}</td>
                    <td style={{ padding: '0.9rem 1.25rem', color: 'var(--text-secondary)' }}>{d.age}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* YAML Editor Modal */}
      {showYamlModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(0, 0, 0, 0.8)',
          backdropFilter: 'blur(10px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 3000,
          padding: '1rem'
        }}>
          <div className="glass-card" style={{ width: '100%', maxWidth: 700, padding: '2rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
              <h3 style={{ color: '#fff', display: 'flex', alignItems: 'center', gap: 6 }}>
                <Code size={18} color="var(--accent-cyan)" /> Apply Resource Manifest (YAML)
              </h3>
              <button onClick={() => setShowYamlModal(false)} style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontSize: '1.2rem' }}>✕</button>
            </div>

            <textarea
              value={yamlContent}
              onChange={e => setYamlContent(e.target.value)}
              rows={12}
              style={{
                width: '100%',
                background: '#040711',
                border: '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '1rem',
                color: '#38bdf8',
                fontFamily: 'var(--font-mono)',
                fontSize: '0.9rem',
                outline: 'none',
                resize: 'vertical',
                marginBottom: '1.5rem'
              }}
            />

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
              <button onClick={() => setShowYamlModal(false)} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>Cancel</button>
              <button onClick={handleApplyYaml} className="btn-primary" style={{ padding: '0.5rem 1.25rem' }}>Apply Manifest</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
