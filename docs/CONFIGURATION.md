# Konfigurasi

Semua konfigurasi backend dibaca dari environment variable. `backend/.env`
dibaca otomatis saat start (jika file ada) — tapi **nilai yang sudah ada di
environment sistem selalu menang**, jadi `backend/.env` bisa dipakai untuk
kebutuhan lokal tanpa mengganggu settingan container atau systemd.

`.env` tidak pernah di-commit. Lihat `.env.example` untuk daftar lengkap.

---

## Environment variable

### Server

| Variable | Default | Keterangan |
|---|---|---|
| `ENV` | `development` | Isi `production` untuk aktifkan validasi ketat |
| `PORT` | `8080` | Port listen backend |

### Database

| Variable | Default | Keterangan |
|---|---|---|
| `DB_PATH` | `./videoplatform.db` | Path file SQLite. Dipakai **hanya** bila `DATABASE_URL` kosong |
| `DATABASE_URL` | *(kosong)* | Connection string PostgreSQL. Kosong = SQLite |

saat `DATABASE_URL` terisi, `DB_PATH` diabaikan dan driver `postgres` yang dipakai.
Skema tabel dibuat otomatis lewat `AutoMigrate` pada tabel Video, Category,
Genre, Banner, dan Ad.

### Storage

| Variable | Default | Keterangan |
|---|---|---|
| `UPLOAD_PATH` | `./uploads` | Folder gambar, thumbnail, dan video mentah |
| `HLS_OUTPUT_PATH` | `./hls_output` | Folder output transcode HLS |

Isi folder setelah upload dan proses HLS:

```
uploads/
├── <uuid>.jpg             thumbnail
├── <uuid>.mp4             file dari endpoint upload umum
├── placeholder.jpg
└── raw/
    └── <uuid>.mp4         video awaiting transcode

hls_output/
├── thumbnails/            thumbnail sementara sebelum dipindahkan
└── hls/
    └── video_<id>/
        ├── index.m3u8
        ├── seg_000.ts ...
        ├── enc.key        kunci AES-128 (bila enkripsi aktif)
        └── enc.keyinfo
```

Folder dibuat dan diperiksa bisa ditulis saat start. Bila gagal, backend berhenti
dengan pesan yang menunjuk ke folder bermasalah.

> Path relatif dihitung dari **working directory**. Jalankan server dari dalam
> `backend/`, bukan dari root repo.

### Auth

| Variable | Default | Keterangan |
|---|---|---|
| `ADMIN_USERNAME` | `admin` | Username admin |
| `ADMIN_PASSWORD` | `admin123` (dev) | Password admin |
| `JWT_SECRET` | fallback dev (dev) | Kunci tanda tangan HS256 |
| `COOKIE_SECURE` | `false` | `true` = cookie hanya lewat HTTPS |

Saat `ENV=production`, backend **menolak start** bila `JWT_SECRET` kosong atau
lebih pendek dari 32 karakter, atau bila `ADMIN_PASSWORD` kosong.

Generate secret:

```bash
openssl rand -hex 32    # untuk JWT_SECRET
openssl rand -hex 16    # untuk HLS_ENCRYPT_KEY
```

### HLS & media

| Variable | Default | Keterangan |
|---|---|---|
| `HLS_ENCRYPT_KEY` | *(kosong)* | 32 karakter hex = 16 byte untuk AES-128 |
| `SERVER_URL` | `http://localhost:8080` | URL publik backend |
| `CLOUDINARY_URL` | *(kosong)* | `cloudinary://<key>:<secret>@<cloud>` |

`HLS_ENCRYPT_KEY` kosong berarti transcode berjalan tanpa enkripsi.

`SERVER_URL` dipakai untuk dua hal: membangun URL key HLS yang ditulis ffmpeg ke
playlist, dan memvalidasi `Referer`/`Origin` saat key diminta.

`CLOUDINARY_URL` opsional. Kalau terisi, upload gambar/videoumum diarahkan ke
Cloudinary. Video tetap diproses lokal menjadi HLS.

### CORS

| Variable | Default | Keterangan |
|---|---|---|
| `CORS_ORIGIN` | `http://localhost:3000` | Origin frontend yang diizinkan, dipisah koma |

```bash
CORS_ORIGIN=http://localhost:3000,http://192.168.100.78:3000
```

Di luar production, `http://localhost:3000` dan `http://localhost:8080`
ditambahkan otomatis. **IP LAN tidak bisa ditebak** — harus ditulis manual.
Kalau origin tidak terdaftar, preflight dibalas 403 dan browser memblokir
request, sehingga upload gagal dengan `NetworkError`.

### Akses dari perangkat lain di jaringan lokal

IP LAN berubah saat ganti router atau hotspot. Kalau tidak sengaja diperbarui,
gambar tidak tampil dan request API gagal. Dua nilai ini harus sinkron:

**`frontend/.env.local`**
```ini
LAN_HOST=192.168.100.78
```

**`backend/.env`**
```ini
CORS_ORIGIN=http://localhost:3000,http://192.168.100.78:3000
```

`LAN_HOST` dibaca `next.config.js` untuk `remotePatterns` next/image.
Kalau tidak diisi, cukup hapus baris `LAN_HOST` — `remotePatterns` LAN
otomatis dilewati dan sisanya tetap jalan.

Restart **frontend** dan **backend** setiap kali IP berubah.
Cara cek IP LAN:

```bash
ip -4 addr show | grep -oP 'inet \K[0-9.]+' | grep -v '^127'
```

---

## Setting development

`backend/.env` untuk lokal:

```ini
ENV=development
PORT=8080

DB_PATH=./videoplatform.db
DATABASE_URL=

UPLOAD_PATH=./uploads
HLS_OUTPUT_PATH=./hls_output

JWT_SECRET=dev-only-secret-tidak-untuk-production-32char
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin

HLS_ENCRYPT_KEY=c848ea11c0ee347eccb882a384c0e9c9

SERVER_URL=http://localhost:8080
CORS_ORIGIN=http://localhost:3000
COOKIE_SECURE=false
CLOUDINARY_URL=
```

`frontend/.env.local`:

```ini
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## Setting production

Disuplai lewat environment variable platform, **jangan** di-commit.

```ini
ENV=production
PORT=8080

DATABASE_URL=postgresql://user:password@host:port/dbname?sslmode=require
UPLOAD_PATH=/var/uploads
HLS_OUTPUT_PATH=/var/hls_output

JWT_SECRET=<openssl rand -hex 32>
ADMIN_USERNAME=<username kuat>
ADMIN_PASSWORD=<password kuat>

HLS_ENCRYPT_KEY=<openssl rand -hex 16>
SERVER_URL=https://api.domain-anda.com
CORS_ORIGIN=https://domain-frontend-anda.com
COOKIE_SECURE=true
CLOUDINARY_URL=cloudinary://<api_key>:<api_secret>@<cloud_name>
```

Yang wajib diubah dari default di production:

| Nilai | Kenapa |
|---|---|
| `JWT_SECRET` | Wajib ≥ 32 karakter, backend menolak start bila tidak |
| `ADMIN_PASSWORD` | Wajib diisi |
| `ADMIN_USERNAME` | Default `admin` hanya memunculkan peringatan |
| `COOKIE_SECURE` | Wajib `true`, kalau tidak cookie tetap terkirim tanpa HTTPS |
| `CORS_ORIGIN` | Wajib domain frontend yang sebenarnya |

> Pada filesystem/container yang immutable, `UPLOAD_PATH` dan `HLS_OUTPUT_PATH`
> harus menunjuk volume persisten. Kalau tidak, file hilang setiap deploy.

---

## Troubleshooting

### Upload gagal dengan `Gagal simpan file` atau `Gagal menyimpan file`

Penyebab: folder storage tidak ada atau tidak writable. Cek nilai `UPLOAD_PATH`,
lalu pastikan proses punya izin tulis:

```bash
ls -ld uploads hls_output
touch uploads/.probe && rm uploads/.probe
```

Karena path relatif terhadap working directory, pastikan server dijalankan dari
dalam `backend/`.

### Upload gagal dengan `NetworkError when attempting to fetch resource`

Penyebab: CORS. Backend menolak origin yang tidak terdaftar. Cek `CORS_ORIGIN`,
kemudian restart backend. Umum terjadi saat frontend dibuka lewat IP LAN
(`http://192.168.x.x:3000`) alih-alih `localhost`.

### `localStorage is not available` di log Node

Peringatan eksperimental Node 26, bukan error aplikasi. Browser tetap
menyediakan localStorage normal.

### `<script>` di dalam komponen React

Ditاخلkan oleh kode iklan (`ad_code`) dan di-render di `<iframe srcDoc>` yang
disandbox — itu pola yang benar dan terisolasi dari halaman utama.

### HLS status macet di `processing`

ffmpeg masih berjalan, atau prosesnya mati. Detail error ada di field
`hls_error` pada endpoint `hls-status`, dan di log backend.

### Transcode gagal: `HLS_ENCRYPT_KEY harus 32 hex chars`

Nilainya harus 32 karakter heksadesimal (16 byte), bukan 32 byte.

```bash
openssl rand -hex 16
```

---

## Verifikasi cepat

```bash
# Backend hidup
curl -s http://localhost:8080/api/videos | head -c 100

# Origin terdaftar (harus 204, bukan 403)
curl -s -o /dev/null -w "%{http_code}\n" -X OPTIONS http://localhost:8080/api/admin/videos \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET"

# ffmpeg tersedia
ffmpeg -version | head -1 && ffprobe -version | head -1
```