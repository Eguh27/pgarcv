"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Ad } from "@/lib/api";

const SESSION_KEY = "ad_shown_session";

export function AdSlot({ position }: { position: "top" | "bottom" | "sidebar" }) {
  const [ad, setAd] = useState<Ad | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const shown = sessionStorage.getItem(SESSION_KEY);
    if (shown) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setReady(true);
      return;
    }

    api.ads
      .serve()
      .then((res) => {
        if (res && res.ad_code) {
          setAd(res);
          sessionStorage.setItem(SESSION_KEY, "1");
        }
      })
      .catch(() => {})
      .finally(() => setReady(true));
  }, []);

  // Belum ready — render placeholder dengan tinggi 0 (tidak geser layout)
  if (!ready || !ad) return null;

  const isTop = position === "top" || position === "bottom";

  return (
    <div
      style={{
        marginBottom: position === "sidebar" ? "16px" : "24px",
        borderRadius: "10px",
        overflow: "hidden",
        border: "1px solid var(--border)",
        position: "relative",
        background: "var(--bg-secondary)",
      }}
    >
      <div
        style={{
          position: "absolute",
          top: "6px",
          right: "8px",
          fontSize: "10px",
          color: "var(--text-muted)",
          background: "var(--bg-secondary)",
          padding: "1px 6px",
          borderRadius: "4px",
          zIndex: 1,
          pointerEvents: "none",
        }}
      >
        Iklan
      </div>
      <iframe
        title={ad.name || "Iklan"}
        srcDoc={ad.ad_code}
        sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox"
        loading="lazy"
        style={{
          width: "100%",
          minHeight: isTop ? "100px" : "250px",
          border: 0,
          display: "block",
        }}
      />
    </div>
  );
}

