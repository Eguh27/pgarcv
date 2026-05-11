"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import useSWR from "swr";

import { VideoCard } from "@/components/video/VideoCard";
import { Filter, X } from "lucide-react";

const BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

function SkeletonCard() {
  return (
    <div
      style={{
        borderRadius: "12px",
        overflow: "hidden",
        background: "var(--bg-card)",
        border: "1px solid var(--border)",
      }}
    >
      <div style={{ aspectRatio: "16/9", background: "var(--surface)" }} />
      <div style={{ padding: "12px", display: "flex", flexDirection: "column", gap: "8px" }}>
        <div style={{ height: "16px", borderRadius: "4px", background: "var(--surface)", width: "80%" }} />
        <div style={{ height: "12px", borderRadius: "4px", background: "var(--surface)", width: "50%" }} />
      </div>
    </div>
  );
}

function FilterChip({ label, onRemove }: { label: string; onRemove: () => void }) {
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: "6px",
        padding: "4px 10px",
        borderRadius: "20px",
        background: "var(--accent)",
        color: "#fff",
        fontSize: "12px",
        fontWeight: 600,
        fontFamily: "var(--font-montserrat)",
      }}
    >
      {label}
      <button
        onClick={onRemove}
        style={{ background: "none", border: "none", cursor: "pointer", color: "#fff", display: "flex", padding: 0 }}
      >
        <X size={12} />
      </button>
    </span>
  );
}

function SearchResults() {
  const searchParams = useSearchParams();
  const router = useRouter();

  const query = searchParams.get("q")?.trim() ?? "";
  const [selectedGenre, setSelectedGenre] = useState(searchParams.get("genre") ?? "");
  const [selectedCategory, setSelectedCategory] = useState(searchParams.get("category") ?? "");
  const [showFilters, setShowFilters] = useState(false);

  const { data: genresData } = useSWR("genres", () =>
    fetch(`${BASE}/api/genres`, { headers: { "ngrok-skip-browser-warning": "true" } })
      .then((r) => r.json())
      .then((r) => r.data ?? [])
  );

  const { data: categoriesData } = useSWR("categories", () =>
    fetch(`${BASE}/api/categories`, { headers: { "ngrok-skip-browser-warning": "true" } })
      .then((r) => r.json())
      .then((r) => r.data ?? [])
  );

  const genres: string[] = (genresData ?? []).map((g: { name: string }) => g.name);
  const categories: string[] = (categoriesData ?? []).map((c: { name: string }) => c.name);

  const swrKey = query || selectedGenre || selectedCategory ? ["videos-search", query, selectedGenre, selectedCategory] : null;

  const { data, isLoading } = useSWR(swrKey, () => {
    const params = new URLSearchParams();
    if (query) params.set("search", query);
    if (selectedGenre) params.set("genre", selectedGenre);
    if (selectedCategory) params.set("category", selectedCategory);

    return fetch(`${BASE}/api/videos?${params}`, {
      headers: { "ngrok-skip-browser-warning": "true" },
    }).then((r) => r.json());
  });

  const videos = data?.data ?? [];
  const total = data?.total ?? 0;
  const hasFilter = selectedGenre || selectedCategory;

  const chipStyle = (active: boolean): React.CSSProperties => ({
    padding: "6px 14px",
    borderRadius: "20px",
    border: "1px solid var(--border)",
    background: active ? "var(--accent)" : "var(--surface)",
    color: active ? "#fff" : "var(--text-secondary)",
    cursor: "pointer",
    fontSize: "13px",
    fontWeight: active ? 600 : 400,
    fontFamily: "var(--font-montserrat)",
    transition: "all 0.15s",
  });

  return (
    <div>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "16px", flexWrap: "wrap", gap: "8px" }}>
        <div>
          <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "20px", marginBottom: "2px" }}>
            {query ? `Hasil: "${query}"` : "Semua Video"}
          </h1>
          <p style={{ fontSize: "13px", color: "var(--text-muted)" }}>{isLoading ? "Mencari..." : `${total} video ditemukan`}</p>
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          style={{
            display: "flex",
            alignItems: "center",
            gap: "6px",
            padding: "8px 14px",
            borderRadius: "20px",
            background: showFilters || hasFilter ? "var(--accent)" : "var(--surface)",
            color: showFilters || hasFilter ? "#fff" : "var(--text-secondary)",
            border: "1px solid var(--border)",
            cursor: "pointer",
            fontSize: "13px",
            fontWeight: 600,
            fontFamily: "var(--font-montserrat)",
          }}
        >
          <Filter size={14} />
          Filter {hasFilter ? "(aktif)" : ""}
        </button>
      </div>

      {showFilters && (
        <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", padding: "16px", marginBottom: "20px" }}>
          {genres.length > 0 && (
            <div style={{ marginBottom: "12px" }}>
              <p style={{ fontSize: "12px", fontWeight: 600, color: "var(--text-muted)", marginBottom: "8px", textTransform: "uppercase", letterSpacing: "0.05em" }}>Genre</p>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "8px" }}>
                {genres.map((g) => (
                  <button key={g} onClick={() => setSelectedGenre(selectedGenre === g ? "" : g)} style={chipStyle(selectedGenre === g)}>
                    {g}
                  </button>
                ))}
              </div>
            </div>
          )}

          {categories.length > 0 && (
            <div>
              <p style={{ fontSize: "12px", fontWeight: 600, color: "var(--text-muted)", marginBottom: "8px", textTransform: "uppercase", letterSpacing: "0.05em" }}>Kategori</p>
              <div style={{ display: "flex", flexWrap: "wrap", gap: "8px" }}>
                {categories.map((cat) => (
                  <button key={cat} onClick={() => setSelectedCategory(selectedCategory === cat ? "" : cat)} style={chipStyle(selectedCategory === cat)}>
                    {cat}
                  </button>
                ))}
              </div>
            </div>
          )}

          {hasFilter && (
            <button
              onClick={() => {
                setSelectedGenre("");
                setSelectedCategory("");
              }}
              style={{ marginTop: "12px", fontSize: "12px", color: "var(--accent)", background: "none", border: "none", cursor: "pointer", padding: 0 }}
            >
              Reset semua filter
            </button>
          )}
        </div>
      )}

      {hasFilter && (
        <div style={{ display: "flex", gap: "8px", marginBottom: "16px", flexWrap: "wrap" }}>
          {selectedGenre && <FilterChip label={`Genre: ${selectedGenre}`} onRemove={() => setSelectedGenre("")} />}
          {selectedCategory && <FilterChip label={`Kategori: ${selectedCategory}`} onRemove={() => setSelectedCategory("")} />}
        </div>
      )}

      {!query && !hasFilter && (
        <div style={{ padding: "48px 0", textAlign: "center" }}>
          <p style={{ fontSize: "40px", marginBottom: "12px" }}>🔍</p>
          <p style={{ fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "18px", marginBottom: "8px" }}>Cari video</p>
          <p style={{ fontSize: "14px", color: "var(--text-muted)" }}>Ketik di search bar di atas, atau pilih filter genre/kategori</p>
        </div>
      )}

      {(query || hasFilter) && (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(260px, 1fr))", gap: "16px" }}>
          {isLoading ? (
            Array.from({ length: 8 }).map((_, i) => <SkeletonCard key={i} />)
          ) : videos.length > 0 ? (
            videos.map((video: Parameters<typeof VideoCard>[0]["video"]) => (
              <VideoCard key={video.id} video={video} onClick={(id) => router.push(`/watch/${id}`)} />
            ))
          ) : (
            <div style={{ gridColumn: "1/-1", padding: "64px 0", textAlign: "center" }}>
              <p style={{ fontSize: "40px", marginBottom: "12px" }}>😕</p>
              <p style={{ fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "18px", marginBottom: "8px" }}>Tidak ada hasil</p>
              <p style={{ fontSize: "14px", color: "var(--text-muted)" }}>Coba kata kunci atau filter yang berbeda</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense
      fallback={<div style={{ padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>Memuat...</div>}
    >
      <SearchResults />
    </Suspense>
  );
}

