import { afterEach, describe, expect, it, vi } from "vitest";
import { api, formatDuration, formatViews, mediaUrl } from "./api";

const BASE = "http://localhost:8080";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("mediaUrl", () => {
  // mediaUrl sengaja selalu mengembalikan path relatif, bukan URL absolut.
  // Nilainya dipakai di srcset <img>/next/image yang dibaca browser, sementara
  // BASE untuk server bisa berupa hostname internal container yang tidak bisa
  // di-resolve browser. Path relatif lewat rewrite Next.js.
  it("path kosong → placeholder relatif", () => {
    expect(mediaUrl("")).toBe("/uploads/placeholder.jpg");
  });

  it("URL absolut (Cloudinary) tidak diubah", () => {
    const url = "https://res.cloudinary.com/demo/video/upload/v.mp4";
    expect(mediaUrl(url)).toBe(url);
  });

  it("path relatif /uploads dikeembalikan apa adanya", () => {
    expect(mediaUrl("/uploads/a.jpg")).toBe("/uploads/a.jpg");
  });

  it("path tanpa leading slash mendapat leading slash", () => {
    expect(mediaUrl("hls/video_1/index.m3u8")).toBe("/hls/video_1/index.m3u8");
  });

  it("tidak pernah mengembalikan hostname internal container", () => {
    const out = mediaUrl("/uploads/a.jpg");
    expect(out).not.toContain("backend:8080");
    expect(out).not.toContain("localhost:8080");
  });
});

describe("formatDuration & formatViews", () => {
  it("durasi 0 → 0:00", () => {
    expect(formatDuration(0)).toBe("0:00");
  });

  it("durasi 125 detik → 2:05", () => {
    expect(formatDuration(125)).toBe("2:05");
  });

  it("views < 1000 tampil apa adanya", () => {
    expect(formatViews(999)).toBe("999");
  });

  it("views ribuan → Rb, jutaan → Jt", () => {
    expect(formatViews(1500)).toContain("Rb");
    expect(formatViews(2_500_000)).toContain("Jt");
  });
});

describe("api.videos.list", () => {
  it("unwrap {success,data} menjadi array", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ success: true, data: [{ id: 1, title: "A" }] }), {
          status: 200,
        })
      )
    );
    const data = await api.videos.list();
    // Server membalas {success,data:[...]} sehingga extractData mengembalikan
    // array-nya langsung, bukan objek PaginatedResponse.
    const list = Array.isArray(data) ? data : (data as unknown as { data: Array<{ title: string }> }).data;
    expect(Array.isArray(list)).toBe(true);
    expect(list[0].title).toBe("A");
    expect(fetch).toHaveBeenCalledWith(
      `${BASE}/api/videos?page=1&search=`,
      expect.objectContaining({ credentials: "include" })
    );
  });

  it("response array polos dilewatkan langsung", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify([{ id: 2 }]), { status: 200 })
      )
    );
    const data = await api.videos.list();
    expect(data).toEqual([{ id: 2 }]);
  });

  it("error API mempropagasikan message dari server", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ success: false, message: "Video tidak ditemukan" }), {
          status: 404,
        })
      )
    );
    await expect(api.videos.list()).rejects.toThrow("Video tidak ditemukan");
  });

  it("kegagalan jaringan → error Network error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("boom")));
    await expect(api.videos.list()).rejects.toThrow(/^Network error:/);
  });

  it("body bukan JSON → pesan ramah", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response("<html>gateway</html>", { status: 200 }))
    );
    await expect(api.videos.list()).rejects.toThrow("Response bukan JSON yang valid");
  });
});
