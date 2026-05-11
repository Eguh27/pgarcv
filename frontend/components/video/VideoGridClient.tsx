"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { VideoCard } from "./VideoCard";
import type { Video } from "@/lib/api";

export function VideoGridClient({ initialVideos }: { initialVideos: Video[] }) {
  const [videos] = useState(initialVideos);
  const router = useRouter();

  return (
    <>
      <style>{`
        .video-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
          gap: 20px;
        }
        @media (max-width: 640px) {
          .video-grid {
            grid-template-columns: repeat(2, 1fr);
            gap: 10px;
          }
        }
      `}</style>
      <div className="video-grid">
        {videos.map((v) => (
          <VideoCard
            key={v.id}
            video={v}
            onClick={(id) => router.push(`/watch/${id}`)}
          />
        ))}
        {videos.length === 0 && (
          <p
            style={{
              color: "var(--text-muted)",
              gridColumn: "1/-1",
              textAlign: "center",
              padding: "48px 0",
              fontFamily: "var(--font-montserrat)",
            }}
          >
            Belum ada video yang dipublikasikan.
          </p>
        )}
      </div>
    </>
  );
}

