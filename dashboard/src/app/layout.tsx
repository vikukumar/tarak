import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "TARAK — High-Performance Native Container Orchestration Platform",
  description: "Enterprise Cluster Control Plane, ArgoCD GitOps, Zero-Trust Multi-Mesh, and Hubble Telemetry Dashboard.",
  icons: {
    icon: "/assets/tarak_icon.jpg",
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
