# Menjalankan dengan Docker

Stack produksi lokal: backend Go + frontend Next.js dalam container.

```bash
cp .env.docker.example .env.docker

# Isi secret sebelum pertama kali jalan
openssl rand -hex 32   # JWT_SECRET
openssl rand -hex 16   # HLS_ENCRYPT_KEY
# ADMIN_PASSWORD juga wajib diganti

docker compose --env-file .env.docker up -d --build
```

Frontend: <http://localhost:3000>

---

## Perintah

```bash
docker compose --env-file .env.docker up -d --build   # build + start
docker compose --env-file .env.docker logs -f backend
docker compose --env-file .env.docker ps
docker compose --env-file .env.docker down             # stop
docker compose --env-file .env.docker down -v          # stop + HAPUS semua data
```

> Selalu pakai `--env-file .env.docker`. Compose tidak membaca `.env` milik
> backend, jadi tanpa flag itu semua secret dianggap kosong dan container
> gagal start.

---

## Cara kerja request

```
Browser  →  localhost:3000/api/videos
         →  rewrite Next.js  →  http://backend:8080/api/videos
         →  container backend
```

Browser **tidak pernah** bicara langsung ke backend, dan tidak perlu tahu
alamat internal container. Backend hanya di-bind ke `127.0.0.1`, jadi tidak
terekspos ke jaringan lokal.

Dua aturan yang mudah terlewat:

| Nilai | Dievaluasi | Dipakai untuk |
|---|---|---|
| `API_URL` (build arg **dan** runtime env) | build + runtime | rewrite Next.js, fetch SSR |
| `NEXT_PUBLIC_API_URL` | build, di-inline ke browser | path relatif, dibiarkan kosong |

`rewrites()` dievaluasi saat **build** lalu di-freeze ke `routes-manifest.json` —
env var runtime tidak akan terbaca di sana. Karena itu `API_URL` perlu
sekali sebagai build arg dan sekali lagi sebagai environment variable runtime.

---

## Data persisten

Tiga named volume, isinya tidak hilang saat container di-restart:

| Volume | Isi |
|---|---|
| `db` | SQLite (`videoplatform.db`) |
| `uploads` | Gambar, thumbnail, video mentah |
| `hls` | Segmen HLS dan kunci enkripsi |

`down -v` menghapus ketiganya, termasuk database.

Untuk PostgreSQL, isi `DATABASE_URL` di `.env.docker` — volume `db` lalu tidak
dipakai.

---

## Image

| Service | Base | Ukuran | Catatan |
|---|---|---|---|
| backend | `golang:1.25-alpine` → `alpine:3.20` | ~280MB | Binary statis (`CGO_ENABLED=0`), ffmpeg terpasang |
| frontend | `node:20-alpine` | ~450MB | `npm ci` + build terpisah dari runtime |

Kedua container berjalan sebagai user non-root (uid 10001) dan punya
HEALTHCHECK yang memanggil endpoint publik, jadi container ditandai unhealthy
bila DB atau routing bermasalah — bukan cuma saat proses mati.

---

## Troubleshooting

**`dependency failed to start: container pgarcv-backend is unhealthy`**

```bash
docker compose --env-file .env.docker logs backend
```

Penyebab yang sering terjadi:

| Pesan | Penyebab | Solusi |
|---|---|---|
| `JWT_SECRET wajib diisi` | `.env.docker` belum ada / kosong | `cp .env.docker.example .env.docker` lalu isi |
| `unable to open database file: out of memory (14)` | volume `db` milik root, proses non-root tak bisa tulis | `docker compose down -v` lalu build ulang (nama volume yang sudah ada tidak mewarisi ownership dari image) |
| `JWT_SECRET harus minimal 32 karakter` | secret terlalu pendek | `openssl rand -hex 32` |

**Thumbnail 404**

`placeholder.jpg` dibuat otomatis saat seeding. Kalau volume `uploads` sudah
terisi tapi file-nya hilang, hapus volumenya agar dibuat ulang.

**`ECONNREFUSED backend:8080` dari frontend**

Pastikan service-nya bernama `backend` di Compose dan keduanya ada di network
yang sama.

**Perubahan `next.config.js` tidak efek**

Config di-bake saat build. Jalankan `up -d --build frontend`, bukan cuma `restart`.