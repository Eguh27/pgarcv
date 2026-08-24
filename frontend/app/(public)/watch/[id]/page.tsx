import { api, mediaUrl, formatViews, type Video } from "@/lib/api";
import { notFound } from "next/navigation";
import { RelatedVideos } from "@/components/video/RelatedVideos";
import { VideoPlayerClient } from "@/components/video/VideoPlayerClient";
import styles from "./WatchPage.module.css";
import { WatchActions } from "./WatchActions";

type Params = Promise<{ id: string }>;

function toVideoArray(res: unknown): Video[] {
  const arr = Array.isArray(res)
    ? res
    : Array.isArray((res as Record<string, unknown>)?.data)
    ? (res as Record<string, unknown>).data
    : Array.isArray((res as Record<string, unknown>)?.videos)
    ? (res as Record<string, unknown>).videos
    : [];
  return (arr as unknown[]).filter(
    (x): x is Video =>
      Boolean(x) && typeof x === "object" && typeof (x as Video).id === "number"
  );
}

export default async function WatchPage({ params }: { params: Params }) {
  const { id } = await params;

  let video: Video;
  let relatedRes: unknown;
  try {
    [video, relatedRes] = await Promise.all([
      api.videos.get(Number(id)),
      api.videos.list(1),
    ]);
  } catch {
    notFound();
  }

  const relatedVideos = toVideoArray(relatedRes)
    .filter((v) => v.id !== video!.id)
    .slice(0, 8);

  return (
    <div className={styles.layout}>
      <div className={styles.main}>
        <div style={{ marginBottom: "16px" }}>
          <VideoPlayerClient
            videoUrl={
              video!.video_url
                ? mediaUrl(video!.video_url)
                : video!.hls_status === "done"
                  ? `/api/hls/${video!.id}/index.m3u8`
                  : ""
            }
            title={video!.title}
          />
        </div>
        <h1
          style={{
            fontFamily: "var(--font-poppins)",
            fontWeight: 700,
            fontSize: "22px",
            marginBottom: "4px",
          }}
        >
          {video!.title}
        </h1>
        <p style={{ fontSize: "13px", color: "var(--text-secondary)", marginBottom: "8px" }}>
          {video!.subtitle}
        </p>
        <WatchActions
          videoId={video!.id}
          title={video!.title}
          allowDownload={Boolean(video!.allow_download)}
        />
        <p style={{ fontSize: "13px", color: "var(--text-muted)", marginBottom: "16px" }}>
          {formatViews(video!.views)} penonton ·{" "}
          {new Date(video!.created_at).toLocaleDateString("id-ID", {
            day: "numeric",
            month: "long",
            year: "numeric",
          })}
        </p>
        <div style={{ background: "var(--bg-secondary)", borderRadius: "12px", padding: "16px" }}>
          <p
            style={{
              fontSize: "14px",
              lineHeight: 1.7,
              color: "var(--text-primary)",
              whiteSpace: "pre-wrap",
            }}
          >
            {video!.description || "Tidak ada deskripsi."}
          </p>
        </div>
      </div> {/* ← div main ditutup di sini */}

      <div className={styles.sidebar}>
        <h3
          style={{
            fontFamily: "var(--font-poppins)",
            fontWeight: 600,
            fontSize: "16px",
            marginBottom: "12px",
          }}
        >
          Video Lainnya
        </h3>
        <RelatedVideos videos={relatedVideos} />
      </div>
    </div>
  );
}