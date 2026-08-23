import React, { useState, useEffect } from 'react';
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
  Filter,
  RefreshCw
} from 'lucide-react';

interface Props {
  onToast: (msg: string) => void;
}

interface MeshItem {
  name: string;
  mtls?: {
    enabled: boolean;
    mode: string;
    trustDomain: string;
    backend: string;
  };
  passthrough?: string;
  metrics?: string;
  tracing?: string;
  logging?: string;
}

interface MeshService {
  name: string;
  mesh: string;
  namespace: string;
  hostnames: string[];
  virtualVIP: string;
  port: number;
  protocol: string;
  endpoints: string[];
  status: string;
  spiffeId: string;
}

interface MeshExternalService {
  name: string;
  mesh: string;
  host: string;
  port: number;
  tlsRequired: boolean;
  sni?: string;
}

interface MeshTrafficPermission {
  name: string;
  mesh: string;
  from: Array<{ service: string }>;
  to: Array<{ service: string }>;
  action: string;
}

interface MeshPassthroughPolicy {
  name: string;
  mesh: string;
  allowedCIDRs: string[];
  allowedHosts: string[];
}

interface MeshProxyPatch {
  name: string;
  mesh: string;
  target: string;
  connectTimeoutMs?: number;
  idleTimeoutMs?: number;
  http2Enabled: boolean;
}

export const MeshManager: React.FC<Props> = ({ onToast }) => {
  const [selectedMesh, setSelectedMesh] = useState<string>('default');
  const [meshTab, setMeshTab] = useState<'services' | 'external' | 'permissions' | 'routing' | 'passthrough' | 'patches'>('services');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);

  // Live cluster data states
  const [meshes, setMeshes] = useState<MeshItem[]>([]);
  const [services, setServices] = useState<MeshService[]>([]);
  const [externalServices, setExternalServices] = useState<MeshExternalService[]>([]);
  const [permissions, setPermissions] = useState<MeshTrafficPermission[]>([]);
  const [passthroughPolicies, setPassthroughPolicies] = useState<MeshPassthroughPolicy[]>([]);
  const [proxyPatches, setProxyPatches] = useState<MeshProxyPatch[]>([]);

  // Modal state
  const [showCreateModal, setShowCreateModal] = useState<boolean>(false);
  const [newMeshName, setNewMeshName] = useState<string>('');
  const [newMeshMode, setNewMeshMode] = useState<string>('Strict');

  const fetchMeshData = async () => {
    setIsLoading(true);
    try {
      // 1. Fetch Meshes
      const meshesRes = await fetch('/apis/mesh.tarak.io/v1/meshes');
      if (meshesRes.ok) {
        const data = await meshesRes.json();
        setMeshes(data.items || []);
      }

      // 2. Fetch Services
      const svcsRes = await fetch(`/apis/mesh.tarak.io/v1/meshes/${selectedMesh}/services`);
      if (svcsRes.ok) {
        const data = await svcsRes.json();
        setServices(data.items || []);
      }

      // 3. Fetch External Services
      const extRes = await fetch(`/apis/mesh.tarak.io/v1/meshes/${selectedMesh}/external-services`);
      if (extRes.ok) {
        const data = await extRes.json();
        setExternalServices(data.items || []);
      }

      // 4. Fetch Traffic Permissions
      const permRes = await fetch(`/apis/mesh.tarak.io/v1/meshes/${selectedMesh}/traffic-permissions`);
      if (permRes.ok) {
        const data = await permRes.json();
        setPermissions(data.items || []);
      }

      // 5. Fetch Passthrough Policies
      const passRes = await fetch(`/apis/mesh.tarak.io/v1/meshes/${selectedMesh}/passthrough-policies`);
      if (passRes.ok) {
        const data = await passRes.json();
        setPassthroughPolicies(data.items || []);
      }

      // 6. Fetch Proxy Patches
      const patchRes = await fetch(`/apis/mesh.tarak.io/v1/meshes/${selectedMesh}/proxy-patches`);
      if (patchRes.ok) {
        const data = await patchRes.json();
        setProxyPatches(data.items || []);
      }
    } catch {
      // Fallback empty gracefully
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMeshData();
  }, [selectedMesh]);

  const handleCreateMesh = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMeshName.trim()) return;

    try {
      const res = await fetch('/apis/mesh.tarak.io/v1/meshes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: newMeshName.trim().toLowerCase(),
          mtls: {
            enabled: true,
            mode: newMeshMode,
            trustDomain: `${newMeshName.trim().toLowerCase()}.tarak.mesh`,
            backend: 'builtin'
          },
          passthrough: 'Passthrough',
          metrics: 'Prometheus',
          tracing: 'OpenTelemetry',
          logging: 'StructuredJSON'
        })
      });

      if (res.ok) {
        onToast(`Mesh '${newMeshName}' created successfully!`);
        setShowCreateModal(false);
        setSelectedMesh(newMeshName.trim().toLowerCase());
        setNewMeshName('');
        fetchMeshData();
      } else {
        onToast('Failed to create mesh');
      }
    } catch {
      onToast('Error creating mesh');
    }
  };

  const filteredServices = services.filter(s => 
    s.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
    (s.hostnames && s.hostnames.some(h => h.toLowerCase().includes(searchQuery.toLowerCase())))
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
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center', flexWrap: 'wrap' }}>
          <button onClick={fetchMeshData} className="btn-secondary" style={{ padding: '0.45rem 0.75rem', fontSize: '0.85rem' }}>
            <RefreshCw size={14} className={isLoading ? 'spin' : ''} />
          </button>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>Active Mesh:</span>
          <select 
            value={selectedMesh} 
            onChange={e => setSelectedMesh(e.target.value)}
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
            {meshes.length === 0 && <option value="default">default (Strict mTLS)</option>}
            {meshes.map((m, idx) => (
              <option key={idx} value={m.name}>{m.name} ({m.mtls?.mode || 'Strict'})</option>
            ))}
          </select>

          <button onClick={() => setShowCreateModal(true)} className="btn-primary" style={{ padding: '0.45rem 0.9rem', fontSize: '0.85rem' }}>
            <Plus size={14} /> New Mesh
          </button>
        </div>
      </div>

      {/* Create Mesh Modal */}
      {showCreateModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'rgba(0, 0, 0, 0.75)',
          backdropFilter: 'blur(6px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 1000
        }}>
          <div className="glass-card" style={{ width: 440, padding: '2rem', border: '1px solid var(--accent-cyan)' }}>
            <h3 style={{ color: '#fff', fontSize: '1.2rem', marginBottom: '1.25rem' }}>Create New Mesh Tenant</h3>
            <form onSubmit={handleCreateMesh} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              <div>
                <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                  Mesh Name (e.g. finance-mesh, prod-mesh):
                </label>
                <input
                  type="text"
                  required
                  placeholder="mesh-name"
                  value={newMeshName}
                  onChange={e => setNewMeshName(e.target.value)}
                  style={{
                    width: '100%',
                    background: 'rgba(15, 23, 42, 0.8)',
                    border: '1px solid var(--border-glass)',
                    color: '#fff',
                    padding: '0.5rem 0.75rem',
                    borderRadius: 6,
                    fontSize: '0.9rem',
                    outline: 'none'
                  }}
                />
              </div>

              <div>
                <label style={{ display: 'block', color: 'var(--text-secondary)', fontSize: '0.82rem', marginBottom: '0.35rem' }}>
                  mTLS Zero-Trust Mode:
                </label>
                <select
                  value={newMeshMode}
                  onChange={e => setNewMeshMode(e.target.value)}
                  style={{
                    width: '100%',
                    background: 'rgba(15, 23, 42, 0.8)',
                    border: '1px solid var(--border-glass)',
                    color: '#fff',
                    padding: '0.5rem 0.75rem',
                    borderRadius: 6,
                    fontSize: '0.9rem',
                    outline: 'none'
                  }}
                >
                  <option value="Strict">Strict (Default Deny & Mandatory mTLS)</option>
                  <option value="Permissive">Permissive (Transition & Diagnostic)</option>
                </select>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem', marginTop: '1rem' }}>
                <button type="button" onClick={() => setShowCreateModal(false)} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
                  Cancel
                </button>
                <button type="submit" className="btn-primary" style={{ padding: '0.5rem 1.25rem' }}>
                  Create Mesh
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Sub Navigation Tabs */}
      <div style={{ display: 'flex', gap: '0.5rem', borderBottom: '1px solid var(--border-glass)', paddingBottom: '0.5rem', overflowX: 'auto' }}>
        <button onClick={() => setMeshTab('services')} className={`tab-btn ${meshTab === 'services' ? 'active' : ''}`}>
          <Server size={15} /> Mesh Services ({services.length})
        </button>
        <button onClick={() => setMeshTab('external')} className={`tab-btn ${meshTab === 'external' ? 'active' : ''}`}>
          <Globe size={15} /> External Services ({externalServices.length})
        </button>
        <button onClick={() => setMeshTab('permissions')} className={`tab-btn ${meshTab === 'permissions' ? 'active' : ''}`}>
          <ShieldCheck size={15} /> Traffic Permissions ({permissions.length})
        </button>
        <button onClick={() => setMeshTab('routing')} className={`tab-btn ${meshTab === 'routing' ? 'active' : ''}`}>
          <Sliders size={15} /> Canary & Routing
        </button>
        <button onClick={() => setMeshTab('passthrough')} className={`tab-btn ${meshTab === 'passthrough' ? 'active' : ''}`}>
          <Zap size={15} /> Passthrough Policies ({passthroughPolicies.length})
        </button>
        <button onClick={() => setMeshTab('patches')} className={`tab-btn ${meshTab === 'patches' ? 'active' : ''}`}>
          <FileCode size={15} /> Proxy Patches ({proxyPatches.length})
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

          {filteredServices.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <Network size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>No workloads enrolled in mesh '{selectedMesh}'</p>
              <p style={{ fontSize: '0.82rem', marginTop: '0.25rem' }}>
                Add the annotation <code>tarak.io/mesh: {selectedMesh}</code> to any Pod or Namespace to automatically enroll it into the mesh.
              </p>
            </div>
          ) : (
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
                        <code style={{ color: 'var(--accent-cyan)' }}>{svc.hostnames?.[0] || `${svc.name}.${svc.namespace}.mesh`}</code>
                      </td>
                      <td style={{ padding: '0.75rem' }}>
                        <span style={{ color: 'var(--accent-green)', fontFamily: 'var(--font-mono)' }}>{svc.virtualVIP}</span>
                      </td>
                      <td style={{ padding: '0.75rem', color: '#fff' }}>{svc.port} / {svc.protocol}</td>
                      <td style={{ padding: '0.75rem' }}>
                        <code style={{ fontSize: '0.78rem', color: 'var(--text-secondary)' }}>{svc.spiffeId}</code>
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
                          {svc.status || 'Healthy'}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* TAB 2: External Services */}
      {meshTab === 'external' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <Globe size={18} color="var(--accent-pink)" /> External Non-Mesh Services (TLS Origination)
          </h3>
          {externalServices.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <Globe size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>No external dependencies configured</p>
              <p style={{ fontSize: '0.82rem', marginTop: '0.25rem' }}>
                Define <code>MeshExternalService</code> to automatically originate TLS for outbound dependencies like Stripe, AWS, and external databases.
              </p>
            </div>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1rem' }}>
              {externalServices.map((ext, idx) => (
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
                  {ext.sni && <div style={{ color: 'var(--text-muted)', fontSize: '0.78rem' }}>SNI: {ext.sni}</div>}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* TAB 3: Traffic Permissions */}
      {meshTab === 'permissions' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <ShieldCheck size={18} color="var(--accent-green)" /> mTLS Zero-Trust Access Control (Traffic Permissions)
          </h3>
          {permissions.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <ShieldCheck size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>Default Deny Microsegmentation Active</p>
              <p style={{ fontSize: '0.82rem', marginTop: '0.25rem' }}>
                Create <code>MeshTrafficPermission</code> rules to explicitly authorize communication between service pairs.
              </p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              {permissions.map((perm, idx) => (
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
                      <code style={{ color: 'var(--accent-cyan)' }}>From: {perm.from?.[0]?.service || '*'}</code>
                      <ArrowRight size={14} color="var(--text-muted)" />
                      <code style={{ color: 'var(--accent-green)' }}>To: {perm.to?.[0]?.service || '*'}</code>
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
          )}
        </div>
      )}

      {/* TAB 4: Canary & Routing */}
      {meshTab === 'routing' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <Sliders size={18} color="var(--accent-purple)" /> Dynamic Canary Traffic Splitting
          </h3>
          <div style={{ background: 'rgba(10, 15, 30, 0.6)', border: '1px solid var(--border-glass)', borderRadius: 8, padding: '1.25rem' }}>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginBottom: '1rem' }}>
              Create <code>MeshTrafficRoute</code> to split live traffic percentage between release subsets with zero downtime.
            </p>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem', fontSize: '0.9rem' }}>
              <span style={{ color: 'var(--accent-cyan)', fontWeight: 600 }}>v1 (Baseline / Stable): 90%</span>
              <span style={{ color: 'var(--accent-pink)', fontWeight: 600 }}>v2 (Canary Release): 10%</span>
            </div>
            <div style={{ height: 10, background: 'rgba(255, 255, 255, 0.1)', borderRadius: 5, overflow: 'hidden', display: 'flex', marginBottom: '1rem' }}>
              <div style={{ width: '90%', background: 'var(--accent-cyan)' }}></div>
              <div style={{ width: '10%', background: 'var(--accent-pink)' }}></div>
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
          {passthroughPolicies.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <Zap size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>Default Passthrough Mode</p>
              <p style={{ fontSize: '0.82rem', marginTop: '0.25rem' }}>
                When the mesh is set to <code>DenyAll</code>, declare <code>MeshPassthroughPolicy</code> resources to allow specific CIDRs or wildcard domains.
              </p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              {passthroughPolicies.map((pass, idx) => (
                <div key={idx} style={{
                  background: 'rgba(10, 15, 30, 0.6)',
                  border: '1px solid var(--border-glass)',
                  borderRadius: 8,
                  padding: '1rem 1.25rem'
                }}>
                  <h4 style={{ color: '#fff', fontSize: '0.95rem', marginBottom: '0.4rem' }}>{pass.name}</h4>
                  <div style={{ fontSize: '0.82rem', color: 'var(--text-secondary)', marginBottom: '0.2rem' }}>
                    Allowed CIDRs: <code style={{ color: 'var(--accent-green)' }}>{pass.allowedCIDRs?.join(', ')}</code>
                  </div>
                  <div style={{ fontSize: '0.82rem', color: 'var(--text-secondary)' }}>
                    Allowed Domains: <code style={{ color: 'var(--accent-cyan)' }}>{pass.allowedHosts?.join(', ')}</code>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* TAB 6: Proxy Patches */}
      {meshTab === 'patches' && (
        <div className="glass-card" style={{ padding: '1.5rem' }}>
          <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
            <FileCode size={18} color="var(--accent-pink)" /> Custom Proxy Patches (HTTP/2 & KeepAlive Tuning)
          </h3>
          {proxyPatches.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <FileCode size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>Default High-Performance Proxy Settings Active</p>
              <p style={{ fontSize: '0.82rem', marginTop: '0.25rem' }}>
                Use <code>MeshProxyPatch</code> to customize timeouts, buffer sizes, and HTTP/2 connection reuse per service.
              </p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              {proxyPatches.map((patch, idx) => (
                <div key={idx} style={{
                  background: 'rgba(10, 15, 30, 0.6)',
                  border: '1px solid var(--border-glass)',
                  borderRadius: 8,
                  padding: '1rem 1.25rem'
                }}>
                  <h4 style={{ color: '#fff', fontSize: '0.95rem', marginBottom: '0.4rem' }}>{patch.name}</h4>
                  <div style={{ display: 'flex', gap: '1.5rem', fontSize: '0.82rem', color: 'var(--text-secondary)' }}>
                    <span>Target: <code style={{ color: 'var(--accent-cyan)' }}>{patch.target}</code></span>
                    {patch.connectTimeoutMs && <span>Connect Timeout: <code style={{ color: 'var(--accent-purple)' }}>{patch.connectTimeoutMs}ms</code></span>}
                    <span>HTTP/2 Multiplexing: <span style={{ color: patch.http2Enabled ? 'var(--accent-green)' : 'var(--text-muted)' }}>{patch.http2Enabled ? 'Enabled' : 'Disabled'}</span></span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};
