"use client";

import { useState } from "react";
import { Download, Share2, Copy } from "lucide-react";

function btnStyle(primary = false): React.CSSProperties {
  return {
    display: "inline-flex",
    alignItems: "center",
    gap: "6px",
    padding: "8px 14px",
    borderRadius: "20px",
    border: primary ? "none" : "1px solid var(--border)",
    background: primary ? "var(--accent)" : "var(--surface)",
    color: primary ? "#fff" : "var(--text-primary)",
    cursor: "pointer",
    fontFamily: "var(--font-montserrat)",
    fontSize: "13px",
    fontWeight: 600,
    transition: "all 0.15s",
    whiteSpace: "nowrap" as const,
  };
}

export function WatchActions({
  videoId,
  title,
  allowDownload,
}: {
  videoId: number;
  title: string;
  allowDownload: boolean;
}) {
  const [downloading, setDownloading] = useState(false);
  const [copied, setCopied] = useState(false);

  const getShareUrl = () => {
    // Menghindari beberapa edge-case saat hanya pakai window.location.href
    const { origin, pathname, search, hash } = window.location;
    return `${origin}${pathname}${search}${hash}`;
  };

  const copyLink = async () => {
    const url = getShareUrl();

    // 1) Prefer modern Clipboard API
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(url);
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
        return;
      }
    } catch {
      // lanjut ke fallback
    }

    // 2) Fallback: execCommand copy (untuk browser yang tidak support navigator.clipboard)
    try {
      const textarea = document.createElement("textarea");
      textarea.value = url;
      textarea.setAttribute("readonly", "");
      textarea.style.position = "absolute";
      textarea.style.left = "-9999px";
      document.body.appendChild(textarea);
      textarea.select();
      const ok = document.execCommand("copy");
      textarea.remove();

      if (ok) {
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
        return;
      }
    } catch {
      // lanjut ke prompt
    }

    // 3) Last resort
    prompt("Salin link ini:", url);
  };

  const share = async () => {
    const url = getShareUrl();
    try {
      if (navigator.share) {
        await navigator.share({ title, url });
        return;
      }
    } catch {
      // user cancelled or share failed -> fallback to copy
    }
    await copyLink();
  };

  const download = async () => {
    setDownloading(true);
    try {
      const res = await fetch(`/api/videos/${videoId}/download`);
      if (!res.ok) {
        const msg = await res.text().catch(() => "");
        throw new Error(msg || `Gagal download (${res.status})`);
      }
      const blob = await res.blob();
      const cd = res.headers.get("content-disposition") || "";
      const match = /filename=([^;]+)/i.exec(cd);
      const filename = match?.[1]?.replaceAll('"', "") || `video-${videoId}.mp4`;

      const a = document.createElement("a");
      const url = URL.createObjectURL(blob);
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (e) {
      alert(e instanceof Error ? e.message : "Gagal download");
    } finally {
      setDownloading(false);
    }
  };

  return (
    <div style={{ display: "flex", gap: "8px", flexWrap: "wrap", margin: "10px 0 14px" }}>
      <button
        type="button"
        onClick={share}
        style={btnStyle(false)}
      >
        <Share2 size={16} />
        Bagikan
      </button>

      <button
        type="button"
        onClick={copyLink}
        style={btnStyle(false)}
      >
        <Copy size={16} />
        {copied ? "✓ Tersalin" : "Salin Link"}
      </button>

      {allowDownload ? (
        <button
          type="button"
          onClick={download}
          disabled={downloading}
          style={{ ...btnStyle(true), opacity: downloading ? 0.7 : 1 }}
        >
          <Download size={16} />
          {downloading ? "Mengunduh..." : "Download"}
        </button>
      ) : null}
    </div>
  );
}

