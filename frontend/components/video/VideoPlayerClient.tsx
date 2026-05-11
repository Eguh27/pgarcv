"use client";

import { useEffect, useRef, useCallback } from "react";
import Hls from "hls.js";

interface Props {
  videoUrl: string;
  title: string;
}

export function VideoPlayerClient({ videoUrl, title }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const hlsRef = useRef<Hls | null>(null);

  // ✅ FIX AbortError saat backend restart/unmount: buat cleanup aman dan tidak memicu abort yang tidak tertangani
  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    // Cleanup HLS instance sebelumnya
    if (hlsRef.current) {
      try {
        hlsRef.current.destroy();
      } catch {
        // ignore AbortError-like exceptions during cleanup
      }
      hlsRef.current = null;
    }

    if (!videoUrl) return;

    const isHLS = videoUrl.includes(".m3u8");

    if (isHLS && Hls.isSupported()) {
      const hls = new Hls({ enableWorker: true });
      hls.loadSource(videoUrl);
      hls.attachMedia(video);
      hlsRef.current = hls;
    } else if (isHLS && video.canPlayType("application/vnd.apple.mpegurl")) {
      // Safari native HLS
      video.src = videoUrl;
      video.load();
    } else {
      // Video biasa (mp4, webm, dll)
      video.pause();
      video.removeAttribute("src");
      video.load();
      video.src = videoUrl;
      video.load();
    }

    return () => {
      // Cleanup HLS: panggil destroy di microtask agar error/abort tidak terpropagasi ke React unmount
      const current = hlsRef.current;
      hlsRef.current = null;

      if (current) {
        queueMicrotask(() => {
          try {
            current.destroy();
          } catch {
            // ignore AbortError-like exceptions
          }
        });
      }

      // Cleanup video element
      try {
        video.pause();
        video.removeAttribute("src");
        video.load();
      } catch {
        // ignore
      }
    };
  }, [videoUrl]);

  // ✅ FIX SPACE = FULLSCREEN: intercept keyboard di container
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLDivElement>) => {
      const video = videoRef.current;
      if (!video) return;

      switch (e.key) {
        case " ":
        case "k": // YouTube shortcut
          // Cegah scroll halaman dan fullscreen browser default
          e.preventDefault();
          e.stopPropagation();
          if (video.paused) {
            video.play();
          } else {
            video.pause();
          }
          break;

        case "f":
        case "F":
          e.preventDefault();
          // Fullscreen manual — hanya saat user sengaja tekan F
          if (!document.fullscreenElement) {
            containerRef.current?.requestFullscreen().catch(() => {});
          } else {
            document.exitFullscreen().catch(() => {});
          }
          break;

        case "ArrowRight":
          e.preventDefault();
          video.currentTime = Math.min(
            video.currentTime + 10,
            video.duration || 0
          );
          break;

        case "ArrowLeft":
          e.preventDefault();
          video.currentTime = Math.max(video.currentTime - 10, 0);
          break;

        case "ArrowUp":
          e.preventDefault();
          video.volume = Math.min(video.volume + 0.1, 1);
          break;

        case "ArrowDown":
          e.preventDefault();
          video.volume = Math.max(video.volume - 0.1, 0);
          break;

        case "m":
        case "M":
          e.preventDefault();
          video.muted = !video.muted;
          break;

        default:
          break;
      }
    },
    []
  );

  if (!videoUrl) {
    return (
      <div
        style={{
          aspectRatio: "16/9",
          background: "var(--surface)",
          borderRadius: "12px",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: "var(--text-muted)",
          fontSize: "14px",
        }}
      >
        Video belum tersedia
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      // tabIndex supaya div bisa menerima keyboard event
      tabIndex={0}
      onKeyDown={handleKeyDown}
      style={{
        aspectRatio: "16/9",
        background: "#000",
        borderRadius: "12px",
        overflow: "hidden",
        outline: "none", // hilangkan focus ring default
        position: "relative",
      }}
      // Auto-focus container saat mount supaya keyboard langsung aktif
      // tanpa perlu user klik dulu
      onClick={() => containerRef.current?.focus()}
    >
      <video
        ref={videoRef}
        controls
        // ❌ Hapus autoPlay — ini salah satu penyebab suara dobel
        // autoPlay
        style={{
          width: "100%",
          height: "100%",
          display: "block",
        }}
        aria-label={title}
        // Preload metadata saja dulu, bukan full video
        preload="metadata"
        // Cegah Space di dalam video element native trigger fullscreen
        onKeyDown={(e) => {
          if (e.key === " " || e.key === "f") {
            e.preventDefault();
            e.stopPropagation();
          }
        }}
      />

      {/* Hint keyboard shortcut — muncul sebentar lalu fade */}
      <style>{`
        .player-hint {
          position: absolute;
          bottom: 60px;
          left: 50%;
          transform: translateX(-50%);
          background: rgba(0,0,0,0.7);
          color: #fff;
          padding: 6px 14px;
          border-radius: 20px;
          font-size: 12px;
          pointer-events: none;
          opacity: 0;
          animation: hintFade 3s ease forwards;
          white-space: nowrap;
        }
        @keyframes hintFade {
          0% { opacity: 0; }
          20% { opacity: 1; }
          70% { opacity: 1; }
          100% { opacity: 0; }
        }
      `}</style>
    </div>
  );
}
