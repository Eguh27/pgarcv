"use client";
import { useEffect, useMemo, useState, useRef } from "react";
import { X, Film, Image as ImageIcon } from "lucide-react";
import { adminApi, mediaUrl } from "@/lib/api";
import type { Video } from "@/lib/api";

type FormData = {
  title: string; subtitle: string; description: string;
  thumbnail_url: string; video_url: string; preview_url: string;
  duration: number; is_published: boolean;
  allow_download: boolean;
  categories: string[];
  genres: string[];
  rawPath?: string; // For HLS processing
};

interface Props {
  initial?: Partial<Video>;
  onSubmit: (data: FormData) => Promise<void>;
  loading?: boolean;
}

function UploadZone({
  label,
  accept,
  icon,
  currentUrl,
  onUploaded,
  onAutoThumbnail,
  onVideoUploaded,
}: {
  label: string;
  accept: string;
  icon: React.ReactNode;
  currentUrl: string;
  onUploaded: (url: string) => void;
  onAutoThumbnail?: (thumbnailUrl: string) => void;
  onVideoUploaded?: (data: { url: string; thumbnailUrl: string; duration: number; rawPath: string }) => void;
}) {
  const [dragging, setDragging] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const uploadingRef = useRef(false);

  const upload = async (file: File) => {
    if (uploadingRef.current) return;
    uploadingRef.current = true;
    setUploading(true);
    setProgress(10);
    try {
      // Simulasi progress
      const interval = setInterval(() => setProgress((p) => Math.min(p + 10, 80)), 200);

      let res;
      if (accept.includes("video")) {
        // Upload video menggunakan endpoint khusus
        res = await adminApi.uploadVideo(file);
        if (onVideoUploaded) {
          onVideoUploaded({
            url: res.url,
            thumbnailUrl: res.thumbnail_url,
            duration: res.duration,
            rawPath: res.raw_path,
          });
        }
      } else {
        // Upload gambar menggunakan endpoint biasa
        res = await adminApi.upload(file);
      }

      clearInterval(interval);
      setProgress(100);
      onUploaded(res.url);
      if (onAutoThumbnail && res.thumbnail_url) onAutoThumbnail(res.thumbnail_url);
    } catch (e) {
      alert("Upload gagal: " + (e instanceof Error ? e.message : "Error"));
    } finally {
      uploadingRef.current = false;
      setTimeout(() => {
        setUploading(false);
        setProgress(0);
      }, 500);
    }
  };

  return (
    <div>
      <label
        style={{
          display: "block",
          fontSize: "13px",
          fontWeight: 500,
          marginBottom: "8px",
          color: "var(--text-secondary)",
        }}
      >
        {label}
      </label>
      <div
        onClick={() => inputRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault();
          setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={(e) => {
          e.preventDefault();
          e.stopPropagation();
          setDragging(false);
          const f = e.dataTransfer.files[0];
          if (f && !uploadingRef.current) upload(f);
        }}
        style={{
          border: `2px dashed ${dragging ? "var(--accent)" : "var(--border)"}`,
          borderRadius: "12px",
          padding: "24px",
          cursor: "pointer",
          textAlign: "center",
          background: dragging ? "var(--accent)10" : "var(--bg-secondary)",
          transition: "all 0.2s",
          position: "relative",
          minHeight: "100px",
        }}
      >
        {currentUrl ? (
          <div style={{ position: "relative", display: "inline-block" }}>
            {accept.includes("image") ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={mediaUrl(currentUrl)}
                alt=""
                style={{
                  maxHeight: "120px",
                  maxWidth: "100%",
                  borderRadius: "8px",
                  objectFit: "cover",
                }}
              />
            ) : (
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "8px",
                  color: "var(--text-secondary)",
                }}
              >
                <Film size={20} />{" "}
                <span style={{ fontSize: "13px" }}>{currentUrl.split("/").pop()}</span>
              </div>
            )}
            <button
              onClick={(e) => {
                e.stopPropagation();
                onUploaded("");
              }}
              style={{
                position: "absolute",
                top: "-8px",
                right: "-8px",
                background: "#dc2626",
                border: "none",
                borderRadius: "50%",
                width: "22px",
                height: "22px",
                cursor: "pointer",
                color: "#fff",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <X size={12} />
            </button>
          </div>
        ) : (
          <div style={{ color: "var(--text-muted)" }}>
            <div style={{ marginBottom: "8px" }}>{icon}</div>
            <p style={{ fontSize: "14px" }}>
              {uploading ? `Mengupload... ${progress}%` : "Klik atau drag file ke sini"}
            </p>
            <p style={{ fontSize: "12px", marginTop: "4px" }}>
              {accept.includes("image") ? "JPG, PNG, GIF" : "MP4, WebM, MOV"} (maks 500MB)
            </p>
          </div>
        )}

        {/* Progress bar */}
        {uploading && (
          <div
            style={{
              position: "absolute",
              bottom: 0,
              left: 0,
              right: 0,
              height: "4px",
              background: "var(--surface)",
              borderRadius: "0 0 10px 10px",
            }}
          >
            <div
              style={{
                height: "100%",
                width: `${progress}%`,
                background: "var(--accent)",
                borderRadius: "inherit",
                transition: "width 0.2s",
              }}
            />
          </div>
        )}
      </div>
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        style={{ display: "none" }}
        onChange={(e) => {
          const f = e.target.files?.[0];
          if (f && !uploadingRef.current) {
            upload(f);
            e.target.value = "";
          }
        }}
      />
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  width: "100%", padding: "10px 14px",
  borderRadius: "8px", border: "1px solid var(--border)",
  background: "var(--bg-secondary)", color: "var(--text-primary)",
  fontSize: "14px", outline: "none",
  fontFamily: "var(--font-montserrat)",
  transition: "border-color 0.15s",
};

function uniqLower(items: string[]) {
  const seen = new Set<string>();
  const out: string[] = [];
  for (const it of items) {
    const n = it.trim();
    if (!n) continue;
    const k = n.toLowerCase();
    if (seen.has(k)) continue;
    seen.add(k);
    out.push(n);
  }
  return out;
}

function TagInput({
  label,
  value,
  onChange,
  fetchSuggestions,
  placeholder,
}: {
  label: string;
  value: string[];
  onChange: (next: string[]) => void;
  fetchSuggestions: (q: string) => Promise<string[]>;
  placeholder?: string;
}) {
  const [q, setQ] = useState("");
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    let alive = true;
    if (!open) return;
    fetchSuggestions(q)
      .then((items) => {
        if (!alive) return;
        setSuggestions(items);
      })
      .catch(() => {
        if (!alive) return;
        setSuggestions([]);
      });
    return () => {
      alive = false;
    };
  }, [q, open, fetchSuggestions]);

  const add = (name: string) => {
    onChange(uniqLower([...value, name]));
    setQ("");
    setOpen(false);
  };
  const remove = (name: string) => {
    const key = name.toLowerCase();
    onChange(value.filter((v) => v.toLowerCase() !== key));
  };

  const filteredSuggestions = useMemo(() => {
    const selected = new Set(value.map((v) => v.toLowerCase()));
    return suggestions.filter((s) => !selected.has(s.toLowerCase())).slice(0, 8);
  }, [suggestions, value]);

  return (
    <div style={{ position: "relative" }}>
      <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>
        {label}
      </label>

      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          gap: "8px",
          padding: "10px 12px",
          borderRadius: "8px",
          border: "1px solid var(--border)",
          background: "var(--bg-secondary)",
        }}
        onClick={() => setOpen(true)}
      >
        {value.map((tag) => (
          <span
            key={tag}
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: "6px",
              padding: "6px 10px",
              borderRadius: "999px",
              background: "var(--surface)",
              border: "1px solid var(--border)",
              fontSize: "12px",
              color: "var(--text-primary)",
              fontFamily: "var(--font-montserrat)",
            }}
          >
            {tag}
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                remove(tag);
              }}
              style={{
                border: "none",
                background: "transparent",
                cursor: "pointer",
                color: "var(--text-muted)",
                display: "inline-flex",
                padding: 0,
              }}
              aria-label={`hapus ${tag}`}
            >
              <X size={14} />
            </button>
          </span>
        ))}

        <input
          value={q}
          onChange={(e) => setQ(e.target.value)}
          onFocus={() => setOpen(true)}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          placeholder={placeholder || "Ketik lalu Enter"}
          style={{
            flex: "1 1 140px",
            minWidth: "140px",
            border: "none",
            outline: "none",
            background: "transparent",
            color: "var(--text-primary)",
            fontSize: "14px",
            fontFamily: "var(--font-montserrat)",
            padding: "2px 4px",
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              const name = q.trim();
              if (name) add(name);
            }
          }}
        />
      </div>

      {open && filteredSuggestions.length > 0 && (
        <div
          style={{
            position: "absolute",
            zIndex: 10,
            top: "100%",
            left: 0,
            right: 0,
            marginTop: "6px",
            background: "var(--bg-card)",
            border: "1px solid var(--border)",
            borderRadius: "12px",
            overflow: "hidden",
            boxShadow: "var(--shadow-md)",
          }}
        >
          {filteredSuggestions.map((s) => (
            <button
              key={s}
              type="button"
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => add(s)}
              style={{
                width: "100%",
                textAlign: "left",
                padding: "10px 12px",
                border: "none",
                background: "transparent",
                cursor: "pointer",
                color: "var(--text-primary)",
                fontFamily: "var(--font-montserrat)",
                fontSize: "13px",
              }}
            >
              {s}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

export function VideoForm({ initial, onSubmit, loading }: Props) {
  const initialCategories = useMemo(() => {
    const names = (initial?.categories ?? []).map((c) => c.name).filter(Boolean);
    if (names.length > 0) return names;
    return initial?.category ? [initial.category] : [];
  }, [initial]);
  const initialGenres = useMemo(() => {
    const names = (initial?.genres ?? []).map((g) => g.name).filter(Boolean);
    if (names.length > 0) return names;
    return initial?.genre ? [initial.genre] : [];
  }, [initial]);

  const [form, setForm] = useState<FormData>({
    title: initial?.title ?? "",
    subtitle: initial?.subtitle ?? "",
    description: initial?.description ?? "",
    thumbnail_url: initial?.thumbnail_url ?? "",
    video_url: initial?.video_url ?? "",
    preview_url: initial?.preview_url ?? "",
    duration: initial?.duration ?? 0,
    is_published: initial?.is_published ?? false,
    allow_download: initial?.allow_download ?? false,
    categories: initialCategories,
    genres: initialGenres,
    rawPath: undefined,
  });

  const [showHLSInfo, setShowHLSInfo] = useState(false);

  const set = (key: keyof FormData) => (val: FormData[keyof FormData]) =>
    setForm((prev) => ({ ...prev, [key]: val }));

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.title.trim()) { alert("Judul harus diisi"); return; }
    onSubmit({
      ...form,
      rawPath: form.rawPath, // Include rawPath for HLS processing
    });
  };

  return (
    <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: "20px" }}>
      {/* Title */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Judul *</label>
        <input value={form.title} onChange={(e) => set("title")(e.target.value)} required placeholder="Masukkan judul video" style={inputStyle} />
      </div>

      {/* Subtitle */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Subtitle <span style={{ color: "var(--text-muted)" }}>(opsional)</span></label>
        <input value={form.subtitle} onChange={(e) => set("subtitle")(e.target.value)} placeholder="Teks kecil di bawah judul" style={inputStyle} />
      </div>

      {/* Deskripsi */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Deskripsi</label>
        <textarea value={form.description} onChange={(e) => set("description")(e.target.value)} rows={4}
          placeholder="Deskripsi lengkap video..."
          style={{ ...inputStyle, resize: "vertical", minHeight: "100px" }} />
      </div>

      <TagInput
        label="Category (bisa lebih dari satu)"
        value={form.categories}
        onChange={(v) => set("categories")(v)}
        fetchSuggestions={async (q) => (await adminApi.categories.list(q)).map((x) => x.name)}
        placeholder="Ketik category lalu Enter"
      />

      <TagInput
        label="Genre (bisa lebih dari satu)"
        value={form.genres}
        onChange={(v) => set("genres")(v)}
        fetchSuggestions={async (q) => (await adminApi.genres.list(q)).map((x) => x.name)}
        placeholder="Ketik genre lalu Enter"
      />

      {/* Durasi */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Durasi (detik)</label>
        <input type="number" value={form.duration} onChange={(e) => set("duration")(Number(e.target.value))} min={0} style={{ ...inputStyle, maxWidth: "160px" }} />
        <p style={{ fontSize: "12px", color: "var(--text-muted)", marginTop: "4px" }}>
          = {Math.floor(form.duration / 60)}:{String(form.duration % 60).padStart(2, "0")} menit
        </p>
      </div>

      {/* Upload thumbnail */}
      <UploadZone label="Thumbnail" accept="image/*"
        icon={<ImageIcon size={28} style={{ margin: "0 auto" }} />}
        currentUrl={form.thumbnail_url}
        onUploaded={(url) => set("thumbnail_url")(url)} />

      {/* Upload video */}
      <UploadZone
        label="File Video"
        accept="video/*"
        icon={<Film size={28} style={{ margin: "0 auto" }} />}
        currentUrl={form.video_url}
        onUploaded={(url) => set("video_url")(url)}
        onAutoThumbnail={(thumb) => {
          if (!form.thumbnail_url) set("thumbnail_url")(thumb);
        }}
        onVideoUploaded={(data) => {
          // Don't set video_url to rawPath - that will be set after HLS processing
          // Just store the raw path for later HLS processing
          setForm((prev) => ({ ...prev, rawPath: data.rawPath }));
          if (!form.thumbnail_url) set("thumbnail_url")(data.thumbnailUrl);
          if (form.duration === 0) set("duration")(data.duration);
          setShowHLSInfo(true);
        }}
      />

      {showHLSInfo && (
        <div style={{
          padding: "12px 16px",
          background: "var(--accent)10",
          border: "1px solid var(--accent)30",
          borderRadius: "8px",
          fontSize: "13px",
          color: "var(--text-primary)",
        }}>
          ⚙️ HLS akan diproses setelah video disimpan
        </div>
      )}

      {/* Preview URL */}
      <div>
        <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>URL Preview <span style={{ color: "var(--text-muted)" }}>(clip pendek untuk hover)</span></label>
        <input value={form.preview_url} onChange={(e) => set("preview_url")(e.target.value)} placeholder="/uploads/preview-clip.mp4" style={inputStyle} />
      </div>

      {/* Toggle Published */}
      <div style={{ display: "flex", alignItems: "center", gap: "12px", padding: "16px", background: "var(--bg-secondary)", borderRadius: "12px" }}>
        <button type="button"
          onClick={() => set("is_published")(!form.is_published)}
          style={{
            width: "48px", height: "26px", borderRadius: "13px",
            background: form.is_published ? "var(--accent)" : "var(--surface)",
            border: "none", cursor: "pointer", position: "relative",
            transition: "background 0.2s",
          }}
        >
          <span style={{
            position: "absolute", top: "3px",
            left: form.is_published ? "24px" : "3px",
            width: "20px", height: "20px", borderRadius: "50%",
            background: "#fff", transition: "left 0.2s",
          }} />
        </button>
        <div>
          <p style={{ fontSize: "14px", fontWeight: 500, color: "var(--text-primary)" }}>
            {form.is_published ? "Publik" : "Draft"}
          </p>
          <p style={{ fontSize: "12px", color: "var(--text-muted)" }}>
            {form.is_published ? "Video terlihat di halaman utama" : "Video hanya bisa dilihat admin"}
          </p>
        </div>
      </div>

      {/* Toggle Download */}
      <div style={{ display: "flex", alignItems: "center", gap: "12px", padding: "16px", background: "var(--bg-secondary)", borderRadius: "12px" }}>
        <button type="button"
          onClick={() => set("allow_download")(!form.allow_download)}
          style={{
            width: "48px", height: "26px", borderRadius: "13px",
            background: form.allow_download ? "var(--accent)" : "var(--surface)",
            border: "none", cursor: "pointer", position: "relative",
            transition: "background 0.2s",
          }}
        >
          <span style={{
            position: "absolute", top: "3px",
            left: form.allow_download ? "24px" : "3px",
            width: "20px", height: "20px", borderRadius: "50%",
            background: "#fff", transition: "left 0.2s",
          }} />
        </button>
        <div>
          <p style={{ fontSize: "14px", fontWeight: 500, color: "var(--text-primary)" }}>
            {form.allow_download ? "Download diizinkan" : "Download dimatikan"}
          </p>
          <p style={{ fontSize: "12px", color: "var(--text-muted)" }}>
            {form.allow_download ? "Pengunjung bisa mengunduh video ini" : "Tombol download tidak muncul di publik"}
          </p>
        </div>
      </div>

      {/* Submit */}
      <div style={{ display: "flex", gap: "12px", paddingTop: "8px" }}>
        <button type="submit" disabled={loading}
          style={{
            padding: "12px 28px", borderRadius: "10px",
            background: "var(--accent)", color: "#fff", border: "none",
            fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "15px",
            cursor: loading ? "not-allowed" : "pointer",
            opacity: loading ? 0.7 : 1,
          }}>
          {loading ? "Menyimpan..." : "Simpan Video"}
        </button>
        <button type="button" onClick={() => history.back()}
          style={{ padding: "12px 20px", borderRadius: "10px", background: "var(--surface)", color: "var(--text-primary)", border: "none", cursor: "pointer", fontSize: "14px" }}>
          Batal
        </button>
      </div>
    </form>
  );
}
