import React, { useState } from 'react';
import { Key, Shield, Github, Globe, CheckCircle2, Copy, Plus, Lock } from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

export const SsoSecurity: React.FC<Props> = ({ onToast }) => {
  const [currentUser, setCurrentUser] = useState({
    username: 'admin',
    email: 'admin@tarak.local',
    roles: ['cluster-admin'],
    provider: 'local-mtls'
  });

  const [tokens, setTokens] = useState([
    { id: 'tok-9fa2', desc: 'CI/CD GitHub Actions Deploy Token', created: '2026-08-20', expires: '2026-09-20', val: 'trk_sec_9fa281c7e2b10...' },
    { id: 'tok-31b0', desc: 'ArgoCD GitOps Sync Token', created: '2026-08-22', expires: '2027-08-22', val: 'trk_sec_31b090a1f77d...' }
  ]);

  const [newTokenDesc, setNewTokenDesc] = useState('');

  const providers = [
    { name: 'GitHub Enterprise & Cloud', type: 'github', icon: 'github', status: 'Connected' },
    { name: 'Google Workspace', type: 'google', icon: 'globe', status: 'Ready' },
    { name: 'Microsoft Entra ID (Azure AD)', type: 'microsoft', icon: 'shield', status: 'Ready' },
    { name: 'Okta Workforce Identity', type: 'okta', icon: 'lock', status: 'Ready' },
    { name: 'Keycloak Identity Provider', type: 'keycloak', icon: 'key', status: 'Ready' },
    { name: 'Generic OIDC (OpenID Connect)', type: 'oidc', icon: 'globe', status: 'Ready' }
  ];

  const handleCreateToken = () => {
    if (!newTokenDesc.trim()) return;
    const newToken = {
      id: `tok-${Math.random().toString(36).substring(2, 6)}`,
      desc: newTokenDesc,
      created: 'Today',
      expires: '30 Days',
      val: `trk_sec_${Math.random().toString(36).substring(2, 14)}...`
    };
    setTokens([...tokens, newToken]);
    setNewTokenDesc('');
    onToast('API Service Token generated!');
  };

  const copyToken = (val: string) => {
    navigator.clipboard.writeText(val);
    onToast('Token copied to clipboard!');
  };

  const handleSsoConnect = (name: string) => {
    onToast(`Initiated SSO OAuth2 flow with ${name}`);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Current Active Identity */}
      <div className="glass-card" style={{ padding: '1.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.8rem', textTransform: 'uppercase' }}>Active Authentication Profile</span>
          <h3 style={{ color: '#fff', fontSize: '1.3rem', marginTop: '0.2rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span>@{currentUser.username}</span>
            <span className="badge badge-emerald"><CheckCircle2 size={12} /> {currentUser.roles[0]}</span>
          </h3>
          <small style={{ color: 'var(--text-secondary)' }}>Provider: {currentUser.provider} • Zero-Trust PKI mTLS</small>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <span className="badge badge-cyan">Cluster Superuser</span>
        </div>
      </div>

      {/* SSO Providers Grid */}
      <div className="glass-card" style={{ padding: '1.5rem' }}>
        <h3 style={{ color: '#fff', fontSize: '1.15rem', marginBottom: '1.25rem', display: 'flex', alignItems: 'center', gap: 6 }}>
          <Shield size={18} color="var(--accent-cyan)" /> Supported Universal SSO Providers
        </h3>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1rem' }}>
          {providers.map((p, idx) => (
            <div key={idx} style={{
              background: 'rgba(15, 23, 42, 0.7)',
              border: '1px solid var(--border-glass)',
              borderRadius: 10,
              padding: '1.25rem',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center'
            }}>
              <div>
                <h4 style={{ color: '#fff', fontSize: '0.95rem', margin: '0 0 0.3rem 0' }}>{p.name}</h4>
                <span className={p.status === 'Connected' ? 'badge badge-emerald' : 'badge badge-cyan'} style={{ fontSize: '0.72rem' }}>
                  {p.status}
                </span>
              </div>

              <button
                onClick={() => handleSsoConnect(p.name)}
                style={{
                  background: 'rgba(0, 240, 255, 0.1)',
                  border: '1px solid rgba(0, 240, 255, 0.3)',
                  color: 'var(--accent-cyan)',
                  padding: '0.35rem 0.75rem',
                  borderRadius: 6,
                  fontSize: '0.8rem',
                  fontWeight: 600,
                  cursor: 'pointer'
                }}
              >
                Login
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* API Service Tokens */}
      <div className="glass-card" style={{ padding: '1.5rem' }}>
        <h3 style={{ color: '#fff', fontSize: '1.15rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
          <Key size={18} color="var(--accent-purple)" /> Cluster API Service Tokens
        </h3>

        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
          <input
            type="text"
            placeholder="New token description (e.g. Jenkins CI, Terraform)..."
            value={newTokenDesc}
            onChange={e => setNewTokenDesc(e.target.value)}
            style={{
              flex: 1,
              minWidth: 240,
              background: 'rgba(15, 23, 42, 0.8)',
              border: '1px solid var(--border-glass)',
              color: '#fff',
              padding: '0.45rem 0.85rem',
              borderRadius: 6,
              fontSize: '0.88rem',
              outline: 'none'
            }}
          />
          <button onClick={handleCreateToken} className="btn-primary" style={{ padding: '0.45rem 1rem', fontSize: '0.85rem' }}>
            <Plus size={14} /> Generate Token
          </button>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.6rem' }}>
          {tokens.map((t, idx) => (
            <div key={idx} style={{
              background: 'rgba(10, 15, 30, 0.6)',
              border: '1px solid var(--border-glass)',
              borderRadius: 8,
              padding: '0.85rem 1.2rem',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              flexWrap: 'wrap',
              gap: '0.6rem'
            }}>
              <div>
                <div style={{ color: '#fff', fontWeight: 600, fontSize: '0.9rem' }}>{t.desc}</div>
                <small style={{ color: 'var(--text-muted)' }}>Created: {t.created} • Expires: {t.expires}</small>
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                <code style={{ fontSize: '0.82rem', color: 'var(--accent-cyan)' }}>{t.val}</code>
                <button
                  onClick={() => copyToken(t.val)}
                  style={{
                    background: 'rgba(255, 255, 255, 0.06)',
                    border: '1px solid var(--border-glass)',
                    color: '#fff',
                    borderRadius: 4,
                    padding: '3px 8px',
                    fontSize: '0.75rem',
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 4
                  }}
                >
                  <Copy size={12} />
                  <span>Copy</span>
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
