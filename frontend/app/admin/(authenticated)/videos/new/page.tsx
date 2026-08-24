"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { VideoForm } from "@/components/admin/VideoForm";
import { adminApi } from "@/lib/api";

export default function NewVideoPage() {
  const [loading, setLoading] = useState(false);
  const [hlsStatus, setHlsStatus] = useState<string>("");
  const router = useRouter();

  // ✅ Poll HLS status
  const pollHLSStatus = async (videoId: number, maxAttempts = 60) => {
    for (let i = 0; i < maxAttempts; i++) {
      try {
        const status = await adminApi.getHLSStatus(videoId);
        setHlsStatus(status.hls_status);

        if (status.is_complete) {
          return true;
        }

        if (status.hls_status === "error") {
          alert(`❌ HLS processing gagal: ${status.hls_error}`);
          return false;
        }

        // Wait 2 seconds before retry
        await new Promise((r) => setTimeout(r, 2000));
      } catch (err) {
        console.error("Status check failed:", err);
      }
    }
    return false;
  };

  const handleSubmit = async (data: Parameters<typeof adminApi.videos.create>[0]) => {
    setLoading(true);
    try {
      const savedVideo = await adminApi.videos.create(data);

      // Check if HLS processing is needed (rawPath exists)
      if (data.rawPath) {
        try {
          await adminApi.processHLS(savedVideo.id, data.rawPath);
          setHlsStatus("processing");

          // ✅ Poll for completion
          const success = await pollHLSStatus(savedVideo.id);

          if (success) {
            alert("✅ Video disimpan dan HLS siap diplay!");
          } else {
            alert("⚠️ Video disimpan, tapi HLS processing masih berlangsung. Cek nanti.");
          }
        } catch (hlsError) {
          console.error("HLS processing failed:", hlsError);
          alert("Video disimpan, tapi HLS processing gagal. Coba lagi nanti.");
        }
      } else {
        alert("✅ Video disimpan!");
      }

      router.push("/admin/videos");
    } catch (e) {
      alert("Gagal simpan: " + (e instanceof Error ? e.message : "Error"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: "720px" }}>
      <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px", marginBottom: "24px" }}>
        Tambah Video Baru
        {hlsStatus && <span style={{ fontSize: "14px", marginLeft: "12px", color: "var(--text-secondary)" }}>HLS: {hlsStatus}</span>}
      </h1>
      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", padding: "28px" }}>
        <VideoForm onSubmit={handleSubmit} loading={loading} />
      </div>
    </div>
  );
}
