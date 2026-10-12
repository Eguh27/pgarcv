# Deploy

Ringkasan batas free tier per platform, dan kenapa project ini tidak serta-merta
bisa di-host gratis.

> Status: **saat ini project ini dijalankan lokal** (dev mode atau Docker).
> Dokumen ini hanya referensi kalau nanti mau deploy ke server sungguhan.

---

## Ringkasan

| Bagian | Paket | Rekomendasi |
|---|---|---|
| **Frontend** (Next.js) | Vercel Hobby | ✅ Cocok — 100GB bandwidth, 5.000 image transform/bulan |
| **Backend** (Go + ffmpeg) | — | ⚠️ Tidak ada free tier yang benar-benar pas |

---

## Kenapa backend sulit di-host gratis

Backend ini bukan API JSON biasa. Ia butuh:

| Kebutuhan | Alasan |
|---|---|
| **ffmpeg + ffprobe** | Transcode HLS + ambil thumbnail |
| **Persistent disk** | Upload & segmen HLS harus bertahan setelah restart |
| **≥ 2 GB RAM** | ffmpeg `libx264` butuh memori |
| **Proses background** | Transcode berjalan di goroutine terpisah |

 Bandingkan dengan spesifikasi free tier yang benar-benar tersedia:

| Platform | Free tier | Masalah untuk project ini |
|---|---|---|
| **Render** | 512MB RAM, 0.1 CPU | ⚠️ Filesystem **ephemeral** — file hilang tiap redeploy/spin-down |
| **Koyeb** | 512MB, 0.1 vCPU, 2GB SSD | ❌ **Tidak bisa pasang volume** di free instance |
| **Fly.io** | — | ❌ Tidak ada free tier untuk user baru sejak 2024 |
| **Railway** | $1 credit/bulan | ❌ Hanya trial, lalu berbayar |
| **Vercel** | 100GB bandwidth | ❌ Timeout 60 detik & 4 jam CPU — ffmpeg tidak cukup |

Poin pembatal terbesar untuk Render & Koyeb sama: **keduanya punya filesystem
ephemeral di paket free**. Video dan segmen HLS akan hilang setiap kali
container di-restart.

---

## Opsi backend

### 1. Oracle Cloud Always Free — satu-satunya yang cocok gratis

| Resource | Kuota |
|---|---|
| Ampere A1 (ARM) | 2 OCPU / 12 GB RAM |
| Block storage | 200 GB |
| Masa berlaku | Sepanjang umur akun (tidak ada kedaluwarsa) |

✅ Cukup untuk ffmpeg, disk persisten, dan binary kamu jalan di ARM karena
dikompilasi `CGO_ENABLED=0` (statis).

⚠️ Kendala:
- Regional tertentu sering **kehabisan kapasitas** saat provisioning
- Memerlukan **kartu kredit** untuk verifikasi (tidak ditagih)
- Kuota A1 turun dari 4 OCPU/24 GB menjadi 2 OCPU/12 GB (Juni 2026)

### 2. Ubah arsitektur

Videonya tetap di komputer lokal; hosting gratis hanya untuk API + metadata.
Butuh penyesuaian pada kode agar storage tidak lokal.

### 3. VPS kecil berbayar

Kalau budget > 0, ini paling andal. VPS ~Rp 30–50rb/bulan jauh lebih nyaman
daripada free tier yang sering spin-down.

---

## Kenapa Supabase/Neon gratis bukan pilihan lebih baik

Free tier database **tidak menggantikan** persistent disk yang dibutuhkan
project ini:

| Layanan | Free tier | Masalah |
|---|---|---|
| **Supabase** | 500 MB | Project **di-pause setelah 1 minggu idle**, harus restore manual |
| **Neon** | 0.5 GB/project | Tidak ada expiry, tapi scale-to-zero + storage kecil |
| **Render Postgres** | 1 GB | **Dihapus setelah 30 hari** |
| **Aiven** | 5 GB | Single node, paling besar di daftar |

Kalau nanti tetap memakai DB gratis: **Neon** lebih aman daripada Supabase
karena tidak ada auto-pause.

---

## Checklist sebelum deploy ke server sungguhan

- [ ] Ganti `JWT_SECRET` dan `ADMIN_PASSWORD` (minimal 32 karakter untuk JWT)
- [ ] Set `ENV=production`
- [ ] Set `COOKIE_SECURE=true` (frontend harus diakses via HTTPS)
- [ ] Set `SERVER_URL` dan `CORS_ORIGIN` ke domain yang sebenarnya
- [ ] Pastikan `UPLOAD_PATH` & `HLS_OUTPUT_PATH` menunjuk **volume persisten**
- [ ] **Hapus `dangerouslyAllowLocalIP: true`** dari `frontend/next.config.js`
- [ ] Update `remotePatterns` di `next.config.js` untuk domain backend sebenarnya
- [ ] Set `NEXT_PUBLIC_API_URL` ke URL backend publik

> 🔒 `dangerouslyAllowLocalIP` hanya aman selama backend berada di jaringan
> privat. Backend publik + flag ini = SSRF.

---

## Catatan arsitektur khusus image & video

- Bandwidth adalah biaya utama. Video yang sirve langsung dari backend boros;
  pertimbangkan CDN atau signed URL untuk HLS.
- Transcode HLS berat. Untuk traffic nyata, pindahkan ke worker terpisah
  (queue) supaya request upload tidak menunggu proses ffmpeg.
- Kalau `HLS_ENCRYPT_KEY` aktif, key endpoint sudah ada proteksi referer —
  pastikan domain frontend terdaftar dengan benar.