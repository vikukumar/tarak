import React, { useEffect, useRef } from "react";

export const ClusterCanvas: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

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

    window.addEventListener("resize", handleResize);
    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseleave", handleMouseLeave);

    let nodes: any[] = [];
    let packets: any[] = [];

    function initNodes() {
      nodes = [];
      packets = [];

      const isMobile = width < 768;
      const nodeCount = isMobile ? 6 : 14;

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

      // Background grid
      ctx.strokeStyle = "rgba(255, 255, 255, 0.018)";
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
          if (dist < 180 && dist > 1) {
            const force = (180 - dist) / 180;
            n.x -= (dx / dist) * force * 1.5;
            n.y -= (dy / dist) * force * 1.5;
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
            const alpha = (1 - dist / maxDist) * 0.22;
            ctx.strokeStyle = `rgba(0, 240, 255, ${alpha})`;
            ctx.lineWidth = 1.2;
            ctx.beginPath();
            ctx.moveTo(n.x, n.y);
            ctx.lineTo(other.x, other.y);
            ctx.stroke();
          }
        }

        // Node Glow Halo
        const haloSize = n.radius + 6 + Math.sin(n.pulse) * 3;
        ctx.beginPath();
        ctx.arc(n.x, n.y, haloSize, 0, Math.PI * 2);
        ctx.fillStyle = n.type === "control-plane" ? "rgba(0, 240, 255, 0.08)" : "rgba(168, 85, 247, 0.06)";
        ctx.fill();

        // Node Core
        ctx.beginPath();
        ctx.arc(n.x, n.y, n.radius, 0, Math.PI * 2);
        ctx.fillStyle = n.color;
        ctx.shadowColor = n.color;
        ctx.shadowBlur = 12;
        ctx.fill();
        ctx.shadowBlur = 0;

        // Orbiting pods
        const podOrbitRadius = n.radius + 10;
        for (let p = 0; p < n.pods; p++) {
          const podAngle = n.pulse * 0.8 + (p * (Math.PI * 2)) / n.pods;
          const px = n.x + Math.cos(podAngle) * podOrbitRadius;
          const py = n.y + Math.sin(podAngle) * podOrbitRadius;

          ctx.beginPath();
          ctx.arc(px, py, 1.8, 0, Math.PI * 2);
          ctx.fillStyle = p === 0 ? "#10b981" : p === 1 ? "#00f0ff" : "#a855f7";
          ctx.fill();
        }
      });

      // Packets
      packets.forEach((pkt) => {
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
        ctx.shadowBlur = 8;
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
    };
  }, []);

  return (
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
        opacity: 0.75,
      }}
    />
  );
};
