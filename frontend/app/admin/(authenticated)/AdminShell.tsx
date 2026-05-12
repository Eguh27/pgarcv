"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import {
  ChevronRight,
  Image,
  LayoutDashboard,
  LogOut,
  Megaphone,
  Menu,
  Video,
  X,
} from "lucide-react";
import { adminApi } from "@/lib/api";
import { ThemeToggle } from "@/components/ui/ThemeToggle";

const navItems = [
  { href: "/admin/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/admin/videos", label: "Video", icon: Video },
  { href: "/admin/banners", label: "Banner", icon: Image },
  { href: "/admin/ads", label: "Iklan", icon: Megaphone },
];

function SidebarContent({
  pathname,
  onNavigate,
  onLogout,
}: {
  pathname: string;
  onNavigate: () => void;
  onLogout: () => void;
}) {
  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <div
        style={{
          padding: "24px 20px",
          borderBottom: "1px solid var(--border)",
          display: "flex",
          alignItems: "center",
          gap: "12px",
        }}
      >
        <div
          style={{
            width: "36px",
            height: "36px",
            borderRadius: "10px",
            background: "var(--accent)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <Video size={20} color="#fff" />
        </div>
        <div>
          <p style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "15px", color: "var(--text-primary)" }}>
            PGarcv
          </p>
          <p style={{ fontSize: "11px", color: "var(--text-muted)" }}>Admin Panel</p>
        </div>
      </div>

      <nav style={{ flex: 1, padding: "16px 12px", display: "flex", flexDirection: "column", gap: "4px" }}>
        {navItems.map(({ href, label, icon: Icon }) => {
          const active = pathname.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              onClick={onNavigate}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "12px",
                padding: "10px 12px",
                borderRadius: "10px",
                textDecoration: "none",
                background: active ? "var(--accent)" : "transparent",
                color: active ? "#fff" : "var(--text-secondary)",
                fontFamily: "var(--font-montserrat)",
                fontSize: "14px",
                fontWeight: active ? 600 : 400,
                transition: "all 0.15s",
              }}
            >
              <Icon size={18} />
              {label}
              {active && <ChevronRight size={14} style={{ marginLeft: "auto" }} />}
            </Link>
          );
        })}
      </nav>

      <div style={{ padding: "16px 12px", borderTop: "1px solid var(--border)" }}>
        <button
          onClick={onLogout}
          style={{
            width: "100%",
            display: "flex",
            alignItems: "center",
            gap: "12px",
            padding: "10px 12px",
            borderRadius: "10px",
            border: "none",
            background: "transparent",
            color: "var(--text-secondary)",
            cursor: "pointer",
            fontFamily: "var(--font-montserrat)",
            fontSize: "14px",
            transition: "all 0.15s",
          }}
        >
          <LogOut size={18} />
          Logout
        </button>
      </div>
    </div>
  );
}

export function AdminShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [isAuthed, setIsAuthed] = useState(false);
  const [isChecking, setIsChecking] = useState(true);

  // ✅ Proper auth check - wait untuk token di localStorage
  useEffect(() => {
    try {
      const token = localStorage.getItem("admin_token");
      if (token) {
        setIsAuthed(true);
      } else if (pathname !== "/admin/login") {
        // Redirect ke login kalau tidak ada token
        window.location.href = "/admin/login";
        return;
      }
    } catch (err) {
      console.error("Auth check error:", err);
    }
    setIsChecking(false);
  }, [pathname]);

  // ✅ Show loading sampai auth check selesai
  if (isChecking) {
    return (
      <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", background: "var(--bg-secondary)" }}>
        <p style={{ color: "var(--text-muted)", fontFamily: "var(--font-poppins)" }}>Memverifikasi akses...</p>
      </div>
    );
  }

  // ✅ Jangan render children kalau belum authed
  if (!isAuthed && pathname !== "/admin/login") {
    return null;
  }

  const handleLogout = async () => {
    await adminApi.auth.logout();
    try {
      localStorage.removeItem("admin_token");
    } catch {}
    window.location.href = "/admin/login";
  };

  const closeSidebar = () => setSidebarOpen(false);

  return (
    <div style={{ display: "flex", minHeight: "100vh", background: "var(--bg-secondary)" }}>
      {/* Sidebar desktop */}
      <aside
        style={{
          width: "240px",
          flexShrink: 0,
          background: "var(--bg-card)",
          borderRight: "1px solid var(--border)",
          position: "sticky",
          top: 0,
          height: "100vh",
          display: "flex",
          flexDirection: "column",
        }}
        className="sidebar-desktop"
      >
        <SidebarContent pathname={pathname} onNavigate={closeSidebar} onLogout={handleLogout} />
      </aside>

      {/* Overlay mobile */}
      {sidebarOpen && (
        <div
          onClick={closeSidebar}
          style={{
            position: "fixed",
            inset: 0,
            zIndex: 40,
            background: "var(--overlay)",
          }}
        />
      )}

      {/* Sidebar mobile */}
      <aside
        style={{
          position: "fixed",
          left: 0,
          top: 0,
          bottom: 0,
          width: "240px",
          zIndex: 50,
          background: "var(--bg-card)",
          borderRight: "1px solid var(--border)",
          transform: sidebarOpen ? "translateX(0)" : "translateX(-100%)",
          transition: "transform 0.25s",
          display: "flex",
          flexDirection: "column",
        }}
        className="sidebar-mobile"
      >
        <button
          onClick={closeSidebar}
          style={{
            position: "absolute",
            top: "16px",
            right: "16px",
            background: "none",
            border: "none",
            cursor: "pointer",
            color: "var(--text-secondary)",
          }}
          aria-label="Tutup menu"
        >
          <X size={20} />
        </button>
        <SidebarContent pathname={pathname} onNavigate={closeSidebar} onLogout={handleLogout} />
      </aside>

      {/* Main content */}
      <div style={{ flex: 1, display: "flex", flexDirection: "column", minWidth: 0 }}>
        <header
          style={{
            height: "64px",
            background: "var(--bg-card)",
            borderBottom: "1px solid var(--border)",
            display: "flex",
            alignItems: "center",
            padding: "0 24px",
            gap: "16px",
            position: "sticky",
            top: 0,
            zIndex: 30,
          }}
        >
          <button
            onClick={() => setSidebarOpen(true)}
            style={{
              background: "none",
              border: "none",
              cursor: "pointer",
              color: "var(--text-secondary)",
              display: "none",
            }}
            className="menu-btn"
            aria-label="Buka menu"
          >
            <Menu size={22} />
          </button>
          <span
            style={{
              fontFamily: "var(--font-poppins)",
              fontWeight: 600,
              fontSize: "16px",
              color: "var(--text-primary)",
            }}
          >
            {navItems.find((n) => pathname.startsWith(n.href))?.label ?? "Admin"}
          </span>
          <div style={{ marginLeft: "auto" }}>
            <ThemeToggle />
          </div>
        </header>

        <main style={{ padding: "24px", flex: 1 }}>{children}</main>
      </div>

      <style>{`
        @media (max-width: 768px) {
          .sidebar-desktop { display: none !important; }
          .menu-btn { display: flex !important; }
        }
        @media (min-width: 769px) {
          .sidebar-mobile { display: none !important; }
        }
      `}</style>
    </div>
  );
}

