import React, { useState, useEffect } from 'react';
import { Shield, Key, Lock, User, CheckCircle2, ArrowRight, Github, Globe, RefreshCw, Zap } from 'lucide-react';

interface Props {
  onAuthenticated: (user: any, token: string) => void;
  onToast: (msg: string) => void;
}

export const AuthPortal: React.FC<Props> = ({ onAuthenticated, onToast }) => {
  const [isSetupRequired, setIsSetupRequired] = useState<boolean>(false);
  const [checkingStatus, setCheckingStatus] = useState<boolean>(true);
  const [authTab, setAuthTab] = useState<'credentials' | 'token' | 'sso'>('credentials');

  // Form states
  const [username, setUsername] = useState<string>('admin');
  const [password, setPassword] = useState<string>('');
  const [confirmPassword, setConfirmPassword] = useState<string>('');
  const [email, setEmail] = useState<string>('admin@tarak.local');
  const [tokenInput, setTokenInput] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [errorMessage, setErrorMessage] = useState<string>('');

  useEffect(() => {
    checkClusterAuthStatus();
  }, []);

  const checkClusterAuthStatus = async () => {
    setCheckingStatus(true);
    try {
      const res = await fetch('/apis/auth.tarak.io/v1/status');
      if (res.ok) {
        const data = await res.json();
        setIsSetupRequired(data.setupRequired ?? false);
      }
    } catch {
      // Fallback
    } finally {
      setCheckingStatus(false);
    }
  };

  const handleSetup = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage('');

    if (password !== confirmPassword) {
      setErrorMessage('Passwords do not match');
      return;
    }
    if (password.length < 4) {
      setErrorMessage('Password must be at least 4 characters');
      return;
    }

    setIsSubmitting(true);
    try {
      const res = await fetch('/apis/auth.tarak.io/v1/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password, email })
      });

      if (res.ok) {
        const data = await res.json();
        localStorage.setItem('tarak_token', data.token);
        onToast('Super-Admin account initialized successfully!');
        onAuthenticated(data.user || { username, roles: ['cluster-admin'] }, data.token);
      } else {
        const err = await res.json();
        setErrorMessage(err.error || 'Setup failed');
      }
    } catch {
      setErrorMessage('Failed to connect to cluster API');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage('');
    setIsSubmitting(true);

    try {
      let body: any = { provider: 'local' };
      if (authTab === 'credentials') {
        body.username = username;
        body.password = password;
      } else if (authTab === 'token') {
        body.token = tokenInput.trim();
        body.username = 'token-user';
      }

      const res = await fetch('/apis/auth.tarak.io/v1/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });

      if (res.ok) {
        const data = await res.json();
        localStorage.setItem('tarak_token', data.token);
        onToast(`Welcome, ${data.user?.username || username}!`);
        onAuthenticated(data.user || { username, roles: ['cluster-admin'] }, data.token);
      } else {
        const err = await res.json();
        setErrorMessage(err.error || 'Invalid credentials');
      }
    } catch {
      setErrorMessage('Unable to connect to cluster auth service');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (checkingStatus) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '60vh' }}>
        <RefreshCw size={28} className="spin" color="var(--accent-cyan)" />
      </div>
    );
  }

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '75vh',
      padding: '2rem 1rem'
    }}>
      <div className="glass-card" style={{
        maxWidth: 480,
        width: '100%',
        padding: '2.5rem 2rem',
        border: '1px solid var(--border-glass)',
        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.4)'
      }}>
        {/* Header Icon & Brand */}
        <div style={{ textAlign: 'center', marginBottom: '1.75rem' }}>
          <div style={{
            width: 64,
            height: 64,
            borderRadius: 16,
            background: 'rgba(0, 240, 255, 0.12)',
            border: '1px solid rgba(0, 240, 255, 0.3)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 1rem auto'
          }}>
            {isSetupRequired ? <Zap size={32} color="var(--accent-green)" /> : <Shield size={32} color="var(--accent-cyan)" />}
          </div>
          <h2 style={{ fontSize: '1.5rem', fontWeight: 800, color: '#fff' }}>
            {isSetupRequired ? 'Initialize Super-Admin' : 'Cluster Dashboard Login'}
          </h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: '0.35rem' }}>
            {isSetupRequired 
              ? 'First-time setup detected: Create the root Super-Admin credentials.'
              : 'Authenticate to access the Tarak cluster control plane.'}
          </p>
        </div>

        {errorMessage && (
          <div style={{
            background: 'rgba(255, 0, 85, 0.12)',
            border: '1px solid rgba(255, 0, 85, 0.3)',
            color: 'var(--accent-pink)',
            padding: '0.65rem 1rem',
            borderRadius: 6,
            fontSize: '0.85rem',
            marginBottom: '1.25rem',
            textAlign: 'center'
          }}>
            {errorMessage}
          </div>
        )}

        {/* 1st-Time Setup Form */}
        {isSetupRequired ? (
          <form onSubmit={handleSetup} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <div>
              <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                Super-Admin Username:
              </label>
              <input
                type="text"
                required
                value={username}
                onChange={e => setUsername(e.target.value)}
                style={{
                  width: '100%',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-glass)',
                  color: '#fff',
                  padding: '0.55rem 0.85rem',
                  borderRadius: 6,
                  fontSize: '0.9rem',
                  outline: 'none'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                Admin Email:
              </label>
              <input
                type="email"
                required
                value={email}
                onChange={e => setEmail(e.target.value)}
                style={{
                  width: '100%',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-glass)',
                  color: '#fff',
                  padding: '0.55rem 0.85rem',
                  borderRadius: 6,
                  fontSize: '0.9rem',
                  outline: 'none'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                Master Password:
              </label>
              <input
                type="password"
                required
                placeholder="••••••••"
                value={password}
                onChange={e => setPassword(e.target.value)}
                style={{
                  width: '100%',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-glass)',
                  color: '#fff',
                  padding: '0.55rem 0.85rem',
                  borderRadius: 6,
                  fontSize: '0.9rem',
                  outline: 'none'
                }}
              />
            </div>

            <div>
              <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                Confirm Password:
              </label>
              <input
                type="password"
                required
                placeholder="••••••••"
                value={confirmPassword}
                onChange={e => setConfirmPassword(e.target.value)}
                style={{
                  width: '100%',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-glass)',
                  color: '#fff',
                  padding: '0.55rem 0.85rem',
                  borderRadius: 6,
                  fontSize: '0.9rem',
                  outline: 'none'
                }}
              />
            </div>

            <button type="submit" disabled={isSubmitting} className="btn-primary" style={{ marginTop: '0.75rem', padding: '0.65rem', fontSize: '0.92rem', justifyContent: 'center' }}>
              {isSubmitting ? 'Initializing Cluster...' : 'Create Super-Admin & Enter'}
            </button>
          </form>
        ) : (
          /* Login Form */
          <div>
            {/* Auth Sub-tabs */}
            <div style={{ display: 'flex', gap: '0.4rem', borderBottom: '1px solid var(--border-glass)', paddingBottom: '0.75rem', marginBottom: '1.25rem' }}>
              <button
                type="button"
                onClick={() => setAuthTab('credentials')}
                className={`tab-btn ${authTab === 'credentials' ? 'active' : ''}`}
                style={{ flex: 1, justifyContent: 'center' }}
              >
                <User size={14} /> Password
              </button>
              <button
                type="button"
                onClick={() => setAuthTab('token')}
                className={`tab-btn ${authTab === 'token' ? 'active' : ''}`}
                style={{ flex: 1, justifyContent: 'center' }}
              >
                <Key size={14} /> Token
              </button>
              <button
                type="button"
                onClick={() => setAuthTab('sso')}
                className={`tab-btn ${authTab === 'sso' ? 'active' : ''}`}
                style={{ flex: 1, justifyContent: 'center' }}
              >
                <Globe size={14} /> SSO
              </button>
            </div>

            {authTab === 'credentials' && (
              <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                    Username:
                  </label>
                  <input
                    type="text"
                    required
                    value={username}
                    onChange={e => setUsername(e.target.value)}
                    style={{
                      width: '100%',
                      background: 'rgba(15, 23, 42, 0.8)',
                      border: '1px solid var(--border-glass)',
                      color: '#fff',
                      padding: '0.55rem 0.85rem',
                      borderRadius: 6,
                      fontSize: '0.9rem',
                      outline: 'none'
                    }}
                  />
                </div>

                <div>
                  <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                    Password:
                  </label>
                  <input
                    type="password"
                    placeholder="••••••••"
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                    style={{
                      width: '100%',
                      background: 'rgba(15, 23, 42, 0.8)',
                      border: '1px solid var(--border-glass)',
                      color: '#fff',
                      padding: '0.55rem 0.85rem',
                      borderRadius: 6,
                      fontSize: '0.9rem',
                      outline: 'none'
                    }}
                  />
                </div>

                <button type="submit" disabled={isSubmitting} className="btn-primary" style={{ marginTop: '0.75rem', padding: '0.65rem', fontSize: '0.92rem', justifyContent: 'center' }}>
                  {isSubmitting ? 'Verifying...' : 'Login to Dashboard'}
                </button>
              </form>
            )}

            {authTab === 'token' && (
              <form onSubmit={handleLogin} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                <div>
                  <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                    Bearer / Master Secret Token:
                  </label>
                  <textarea
                    rows={4}
                    required
                    placeholder="Paste trk_sec_... or JWT token"
                    value={tokenInput}
                    onChange={e => setTokenInput(e.target.value)}
                    style={{
                      width: '100%',
                      background: 'rgba(15, 23, 42, 0.8)',
                      border: '1px solid var(--border-glass)',
                      color: 'var(--accent-green)',
                      fontFamily: 'var(--font-mono)',
                      padding: '0.55rem 0.85rem',
                      borderRadius: 6,
                      fontSize: '0.85rem',
                      outline: 'none'
                    }}
                  />
                </div>

                <button type="submit" disabled={isSubmitting || !tokenInput.trim()} className="btn-primary" style={{ marginTop: '0.75rem', padding: '0.65rem', fontSize: '0.92rem', justifyContent: 'center' }}>
                  {isSubmitting ? 'Validating Token...' : 'Authenticate with Token'}
                </button>
              </form>
            )}

            {authTab === 'sso' && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                <button
                  type="button"
                  onClick={() => onToast('Redirecting to GitHub SSO...')}
                  className="btn-secondary"
                  style={{ padding: '0.6rem 1rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}
                >
                  <Github size={16} /> Continue with GitHub
                </button>
                <button
                  type="button"
                  onClick={() => onToast('Redirecting to Google Workspace...')}
                  className="btn-secondary"
                  style={{ padding: '0.6rem 1rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}
                >
                  <Globe size={16} /> Continue with Google
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};
