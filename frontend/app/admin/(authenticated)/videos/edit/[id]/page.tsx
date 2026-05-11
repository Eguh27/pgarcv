"use client";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { VideoForm } from "@/components/admin/VideoForm";
import { adminApi } from "@/lib/api";
import type { Video } from "@/lib/api";

export default function EditVideoPage() {
  const { id } = useParams<{ id: string }>();
  const [video, setVideo] = useState<Video | null>(null);
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  useEffect(() => {
    adminApi.videos.get(Number(id)).then(setVideo).catch(() => router.push("/admin/videos"));
  }, [id, router]);

  const handleSubmit = async (data: Parameters<typeof adminApi.videos.update>[1]) => {
    setLoading(true);
    try {
      const updatedVideo = await adminApi.videos.update(Number(id), data);

      // Check if HLS processing is needed (rawPath exists)
      if (data.rawPath) {
        try {
          await adminApi.processHLS(updatedVideo.id, data.rawPath);
          alert("✅ Video diperbarui! HLS sedang diproses di background...");
        } catch (hlsError) {
          console.error("HLS processing failed:", hlsError);
          alert("Video diperbarui, tapi HLS processing gagal. Coba lagi nanti.");
        }
      } else {
        alert("✅ Video diperbarui!");
      }

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
