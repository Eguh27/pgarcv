"use client";
import { useRef, useState, useEffect, useCallback } from "react";
import Image from "next/image";
import { Eye, Clock } from "lucide-react";
import type { Video } from "@/lib/api";
import { mediaUrl, formatDuration, formatViews } from "@/lib/api";

// Manager global: maks 3 preview aktif
const activeVideos = new Set<HTMLVideoElement>();
const MAX_ACTIVE = 3;

interface Props {
  video: Video;
  onClick?: (id: number) => void;
}

export function VideoCard({ video, onClick }: Props) {
  const [isHovered, setIsHovered] = useState(false);
  const [showPreview, setShowPreview] = useState(false);
  const [imgLoaded, setImgLoaded] = useState(false);

  const videoRef = useRef<HTMLVideoElement | null>(null);
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const cardRef = useRef<HTMLDivElement | null>(null);

  // IntersectionObserver: stop video saat keluar viewport
  useEffect(() => {
    const el = cardRef.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry.isIntersecting && videoRef.current) {
          try {
            videoRef.current.pause();
          } catch {}
          activeVideos.delete(videoRef.current);
          setShowPreview(false);
        }
      },
      { threshold: 0.3 }
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  const handleMouseEnter = useCallback(() => {
    setIsHovered(true);
    if (!video.preview_url) return;

    hoverTimer.current = setTimeout(() => {
      if (activeVideos.size >= MAX_ACTIVE) {
        const oldest = activeVideos.values().next().value;
        if (oldest) {
          try {
            oldest.pause();
          } catch {}
          activeVideos.delete(oldest);
        }
      }

      setShowPreview(true);
      if (videoRef.current) {
        try {
          videoRef.current.play().catch(() => {});
        } catch {}
        activeVideos.add(videoRef.current);
      }
    }, 800);
  }, [video.preview_url]);

  const handleMouseLeave = useCallback(() => {
    setIsHovered(false);
    if (hoverTimer.current) clearTimeout(hoverTimer.current);

    setShowPreview(false);

    if (videoRef.current) {
      try {
        videoRef.current.pause();
        videoRef.current.currentTime = 0;
      } catch {}
      activeVideos.delete(videoRef.current);
    }
  }, []);

  // Cleanup
  useEffect(() => {
    const node = videoRef.current;
    return () => {
      if (hoverTimer.current) clearTimeout(hoverTimer.current);
      if (node) activeVideos.delete(node);
    };
  }, []);

  return (
    <>
      <style>{`
        .video-card-title {
          font-family: var(--font-poppins);
          font-weight: 600;
          font-size: 14px;
          color: var(--text-primary);
          line-height: 1.4;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
          margin-bottom: 4px;
        }
        .video-card-subtitle {
          font-size: 12px;
          color: var(--text-secondary);
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          margin-bottom: 6px;
        }
        .video-card-meta {
          display: flex;
          gap: 8px;
          align-items: center;
          flex-wrap: wrap;
        }
        .video-card-meta span {
          display: flex;
          align-items: center;
          gap: 3px;
          font-size: 11px;
          color: var(--text-muted);
        }
        @media (max-width: 640px) {
          .video-card-title { font-size: 12px; }
          .video-card-subtitle { font-size: 11px; }
          .video-card-meta span { font-size: 10px; }
        }
      `}</style>

      <div
        ref={cardRef}
        onClick={() => onClick?.(video.id)}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        style={{
          cursor: "pointer",
          borderRadius: "12px",
          overflow: "hidden",
          background: "var(--bg-card)",
          border: "1px solid var(--border)",
          boxShadow: isHovered ? "var(--shadow-md)" : "var(--shadow)",
          transition: "transform 0.2s, box-shadow 0.2s",
          transform: isHovered ? "translateY(-2px)" : "translateY(0)",
        }}
      >
        {/* Thumbnail area */}
        <div
          style={{
            position: "relative",
            aspectRatio: "16/9",
            background: "var(--surface)",
            overflow: "hidden",
          }}
        >
          <Image
            src={mediaUrl(video.thumbnail_url)}
            alt={video.title}
            fill
            sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
            style={{
              objectFit: "cover",
              opacity: showPreview ? 0 : imgLoaded ? 1 : 0,
              transition: "opacity 0.3s",
            }}
            onLoad={() => setImgLoaded(true)}
          />

          {video.preview_url && (
            <video
              ref={videoRef}
              src={mediaUrl(video.preview_url)}
              muted
              loop
              playsInline
              style={{
                position: "absolute",
                inset: 0,
                width: "100%",
                height: "100%",
                objectFit: "cover",
                opacity: showPreview ? 1 : 0,
                transition: "opacity 0.3s",
              }}
            />
          )}

          {/* Duration badge */}
          <span
            style={{
              position: "absolute",
              bottom: "8px",
              right: "8px",
              background: "rgba(0,0,0,0.8)",
              color: "#fff",
              borderRadius: "4px",
              padding: "2px 6px",
              fontSize: "12px",
              fontWeight: 600,
              fontFamily: "var(--font-poppins)",
            }}
          >
            {formatDuration(video.duration)}
          </span>
        </div>

        {/* Info area */}
        <div style={{ padding: "12px" }}>
          <p className="video-card-title">{video.title}</p>
          <p className="video-card-subtitle">{video.subtitle}</p>

          {/* Deskripsi — muncul saat hover */}
          <div
            style={{
              overflow: "hidden",
              maxHeight: isHovered ? "60px" : "0",
              transition: "max-height 0.3s ease",
              marginBottom: isHovered ? "8px" : "0",
            }}
          >
            <p
              style={{
                fontSize: "11px",
                color: "var(--text-muted)",
                lineHeight: 1.5,
                display: "-webkit-box",
                WebkitLineClamp: 3,
                WebkitBoxOrient: "vertical",
                overflow: "hidden",
              }}
            >
              {video.description}
            </p>
          </div>

          {/* Meta: views & waktu */}
          <div className="video-card-meta">
            <span>
              <Eye size={12} /> {formatViews(video.views)} penonton
            </span>
            <span>
              <Clock size={12} />{" "}
              {new Date(video.created_at).toLocaleDateString("id-ID", {
                day: "numeric",
                month: "short",
                year: "numeric",
              })}
            </span>
          </div>
        </div>
      </div>
    </>
  );
}

