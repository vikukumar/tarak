// Tarak Interactive Documentation & Releases App
const REPO = 'vikukumar/tarak';

document.addEventListener('DOMContentLoaded', () => {
  setupTerminalTabs();
  fetchLatestRelease();
  fetchAllReleases();
});

// Terminal Tabs switcher
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

// Copy to Clipboard
function copyInstallCommand() {
  const codeEl = document.getElementById('install-cmd');
  if (!codeEl) return;
  navigator.clipboard.writeText(codeEl.innerText).then(() => {
    const btn = document.querySelector('.btn-copy');
    if (btn) {
      const orig = btn.innerText;
      btn.innerText = 'Copied!';
      setTimeout(() => btn.innerText = orig, 2000);
    }
  });
}

// Fetch Latest Release Version for Badge
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
  } catch (e) {
    // fallback to local data
  }

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

// Fetch All Releases & Render Version History
async function fetchAllReleases() {
  const container = document.getElementById('releases-container');
  if (!container) return;

  let localReleases = [];
  try {
    const res = await fetch('data/releases.json');
    if (res.ok) {
      localReleases = await res.json();
    }
  } catch (e) {}

  let ghReleases = [];
  try {
    const res = await fetch(`https://api.github.com/repos/${REPO}/releases`);
    if (res.ok) {
      ghReleases = await res.json();
    }
  } catch (e) {}

  // If we have GitHub releases, render them with live assets
  if (ghReleases && ghReleases.length > 0) {
    container.innerHTML = ghReleases.map(rel => `
      <div class="card" style="margin-bottom: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; flex-wrap: wrap; gap: 0.5rem;">
          <h2>⚡ ${rel.name || rel.tag_name}</h2>
          <span class="badge-win">${new Date(rel.published_at).toLocaleDateString()}</span>
        </div>
        <div style="background: rgba(0,0,0,0.35); padding: 1.25rem; border-radius: 8px; margin-bottom: 1.5rem; white-space: pre-wrap; font-family: var(--font-mono); font-size: 0.88rem; color: var(--text-secondary);">
          ${escapeHTML(rel.body || 'Production release with multi-platform cross-compilation and automated SemVer.')}
        </div>
        <h3 style="margin-bottom: 0.75rem;">📦 Ready-to-Run Binary Archives:</h3>
        <div style="display: flex; gap: 0.75rem; flex-wrap: wrap;">
          ${(rel.assets || []).map(a => `
            <a href="${a.browser_download_url}" class="btn-github" style="font-size: 0.85rem;" target="_blank">
              💾 ${a.name} (${(a.size / 1024 / 1024).toFixed(1)} MB)
            </a>
          `).join('')}
        </div>
      </div>
    `).join('');
    return;
  }

  // Fallback to static releases.json history
  if (localReleases && localReleases.length > 0) {
    container.innerHTML = localReleases.map(rel => `
      <div class="card" style="margin-bottom: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; flex-wrap: wrap; gap: 0.5rem;">
          <h2>⚡ ${rel.name}</h2>
          <span class="badge-win">${rel.date}</span>
        </div>
        <h4 style="color: var(--accent-cyan); margin-bottom: 0.5rem;">🚀 Release Highlights & Features:</h4>
        <ul style="margin-left: 1.5rem; color: var(--text-secondary); margin-bottom: 1.5rem; line-height: 1.8;">
          ${rel.highlights.map(h => `<li>${h}</li>`).join('')}
        </ul>
        <h4 style="color: var(--accent-purple); margin-bottom: 0.5rem;">📦 Included Binaries:</h4>
        <ul style="margin-left: 1.5rem; color: var(--text-secondary); margin-bottom: 1.5rem;">
          ${rel.binaries.map(b => `<li><code>${b}</code></li>`).join('')}
        </ul>
        <div style="display: flex; gap: 0.75rem; flex-wrap: wrap;">
          <a href="https://github.com/${REPO}/releases/tag/${rel.tag}" class="btn-github" target="_blank">
            💾 Download ${rel.tag} Assets on GitHub
          </a>
        </div>
      </div>
    `).join('');
    return;
  }

  container.innerHTML = '<p style="color: var(--text-secondary);">No release history found.</p>';
}

function escapeHTML(str) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
