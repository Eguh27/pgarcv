"use client";

import Link from "next/link";
import { useState } from "react";
import useSWR from "swr";
import { Eye, Pencil, Plus, Trash2 } from "lucide-react";
import { adminApi, formatViews, mediaUrl } from "@/lib/api";
import type { Video } from "@/lib/api";

function normalizeVideoList(res: unknown): Video[] {
  let data: unknown[] = [];
  if (Array.isArray(res)) {
    data = res;
  } else if (res !== null && typeof res === "object") {
    const r = res as Record<string, unknown>;
    if (Array.isArray(r.data)) data = r.data;
    else if (Array.isArray(r.videos)) data = r.videos;
  }
  return data.filter(
    (v): v is Video =>
      typeof v === "object" &&
      v !== null &&
      typeof (v as Video).id === "number"
  );
}

const fetcher = async (): Promise<Video[]> =>
  normalizeVideoList(await adminApi.videos.list());

export default function AdminVideosPage() {
  const { data: videos = [], isLoading, error, mutate } = useSWR<Video[]>(
    "admin/videos",
    fetcher,
    { revalidateOnFocus: false }
  );
  const [deletingId, setDeletingId] = useState<number | null>(null);

  const handleDelete = async (id: number, title: string) => {
    if (!confirm(`Hapus video "${title}"?`)) return;
    setDeletingId(id);
    try {
      await adminApi.videos.delete(id);
      await mutate();
    } catch {
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
        <Link
          href="/admin/videos/new"
          style={{
            display: "flex",
            alignItems: "center",
            gap: "8px",
            padding: "10px 18px",
            borderRadius: "10px",
            background: "var(--accent)",
            color: "#fff",
            textDecoration: "none",
            fontFamily: "var(--font-poppins)",
            fontWeight: 600,
            fontSize: "14px",
          }}
        >
          <Plus size={16} /> Tambah Video
        </Link>
      </div>

      {error && (
        <p style={{ marginBottom: "12px", fontSize: "13px", color: "#dc2626" }}>
          Gagal memuat video. Coba refresh halaman.
        </p>
      )}

      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", overflow: "hidden" }}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse" }}>
            <thead>
              <tr style={{ background: "var(--bg-secondary)" }}>
                {["#", "Video", "Views", "Durasi", "Status", "Aksi"].map((h) => (
                  <th key={h} style={{ padding: "12px 16px", textAlign: "left", fontSize: "12px", fontWeight: 600, color: "var(--text-muted)", textTransform: "uppercase" }}>
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={6} style={{ padding: "32px", textAlign: "center", color: "var(--text-muted)" }}>
                    Memuat...
                  </td>
                </tr>
              ) : (
                videos.map((v, idx) => (
                  <tr key={v.id} style={{ borderTop: "1px solid var(--border)" }}>
                    <td style={{ padding: "12px 16px", fontSize: "13px", color: "var(--text-muted)" }}>{idx + 1}</td>
                    <td style={{ padding: "12px 16px" }}>
                      <div style={{ display: "flex", alignItems: "center", gap: "12px" }}>
                        <div style={{ width: "72px", height: "40px", borderRadius: "6px", background: "var(--surface)", overflow: "hidden", flexShrink: 0 }}>
                          {v.thumbnail_url && (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img src={mediaUrl(v.thumbnail_url)} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                          )}
                        </div>
                        <div style={{ minWidth: 0 }}>
                          <p style={{ fontSize: "14px", fontWeight: 500, color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", maxWidth: "240px" }}>{v.title}</p>
                          <p style={{ fontSize: "12px", color: "var(--text-muted)" }}>{v.subtitle || "-"}</p>
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
                      <span
                        style={{
                          padding: "3px 10px",
                          borderRadius: "20px",
                          fontSize: "12px",
                          fontWeight: 600,
                          background: v.is_published ? "#d1fae5" : "#fef3c7",
                          color: v.is_published ? "#065f46" : "#92400e",
                        }}
                      >
                        {v.is_published ? "Publik" : "Draft"}
                      </span>
                    </td>
                    <td style={{ padding: "12px 16px" }}>
                      <div style={{ display: "flex", gap: "8px" }}>
                        <Link href={`/admin/videos/edit/${v.id}`} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "var(--surface)", color: "var(--text-primary)", textDecoration: "none", fontSize: "13px" }}>
                          <Pencil size={13} /> Edit
                        </Link>
                        <button onClick={() => handleDelete(v.id, v.title)} disabled={deletingId === v.id} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "#fee2e2", color: "#dc2626", border: "none", cursor: "pointer", fontSize: "13px" }}>
                          <Trash2 size={13} /> {deletingId === v.id ? "..." : "Hapus"}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
              {!isLoading && videos.length === 0 && (
                <tr>
                  <td colSpan={6} style={{ padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>
                    Belum ada video. <Link href="/admin/videos/new" style={{ color: "var(--accent)" }}>Tambah sekarang</Link>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
