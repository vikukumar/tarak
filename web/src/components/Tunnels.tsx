import React from 'react';
import { CloudLightning, Globe, Shield, Terminal, ArrowRight } from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

export const Tunnels: React.FC<Props> = ({ onToast }) => {
  const copy = (txt: string) => {
    navigator.clipboard.writeText(txt);
    onToast('Copied to clipboard!');
  };

  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: '2rem 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <span className="badge badge-cyan" style={{ marginBottom: '1rem' }}>
          <CloudLightning size={14} /> Edge Networking
        </span>
        <h1 style={{ fontSize: 'clamp(2.2rem, 4vw, 3rem)', fontWeight: 800, color: '#fff' }}>
          Inbuilt <span className="text-gradient">Cloudflare & Tailscale Tunnels</span>
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.1rem', marginTop: '0.5rem' }}>
          Publish your apps to the internet or connect your private team mesh with zero port forwarding.
        </p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '1.75rem', marginBottom: '2.5rem' }}>
        {/* Cloudflare Card */}
        <div className="glass-card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '1rem' }}>
            <Globe size={24} color="var(--accent-cyan)" />
            <h3 style={{ color: '#fff', fontSize: '1.25rem' }}>Cloudflare Tunnels</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.92rem', marginBottom: '1.25rem' }}>
            Generates instant public HTTPS URLs (<code>*.trycloudflare.com</code>) or binds to your production custom domains.
          </p>
          <div className="code-box" style={{ fontSize: '0.82rem', marginBottom: '0.75rem' }}>
            tarak --cloudflare-tunnel
          </div>
          <div className="code-box" style={{ fontSize: '0.82rem' }}>
            tarak --cloudflare-tunnel --cloudflare-token &lt;TOKEN&gt;
          </div>
        </div>

        {/* Tailscale Card */}
        <div className="glass-card" style={{ padding: '2rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '1rem' }}>
            <Shield size={24} color="var(--accent-purple)" />
            <h3 style={{ color: '#fff', fontSize: '1.25rem' }}>Tailscale Mesh</h3>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.92rem', marginBottom: '1.25rem' }}>
            Secures inter-service communications and provides zero-trust private access via MagicDNS (<code>*.ts.net</code>).
          </p>
          <div className="code-box" style={{ fontSize: '0.82rem' }}>
            tarak --tailscale --tailscale-authkey &lt;AUTH_KEY&gt;
          </div>
        </div>
      </div>

      {/* Native Ingress YAML */}
      <div className="glass-card" style={{ padding: '2rem', marginBottom: '2rem' }}>
        <h3 style={{ color: '#fff', fontSize: '1.25rem', marginBottom: '0.75rem' }}>
          ⚡ Native Kubernetes Ingress Route
        </h3>
        <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
          Declare your routing rules with <code>ingressClassName: tarak-cloudflare</code> or <code>tarak-tailscale</code>:
        </p>
        <pre className="code-box">{`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-ingress
  namespace: default
spec:
  ingressClassName: tarak-cloudflare
  rules:
  - host: app.vikshro.in
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-app-svc
            port:
              number: 80`}</pre>
      </div>

      {/* CLI Inspection */}
      <div className="glass-card" style={{ padding: '2rem' }}>
        <h3 style={{ color: '#fff', fontSize: '1.25rem', marginBottom: '0.75rem' }}>
          🔍 Inspect Tunnel Status
        </h3>
        <pre className="code-box">{`tarakctl tunnel list

# Output:
# TYPE         STATUS   MODE           PUBLIC URL                               AGE
# CLOUDFLARE   Active   quick-tunnel   https://demo-app-7fa2.trycloudflare.com   5m
# TAILSCALE    Active   magic-dns      https://tarak-cluster.ts.net             5m`}</pre>
      </div>
    </div>
  );
};
