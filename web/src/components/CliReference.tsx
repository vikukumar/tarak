import React, { useState } from 'react';
import { Terminal, Search, Copy, Check } from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

export const CliReference: React.FC<Props> = ({ onToast }) => {
  const [search, setSearch] = useState('');

  const commands = [
    { cmd: 'tarakctl get pods', alias: 'po', desc: 'List all running and completed pods in current namespace' },
    { cmd: 'tarakctl get nodes', alias: 'no', desc: 'List master control plane and worker nodes' },
    { cmd: 'tarakctl get services', alias: 'svc', desc: 'List ClusterIP and NodePort services' },
    { cmd: 'tarakctl apply -f <file.yml>', alias: '-', desc: 'Create or update resources declaratively from YAML' },
    { cmd: 'tarakctl delete <kind> <name>', alias: '-', desc: 'Delete a pod, deployment, service, or ingress' },
    { cmd: 'tarakctl logs <pod-name>', alias: '-', desc: 'Stream real-time stdout/stderr container logs' },
    { cmd: 'tarakctl tunnel list', alias: 'tun ls', desc: 'Inspect active Cloudflare & Tailscale tunnel endpoints' },
    { cmd: 'tarakctl version', alias: 'v', desc: 'Print client and server version information' }
  ];

  const filtered = commands.filter(c => 
    c.cmd.toLowerCase().includes(search.toLowerCase()) || 
    c.desc.toLowerCase().includes(search.toLowerCase())
  );

  const copy = (txt: string) => {
    navigator.clipboard.writeText(txt);
    onToast('Command copied to clipboard!');
  };

  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: '2rem 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <span className="badge badge-purple" style={{ marginBottom: '1rem' }}>
          <Terminal size={14} /> Command Cheat Sheet
        </span>
        <h1 style={{ fontSize: 'clamp(2.2rem, 4vw, 3rem)', fontWeight: 800, color: '#fff' }}>
          <code>tarakctl</code> <span className="text-gradient">CLI Reference</span>
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.1rem', marginTop: '0.5rem' }}>
          Complete command-line interface guide with syntax, aliases, and flags.
        </p>
      </div>

      {/* Search Box */}
      <div className="glass-card" style={{ padding: '1rem 1.5rem', marginBottom: '2rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
        <Search size={18} color="var(--text-muted)" />
        <input
          type="text"
          placeholder="Filter commands (e.g. pods, tunnel, logs)..."
          value={search}
          onChange={e => setSearch(e.target.value)}
          style={{
            background: 'transparent',
            border: 'none',
            color: '#fff',
            outline: 'none',
            width: '100%',
            fontSize: '0.95rem'
          }}
        />
      </div>

      {/* Commands List */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        {filtered.map((item, idx) => (
          <div key={idx} className="glass-card" style={{
            padding: '1.2rem 1.5rem',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: '0.75rem'
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
              <code style={{ fontSize: '0.95rem', color: 'var(--accent-cyan)', fontWeight: 600 }}>
                {item.cmd}
              </code>
              {item.alias !== '-' && (
                <span className="badge badge-purple" style={{ fontSize: '0.75rem' }}>
                  alias: {item.alias}
                </span>
              )}
            </div>

            <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
              <span style={{ color: 'var(--text-secondary)', fontSize: '0.88rem' }}>{item.desc}</span>
              <button
                onClick={() => copy(item.cmd)}
                style={{
                  background: 'rgba(255, 255, 255, 0.08)',
                  border: '1px solid var(--border-glass)',
                  color: 'var(--text-primary)',
                  padding: '0.35rem 0.65rem',
                  borderRadius: 6,
                  cursor: 'pointer',
                  fontSize: '0.8rem',
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
  );
};
