import { api } from "@/lib/api";
import { Navbar } from "@/components/layout/Navbar";
import { VideoGridClient } from "@/components/video/VideoGridClient";
import { BannerSlider } from "@/components/video/BannerSlider";
import { AdSlot } from "@/components/ui/AdSlot";
import type { Video, Banner } from "@/lib/api";

// ✅ Jangan prerender di build time
export const dynamic = "force-dynamic";

export default async function RootPage() {
  let videos: Video[] = [];
  let banners: Banner[] = [];

  try {
    const [videosRes, bannersRes] = await Promise.allSettled([
      api.videos.list(1),
      api.banners.list(),
    ]);

    videos = (() => {
      if (videosRes.status !== "fulfilled") return [];
      const v: unknown = videosRes.value;
      if (Array.isArray(v)) return v as Video[];
      if (v && typeof v === "object") {
        const r = v as Record<string, unknown>;
        if (Array.isArray(r.data)) return r.data as Video[];
        if (Array.isArray(r.videos)) return r.videos as Video[];
      }
      return [];
    })();

    banners = (() => {
      if (bannersRes.status !== "fulfilled") return [];
      return (bannersRes.value as Banner[]) || [];
    })();
  } catch (err) {
    console.error("Error loading homepage:", err);
  }

  return (
    <>
      <Navbar />
      <main style={{ maxWidth: "1400px", margin: "0 auto", padding: "16px 12px 32px" }}>
        <section style={{ marginBottom: "20px" }}>
          <BannerSlider banners={banners} />
        </section>

        <AdSlot position="top" />

        <section>
          <h2
            style={{
              fontFamily: "var(--font-poppins)",
              fontWeight: 700,
              fontSize: "18px",
              marginBottom: "12px",
              color: "var(--text-primary)",
            }}
          >
            Video Terbaru
          </h2>
          <VideoGridClient initialVideos={videos} />
        </section>
      </main>
    </>
  );
}



 