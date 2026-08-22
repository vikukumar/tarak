import React, { useState } from 'react';
import { Terminal as TermIcon, Play, RotateCcw, FileText, Search, Copy, Check } from 'lucide-react';

interface Props {
  namespace: string;
  onToast: (msg: string) => void;
}

export const WebTerminal: React.FC<Props> = ({ namespace, onToast }) => {
  const [selectedPod, setSelectedPod] = useState<string>('storefront-web-6d9b7c-7d2x1');
  const [activeMode, setActiveMode] = useState<'exec' | 'logs'>('exec');
  const [command, setCommand] = useState<string>('uname -a');
  const [history, setHistory] = useState<Array<{ cmd: string; out: string; code: number }>>([
    { cmd: 'uname -a', out: 'Linux tarak-worker-01 6.8.0-tarak-x86_64 #1 SMP PREEMPT GNU/Linux', code: 0 },
    { cmd: 'ps aux', out: 'USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND\nroot         1  0.0  0.1  22104  3200 ?        Ss   04:00   0:01 /tcr/app-entrypoint\nroot        14  0.0  0.2  45120  4800 ?        S    04:00   0:00 nginx: master process', code: 0 }
  ]);

  const [logs, setLogs] = useState<string>(`[04:40:01.120] 2026-08-23T04:40:01Z INFO  [server] Ingress listener bound to port 80 (HTTP)
[04:40:01.122] 2026-08-23T04:40:01Z INFO  [auth] Zero-trust mTLS CA certificate loaded successfully
[04:40:01.125] 2026-08-23T04:40:01Z DEBUG [router] Matched Host: store.vikshro.in -> backend: storefront-svc:80
[04:40:02.341] 2026-08-23T04:40:02Z INFO  [audit] 10.244.0.12 "GET /api/v1/healthz HTTP/1.1" 200 48 0.4ms
[04:40:05.890] 2026-08-23T04:40:05Z INFO  [audit] 10.244.0.15 "GET /products/featured HTTP/1.1" 200 4096 1.8ms`);

  const runExec = () => {
    if (!command.trim()) return;
    const newEntry = {
      cmd: command,
      out: `Executed '${command}' inside ${selectedPod} (${namespace})\nExit Code: 0\nResult: Execution completed in 1.2ms with zero overhead.`,
      code: 0
    };
    setHistory([...history, newEntry]);
    setCommand('');
  };

  const presetCmds = ['uname -a', 'env', 'ps aux', 'df -h', 'cat /etc/hosts', 'curl -I http://localhost:80'];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Terminal Toolbar */}
      <div className="glass-card" style={{ padding: '1.25rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
          <div style={{ display: 'flex', gap: '0.4rem' }}>
            <button
              onClick={() => setActiveMode('exec')}
              style={{
                background: activeMode === 'exec' ? 'rgba(0, 240, 255, 0.15)' : 'transparent',
                color: activeMode === 'exec' ? 'var(--accent-cyan)' : 'var(--text-muted)',
                border: activeMode === 'exec' ? '1px solid rgba(0, 240, 255, 0.4)' : '1px solid transparent',
                borderRadius: 6,
                padding: '0.4rem 0.85rem',
                fontSize: '0.85rem',
                fontWeight: 600,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 6
              }}
            >
              <TermIcon size={14} />
              <span>Container Exec Console</span>
            </button>
            <button
              onClick={() => setActiveMode('logs')}
              style={{
                background: activeMode === 'logs' ? 'rgba(0, 240, 255, 0.15)' : 'transparent',
                color: activeMode === 'logs' ? 'var(--accent-cyan)' : 'var(--text-muted)',
                border: activeMode === 'logs' ? '1px solid rgba(0, 240, 255, 0.4)' : '1px solid transparent',
                borderRadius: 6,
                padding: '0.4rem 0.85rem',
                fontSize: '0.85rem',
                fontWeight: 600,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 6
              }}
            >
              <FileText size={14} />
              <span>Live Logs Tail</span>
            </button>
          </div>

          {/* Pod Selector */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.82rem' }}>Pod:</span>
            <select
              value={selectedPod}
              onChange={e => setSelectedPod(e.target.value)}
              style={{
                background: 'rgba(15, 23, 42, 0.8)',
                border: '1px solid var(--border-glass)',
                color: '#fff',
                borderRadius: 6,
                padding: '0.35rem 0.6rem',
                fontSize: '0.82rem',
                outline: 'none'
              }}
            >
              <option value="storefront-web-6d9b7c-7d2x1">storefront-web-6d9b7c-7d2x1</option>
              <option value="storefront-web-6d9b7c-8m4k9">storefront-web-6d9b7c-8m4k9</option>
              <option value="auth-service-58f79-22a1">auth-service-58f79-22a1</option>
              <option value="db-primary-0">db-primary-0</option>
            </select>
          </div>
        </div>
      </div>

      {/* Mode View */}
      {activeMode === 'exec' ? (
        <div className="glass-card" style={{ padding: '0', background: '#050811', overflow: 'hidden' }}>
          {/* Quick Presets */}
          <div style={{ padding: '0.6rem 1rem', background: 'rgba(15, 23, 42, 0.8)', borderBottom: '1px solid var(--border-glass)', display: 'flex', gap: '0.5rem', overflowX: 'auto' }}>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.78rem', alignSelf: 'center' }}>Presets:</span>
            {presetCmds.map((cmd, idx) => (
              <button
                key={idx}
                onClick={() => setCommand(cmd)}
                style={{
                  background: 'rgba(255, 255, 255, 0.05)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  color: 'var(--accent-cyan)',
                  borderRadius: 4,
                  padding: '2px 8px',
                  fontSize: '0.75rem',
                  fontFamily: 'var(--font-mono)',
                  cursor: 'pointer'
                }}
              >
                {cmd}
              </button>
            ))}
          </div>

          {/* Terminal Output */}
          <div style={{ padding: '1.25rem', minHeight: 280, maxHeight: 420, overflowY: 'auto', fontFamily: 'var(--font-mono)', fontSize: '0.88rem', color: '#38bdf8' }}>
            {history.map((h, idx) => (
              <div key={idx} style={{ marginBottom: '1.25rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#fff', fontWeight: 600 }}>
                  <span style={{ color: 'var(--accent-purple)' }}>root@{selectedPod}:/#</span>
                  <span>{h.cmd}</span>
                </div>
                <pre style={{ color: '#94a3b8', whiteSpace: 'pre-wrap', marginTop: '0.4rem', fontSize: '0.85rem' }}>{h.out}</pre>
              </div>
            ))}
          </div>

          {/* Terminal Input Line */}
          <div style={{ padding: '0.75rem 1.25rem', borderTop: '1px solid var(--border-glass)', display: 'flex', alignItems: 'center', gap: '0.75rem', background: '#020408' }}>
            <span style={{ color: 'var(--accent-purple)', fontFamily: 'var(--font-mono)', fontWeight: 700 }}>❯</span>
            <input
              type="text"
              placeholder="Enter command to exec (e.g. env, sh, curl, netstat)..."
              value={command}
              onChange={e => setCommand(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && runExec()}
              style={{
                flex: 1,
                background: 'transparent',
                border: 'none',
                color: '#fff',
                fontFamily: 'var(--font-mono)',
                fontSize: '0.9rem',
                outline: 'none'
              }}
            />
            <button onClick={runExec} className="btn-primary" style={{ padding: '0.35rem 0.85rem', fontSize: '0.8rem' }}>
              <Play size={12} />
              <span>Run</span>
            </button>
          </div>
        </div>
      ) : (
        <div className="glass-card" style={{ padding: '1.25rem', background: '#050811' }}>
          <pre style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '0.85rem',
            color: '#a7f3d0',
            lineHeight: 1.7,
            whiteSpace: 'pre-wrap',
            maxHeight: 400,
            overflowY: 'auto'
          }}>
            {logs}
          </pre>
        </div>
      )}
    </div>
  );
};
