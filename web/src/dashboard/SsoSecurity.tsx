import React, { useState, useEffect } from 'react';
import { Key, Shield, Github, Globe, CheckCircle2, Copy, Plus, Lock, RefreshCw } from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

interface Provider {
  name: string;
  type: string;
  enabled: boolean;
  issuer?: string;
  status?: string;
}

export const SsoSecurity: React.FC<Props> = ({ onToast }) => {
  const [currentUser, setCurrentUser] = useState({
    username: 'tarak-admin',
    email: 'admin@tarak.local',
    roles: ['system:masters', 'cluster-admin'],
    provider: 'local-mtls'
  });

  const [providers, setProviders] = useState<Provider[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [tokens, setTokens] = useState<Array<{ id: string; desc: string; created: string; expires: string; val: string }>>([]);
  const [newTokenDesc, setNewTokenDesc] = useState('');

  const fetchProviders = async () => {
    setIsLoading(true);
    try {
      const res = await fetch('/apis/auth.tarak.io/v1/providers');
      if (res.ok) {
        const data = await res.json();
        setProviders(data.items || []);
      }
    } catch {
      // Fallback
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchProviders();
  }, []);

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

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Active User Security Identity */}
      <div className="glass-card" style={{ padding: '1.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{
            background: 'rgba(57, 255, 20, 0.12)',
            border: '1px solid rgba(57, 255, 20, 0.3)',
            borderRadius: 12,
            padding: '0.85rem',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>
            <Shield size={28} color="var(--accent-green)" />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <h3 style={{ color: '#fff', fontSize: '1.2rem' }}>{currentUser.username}</h3>
              <span style={{
                background: 'rgba(57, 255, 20, 0.15)',
                color: 'var(--accent-green)',
                padding: '2px 8px',
                borderRadius: 4,
                fontSize: '0.72rem',
                fontWeight: 700
              }}>
                SUPER-ADMIN (mTLS)
              </span>
            </div>
            <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.2rem' }}>
              {currentUser.email} • Assigned Groups: <code>{currentUser.roles.join(', ')}</code>
            </div>
          </div>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button onClick={() => onToast('mTLS client cert valid for 365 days')} className="btn-secondary" style={{ padding: '0.45rem 0.9rem', fontSize: '0.85rem' }}>
            <Lock size={14} /> Certificate Details
          </button>
        </div>
      </div>

      {/* Remote CLI Access & Login Guide */}
      <div className="glass-card" style={{ padding: '1.5rem', border: '1px solid rgba(0, 240, 255, 0.25)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.75rem' }}>
          <Key size={20} color="var(--accent-cyan)" />
          <h3 style={{ color: '#fff', fontSize: '1.1rem' }}>Remote CLI Access & Authentication</h3>
        </div>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginBottom: '1rem', lineHeight: 1.5 }}>
          Local CLI access functions as Super-Admin by default. For remote developers or CI/CD pipelines, connect using <code>tarakctl login</code>:
        </p>
        <div style={{ background: 'rgba(10, 15, 30, 0.9)', padding: '1rem', borderRadius: 8, fontFamily: 'var(--font-mono)', fontSize: '0.85rem', color: 'var(--accent-cyan)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>tarakctl login https://cluster-api.domain.com --username developer --password *****</span>
          <button onClick={() => { navigator.clipboard.writeText('tarakctl login https://127.0.0.1:18443'); onToast('Command copied!'); }} style={{ background: 'transparent', border: 'none', color: '#fff', cursor: 'pointer' }}>
            <Copy size={16} />
          </button>
        </div>
      </div>

      {/* SSO Identity Providers */}
      <div className="glass-card" style={{ padding: '1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '1rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <Globe size={18} color="var(--accent-cyan)" /> Single Sign-On (SSO) & OIDC Identity Providers
          </h3>
          <button onClick={fetchProviders} className="btn-secondary" style={{ padding: '0.4rem 0.75rem', fontSize: '0.8rem' }}>
            <RefreshCw size={13} className={isLoading ? 'spin' : ''} /> Refresh Providers
          </button>
        </div>

        {providers.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '2rem 1rem', color: 'var(--text-muted)' }}>
            <Globe size={32} style={{ margin: '0 auto 0.5rem auto', opacity: 0.4 }} />
            <p style={{ fontSize: '0.92rem', color: '#fff', fontWeight: 600 }}>Default Local PKI & mTLS Active</p>
            <p style={{ fontSize: '0.8rem', marginTop: '0.2rem' }}>Configure GitHub, Google, or Okta OIDC in server config for enterprise single sign-on.</p>
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: '1rem' }}>
            {providers.map((p, idx) => (
              <div key={idx} style={{
                background: 'rgba(10, 15, 30, 0.6)',
                border: '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '1.25rem'
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                  <h4 style={{ color: '#fff', fontSize: '0.95rem' }}>{p.name}</h4>
                  <span style={{
                    background: p.enabled ? 'rgba(57, 255, 20, 0.15)' : 'rgba(255, 255, 255, 0.05)',
                    color: p.enabled ? 'var(--accent-green)' : 'var(--text-muted)',
                    padding: '2px 8px',
                    borderRadius: 4,
                    fontSize: '0.75rem',
                    fontWeight: 600
                  }}>
                    {p.enabled ? 'Connected' : 'Configured'}
                  </span>
                </div>
                <div style={{ color: 'var(--text-secondary)', fontSize: '0.82rem' }}>Type: <code>{p.type}</code></div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Service Tokens & API Keys */}
      <div className="glass-card" style={{ padding: '1.5rem' }}>
        <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
          <Key size={18} color="var(--accent-purple)" /> API Service Accounts & CI/CD Tokens
        </h3>

        <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.25rem' }}>
          <input
            type="text"
            placeholder="Token description (e.g. GitHub Actions Sync Token)..."
            value={newTokenDesc}
            onChange={e => setNewTokenDesc(e.target.value)}
            style={{
              flex: 1,
              background: 'rgba(15, 23, 42, 0.8)',
              border: '1px solid var(--border-glass)',
              color: '#fff',
              padding: '0.45rem 0.85rem',
              borderRadius: 6,
              fontSize: '0.85rem',
              outline: 'none'
            }}
          />
          <button onClick={handleCreateToken} className="btn-primary" style={{ padding: '0.45rem 1.1rem', fontSize: '0.85rem' }}>
            <Plus size={15} /> Generate Token
          </button>
        </div>

        {tokens.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '1.5rem 1rem', color: 'var(--text-muted)', fontSize: '0.85rem' }}>
            No custom service tokens created yet. Generate one above for automated CLI or CI/CD access.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {tokens.map((tok, idx) => (
              <div key={idx} style={{
                background: 'rgba(10, 15, 30, 0.6)',
                border: '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '1rem',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                flexWrap: 'wrap',
                gap: '0.5rem'
              }}>
                <div>
                  <div style={{ color: '#fff', fontWeight: 600, fontSize: '0.92rem' }}>{tok.desc}</div>
                  <div style={{ color: 'var(--accent-cyan)', fontFamily: 'var(--font-mono)', fontSize: '0.82rem', marginTop: '0.2rem' }}>{tok.val}</div>
                </div>
                <button onClick={() => copyToken(tok.val)} className="btn-secondary" style={{ padding: '0.35rem 0.75rem', fontSize: '0.78rem' }}>
                  <Copy size={13} /> Copy Key
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
