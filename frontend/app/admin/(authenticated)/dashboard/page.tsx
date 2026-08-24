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
      .then((res: unknown) => {
        // Handle semua kemungkinan struktur response
        let data: VideoType[] = [];
        if (Array.isArray(res)) {
          data = res as VideoType[];
        } else if (res && typeof res === "object") {
          const r = res as Record<string, unknown>;
          if (Array.isArray(r.data)) data = r.data as VideoType[];
          else if (Array.isArray(r.videos)) data = r.videos as VideoType[];
        }
        setVideos(data);
      })
      .catch((err) => {
        console.error("Dashboard fetch error:", err);
        setVideos([]);
      })
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
          <p style={{ fontSize: "13px", color: "var(--text-muted)", marginTop: "2px" }}>Selamat datang kembali</p>
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
                {loading ? "-" : s.value}
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
          <Link href="/admin/videos" style={{ fontSize: "13px", color: "var(--accent)", textDecoration: "none" }}>Lihat semua -&gt;</Link>
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
