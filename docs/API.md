# Referensi API

Base URL lokal: `http://localhost:8080`

Semua response memakai amplop yang konsisten:

```json
// sukses
{ "success": true, "data": { } }

// gagal
{ "success": false, "message": "Video tidak ditemukan" }
```

Daftar berpaginasi:

```json
{ "success": true, "data": [], "total": 42, "page": 1, "limit": 12, "total_pages": 4 }
```

---

## Autentikasi admin

Endpoint di bawah `/api/admin/*` (kecuali login/logout) butuh token JWT HS256.

### Login

```http
POST /api/admin/login
Content-Type: application/json

{ "username": "admin", "password": "admin" }
```

```json
{ "success": true, "data": { "token": "eyJhbGciOiJIUzI1NiIs...", "message": "Login berhasil" } }
```

Respons juga memasang cookie `admin_token` (httpOnly, `SameSite=Lax`, berlaku 24 jam).
Masa berlaku token ikut mengikuti `COOKIE_SECURE` — saat `true`, cookie hanya
dikirim lewat HTTPS.

### Mengirim token

Cara pertama (disarankan — cookie ikut terkirim otomatis):

```http
Cookie: admin_token=eyJhbGciOiJIUzI1NiIs...
```

Cara kedua:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

Cookie diperiksa lebih dulu, baru header `Authorization`.

### Logout

```http
POST /api/admin/logout
```

Menghapus cookie (MaxAge −1).

### Response error autentikasi

| Kode | Arti |
|---|---|
| 401 | Token tidak ada, tidak valid, atau kedaluwarsa |
| 403 | Wajib login, tapi tidak punya akses |
| 404 | Endpoint tidak ada |

---

## Endpoint publik

Tanpa autentikasi.

### Video

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/videos` | Daftar video published, berpaginasi |
| GET | `/api/videos/featured` | Video unggulan |
| GET | `/api/videos/:id` | Detail satu video |
| GET | `/api/videos/:id/download` | Unduh file asli (butuh `allow_download`) |
| GET | `/api/videos/:id/hls-status` | Status transcode HLS |
| GET | `/api/videos/suggest` | Rekomendasi video |

#### `GET /api/videos` — parameter query

| Parameter | Tipe | Default | Keterangan |
|---|---|---|---|
| `page` | int | `1` | Nomor halaman |
| `limit` | int | `12` | Maksimal 50; nilai di luar itu dikembalikan ke default |
| `search` | string | — | Pencarian di title, subtitle, category, genre |
| `category` | string | — | Filter kategori (persis, case-sensitive di SQL) |
| `genre` | string | — | Filter genre |

```bash
curl "http://localhost:8080/api/videos?page=1&limit=12&search=tutorial&category=Teknologi"
```

Hanya video `is_published = true` yang dikembalikan.

#### `GET /api/videos/:id/download`

| Syarat | Hasil |
|---|---|
| Video tidak ada | 404 |
| Video draft | 404 |
| `allow_download = false` | 403 |
| `video_url` kosong | 404 |

Mengirim file sebagai attachment. Path lokal dicek terhadap path traversal
(nilai berisi `..` atau dotfile ditolak).

### Taksonomi

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/categories` | Daftar kategori. Opsional: `?q=` untuk cari |
| GET | `/api/genres` | Daftar genre. Opsional: `?q=` untuk cari |

### Banner & iklan

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/banners` | Banner aktif, urut `sort_order` |
| GET | `/api/ads/serve` | Satu iklan sesuai perangkat |

`/api/ads/serve` mendeteksi perangkat dari header `User-Agent` (`mobile`/`android`/`iphone` → mobile, selain itu desktop) lalu mengembalikan iklan acak yang cocok, atau `data: null` dengan status 200 bila tidak ada.

### Streaming HLS

Middleware `OptionalAuth` — endpoint tetap publik untuk video published, tapi
admin yang login juga boleh mengakses preview video draft.

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/hls/:id/index.m3u8` | Playlist master |
| GET | `/api/hls/:id/:segment` | Segmen `.ts` |
| GET | `/api/hls/key/:id` | Kunci dekripsi AES-128 |

Aturan akses:

| Kondisi | Hasil |
|---|---|
| Published | 200 untuk siapa pun |
| Draft, tanpa login | 403 |
| Draft, admin login | 200 |

`/api/hls/key/:id` juga memvalidasi header `Referer`/`Origin` — key hanya
disajikan jika berasal dari `SERVER_URL` atau localhost.

Parameter `:segment` divalidasi: nilai kosong, berawalan titik, atau berisi
subpath ditolak dengan 400.

---

## Endpoint admin

Semua butuh autentikasi. Prefix: `/api/admin`.

### Video CRUD

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/admin/videos` | Daftar semua video (termasuk draft) |
| GET | `/api/admin/videos/:id` | Detail video |
| POST | `/api/admin/videos` | Buat video baru |
| PUT | `/api/admin/videos/:id` | Update video |
| DELETE | `/api/admin/videos/:id` | Hapus video |

#### Body create / update

```json
{
  "title": "Judul Video",
  "subtitle": "Subtitle opsional",
  "description": "Deskripsi lengkap",
  "thumbnail_url": "/uploads/abc.jpg",
  "video_url": "/api/hls/12/index.m3u8",
  "preview_url": "/uploads/preview.mp4",
  "duration": 300,
  "categories": ["Teknologi", "Tutorial"],
  "genres": ["Programming"],
  "is_published": true,
  "allow_download": false
}
```

Seluruh field bersifat **partial update** — field yang tidak dikirim tidak
menimpa nilai lama. `POST` mewajibkan `title`.

`categories` dan `genres` berupa array nama. Nama yang belum ada **dibuat
otomatis**; pencocokan case-insensitive sehingga `"Tutorial"` dan `"tutorial"`
tidak tercipta dua kali.

### Kategori & genre

| Method | Path |
|---|---|
| GET | `/api/admin/categories` |
| GET | `/api/admin/genres` |

### Banner CRUD

| Method | Path |
|---|---|
| GET | `/api/admin/banners` |
| POST | `/api/admin/banners` |
| PUT | `/api/admin/banners/:id` |
| DELETE | `/api/admin/banners/:id` |

```json
{
  "title": "Banner Utama",
  "image_url": "/uploads/banner.jpg",
  "link_url": "https://example.com",
  "is_active": true,
  "sort_order": 1
}
```

### Iklan CRUD

| Method | Path |
|---|---|
| GET | `/api/admin/ads` |
| POST | `/api/admin/ads` |
| PUT | `/api/admin/ads/:id` |
| DELETE | `/api/admin/ads/:id` |

```json
{
  "name": "Iklan Utama",
  "ad_code": "<div>...</div>",
  "device_type": "all",
  "is_active": true
}
```

`device_type`: `all` | `mobile` | `desktop`.

> Keamanan: `ad_code` di-render lewat `<iframe srcDoc>` dengan atribut
> `sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox"`.
> Jangan ganti ini dengan `dangerouslySetInnerHTML`.

---

## Upload

### Upload umum (gambar / poster)

```http
POST /api/admin/upload
Content-Type: multipart/form-data

file=<binary>
```

```json
{
  "success": true,
  "data": { "url": "/uploads/uuid.jpg", "filename": "uuid.jpg", "thumbnail_url": "" }
}
```

Format: `.jpg` `.jpeg` `.png` `.gif` `.mp4` `.webm` `.mov` `.avi`. Maks **500 MB**.
Tipe konten dicek lewat `http.DetectContentType`, bukan sekadar ekstensi.

Bila `CLOUDINARY_URL` terisi, file diunggah ke sana dan `url` berisi URL Cloudinary.
Thumbnail otomatis hanya berlaku untuk penyimpanan lokal.

### Upload video (langkah 1)

```http
POST /api/admin/upload/video
Content-Type: multipart/form-data

file=<binary>
```

```json
{
  "success": true,
  "data": {
    "url": "",
    "raw_path": "uploads/raw/uuid.mp4",
    "thumbnail_url": "/uploads/uuid.jpg",
    "duration": 300,
    "message": "Upload selesai. HLS akan diproses otomatis setelah video disimpan."
  }
}
```

`url` sengaja kosong — video baru bisa diputar setelah transcode HLS selesai.
`raw_path` dipakai pada langkah 3.

Thumbnail diambil pada 10% durasi untuk video > 50 detik, atau detik ke-5 untuk
yang lebih pendek. Video di bawah 5 detik mengambil frame dari tengah durasinya.

### Proses HLS (langkah 3)

```http
POST /api/admin/upload/process-hls
Content-Type: application/json

{ "video_id": 12, "raw_path": "uploads/raw/uuid.mp4" }
```

```json
{ "success": true, "data": { "message": "HLS processing dimulai", "video_id": 12 } }
```

Transcode berjalan di **background** — endpoint ini langsung mengembalikan, tidak
menunggu ffmpeg selesai. Field `video_url`, `hls_path`, dan `is_encrypted` terisi
saat proses selesai.

### Status HLS

```http
GET /api/videos/:id/hls-status
```

```json
{
  "success": true,
  "data": { "id": 12, "hls_status": "done", "hls_error": "", "video_url": "/api/hls/12/index.m3u8", "is_complete": true }
}
```

`hls_status`: `pending` → `processing` → `done` | `error`.
Saat `error`, `hls_error` berisi pesan dari ffmpeg.

### Upload chunked

Untuk file besar. Maksimal 500 chunk.

| Method | Path | Keterangan |
|---|---|---|
| POST | `/api/admin/upload/chunked/initiate` | Mulai sesi, dapat `upload_id` |
| POST | `/api/admin/upload/chunked/:upload_id/chunk` | Unggah satu chunk |
| POST | `/api/admin/upload/chunked/:upload_id/complete` | Gabungkan chunk |
| GET | `/api/admin/upload/chunked/:upload_id/progress` | Progres |
| DELETE | `/api/admin/upload/chunked/:upload_id` | Batalkan |

```bash
# 1. initiate
curl -X POST http://localhost:8080/api/admin/upload/chunked/initiate \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"filename":"video.mp4","file_size":104857600,"total_chunks":20}'

# 2. chunk (form field: chunk, chunk_index)
curl -X POST http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID/chunk \
  -H "Authorization: Bearer $TOKEN" \
  -F "chunk=@chunk_0" -F "chunk_index=0"

# 3. complete
curl -X POST http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID/complete \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{}'
```

> Sesi chunked disimpan di memori proses. Upload yang belum `complete` hilang
> setelah backend restart.

---

## File statis

| Path | Isi |
|---|---|
| `/uploads/*` | Gambar, thumbnail, dan file video yang sudah diunggah |
| `/favicon.ico` | Selalu 204 (menghindari 404 di log) |

---

## Contoh alur lengkap

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['token'])")

# Upload video
RAW=$(curl -s -X POST http://localhost:8080/api/admin/upload/video \
  -H "Authorization: Bearer $TOKEN" -F "file=@video.mp4" \
  | python3 -c "import sys,json;d=json.load(sys.stdin)['data'];print(d['raw_path'])")

# Buat record video
VIDEO_ID=$(curl -s -X POST http://localhost:8080/api/admin/videos \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"title\":\"Demo\",\"categories\":[\"Umum\"],\"is_published\":true}" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['id'])")

# Proses HLS
curl -X POST http://localhost:8080/api/admin/upload/process-hls \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"video_id\":$VIDEO_ID,\"raw_path\":\"$RAW\"}"

# Poll status
curl http://localhost:8080/api/videos/$VIDEO_ID/hls-status
```