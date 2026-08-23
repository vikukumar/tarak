import React, { useState } from 'react';
import { 
  Network, 
  ShieldCheck, 
  Globe, 
  Sliders, 
  Layers, 
  Activity, 
  Lock, 
  ArrowRight, 
  Plus, 
  CheckCircle2, 
  Server, 
  FileCode, 
  Zap,
  ExternalLink,
  Search,
  Filter
} from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

export const MeshManager: React.FC<Props> = ({ onToast }) => {
  const [selectedMesh, setSelectedMesh] = useState<string>('default');
  const [meshTab, setMeshTab] = useState<'services' | 'external' | 'permissions' | 'routing' | 'passthrough' | 'patches'>('services');
  const [searchQuery, setSearchQuery] = useState<string>('');

  const meshes = [
    { name: 'default', mode: 'Strict mTLS', trustDomain: 'tarak.mesh', servicesCount: 4, passthrough: 'Passthrough', status: 'Healthy' },
    { name: 'finance-mesh', mode: 'Strict mTLS', trustDomain: 'finance.tarak.mesh', servicesCount: 2, passthrough: 'DenyAll', status: 'Healthy' },
    { name: 'prod-mesh', mode: 'Permissive', trustDomain: 'prod.tarak.mesh', servicesCount: 6, passthrough: 'Passthrough', status: 'Healthy' }
  ];

  const meshServices = [
    { name: 'frontend', mesh: 'default', namespace: 'default', dns: 'frontend.default.mesh', vip: '240.240.0.1', port: 80, proto: 'HTTP', endpoints: '10.244.0.12:80', status: 'Healthy', spiffe: 'spiffe://tarak.mesh/ns/default/sa/frontend' },
    { name: 'api-service', mesh: 'default', namespace: 'default', dns: 'api-service.default.mesh', vip: '240.240.0.2', port: 8080, proto: 'HTTP', endpoints: '10.244.0.14:8080', status: 'Healthy', spiffe: 'spiffe://tarak.mesh/ns/default/sa/api-service' },
    { name: 'auth-service', mesh: 'default', namespace: 'default', dns: 'auth-service.default.mesh', vip: '240.240.0.3', port: 9000, proto: 'gRPC', endpoints: '10.244.0.19:9000', status: 'Healthy', spiffe: 'spiffe://tarak.mesh/ns/default/sa/auth-service' },
    { name: 'payments', mesh: 'finance-mesh', namespace: 'finance', dns: 'payments.finance.mesh', vip: '240.240.1.1', port: 8443, proto: 'HTTPS', endpoints: '10.244.1.8:8443', status: 'Healthy', spiffe: 'spiffe://finance.tarak.mesh/ns/finance/sa/payments' }
  ];

  const externalServices = [
    { name: 'stripe-api', mesh: 'default', host: 'api.stripe.com', port: 443, tls: true, sni: 'api.stripe.com' },
    { name: 'aws-rds-postgres', mesh: 'default', host: 'db-prod.internal.aws', port: 5432, tls: true, sni: 'db-prod.internal.aws' },
    { name: 'banking-gateway', mesh: 'finance-mesh', host: 'gw.swift-network.com', port: 443, tls: true, sni: 'gw.swift-network.com' }
  ];

  const permissions = [
    { name: 'allow-frontend-to-api', mesh: 'default', from: 'frontend', to: 'api-service', action: 'ALLOW' },
    { name: 'allow-api-to-auth', mesh: 'default', from: 'api-service', to: 'auth-service', action: 'ALLOW' },
    { name: 'strict-deny-external', mesh: 'finance-mesh', from: '*', to: 'payments', action: 'DENY' }
  ];

  const passthroughPolicies = [
    { name: 'allow-public-web-egress', mesh: 'default', cidrs: ['0.0.0.0/0'], hosts: ['*.github.com', '*.docker.io', '*.cloudflare.com'] },
    { name: 'finance-restricted-egress', mesh: 'finance-mesh', cidrs: ['10.100.0.0/16'], hosts: ['api.stripe.com'] }
  ];

  const proxyPatches = [
    { name: 'http2-multiplexing-tuning', mesh: 'default', target: 'all', connectTimeout: '3000ms', idleTimeout: '60000ms', http2: true },
    { name: 'grpc-keepalive-patch', mesh: 'finance-mesh', target: 'payments', connectTimeout: '1500ms', idleTimeout: '120000ms', http2: true }
  ];

  const filteredServices = meshServices.filter(s => 
    s.mesh === selectedMesh && (s.name.includes(searchQuery) || s.dns.includes(searchQuery))
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Top Banner: Mesh Tenant Selector & Overview */}
      <div className="glass-card" style={{ padding: '1.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{
            background: 'rgba(0, 240, 255, 0.12)',
            border: '1px solid rgba(0, 240, 255, 0.3)',
            borderRadius: 12,
            padding: '0.85rem',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>
            <Network size={28} color="var(--accent-cyan)" />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
              <h2 style={{ fontSize: '1.4rem', fontWeight: 800, color: '#fff' }}>Native Multi-Mesh</h2>
              <span style={{
                background: 'rgba(57, 255, 20, 0.15)',
                color: 'var(--accent-green)',
                border: '1px solid rgba(57, 255, 20, 0.3)',
                padding: '2px 8px',
                borderRadius: 6,
                fontSize: '0.75rem',
                fontWeight: 700
              }}>
                Zero Third-Party (Kuma Equivalent)
              </span>
            </div>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
              Multi-tenant isolated meshes, automatic SPIFFE mTLS Zero-Trust, virtual <code>.mesh</code> DNS, and egress passthrough.
            </p>
          </div>
        </div>

        {/* Mesh Tenant Switcher */}
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Active Mesh:</span>
          <select 
            value={selectedMesh} 
            onChange={e => { setSelectedMesh(e.target.value); onToast(`Switched to mesh: ${e.target.value}`); }}
            style={{
              background: 'rgba(15, 23, 42, 0.8)',
              border: '1px solid var(--accent-cyan)',
              color: '#fff',
              padding: '0.45rem 1rem',
              borderRadius: 6,
              fontSize: '0.88rem',
              fontWeight: 600,
              outline: 'none',
              cursor: 'pointer'
            }}
          >
            {meshes.map((m, idx) => (
              <option key={idx} value={m.name}>{m.name} ({m.mode})</option>
            ))}
          </select>

          <button onClick={() => onToast('Create Mesh modal ready')} className="btn-primary" style={{ padding: '0.45rem 0.9rem', fontSize: '0.85rem' }}>
            <Plus size={14} /> New Mesh
          </button>
        </div>
      </div>

      {/* Sub Navigation Tabs */}
      <div style={{ display: 'flex', gap: '0.5rem', borderBottom: '1px solid var(--border-glass)', paddingBottom: '0.5rem', overflowX: 'auto' }}>
        <button onClick={() => setMeshTab('services')} className={`tab-btn ${meshTab === 'services' ? 'active' : ''}`}>
          <Server size={15} /> Mesh Services ({meshServices.filter(s => s.mesh === selectedMesh).length})
        </button>
        <button onClick={() => setMeshTab('external')} className={`tab-btn ${meshTab === 'external' ? 'active' : ''}`}>
          <Globe size={15} /> External Services ({externalServices.filter(s => s.mesh === selectedMesh).length})
        </button>
        <button onClick={() => setMeshTab('permissions')} className={`tab-btn ${meshTab === 'permissions' ? 'active' : ''}`}>
          <ShieldCheck size={15} /> Traffic Permissions
        </button>
        <button onClick={() => setMeshTab('routing')} className={`tab-btn ${meshTab === 'routing' ? 'active' : ''}`}>
          <Sliders size={15} /> Canary & Routing
        </button>
        <button onClick={() => setMeshTab('passthrough')} className={`tab-btn ${meshTab === 'passthrough' ? 'active' : ''}`}>
          <Zap size={15} /> Passthrough Policies
        </button>
        <button onClick={() => setMeshTab('patches')} className={`tab-btn ${meshTab === 'patches' ? 'active' : ''}`}>
          <FileCode size={15} /> Proxy Patches
        </button>
      </div>

      {/* TAB 1: Mesh Services & DNS */}
      {meshTab === 'services' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '0.75rem' }}>
            <h3 style={{ color: '#fff', fontSize: '1.1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
              <Server size={18} color="var(--accent-cyan)" /> Auto-Enrolled Mesh Services & Inbuilt Virtual DNS
            </h3>
            <div style={{ position: 'relative', minWidth: 220 }}>
              <Search size={14} style={{ position: 'absolute', left: 10, top: 10, color: 'var(--text-muted)' }} />
              <input
                type="text"
                placeholder="Search services or .mesh DNS..."
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                style={{
                  width: '100%',
                  background: 'rgba(15, 23, 42, 0.8)',
                  border: '1px solid var(--border-glass)',
                  color: '#fff',
                  padding: '0.4rem 0.75rem 0.4rem 2rem',
                  borderRadius: 6,
                  fontSize: '0.85rem',
                  outline: 'none'
                }}
              />
            </div>
          </div>

          <div style={{ overflowX: 'auto' }}>
            <table className="data-table" style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ color: 'var(--text-muted)', fontSize: '0.8rem', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.75rem' }}>SERVICE NAME</th>
                  <th style={{ padding: '0.75rem' }}>VIRTUAL .MESH DNS</th>
                  <th style={{ padding: '0.75rem' }}>VIRTUAL VIP</th>
                  <th style={{ padding: '0.75rem' }}>PORT / PROTO</th>
                  <th style={{ padding: '0.75rem' }}>SPIFFE WORKLOAD IDENTITY</th>
                  <th style={{ padding: '0.75rem' }}>STATUS</th>
                </tr>
              </thead>
              <tbody>
                {filteredServices.map((svc, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)', fontSize: '0.88rem' }}>
                    <td style={{ padding: '0.75rem', color: '#fff', fontWeight: 600 }}>{svc.name}</td>
                    <td style={{ padding: '0.75rem' }}>
                      <code style={{ color: 'var(--accent-cyan)' }}>{svc.dns}</code>
                    </td>
                    <td style={{ padding: '0.75rem' }}>
                      <span style={{ color: 'var(--accent-green)', fontFamily: 'var(--font-mono)' }}>{svc.vip}</span>
                    </td>
                    <td style={{ padding: '0.75rem', color: '#fff' }}>{svc.port} / {svc.proto}</td>
                    <td style={{ padding: '0.75rem' }}>
                      <code style={{ fontSize: '0.78rem', color: 'var(--text-secondary)' }}>{svc.spiffe}</code>
                    </td>
                    <td style={{ padding: '0.75rem' }}>
                      <span style={{
                        background: 'rgba(57, 255, 20, 0.15)',
                        color: 'var(--accent-green)',
                        padding: '2px 8px',
                        borderRadius: 4,
                        fontSize: '0.75rem',
                        fontWeight: 600
                      }}>
                        {svc.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* TAB 2: External Services */}
      {meshTab === 'external' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <Globe size={18} color="var(--accent-pink)" /> External Non-Mesh Services (TLS Origination)
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1rem' }}>
            {externalServices.filter(e => e.mesh === selectedMesh).map((ext, idx) => (
              <div key={idx} style={{
                background: 'rgba(10, 15, 30, 0.6)',
                border: '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '1.2rem'
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
                  <h4 style={{ color: '#fff', fontSize: '1rem' }}>{ext.name}</h4>
                  <span style={{
                    background: 'rgba(0, 240, 255, 0.12)',
                    color: 'var(--accent-cyan)',
                    padding: '2px 6px',
                    borderRadius: 4,
                    fontSize: '0.72rem'
                  }}>
                    TLS Origination
                  </span>
                </div>
                <div style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginBottom: '0.3rem' }}>
                  Host: <code style={{ color: 'var(--accent-pink)' }}>{ext.host}:{ext.port}</code>
                </div>
                <div style={{ color: 'var(--text-muted)', fontSize: '0.78rem' }}>SNI: {ext.sni}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 3: Traffic Permissions */}
      {meshTab === 'permissions' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <ShieldCheck size={18} color="var(--accent-green)" /> mTLS Zero-Trust Access Control (Traffic Permissions)
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {permissions.filter(p => p.mesh === selectedMesh).map((perm, idx) => (
              <div key={idx} style={{
                background: 'rgba(10, 15, 30, 0.6)',
                border: '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '1rem 1.25rem',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                flexWrap: 'wrap',
                gap: '0.5rem'
              }}>
                <div>
                  <div style={{ color: '#fff', fontWeight: 600, fontSize: '0.92rem' }}>{perm.name}</div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.3rem', fontSize: '0.82rem' }}>
                    <code style={{ color: 'var(--accent-cyan)' }}>From: {perm.from}</code>
                    <ArrowRight size={14} color="var(--text-muted)" />
                    <code style={{ color: 'var(--accent-green)' }}>To: {perm.to}</code>
                  </div>
                </div>

                <span style={{
                  background: perm.action === 'ALLOW' ? 'rgba(57, 255, 20, 0.15)' : 'rgba(255, 0, 85, 0.15)',
                  color: perm.action === 'ALLOW' ? 'var(--accent-green)' : 'var(--accent-pink)',
                  border: `1px solid ${perm.action === 'ALLOW' ? 'rgba(57, 255, 20, 0.3)' : 'rgba(255, 0, 85, 0.3)'}`,
                  padding: '3px 10px',
                  borderRadius: 4,
                  fontSize: '0.78rem',
                  fontWeight: 700
                }}>
                  {perm.action}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 4: Canary & Routing */}
      {meshTab === 'routing' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <Sliders size={18} color="var(--accent-purple)" /> Dynamic Canary Traffic Splitting
          </h3>
          <div style={{ background: 'rgba(10, 15, 30, 0.6)', border: '1px solid var(--border-glass)', borderRadius: 8, padding: '1.25rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem', fontSize: '0.9rem' }}>
              <span style={{ color: 'var(--accent-cyan)', fontWeight: 600 }}>v1 (Baseline / Stable): 85%</span>
              <span style={{ color: 'var(--accent-pink)', fontWeight: 600 }}>v2 (Canary Release): 15%</span>
            </div>
            <div style={{ height: 10, background: 'rgba(255, 255, 255, 0.1)', borderRadius: 5, overflow: 'hidden', display: 'flex', marginBottom: '1rem' }}>
              <div style={{ width: '85%', background: 'var(--accent-cyan)' }}></div>
              <div style={{ width: '15%', background: 'var(--accent-pink)' }}></div>
            </div>
            <button onClick={() => onToast('Canary weight applied!')} className="btn-primary" style={{ padding: '0.45rem 1rem', fontSize: '0.85rem' }}>
              Save Route Split
            </button>
          </div>
        </div>
      )}

      {/* TAB 5: Passthrough Policies */}
      {meshTab === 'passthrough' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <Zap size={18} color="var(--accent-cyan)" /> Egress Passthrough Policies
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {passthroughPolicies.filter(p => p.mesh === selectedMesh).map((pass, idx) => (
              <div key={idx} style={{
                background: 'rgba(10, 15, 30, 0.6)',
                border: '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '1rem 1.25rem'
              }}>
                <h4 style={{ color: '#fff', fontSize: '0.95rem', marginBottom: '0.4rem' }}>{pass.name}</h4>
                <div style={{ fontSize: '0.82rem', color: 'var(--text-secondary)', marginBottom: '0.2rem' }}>
                  Allowed CIDRs: <code style={{ color: 'var(--accent-green)' }}>{pass.cidrs.join(', ')}</code>
                </div>
                <div style={{ fontSize: '0.82rem', color: 'var(--text-secondary)' }}>
                  Allowed Domains: <code style={{ color: 'var(--accent-cyan)' }}>{pass.hosts.join(', ')}</code>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* TAB 6: Proxy Patches */}
      {meshTab === 'patches' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <FileCode size={18} color="var(--accent-pink)" /> Custom Proxy Patches (HTTP/2 & KeepAlive Tuning)
          </h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {proxyPatches.filter(p => p.mesh === selectedMesh).map((patch, idx) => (
              <div key={idx} style={{
                background: 'rgba(10, 15, 30, 0.6)',
                border: '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '1rem 1.25rem'
              }}>
                <h4 style={{ color: '#fff', fontSize: '0.95rem', marginBottom: '0.4rem' }}>{patch.name}</h4>
                <div style={{ display: 'flex', gap: '1.5rem', fontSize: '0.82rem', color: 'var(--text-secondary)' }}>
                  <span>Target: <code style={{ color: 'var(--accent-cyan)' }}>{patch.target}</code></span>
                  <span>Connect Timeout: <code style={{ color: 'var(--accent-purple)' }}>{patch.connectTimeout}</code></span>
                  <span>HTTP/2 Multiplexing: <span style={{ color: 'var(--accent-green)' }}>Enabled</span></span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
