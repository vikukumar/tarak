import React, { useState, useEffect } from 'react';
import { Background3D } from './components/Background3D';
import { Navbar } from './components/Navbar';
import { Hero } from './components/Hero';
import { Benchmarks } from './components/Benchmarks';
import { GettingStarted } from './components/GettingStarted';
import { MultiNode } from './components/MultiNode';
import { Tunnels } from './components/Tunnels';
import { Architecture } from './components/Architecture';
import { ApiReference } from './components/ApiReference';
import { CliReference } from './components/CliReference';
import { Releases } from './components/Releases';
import { DashboardLayout } from './dashboard/DashboardLayout';
import { Footer } from './components/Footer';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string>('home');
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [isDashboardPath, setIsDashboardPath] = useState<boolean>(() => {
    if (typeof window === 'undefined') return false;
    return window.location.pathname.startsWith('/dashboard') || window.location.hash === '#dashboard';
  });

  useEffect(() => {
    const handleLocationChange = () => {
      const isDash = window.location.pathname.startsWith('/dashboard') || window.location.hash === '#dashboard';
      setIsDashboardPath(isDash);
      if (!isDash) {
        const hash = window.location.hash.replace('#', '');
        setActiveTab(hash || 'home');
      }
    };

    handleLocationChange();
    window.addEventListener('popstate', handleLocationChange);
    window.addEventListener('hashchange', handleLocationChange);
    return () => {
      window.removeEventListener('popstate', handleLocationChange);
      window.removeEventListener('hashchange', handleLocationChange);
    };
  }, []);

  const handleTabChange = (tab: string) => {
    if (tab === 'dashboard') {
      window.history.pushState({}, '', '/dashboard');
      setIsDashboardPath(true);
      return;
    }
    setIsDashboardPath(false);
    setActiveTab(tab);
    window.location.hash = tab === 'home' ? '' : tab;
  };

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 2500);
  };

  // Pure Standalone Cluster Dashboard (Zero Documentation UI)
  if (isDashboardPath) {
    return (
      <div style={{
        minHeight: '100vh',
        background: '#070c18',
        color: '#f8fafc',
        fontFamily: 'var(--font-sans)',
        position: 'relative'
      }}>
        <Background3D />
        <div style={{ position: 'relative', zIndex: 10, padding: '1rem 0' }}>
          <DashboardLayout onToast={showToast} />
        </div>

        {/* Floating Toast Alert */}
        {toastMessage && (
          <div style={{
            position: 'fixed',
            bottom: '2rem',
            right: '2rem',
            zIndex: 2000,
            background: 'rgba(15, 23, 42, 0.95)',
            border: '1px solid var(--accent-cyan)',
            color: '#fff',
            padding: '0.75rem 1.25rem',
            borderRadius: 8,
            fontSize: '0.88rem',
            fontWeight: 600,
            boxShadow: '0 10px 30px rgba(0, 0, 0, 0.5)',
            animation: 'fadeIn 0.2s ease-out'
          }}>
            ✓ {toastMessage}
          </div>
        )}
      </div>
    );
  }

  // Documentation Portal
  return (
    <div style={{ position: 'relative', minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Background3D />
      <Navbar activeTab={activeTab} setActiveTab={handleTabChange} />

      <main style={{ flex: 1, position: 'relative', zIndex: 1 }}>
        {activeTab === 'home' && (
          <>
            <Hero onExplore={handleTabChange} onToast={showToast} />
            <Benchmarks />
          </>
        )}
        {activeTab === 'getting-started' && <GettingStarted onToast={showToast} />}
        {activeTab === 'multi-node' && <MultiNode onToast={showToast} />}
        {activeTab === 'tunnels' && <Tunnels onToast={showToast} />}
        {activeTab === 'architecture' && <Architecture />}
        {activeTab === 'api-reference' && <ApiReference />}
        {activeTab === 'cli-reference' && <CliReference onToast={showToast} />}
        {activeTab === 'releases' && <Releases />}
      </main>

      <Footer />

      {/* Floating Toast Alert */}
      {toastMessage && (
        <div style={{
          position: 'fixed',
          bottom: '2rem',
          right: '2rem',
          zIndex: 2000,
          background: 'rgba(15, 23, 42, 0.95)',
          border: '1px solid var(--accent-cyan)',
          color: '#fff',
          padding: '0.75rem 1.25rem',
          borderRadius: 8,
          fontSize: '0.88rem',
          fontWeight: 600,
          boxShadow: '0 10px 30px rgba(0, 0, 0, 0.5)',
          animation: 'fadeIn 0.2s ease-out'
        }}>
          ✓ {toastMessage}
        </div>
      )}
    </div>
  );
};
