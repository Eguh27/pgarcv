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

  const videos = (() => {
    if (videosRes.status !== "fulfilled") return [];
    const v: unknown = videosRes.value;
    if (Array.isArray(v)) return v;
    if (v && typeof v === "object") {
      const r = v as Record<string, unknown>;
      if (Array.isArray(r.data)) return r.data;
      if (Array.isArray(r.videos)) return r.videos;
    }
    return [];
  })();

  const banners = (() => {
    if (bannersRes.status !== "fulfilled") return [];
    return bannersRes.value;
  })();

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
        <VideoGridClient initialVideos={videos} />
      </section>
    </div>
  );
}
