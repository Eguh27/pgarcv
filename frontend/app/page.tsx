import { Suspense } from "react";
import Link from "next/link";
import { api, mediaUrl, type Video } from "@/lib/api";

async function VideoGrid() {
  // Public videos
  const res = await api.videos.list(1, "");
  const videos = (res?.data ?? []).slice(0, 24);

  return (
    <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(240px, 1fr))", gap: 16 }}>
      {videos.map((v: Video) => (
        <Link
          key={v.id}
          href={`/watch/${v.id}`}
          style={{
            textDecoration: "none",
            color: "inherit",
            border: "1px solid var(--border)",
            borderRadius: 16,
            overflow: "hidden",
            background: "var(--bg-card)",
          }}
        >
          <div style={{ position: "relative", aspectRatio: "16 / 9", background: "#111" }}>
            <img
              src={mediaUrl(v.thumbnail_url || v.preview_url)}
              alt={v.title}
              style={{ width: "100%", height: "100%", objectFit: "cover" }}
            />
          </div>
          <div style={{ padding: 12 }}>
            <div style={{ fontWeight: 700, fontSize: 14, lineHeight: "20px" }}>{v.title}</div>
            {typeof v.views === "number" ? (
              <div style={{ marginTop: 6, fontSize: 12, color: "var(--text-secondary)" }}>{v.views} views</div>
            ) : null}
          </div>
        </Link>
      ))}
    </div>
  );
}

export default function HomePage() {
  return (
    <div>
      <div style={{ margin: "12px 0 18px" }}>
        <h1 style={{ fontSize: 28, fontWeight: 800, marginBottom: 6 }}>Video Terbaik</h1>
        <p style={{ color: "var(--text-secondary)", margin: 0 }}>
          Pilih video untuk mulai menonton.
        </p>
      </div>

      <Suspense fallback={<div style={{ color: "var(--text-secondary)" }}>Memuat video...</div>}>
        {/* Server component fetch */}
        <VideoGrid />
      </Suspense>
    </div>
  );
}

 