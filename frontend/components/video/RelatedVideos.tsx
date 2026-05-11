"use client";

import { useRouter } from "next/navigation";
import type { Video } from "@/lib/api";
import { VideoCard } from "./VideoCard";

export function RelatedVideos({ videos }: { videos: Video[] }) {
  const router = useRouter();

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
      {videos.map((video) => (
        <VideoCard
          key={video.id}
          video={video}
          onClick={(id) => router.push(`/watch/${id}`)}
        />
      ))}
    </div>
  );
}
