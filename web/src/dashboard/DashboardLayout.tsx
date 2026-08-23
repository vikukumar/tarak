import React, { useState, useEffect } from 'react';
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
  LogOut,
  User,
  Shield,
  LucideIcon
} from 'lucide-react';
import { AppTopologyView } from './AppTopologyView';
import { WorkloadManager } from './WorkloadManager';
import { WebTerminal } from './WebTerminal';
import { HubbleVisualizer } from './HubbleVisualizer';
import { RbacMatrix } from './RbacMatrix';
import { SsoSecurity } from './SsoSecurity';
import { MeshManager } from './MeshManager';
import { AuthPortal } from './AuthPortal';
import { ErrorPage } from './ErrorPages';

// Setup global fetch interceptor to attach Bearer token
if (typeof window !== 'undefined') {
  const origFetch = window.fetch;
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const token = localStorage.getItem('tarak_token');
    const headers = new Headers(init?.headers || {});
    if (token && typeof input === 'string' && input.startsWith('/')) {
      if (!headers.has('Authorization')) {
        headers.set('Authorization', `Bearer ${token}`);
      }
    }
    return origFetch(input, { ...init, headers });
  };
}

interface Props {
  onToast: (msg: string) => void;
}

export const DashboardLayout: React.FC<Props> = ({ onToast }) => {
  const [authToken, setAuthToken] = useState<string | null>(() => {
    return typeof window !== 'undefined' ? localStorage.getItem('tarak_token') : null;
  });
  const [currentUser, setCurrentUser] = useState<any>(() => {
    return { username: 'admin', roles: ['cluster-admin', 'system:masters'] };
  });
  const [activeSubTab, setActiveSubTab] = useState<'topology' | 'workloads' | 'exec' | 'mesh' | 'hubble' | 'rbac' | 'sso'>('topology');
  const [namespace, setNamespace] = useState<string>('default');
  const [namespaces, setNamespaces] = useState<string[]>(['default', 'tarak-system', 'tarak-public']);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [currentErrorCode, setCurrentErrorCode] = useState<401 | 403 | 404 | 500 | null>(null);

  const fetchNamespaces = async () => {
    try {
      const res = await fetch('/api/v1/namespaces');
      if (res.ok) {
        const data = await res.json();
        if (data.items && Array.isArray(data.items)) {
          const names = data.items.map((it: any) => it.metadata?.name).filter(Boolean);
          if (names.length > 0) {
            setNamespaces(names);
          }
        }
      }
    } catch {}
  };

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const pathParts = window.location.pathname.split('/').filter(Boolean);
      if (pathParts.length > 1 && ['topology', 'workloads', 'exec', 'mesh', 'hubble', 'rbac', 'sso'].includes(pathParts[1])) {
        setActiveSubTab(pathParts[1] as any);
      }
    }
    fetchNamespaces();
  }, []);

  const handleSubTabChange = (tabId: 'topology' | 'workloads' | 'exec' | 'mesh' | 'hubble' | 'rbac' | 'sso') => {
    setActiveSubTab(tabId);
    if (typeof window !== 'undefined') {
      window.history.pushState({}, '', `/dashboard/${tabId}`);
    }
  };

  const handleRefresh = () => {
    setIsRefreshing(true);
    fetchNamespaces();
    setTimeout(() => {
      setIsRefreshing(false);
      onToast('Cluster state synced');
    }, 600);
  };

  const handleLogout = () => {
    localStorage.removeItem('tarak_token');
    setAuthToken(null);
    onToast('Logged out of cluster');
  };

  // If unauthenticated, show First-Time Setup Wizard or Login Portal
  if (!authToken) {
    return (
      <AuthPortal
        onAuthenticated={(user, token) => {
          setAuthToken(token);
          setCurrentUser(user);
        }}
        onToast={onToast}
      />
    );
  }

  // If a custom error state is active, render the dedicated error page
  if (currentErrorCode) {
    return (
      <div style={{ maxWidth: 1280, margin: '2rem auto', padding: '0 1rem' }}>
        <ErrorPage
          code={currentErrorCode}
          onAction={() => setCurrentErrorCode(null)}
          actionText="Return to Cluster Dashboard"
        />
      </div>
    );
  }

  interface NavItem {
    id: 'topology' | 'workloads' | 'exec' | 'mesh' | 'hubble' | 'rbac' | 'sso';
    label: string;
    icon: LucideIcon;
    badge?: string;
  }

  const navItems: NavItem[] = [
    { id: 'topology', label: 'GitOps Topology (ArgoCD)', icon: GitBranch, badge: 'Live' },
    { id: 'workloads', label: 'Workloads & Resources', icon: Layers },
    { id: 'exec', label: 'Container Exec & Logs', icon: Terminal },
    { id: 'mesh', label: 'Service Mesh (Kuma)', icon: Server, badge: 'Multi-Mesh' },
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
          {/* Dynamic Namespace Filter */}
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
              {namespaces.map(ns => (
                <option key={ns} value={ns} style={{ background: '#0f172a', color: '#fff' }}>
                  {ns}
                </option>
              ))}
            </select>
          </div>

          {/* Sync Button */}
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

          {/* Super-Admin User Pill & Logout */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            background: 'rgba(15, 23, 42, 0.8)',
            border: '1px solid var(--border-glass)',
            padding: '0.25rem 0.75rem',
            borderRadius: 8
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
              <Shield size={14} color="var(--accent-green)" />
              <span style={{ color: '#fff', fontSize: '0.85rem', fontWeight: 600 }}>
                {currentUser?.username || 'Super-Admin'}
              </span>
            </div>
            <button
              onClick={handleLogout}
              title="Logout"
              style={{
                background: 'transparent',
                border: 'none',
                color: 'var(--text-muted)',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                padding: '2px'
              }}
            >
              <LogOut size={14} />
            </button>
          </div>
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
              onClick={() => handleSubTabChange(item.id)}
              style={{
                background: isActive ? 'rgba(0, 240, 255, 0.12)' : 'rgba(15, 23, 42, 0.4)',
                color: isActive ? 'var(--accent-cyan)' : 'var(--text-secondary)',
                border: isActive ? '1px solid rgba(0, 240, 255, 0.4)' : '1px solid var(--border-glass)',
                borderRadius: 8,
                padding: '0.55rem 1rem',
                fontSize: '0.88rem',
                fontWeight: 600,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                whiteSpace: 'nowrap',
                transition: 'all 0.2s ease'
              }}
            >
              <Icon size={16} />
              <span>{item.label}</span>
              {item.badge && (
                <span style={{
                  background: item.badge === 'Realtime' ? 'var(--accent-pink)' : (item.badge === 'Multi-Mesh' ? 'var(--accent-purple)' : 'var(--accent-green)'),
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
        {activeSubTab === 'mesh' && <MeshManager onToast={onToast} />}
        {activeSubTab === 'hubble' && <HubbleVisualizer namespace={namespace} />}
        {activeSubTab === 'rbac' && <RbacMatrix onToast={onToast} />}
        {activeSubTab === 'sso' && <SsoSecurity onToast={onToast} />}
      </div>
    </div>
  );
};
