import React, { useState } from 'react';
import { 
  Menu, 
  X, 
  Star, 
  Home, 
  Rocket, 
  Network, 
  CloudLightning, 
  Cpu, 
  Terminal, 
  Code2, 
  PackageCheck 
} from 'lucide-react';

interface NavbarProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
}

export const Navbar: React.FC<NavbarProps> = ({ activeTab, setActiveTab }) => {
  const [mobileOpen, setMobileOpen] = useState(false);

  const navItems = [
    { id: 'home', label: 'Home', icon: Home },
    { id: 'getting-started', label: 'Getting Started', icon: Rocket },
    { id: 'multi-node', label: 'Multi-Node', icon: Network },
    { id: 'tunnels', label: 'Tunnels & Ingress', icon: CloudLightning },
    { id: 'architecture', label: 'Architecture', icon: Cpu },
    { id: 'api-reference', label: 'API Reference', icon: Code2 },
    { id: 'cli-reference', label: 'CLI Reference', icon: Terminal },
    { id: 'releases', label: 'Releases', icon: PackageCheck }
  ];

  const handleNavClick = (id: string) => {
    setActiveTab(id);
    setMobileOpen(false);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <>
      <header style={{
        position: 'sticky',
        top: 0,
        zIndex: 1000,
        background: 'rgba(7, 9, 14, 0.8)',
        backdropFilter: 'blur(20px)',
        WebkitBackdropFilter: 'blur(20px)',
        borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
        padding: '0.85rem 1.5rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center'
      }}>
        {/* Logo */}
        <div 
          onClick={() => handleNavClick('home')}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.75rem',
            cursor: 'pointer',
            fontWeight: 800,
            fontSize: '1.35rem',
            letterSpacing: '0.5px'
          }}
        >
          <img 
            src="assets/tarak_icon.jpg" 
            alt="Tarak Logo" 
            style={{
              width: 36,
              height: 36,
              borderRadius: 10,
              boxShadow: '0 0 15px rgba(0, 240, 255, 0.35)'
            }} 
          />
          <span className="text-gradient">TARAK</span>
        </div>

        {/* Desktop Navigation */}
        <nav style={{
          display: 'flex',
          gap: '0.4rem',
          alignItems: 'center'
        }} className="desktop-nav-container">
          {navItems.map(item => (
            <button
              key={item.id}
              onClick={() => handleNavClick(item.id)}
              style={{
                background: activeTab === item.id ? 'rgba(0, 240, 255, 0.1)' : 'transparent',
                color: activeTab === item.id ? 'var(--accent-cyan)' : 'var(--text-secondary)',
                border: activeTab === item.id ? '1px solid rgba(0, 240, 255, 0.3)' : '1px solid transparent',
                borderRadius: 8,
                padding: '0.45rem 0.85rem',
                fontSize: '0.9rem',
                fontWeight: activeTab === item.id ? 600 : 500,
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                display: 'flex',
                alignItems: 'center',
                gap: '0.4rem'
              }}
            >
              <item.icon size={15} />
              <span>{item.label}</span>
            </button>
          ))}

          <a
            href="https://github.com/vikukumar/tarak"
            target="_blank"
            rel="noopener noreferrer"
            style={{
              background: 'rgba(255, 255, 255, 0.08)',
              border: '1px solid rgba(255, 255, 255, 0.15)',
              color: '#fff',
              borderRadius: 8,
              padding: '0.45rem 0.95rem',
              fontSize: '0.88rem',
              fontWeight: 600,
              textDecoration: 'none',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              marginLeft: '0.5rem',
              transition: 'all 0.2s ease'
            }}
          >
            <Star size={15} color="#fbbf24" fill="#fbbf24" />
            <span>GitHub</span>
          </a>
        </nav>

        {/* Mobile Hamburger Toggle */}
        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          style={{
            display: 'none',
            background: 'transparent',
            border: '1px solid var(--border-glass)',
            color: '#fff',
            padding: '0.45rem',
            borderRadius: 8,
            cursor: 'pointer'
          }}
          className="mobile-toggle-btn"
          aria-label="Toggle menu"
        >
          {mobileOpen ? <X size={22} /> : <Menu size={22} />}
        </button>
      </header>

      {/* Mobile Backdrop & Drawer */}
      {mobileOpen && (
        <div 
          onClick={() => setMobileOpen(false)}
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0, 0, 0, 0.7)',
            backdropFilter: 'blur(6px)',
            zIndex: 1001
          }} 
        />
      )}

      <aside style={{
        position: 'fixed',
        top: 0,
        right: mobileOpen ? 0 : '-100%',
        width: 290,
        height: '100vh',
        background: 'rgba(11, 17, 33, 0.96)',
        backdropFilter: 'blur(25px)',
        borderLeft: '1px solid var(--border-glass)',
        padding: '5rem 1.25rem 2rem 1.25rem',
        display: 'flex',
        flexDirection: 'column',
        gap: '0.5rem',
        zIndex: 1002,
        transition: 'right 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
        boxShadow: '-10px 0 30px rgba(0,0,0,0.8)'
      }}>
        {navItems.map(item => (
          <button
            key={item.id}
            onClick={() => handleNavClick(item.id)}
            style={{
              background: activeTab === item.id ? 'rgba(0, 240, 255, 0.12)' : 'transparent',
              color: activeTab === item.id ? 'var(--accent-cyan)' : 'var(--text-secondary)',
              border: activeTab === item.id ? '1px solid rgba(0, 240, 255, 0.3)' : '1px solid transparent',
              borderRadius: 8,
              padding: '0.75rem 1rem',
              fontSize: '1rem',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.75rem',
              textAlign: 'left',
              width: '100%'
            }}
          >
            <item.icon size={18} />
            <span>{item.label}</span>
          </button>
        ))}

        <a
          href="https://github.com/vikukumar/tarak"
          target="_blank"
          rel="noopener noreferrer"
          style={{
            marginTop: '1.5rem',
            background: 'var(--grad-cyan-purple)',
            color: '#000',
            fontWeight: 700,
            borderRadius: 8,
            padding: '0.75rem 1rem',
            textDecoration: 'none',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '0.5rem'
          }}
        >
          <Star size={18} />
          <span>Star on GitHub</span>
        </a>
      </aside>

      <style>{`
        @media (max-width: 960px) {
          .desktop-nav-container { display: none !important; }
          .mobile-toggle-btn { display: flex !important; }
        }
      `}</style>
    </>
  );
};
