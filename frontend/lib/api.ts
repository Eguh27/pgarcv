const BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const NGROK_HEADER: Record<string, string> = BASE.includes("ngrok")
  ? { "ngrok-skip-browser-warning": "true" }
  : {};

function saveToken(token: string) {
  try { localStorage.setItem("admin_token", token); } catch {}
}
function getToken(): string {
  try { return localStorage.getItem("admin_token") || ""; } catch { return ""; }
}
function clearToken() {
  try { localStorage.removeItem("admin_token"); } catch {}
}

function extractData<T>(json: unknown): T {
  if (json === null || json === undefined) throw new Error("Response kosong");
  const j = json as Record<string, unknown>;
  if ("success" in j && "data" in j) return j.data as T;
  if (Array.isArray(json)) return json as T;
  return json as T;
}

export interface Video {
  id: number;
  title: string;
  subtitle: string;
  description: string;
  thumbnail_url: string;
  video_url: string;
  preview_url: string;
  duration: number;
  views: number;
  is_published: boolean;
  allow_download?: boolean;
  category?: string;
  genre?: string;
  categories?: Array<{ id: number; name: string }>;
  genres?: Array<{ id: number; name: string }>;
  created_at: string;
  hls_status?: string;
  hls_error_msg?: string;
}

export interface TaxonomyItem {
  id: number;
  name: string;
}

export type VideoUpsertRequest = {
  title: string;
  subtitle: string;
  description: string;
  thumbnail_url: string;
  video_url: string;
  preview_url: string;
  duration: number;
  is_published: boolean;
  allow_download: boolean;
  categories: string[];
  genres: string[];
  rawPath?: string;
};

export interface Banner {
  id: number;
  title: string;
  image_url: string;
  link_url: string;
  is_active: boolean;
  sort_order: number;
}

export interface Ad {
  id: number;
  name: string;
  ad_code: string;
  device_type: string;
  is_active: boolean;
}

export interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

async function get<T>(path: string): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = { ...NGROK_HEADER };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  let res: Response;
  try {
    res = await fetch(`${BASE}${path}`, { credentials: "include", headers });
  } catch (e) {
    throw new Error(`Network error: ${e instanceof Error ? e.message : "gagal"}`);
  }

  if (!res.ok) {
    let msg = `API Error: ${res.status}`;
    try {
      const err: unknown = await res.json();
      if (err && typeof err === "object" && "message" in (err as Record<string, unknown>)) {
        const m = (err as Record<string, unknown>).message;
        if (typeof m === "string") msg = m;
      }
    } catch {}
    throw new Error(msg);
  }

  let json: unknown;
  try {
    json = await res.json();
  } catch {
    throw new Error("Response bukan JSON yang valid");
  }

  return extractData<T>(json);
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...NGROK_HEADER,
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, {
    method: "POST",
    headers,
    credentials: "include",
    body: JSON.stringify(body),
  });
  const json: unknown = await res.json();
  if (!json || typeof json !== "object") throw new Error("Error");
  const r = json as { success?: unknown; message?: unknown; data?: unknown };
  if (r.success !== true) throw new Error(typeof r.message === "string" ? r.message : "Error");
  return r.data as T;
}

async function put<T>(path: string, body: unknown): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...NGROK_HEADER,
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, {
    method: "PUT",
    headers,
    credentials: "include",
    body: JSON.stringify(body),
  });
  const json = await res.json();
  if (!json.success) throw new Error(json.message || "Error");
  return json.data;
}

async function del(path: string): Promise<void> {
  const token = getToken();
  const headers: Record<string, string> = { ...NGROK_HEADER };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  let res: Response;
  try {
    res = await fetch(`${BASE}${path}`, {
      method: "DELETE",
      credentials: "include",
      headers,
    });
  } catch (e) {
    throw new Error(`Network error: ${e instanceof Error ? e.message : "gagal"}`);
  }

  if (!res.ok) {
    let msg = `Delete gagal: ${res.status}`;
    try {
      const err = await res.json();
      msg = err.message || msg;
    } catch {}
    throw new Error(msg);
  }
}

// Public APIs
export const api = {
  videos: {
    list: (page = 1, search = "") =>
      get<PaginatedResponse<Video>>(`/api/videos?page=${page}&search=${search}`),
    get: (id: number) => get<Video>(`/api/videos/${id}`),
    featured: () => get<Video[]>(`/api/videos/featured`),
  },
  categories: {
    list: (q = "") => get<TaxonomyItem[]>(`/api/categories${q ? `?q=${encodeURIComponent(q)}` : ""}`),
  },
  genres: {
    list: (q = "") => get<TaxonomyItem[]>(`/api/genres${q ? `?q=${encodeURIComponent(q)}` : ""}`),
  },
  banners: {
    list: () => get<Banner[]>(`/api/banners`),
  },
  ads: {
    serve: () => get<Ad | null>(`/api/ads/serve`),
  },
};

// Admin APIs
export const adminApi = {
  auth: {
    login: (username: string, password: string) =>
      post<{ token: string }>("/api/admin/login", { username, password }).then((json) => {
        if (json?.token) saveToken(json.token);
        return json;
      }),
    logout: () => {
      clearToken();
      return post("/api/admin/logout", {});
    },
  },
  videos: {
    list: () => get<PaginatedResponse<Video>>(`/api/admin/videos`),
    get: (id: number) => get<Video>(`/api/admin/videos/${id}`),
    create: (data: VideoUpsertRequest) => post<Video>("/api/admin/videos", data),
    update: (id: number, data: VideoUpsertRequest) => put<Video>(`/api/admin/videos/${id}`, data),
    delete: (id: number) => del(`/api/admin/videos/${id}`),
  },
  categories: {
    list: (q = "") => get<TaxonomyItem[]>(`/api/admin/categories${q ? `?q=${encodeURIComponent(q)}` : ""}`),
  },
  genres: {
    list: (q = "") => get<TaxonomyItem[]>(`/api/admin/genres${q ? `?q=${encodeURIComponent(q)}` : ""}`),
  },
  banners: {
    list: () => get<Banner[]>(`/api/admin/banners`),
    create: (data: Partial<Banner>) => post<Banner>("/api/admin/banners", data),
    update: (id: number, data: Partial<Banner>) => put<Banner>(`/api/admin/banners/${id}`, data),
    delete: (id: number) => del(`/api/admin/banners/${id}`),
  },
  ads: {
    list: () => get<Ad[]>(`/api/admin/ads`),
    create: (data: Partial<Ad>) => post<Ad>("/api/admin/ads", data),
    update: (id: number, data: Partial<Ad>) => put<Ad>(`/api/admin/ads/${id}`, data),
    delete: (id: number) => del(`/api/admin/ads/${id}`),
  },
  upload: async (file: File): Promise<{ url: string; filename: string; thumbnail_url?: string }> => {
    const MAX_SIZE = 500 * 1024 * 1024;
    const ALLOWED_IMAGE = [".jpg", ".jpeg", ".png", ".gif"];
    const ALLOWED_VIDEO = [".mp4", ".webm", ".mov", ".avi"];
    const ext = "." + file.name.split(".").pop()?.toLowerCase();
    const allowed = [...ALLOWED_IMAGE, ...ALLOWED_VIDEO];
    if (!allowed.includes(ext)) throw new Error(`Format tidak didukung: ${ext}`);
    if (file.size > MAX_SIZE) throw new Error("File terlalu besar (maks 500MB)");
    if (file.size === 0) throw new Error("File kosong");

    const form = new FormData();
    form.append("file", file);
    const token = getToken();
    const headers: Record<string, string> = { ...NGROK_HEADER };
    if (token) headers["Authorization"] = `Bearer ${token}`;

    let res: Response;
    try {
      res = await fetch(`${BASE}/api/admin/upload`, {
        method: "POST",
        credentials: "include",
        headers,
        body: form,
      });
    } catch (e) {
      throw new Error(`Upload network error: ${e instanceof Error ? e.message : "gagal"}`);
    }

    if (!res.ok) {
      let msg = `Upload gagal (${res.status})`;
      try {
        const err = await res.json();
        msg = err.message || msg;
      } catch {}
      throw new Error(msg);
    }

    const json = await res.json();
    if (!json.success) throw new Error(json.message || "Upload gagal");
    return json.data;
  },
  uploadVideo: async (file: File): Promise<{ url: string; thumbnail_url: string; duration: number; raw_path: string; message: string }> => {
    const token = getToken();
    const form = new FormData();
    form.append("file", file);
    const headers: Record<string, string> = { ...NGROK_HEADER };
    if (token) headers["Authorization"] = `Bearer ${token}`;

    const res = await fetch(`${BASE}/api/admin/upload/video`, {
      method: "POST",
      credentials: "include",
      headers,
      body: form,
    });

    if (!res.ok) {
      const text = await res.text();
      throw new Error(`Upload video gagal (${res.status}): ${text}`);
    }

    const json = await res.json();
    if (!json.success) throw new Error(json.message || "Upload video gagal");
    return json.data;
  },
  processHLS: async (videoId: number, rawPath: string): Promise<{ message: string; video_id: number }> => {
    return post<{ message: string; video_id: number }>("/api/admin/upload/process-hls", { video_id: videoId, raw_path: rawPath });
  },
  getHLSStatus: async (videoId: number): Promise<{ id: number; hls_status: string; hls_error: string; video_url: string; is_complete: boolean }> => {
    return get(`/api/videos/${videoId}/hls-status`);
  },
  chunkedUpload: {
    initiate: async (filename: string, fileSize: number, totalChunks: number): Promise<{ upload_id: string; total_chunks: number; chunk_size: number }> => {
      return post(`/api/admin/upload/chunked/initiate`, { filename, file_size: fileSize, total_chunks: totalChunks });
    },
    uploadChunk: async (uploadId: string, chunkIndex: number, chunkFile: Blob): Promise<{ chunk_index: number; chunks_received: number; total_chunks: number; progress_pct: number }> => {
      const token = getToken();
      const form = new FormData();
      form.append("chunk", chunkFile);
      form.append("chunk_index", chunkIndex.toString());
      const headers: Record<string, string> = { ...NGROK_HEADER };
      if (token) headers["Authorization"] = `Bearer ${token}`;

      const res = await fetch(`${BASE}/api/admin/upload/chunked/${uploadId}/chunk`, {
        method: "POST",
        credentials: "include",
        headers,
        body: form,
      });

      if (!res.ok) throw new Error(`Chunk upload failed: ${res.status}`);
      const json = await res.json();
      if (!json.success) throw new Error(json.message || "Chunk upload failed");
      return json.data;
    },
    complete: async (uploadId: string): Promise<{ filename: string; path: string; size: number }> => {
      return post(`/api/admin/upload/chunked/${uploadId}/complete`, {});
    },
    getProgress: async (uploadId: string): Promise<{ upload_id: string; chunks_received: number; total_chunks: number; progress_pct: number; status: string }> => {
      return get(`/api/admin/upload/chunked/${uploadId}/progress`);
    },
    abort: async (uploadId: string): Promise<{ message: string }> => {
      const token = getToken();
      const headers: Record<string, string> = { ...NGROK_HEADER };
      if (token) headers["Authorization"] = `Bearer ${token}`;
      const res = await fetch(`${BASE}/api/admin/upload/chunked/${uploadId}`, {
        method: "DELETE",
        credentials: "include",
        headers,
      });
      if (!res.ok) throw new Error("Abort failed");
      return res.json();
    },
  },
};

// ✅ FIX: mediaUrl sekarang prefix dengan BASE untuk path relatif
export function mediaUrl(path: string): string {
  if (!path) return "/placeholder.jpg";
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  // Path relatif → prefix dengan BASE supaya gambar dari backend bisa diakses
  if (path.startsWith("/")) return `${BASE}${path}`;
  return `${BASE}/${path}`;
}

export function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function formatViews(views: number): string {
  if (views >= 1_000_000) return `${(views / 1_000_000).toFixed(1)}Jt`;
  if (views >= 1_000) return `${(views / 1_000).toFixed(1)}Rb`;
  return views.toString();
}