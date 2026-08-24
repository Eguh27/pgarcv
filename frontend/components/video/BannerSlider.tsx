"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import Image from "next/image";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { Banner } from "@/lib/api";
import { mediaUrl } from "@/lib/api";

export function BannerSlider({ banners }: { banners: Banner[] }) {
  const [current, setCurrent] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const pointerRef = useRef({
    startX: 0,
    startY: 0,
    currentX: 0,
    currentY: 0,
    swiped: false,
  });

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

  const handlePointerDown = (event: React.PointerEvent<HTMLDivElement>) => {
    if (banners.length <= 1 || (event.pointerType === "mouse" && event.button !== 0)) return;
    pointerRef.current = {
      startX: event.clientX,
      startY: event.clientY,
      currentX: event.clientX,
      currentY: event.clientY,
      swiped: false,
    };
    event.currentTarget.setPointerCapture(event.pointerId);
  };

  const handlePointerMove = (event: React.PointerEvent<HTMLDivElement>) => {
    pointerRef.current.currentX = event.clientX;
    pointerRef.current.currentY = event.clientY;
  };

  const handlePointerUp = (event: React.PointerEvent<HTMLDivElement>) => {
    if (banners.length <= 1) return;

    const deltaX = pointerRef.current.currentX - pointerRef.current.startX;
    const deltaY = pointerRef.current.currentY - pointerRef.current.startY;
    const isHorizontalSwipe = Math.abs(deltaX) > 48 && Math.abs(deltaX) > Math.abs(deltaY) * 1.2;

    if (isHorizontalSwipe) {
      pointerRef.current.swiped = true;
      event.preventDefault();
      event.stopPropagation();
      go(deltaX < 0 ? 1 : -1);
    }

    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
  };

  const handleClickCapture = (event: React.MouseEvent<HTMLDivElement>) => {
    if (!pointerRef.current.swiped) return;
    event.preventDefault();
    event.stopPropagation();
    window.setTimeout(() => {
      pointerRef.current.swiped = false;
    }, 0);
  };

  if (!banners.length) return (
    <div style={{ height: "240px", borderRadius: "16px", background: "var(--surface)", display: "flex", alignItems: "center", justifyContent: "center" }}>
      <p style={{ color: "var(--text-muted)" }}>Belum ada banner aktif</p>
    </div>
  );

  return (
    <div
      onPointerDownCapture={handlePointerDown}
      onPointerMoveCapture={handlePointerMove}
      onPointerUpCapture={handlePointerUp}
      onPointerCancelCapture={handlePointerUp}
      onClickCapture={handleClickCapture}
      style={{
        position: "relative",
        borderRadius: "16px",
        overflow: "hidden",
        aspectRatio: "16/5",
        background: "var(--surface)",
        touchAction: "pan-y",
      }}
    >
      {/* Slides */}
      <div
        style={{
          display: "flex",
          height: "100%",
          transform: `translateX(-${current * 100}%)`,
          transition: "transform 0.45s ease",
        }}
      >
        {banners.map((b, i) => (
          <a
            key={b.id}
            href={b.link_url || "#"}
            target="_blank"
            rel="noopener noreferrer"
            draggable={false}
            style={{
              position: "relative",
              flex: "0 0 100%",
              height: "100%",
              display: "block",
              userSelect: "none",
            }}
          >
            <Image src={mediaUrl(b.image_url)} alt={b.title} fill style={{ objectFit: "cover" }} priority={i === 0} draggable={false} />
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
      </div>

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
