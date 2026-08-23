import React, { useEffect, useRef, useState } from "react";

export const ClusterCanvas: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [scrollProgress, setScrollProgress] = useState(0);
  const [showBackToTop, setShowBackToTop] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      const scrollTop = window.scrollY || document.documentElement.scrollTop;
      const docHeight = document.documentElement.scrollHeight - document.documentElement.clientHeight;
      const progress = docHeight > 0 ? (scrollTop / docHeight) * 100 : 0;
      setScrollProgress(progress);
      setShowBackToTop(scrollTop > 350);
    };

    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let width = (canvas.width = window.innerWidth);
    let height = (canvas.height = window.innerHeight);

    let mouse = { x: width / 2, y: height / 2, active: false };

    const handleResize = () => {
      width = canvas.width = window.innerWidth;
      height = canvas.height = window.innerHeight;
      initNodes();
    };

    const handleMouseMove = (e: MouseEvent) => {
      mouse.x = e.clientX;
      mouse.y = e.clientY;
      mouse.active = true;
    };

    const handleMouseLeave = () => {
      mouse.active = false;
    };

    const handleClick = (e: MouseEvent) => {
      // Burst packets on click
      for (let i = 0; i < 8; i++) {
        packets.push({
          from: 0,
          to: Math.floor(Math.random() * nodes.length),
          progress: 0,
          speed: 0.01 + Math.random() * 0.015,
          color: Math.random() > 0.5 ? "#00f0ff" : "#ec4899",
          size: 3,
        });
      }
    };

    window.addEventListener("resize", handleResize);
    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseleave", handleMouseLeave);
    window.addEventListener("click", handleClick);

    let nodes: any[] = [];
    let packets: any[] = [];
    let starDust: any[] = [];

    function initNodes() {
      nodes = [];
      packets = [];
      starDust = [];

      const isMobile = width < 768;
      const nodeCount = isMobile ? 6 : 14;

      // Ambient background stardust particles
      const starCount = isMobile ? 30 : 70;
      for (let i = 0; i < starCount; i++) {
        starDust.push({
          x: Math.random() * width,
          y: Math.random() * height,
          radius: Math.random() * 1.5 + 0.5,
          alpha: Math.random() * 0.6 + 0.2,
          speed: Math.random() * 0.02 + 0.005,
          phase: Math.random() * Math.PI * 2,
        });
      }

      nodes.push({
        x: width * 0.5,
        y: height * 0.35,
        vx: (Math.random() - 0.5) * 0.4,
        vy: (Math.random() - 0.5) * 0.4,
        radius: 9,
        type: "control-plane",
        color: "#00f0ff",
        pulse: 0,
        pods: 4,
      });

      for (let i = 1; i < nodeCount; i++) {
        const type = i % 4 === 0 ? "edge-gateway" : "worker-node";
        const color = type === "edge-gateway" ? "#a855f7" : "#38bdf8";
        nodes.push({
          x: Math.random() * width,
          y: Math.random() * height,
          vx: (Math.random() - 0.5) * 0.6,
          vy: (Math.random() - 0.5) * 0.6,
          radius: type === "edge-gateway" ? 7.5 : 6,
          type: type,
          color: color,
          pulse: Math.random() * Math.PI * 2,
          pods: Math.floor(Math.random() * 3) + 2,
        });
      }

      const packetCount = isMobile ? 8 : 22;
      for (let i = 0; i < packetCount; i++) {
        packets.push({
          from: Math.floor(Math.random() * nodes.length),
          to: Math.floor(Math.random() * nodes.length),
          progress: Math.random(),
          speed: 0.003 + Math.random() * 0.006,
          color: Math.random() > 0.4 ? "#00f0ff" : "#a855f7",
          size: 2.2,
        });
      }
    }

    initNodes();

    let animId: number;
    function animate() {
      if (!ctx) return;
      ctx.clearRect(0, 0, width, height);

      // Star dust twinkle
      starDust.forEach((s) => {
        s.phase += s.speed;
        const currentAlpha = s.alpha * (0.5 + 0.5 * Math.sin(s.phase));
        ctx.beginPath();
        ctx.arc(s.x, s.y, s.radius, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(148, 163, 184, ${currentAlpha})`;
        ctx.fill();
      });

      // Background perspective cyber grid
      ctx.strokeStyle = "rgba(255, 255, 255, 0.02)";
      ctx.lineWidth = 1;
      const gridSize = 75;
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

      // Nodes
      nodes.forEach((n, idx) => {
        n.x += n.vx;
        n.y += n.vy;
        n.pulse += 0.04;

        if (n.x < 50 || n.x > width - 50) n.vx *= -1;
        if (n.y < 50 || n.y > height - 50) n.vy *= -1;

        if (mouse.active) {
          const dx = mouse.x - n.x;
          const dy = mouse.y - n.y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < 190 && dist > 1) {
            const force = (190 - dist) / 190;
            n.x -= (dx / dist) * force * 1.6;
            n.y -= (dy / dist) * force * 1.6;
          }
        }

        // Interconnect lines
        for (let j = idx + 1; j < nodes.length; j++) {
          const other = nodes[j];
          const dx = other.x - n.x;
          const dy = other.y - n.y;
          const dist = Math.sqrt(dx * dx + dy * dy);

          const maxDist = width < 768 ? 160 : 260;
          if (dist < maxDist) {
            const alpha = (1 - dist / maxDist) * 0.25;
            ctx.strokeStyle = `rgba(0, 240, 255, ${alpha})`;
            ctx.lineWidth = 1.2;
            ctx.beginPath();
            ctx.moveTo(n.x, n.y);
            ctx.lineTo(other.x, other.y);
            ctx.stroke();
          }
        }

        // Pulsing Sonar Ring for Leader
        if (n.type === "control-plane") {
          const waveRadius = n.radius + 12 + ((n.pulse * 8) % 24);
          const waveAlpha = Math.max(0, 0.35 - (waveRadius / 36));
          ctx.beginPath();
          ctx.arc(n.x, n.y, waveRadius, 0, Math.PI * 2);
          ctx.strokeStyle = `rgba(0, 240, 255, ${waveAlpha})`;
          ctx.lineWidth = 1.5;
          ctx.stroke();
        }

        // Node Outer Glow Halo
        const haloSize = n.radius + 6 + Math.sin(n.pulse) * 3;
        ctx.beginPath();
        ctx.arc(n.x, n.y, haloSize, 0, Math.PI * 2);
        ctx.fillStyle = n.type === "control-plane" ? "rgba(0, 240, 255, 0.1)" : "rgba(168, 85, 247, 0.08)";
        ctx.fill();

        // Node Core
        ctx.beginPath();
        ctx.arc(n.x, n.y, n.radius, 0, Math.PI * 2);
        ctx.fillStyle = n.color;
        ctx.shadowColor = n.color;
        ctx.shadowBlur = 14;
        ctx.fill();
        ctx.shadowBlur = 0;

        // Orbiting pods
        const podOrbitRadius = n.radius + 11;
        for (let p = 0; p < n.pods; p++) {
          const podAngle = n.pulse * 0.8 + (p * (Math.PI * 2)) / n.pods;
          const px = n.x + Math.cos(podAngle) * podOrbitRadius;
          const py = n.y + Math.sin(podAngle) * podOrbitRadius;

          ctx.beginPath();
          ctx.arc(px, py, 2, 0, Math.PI * 2);
          ctx.fillStyle = p === 0 ? "#10b981" : p === 1 ? "#00f0ff" : "#a855f7";
          ctx.shadowColor = ctx.fillStyle;
          ctx.shadowBlur = 6;
          ctx.fill();
          ctx.shadowBlur = 0;
        }
      });

      // Packets
      packets.forEach((pkt, pIdx) => {
        pkt.progress += pkt.speed;
        if (pkt.progress >= 1) {
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
        ctx.shadowBlur = 10;
        ctx.fill();
        ctx.shadowBlur = 0;
      });

      animId = requestAnimationFrame(animate);
    }

    animate();

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener("resize", handleResize);
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseleave", handleMouseLeave);
      window.removeEventListener("click", handleClick);
    };
  }, []);

  return (
    <>
      {/* Scroll Progress Bar */}
      <div
        className="fixed top-0 left-0 h-[3px] z-[99999] pointer-events-none transition-all duration-100 bg-gradient-to-r from-cyan-400 via-purple-400 to-indigo-400 shadow-[0_0_12px_rgba(0,240,255,0.9)]"
        style={{ width: `${scrollProgress}%` }}
      />

      {/* Floating Back to Top Button */}
      {showBackToTop && (
        <button
          onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
          className="fixed bottom-6 right-6 p-3 rounded-xl bg-slate-900/80 backdrop-blur-xl border border-cyan-500/40 text-cyan-300 shadow-[0_8px_24px_rgba(0,0,0,0.7),0_0_15px_rgba(0,240,255,0.3)] hover:scale-110 hover:border-cyan-400 hover:shadow-[0_12px_30px_rgba(0,240,255,0.5)] transition-all z-50 animate-fade-in"
          aria-label="Back to Top"
        >
          ↑
        </button>
      )}

      {/* Ambient Glowing Orbs */}
      <div className="fixed -top-36 -left-36 w-[550px] h-[550px] rounded-full bg-cyan-500/15 blur-[140px] pointer-events-none z-0 animate-pulse" />
      <div className="fixed -bottom-36 -right-36 w-[550px] h-[550px] rounded-full bg-purple-500/15 blur-[140px] pointer-events-none z-0 animate-pulse" />

      {/* Cluster Canvas */}
      <canvas
        ref={canvasRef}
        id="cluster-canvas"
        style={{
          position: "fixed",
          top: 0,
          left: 0,
          width: "100vw",
          height: "100vh",
          pointerEvents: "none",
          zIndex: 0,
          opacity: 0.85,
        }}
      />
    </>
  );
};
