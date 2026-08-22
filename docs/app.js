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
    bash: 'curl -fsSL https://raw.githubusercontent.com/vikukumar/tarak/main/install.sh | bash',
    powershell: 'irm https://raw.githubusercontent.com/vikukumar/tarak/main/install.ps1 | iex',
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
    } else {
      badge.innerHTML = `⚡ Latest Release: <b>v1.0.0</b>`;
    }
  } catch (e) {
    badge.innerHTML = `⚡ Production Ready: <b>v1.0.0</b>`;
  }
}

// Fetch All Releases and Render in Releases Page
async function fetchAllReleases() {
  const container = document.getElementById('releases-container');
  if (!container) return;

  container.innerHTML = '<p style="color: var(--text-secondary);">Loading releases from GitHub...</p>';

  try {
    const res = await fetch(`https://api.github.com/repos/${REPO}/releases`);
    if (!res.ok) throw new Error('Failed to fetch');
    const releases = await res.json();

    if (releases.length === 0) {
      container.innerHTML = '<p style="color: var(--text-secondary);">No releases published yet.</p>';
      return;
    }

    container.innerHTML = releases.map(rel => `
      <div class="card" style="margin-bottom: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
          <h2>${rel.name || rel.tag_name}</h2>
          <span class="badge-win">${new Date(rel.published_at).toLocaleDateString()}</span>
        </div>
        <div style="background: rgba(0,0,0,0.3); padding: 1rem; border-radius: 8px; margin-bottom: 1.5rem; white-space: pre-wrap; font-family: var(--font-mono); font-size: 0.85rem; color: var(--text-secondary);">
          ${escapeHTML(rel.body || 'No release notes provided.')}
        </div>
        <h3>📦 Download Binaries:</h3>
        <div style="display: flex; gap: 0.75rem; flex-wrap: wrap; margin-top: 0.75rem;">
          ${(rel.assets || []).map(a => `
            <a href="${a.browser_download_url}" class="btn-github" style="font-size: 0.8rem;" target="_blank">
              💾 ${a.name} (${(a.size / 1024 / 1024).toFixed(1)} MB)
            </a>
          `).join('')}
        </div>
      </div>
    `).join('');
  } catch (e) {
    container.innerHTML = `
      <div class="card">
        <h3>Release v1.0.0 (Initial Release)</h3>
        <p style="color: var(--text-secondary); margin: 0.5rem 0 1rem;">Initial production release with native OCI container engine, Virtual Pod Bridge, and pure-Go control plane.</p>
        <div style="display: flex; gap: 0.75rem; flex-wrap: wrap;">
          <a href="https://github.com/${REPO}/releases" class="btn-github" target="_blank">View GitHub Releases</a>
        </div>
      </div>
    `;
  }
}

function escapeHTML(str) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
