"use client";

import { useRouter } from "next/navigation";
import { adminApi } from "@/lib/api";
import { useState } from "react";

export function AdminNavbar() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);

  const handleLogout = async () => {
    setLoading(true);
    try {
      await adminApi.auth.logout();
      router.push("/admin/login");
    } catch (err) {
      console.error("Logout failed:", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <nav
      style={{
        background: "var(--bg-card)",
        borderBottom: "1px solid var(--border)",
        height: "80px",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        padding: "0 24px",
        position: "sticky",
        top: 0,
        zIndex: 100,
      }}
    >
      <h1
        style={{
          fontFamily: "var(--font-poppins)",
          fontSize: "20px",
          fontWeight: 700,
        }}
      >
        Admin Panel
      </h1>

      <div style={{ display: "flex", gap: "16px", alignItems: "center" }}>
        <div style={{ display: "flex", gap: "24px" }}>
          {[
            { label: "Dashboard", href: "/admin/dashboard" },
            { label: "Video", href: "/admin/videos" },
            { label: "Banner", href: "/admin/banners" },
            { label: "Iklan", href: "/admin/ads" },
          ].map((item) => (
            <a
              key={item.href}
              href={item.href}
              style={{
                color: "var(--text-secondary)",
                textDecoration: "none",
                fontSize: "14px",
                fontWeight: 500,
                transition: "color 0.2s",
                cursor: "pointer",
              }}
              onMouseEnter={(e) =>
                (e.currentTarget.style.color = "var(--accent)")
              }
              onMouseLeave={(e) =>
                (e.currentTarget.style.color = "var(--text-secondary)")
              }
            >
              {item.label}
            </a>
          ))}
        </div>

        <button
          onClick={handleLogout}
          disabled={loading}
          style={{
            padding: "8px 16px",
            background: "#ef4444",
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            fontSize: "13px",
            fontWeight: 500,
            cursor: loading ? "not-allowed" : "pointer",
            opacity: loading ? 0.7 : 1,
            transition: "opacity 0.2s",
          }}
        >
          {loading ? "Logout..." : "Logout"}
        </button>
      </div>
    </nav>
  );
}
