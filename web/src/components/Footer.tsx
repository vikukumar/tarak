import React from 'react';
import { Heart, Github, Globe } from 'lucide-react';

export const Footer: React.FC = () => {
  return (
    <footer style={{
      borderTop: '1px solid var(--border-glass)',
      background: 'rgba(7, 9, 14, 0.9)',
      backdropFilter: 'blur(20px)',
      padding: '3rem 1.5rem',
      marginTop: '6rem',
      position: 'relative',
      zIndex: 1
    }}>
      <div style={{
        maxWidth: 1200,
        margin: '0 auto',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        flexWrap: 'wrap',
        gap: '1.5rem'
      }}>
        {/* Left branding & Developer info */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
          <img src="assets/tarak_icon.jpg" alt="Tarak" style={{ width: 28, height: 28, borderRadius: 6 }} />
          <span style={{ fontWeight: 700, color: '#fff' }}>TARAK</span>
          <span style={{ color: 'var(--text-muted)' }}>•</span>
          <span style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
            Developed with <Heart size={14} color="#f43f5e" fill="#f43f5e" style={{ display: 'inline', margin: '0 2px' }} /> by{' '}
            <a 
              href="https://github.com/vikukumar" 
              target="_blank" 
              rel="noopener noreferrer"
              style={{ color: 'var(--accent-cyan)', fontWeight: 600, textDecoration: 'none' }}
            >
              @vikukumar
            </a>
          </span>
          <span style={{ color: 'var(--text-muted)' }}>•</span>
          <span style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
            Made in India 🇮🇳
          </span>
        </div>

        {/* Right Links */}
        <div style={{ display: 'flex', gap: '1.5rem', alignItems: 'center' }}>
          <a
            href="https://github.com/vikukumar/tarak"
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: 'var(--text-secondary)', textDecoration: 'none', display: 'flex', alignItems: 'center', gap: 6, fontSize: '0.9rem' }}
          >
            <Github size={16} /> GitHub
          </a>
          <a
            href="https://tarak.vikshro.in"
            style={{ color: 'var(--text-secondary)', textDecoration: 'none', display: 'flex', alignItems: 'center', gap: 6, fontSize: '0.9rem' }}
          >
            <Globe size={16} /> Docs
          </a>
        </div>
      </div>
    </footer>
  );
};
