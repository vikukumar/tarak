import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "TARAK — High-Performance Native Container Orchestration Platform",
  description:
    "Enterprise Cluster Control Plane, ArgoCD GitOps, Zero-Trust Multi-Mesh, and Hubble Telemetry Dashboard.",
  icons: {
    icon: [
      { url: "/assets/favicon.ico" },
      { url: "/assets/favicon-16x16.png", sizes: "16x16", type: "image/png" },
      { url: "/assets/favicon-32x32.png", sizes: "32x32", type: "image/png" },
      { url: "/assets/icon.png", sizes: "512x512", type: "image/png" },
    ],
    apple: [
      { url: "/assets/apple-touch-icon.png", sizes: "180x180", type: "image/png" },
    ],
  },
  manifest: "/site.webmanifest",
  openGraph: {
    title: "TARAK — Ultra-Lightweight Zero-Dependency Kubernetes Alternative",
    description:
      "Next-generation pure Go container orchestrator. 10x faster & 20x lighter than K8s and K3s. Inbuilt Cloudflare Tunnels, Tailscale private mesh, zero external daemons.",
    url: "https://tarak.vikshro.in/",
    siteName: "TARAK Platform",
    images: [
      {
        url: "/assets/og-image.jpg",
        width: 1200,
        height: 630,
        alt: "TARAK Container Orchestration Platform",
      },
    ],
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className="bg-[#070c18] text-slate-100 min-h-screen antialiased selection:bg-cyan-500/30 selection:text-cyan-300">
        {children}
      </body>
    </html>
  );
}
