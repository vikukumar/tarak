import React, { useState, useEffect } from 'react';
import { Sparkles, Copy, Check, Terminal, ExternalLink, ShieldCheck, Zap, Globe } from 'lucide-react';

interface HeroProps {
  onExplore: (tab: string) => void;
  onToast: (msg: string) => void;
}

export const Hero: React.FC<HeroProps> = ({ onExplore, onToast }) => {
  const [activeOS, setActiveOS] = useState<'bash' | 'powershell' | 'go'>('bash');
  const [copied, setCopied] = useState(false);
  const [latestVersion, setLatestVersion] = useState('v1.0.6');

  const commands = {
    bash: 'curl -fsSL https://tarak.vikshro.in/install.sh | bash',
    powershell: 'irm https://tarak.vikshro.in/install.ps1 | iex',
    go: 'go get -u github.com/vikukumar/tarak/pkg/client@latest'
  };

  useEffect(() => {
    fetch('https://api.github.com/repos/vikukumar/tarak/releases/latest')
      .then(res => res.json())
      .then(data => {
        if (data.tag_name) setLatestVersion(data.tag_name);
      })
      .catch(() => {});
  }, []);

  const handleCopy = () => {
    navigator.clipboard.writeText(commands[activeOS]);
    setCopied(true);
    onToast('Command copied to clipboard!');
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section style={{
      textAlign: 'center',
      padding: '4rem 1rem 3rem 1rem',
      maxWidth: 960,
      margin: '0 auto'
    }}>
      {/* Release Badge */}
      <div style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '0.5rem',
        background: 'rgba(0, 240, 255, 0.08)',
        border: '1px solid rgba(0, 240, 255, 0.3)',
        color: 'var(--accent-cyan)',
        padding: '0.4rem 1.2rem',
        borderRadius: 9999,
        fontSize: '0.88rem',
        fontWeight: 600,
        marginBottom: '1.75rem',
        boxShadow: '0 0 20px rgba(0, 240, 255, 0.15)'
      }}>
        <Sparkles size={16} />
        <span>Latest Release: <b>{latestVersion}</b></span>
      </div>

      {/* Main Title */}
      <h1 style={{
        fontSize: 'clamp(2.5rem, 5.5vw, 4.4rem)',
        fontWeight: 800,
        lineHeight: 1.12,
        letterSpacing: '-1.5px',
        marginBottom: '1.5rem'
      }}>
        Next-Gen <span className="text-gradient-silver">Container Orchestrator</span>
      </h1>

      {/* Subtitle */}
      <p style={{
        fontSize: 'clamp(1.05rem, 2vw, 1.25rem)',
        color: 'var(--text-secondary)',
        lineHeight: 1.7,
        marginBottom: '2.5rem',
        maxWidth: 780,
        margin: '0 auto 2.5rem auto'
      }}>
        A pure Go, zero-dependency orchestrator that runs anywhere. <b>10x faster</b> and <b>20x lighter</b> than Kubernetes & K3s with native TCR runtime, inbuilt Cloudflare tunnels, and Tailscale private mesh.
      </p>

      {/* Action Buttons */}
      <div style={{
        display: 'flex',
        gap: '1rem',
        justifyContent: 'center',
        flexWrap: 'wrap',
        marginBottom: '3rem'
      }}>
        <button onClick={() => onExplore('getting-started')} className="btn-primary">
          <Zap size={18} />
          <span>Get Started in 30s</span>
        </button>
        <a 
          href="https://github.com/vikukumar/tarak" 
          target="_blank" 
          rel="noopener noreferrer" 
          className="btn-secondary"
        >
          <ExternalLink size={18} />
          <span>Star on GitHub</span>
        </a>
      </div>

      {/* Interactive Glass Terminal */}
      <div className="glass-card" style={{
        textAlign: 'left',
        maxWidth: 780,
        margin: '0 auto',
        borderRadius: 14,
        background: 'rgba(8, 12, 22, 0.92)'
      }}>
        {/* Terminal Header */}
        <div style={{
          background: 'rgba(15, 23, 42, 0.85)',
          borderBottom: '1px solid var(--border-glass)',
          padding: '0.75rem 1.25rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '0.5rem'
        }}>
          <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
            <span style={{ width: 11, height: 11, borderRadius: '50%', background: '#ff5f56' }} />
            <span style={{ width: 11, height: 11, borderRadius: '50%', background: '#ffbd2e' }} />
            <span style={{ width: 11, height: 11, borderRadius: '50%', background: '#27c93f' }} />
            <span style={{ marginLeft: 8, fontSize: '0.8rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>
              quick-install
            </span>
          </div>

          <div style={{ display: 'flex', gap: '0.35rem' }}>
            {(['bash', 'powershell', 'go'] as const).map(os => (
              <button
                key={os}
                onClick={() => setActiveOS(os)}
                style={{
                  background: activeOS === os ? 'rgba(0, 240, 255, 0.12)' : 'transparent',
                  color: activeOS === os ? 'var(--accent-cyan)' : 'var(--text-muted)',
                  border: activeOS === os ? '1px solid rgba(0, 240, 255, 0.3)' : '1px solid transparent',
                  padding: '0.25rem 0.65rem',
                  borderRadius: 6,
                  fontSize: '0.8rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                  transition: 'all 0.2s ease'
                }}
              >
                {os === 'bash' ? 'Linux / macOS' : os === 'powershell' ? 'PowerShell' : 'Go SDK'}
              </button>
            ))}
          </div>
        </div>

        {/* Terminal Body */}
        <div style={{
          padding: '1.25rem 1.5rem',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          gap: '1rem',
          fontFamily: 'var(--font-mono)',
          fontSize: '0.92rem',
          color: '#38bdf8',
          overflowX: 'auto'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', minWidth: 0 }}>
            <span style={{ color: 'var(--accent-purple)' }}>❯</span>
            <span style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
              {commands[activeOS]}
            </span>
          </div>

          <button
            onClick={handleCopy}
            style={{
              background: copied ? 'var(--accent-emerald)' : 'rgba(255, 255, 255, 0.08)',
              border: '1px solid var(--border-glass)',
              color: copied ? '#000' : 'var(--text-primary)',
              padding: '0.4rem 0.85rem',
              borderRadius: 6,
              fontSize: '0.82rem',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              whiteSpace: 'nowrap',
              transition: 'all 0.2s ease'
            }}
          >
            {copied ? <Check size={14} /> : <Copy size={14} />}
            <span>{copied ? 'Copied' : 'Copy'}</span>
          </button>
        </div>
      </div>

      {/* Feature Highlights Pill Grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
        gap: '1rem',
        marginTop: '3.5rem'
      }}>
        <div className="glass-card" style={{ padding: '1.25rem', textAlign: 'left' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '0.5rem' }}>
            <Zap size={20} color="var(--accent-cyan)" />
            <h4 style={{ color: '#fff', fontSize: '1rem' }}>Instant Boot</h4>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.88rem' }}>Under 180ms startup with ~22MB RAM idle footprint.</p>
        </div>

        <div className="glass-card" style={{ padding: '1.25rem', textAlign: 'left' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '0.5rem' }}>
            <ShieldCheck size={20} color="var(--accent-purple)" />
            <h4 style={{ color: '#fff', fontSize: '1rem' }}>Zero Dependencies</h4>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.88rem' }}>No Docker, containerd, WSL2, or external daemons required.</p>
        </div>

        <div className="glass-card" style={{ padding: '1.25rem', textAlign: 'left' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '0.5rem' }}>
            <Globe size={20} color="var(--accent-emerald)" />
            <h4 style={{ color: '#fff', fontSize: '1rem' }}>Inbuilt Mesh</h4>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.88rem' }}>Cloudflare Quick Tunnels & Tailscale MagicDNS out of the box.</p>
        </div>
      </div>
    </section>
  );
};
