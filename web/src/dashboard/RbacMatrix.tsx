import React, { useState } from 'react';
import { ShieldCheck, UserCheck, Lock, Check, X, Plus, Users, Key } from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

export const RbacMatrix: React.FC<Props> = ({ onToast }) => {
  const [selectedRole, setSelectedRole] = useState<string>('developer');
  const [simulatorUser, setSimulatorUser] = useState<string>('alice');
  const [simulatorVerb, setSimulatorVerb] = useState<string>('create');
  const [simulatorResource, setSimulatorResource] = useState<string>('pods');
  const [testResult, setTestResult] = useState<{ allowed: boolean; reason: string } | null>(null);

  const roles = [
    { name: 'cluster-admin', desc: 'Full superuser root access across all API groups and namespaces', users: ['admin', 'tarak-admin'] },
    { name: 'developer', desc: 'Read and write access to Pods, Deployments, Services, and Ingresses in default namespace', users: ['alice', 'bob'] },
    { name: 'auditor', desc: 'Read-only access to cluster audit logs, security policies, and telemetry', users: ['security-bot'] },
    { name: 'viewer', desc: 'Read-only view access across non-sensitive core namespaces', users: ['guest'] }
  ];

  const resources = ['pods', 'deployments', 'services', 'ingresses', 'secrets', 'nodes', 'tunnels'];
  const verbs = ['get', 'list', 'watch', 'create', 'update', 'delete', 'exec'];

  // Matrix permission lookup function
  const hasPermission = (role: string, res: string, verb: string): boolean => {
    if (role === 'cluster-admin') return true;
    if (role === 'viewer') return ['get', 'list', 'watch'].includes(verb) && res !== 'secrets';
    if (role === 'auditor') return ['get', 'list'].includes(verb);
    if (role === 'developer') {
      if (res === 'secrets') return ['get'].includes(verb);
      if (res === 'nodes') return ['get', 'list'].includes(verb);
      return true;
    }
    return false;
  };

  const handleTestPermission = () => {
    const isAllowed = hasPermission(selectedRole, simulatorResource, simulatorVerb);
    setTestResult({
      allowed: isAllowed,
      reason: isAllowed 
        ? `User '${simulatorUser}' with Role '${selectedRole}' is AUTHORIZED for verb '${simulatorVerb}' on '${simulatorResource}'.`
        : `Access DENIED: Role '${selectedRole}' does not grant '${simulatorVerb}' permission on '${simulatorResource}'.`
    });
    onToast(isAllowed ? 'RBAC Check Passed!' : 'RBAC Check: Denied');
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Roles Header */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1rem' }}>
        {roles.map(r => (
          <div
            key={r.name}
            onClick={() => setSelectedRole(r.name)}
            className="glass-card"
            style={{
              padding: '1.25rem',
              cursor: 'pointer',
              borderColor: selectedRole === r.name ? 'var(--accent-cyan)' : 'var(--border-glass)',
              background: selectedRole === r.name ? 'rgba(0, 240, 255, 0.08)' : 'rgba(15, 23, 42, 0.6)'
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.4rem' }}>
              <h4 style={{ color: '#fff', margin: 0, fontSize: '1rem' }}>{r.name}</h4>
              <ShieldCheck size={16} color={selectedRole === r.name ? 'var(--accent-cyan)' : 'var(--text-muted)'} />
            </div>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.6rem' }}>{r.desc}</p>
            <div style={{ display: 'flex', gap: '0.3rem', flexWrap: 'wrap' }}>
              {r.users.map(u => (
                <span key={u} style={{ background: 'rgba(255, 255, 255, 0.08)', borderRadius: 4, padding: '1px 6px', fontSize: '0.72rem', color: '#94a3b8' }}>
                  @{u}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Permission Matrix Grid */}
      <div className="glass-card" style={{ padding: 0, overflow: 'hidden' }}>
        <div style={{ padding: '1rem 1.5rem', background: 'rgba(15, 23, 42, 0.8)', borderBottom: '1px solid var(--border-glass)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h3 style={{ color: '#fff', fontSize: '1.05rem', margin: 0, display: 'flex', alignItems: 'center', gap: 6 }}>
            <Lock size={16} color="var(--accent-cyan)" /> Permission Matrix for <code>{selectedRole}</code>
          </h3>
          <span className="badge badge-cyan">RBAC Active</span>
        </div>

        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'center', minWidth: 600 }}>
            <thead>
              <tr style={{ background: 'rgba(10, 15, 30, 0.9)', borderBottom: '1px solid var(--border-glass)' }}>
                <th style={{ padding: '0.8rem 1.25rem', textAlign: 'left', color: 'var(--text-secondary)', fontSize: '0.8rem' }}>RESOURCE</th>
                {verbs.map(v => (
                  <th key={v} style={{ padding: '0.8rem 0.6rem', color: 'var(--text-secondary)', fontSize: '0.78rem', textTransform: 'uppercase' }}>
                    {v}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {resources.map(res => (
                <tr key={res} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.04)' }}>
                  <td style={{ padding: '0.85rem 1.25rem', textAlign: 'left', color: '#fff', fontWeight: 600, fontFamily: 'var(--font-mono)', fontSize: '0.85rem' }}>
                    {res}
                  </td>
                  {verbs.map(v => {
                    const ok = hasPermission(selectedRole, res, v);
                    return (
                      <td key={v} style={{ padding: '0.85rem 0.6rem' }}>
                        {ok ? (
                          <span style={{ color: 'var(--accent-emerald)', display: 'inline-flex', alignItems: 'center' }}>
                            <Check size={16} />
                          </span>
                        ) : (
                          <span style={{ color: 'rgba(255, 255, 255, 0.2)', display: 'inline-flex', alignItems: 'center' }}>
                            <X size={16} />
                          </span>
                        )}
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Permission Tester Simulator */}
      <div className="glass-card" style={{ padding: '1.5rem' }}>
        <h4 style={{ color: '#fff', fontSize: '1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
          <Key size={16} color="var(--accent-purple)" /> RBAC Policy Access Simulator (SelfSubjectAccessReview)
        </h4>

        <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'center', marginBottom: '1rem' }}>
          <input
            type="text"
            value={simulatorUser}
            onChange={e => setSimulatorUser(e.target.value)}
            placeholder="User (e.g. alice)"
            style={{
              background: 'rgba(15, 23, 42, 0.8)',
              border: '1px solid var(--border-glass)',
              color: '#fff',
              padding: '0.4rem 0.75rem',
              borderRadius: 6,
              fontSize: '0.85rem'
            }}
          />

          <span style={{ color: 'var(--text-muted)' }}>can</span>

          <select
            value={simulatorVerb}
            onChange={e => setSimulatorVerb(e.target.value)}
            style={{
              background: 'rgba(15, 23, 42, 0.8)',
              border: '1px solid var(--border-glass)',
              color: '#fff',
              padding: '0.4rem 0.75rem',
              borderRadius: 6,
              fontSize: '0.85rem'
            }}
          >
            {verbs.map(v => <option key={v} value={v}>{v}</option>)}
          </select>

          <span style={{ color: 'var(--text-muted)' }}>resource</span>

          <select
            value={simulatorResource}
            onChange={e => setSimulatorResource(e.target.value)}
            style={{
              background: 'rgba(15, 23, 42, 0.8)',
              border: '1px solid var(--border-glass)',
              color: '#fff',
              padding: '0.4rem 0.75rem',
              borderRadius: 6,
              fontSize: '0.85rem'
            }}
          >
            {resources.map(r => <option key={r} value={r}>{r}</option>)}
          </select>

          <button onClick={handleTestPermission} className="btn-primary" style={{ padding: '0.4rem 1rem', fontSize: '0.85rem' }}>
            Check Access
          </button>
        </div>

        {testResult && (
          <div style={{
            background: testResult.allowed ? 'rgba(16, 185, 129, 0.1)' : 'rgba(244, 63, 94, 0.1)',
            border: `1px solid ${testResult.allowed ? 'rgba(16, 185, 129, 0.4)' : 'rgba(244, 63, 94, 0.4)'}`,
            color: testResult.allowed ? '#34d399' : '#fb7185',
            padding: '0.75rem 1rem',
            borderRadius: 8,
            fontSize: '0.85rem'
          }}>
            {testResult.reason}
          </div>
        )}
      </div>
    </div>
  );
};
