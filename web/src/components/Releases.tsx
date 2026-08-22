import React, { useEffect, useState } from 'react';
import { PackageCheck, Download, Calendar, Tag, ExternalLink } from 'lucide-react';
import { marked } from 'marked';

interface ReleaseAsset {
  name: string;
  browser_download_url: string;
  size: number;
}

interface ReleaseItem {
  id?: number;
  name: string;
  tag_name?: string;
  tag?: string;
  published_at?: string;
  date?: string;
  body?: string;
  highlights?: string[];
  binaries?: string[];
  assets?: ReleaseAsset[];
}

export const Releases: React.FC = () => {
  const [releases, setReleases] = useState<ReleaseItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('https://api.github.com/repos/vikukumar/tarak/releases')
      .then(res => res.json())
      .then(data => {
        if (Array.isArray(data) && data.length > 0) {
          setReleases(data);
          setLoading(false);
          return;
        }
        throw new Error('No GitHub releases found');
      })
      .catch(() => {
        // Fallback to local static data
        fetch('data/releases.json')
          .then(res => res.json())
          .then(localData => {
            if (Array.isArray(localData)) {
              setReleases(localData);
            }
          })
          .catch(() => {})
          .finally(() => setLoading(false));
      });
  }, []);

  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: '2rem 1rem' }}>
      <div style={{ textAlign: 'center', marginBottom: '3rem' }}>
        <span className="badge badge-cyan" style={{ marginBottom: '1rem' }}>
          <PackageCheck size={14} /> Version History
        </span>
        <h1 style={{ fontSize: 'clamp(2.2rem, 4vw, 3rem)', fontWeight: 800, color: '#fff' }}>
          Releases & <span className="text-gradient">Compiled Changelog</span>
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '1.1rem', marginTop: '0.5rem' }}>
          Complete version history with live compiled markdown notes and cross-platform binary downloads.
        </p>
      </div>

      {loading ? (
        <div className="glass-card" style={{ padding: '3rem', textAlign: 'center' }}>
          <p style={{ color: 'var(--text-secondary)' }}>Loading latest releases...</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
          {releases.map((rel, idx) => {
            const tagName = rel.tag_name || rel.tag || 'v1.0.0';
            const title = rel.name || tagName;
            const pubDate = rel.published_at ? new Date(rel.published_at).toLocaleDateString() : rel.date || 'Recent';
            const bodyHtml = rel.body ? marked.parse(rel.body) as string : '';

            return (
              <div key={idx} className="glass-card" style={{ padding: '2rem' }}>
                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  flexWrap: 'wrap',
                  gap: '0.75rem',
                  marginBottom: '1.5rem',
                  borderBottom: '1px solid var(--border-glass)',
                  paddingBottom: '1rem'
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                    <Tag size={20} color="var(--accent-cyan)" />
                    <h2 style={{ fontSize: '1.45rem', fontWeight: 700, color: '#fff' }}>{title}</h2>
                  </div>
                  <span className="badge badge-cyan" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                    <Calendar size={14} /> {pubDate}
                  </span>
                </div>

                {/* Compiled Markdown Body or Highlights */}
                {bodyHtml ? (
                  <div 
                    className="markdown-body" 
                    dangerouslySetInnerHTML={{ __html: bodyHtml }}
                    style={{
                      background: 'rgba(4, 6, 12, 0.6)',
                      padding: '1.5rem',
                      borderRadius: 12,
                      border: '1px solid var(--border-glass)',
                      marginBottom: '1.5rem'
                    }}
                  />
                ) : rel.highlights ? (
                  <div style={{ marginBottom: '1.5rem' }}>
                    <h4 style={{ color: 'var(--accent-cyan)', marginBottom: '0.5rem', fontSize: '1rem' }}>🚀 Release Highlights:</h4>
                    <ul style={{ marginLeft: '1.5rem', color: 'var(--text-secondary)', lineHeight: 1.8 }}>
                      {rel.highlights.map((h, hIdx) => (
                        <li key={hIdx}>{h}</li>
                      ))}
                    </ul>
                  </div>
                ) : null}

                {/* Assets / Binary Downloads */}
                {rel.assets && rel.assets.length > 0 ? (
                  <div>
                    <h4 style={{ color: '#fff', fontSize: '1rem', marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: 6 }}>
                      <Download size={16} /> Multi-Platform Binaries:
                    </h4>
                    <div style={{ display: 'flex', gap: '0.6rem', flexWrap: 'wrap' }}>
                      {rel.assets.map((asset, aIdx) => (
                        <a
                          key={aIdx}
                          href={asset.browser_download_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="btn-secondary"
                          style={{ fontSize: '0.82rem', padding: '0.5rem 0.9rem' }}
                        >
                          <span>💾 {asset.name}</span>
                          <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>
                            ({(asset.size / 1024 / 1024).toFixed(1)} MB)
                          </span>
                        </a>
                      ))}
                    </div>
                  </div>
                ) : (
                  <a
                    href={`https://github.com/vikukumar/tarak/releases/tag/${tagName}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="btn-primary"
                    style={{ display: 'inline-flex', fontSize: '0.88rem', padding: '0.6rem 1.2rem' }}
                  >
                    <ExternalLink size={15} />
                    <span>Download {tagName} Assets on GitHub</span>
                  </a>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
