// ==========================================================================
// TARAK — Modern Documentation & Interactive App Controller
// ==========================================================================

// 0. Client-Side HTTPS Enforcer (Redirects HTTP -> HTTPS in production)
if (window.location.protocol === 'http:' && !window.location.hostname.includes('localhost') && !window.location.hostname.includes('127.0.0.1')) {
  window.location.replace(window.location.href.replace('http:', 'https:'));
}

const REPO = 'vikukumar/tarak';

document.addEventListener('DOMContentLoaded', () => {
  setupMobileDrawer();
  setupTerminalTabs();
  setup3DTiltEffect();
  fetchLatestRelease();
  fetchAllReleases();
});

// 1. Mobile Drawer Navigation Controller
function setupMobileDrawer() {
  const toggleBtn = document.querySelector('.mobile-toggle');
  const drawer = document.querySelector('.mobile-drawer');
  const backdrop = document.querySelector('.drawer-backdrop');

  if (!toggleBtn || !drawer || !backdrop) return;

  function openDrawer() {
    drawer.classList.add('open');
    backdrop.classList.add('active');
    document.body.style.overflow = 'hidden';
  }

  function closeDrawer() {
    drawer.classList.remove('open');
    backdrop.classList.remove('active');
    document.body.style.overflow = '';
  }

  toggleBtn.addEventListener('click', () => {
    if (drawer.classList.contains('open')) {
      closeDrawer();
    } else {
      openDrawer();
    }
  });

  backdrop.addEventListener('click', closeDrawer);

  drawer.querySelectorAll('a').forEach(link => {
    link.addEventListener('click', closeDrawer);
  });
}

// 2. Interactive Terminal Tabs
function setupTerminalTabs() {
  const tabs = document.querySelectorAll('.terminal-tab');
  const codeEl = document.getElementById('install-cmd');

  const commands = {
    bash: 'curl -fsSL https://tarak.vikshro.in/install.sh | bash',
    powershell: 'irm https://tarak.vikshro.in/install.ps1 | iex',
    go: 'go get -u github.com/vikukumar/tarak/pkg/client@latest'
  };

  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      tabs.forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      const type = tab.getAttribute('data-type');
      if (codeEl && commands[type]) {
        codeEl.innerText = commands[type];
      }
    });
  });
}

// 3. Copy to Clipboard with Toast Notification
function copyInstallCommand() {
  const codeEl = document.getElementById('install-cmd');
  if (!codeEl) return;

  navigator.clipboard.writeText(codeEl.innerText).then(() => {
    const btn = document.querySelector('.btn-copy');
    if (btn) {
      const orig = btn.innerText;
      btn.innerText = '✓ Copied!';
      setTimeout(() => btn.innerText = orig, 2000);
    }
    showToast('Command copied to clipboard!');
  }).catch(() => {
    showToast('Failed to copy command');
  });
}

function showToast(message) {
  let container = document.querySelector('.toast-container');
  if (!container) {
    container = document.createElement('div');
    container.className = 'toast-container';
    document.body.appendChild(container);
  }

  const toast = document.createElement('div');
  toast.className = 'toast';
  toast.innerText = message;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateY(10px)';
    setTimeout(() => toast.remove(), 300);
  }, 2500);
}

// 4. Subtle 3D Card Tilt Effect on Desktop
function setup3DTiltEffect() {
  if (window.innerWidth < 900) return;

  const cards = document.querySelectorAll('.card, .terminal-card');
  cards.forEach(card => {
    card.addEventListener('mousemove', e => {
      const rect = card.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      const centerX = rect.width / 2;
      const centerY = rect.height / 2;
      const rotateX = ((y - centerY) / centerY) * -5;
      const rotateY = ((x - centerX) / centerX) * 5;

      card.style.transform = `perspective(1000px) rotateX(${rotateX.toFixed(2)}deg) rotateY(${rotateY.toFixed(2)}deg) translateY(-4px)`;
    });

    card.addEventListener('mouseleave', () => {
      card.style.transform = 'perspective(1000px) rotateX(0) rotateY(0) translateY(0)';
    });
  });
}

// 5. Lightweight Fast Markdown Compiler for Release Changelogs
function compileMarkdown(md) {
  if (!md) return '';

  let html = md
    // Escape angle brackets for code security
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

    // Headers
    .replace(/^### (.*$)/gim, '<h3 style="color: var(--accent-cyan); margin: 1.25rem 0 0.5rem 0; font-size: 1.15rem;">$1</h3>')
    .replace(/^## (.*$)/gim, '<h2 style="color: #fff; margin: 1.5rem 0 0.75rem 0; font-size: 1.35rem;">$1</h2>')
    .replace(/^# (.*$)/gim, '<h1 style="color: #fff; margin: 1.75rem 0 1rem 0; font-size: 1.6rem;">$1</h1>')

    // Code blocks with syntax box
    .replace(/```([a-z]*)\n([\s\S]*?)```/gim, '<div class="code-block">$2</div>')

    // Inline code
    .replace(/`([^`]+)`/gim, '<code>$1</code>')

    // Bold & Italic
    .replace(/\*\*([^*]+)\*\*/gim, '<strong style="color: #f1f5f9;">$1</strong>')
    .replace(/\*([^*]+)\*/gim, '<em>$1</em>')

    // Blockquotes
    .replace(/^\> (.*$)/gim, '<blockquote style="border-left: 3px solid var(--accent-cyan); padding-left: 1rem; color: var(--text-secondary); margin: 0.75rem 0;">$1</blockquote>')

    // Unordered Lists
    .replace(/^\s*-\s+(.*$)/gim, '<li style="margin-left: 1.25rem; margin-bottom: 0.35rem;">$1</li>')
    .replace(/^\s*\*\s+(.*$)/gim, '<li style="margin-left: 1.25rem; margin-bottom: 0.35rem;">$1</li>')

    // Links
    .replace(/\[([^\]]+)\]\(([^)]+)\)/gim, '<a href="$2" target="_blank" style="color: var(--accent-cyan); text-decoration: underline;">$1</a>')

    // Line breaks
    .replace(/\n\n/gim, '<br><br>');

  return `<div class="markdown-body">${html}</div>`;
}

// 6. Fetch Latest Release Version Badge
async function fetchLatestRelease() {
  const badge = document.getElementById('release-badge');
  if (!badge) return;

  try {
    const res = await fetch(`https://api.github.com/repos/${REPO}/releases/latest`);
    if (res.ok) {
      const data = await res.json();
      badge.innerHTML = `⚡ Latest Release: <b>${data.tag_name}</b>`;
      return;
    }
  } catch (e) {}

  try {
    const localRes = await fetch('data/releases.json');
    if (localRes.ok) {
      const list = await localRes.json();
      if (list && list.length > 0) {
        badge.innerHTML = `⚡ Latest Release: <b>${list[0].tag}</b>`;
        return;
      }
    }
  } catch (e) {}

  badge.innerHTML = `⚡ Latest Release: <b>v1.0.1</b>`;
}

// 7. Fetch All Releases & Render Compiled Markdown Changelogs
async function fetchAllReleases() {
  const container = document.getElementById('releases-container');
  if (!container) return;

  let ghReleases = [];
  try {
    const res = await fetch(`https://api.github.com/repos/${REPO}/releases`);
    if (res.ok) {
      ghReleases = await res.json();
    }
  } catch (e) {}

  if (ghReleases && ghReleases.length > 0) {
    container.innerHTML = ghReleases.map(rel => `
      <div class="card" style="margin-bottom: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.25rem; flex-wrap: wrap; gap: 0.75rem;">
          <h2 style="font-size: 1.45rem; font-weight: 700; color: #fff;">⚡ ${rel.name || rel.tag_name}</h2>
          <span class="badge badge-cyan">${new Date(rel.published_at).toLocaleDateString()}</span>
        </div>
        
        <div style="background: rgba(4, 6, 12, 0.65); padding: 1.5rem; border-radius: 12px; border: 1px solid var(--border-glass); margin-bottom: 1.75rem;">
          ${compileMarkdown(rel.body || 'Production release with multi-platform cross-compilation and automated SemVer.')}
        </div>

        <h3 style="font-size: 1.05rem; margin-bottom: 0.85rem; color: #f8fafc;">📦 Multi-Platform Binaries:</h3>
        <div style="display: flex; gap: 0.75rem; flex-wrap: wrap;">
          ${(rel.assets || []).map(a => `
            <a href="${a.browser_download_url}" class="btn-github" style="font-size: 0.85rem; padding: 0.55rem 1rem;" target="_blank">
              💾 ${a.name} <span style="color: var(--text-muted); font-size: 0.78rem;">(${(a.size / 1024 / 1024).toFixed(1)} MB)</span>
            </a>
          `).join('')}
        </div>
      </div>
    `).join('');
    setup3DTiltEffect();
    return;
  }

  // Fallback to static data/releases.json
  let localReleases = [];
  try {
    const res = await fetch('data/releases.json');
    if (res.ok) {
      localReleases = await res.json();
    }
  } catch (e) {}

  if (localReleases && localReleases.length > 0) {
    container.innerHTML = localReleases.map(rel => `
      <div class="card" style="margin-bottom: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.25rem; flex-wrap: wrap; gap: 0.75rem;">
          <h2 style="font-size: 1.45rem; font-weight: 700; color: #fff;">⚡ ${rel.name}</h2>
          <span class="badge badge-cyan">${rel.date}</span>
        </div>

        <h4 style="color: var(--accent-cyan); margin-bottom: 0.75rem; font-size: 1.05rem;">🚀 Release Highlights & Features:</h4>
        <ul style="margin-left: 1.5rem; color: var(--text-secondary); margin-bottom: 1.5rem; line-height: 1.8;">
          ${rel.highlights.map(h => `<li>${h}</li>`).join('')}
        </ul>

        <h4 style="color: var(--accent-purple); margin-bottom: 0.75rem; font-size: 1.05rem;">📦 Included Binaries:</h4>
        <div style="display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 1.5rem;">
          ${rel.binaries.map(b => `<span class="badge badge-purple">${b}</span>`).join('')}
        </div>

        <div style="display: flex; gap: 0.75rem; flex-wrap: wrap;">
          <a href="https://github.com/${REPO}/releases/tag/${rel.tag}" class="btn-primary" target="_blank">
            💾 Download ${rel.tag} Assets on GitHub
          </a>
        </div>
      </div>
    `).join('');
    setup3DTiltEffect();
  }
}
