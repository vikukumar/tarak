import React from 'react';
import { Rocket, Download, Play, Terminal, Code, Check } from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

export const GettingStarted: React.FC<Props> = ({ onToast }) => {
  const copy = (txt: string) => {
    navigator.clipboard.writeText(txt);
    onToast('Copied to clipboard!');
  };

  return (
    <div style={{ maxWidth: 900, margin: '0 auto', padding: '2rem 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <span className="badge badge-cyan" style={{ marginBottom: '1rem' }}>
          <Rocket size={14} /> Quickstart Guide
        </span>
        <h1 style={{ fontSize: 'clamp(2.2rem, 4vw, 3rem)', fontWeight: 800, color: '#fff' }}>
          Get Started in <span className="text-gradient">30 Seconds</span>
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.1rem', marginTop: '0.5rem' }}>
          Install Tarak, boot a local cluster, and deploy your first workload.
        </p>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
        {/* Step 1 */}
        <div className="glass-card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1rem' }}>
            <span style={{
              background: 'var(--accent-cyan)',
              color: '#000',
              width: 28,
              height: 28,
              borderRadius: '50%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 800,
              fontSize: '0.9rem'
            }}>1</span>
            <h3 style={{ color: '#fff', fontSize: '1.3rem' }}>Install Tarak Binaries</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
            Run the automated installation script to download all 4 core binaries (<code>tarak</code>, <code>tarakd</code>, <code>taraks</code>, <code>tarakctl</code>):
          </p>
          <div className="code-box" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>curl -fsSL https://tarak.vikshro.in/install.sh | bash</span>
            <button onClick={() => copy('curl -fsSL https://tarak.vikshro.in/install.sh | bash')} style={{ background: 'transparent', border: 'none', color: 'var(--accent-cyan)', cursor: 'pointer', fontWeight: 600 }}>Copy</button>
          </div>
        </div>

        {/* Step 2 */}
        <div className="glass-card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1rem' }}>
            <span style={{
              background: 'var(--accent-purple)',
              color: '#fff',
              width: 28,
              height: 28,
              borderRadius: '50%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 800,
              fontSize: '0.9rem'
            }}>2</span>
            <h3 style={{ color: '#fff', fontSize: '1.3rem' }}>Launch Tarak All-in-One Engine</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
            Start the control plane API server and node runtime in a single standalone process:
          </p>
          <div className="code-box" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>tarak</span>
            <button onClick={() => copy('tarak')} style={{ background: 'transparent', border: 'none', color: 'var(--accent-cyan)', cursor: 'pointer', fontWeight: 600 }}>Copy</button>
          </div>
        </div>

        {/* Step 3 */}
        <div className="glass-card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1rem' }}>
            <span style={{
              background: 'var(--accent-emerald)',
              color: '#000',
              width: 28,
              height: 28,
              borderRadius: '50%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 800,
              fontSize: '0.9rem'
            }}>3</span>
            <h3 style={{ color: '#fff', fontSize: '1.3rem' }}>Deploy Applications & Ingress</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
            Create and manage Kubernetes-compatible workloads declaratively using <code>tarakctl</code>:
          </p>
          <div className="code-box" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <span>tarakctl apply -f app-sample.yml</span>
            <button onClick={() => copy('tarakctl apply -f app-sample.yml')} style={{ background: 'transparent', border: 'none', color: 'var(--accent-cyan)', cursor: 'pointer', fontWeight: 600 }}>Copy</button>
          </div>
          <div className="code-box" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>tarakctl get pods</span>
            <button onClick={() => copy('tarakctl get pods')} style={{ background: 'transparent', border: 'none', color: 'var(--accent-cyan)', cursor: 'pointer', fontWeight: 600 }}>Copy</button>
          </div>
        </div>

        {/* Step 4 */}
        <div className="glass-card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1rem' }}>
            <span style={{
              background: 'var(--accent-blue)',
              color: '#fff',
              width: 28,
              height: 28,
              borderRadius: '50%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 800,
              fontSize: '0.9rem'
            }}>4</span>
            <h3 style={{ color: '#fff', fontSize: '1.3rem' }}>Use the Go SDK Client</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
            Import Tarak directly into your Go services with zero overhead:
          </p>
          <div className="code-box" style={{ marginBottom: '1rem' }}>
            go get -u github.com/vikukumar/tarak/pkg/client@latest
          </div>
          <pre className="code-box">{`package main

import (
    "context"
    "fmt"
    "github.com/vikukumar/tarak/pkg/client"
)

func main() {
    c, err := client.NewClient("https://127.0.0.1:8443", client.WithInsecure())
    if err != nil {
        panic(err)
    }

    pods, _ := c.Pods("default").List(context.Background())
    fmt.Printf("Active pods: %d\\n", len(pods.Items))
}`}</pre>
        </div>
      </div>
    </div>
  );
};
