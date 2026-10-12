# pgarcv

Platform video dengan streaming HLS terenkripsi. Backend Go (Gin + GORM), frontend Next.js.

- **Backend** — Go 1.25, Gin, GORM, SQLite (lokal) / PostgreSQL (production)
- **Frontend** — Next.js 16 (App Router), React 18, Tailwind CSS 4
- **Streaming** — HLS + AES-128, transcode via ffmpeg/ffprobe

Dokumentasi lengkap:

| Dokumen | Isi |
|---|---|
| [docs/API.md](docs/API.md) | Referensi seluruh endpoint (publik + admin), auth, contoh request |
| [docs/CONFIGURATION.md](docs/CONFIGURATION.md) | Environment variable, database, storage, CORS, HLS |

---

## Menjalankan di lokal (dev)

Prasyarat: **Go ≥ 1.25**, **Node.js ≥ 20**, **ffmpeg + ffprobe** di `PATH`.

### 1. Backend

```bash
cd backend
cp .env.example .env      # lalu sesuaikan bila perlu
go run cmd/server/main.go
```

Backend berjalan di `http://localhost:8080`.

> **Penting:** jalankan perintah ini dari dalam folder `backend/`. Nilai default
> `UPLOAD_PATH=./uploads`, `HLS_OUTPUT_PATH=./hls_output`, dan
> `DB_PATH=./videoplatform.db` bersifat relatif terhadap working directory.
> Menjalankan dari root repo akan membuat folder storage di lokasi yang salah.

Folder storage otomatis dibuat saat start. Backend berhenti dengan pesan jelas
bila folder tidak bisa dibuat atau tidak writable — bukan diam-diam gagal saat upload.

### 2. Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend berjalan di `http://localhost:3000`, dan rewrite `/api/*` serta
`*/uploads/*` diteruskan ke backend.

### 3. Login admin

Buka `http://localhost:3000/admin/login`.
Default dev: **admin / admin** (lihat `backend/.env`).

### Akses dari perangkat lain di jaringan lokal

Kalau membuka lewat IP LAN (mis. `http://192.168.1.16:3000` untuk tes dari HP),
origin tersebut **wajib** didaftarkan di `CORS_ORIGIN` pada `backend/.env`,
dipisahkan koma. Tanpa itu browser memblokir request dan upload gagal dengan
`NetworkError`. Restart backend setelah mengubahnya.

---

## Perintah

### Backend

```bash
go run cmd/server/main.go   # jalankan server
go build ./...              # compile semua paket
go vet ./...                # static analysis
go test ./...               # unit test
```

### Frontend

```bash
npm run dev                 # dev server (Turbopack)
npm run build               # production build
npm run lint                # ESLint
npm test                    # Vitest
npx tsc --noEmit            # type check
```

---

## Struktur project

```
backend/
  cmd/server/main.go        # entrypoint: rakit dependensi, jalankan server
  internal/
    config/                 # baca + validasi environment
    model/                  # entitas domain + mapping GORM
    repository/             # akses database, seed
    handler/                # HTTP handler per resource
    service/                # logika di luar HTTP (HLS job, sesi upload)
    middleware/             # auth JWT
    router/                 # definisi seluruh route
    hls/                    # processor ffmpeg (transcode, thumbnail, key)
    response/               # pembungkus JSON response
    storage/                # adapter Cloudinary
  .env                      # konfigurasi lokal (tidak di-commit)

frontend/
  app/
    (public)/               # halaman publik: home, search, watch
    admin/                  # area admin (dijaga proxy.ts)
  components/               # komponen UI
  hooks/useAdminAuth.ts     # cek sesi admin di client
  lib/api.ts                # klien API + helper URL
  proxy.ts                  # guard route /admin/* (Next 16 Proxy)
  next.config.js            # rewrite /api dan /uploads ke backend
```

---

## Alur upload video

Upload sengaja dipecah dua langkah supaya UI tidak menunggu transcode.

```
1. POST /api/admin/upload/video      → simpan file mentah + buat thumbnail
2. POST /api/admin/videos            → buat record video, kirim raw_path
3. POST /api/admin/upload/process-hls→ jalankan ffmpeg di background
4. GET  /api/videos/:id/hls-status   → polling sampai done / error
```

`video_url` baru terisi setelah transcode selesai, yaitu
`/api/hls/<id>/index.m3u8`. Video tidak bisa diputar sebelum itu.

Halaman admin mem-poll `hls-status` setiap 2 detik. Jika proses masih berjalan
setelah 60 kali poll (±2 menit), UI memberi tahu proses masih berjalan di latar.

---

## Catatan keamanan

- `.env` tidak pernah di-commit (`.gitignore`). Jangan masukkan kredensial production ke dalam repo.
- `ADMIN_PASSWORD` dan `JWT_SECRET` **wajib diganti** di production. Backend menolak start bila `JWT_SECRET` kosong atau < 32 karakter saat `ENV=production`.
- Route `/api/admin/*` dilindungi JWT HS256. Token diterima lewat cookie `admin_token` (httpOnly) maupun header `Authorization: Bearer`.
- Penyerang harus login untuk menjangkau endpoint admin — termasuk listing video draft.
- Nilai `ADMIN_USERNAME=admin` di production hanya memunculkan peringatan, tidak memblokir start.