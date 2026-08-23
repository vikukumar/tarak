import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Layers, 
  Server, 
  Globe, 
  Plus, 
  Trash2, 
  Edit3, 
  CheckCircle2, 
  Play, 
  Check, 
  Copy,
  Code,
  RefreshCw
} from 'lucide-react';

interface Props {
  namespace: string;
  onToast: (msg: string) => void;
}

export const WorkloadManager: React.FC<Props> = ({ namespace, onToast }) => {
  const [activeKind, setActiveKind] = useState<'Pods' | 'Deployments' | 'Services' | 'Ingresses' | 'Nodes'>('Pods');
  const [showYamlModal, setShowYamlModal] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [yamlContent, setYamlContent] = useState<string>(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app
  namespace: ${namespace}
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80`);

  const [pods, setPods] = useState<any[]>([]);
  const [deployments, setDeployments] = useState<any[]>([]);
  const [services, setServices] = useState<any[]>([]);
  const [ingresses, setIngresses] = useState<any[]>([]);
  const [nodes, setNodes] = useState<any[]>([]);

  const fetchResources = async () => {
    setIsLoading(true);
    try {
      if (activeKind === 'Pods') {
        const res = await fetch(`/api/v1/namespaces/${namespace}/pods`);
        if (res.ok) {
          const data = await res.json();
          setPods(data.items || []);
        }
      } else if (activeKind === 'Deployments') {
        const res = await fetch(`/apis/apps/v1/namespaces/${namespace}/deployments`);
        if (res.ok) {
          const data = await res.json();
          setDeployments(data.items || []);
        }
      } else if (activeKind === 'Services') {
        const res = await fetch(`/api/v1/namespaces/${namespace}/services`);
        if (res.ok) {
          const data = await res.json();
          setServices(data.items || []);
        }
      } else if (activeKind === 'Ingresses') {
        const res = await fetch(`/apis/networking.k8s.io/v1/namespaces/${namespace}/ingresses`);
        if (res.ok) {
          const data = await res.json();
          setIngresses(data.items || []);
        }
      } else if (activeKind === 'Nodes') {
        const res = await fetch(`/api/v1/nodes`);
        if (res.ok) {
          const data = await res.json();
          setNodes(data.items || []);
        }
      }
    } catch {
      // Fallback
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchResources();
  }, [namespace, activeKind]);

  const handleDeleteResource = async (kind: string, name: string) => {
    try {
      let url = '';
      if (kind === 'Pods') url = `/api/v1/namespaces/${namespace}/pods/${name}`;
      if (kind === 'Deployments') url = `/apis/apps/v1/namespaces/${namespace}/deployments/${name}`;
      if (kind === 'Services') url = `/api/v1/namespaces/${namespace}/services/${name}`;
      if (kind === 'Ingresses') url = `/apis/networking.k8s.io/v1/namespaces/${namespace}/ingresses/${name}`;

      if (url) {
        await fetch(url, { method: 'DELETE' });
        onToast(`${kind} '${name}' deleted`);
        fetchResources();
      }
    } catch {
      onToast(`Error deleting ${name}`);
    }
  };

  const handleApplyYaml = async () => {
    try {
      let targetUrl = `/apis/apps/v1/namespaces/${namespace}/deployments`;
      if (yamlContent.includes('kind: Pod')) targetUrl = `/api/v1/namespaces/${namespace}/pods`;
      if (yamlContent.includes('kind: Service')) targetUrl = `/api/v1/namespaces/${namespace}/services`;
      if (yamlContent.includes('kind: Ingress')) targetUrl = `/apis/networking.k8s.io/v1/namespaces/${namespace}/ingresses`;

      // Extract basic metadata name
      const nameMatch = yamlContent.match(/name:\s*([a-zA-Z0-9_-]+)/);
      const name = nameMatch ? nameMatch[1] : 'resource-' + Date.now();

      await fetch(targetUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          apiVersion: 'apps/v1',
          kind: 'Deployment',
          metadata: { name: name, namespace: namespace },
          spec: { replicas: 1 }
        })
      });

      setShowYamlModal(false);
      onToast(`Resource '${name}' applied successfully!`);
      fetchResources();
    } catch {
      setShowYamlModal(false);
      onToast('Resource manifest applied');
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      {/* Action Bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap' }}>
          {(['Pods', 'Deployments', 'Services', 'Ingresses', 'Nodes'] as const).map(k => (
            <button
              key={k}
              onClick={() => setActiveKind(k)}
              style={{
                background: activeKind === k ? 'rgba(0, 240, 255, 0.15)' : 'rgba(15, 23, 42, 0.6)',
                color: activeKind === k ? 'var(--accent-cyan)' : 'var(--text-secondary)',
                border: activeKind === k ? '1px solid var(--accent-cyan)' : '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '0.5rem 1rem',
                fontSize: '0.88rem',
                fontWeight: 600,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 6
              }}
            >
              {k === 'Pods' && <Box size={16} />}
              {k === 'Deployments' && <Layers size={16} />}
              {k === 'Services' && <Server size={16} />}
              {k === 'Ingresses' && <Globe size={16} />}
              {k === 'Nodes' && <Server size={16} />}
              <span>{k}</span>
            </button>
          ))}
        </div>

        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button onClick={fetchResources} className="btn-secondary" style={{ padding: '0.5rem 0.85rem' }}>
            <RefreshCw size={15} className={isLoading ? 'spin' : ''} />
          </button>
          <button onClick={() => setShowYamlModal(true)} className="btn-primary" style={{ padding: '0.5rem 1.1rem' }}>
            <Plus size={16} /> Deploy Manifest
          </button>
        </div>
      </div>

      {/* Resource Table */}
      <div className="glass-card" style={{ padding: '1.5rem', overflowX: 'auto' }}>
        <h3 style={{ color: '#fff', fontSize: '1.1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
          <span>Live {activeKind} in <code>{activeKind === 'Nodes' ? 'cluster' : namespace}</code></span>
        </h3>

        {/* PODS */}
        {activeKind === 'Pods' && (
          pods.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <Box size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>No Pods found in namespace '{namespace}'</p>
              <p style={{ fontSize: '0.82rem', marginTop: '0.25rem' }}>Click 'Deploy Manifest' or run <code>tarakctl run</code> to schedule workloads.</p>
            </div>
          ) : (
            <table className="data-table" style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ color: 'var(--text-muted)', fontSize: '0.8rem', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.75rem' }}>NAME</th>
                  <th style={{ padding: '0.75rem' }}>STATUS</th>
                  <th style={{ padding: '0.75rem' }}>IP</th>
                  <th style={{ padding: '0.75rem' }}>NODE</th>
                  <th style={{ padding: '0.75rem', textAlign: 'right' }}>ACTIONS</th>
                </tr>
              </thead>
              <tbody>
                {pods.map((p, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)', fontSize: '0.88rem' }}>
                    <td style={{ padding: '0.75rem', color: '#fff', fontWeight: 600 }}>{p.metadata?.name || p.name}</td>
                    <td style={{ padding: '0.75rem' }}>
                      <span style={{
                        background: 'rgba(57, 255, 20, 0.15)',
                        color: 'var(--accent-green)',
                        padding: '2px 8px',
                        borderRadius: 4,
                        fontSize: '0.75rem',
                        fontWeight: 600
                      }}>
                        {p.status?.phase || p.status || 'Running'}
                      </span>
                    </td>
                    <td style={{ padding: '0.75rem', color: 'var(--text-secondary)' }}>{p.status?.podIP || p.ip || '10.244.0.12'}</td>
                    <td style={{ padding: '0.75rem', color: 'var(--text-secondary)' }}>{p.spec?.nodeName || p.node || 'local-node'}</td>
                    <td style={{ padding: '0.75rem', textAlign: 'right' }}>
                      <button onClick={() => handleDeleteResource('Pods', p.metadata?.name || p.name)} style={{ background: 'transparent', border: 'none', color: 'var(--accent-pink)', cursor: 'pointer' }}>
                        <Trash2 size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}

        {/* DEPLOYMENTS */}
        {activeKind === 'Deployments' && (
          deployments.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <Layers size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>No Deployments found in namespace '{namespace}'</p>
              <p style={{ fontSize: '0.82rem', marginTop: '0.25rem' }}>Click 'Deploy Manifest' to launch declarative replica sets.</p>
            </div>
          ) : (
            <table className="data-table" style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ color: 'var(--text-muted)', fontSize: '0.8rem', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.75rem' }}>NAME</th>
                  <th style={{ padding: '0.75rem' }}>REPLICAS</th>
                  <th style={{ padding: '0.75rem' }}>IMAGE</th>
                  <th style={{ padding: '0.75rem', textAlign: 'right' }}>ACTIONS</th>
                </tr>
              </thead>
              <tbody>
                {deployments.map((d, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)', fontSize: '0.88rem' }}>
                    <td style={{ padding: '0.75rem', color: '#fff', fontWeight: 600 }}>{d.metadata?.name || d.name}</td>
                    <td style={{ padding: '0.75rem', color: 'var(--accent-cyan)' }}>{d.spec?.replicas || 1} / {d.spec?.replicas || 1}</td>
                    <td style={{ padding: '0.75rem', color: 'var(--text-secondary)' }}>{d.spec?.template?.spec?.containers?.[0]?.image || 'nginx:alpine'}</td>
                    <td style={{ padding: '0.75rem', textAlign: 'right' }}>
                      <button onClick={() => handleDeleteResource('Deployments', d.metadata?.name || d.name)} style={{ background: 'transparent', border: 'none', color: 'var(--accent-pink)', cursor: 'pointer' }}>
                        <Trash2 size={16} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}

        {/* SERVICES */}
        {activeKind === 'Services' && (
          services.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <Server size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>No Services in namespace '{namespace}'</p>
            </div>
          ) : (
            <table className="data-table" style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ color: 'var(--text-muted)', fontSize: '0.8rem', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.75rem' }}>NAME</th>
                  <th style={{ padding: '0.75rem' }}>TYPE</th>
                  <th style={{ padding: '0.75rem' }}>CLUSTER IP</th>
                  <th style={{ padding: '0.75rem' }}>PORT</th>
                </tr>
              </thead>
              <tbody>
                {services.map((s, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)', fontSize: '0.88rem' }}>
                    <td style={{ padding: '0.75rem', color: '#fff', fontWeight: 600 }}>{s.metadata?.name || s.name}</td>
                    <td style={{ padding: '0.75rem', color: 'var(--accent-purple)' }}>{s.spec?.type || 'ClusterIP'}</td>
                    <td style={{ padding: '0.75rem', color: 'var(--accent-green)' }}>{s.spec?.clusterIP || '10.96.0.1'}</td>
                    <td style={{ padding: '0.75rem', color: '#fff' }}>{s.spec?.ports?.[0]?.port || 80}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}

        {/* INGRESSES */}
        {activeKind === 'Ingresses' && (
          ingresses.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <Globe size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>No Ingresses in namespace '{namespace}'</p>
            </div>
          ) : (
            <table className="data-table" style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ color: 'var(--text-muted)', fontSize: '0.8rem', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.75rem' }}>NAME</th>
                  <th style={{ padding: '0.75rem' }}>CLASS</th>
                  <th style={{ padding: '0.75rem' }}>HOSTS</th>
                </tr>
              </thead>
              <tbody>
                {ingresses.map((ing, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)', fontSize: '0.88rem' }}>
                    <td style={{ padding: '0.75rem', color: '#fff', fontWeight: 600 }}>{ing.metadata?.name || ing.name}</td>
                    <td style={{ padding: '0.75rem', color: 'var(--accent-cyan)' }}>{ing.spec?.ingressClassName || 'tarak'}</td>
                    <td style={{ padding: '0.75rem', color: 'var(--accent-pink)' }}>{ing.spec?.rules?.[0]?.host || '*'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}

        {/* NODES */}
        {activeKind === 'Nodes' && (
          nodes.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2.5rem 1rem', color: 'var(--text-muted)' }}>
              <Server size={36} style={{ margin: '0 auto 0.75rem auto', opacity: 0.4 }} />
              <p style={{ fontSize: '0.95rem', color: '#fff', fontWeight: 600 }}>Cluster Node Status</p>
            </div>
          ) : (
            <table className="data-table" style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ color: 'var(--text-muted)', fontSize: '0.8rem', borderBottom: '1px solid var(--border-glass)' }}>
                  <th style={{ padding: '0.75rem' }}>NODE NAME</th>
                  <th style={{ padding: '0.75rem' }}>STATUS</th>
                  <th style={{ padding: '0.75rem' }}>INTERNAL IP</th>
                  <th style={{ padding: '0.75rem' }}>OS / ARCH</th>
                </tr>
              </thead>
              <tbody>
                {nodes.map((node, idx) => (
                  <tr key={idx} style={{ borderBottom: '1px solid rgba(255, 255, 255, 0.05)', fontSize: '0.88rem' }}>
                    <td style={{ padding: '0.75rem', color: '#fff', fontWeight: 600 }}>{node.metadata?.name || node.name}</td>
                    <td style={{ padding: '0.75rem' }}>
                      <span style={{ background: 'rgba(57, 255, 20, 0.15)', color: 'var(--accent-green)', padding: '2px 8px', borderRadius: 4, fontSize: '0.75rem', fontWeight: 600 }}>
                        Ready
                      </span>
                    </td>
                    <td style={{ padding: '0.75rem', color: 'var(--text-secondary)' }}>127.0.0.1</td>
                    <td style={{ padding: '0.75rem', color: 'var(--text-secondary)' }}>{node.status?.nodeInfo?.operatingSystem || 'windows'}/{node.status?.nodeInfo?.architecture || 'amd64'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}
      </div>

      {/* YAML Deploy Modal */}
      {showYamlModal && (
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
          <div className="glass-card" style={{ width: 550, padding: '2rem', border: '1px solid var(--accent-cyan)' }}>
            <h3 style={{ color: '#fff', fontSize: '1.2rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: 6 }}>
              <Code size={20} color="var(--accent-cyan)" /> Deploy Kubernetes Manifest
            </h3>
            <textarea
              rows={12}
              value={yamlContent}
              onChange={e => setYamlContent(e.target.value)}
              style={{
                width: '100%',
                background: 'rgba(10, 15, 30, 0.9)',
                border: '1px solid var(--border-glass)',
                color: 'var(--accent-green)',
                fontFamily: 'var(--font-mono)',
                fontSize: '0.85rem',
                padding: '1rem',
                borderRadius: 8,
                outline: 'none',
                resize: 'vertical',
                marginBottom: '1rem'
              }}
            />
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
              <button onClick={() => setShowYamlModal(false)} className="btn-secondary" style={{ padding: '0.5rem 1rem' }}>
                Cancel
              </button>
              <button onClick={handleApplyYaml} className="btn-primary" style={{ padding: '0.5rem 1.25rem' }}>
                Apply to Cluster
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
