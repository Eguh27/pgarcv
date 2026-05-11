"use client";

import dynamic from "next/dynamic";

const AdminShell = dynamic(() => import("./AdminShell").then((m) => m.AdminShell), {
  ssr: false,
  loading: () => (
    <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", background: "var(--bg-secondary)" }}>
      <p style={{ color: "var(--text-muted)", fontFamily: "var(--font-poppins)" }}>Memuat...</p>
    </div>
  ),
});

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return <AdminShell>{children}</AdminShell>;
}