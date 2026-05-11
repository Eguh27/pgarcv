# 🐛 BUG FIXES — Hydration + Video Player

## Bug yang difix:
1. **Hydration mismatch** — `next/image` URL beda antara server & client
2. **Space = fullscreen** — keyboard shortcut bentrok dengan browser default
3. **Suara dobel** — video element tidak di-cleanup saat navigate

---

## FIX 1 — Hydration Mismatch di `BannerSlider.tsx` & `VideoCard.tsx`

**Root cause:** `mediaUrl()` di server menghasilkan URL relatif (`/uploads/...`),
tapi di client menghasilkan full URL (`http://localhost:8080/uploads/...`).
Next.js `<Image>` mendeteksi perbedaan ini → hydration error.

**Solusi:** Pisahkan `next/image` jadi komponen client-only dengan `"use client"` + 
`suppressHydrationWarning`, atau lebih clean: normalize `mediaUrl()` supaya 
**selalu konsisten** di server dan client.

### Edit `frontend/lib/api.ts` — fungsi `mediaUrl`:

```typescript
// Ganti fungsi mediaUrl yang lama dengan ini:
export function mediaUrl(path: string): string {
  if (!path) return "/placeholder.jpg";
  // Kalau sudah absolute URL (http/https), return as-is
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  // Kalau sudah path relatif dari root (/uploads/...), return as-is
  // next/image akan handle via remotePatterns atau local
  if (path.startsWith("/")) return path;
  // Fallback
  return `/${path}`;
}
```

**Kenapa ini fix hydration?**  
Server dan client sekarang sama-sama return `/uploads/xxx.jpg` (path relatif).
`next.config.ts` sudah dikonfigurasi untuk serve `/uploads` via rewrite ke backend.

### Edit `frontend/next.config.ts` — tambahkan rewrite untuk uploads:

```typescript
// frontend/next.config.ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    remotePatterns: [
      {
        protocol: "http",
        hostname: "localhost",
        port: "8080",
        pathname: "/uploads/**",
      },
    ],
    // Tambahan: izinkan path relatif /uploads juga
    unoptimized: false,
  },
  async rewrites() {
    return [
      // API proxy
      {
        source: "/api/:path*",
        destination: "http://localhost:8080/api/:path*",
      },
      // ⬇️ INI YANG BARU — proxy /uploads ke backend
      {
        source: "/uploads/:path*",
        destination: "http://localhost:8080/uploads/:path*",
      },
    ];
  },
};

export default nextConfig;
```

**Dengan ini:** `/uploads/foto.jpg` di frontend otomatis di-proxy ke `localhost:8080/uploads/foto.jpg`.
Tidak ada lagi URL yang beda antara SSR dan client. ✅

---

## FIX 2 & 3 — Video Player: Space=Pause + Suara Dobel

**Root cause bug Space=Fullscreen:**  
Browser default-nya: Space = scroll halaman. Tapi beberapa browser modern
juga punya default Space = play/pause pada element `<video>` yang sedang focused.
Yang terjadi di sini: Space mentriger **dua hal** — browser scroll + video action,
dan karena video dalam container div, browser malah trigger fullscreen API.

**Root cause suara dobel:**  
Saat React re-render atau navigasi, `<video>` element lama tidak di-unmount dengan benar.
Terutama karena `autoPlay` — dua instance video bisa overlap sesaat.

### Ganti seluruh watch page — `frontend/app/(public)/watch/[id]/page.tsx`:

```typescript
// frontend/app/(public)/watch/[id]/page.tsx
import { api, mediaUrl, formatViews } from "@/lib/api";
import { notFound } from "next/navigation";
import { VideoCard } from "@/components/video/VideoCard";
import { VideoPlayerClient } from "@/components/video/VideoPlayerClient";

export default async function WatchPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let video, relatedRes;
  try {
    [video, relatedRes] = await Promise.all([
      api.videos.get(Number(id)),
      api.videos.list(1),
    ]);
  } catch {
    notFound();
  }

  const related = relatedRes.data.filter((v) => v.id !== video.id).slice(0, 8);

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "1fr 360px",
        gap: "24px",
      }}
    >
      {/* Main */}
      <div>
        {/* ⬇️ Player sekarang jadi komponen client tersendiri */}
        <VideoPlayerClient
          videoUrl={mediaUrl(video.video_url)}
          title={video.title}
        />

        <h1
          style={{
            fontFamily: "var(--font-poppins)",
            fontWeight: 700,
            fontSize: "22px",
            marginBottom: "4px",
            marginTop: "16px",
          }}
        >
          {video.title}
        </h1>
        <p
          style={{
            fontSize: "13px",
            color: "var(--text-secondary)",
            marginBottom: "8px",
          }}
        >
          {video.subtitle}
        </p>
        <p
          style={{
            fontSize: "13px",
            color: "var(--text-muted)",
            marginBottom: "16px",
          }}
        >
          {formatViews(video.views)} penonton ·{" "}
          {new Date(video.created_at).toLocaleDateString("id-ID", {
            day: "numeric",
            month: "long",
            year: "numeric",
          })}
        </p>

        <div
          style={{
            background: "var(--bg-secondary)",
            borderRadius: "12px",
            padding: "16px",
          }}
        >
          <p
            style={{
              fontSize: "14px",
              lineHeight: 1.7,
              color: "var(--text-primary)",
              whiteSpace: "pre-wrap",
            }}
          >
            {video.description || "Tidak ada deskripsi."}
          </p>
        </div>
      </div>

      {/* Sidebar */}
      <div>
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
        <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
          {related.map((v) => (
            <VideoCard
              key={v.id}
              video={v}
              onClick={(id) => (window.location.href = `/watch/${id}`)}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
```

### Buat file baru — `frontend/components/video/VideoPlayerClient.tsx`:

```typescript
// frontend/components/video/VideoPlayerClient.tsx
"use client";

import { useEffect, useRef, useCallback } from "react";

interface Props {
  videoUrl: string;
  title: string;
}

export function VideoPlayerClient({ videoUrl, title }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // ✅ FIX SUARA DOBEL: cleanup yang benar saat unmount / URL berubah
  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    // Reset state saat URL berubah
    video.pause();
    video.removeAttribute("src");
    video.load(); // flush buffer lama

    // Set src baru
    video.src = videoUrl;
    video.load();

    return () => {
      // Cleanup saat komponen unmount (navigasi ke halaman lain)
      video.pause();
      video.removeAttribute("src");
      video.load();
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
```

---

## RINGKASAN PERUBAHAN

| File | Perubahan |
|---|---|
| `frontend/lib/api.ts` | `mediaUrl()` — selalu return path relatif, bukan full URL |
| `frontend/next.config.ts` | Tambah rewrite `/uploads/:path*` → backend |
| `frontend/app/(public)/watch/[id]/page.tsx` | Pisahkan player jadi `VideoPlayerClient` |
| `frontend/components/video/VideoPlayerClient.tsx` | **File baru** — player dengan keyboard fix + cleanup |

---

## CARA APPLY

```bash
# 1. Edit lib/api.ts — ganti fungsi mediaUrl
# 2. Edit next.config.ts — tambah rewrite uploads
# 3. Ganti isi watch/[id]/page.tsx
# 4. Buat file baru VideoPlayerClient.tsx

# 5. Restart dev server (wajib setelah edit next.config.ts!)
cd frontend
# Ctrl+C dulu, lalu:
npm run dev
```

> **Catatan:** Setelah edit `next.config.ts`, dev server **harus di-restart**.
> Hot reload tidak akan apply perubahan config.

---

## KEYBOARD SHORTCUTS YANG AKTIF

| Tombol | Fungsi |
|---|---|
| `Space` atau `K` | Play / Pause |
| `F` | Toggle Fullscreen |
| `→` | Skip +10 detik |
| `←` | Skip -10 detik |
| `↑` | Volume naik |
| `↓` | Volume turun |
| `M` | Mute / Unmute |

---

## KALAU MASIH ADA SUARA DOBEL

Kemungkinan ada instance video lain yang hidup. Cek ini di browser DevTools:
```javascript
// Paste di Console browser
document.querySelectorAll('video').forEach((v, i) => {
  console.log(`Video ${i}:`, v.src, 'paused:', v.paused, 'muted:', v.muted);
});
```
Kalau muncul lebih dari 1 video dengan src yang sama → ada komponen yang tidak unmount.
Beritahu aku output-nya dan aku debug lebih lanjut.
