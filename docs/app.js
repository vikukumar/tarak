// ==========================================================================
// TARAK — Modern Documentation & Interactive Cluster Simulator
// ==========================================================================

// 0. Client-Side HTTPS Enforcer (Redirects HTTP -> HTTPS in production)
if (window.location.protocol === 'http:' && !window.location.hostname.includes('localhost') && !window.location.hostname.includes('127.0.0.1')) {
  window.location.replace(window.location.href.replace('http:', 'https:'));
}

const REPO = 'vikukumar/tarak';
let latestVersion = 'v1.0.6';
let allReleasesData = [];

document.addEventListener('DOMContentLoaded', () => {
  setupClusterCanvas();
  setupMobileDrawer();
  setupTerminalTabs();
  setup3DTiltEffect();
  syncDynamicReleases();
});

// ==========================================================================
// 1. Interactive Cluster Topology & Packet Stream Background Simulation
// ==========================================================================
function setupClusterCanvas() {
  let canvas = document.getElementById('cluster-canvas');
  if (!canvas) {
    canvas = document.createElement('canvas');
    canvas.id = 'cluster-canvas';
    document.body.prepend(canvas);
  }

  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  let width = (canvas.width = window.innerWidth);
  let height = (canvas.height = window.innerHeight);

  let mouse = { x: width / 2, y: height / 2, active: false };

  window.addEventListener('resize', () => {
    width = canvas.width = window.innerWidth;
    height = canvas.height = window.innerHeight;
    initNodes();
  });

  window.addEventListener('mousemove', (e) => {
    mouse.x = e.clientX;
    mouse.y = e.clientY;
    mouse.active = true;
  });

  window.addEventListener('mouseleave', () => {
    mouse.active = false;
  });

  // Cluster Nodes & Topology definition
  let nodes = [];
  let packets = [];

  function initNodes() {
    nodes = [];
    packets = [];

    const isMobile = width < 768;
    const nodeCount = isMobile ? 6 : 14;

    // Control Plane Leader
    nodes.push({
      x: width * 0.5,
      y: height * 0.35,
      vx: (Math.random() - 0.5) * 0.4,
      vy: (Math.random() - 0.5) * 0.4,
      radius: 9,
      type: 'control-plane',
      label: 'tarak-control-plane',
      color: '#00f0ff',
      pulse: 0,
      pods: 4,
    });

    // Worker & Edge Nodes
    for (let i = 1; i < nodeCount; i++) {
      const type = i % 4 === 0 ? 'edge-gateway' : 'worker-node';
      const color = type === 'edge-gateway' ? '#a855f7' : '#38bdf8';
      nodes.push({
        x: Math.random() * width,
        y: Math.random() * height,
        vx: (Math.random() - 0.5) * 0.6,
        vy: (Math.random() - 0.5) * 0.6,
        radius: type === 'edge-gateway' ? 7.5 : 6,
        type: type,
        label: `worker-${i}`,
        color: color,
        pulse: Math.random() * Math.PI * 2,
        pods: Math.floor(Math.random() * 3) + 2,
      });
    }

    // Spawn packets traveling across pod bridges
    const packetCount = isMobile ? 8 : 22;
    for (let i = 0; i < packetCount; i++) {
      spawnPacket();
    }
  }

  function spawnPacket() {
    if (nodes.length < 2) return;
    const fromIdx = Math.floor(Math.random() * nodes.length);
    let toIdx = Math.floor(Math.random() * nodes.length);
    while (toIdx === fromIdx) {
      toIdx = Math.floor(Math.random() * nodes.length);
    }
    packets.push({
      from: fromIdx,
      to: toIdx,
      progress: Math.random(),
      speed: 0.003 + Math.random() * 0.006,
      color: Math.random() > 0.4 ? '#00f0ff' : '#a855f7',
      size: 2.2,
    });
  }

  initNodes();

  let animFrame;
  function animate() {
    ctx.clearRect(0, 0, width, height);

    // 1. Draw subtle background coordinate grid
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.018)';
    ctx.lineWidth = 1;
    const gridSize = 80;
    for (let x = 0; x < width; x += gridSize) {
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, height);
      ctx.stroke();
    }
    for (let y = 0; y < height; y += gridSize) {
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(width, y);
      ctx.stroke();
    }

    // 2. Update and draw nodes
    nodes.forEach((n, idx) => {
      n.x += n.vx;
      n.y += n.vy;
      n.pulse += 0.04;

      // Bounce on edges
      if (n.x < 50 || n.x > width - 50) n.vx *= -1;
      if (n.y < 50 || n.y > height - 50) n.vy *= -1;

      // Mouse influence
      if (mouse.active) {
        const dx = mouse.x - n.x;
        const dy = mouse.y - n.y;
        const dist = Math.sqrt(dx * dx + dy * dy);
        if (dist < 180 && dist > 1) {
          const force = (180 - dist) / 180;
          n.x -= (dx / dist) * force * 1.5;
          n.y -= (dy / dist) * force * 1.5;
        }
      }

      // Draw Virtual Pod Interconnect lines between nearby nodes
      for (let j = idx + 1; j < nodes.length; j++) {
        const other = nodes[j];
        const dx = other.x - n.x;
        const dy = other.y - n.y;
        const dist = Math.sqrt(dx * dx + dy * dy);

        const maxDist = width < 768 ? 160 : 260;
        if (dist < maxDist) {
          const alpha = (1 - dist / maxDist) * 0.22;
          ctx.strokeStyle = `rgba(0, 240, 255, ${alpha})`;
          ctx.lineWidth = 1.2;
          ctx.beginPath();
          ctx.moveTo(n.x, n.y);
          ctx.lineTo(other.x, other.y);
          ctx.stroke();
        }
      }

      // Draw Node Outer Glow Halo
      const haloSize = n.radius + 6 + Math.sin(n.pulse) * 3;
      ctx.beginPath();
      ctx.arc(n.x, n.y, haloSize, 0, Math.PI * 2);
      ctx.fillStyle = n.type === 'control-plane' ? 'rgba(0, 240, 255, 0.08)' : 'rgba(168, 85, 247, 0.06)';
      ctx.fill();

      // Draw Node Core
      ctx.beginPath();
      ctx.arc(n.x, n.y, n.radius, 0, Math.PI * 2);
      ctx.fillStyle = n.color;
      ctx.shadowColor = n.color;
      ctx.shadowBlur = 12;
      ctx.fill();
      ctx.shadowBlur = 0; // reset

      // Orbiting Pod Dots around each node
      const podOrbitRadius = n.radius + 10;
      for (let p = 0; p < n.pods; p++) {
        const podAngle = n.pulse * 0.8 + (p * (Math.PI * 2)) / n.pods;
        const px = n.x + Math.cos(podAngle) * podOrbitRadius;
        const py = n.y + Math.sin(podAngle) * podOrbitRadius;

        ctx.beginPath();
        ctx.arc(px, py, 1.8, 0, Math.PI * 2);
        ctx.fillStyle = p === 0 ? '#10b981' : p === 1 ? '#00f0ff' : '#a855f7';
        ctx.fill();
      }
    });

    // 3. Update and draw encrypted mTLS data packets
    packets.forEach((pkt, idx) => {
      pkt.progress += pkt.speed;
      if (pkt.progress >= 1) {
        // Reset packet with new path
        pkt.from = Math.floor(Math.random() * nodes.length);
        pkt.to = Math.floor(Math.random() * nodes.length);
        while (pkt.to === pkt.from) {
          pkt.to = Math.floor(Math.random() * nodes.length);
        }
        pkt.progress = 0;
      }

      const fromNode = nodes[pkt.from];
      const toNode = nodes[pkt.to];
      if (!fromNode || !toNode) return;

      const px = fromNode.x + (toNode.x - fromNode.x) * pkt.progress;
      const py = fromNode.y + (toNode.y - fromNode.y) * pkt.progress;

      ctx.beginPath();
      ctx.arc(px, py, pkt.size, 0, Math.PI * 2);
      ctx.fillStyle = pkt.color;
      ctx.shadowColor = pkt.color;
      ctx.shadowBlur = 8;
      ctx.fill();
      ctx.shadowBlur = 0;
    });

    animFrame = requestAnimationFrame(animate);
  }

  animate();
}

// ==========================================================================
// 2. Mobile Navigation Drawer
// ==========================================================================
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

  drawer.querySelectorAll('a').forEach((link) => {
    link.addEventListener('click', closeDrawer);
  });
}

// ==========================================================================
// 3. Interactive Terminal Tabs
// ==========================================================================
function setupTerminalTabs() {
  const tabs = document.querySelectorAll('.terminal-tab');
  const codeEl = document.getElementById('install-cmd');

  const commands = {
    bash: 'curl -fsSL https://tarak.vikshro.in/install.sh | bash',
    powershell: 'irm https://tarak.vikshro.in/install.ps1 | iex',
    go: 'go get -u github.com/vikukumar/tarak/pkg/client@latest',
  };

  tabs.forEach((tab) => {
    tab.addEventListener('click', () => {
      tabs.forEach((t) => t.classList.remove('active'));
      tab.classList.add('active');
      const type = tab.getAttribute('data-type');
      if (codeEl && commands[type]) {
        codeEl.innerText = commands[type];
      }
    });
  });
}

// ==========================================================================
// 4. Clipboard Copy with Toast
// ==========================================================================
function copyInstallCommand() {
  const codeEl = document.getElementById('install-cmd');
  if (!codeEl) return;

  navigator.clipboard
    .writeText(codeEl.innerText)
    .then(() => {
      const btn = document.querySelector('.btn-copy');
      if (btn) {
        const orig = btn.innerText;
        btn.innerText = '✓ Copied!';
        setTimeout(() => (btn.innerText = orig), 2000);
      }
      showToast('Command copied to clipboard!');
    })
    .catch(() => {
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

// ==========================================================================
// 5. 3D Card Hover Tilt Effect
// ==========================================================================
function setup3DTiltEffect() {
  if (window.innerWidth < 900) return;

  const cards = document.querySelectorAll('.card, .terminal-card');
  cards.forEach((card) => {
    card.addEventListener('mousemove', (e) => {
      const rect = card.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;
      const centerX = rect.width / 2;
      const centerY = rect.height / 2;
      const rotateX = ((y - centerY) / centerY) * -4;
      const rotateY = ((x - centerX) / centerX) * 4;

      card.style.transform = `perspective(1000px) rotateX(${rotateX.toFixed(2)}deg) rotateY(${rotateY.toFixed(2)}deg) translateY(-4px)`;
    });

    card.addEventListener('mouseleave', () => {
      card.style.transform = 'perspective(1000px) rotateX(0) rotateY(0) translateY(0)';
    });
  });
}

// ==========================================================================
// 6. Dynamic Version & Releases Synchronization
// ==========================================================================
async function syncDynamicReleases() {
  // First load local data/releases.json for immediate instant display
  try {
    const res = await fetch('data/releases.json');
    if (res.ok) {
      allReleasesData = await res.json();
      if (allReleasesData && allReleasesData.length > 0) {
        latestVersion = allReleasesData[0].tag;
        applyDynamicVersionUI(allReleasesData[0], allReleasesData);
      }
    }
  } catch (e) {}

  // Then query GitHub API for live real-time sync if online
  try {
    const [latestRes, listRes] = await Promise.allSettled([
      fetch(`https://api.github.com/repos/${REPO}/releases/latest`),
      fetch(`https://api.github.com/repos/${REPO}/releases`),
    ]);

    if (latestRes.status === 'fulfilled' && latestRes.value.ok) {
      const ghLatest = await latestRes.value.json();
      latestVersion = ghLatest.tag_name;
      applyDynamicVersionUI(
        {
          tag: ghLatest.tag_name,
          version: ghLatest.tag_name.replace(/^v/, ''),
          name: ghLatest.name || `Tarak ${ghLatest.tag_name}`,
          date: new Date(ghLatest.published_at).toISOString().split('T')[0],
          status: 'Production Ready',
        },
        allReleasesData
      );
    }

    if (listRes.status === 'fulfilled' && listRes.value.ok) {
      const ghList = await listRes.value.json();
      if (ghList && ghList.length > 0) {
        renderReleasesChangelog(ghList, true);
        return;
      }
    }
  } catch (e) {}

  // Fallback render using local data
  if (allReleasesData && allReleasesData.length > 0) {
    renderReleasesChangelog(allReleasesData, false);
  }
}

function applyDynamicVersionUI(latestObj, allReleases) {
  const versionTag = latestObj.tag || 'v1.0.6';
  const cleanVer = versionTag.replace(/^v/, '');

  // 1. Update Hero Badge & Elements
  const heroBadge = document.getElementById('release-badge');
  if (heroBadge) {
    heroBadge.innerHTML = `⚡ <b>${versionTag}</b> Production Ready`;
  }

  document.querySelectorAll('[data-latest-version]').forEach((el) => {
    el.innerText = versionTag;
  });

  document.querySelectorAll('[data-latest-clean-version]').forEach((el) => {
    el.innerText = cleanVer;
  });

  document.querySelectorAll('[data-release-date]').forEach((el) => {
    el.innerText = latestObj.date || '2026-08-23';
  });

  // Update dynamic download buttons
  document.querySelectorAll('a[data-latest-download]').forEach((a) => {
    a.href = `https://github.com/${REPO}/releases/tag/${versionTag}`;
  });
}

// ==========================================================================
// 7. Render Releases & Compiled Markdown Changelog
// ==========================================================================
function renderReleasesChangelog(releases, isGitHubApi) {
  const container = document.getElementById('releases-container');
  if (!container) return;

  if (isGitHubApi) {
    container.innerHTML = releases
      .map(
        (rel) => `
      <div class="card" style="margin-bottom: 2rem;">
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.25rem; flex-wrap: wrap; gap: 0.75rem;">
          <h2 style="font-size: 1.45rem; font-weight: 700; color: #fff;">⚡ ${rel.name || rel.tag_name}</h2>
          <div style="display: flex; gap: 0.5rem; align-items: center;">
            <span class="badge badge-cyan">${rel.tag_name}</span>
            <span class="badge badge-purple">${new Date(rel.published_at).toLocaleDateString()}</span>
          </div>
        </div>
        
        <div style="background: rgba(4, 6, 12, 0.65); padding: 1.5rem; border-radius: 12px; border: 1px solid var(--border-glass); margin-bottom: 1.75rem;">
          ${compileMarkdown(rel.body || 'Production release with multi-platform cross-compilation and automated SemVer.')}
        </div>

        <h3 style="font-size: 1.05rem; margin-bottom: 0.85rem; color: #f8fafc;">📦 Multi-Platform Release Binaries:</h3>
        <div style="display: flex; gap: 0.75rem; flex-wrap: wrap;">
          ${(rel.assets || [])
            .map(
              (a) => `
            <a href="${a.browser_download_url}" class="btn-github" style="font-size: 0.85rem; padding: 0.55rem 1rem;" target="_blank">
              💾 ${a.name} <span style="color: var(--text-muted); font-size: 0.78rem;">(${(a.size / 1024 / 1024).toFixed(1)} MB)</span>
            </a>
          `
            )
            .join('')}
          <a href="${rel.html_url}" class="btn-primary" style="font-size: 0.85rem; padding: 0.55rem 1.2rem;" target="_blank">
            ⭐ View on GitHub
          </a>
        </div>
      </div>
    `
      )
      .join('');
    setup3DTiltEffect();
    return;
  }

  // Local data/releases.json fallback
  container.innerHTML = releases
    .map(
      (rel) => `
    <div class="card" style="margin-bottom: 2rem;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.25rem; flex-wrap: wrap; gap: 0.75rem;">
        <h2 style="font-size: 1.45rem; font-weight: 700; color: #fff;">⚡ ${rel.name}</h2>
        <div style="display: flex; gap: 0.5rem; align-items: center;">
          <span class="badge ${rel.isLatest ? 'badge-cyan' : 'badge-purple'}">${rel.tag}</span>
          <span class="badge badge-emerald">${rel.status || 'Stable'}</span>
          <span class="badge" style="background: rgba(255,255,255,0.06); color: var(--text-muted);">${rel.date}</span>
        </div>
      </div>

      <h4 style="color: var(--accent-cyan); margin-bottom: 0.75rem; font-size: 1.05rem;">🚀 Release Highlights & Features:</h4>
      <ul style="margin-left: 1.5rem; color: var(--text-secondary); margin-bottom: 1.5rem; line-height: 1.8;">
        ${rel.highlights.map((h) => `<li>${h}</li>`).join('')}
      </ul>

      <h4 style="color: var(--accent-purple); margin-bottom: 0.75rem; font-size: 1.05rem;">📦 Included Multi-Platform Binaries:</h4>
      <div style="display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 1.5rem;">
        ${rel.binaries.map((b) => `<span class="badge badge-purple">${b}</span>`).join('')}
      </div>

      <div style="display: flex; gap: 0.75rem; flex-wrap: wrap;">
        <a href="${rel.downloadUrl || `https://github.com/${REPO}/releases/tag/${rel.tag}`}" class="btn-primary" target="_blank">
          💾 Download ${rel.tag} Assets on GitHub
        </a>
      </div>
    </div>
  `
    )
    .join('');
  setup3DTiltEffect();
}

// Lightweight Fast Markdown Compiler
function compileMarkdown(md) {
  if (!md) return '';

  let html = md
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/^### (.*$)/gim, '<h3 style="color: var(--accent-cyan); margin: 1.25rem 0 0.5rem 0; font-size: 1.15rem;">$1</h3>')
    .replace(/^## (.*$)/gim, '<h2 style="color: #fff; margin: 1.5rem 0 0.75rem 0; font-size: 1.35rem;">$1</h2>')
    .replace(/^# (.*$)/gim, '<h1 style="color: #fff; margin: 1.75rem 0 1rem 0; font-size: 1.6rem;">$1</h1>')
    .replace(/```([a-z]*)\n([\s\S]*?)```/gim, '<div class="code-block">$2</div>')
    .replace(/`([^`]+)`/gim, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/gim, '<strong style="color: #f1f5f9;">$1</strong>')
    .replace(/\*([^*]+)\*/gim, '<em>$1</em>')
    .replace(/^\> (.*$)/gim, '<blockquote style="border-left: 3px solid var(--accent-cyan); padding-left: 1rem; color: var(--text-secondary); margin: 0.75rem 0;">$1</blockquote>')
    .replace(/^\s*-\s+(.*$)/gim, '<li style="margin-left: 1.25rem; margin-bottom: 0.35rem;">$1</li>')
    .replace(/^\s*\*\s+(.*$)/gim, '<li style="margin-left: 1.25rem; margin-bottom: 0.35rem;">$1</li>')
    .replace(/\[([^\]]+)\]\(([^)]+)\)/gim, '<a href="$2" target="_blank" style="color: var(--accent-cyan); text-decoration: underline;">$1</a>')
    .replace(/\n\n/gim, '<br><br>');

  return `<div class="markdown-body">${html}</div>`;
}
