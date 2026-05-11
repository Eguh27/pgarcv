"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import Image from "next/image";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { Banner } from "@/lib/api";
import { mediaUrl } from "@/lib/api";

export function BannerSlider({ banners }: { banners: Banner[] }) {
  const [current, setCurrent] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const resetTimer = useCallback(() => {
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = setInterval(() => setCurrent((c) => (c + 1) % banners.length), 5000);
  }, [banners.length]);

  useEffect(() => {
    if (banners.length <= 1) return;
    resetTimer();
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [banners.length, resetTimer]);

  const go = (dir: 1 | -1) => {
    setCurrent((c) => (c + dir + banners.length) % banners.length);
    resetTimer();
  };

  if (!banners.length) return (
    <div style={{ height: "240px", borderRadius: "16px", background: "var(--surface)", display: "flex", alignItems: "center", justifyContent: "center" }}>
      <p style={{ color: "var(--text-muted)" }}>Belum ada banner aktif</p>
    </div>
  );

  return (
    <div style={{ position: "relative", borderRadius: "16px", overflow: "hidden", aspectRatio: "16/5", background: "var(--surface)" }}>
      {/* Slides */}
      {banners.map((b, i) => (
        <a key={b.id} href={b.link_url || "#"} target="_blank" rel="noopener noreferrer"
          style={{
            position: "absolute", inset: 0,
            opacity: i === current ? 1 : 0,
            transition: "opacity 0.6s ease",
            pointerEvents: i === current ? "auto" : "none",
          }}
        >
          <Image src={mediaUrl(b.image_url)} alt={b.title} fill style={{ objectFit: "cover" }} priority={i === 0} />
          {b.title && (
            <div style={{
              position: "absolute", bottom: 0, left: 0, right: 0,
              background: "linear-gradient(transparent, rgba(0,0,0,0.6))",
              padding: "32px 24px 16px",
            }}>
              <p style={{ color: "#fff", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "18px" }}>{b.title}</p>
            </div>
          )}
        </a>
      ))}

      {/* Arrows */}
      {banners.length > 1 && (
        <>
          {[{ dir: -1 as const, side: "left" }, { dir: 1 as const, side: "right" }].map(({ dir, side }) => (
            <button key={side} onClick={(e) => { e.preventDefault(); go(dir); }}
              style={{
                position: "absolute", top: "50%", [side]: "12px",
                transform: "translateY(-50%)",
                background: "rgba(0,0,0,0.45)", border: "none",
                borderRadius: "50%", width: "36px", height: "36px",
                display: "flex", alignItems: "center", justifyContent: "center",
                cursor: "pointer", color: "#fff", transition: "background 0.2s",
                backdropFilter: "blur(4px)",
              }}
            >
              {dir === -1 ? <ChevronLeft size={18} /> : <ChevronRight size={18} />}
            </button>
          ))}
        </>
      )}

      {/* Dots */}
      {banners.length > 1 && (
        <div style={{ position: "absolute", bottom: "12px", left: "50%", transform: "translateX(-50%)", display: "flex", gap: "6px" }}>
          {banners.map((_, i) => (
            <button key={i}
              onClick={() => { setCurrent(i); resetTimer(); }}
              style={{
                width: i === current ? "20px" : "6px", height: "6px",
                borderRadius: "3px", border: "none",
                background: i === current ? "#fff" : "rgba(255,255,255,0.4)",
                cursor: "pointer", padding: 0, transition: "all 0.3s",
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}
