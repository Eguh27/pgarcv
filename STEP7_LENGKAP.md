# ⚡ STEP 7 — KODE LENGKAP SIAP PAKAI
> Semua file ini dibuat langsung tanpa perlu agent lain.
> Copy-paste sesuai path masing-masing.

---

## 1️⃣ MIDDLEWARE AUTH — `frontend/middleware.ts`

```typescript
// frontend/middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get("admin_token")?.value;

  const isAdminRoute = pathname.startsWith("/admin");
  const isLoginPage = pathname === "/admin/login";

  if (isAdminRoute && !isLoginPage && !token) {
    return NextResponse.redirect(new URL("/admin/login", request.url));
  }

  if (isLoginPage && token) {
    return NextResponse.redirect(new URL("/admin/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/admin/:path*"],
};
```

---

## 2️⃣ ADMIN LAYOUT — `frontend/app/(admin)/layout.tsx`

```typescript
// frontend/app/(admin)/layout.tsx
"use client";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";
import {
  LayoutDashboard, Video, Image, Megaphone,
  LogOut, Menu, X, ChevronRight,
} from "lucide-react";
import { adminApi } from "@/lib/api";
import { ThemeToggle } from "@/components/ui/ThemeToggle";

const navItems = [
  { href: "/admin/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/admin/videos", label: "Video", icon: Video },
  { href: "/admin/banners", label: "Banner", icon: Image },
  { href: "/admin/ads", label: "Iklan", icon: Megaphone },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  if (pathname === "/admin/login") return <>{children}</>;

  const handleLogout = async () => {
    await adminApi.auth.logout();
    router.push("/admin/login");
  };

  const SidebarContent = () => (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      {/* Logo */}
      <div style={{
        padding: "24px 20px", borderBottom: "1px solid var(--border)",
        display: "flex", alignItems: "center", gap: "12px",
      }}>
        <div style={{
          width: "36px", height: "36px", borderRadius: "10px",
          background: "var(--accent)", display: "flex",
          alignItems: "center", justifyContent: "center",
        }}>
          <Video size={20} color="#fff" />
        </div>
        <div>
          <p style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "15px", color: "var(--text-primary)" }}>
            PGarcv
          </p>
          <p style={{ fontSize: "11px", color: "var(--text-muted)" }}>Admin Panel</p>
        </div>
      </div>

      {/* Nav items */}
      <nav style={{ flex: 1, padding: "16px 12px", display: "flex", flexDirection: "column", gap: "4px" }}>
        {navItems.map(({ href, label, icon: Icon }) => {
          const active = pathname.startsWith(href);
          return (
            <Link key={href} href={href} onClick={() => setSidebarOpen(false)}
              style={{
                display: "flex", alignItems: "center", gap: "12px",
                padding: "10px 12px", borderRadius: "10px",
                textDecoration: "none",
                background: active ? "var(--accent)" : "transparent",
                color: active ? "#fff" : "var(--text-secondary)",
                fontFamily: "var(--font-montserrat)",
                fontSize: "14px", fontWeight: active ? 600 : 400,
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

      {/* Logout */}
      <div style={{ padding: "16px 12px", borderTop: "1px solid var(--border)" }}>
        <button onClick={handleLogout}
          style={{
            width: "100%", display: "flex", alignItems: "center", gap: "12px",
            padding: "10px 12px", borderRadius: "10px",
            border: "none", background: "transparent",
            color: "var(--text-secondary)", cursor: "pointer",
            fontFamily: "var(--font-montserrat)", fontSize: "14px",
            transition: "all 0.15s",
          }}
          onMouseEnter={(e) => { (e.currentTarget as HTMLButtonElement).style.background = "var(--surface)"; }}
          onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.background = "transparent"; }}
        >
          <LogOut size={18} />
          Logout
        </button>
      </div>
    </div>
  );

  return (
    <div style={{ display: "flex", minHeight: "100vh", background: "var(--bg-secondary)" }}>
      {/* Sidebar desktop */}
      <aside style={{
        width: "240px", flexShrink: 0,
        background: "var(--bg-card)",
        borderRight: "1px solid var(--border)",
        position: "sticky", top: 0, height: "100vh",
        display: "flex", flexDirection: "column",
      }}
        className="sidebar-desktop"
      >
        <SidebarContent />
      </aside>

      {/* Sidebar mobile overlay */}
      {sidebarOpen && (
        <div
          onClick={() => setSidebarOpen(false)}
          style={{
            position: "fixed", inset: 0, zIndex: 40,
            background: "var(--overlay)",
          }}
        />
      )}
      <aside style={{
        position: "fixed", left: 0, top: 0, bottom: 0,
        width: "240px", zIndex: 50,
        background: "var(--bg-card)",
        borderRight: "1px solid var(--border)",
        transform: sidebarOpen ? "translateX(0)" : "translateX(-100%)",
        transition: "transform 0.25s",
        display: "flex", flexDirection: "column",
      }}
        className="sidebar-mobile"
      >
        <button
          onClick={() => setSidebarOpen(false)}
          style={{
            position: "absolute", top: "16px", right: "16px",
            background: "none", border: "none", cursor: "pointer",
            color: "var(--text-secondary)",
          }}
        >
          <X size={20} />
        </button>
        <SidebarContent />
      </aside>

      {/* Main */}
      <div style={{ flex: 1, display: "flex", flexDirection: "column", minWidth: 0 }}>
        {/* Topbar */}
        <header style={{
          height: "64px", background: "var(--bg-card)",
          borderBottom: "1px solid var(--border)",
          display: "flex", alignItems: "center",
          padding: "0 24px", gap: "16px",
          position: "sticky", top: 0, zIndex: 30,
        }}>
          <button
            onClick={() => setSidebarOpen(true)}
            style={{ background: "none", border: "none", cursor: "pointer", color: "var(--text-secondary)", display: "none" }}
            className="menu-btn"
          >
            <Menu size={22} />
          </button>
          <span style={{ fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "16px", color: "var(--text-primary)" }}>
            {navItems.find((n) => pathname.startsWith(n.href))?.label ?? "Admin"}
          </span>
          <div style={{ marginLeft: "auto" }}>
            <ThemeToggle />
          </div>
        </header>

        {/* Content */}
        <main style={{ padding: "24px", flex: 1 }}>
          {children}
        </main>
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
```

---

## 3️⃣ ADMIN DASHBOARD — `frontend/app/(admin)/dashboard/page.tsx`

```bash
mkdir -p "frontend/app/(admin)/dashboard"
```

```typescript
// frontend/app/(admin)/dashboard/page.tsx
"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { Eye, Video, TrendingUp, Plus } from "lucide-react";
import { adminApi, mediaUrl } from "@/lib/api";
import type { Video as VideoType } from "@/lib/api";

interface StatCard {
  label: string;
  value: string | number;
  icon: React.ReactNode;
  color: string;
}

export default function DashboardPage() {
  const [videos, setVideos] = useState<VideoType[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    adminApi.videos.list()
      .then((res) => setVideos((res as { data: VideoType[] }).data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const totalViews = videos.reduce((sum, v) => sum + v.views, 0);
  const published = videos.filter((v) => v.is_published).length;

  const stats: StatCard[] = [
    { label: "Total Video", value: videos.length, icon: <Video size={22} />, color: "#6366f1" },
    { label: "Dipublikasikan", value: published, icon: <TrendingUp size={22} />, color: "#10b981" },
    { label: "Total Penonton", value: totalViews.toLocaleString("id-ID"), icon: <Eye size={22} />, color: "var(--accent)" },
  ];

  return (
    <div>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "24px" }}>
        <div>
          <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px", color: "var(--text-primary)" }}>Dashboard</h1>
          <p style={{ fontSize: "13px", color: "var(--text-muted)", marginTop: "2px" }}>Selamat datang kembali 👋</p>
        </div>
        <Link href="/admin/videos/new"
          style={{
            display: "flex", alignItems: "center", gap: "8px",
            padding: "10px 18px", borderRadius: "10px",
            background: "var(--accent)", color: "#fff",
            textDecoration: "none", fontFamily: "var(--font-poppins)",
            fontWeight: 600, fontSize: "14px",
          }}
        >
          <Plus size={16} /> Tambah Video
        </Link>
      </div>

      {/* Stat cards */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))", gap: "16px", marginBottom: "32px" }}>
        {stats.map((s) => (
          <div key={s.label} style={{
            background: "var(--bg-card)", border: "1px solid var(--border)",
            borderRadius: "14px", padding: "20px",
            display: "flex", flexDirection: "column", gap: "12px",
          }}>
            <div style={{
              width: "44px", height: "44px", borderRadius: "12px",
              background: s.color + "20",
              display: "flex", alignItems: "center", justifyContent: "center",
              color: s.color,
            }}>
              {s.icon}
            </div>
            <div>
              <p style={{ fontSize: "26px", fontFamily: "var(--font-poppins)", fontWeight: 700, color: "var(--text-primary)" }}>
                {loading ? "—" : s.value}
              </p>
              <p style={{ fontSize: "13px", color: "var(--text-muted)" }}>{s.label}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Tabel video terbaru */}
      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", overflow: "hidden" }}>
        <div style={{ padding: "16px 20px", borderBottom: "1px solid var(--border)", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <h2 style={{ fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "16px" }}>Video Terbaru</h2>
          <Link href="/admin/videos" style={{ fontSize: "13px", color: "var(--accent)", textDecoration: "none" }}>Lihat semua →</Link>
        </div>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr style={{ background: "var(--bg-secondary)" }}>
                {["Thumbnail", "Judul", "Views", "Status", "Aksi"].map((h) => (
                  <th key={h} style={{ padding: "12px 16px", textAlign: "left", fontSize: "12px", fontWeight: 600, color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.05em" }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr><td colSpan={5} style={{ padding: "32px", textAlign: "center", color: "var(--text-muted)" }}>Memuat...</td></tr>
              ) : videos.slice(0, 5).map((v) => (
                <tr key={v.id} style={{ borderTop: "1px solid var(--border)" }}>
                  <td style={{ padding: "12px 16px" }}>
                    <div style={{ width: "60px", height: "34px", borderRadius: "6px", background: "var(--surface)", overflow: "hidden" }}>
                      {v.thumbnail_url && (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={mediaUrl(v.thumbnail_url)} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                      )}
                    </div>
                  </td>
                  <td style={{ padding: "12px 16px" }}>
                    <p style={{ fontSize: "14px", fontWeight: 500, color: "var(--text-primary)", maxWidth: "260px", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{v.title}</p>
                    <p style={{ fontSize: "12px", color: "var(--text-muted)" }}>{v.subtitle}</p>
                  </td>
                  <td style={{ padding: "12px 16px", fontSize: "13px", color: "var(--text-secondary)" }}>{v.views.toLocaleString("id-ID")}</td>
                  <td style={{ padding: "12px 16px" }}>
                    <span style={{
                      padding: "3px 10px", borderRadius: "20px", fontSize: "12px", fontWeight: 600,
                      background: v.is_published ? "#d1fae5" : "#fef3c7",
                      color: v.is_published ? "#065f46" : "#92400e",
                    }}>
                      {v.is_published ? "Publik" : "Draft"}
                    </span>
                  </td>
                  <td style={{ padding: "12px 16px" }}>
                    <Link href={`/admin/videos/edit/${v.id}`} style={{ fontSize: "13px", color: "var(--accent)", textDecoration: "none" }}>Edit</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
```

---

## 4️⃣ ADMIN VIDEO LIST — `frontend/app/(admin)/videos/page.tsx`

```bash
mkdir -p "frontend/app/(admin)/videos"
```

```typescript
// frontend/app/(admin)/videos/page.tsx
"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus, Pencil, Trash2, Eye } from "lucide-react";
import { adminApi, mediaUrl, formatViews } from "@/lib/api";
import type { Video } from "@/lib/api";

export default function AdminVideosPage() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);
  const [deletingId, setDeletingId] = useState<number | null>(null);

  const fetchVideos = () => {
    setLoading(true);
    adminApi.videos.list()
      .then((res) => setVideos((res as { data: Video[] }).data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchVideos(); }, []);

  const handleDelete = async (id: number, title: string) => {
    if (!confirm(`Hapus video "${title}"?`)) return;
    setDeletingId(id);
    try {
      await adminApi.videos.delete(id);
      setVideos((prev) => prev.filter((v) => v.id !== id));
    } catch (e) {
      alert("Gagal hapus video");
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
        <div>
          <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px" }}>Manajemen Video</h1>
          <p style={{ fontSize: "13px", color: "var(--text-muted)" }}>{videos.length} video total</p>
        </div>
        <Link href="/admin/videos/new"
          style={{
            display: "flex", alignItems: "center", gap: "8px",
            padding: "10px 18px", borderRadius: "10px",
            background: "var(--accent)", color: "#fff",
            textDecoration: "none", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px",
          }}
        >
          <Plus size={16} /> Tambah Video
        </Link>
      </div>

      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", overflow: "hidden" }}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr style={{ background: "var(--bg-secondary)" }}>
                {["#", "Video", "Views", "Durasi", "Status", "Aksi"].map((h) => (
                  <th key={h} style={{ padding: "12px 16px", textAlign: "left", fontSize: "12px", fontWeight: 600, color: "var(--text-muted)", textTransform: "uppercase" }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <tr key={i} style={{ borderTop: "1px solid var(--border)" }}>
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j} style={{ padding: "16px" }}>
                        <div style={{ height: "16px", borderRadius: "4px", background: "var(--surface)", width: j === 1 ? "200px" : "60px" }} />
                      </td>
                    ))}
                  </tr>
                ))
              ) : videos.map((v, idx) => (
                <tr key={v.id} style={{ borderTop: "1px solid var(--border)", transition: "background 0.15s" }}
                  onMouseEnter={(e) => (e.currentTarget.style.background = "var(--bg-secondary)")}
                  onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
                >
                  <td style={{ padding: "12px 16px", fontSize: "13px", color: "var(--text-muted)" }}>{idx + 1}</td>
                  <td style={{ padding: "12px 16px" }}>
                    <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
                      <div style={{ width: "72px", height: "40px", borderRadius: "6px", background: "var(--surface)", overflow: "hidden", flexShrink: 0 }}>
                        {v.thumbnail_url && <img src={mediaUrl(v.thumbnail_url)} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />}
                      </div>
                      <div style={{ minWidth: 0 }}>
                        <p style={{ fontSize: "14px", fontWeight: 500, color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", maxWidth: "240px" }}>{v.title}</p>
                        <p style={{ fontSize: "12px", color: "var(--text-muted)" }}>{v.subtitle || "—"}</p>
                      </div>
                    </div>
                  </td>
                  <td style={{ padding: "12px 16px" }}>
                    <span style={{ display: "flex", alignItems: "center", gap: "4px", fontSize: "13px", color: "var(--text-secondary)" }}>
                      <Eye size={13} /> {formatViews(v.views)}
                    </span>
                  </td>
                  <td style={{ padding: "12px 16px", fontSize: "13px", color: "var(--text-secondary)" }}>
                    {Math.floor(v.duration / 60)}:{String(v.duration % 60).padStart(2, "0")}
                  </td>
                  <td style={{ padding: "12px 16px" }}>
                    <span style={{
                      padding: "3px 10px", borderRadius: "20px", fontSize: "12px", fontWeight: 600,
                      background: v.is_published ? "#d1fae5" : "#fef3c7",
                      color: v.is_published ? "#065f46" : "#92400e",
                    }}>
                      {v.is_published ? "Publik" : "Draft"}
                    </span>
                  </td>
                  <td style={{ padding: "12px 16px" }}>
                    <div style={{ display: "flex", gap: "8px" }}>
                      <Link href={`/admin/videos/edit/${v.id}`}
                        style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "var(--surface)", color: "var(--text-primary)", textDecoration: "none", fontSize: "13px" }}>
                        <Pencil size={13} /> Edit
                      </Link>
                      <button onClick={() => handleDelete(v.id, v.title)} disabled={deletingId === v.id}
                        style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "#fee2e2", color: "#dc2626", border: "none", cursor: "pointer", fontSize: "13px" }}>
                        <Trash2 size={13} /> {deletingId === v.id ? "..." : "Hapus"}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
              {!loading && videos.length === 0 && (
                <tr><td colSpan={6} style={{ padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>Belum ada video. <Link href="/admin/videos/new" style={{ color: "var(--accent)" }}>Tambah sekarang</Link></td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
```

---

## 5️⃣ FORM VIDEO (TAMBAH & EDIT)

### Komponen shared form — `frontend/components/admin/VideoForm.tsx`

```bash
mkdir -p frontend/components/admin
```

```typescript
// frontend/components/admin/VideoForm.tsx
"use client";
import { useState, useRef } from "react";
import { Upload, X, Film, Image as ImageIcon } from "lucide-react";
import { adminApi, mediaUrl } from "@/lib/api";
import type { Video } from "@/lib/api";

type FormData = {
  title: string; subtitle: string; description: string;
  thumbnail_url: string; video_url: string; preview_url: string;
  duration: number; is_published: boolean;
};

interface Props {
  initial?: Partial<Video>;
  onSubmit: (data: FormData) => Promise<void>;
  loading?: boolean;
}

function UploadZone({ label, accept, icon, currentUrl, onUploaded }: {
  label: string; accept: string; icon: React.ReactNode;
  currentUrl: string; onUploaded: (url: string) => void;
}) {
  const [dragging, setDragging] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const upload = async (file: File) => {
    setUploading(true); setProgress(10);
    try {
      // Simulasi progress
      const interval = setInterval(() => setProgress((p) => Math.min(p + 10, 80)), 200);
      const res = await adminApi.upload(file);
      clearInterval(interval); setProgress(100);
      onUploaded(res.url);
    } catch (e) {
      alert("Upload gagal: " + (e instanceof Error ? e.message : "Error"));
    } finally {
      setTimeout(() => { setUploading(false); setProgress(0); }, 500);
    }
  };

  return (
    <div>
      <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "8px", color: "var(--text-secondary)" }}>{label}</label>
      <div
        onClick={() => inputRef.current?.click()}
        onDragOver={(e) => { e.preventDefault(); setDragging(true); }}
        onDragLeave={() => setDragging(false)}
        onDrop={(e) => { e.preventDefault(); setDragging(false); const f = e.dataTransfer.files[0]; if (f) upload(f); }}
        style={{
          border: `2px dashed ${dragging ? "var(--accent)" : "var(--border)"}`,
          borderRadius: "12px", padding: "24px",
          cursor: "pointer", textAlign: "center",
          background: dragging ? "var(--accent)10" : "var(--bg-secondary)",
          transition: "all 0.2s", position: "relative",
          minHeight: "100px",
        }}
      >
        {currentUrl ? (
          <div style={{ position: "relative", display: "inline-block" }}>
            {accept.includes("image") ? (
              <img src={mediaUrl(currentUrl)} alt="" style={{ maxHeight: "120px", maxWidth: "100%", borderRadius: "8px", objectFit: "cover" }} />
            ) : (
              <div style={{ display: "flex", alignItems: "center", gap: "8px", color: "var(--text-secondary)" }}>
                <Film size={20} /> <span style={{ fontSize: "13px" }}>{currentUrl.split("/").pop()}</span>
              </div>
            )}
            <button onClick={(e) => { e.stopPropagation(); onUploaded(""); }}
              style={{ position: "absolute", top: "-8px", right: "-8px", background: "#dc2626", border: "none", borderRadius: "50%", width: "22px", height: "22px", cursor: "pointer", color: "#fff", display: "flex", alignItems: "center", justifyContent: "center" }}>
              <X size={12} />
            </button>
          </div>
        ) : (
          <div style={{ color: "var(--text-muted)" }}>
            <div style={{ marginBottom: "8px" }}>{icon}</div>
            <p style={{ fontSize: "14px" }}>{uploading ? `Mengupload... ${progress}%` : "Klik atau drag file ke sini"}</p>
            <p style={{ fontSize: "12px", marginTop: "4px" }}>{accept.includes("image") ? "JPG, PNG, GIF" : "MP4, WebM, MOV"} (maks 500MB)</p>
          </div>
        )}

        {/* Progress bar */}
        {uploading && (
          <div style={{ position: "absolute", bottom: 0, left: 0, right: 0, height: "4px", background: "var(--surface)", borderRadius: "0 0 10px 10px" }}>
            <div style={{ height: "100%", width: `${progress}%`, background: "var(--accent)", borderRadius: "inherit", transition: "width 0.2s" }} />
          </div>
        )}
      </div>
      <input ref={inputRef} type="file" accept={accept} style={{ display: "none" }}
        onChange={(e) => { const f = e.target.files?.[0]; if (f) upload(f); }} />
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  width: "100%", padding: "10px 14px",
  borderRadius: "8px", border: "1px solid var(--border)",
  background: "var(--bg-secondary)", color: "var(--text-primary)",
  fontSize: "14px", outline: "none",
  fontFamily: "var(--font-montserrat)",
  transition: "border-color 0.15s",
};

export function VideoForm({ initial, onSubmit, loading }: Props) {
  const [form, setForm] = useState<FormData>({
    title: initial?.title ?? "",
    subtitle: initial?.subtitle ?? "",
    description: initial?.description ?? "",
    thumbnail_url: initial?.thumbnail_url ?? "",
    video_url: initial?.video_url ?? "",
    preview_url: initial?.preview_url ?? "",
    duration: initial?.duration ?? 0,
    is_published: initial?.is_published ?? false,
  });

  const set = (key: keyof FormData) => (val: string | number | boolean) =>
    setForm((prev) => ({ ...prev, [key]: val }));

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.title.trim()) { alert("Judul harus diisi"); return; }
    onSubmit(form);
  };

  return (
    <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: "20px" }}>
      {/* Title */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Judul *</label>
        <input value={form.title} onChange={(e) => set("title")(e.target.value)} required placeholder="Masukkan judul video" style={inputStyle} />
      </div>

      {/* Subtitle */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Subtitle <span style={{ color: "var(--text-muted)" }}>(opsional)</span></label>
        <input value={form.subtitle} onChange={(e) => set("subtitle")(e.target.value)} placeholder="Teks kecil di bawah judul" style={inputStyle} />
      </div>

      {/* Deskripsi */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Deskripsi</label>
        <textarea value={form.description} onChange={(e) => set("description")(e.target.value)} rows={4}
          placeholder="Deskripsi lengkap video..."
          style={{ ...inputStyle, resize: "vertical", minHeight: "100px" }} />
      </div>

      {/* Durasi */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Durasi (detik)</label>
        <input type="number" value={form.duration} onChange={(e) => set("duration")(Number(e.target.value))} min={0} style={{ ...inputStyle, maxWidth: "160px" }} />
        <p style={{ fontSize: "12px", color: "var(--text-muted)", marginTop: "4px" }}>
          = {Math.floor(form.duration / 60)}:{String(form.duration % 60).padStart(2, "0")} menit
        </p>
      </div>

      {/* Upload thumbnail */}
      <UploadZone label="Thumbnail" accept="image/*"
        icon={<ImageIcon size={28} style={{ margin: "0 auto" }} />}
        currentUrl={form.thumbnail_url}
        onUploaded={(url) => set("thumbnail_url")(url)} />

      {/* Upload video */}
      <UploadZone label="File Video" accept="video/*"
        icon={<Film size={28} style={{ margin: "0 auto" }} />}
        currentUrl={form.video_url}
        onUploaded={(url) => set("video_url")(url)} />

      {/* Preview URL */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>URL Preview <span style={{ color: "var(--text-muted)" }}>(clip pendek untuk hover)</span></label>
        <input value={form.preview_url} onChange={(e) => set("preview_url")(e.target.value)} placeholder="/uploads/preview-clip.mp4" style={inputStyle} />
      </div>

      {/* Toggle Published */}
      <div style={{ display: "flex", alignItems: "center", gap: "12px", padding: "16px", background: "var(--bg-secondary)", borderRadius: "12px" }}>
        <button type="button"
          onClick={() => set("is_published")(!form.is_published)}
          style={{
            width: "48px", height: "26px", borderRadius: "13px",
            background: form.is_published ? "var(--accent)" : "var(--surface)",
            border: "none", cursor: "pointer", position: "relative",
            transition: "background 0.2s",
          }}
        >
          <span style={{
            position: "absolute", top: "3px",
            left: form.is_published ? "24px" : "3px",
            width: "20px", height: "20px", borderRadius: "50%",
            background: "#fff", transition: "left 0.2s",
          }} />
        </button>
        <div>
          <p style={{ fontSize: "14px", fontWeight: 500, color: "var(--text-primary)" }}>
            {form.is_published ? "Publik" : "Draft"}
          </p>
          <p style={{ fontSize: "12px", color: "var(--text-muted)" }}>
            {form.is_published ? "Video terlihat di halaman utama" : "Video hanya bisa dilihat admin"}
          </p>
        </div>
      </div>

      {/* Submit */}
      <div style={{ display: "flex", gap: "12px", paddingTop: "8px" }}>
        <button type="submit" disabled={loading}
          style={{
            padding: "12px 28px", borderRadius: "10px",
            background: "var(--accent)", color: "#fff", border: "none",
            fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "15px",
            cursor: loading ? "not-allowed" : "pointer",
            opacity: loading ? 0.7 : 1,
          }}>
          {loading ? "Menyimpan..." : "Simpan Video"}
        </button>
        <button type="button" onClick={() => history.back()}
          style={{ padding: "12px 20px", borderRadius: "10px", background: "var(--surface)", color: "var(--text-primary)", border: "none", cursor: "pointer", fontSize: "14px" }}>
          Batal
        </button>
      </div>
    </form>
  );
}
```

### Halaman Tambah — `frontend/app/(admin)/videos/new/page.tsx`

```bash
mkdir -p "frontend/app/(admin)/videos/new"
```

```typescript
// frontend/app/(admin)/videos/new/page.tsx
"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { VideoForm } from "@/components/admin/VideoForm";
import { adminApi } from "@/lib/api";

export default function NewVideoPage() {
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleSubmit = async (data: Parameters<typeof adminApi.videos.create>[0]) => {
    setLoading(true);
    try {
      await adminApi.videos.create(data);
      router.push("/admin/videos");
    } catch (e) {
      alert("Gagal simpan: " + (e instanceof Error ? e.message : "Error"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: "720px" }}>
      <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px", marginBottom: "24px" }}>Tambah Video Baru</h1>
      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", padding: "28px" }}>
        <VideoForm onSubmit={handleSubmit} loading={loading} />
      </div>
    </div>
  );
}
```

### Halaman Edit — `frontend/app/(admin)/videos/edit/[id]/page.tsx`

```bash
mkdir -p "frontend/app/(admin)/videos/edit/[id]"
```

```typescript
// frontend/app/(admin)/videos/edit/[id]/page.tsx
"use client";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { VideoForm } from "@/components/admin/VideoForm";
import { adminApi, api } from "@/lib/api";
import type { Video } from "@/lib/api";

export default function EditVideoPage() {
  const { id } = useParams<{ id: string }>();
  const [video, setVideo] = useState<Video | null>(null);
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  useEffect(() => {
    api.videos.get(Number(id)).then(setVideo).catch(() => router.push("/admin/videos"));
  }, [id, router]);

  const handleSubmit = async (data: Parameters<typeof adminApi.videos.update>[1]) => {
    setLoading(true);
    try {
      await adminApi.videos.update(Number(id), data);
      router.push("/admin/videos");
    } catch (e) {
      alert("Gagal update: " + (e instanceof Error ? e.message : "Error"));
    } finally {
      setLoading(false);
    }
  };

  if (!video) return <div style={{ padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>Memuat...</div>;

  return (
    <div style={{ maxWidth: "720px" }}>
      <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px", marginBottom: "24px" }}>Edit Video</h1>
      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", padding: "28px" }}>
        <VideoForm initial={video} onSubmit={handleSubmit} loading={loading} />
      </div>
    </div>
  );
}
```

---

## 6️⃣ BANNER SLIDER — `frontend/components/video/BannerSlider.tsx`

```typescript
// frontend/components/video/BannerSlider.tsx
"use client";
import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { Banner } from "@/lib/api";
import { mediaUrl } from "@/lib/api";

export function BannerSlider({ banners }: { banners: Banner[] }) {
  const [current, setCurrent] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const resetTimer = () => {
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = setInterval(() => setCurrent((c) => (c + 1) % banners.length), 5000);
  };

  useEffect(() => {
    if (banners.length <= 1) return;
    resetTimer();
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [banners.length]);

  const go = (dir: 1 | -1) => {
    setCurrent((c) => (c + dir + banners.length) % banners.length);
    resetTimer();
  };

  if (!banners.length) return (
    <div style={{ height: "240px", borderRadius: "16px", background: "var(--surface)", display: "flex", alignItems: "center", justifyContent: "center" }}>
      <p style={{ color: "var(--text-muted)" }}>Belum ada banner aktif</p>
    </div>
  );

  return (
    <div style={{ position: "relative", borderRadius: "16px", overflow: "hidden", aspectRatio: "16/5", background: "var(--surface)" }}>
      {/* Slides */}
      {banners.map((b, i) => (
        <a key={b.id} href={b.link_url || "#"} target="_blank" rel="noopener noreferrer"
          style={{
            position: "absolute", inset: 0,
            opacity: i === current ? 1 : 0,
            transition: "opacity 0.6s ease",
            pointerEvents: i === current ? "auto" : "none",
          }}
        >
          <Image src={mediaUrl(b.image_url)} alt={b.title} fill style={{ objectFit: "cover" }} priority={i === 0} />
          {b.title && (
            <div style={{
              position: "absolute", bottom: 0, left: 0, right: 0,
              background: "linear-gradient(transparent, rgba(0,0,0,0.6))",
              padding: "32px 24px 16px",
            }}>
              <p style={{ color: "#fff", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "18px" }}>{b.title}</p>
            </div>
          )}
        </a>
      ))}

      {/* Arrows */}
      {banners.length > 1 && (
        <>
          {[{ dir: -1 as const, side: "left" }, { dir: 1 as const, side: "right" }].map(({ dir, side }) => (
            <button key={side} onClick={(e) => { e.preventDefault(); go(dir); }}
              style={{
                position: "absolute", top: "50%", [side]: "12px",
                transform: "translateY(-50%)",
                background: "rgba(0,0,0,0.45)", border: "none",
                borderRadius: "50%", width: "36px", height: "36px",
                display: "flex", alignItems: "center", justifyContent: "center",
                cursor: "pointer", color: "#fff", transition: "background 0.2s",
                backdropFilter: "blur(4px)",
              }}
            >
              {dir === -1 ? <ChevronLeft size={18} /> : <ChevronRight size={18} />}
            </button>
          ))}
        </>
      )}

      {/* Dots */}
      {banners.length > 1 && (
        <div style={{ position: "absolute", bottom: "12px", left: "50%", transform: "translateX(-50%)", display: "flex", gap: "6px" }}>
          {banners.map((_, i) => (
            <button key={i}
              onClick={() => { setCurrent(i); resetTimer(); }}
              style={{
                width: i === current ? "20px" : "6px", height: "6px",
                borderRadius: "3px", border: "none",
                background: i === current ? "#fff" : "rgba(255,255,255,0.4)",
                cursor: "pointer", padding: 0, transition: "all 0.3s",
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}
```

### Update homepage untuk pakai BannerSlider — `frontend/app/(public)/page.tsx`

```typescript
// frontend/app/(public)/page.tsx  ← GANTI SELURUH ISINYA
import { api } from "@/lib/api";
import { VideoGridClient } from "@/components/video/VideoGridClient";
import { BannerSlider } from "@/components/video/BannerSlider";
import { AdSlot } from "@/components/ui/AdSlot";

export const revalidate = 60;

export default async function HomePage() {
  const [videosRes, bannersRes] = await Promise.allSettled([
    api.videos.list(1),
    api.banners.list(),
  ]);

  const videos = videosRes.status === "fulfilled" ? videosRes.value.data : [];
  const banners = bannersRes.status === "fulfilled" ? bannersRes.value : [];

  return (
    <div>
      <section style={{ marginBottom: "28px" }}>
        <BannerSlider banners={banners} />
      </section>

      <AdSlot position="top" />

      <section>
        <h2 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "20px", marginBottom: "16px", color: "var(--text-primary)" }}>
          Video Terbaru
        </h2>
        <VideoGridClient initialVideos={videos} />
      </section>
    </div>
  );
}
```

---

## 7️⃣ SISTEM IKLAN — `frontend/components/ui/AdSlot.tsx`

```typescript
// frontend/components/ui/AdSlot.tsx
"use client";
import { useEffect, useState } from "react";
import type { Ad } from "@/lib/api";
import { api } from "@/lib/api";

const SESSION_KEY = "ad_shown_session";

export function AdSlot({ position }: { position: "top" | "bottom" | "sidebar" }) {
  const [ad, setAd] = useState<Ad | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    // Cek apakah iklan sudah ditampilkan di sesi ini
    const shown = sessionStorage.getItem(SESSION_KEY);
    if (shown) { setReady(true); return; }

    api.ads.serve()
      .then((res) => {
        if (res) {
          setAd(res);
          sessionStorage.setItem(SESSION_KEY, "1");
        }
      })
      .catch(() => {}) // silent fail — tidak ganggu UX
      .finally(() => setReady(true));
  }, []);

  // Belum ready atau tidak ada iklan
  if (!ready || !ad) return null;

  const wrapperStyle: React.CSSProperties = {
    marginBottom: "24px",
    borderRadius: "10px",
    overflow: "hidden",
    border: "1px solid var(--border)",
    position: "relative",
  };

  if (position === "sidebar") {
    wrapperStyle.marginBottom = "16px";
  }

  return (
    <div style={wrapperStyle}>
      {/* Label "Iklan" kecil */}
      <div style={{
        position: "absolute", top: "6px", right: "8px",
        fontSize: "10px", color: "var(--text-muted)",
        background: "var(--bg-secondary)",
        padding: "1px 6px", borderRadius: "4px",
        zIndex: 1, pointerEvents: "none",
      }}>
        Iklan
      </div>
      <div dangerouslySetInnerHTML={{ __html: ad.ad_code }} />
    </div>
  );
}
```

---

## 8️⃣ HALAMAN SEARCH — `frontend/app/(public)/search/page.tsx`

```bash
mkdir -p "frontend/app/(public)/search"
```

```typescript
// frontend/app/(public)/search/page.tsx
"use client";
import { useSearchParams } from "next/navigation";
import { useEffect, useState, Suspense } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import type { Video } from "@/lib/api";
import { VideoCard } from "@/components/video/VideoCard";

function SkeletonCard() {
  return (
    <div style={{ borderRadius: "12px", overflow: "hidden", background: "var(--bg-card)", border: "1px solid var(--border)" }}>
      <div style={{ aspectRatio: "16/9", background: "var(--surface)" }} />
      <div style={{ padding: "12px", display: "flex", flexDirection: "column", gap: "8px" }}>
        <div style={{ height: "16px", borderRadius: "4px", background: "var(--surface)", width: "80%" }} />
        <div style={{ height: "12px", borderRadius: "4px", background: "var(--surface)", width: "50%" }} />
      </div>
    </div>
  );
}

function SearchResults() {
  const searchParams = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const router = useRouter();

  useEffect(() => {
    if (!query.trim()) return;
    setLoading(true);
    api.videos.list(1, query)
      .then((res) => { setVideos(res.data); setTotal(res.total); })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [query]);

  return (
    <div>
      <div style={{ marginBottom: "20px" }}>
        <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px", marginBottom: "4px" }}>
          Hasil pencarian
        </h1>
        <p style={{ fontSize: "14px", color: "var(--text-muted)" }}>
          {loading ? "Mencari..." : `${total} hasil untuk `}
          {!loading && <strong style={{ color: "var(--text-primary)" }}>"{query}"</strong>}
        </p>
      </div>

      <div style={{
        display: "grid",
        gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))",
        gap: "20px",
      }}>
        {loading ? (
          Array.from({ length: 8 }).map((_, i) => <SkeletonCard key={i} />)
        ) : videos.length > 0 ? (
          videos.map((v) => (
            <VideoCard key={v.id} video={v} onClick={(id) => router.push(`/watch/${id}`)} />
          ))
        ) : (
          <div style={{ gridColumn: "1/-1", padding: "64px 0", textAlign: "center" }}>
            <p style={{ fontSize: "48px", marginBottom: "12px" }}>🔍</p>
            <p style={{ fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "18px", color: "var(--text-primary)", marginBottom: "8px" }}>
              Tidak ada hasil untuk "{query}"
            </p>
            <p style={{ fontSize: "14px", color: "var(--text-muted)" }}>
              Coba kata kunci yang berbeda
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense fallback={<div style={{ padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>Memuat...</div>}>
      <SearchResults />
    </Suspense>
  );
}
```

---

## 9️⃣ ADMIN BANNER PAGE — `frontend/app/(admin)/banners/page.tsx`

```bash
mkdir -p "frontend/app/(admin)/banners"
```

```typescript
// frontend/app/(admin)/banners/page.tsx
"use client";
import { useEffect, useState } from "react";
import { Plus, Pencil, Trash2, ToggleLeft, ToggleRight } from "lucide-react";
import { adminApi, mediaUrl } from "@/lib/api";
import type { Banner } from "@/lib/api";

type BannerForm = { title: string; image_url: string; link_url: string; is_active: boolean; sort_order: number };
const emptyForm: BannerForm = { title: "", image_url: "", link_url: "", is_active: true, sort_order: 0 };

const inputStyle: React.CSSProperties = {
  width: "100%", padding: "9px 12px", borderRadius: "8px",
  border: "1px solid var(--border)", background: "var(--bg-secondary)",
  color: "var(--text-primary)", fontSize: "14px", outline: "none",
  fontFamily: "var(--font-montserrat)",
};

export default function AdminBannersPage() {
  const [banners, setBanners] = useState<Banner[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<Banner | null>(null);
  const [form, setForm] = useState<BannerForm>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);

  const fetch = () => {
    setLoading(true);
    adminApi.banners.list()
      .then((res) => setBanners(Array.isArray(res) ? res : []))
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetch(); }, []);

  const openAdd = () => { setEditing(null); setForm(emptyForm); setShowModal(true); };
  const openEdit = (b: Banner) => {
    setEditing(b);
    setForm({ title: b.title, image_url: b.image_url, link_url: b.link_url, is_active: b.is_active, sort_order: b.sort_order });
    setShowModal(true);
  };

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const res = await adminApi.upload(file);
      setForm((f) => ({ ...f, image_url: res.url }));
    } catch { alert("Upload gagal"); }
    finally { setUploading(false); }
  };

  const handleSave = async () => {
    if (!form.title.trim()) { alert("Judul wajib diisi"); return; }
    setSaving(true);
    try {
      if (editing) await adminApi.banners.update(editing.id, form);
      else await adminApi.banners.create(form);
      setShowModal(false);
      fetch();
    } catch (e) { alert("Gagal simpan"); }
    finally { setSaving(false); }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Hapus banner ini?")) return;
    await adminApi.banners.delete(id);
    fetch();
  };

  const toggleActive = async (b: Banner) => {
    await adminApi.banners.update(b.id, { is_active: !b.is_active });
    setBanners((prev) => prev.map((x) => x.id === b.id ? { ...x, is_active: !x.is_active } : x));
  };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
        <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px" }}>Banner</h1>
        <button onClick={openAdd}
          style={{ display: "flex", alignItems: "center", gap: "8px", padding: "10px 18px", borderRadius: "10px", background: "var(--accent)", color: "#fff", border: "none", cursor: "pointer", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px" }}>
          <Plus size={16} /> Tambah Banner
        </button>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))", gap: "16px" }}>
        {loading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <div key={i} style={{ borderRadius: "14px", background: "var(--bg-card)", border: "1px solid var(--border)", overflow: "hidden" }}>
              <div style={{ aspectRatio: "16/5", background: "var(--surface)" }} />
              <div style={{ padding: "14px", display: "flex", flexDirection: "column", gap: "8px" }}>
                <div style={{ height: "14px", borderRadius: "4px", background: "var(--surface)", width: "60%" }} />
              </div>
            </div>
          ))
        ) : banners.map((b) => (
          <div key={b.id} style={{ borderRadius: "14px", background: "var(--bg-card)", border: "1px solid var(--border)", overflow: "hidden" }}>
            <div style={{ aspectRatio: "16/5", background: "var(--surface)", position: "relative", overflow: "hidden" }}>
              {b.image_url && <img src={mediaUrl(b.image_url)} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />}
              {!b.is_active && (
                <div style={{ position: "absolute", inset: 0, background: "rgba(0,0,0,0.4)", display: "flex", alignItems: "center", justifyContent: "center" }}>
                  <span style={{ color: "#fff", fontSize: "12px", background: "rgba(0,0,0,0.6)", padding: "4px 10px", borderRadius: "20px" }}>Nonaktif</span>
                </div>
              )}
            </div>
            <div style={{ padding: "14px" }}>
              <p style={{ fontSize: "14px", fontWeight: 600, color: "var(--text-primary)", marginBottom: "4px" }}>{b.title}</p>
              <p style={{ fontSize: "12px", color: "var(--text-muted)", marginBottom: "12px" }}>Urutan: {b.sort_order}</p>
              <div style={{ display: "flex", gap: "8px" }}>
                <button onClick={() => openEdit(b)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "var(--surface)", color: "var(--text-primary)", border: "none", cursor: "pointer", fontSize: "13px" }}>
                  <Pencil size={13} /> Edit
                </button>
                <button onClick={() => toggleActive(b)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "var(--surface)", color: b.is_active ? "#10b981" : "var(--text-muted)", border: "none", cursor: "pointer", fontSize: "13px" }}>
                  {b.is_active ? <ToggleRight size={15} /> : <ToggleLeft size={15} />}
                  {b.is_active ? "Aktif" : "Nonaktif"}
                </button>
                <button onClick={() => handleDelete(b.id)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "#fee2e2", color: "#dc2626", border: "none", cursor: "pointer", fontSize: "13px" }}>
                  <Trash2 size={13} />
                </button>
              </div>
            </div>
          </div>
        ))}
        {!loading && banners.length === 0 && (
          <div style={{ gridColumn: "1/-1", padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>
            Belum ada banner. Tambah banner pertamamu!
          </div>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div style={{ position: "fixed", inset: 0, zIndex: 200, display: "flex", alignItems: "center", justifyContent: "center", padding: "16px" }}
          onClick={(e) => { if (e.target === e.currentTarget) setShowModal(false); }}>
          <div style={{ position: "absolute", inset: 0, background: "var(--overlay)" }} />
          <div style={{ position: "relative", background: "var(--bg-card)", borderRadius: "16px", padding: "28px", width: "100%", maxWidth: "480px", border: "1px solid var(--border)", zIndex: 1 }}>
            <h2 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "18px", marginBottom: "20px" }}>
              {editing ? "Edit Banner" : "Tambah Banner"}
            </h2>
            <div style={{ display: "flex", flexDirection: "column", gap: "14px" }}>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Judul *</label>
                <input value={form.title} onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))} style={inputStyle} placeholder="Judul banner" />
              </div>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Gambar Banner</label>
                <div style={{ display: "flex", gap: "8px" }}>
                  <input value={form.image_url} onChange={(e) => setForm((f) => ({ ...f, image_url: e.target.value }))} style={{ ...inputStyle, flex: 1 }} placeholder="/uploads/banner.jpg" />
                  <label style={{ padding: "9px 14px", background: "var(--surface)", borderRadius: "8px", cursor: "pointer", fontSize: "13px", whiteSpace: "nowrap", border: "1px solid var(--border)" }}>
                    {uploading ? "..." : "Upload"}
                    <input type="file" accept="image/*" style={{ display: "none" }} onChange={handleUpload} />
                  </label>
                </div>
                {form.image_url && <img src={mediaUrl(form.image_url)} alt="" style={{ marginTop: "8px", maxHeight: "80px", borderRadius: "6px" }} />}
              </div>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Link URL</label>
                <input value={form.link_url} onChange={(e) => setForm((f) => ({ ...f, link_url: e.target.value }))} style={inputStyle} placeholder="https://..." />
              </div>
              <div style={{ display: "flex", gap: "12px" }}>
                <div style={{ flex: 1 }}>
                  <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Urutan</label>
                  <input type="number" value={form.sort_order} onChange={(e) => setForm((f) => ({ ...f, sort_order: Number(e.target.value) }))} style={{ ...inputStyle }} min={0} />
                </div>
                <div>
                  <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Status</label>
                  <button type="button" onClick={() => setForm((f) => ({ ...f, is_active: !f.is_active }))}
                    style={{ padding: "9px 16px", borderRadius: "8px", border: "1px solid var(--border)", background: form.is_active ? "#d1fae5" : "var(--surface)", color: form.is_active ? "#065f46" : "var(--text-secondary)", cursor: "pointer", fontSize: "13px", fontWeight: 500 }}>
                    {form.is_active ? "✓ Aktif" : "Nonaktif"}
                  </button>
                </div>
              </div>
            </div>
            <div style={{ display: "flex", gap: "10px", marginTop: "24px" }}>
              <button onClick={handleSave} disabled={saving}
                style={{ flex: 1, padding: "11px", borderRadius: "10px", background: "var(--accent)", color: "#fff", border: "none", cursor: saving ? "not-allowed" : "pointer", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px", opacity: saving ? 0.7 : 1 }}>
                {saving ? "Menyimpan..." : "Simpan"}
              </button>
              <button onClick={() => setShowModal(false)}
                style={{ padding: "11px 20px", borderRadius: "10px", background: "var(--surface)", border: "none", cursor: "pointer", fontSize: "14px" }}>
                Batal
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
```

---

## 🔟 ADMIN ADS PAGE — `frontend/app/(admin)/ads/page.tsx`

```bash
mkdir -p "frontend/app/(admin)/ads"
```

```typescript
// frontend/app/(admin)/ads/page.tsx
"use client";
import { useEffect, useState } from "react";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { adminApi } from "@/lib/api";
import type { Ad } from "@/lib/api";

const inputStyle: React.CSSProperties = {
  width: "100%", padding: "9px 12px", borderRadius: "8px",
  border: "1px solid var(--border)", background: "var(--bg-secondary)",
  color: "var(--text-primary)", fontSize: "14px", outline: "none",
  fontFamily: "var(--font-montserrat)",
};

type AdForm = { name: string; ad_code: string; device_type: string; is_active: boolean };
const emptyForm: AdForm = { name: "", ad_code: "", device_type: "all", is_active: true };

export default function AdminAdsPage() {
  const [ads, setAds] = useState<Ad[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<Ad | null>(null);
  const [form, setForm] = useState<AdForm>(emptyForm);
  const [saving, setSaving] = useState(false);

  const fetchAds = () => {
    setLoading(true);
    adminApi.banners.list() // reuse pattern — ganti ke adminApi untuk ads
      .catch(() => [])
      .finally(() => setLoading(false));
    // Karena adminApi.ads belum ada di lib/api.ts, tambahkan ini:
    fetch("http://localhost:8080/api/admin/ads", { credentials: "include" })
      .then((r) => r.json())
      .then((r) => setAds(r.data ?? []))
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchAds(); }, []);

  const openAdd = () => { setEditing(null); setForm(emptyForm); setShowModal(true); };
  const openEdit = (a: Ad) => { setEditing(a); setForm({ name: a.name, ad_code: a.ad_code, device_type: a.device_type, is_active: a.is_active }); setShowModal(true); };

  const handleSave = async () => {
    if (!form.name.trim()) { alert("Nama wajib diisi"); return; }
    setSaving(true);
    const base = "http://localhost:8080/api/admin/ads";
    try {
      if (editing) {
        await fetch(`${base}/${editing.id}`, { method: "PUT", credentials: "include", headers: { "Content-Type": "application/json" }, body: JSON.stringify(form) });
      } else {
        await fetch(base, { method: "POST", credentials: "include", headers: { "Content-Type": "application/json" }, body: JSON.stringify(form) });
      }
      setShowModal(false); fetchAds();
    } catch { alert("Gagal simpan"); }
    finally { setSaving(false); }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Hapus iklan ini?")) return;
    await fetch(`http://localhost:8080/api/admin/ads/${id}`, { method: "DELETE", credentials: "include" });
    fetchAds();
  };

  const deviceLabel: Record<string, string> = { all: "Semua", mobile: "Mobile", desktop: "Desktop" };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
        <div>
          <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px" }}>Manajemen Iklan</h1>
          <p style={{ fontSize: "13px", color: "var(--text-muted)" }}>1 iklan per device per sesi</p>
        </div>
        <button onClick={openAdd}
          style={{ display: "flex", alignItems: "center", gap: "8px", padding: "10px 18px", borderRadius: "10px", background: "var(--accent)", color: "#fff", border: "none", cursor: "pointer", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px" }}>
          <Plus size={16} /> Tambah Iklan
        </button>
      </div>

      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", overflow: "hidden" }}>
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr style={{ background: "var(--bg-secondary)" }}>
              {["Nama", "Device", "Status", "Aksi"].map((h) => (
                <th key={h} style={{ padding: "12px 16px", textAlign: "left", fontSize: "12px", fontWeight: 600, color: "var(--text-muted)", textTransform: "uppercase" }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={4} style={{ padding: "32px", textAlign: "center", color: "var(--text-muted)" }}>Memuat...</td></tr>
            ) : ads.map((a) => (
              <tr key={a.id} style={{ borderTop: "1px solid var(--border)" }}>
                <td style={{ padding: "14px 16px", fontWeight: 500, fontSize: "14px" }}>{a.name}</td>
                <td style={{ padding: "14px 16px" }}>
                  <span style={{ padding: "3px 10px", borderRadius: "20px", fontSize: "12px", fontWeight: 600, background: "var(--surface)", color: "var(--text-secondary)" }}>
                    {deviceLabel[a.device_type] ?? a.device_type}
                  </span>
                </td>
                <td style={{ padding: "14px 16px" }}>
                  <span style={{ padding: "3px 10px", borderRadius: "20px", fontSize: "12px", fontWeight: 600, background: a.is_active ? "#d1fae5" : "#fef3c7", color: a.is_active ? "#065f46" : "#92400e" }}>
                    {a.is_active ? "Aktif" : "Nonaktif"}
                  </span>
                </td>
                <td style={{ padding: "14px 16px" }}>
                  <div style={{ display: "flex", gap: "8px" }}>
                    <button onClick={() => openEdit(a)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "var(--surface)", border: "none", cursor: "pointer", fontSize: "13px", color: "var(--text-primary)" }}>
                      <Pencil size={13} /> Edit
                    </button>
                    <button onClick={() => handleDelete(a.id)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "#fee2e2", color: "#dc2626", border: "none", cursor: "pointer", fontSize: "13px" }}>
                      <Trash2 size={13} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {!loading && ads.length === 0 && (
              <tr><td colSpan={4} style={{ padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>Belum ada iklan.</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Modal */}
      {showModal && (
        <div style={{ position: "fixed", inset: 0, zIndex: 200, display: "flex", alignItems: "center", justifyContent: "center", padding: "16px" }}
          onClick={(e) => { if (e.target === e.currentTarget) setShowModal(false); }}>
          <div style={{ position: "absolute", inset: 0, background: "var(--overlay)" }} />
          <div style={{ position: "relative", background: "var(--bg-card)", borderRadius: "16px", padding: "28px", width: "100%", maxWidth: "520px", border: "1px solid var(--border)", zIndex: 1 }}>
            <h2 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "18px", marginBottom: "20px" }}>
              {editing ? "Edit Iklan" : "Tambah Iklan"}
            </h2>
            <div style={{ display: "flex", flexDirection: "column", gap: "14px" }}>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Nama Iklan *</label>
                <input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} style={inputStyle} placeholder="Mis: Google AdSense Utama" />
              </div>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Target Device</label>
                <select value={form.device_type} onChange={(e) => setForm((f) => ({ ...f, device_type: e.target.value }))} style={inputStyle}>
                  <option value="all">Semua Device</option>
                  <option value="mobile">Mobile Only</option>
                  <option value="desktop">Desktop Only</option>
                </select>
              </div>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Ad Code (HTML)</label>
                <textarea value={form.ad_code} onChange={(e) => setForm((f) => ({ ...f, ad_code: e.target.value }))} rows={5}
                  placeholder='<script async src="https://..."></script>\n<ins class="adsbygoogle" ...></ins>'
                  style={{ ...inputStyle, resize: "vertical", fontFamily: "monospace", fontSize: "12px" }} />
              </div>
              <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
                <button type="button" onClick={() => setForm((f) => ({ ...f, is_active: !f.is_active }))}
                  style={{ width: "44px", height: "24px", borderRadius: "12px", background: form.is_active ? "var(--accent)" : "var(--surface)", border: "none", cursor: "pointer", position: "relative", transition: "background 0.2s", flexShrink: 0 }}>
                  <span style={{ position: "absolute", top: "2px", left: form.is_active ? "22px" : "2px", width: "20px", height: "20px", borderRadius: "50%", background: "#fff", transition: "left 0.2s" }} />
                </button>
                <span style={{ fontSize: "13px", color: "var(--text-secondary)" }}>{form.is_active ? "Aktif — iklan akan ditayangkan" : "Nonaktif"}</span>
              </div>
            </div>
            <div style={{ display: "flex", gap: "10px", marginTop: "24px" }}>
              <button onClick={handleSave} disabled={saving}
                style={{ flex: 1, padding: "11px", borderRadius: "10px", background: "var(--accent)", color: "#fff", border: "none", cursor: saving ? "not-allowed" : "pointer", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px", opacity: saving ? 0.7 : 1 }}>
                {saving ? "Menyimpan..." : "Simpan"}
              </button>
              <button onClick={() => setShowModal(false)}
                style={{ padding: "11px 20px", borderRadius: "10px", background: "var(--surface)", border: "none", cursor: "pointer", fontSize: "14px" }}>
                Batal
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
```

---

## ✅ CHECKLIST FINAL STEP 7

```
[ ] middleware.ts          → proteksi route admin
[ ] (admin)/layout.tsx     → sidebar + topbar admin
[ ] (admin)/dashboard/     → stat cards + tabel video
[ ] (admin)/videos/        → list video + hapus
[ ] (admin)/videos/new/    → form tambah video
[ ] (admin)/videos/edit/   → form edit video
[ ] components/admin/VideoForm.tsx   → shared form + upload
[ ] components/video/BannerSlider.tsx
[ ] components/ui/AdSlot.tsx
[ ] (public)/search/page.tsx
[ ] (admin)/banners/page.tsx
[ ] (admin)/ads/page.tsx
[ ] (public)/page.tsx      → diupdate pakai BannerSlider + AdSlot
```

---

## 🚀 TEST AKHIR

```bash
# Terminal 1
cd backend && go run cmd/server/main.go

# Terminal 2
cd frontend && npm run dev
```

| URL | Yang Harus Tampil |
|---|---|
| `localhost:3000` | Homepage + banner slider + grid video |
| `localhost:3000/watch/1` | Player + info video + sidebar related |
| `localhost:3000/search?q=go` | Hasil pencarian video |
| `localhost:3000/admin/login` | Form login |
| `localhost:3000/admin/dashboard` | Redirect ke login (belum auth) |
| Login dengan `admin` / `admin123` | Masuk ke dashboard |
| `localhost:3000/admin/videos` | Tabel semua video |
| `localhost:3000/admin/videos/new` | Form tambah video dengan upload |
| `localhost:3000/admin/banners` | Grid banner + modal edit |
| `localhost:3000/admin/ads` | Tabel iklan + modal edit |
