import React from 'react';
import { ShieldAlert, AlertTriangle, FileQuestion, ServerCrash, ArrowLeft, RefreshCw, Lock, Key } from 'lucide-react';

interface ErrorProps {
  code: 401 | 403 | 404 | 500;
  message?: string;
  onAction?: () => void;
  actionText?: string;
}

export const ErrorPage: React.FC<ErrorProps> = ({ code, message, onAction, actionText }) => {
  const getDetails = () => {
    switch (code) {
      case 401:
        return {
          title: '401 - Authentication Required',
          subtitle: 'Your session has expired or no valid cryptographic token was supplied.',
          icon: Lock,
          color: 'var(--accent-pink)',
          bg: 'rgba(255, 0, 85, 0.12)',
          defaultAction: 'Login to Cluster'
        };
      case 403:
        return {
          title: '403 - RBAC Permission Forbidden',
          subtitle: 'Your identity lacks the required RoleBinding permissions to access this resource.',
          icon: ShieldAlert,
          color: 'var(--accent-purple)',
          bg: 'rgba(180, 0, 255, 0.12)',
          defaultAction: 'Switch Account / Token'
        };
      case 404:
        return {
          title: '404 - Cluster Resource Not Found',
          subtitle: 'The requested resource, namespace, or API endpoint does not exist.',
          icon: FileQuestion,
          color: 'var(--accent-cyan)',
          bg: 'rgba(0, 240, 255, 0.12)',
          defaultAction: 'Return to Dashboard'
        };
      case 500:
      default:
        return {
          title: '500 - Internal Daemon Error',
          subtitle: 'The cluster control plane encountered an unexpected internal exception.',
          icon: ServerCrash,
          color: 'var(--accent-pink)',
          bg: 'rgba(255, 0, 85, 0.15)',
          defaultAction: 'Retry Connection'
        };
    }
  };

  const details = getDetails();
  const IconComponent = details.icon;

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '60vh',
      padding: '2rem 1rem'
    }}>
      <div className="glass-card" style={{
        maxWidth: 540,
        width: '100%',
        padding: '3rem 2rem',
        textAlign: 'center',
        border: `1px solid ${details.color}`,
        boxShadow: `0 0 30px ${details.bg}`
      }}>
        <div style={{
          width: 80,
          height: 80,
          borderRadius: 20,
          background: details.bg,
          border: `1px solid ${details.color}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          margin: '0 auto 1.5rem auto'
        }}>
          <IconComponent size={40} color={details.color} />
        </div>

        <h2 style={{ color: '#fff', fontSize: '1.5rem', fontWeight: 800, marginBottom: '0.6rem' }}>
          {details.title}
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.92rem', lineHeight: 1.5, marginBottom: '1.75rem' }}>
          {message || details.subtitle}
        </p>

        <div style={{ display: 'flex', justifyContent: 'center', gap: '0.75rem' }}>
          {onAction && (
            <button onClick={onAction} className="btn-primary" style={{ padding: '0.6rem 1.5rem', fontSize: '0.9rem', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Key size={16} />
              <span>{actionText || details.defaultAction}</span>
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
