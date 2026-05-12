"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { VideoGridClient } from "@/components/video/VideoGridClient";
import { BannerSlider } from "@/components/video/BannerSlider";
import { AdSlot } from "@/components/ui/AdSlot";
import type { Video, Banner } from "@/lib/api";

export default function HomePage() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [banners, setBanners] = useState<Banner[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.allSettled([
      api.videos.list(1),
      api.banners.list(),
    ]).then(([videosRes, bannersRes]) => {
      if (videosRes.status === "fulfilled") {
        const v = videosRes.value as unknown;
        if (Array.isArray(v)) setVideos(v as Video[]);
        else if (v && typeof v === "object") {
          const r = v as Record<string, unknown>;
          if (Array.isArray(r.data)) setVideos(r.data as Video[]);
        }
      }
      if (bannersRes.status === "fulfilled") {
        const b = bannersRes.value;
        if (Array.isArray(b)) setBanners(b as Banner[]);
      }
    }).finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <section style={{ marginBottom: "20px" }}>
        <BannerSlider banners={banners} />
      </section>


      <AdSlot position="top" />
      <section>
        <h2 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "18px", marginBottom: "12px", color: "var(--text-primary)" }}>
          Video Terbaru
        </h2>
        {loading ? (
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))", gap: "20px" }}>
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} style={{ borderRadius: "12px", overflow: "hidden", background: "var(--bg-card)", border: "1px solid var(--border)" }}>
                <div style={{ aspectRatio: "16/9", background: "var(--surface)" }} />
                <div style={{ padding: "12px", display: "flex", flexDirection: "column", gap: "8px" }}>
                  <div style={{ height: "16px", borderRadius: "4px", background: "var(--surface)", width: "80%" }} />
                  <div style={{ height: "12px", borderRadius: "4px", background: "var(--surface)", width: "50%" }} />
                </div>
              </div>
            ))}
          </div>
        ) : (
          <VideoGridClient initialVideos={videos} />
        )}
      </section>
    </div>
  );
}