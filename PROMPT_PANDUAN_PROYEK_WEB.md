# 🎬 PANDUAN LENGKAP MEMBANGUN VIDEO PLATFORM WEB
> Stack: Next.js (Frontend) · Go (Backend) · SQLite → PostgreSQL (Database)
> Arsitektur: Monorepo dengan pemisahan `frontend/` dan `backend/`

---

## 📁 ARSITEKTUR YANG DISARANKAN

```
project-root/
├── frontend/                  # Next.js App
│   ├── app/
│   │   ├── (public)/          # Halaman publik (layout utama)
│   │   │   ├── page.tsx       # Beranda
│   │   │   ├── watch/[id]/    # Halaman tonton video
│   │   │   └── search/        # Hasil pencarian
│   │   ├── (admin)/           # Admin panel (layout terpisah)
│   │   │   ├── dashboard/
│   │   │   ├── videos/
│   │   │   ├── banners/
│   │   │   └── ads/
│   │   └── layout.tsx
│   ├── components/
│   │   ├── ui/                # Komponen atom (Button, Input, dll)
│   │   ├── video/             # VideoCard, VideoPlayer, dll
│   │   ├── admin/             # Komponen khusus admin
│   │   └── layout/            # Navbar, Sidebar, Footer
│   ├── lib/
│   │   ├── api.ts             # Semua API call ke backend
│   │   ├── theme.ts           # Konfigurasi tema
│   │   └── utils.ts
│   ├── styles/
│   │   └── globals.css        # CSS variables (light/dark token)
│   └── public/
│       └── logo.svg           # Logo kamu
│
├── backend/                   # Go API Server
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── handler/           # HTTP handler per fitur
│   │   ├── service/           # Business logic
│   │   ├── repository/        # Query database
│   │   └── middleware/        # Auth, CORS, rate limiter
│   ├── pkg/
│   │   ├── storage/           # File upload logic
│   │   └── response/          # Standar response helper
│   ├── migrations/            # SQL migration files
│   └── config/config.go
│
└── docker-compose.yml         # Untuk development lokal
```

> **💡 Rekomendasi Database:**
> - **Sekarang:** SQLite (via `gorm` + `mattn/go-sqlite3`) — zero config, file tunggal, cocok untuk dev
> - **Nanti migrasi ke:** PostgreSQL — robust, support concurrent writes, cocok produksi
> - GORM dipakai agar migrasi tinggal ganti driver, **tanpa ubah query**

---

## 🪜 STEP-BY-STEP PEMBANGUNAN

---

### ✅ STEP 1 — Setup Proyek & Struktur Awal

**Prompt ke AI:**
```
Buatkan struktur folder monorepo untuk proyek video platform dengan:
- frontend/ menggunakan Next.js 14 (App Router, TypeScript)
- backend/ menggunakan Go dengan gin framework
- File: .gitignore, README.md, docker-compose.yml untuk dev lokal
- Di frontend, setup Tailwind CSS dengan font Poppins/Montserrat dari Google Fonts
- Di backend, setup go.mod dengan dependensi: gin, gorm, go-sqlite3, godotenv, cors
Sertakan perintah instalasi lengkap.
```

---

### ✅ STEP 2 — Sistem Tema (Light/Dark) & Design Token

**Prompt ke AI:**
```
Buatkan sistem tema light/dark untuk Next.js menggunakan:
- CSS custom properties (variables) di globals.css sebagai design token
- next-themes untuk toggle tema
- Font default: Poppins (heading) + Montserrat (body), import via next/font/google
- Token warna meliputi: background, surface, text-primary, text-secondary,
  accent, border, card-bg, overlay
- Komponen ThemeToggle dengan animasi smooth
- Dark mode menggunakan class strategy (.dark)
Pastikan semua komponen selanjutnya HANYA menggunakan CSS variables, bukan hardcode warna.
```

---

### ✅ STEP 3 — Logo SVG & Branding

**Prompt ke AI:**
```
Saya sudah punya logo dalam format SVG. Buatkan:
- Komponen Logo.tsx yang me-render SVG logo saya
- Versi: default (full), compact (icon only), light variant, dark variant
- Logo harus responsif dan bisa menerima prop: size, variant
- Dua warna utama mengikuti CSS variable accent dan text-primary
File SVG logo saya: [paste konten SVG kamu di sini]
```

---

### ✅ STEP 4 — Database Schema & Migrasi (Backend Go)

**Prompt ke AI:**
```
Buatkan skema database SQLite menggunakan GORM di Go untuk video platform dengan tabel:
- videos: id, title, subtitle, description, thumbnail_url, video_url,
  duration, views, is_published, created_at, updated_at
- banners: id, image_url, link_url, title, is_active, position, created_at
- ads: id, ad_code, device_type (mobile/desktop/all), is_active,
  max_per_device_per_day, created_at
- admins: id, username, password_hash, created_at
- Buatkan file migration SQL dan GORM AutoMigrate
- Sertakan seed data contoh (3 video, 1 banner, 1 admin)
```

---

### ✅ STEP 5 — Backend API: Video & Upload (Go)

**Prompt ke AI:**
```
Buatkan REST API di Go (gin framework) untuk fitur video dengan:
Endpoint publik:
  GET  /api/videos          → list video (pagination, search)
  GET  /api/videos/:id      → detail video + increment view
  GET  /api/videos/featured → video unggulan untuk homepage

Endpoint admin (butuh JWT middleware):
  POST   /api/admin/videos        → upload video baru
  PUT    /api/admin/videos/:id    → edit video
  DELETE /api/admin/videos/:id    → hapus video
  POST   /api/admin/upload        → upload file (thumbnail/video) ke /uploads/

Tambahan:
- File upload disimpan di backend/uploads/ dengan nama unik (UUID)
- Response standar: { success, data, message, pagination? }
- CORS allow dari localhost:3000
- Maksimal upload: 500MB video, 5MB thumbnail
```

---

### ✅ STEP 6 — VideoCard Component dengan Preview

**Prompt ke AI:**
```
Buatkan komponen VideoCard.tsx di Next.js dengan fitur mirip YouTube tapi dimodifikasi:

Tampilan card:
- Thumbnail dengan aspect ratio 16:9
- FITUR PREVIEW: saat hover 1 detik, putar preview video pendek (gif/video clip)
  menggunakan HTML video element autoplay muted loop
- Badge durasi di pojok kanan bawah thumbnail
- Avatar/icon channel di bawah thumbnail (kiri)
- Title (bold, 2 baris max dengan ellipsis)
- Subtitle (teks kecil di bawah title, warna text-secondary, 1 baris)
- Deskripsi singkat (hidden by default, muncul saat hover card pada desktop)
- Jumlah view + tanggal upload

Props: { id, title, subtitle, description, thumbnail, previewUrl,
         duration, views, uploadDate, onClick }

Gunakan CSS variables untuk warna. Animasi smooth dengan Framer Motion.
Responsive: grid 1 col mobile, 2 col tablet, 3-4 col desktop.
```

---

### ✅ STEP 7 — Halaman Beranda (Homepage)

**Prompt ke AI:**
```
Buatkan halaman beranda (app/(public)/page.tsx) untuk video platform dengan:
- Hero section: Banner/slider dari data API banners (auto-slide 5 detik, manual arrow)
- Section "Video Terbaru": grid VideoCard dengan lazy loading
- Section "Video Populer": horizontal scroll carousel
- Infinite scroll atau tombol "Muat Lebih" untuk pagination
- Loading skeleton yang smooth saat fetch data
- Semua data di-fetch dari backend Go via fetch() dengan revalidate 60 detik
- Gunakan Suspense + loading.tsx untuk streaming
```

---

### ✅ STEP 8 — Halaman Tonton Video (Watch Page)

**Prompt ke AI:**
```
Buatkan halaman watch (app/(public)/watch/[id]/page.tsx) dengan:
- Video player custom menggunakan HTML5 video (fullscreen, volume, progress bar,
  keyboard shortcut: space=pause, f=fullscreen, arrow=seek 10s)
- Di bawah player: title besar, subtitle kecil, view count, tanggal
- Tombol share
- Deskripsi lengkap (collapsible jika panjang)
- Sidebar: "Video Lainnya" (VideoCard versi compact/vertikal)
- Layout: player area 70% + sidebar 30% di desktop, stack di mobile
- Increment view count saat video mulai diputar (POST /api/videos/:id/view)
```

---

### ✅ STEP 9 — Sistem Iklan (1 Iklan per Device)

**Prompt ke AI:**
```
Buatkan sistem iklan untuk video platform dengan ketentuan:
1 device hanya melihat 1 iklan per sesi (bukan per hari dulu, cukup per session)

Backend (Go):
- GET /api/ads/serve → return 1 iklan aktif berdasarkan device type
- Pakai User-Agent untuk deteksi mobile/desktop
- Logic: random iklan aktif yang sesuai device type

Frontend (Next.js):
- Komponen AdSlot.tsx:
  - Cek localStorage: apakah sudah ada "ad_shown" di session ini
  - Jika belum: fetch iklan, render, simpan flag di sessionStorage
  - Jika sudah: render div kosong (tidak muncul iklan)
  - Support posisi: banner-top, banner-bottom, sidebar
  - Iklan bisa berupa: HTML embed code (Google AdSense, dll)
- Tempatkan AdSlot di: homepage (bawah hero) dan watch page (bawah player)
```

---

### ✅ STEP 10 — Admin Authentication (JWT)

**Prompt ke AI:**
```
Buatkan sistem autentikasi admin dengan JWT di Go:
- POST /api/admin/login → validasi username+password, return JWT token (expire 24 jam)
- Middleware AuthRequired: validasi Bearer token di header Authorization
- Di frontend Next.js:
  - Halaman /admin/login dengan form sederhana
  - Simpan JWT di httpOnly cookie via Route Handler (bukan localStorage)
  - Middleware Next.js: redirect ke /admin/login jika tidak ada token
  - Hook useAdmin() untuk cek status auth
```

---

### ✅ STEP 11 — Admin Dashboard

**Prompt ke AI:**
```
Buatkan Admin Dashboard (app/(admin)/dashboard/page.tsx) dengan:
Layout:
- Sidebar navigasi: Dashboard, Video, Banner, Iklan, Pengaturan
- Header dengan nama admin + tombol logout
- Konten utama

Halaman Dashboard (ringkasan):
- Stat cards: total video, total views, video terbaru
- Tabel 5 video terbaru dengan link ke edit

Desain: clean, profesional, gunakan CSS variables tema yang sama.
Sidebar collapsible di mobile.
```

---

### ✅ STEP 12 — Admin Video Management

**Prompt ke AI:**
```
Buatkan halaman manajemen video di admin (app/(admin)/videos/page.tsx):

Halaman list video:
- Tabel dengan kolom: thumbnail kecil, title, subtitle, views, status (published/draft), aksi
- Tombol: Tambah Video, Edit, Hapus (konfirmasi modal)
- Search dan filter status

Halaman tambah/edit video (app/(admin)/videos/new & /edit/[id]):
- Form fields: title, subtitle, deskripsi (rich text sederhana / textarea)
- Upload thumbnail: drag & drop + preview gambar langsung
- Upload video: drag & drop dengan progress bar upload
- Input URL preview (untuk fitur hover preview di VideoCard)
- Toggle: Published / Draft
- Tombol Simpan + Simpan sebagai Draft

Semua upload ke POST /api/admin/upload, response berupa URL file.
```

---

### ✅ STEP 13 — Admin Banner Management

**Prompt ke AI:**
```
Buatkan halaman manajemen banner di admin (app/(admin)/banners/page.tsx):
- List banner: preview gambar kecil, judul, posisi, status aktif/nonaktif
- Form tambah/edit: upload gambar banner (preview langsung), input link URL,
  judul, toggle aktif, urutan tampil
- Drag & drop reorder banner (opsional, atau pakai input angka urutan)
- Endpoint backend yang dipakai: GET/POST/PUT/DELETE /api/admin/banners
```

---

### ✅ STEP 14 — Memory Management & Performance

**Prompt ke AI:**
```
Implementasikan optimasi performa untuk video platform Next.js:

1. Image Optimization:
   - Gunakan next/image untuk semua thumbnail dan banner
   - Konfigurasi domains di next.config.js untuk domain backend
   - Lazy loading otomatis, blur placeholder

2. Video Memory Management:
   - VideoCard: saat tidak di viewport (IntersectionObserver), pause & unload preview video
   - Batasi maksimal 3 preview video yang aktif bersamaan (queue management)
   - Gunakan useRef + cleanup di useEffect

3. Data Fetching:
   - Server Components untuk data awal (SSR/SSG)
   - SWR atau React Query untuk client-side data yang perlu refresh
   - Pagination cursor-based (bukan offset) untuk list video

4. Bundle:
   - Dynamic import untuk VideoPlayer dan komponen berat
   - next.config.js: aktifkan swcMinify, compress

Buatkan implementasi lengkap untuk poin 2 (VideoCard memory management)
karena ini yang paling kritikal.
```

---

### ✅ STEP 15 — Polish, SEO & Deployment Prep

**Prompt ke AI:**
```
Buatkan konfigurasi final untuk video platform:

SEO (Next.js):
- generateMetadata() di setiap page (homepage, watch page)
- Open Graph tags untuk share ke sosial media
- Structured data (JSON-LD) untuk video di watch page

Error Handling:
- error.tsx global dan per-route
- not-found.tsx dengan desain yang sesuai tema
- Loading states dan skeleton di semua halaman

Backend (Go):
- Rate limiting sederhana (gin-contrib/limiter): maks 100 req/menit per IP
- Graceful shutdown
- Log request dengan format yang rapi

Environment:
- .env.example untuk frontend dan backend
- next.config.js production-ready
- README.md dengan instruksi setup lengkap
```

---

## 🗺️ URUTAN PENGERJAAN (REKOMENDASI)

```
Step 1  →  Step 2  →  Step 3     # Fondasi & branding
    ↓
Step 4  →  Step 5               # Backend & database
    ↓
Step 6  →  Step 7  →  Step 8    # Halaman publik
    ↓
Step 9                          # Sistem iklan
    ↓
Step 10 →  Step 11 →  Step 12 → Step 13   # Admin panel
    ↓
Step 14 →  Step 15              # Optimasi & finishing
```

---

## 💡 IDE TAMBAHAN (Catat untuk Nanti)

| Fitur | Kapan dipertimbangkan |
|---|---|
| Komentar video | Setelah user system ada |
| Like / Reaksi | Setelah video stabil |
| Playlist | Konten sudah banyak |
| Notifikasi push | Sudah ada user terdaftar |
| CDN untuk video | Saat traffic mulai tinggi |
| Redis cache | Saat SQLite mulai lambat |
| PostgreSQL migrasi | Sebelum production launch |
| Analytics dashboard | Setelah ada traffic nyata |

---

## 🔧 TECH STACK SUMMARY

| Layer | Teknologi | Catatan |
|---|---|---|
| Frontend | Next.js 14 (App Router) | TypeScript, Tailwind |
| UI | Framer Motion, next-themes | Animasi & dark mode |
| Font | Poppins + Montserrat | Via next/font/google |
| Backend | Go + Gin | REST API |
| ORM | GORM | Mudah migrasi DB |
| Database (dev) | SQLite | Zero config |
| Database (prod) | PostgreSQL | Rekomendasi migrasi |
| File Storage | Local `/uploads` | Nanti bisa S3/R2 |
| Auth | JWT (simpan di cookie) | Hanya untuk admin |
| Deploy | VPS / Railway / Fly.io | Terserah nanti |

---

> **📌 Tips:** Setiap step bisa langsung dipaste ke AI coding assistant (Cursor, Claude, dll).
> Selesaikan satu step sebelum lanjut ke berikutnya agar struktur tetap rapi.
> Kalau ada ide baru di tengah jalan, catat dulu di bagian "IDE TAMBAHAN" — jangan langsung diimplementasi agar fokus tidak pecah.
