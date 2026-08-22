import React, { useState } from 'react';
import { 
  GitBranch, 
  Layers, 
  Terminal, 
  Activity, 
  ShieldAlert, 
  Key, 
  Server, 
  RefreshCw, 
  CheckCircle2, 
  Search,
  Filter,
  LucideIcon
} from 'lucide-react';
import { AppTopologyView } from './AppTopologyView';
import { WorkloadManager } from './WorkloadManager';
import { WebTerminal } from './WebTerminal';
import { HubbleVisualizer } from './HubbleVisualizer';
import { RbacMatrix } from './RbacMatrix';
import { SsoSecurity } from './SsoSecurity';

interface Props {
  onToast: (msg: string) => void;
}

export const DashboardLayout: React.FC<Props> = ({ onToast }) => {
  const [activeSubTab, setActiveSubTab] = useState<'topology' | 'workloads' | 'exec' | 'hubble' | 'rbac' | 'sso'>('topology');
  const [namespace, setNamespace] = useState<string>('default');
  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleRefresh = () => {
    setIsRefreshing(true);
    setTimeout(() => {
      setIsRefreshing(false);
      onToast('Cluster state refreshed');
    }, 600);
  };

  interface NavItem {
    id: 'topology' | 'workloads' | 'exec' | 'hubble' | 'rbac' | 'sso';
    label: string;
    icon: LucideIcon;
    badge?: string;
  }

  const navItems: NavItem[] = [
    { id: 'topology', label: 'GitOps Topology (ArgoCD)', icon: GitBranch, badge: 'Live' },
    { id: 'workloads', label: 'Workloads & Resources', icon: Layers },
    { id: 'exec', label: 'Container Exec & Logs', icon: Terminal },
    { id: 'hubble', label: 'Hubble Network Flows', icon: Activity, badge: 'Realtime' },
    { id: 'rbac', label: 'RBAC Policy Matrix', icon: ShieldAlert },
    { id: 'sso', label: 'SSO & Security Center', icon: Key }
  ];

  return (
    <div style={{ maxWidth: 1280, margin: '1rem auto 3rem auto', padding: '0 1rem' }}>
      {/* Cluster Status Top Bar */}
      <div className="glass-card" style={{
        padding: '1rem 1.5rem',
        marginBottom: '1.5rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        flexWrap: 'wrap',
        gap: '1rem'
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', flexWrap: 'wrap' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem' }}>
            <span style={{ width: 10, height: 10, borderRadius: '50%', background: '#22c55e', boxShadow: '0 0 10px #22c55e' }} />
            <span style={{ fontWeight: 700, color: '#fff', fontSize: '1.05rem' }}>tarak-cluster-prod</span>
          </div>

          <span className="badge badge-cyan" style={{ fontSize: '0.78rem' }}>v1.0.6</span>
          <span className="badge badge-emerald" style={{ fontSize: '0.78rem' }}>Zero-Trust mTLS Active</span>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
          {/* Namespace Filter */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', background: 'rgba(255, 255, 255, 0.05)', padding: '0.3rem 0.75rem', borderRadius: 8, border: '1px solid var(--border-glass)' }}>
            <Filter size={14} color="var(--text-muted)" />
            <select
              value={namespace}
              onChange={e => setNamespace(e.target.value)}
              style={{
                background: 'transparent',
                border: 'none',
                color: '#fff',
                outline: 'none',
                fontSize: '0.85rem',
                cursor: 'pointer'
              }}
            >
              <option value="all" style={{ background: '#0f172a' }}>All Namespaces</option>
              <option value="default" style={{ background: '#0f172a' }}>default</option>
              <option value="tarak-system" style={{ background: '#0f172a' }}>tarak-system</option>
              <option value="tarak-public" style={{ background: '#0f172a' }}>tarak-public</option>
            </select>
          </div>

          <button
            onClick={handleRefresh}
            style={{
              background: 'rgba(0, 240, 255, 0.1)',
              border: '1px solid rgba(0, 240, 255, 0.3)',
              color: 'var(--accent-cyan)',
              padding: '0.4rem 0.85rem',
              borderRadius: 8,
              fontSize: '0.82rem',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: 6
            }}
          >
            <RefreshCw size={14} className={isRefreshing ? 'spin-icon' : ''} />
            <span>Sync</span>
          </button>
        </div>
      </div>

      {/* Sub-Navigation Tabs */}
      <div style={{
        display: 'flex',
        gap: '0.5rem',
        overflowX: 'auto',
        paddingBottom: '0.75rem',
        marginBottom: '1.5rem',
        borderBottom: '1px solid var(--border-glass)'
      }}>
        {navItems.map(item => {
          const Icon = item.icon;
          const isActive = activeSubTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => setActiveSubTab(item.id)}
              style={{
                background: isActive ? 'rgba(0, 240, 255, 0.12)' : 'rgba(255, 255, 255, 0.03)',
                color: isActive ? 'var(--accent-cyan)' : 'var(--text-secondary)',
                border: isActive ? '1px solid rgba(0, 240, 255, 0.4)' : '1px solid var(--border-glass)',
                padding: '0.65rem 1.1rem',
                borderRadius: 10,
                fontSize: '0.88rem',
                fontWeight: 600,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: '0.5rem',
                whiteSpace: 'nowrap',
                transition: 'all 0.2s ease'
              }}
            >
              <Icon size={16} />
              <span>{item.label}</span>
              {item.badge && (
                <span style={{
                  background: 'var(--accent-cyan)',
                  color: '#000',
                  fontSize: '0.65rem',
                  fontWeight: 800,
                  padding: '1px 5px',
                  borderRadius: 4,
                  textTransform: 'uppercase'
                }}>
                  {item.badge}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* View Switcher */}
      <div>
        {activeSubTab === 'topology' && <AppTopologyView namespace={namespace} onToast={onToast} />}
        {activeSubTab === 'workloads' && <WorkloadManager namespace={namespace} onToast={onToast} />}
        {activeSubTab === 'exec' && <WebTerminal namespace={namespace} onToast={onToast} />}
        {activeSubTab === 'hubble' && <HubbleVisualizer namespace={namespace} />}
        {activeSubTab === 'rbac' && <RbacMatrix onToast={onToast} />}
        {activeSubTab === 'sso' && <SsoSecurity onToast={onToast} />}
      </div>
    </div>
  );
};
