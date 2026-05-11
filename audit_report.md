## 🔴 CRITICAL (harus difix sekarang)
- [frontend/lib/api.ts] [post()/put()/del()/get() & upload()] [Response parsing tidak konsisten & raw parsing `.data ?? json` menyebabkan type safety rusak (contoh: adminApi.videos.list() return tipe PaginatedResponse, tapi konsumen kadang mengakses struktur `r.data`, `r.videos`, bahkan `any`)] [Solusi singkat: standarkan contract response di backend (mis. selalu `{success,data,...}`) dan hapus parsing heuristik di frontend; type-kan generik sesuai contract]
- [frontend/app/admin/(authenticated)/videos/page.tsx] [~45-63] [Menggunakan `(r as any).data` setelah pengecekan `"data" in r` → tetap raw `any` dan potensi undefined/shape mismatch saat backend berubah] [Solusi singkat: buat type-guard yang benar untuk PaginatedResponse vs array; hilangkan cast `any`]
- [frontend/app/admin/(authenticated)/layout.tsx] [~70-112] [Hooks dipanggil setelah conditional return? Secara struktur terlihat aman, namun ada potensi pelanggaran Rules of Hooks karena early returns (`if (!checked) return ...` dan `if (!isAuth) return null`) muncul sebelum seluruh hook? (di file ini semua hook memang sudah dipanggil, tapi pola ini sering memicu regresi di perubahan berikutnya)] [Solusi singkat: jadikan pola render tetap setelah semua hook; pertimbangkan memindahkan guard ke komponen wrapper/parent]
- [backend/config/config.go] [JWTSecret default "secret-dev-only" dan ADMIN_PASSWORD default "admin123"] [Berpotensi fatal security: bila env produksi tidak di-set dengan benar, secret/password default bisa lolos ke lingkungan sebenarnya atau terlewat dalam deploy] [Solusi singkat: buat default kosong dan fail-fast; jangan pakai nilai default sensitif]
- [backend/internal/middleware/auth.go] [AuthRequired()] [Middleware mengandalkan cookie `admin_token` atau header `Authorization` tanpa validasi skema/format cookie domain/path; serta JWT secret berasal dari config; namun tidak ada pengecekan issuer/audience dan tidak set expiry handling khusus (hanya jwt.Parse)] [Solusi singkat: tambah claims validation (iss/aud), pastikan cookie options selaras, dan gunakan parsing dengan opsi yang lebih ketat]

## 🟡 WARNING (sebaiknya difix)
- [frontend/lib/api.ts] [get()/post()/put()/del()/upload()] [Error handling: tidak ada try/catch internal; beberapa consumer hanya `console.error` lalu set state; state loading bisa tidak di-reset pada beberapa skenario (terutama bila parsing JSON gagal karena res bukan JSON)] [Solusi singkat: tambahkan try/catch di util fetch wrapper; handle non-JSON response]
- [frontend/lib/api.ts] [del()] [Tidak memeriksa `res.ok`; potensi silent failure saat delete gagal] [Solusi singkat: cek `res.ok` dan parse error body]
- [frontend/lib/api.ts] [upload()] [Tidak memvalidasi bahwa `file` sesuai tipe/ukuran di client (validasi server ada, tapi UX bisa lebih cepat dengan validasi client)] [Solusi singkat: validasi ekstensi/mime dan ukuran sebelum upload]
- [frontend/app/admin/login/page.tsx] [~36-61] [Konsistensi response: ambil token dari `json.token || json.data?.token`—namun backend `auth_handler.go` hanya return `{"token": tokenStr, "message": ...}`. Ini aman, tapi kontrak tidak tegas] [Solusi singkat: buat type response khusus login dan validasi field yang wajib]
- [frontend/app/admin/login/page.tsx] [~27-29] [Tidak ada reset state `error` pada perubahan input (hanya di setError('') saat submit); UX mungkin membingungkan jika user edit setelah error] [Solusi singkat: reset error onChange]
- [frontend/app/admin/(authenticated)/layout.tsx] [~60-76] [Auth check: menggunakan `localStorage` + `useEffect([])` tetapi dependency `pathname` / route changes tidak diikutkan (ambil `usePathname()` namun effect tidak tergantung). Pada navigasi client-side, `pathname` berubah tapi auth-logic tidak re-evaluasi] [Solusi singkat: set dependency `[pathname, router]` atau gunakan logic di hook khusus]
- [frontend/app/admin/(authenticated)/layout.tsx] [~121-185] [Memory leaks: tidak ada cleanup untuk potensi timer/listener (di file ini tampaknya tidak ada), tapi gunakan audit menyeluruh untuk komponen anak] [Solusi singkat: pastikan seluruh komponen yang membuat listener/timer sudah cleanup]
- [frontend/components/video/VideoForm.tsx] [detectVideoDuration()] [Membuat objectUrl untuk metadata dan cleanup di finally (sudah benar), namun pembuatan `video` element tanpa `video.load()`/abort controller; serta tidak ada guard unmount untuk async duration detection] [Solusi singkat: gunakan Abort/flag isMounted]
- [frontend/components/video/VideoCard.tsx] [Cleanup] [IntersectionObserver cleanup sudah benar; setTimeout cleanup benar. Namun Set `activeVideos` global berpotensi berisi element yang sudah unmounted jika browser pause/cancel tak sesuai; ada cleanup di unmount effect, tapi tidak ada guard saat preview belum ada] [Solusi singkat: pastikan `if (videoRef.current)` dan reset state/hapus saat ref null]
- [backend/internal/handler/video_handler.go] [AdminList()] [response.Paginated(c, videos, total, 1, int(math.Max(float64(total), 1))) → `limit`/pagination parameter terlihat salah (limit dibuat dari total, bukan limit request). Ini bisa menyebabkan response total_pages keliru.] [Solusi singkat: perbaiki pagination contract]
- [backend/internal/handler/video_download_handler.go] [Download()] [Jika srcURL adalah absolute URL, menggunakan `http.Get` tanpa timeout → rentan hanging request/thread] [Solusi singkat: gunakan http.Client dengan timeout]
- [backend/internal/handler/video_download_handler.go] [Download()] [Mengakses `Content-Type` dari remote tanpa fallback; bisa jadi empty] [Solusi singkat: fallback ke `application/octet-stream`]

## 🟢 OK (tidak ada masalah)
- [frontend/components/video/BannerSlider.tsx] [useEffect cleanup] [Timer interval dibersihkan via cleanup return fungsi pada effect] 
- [frontend/components/video/VideoCard.tsx] [useEffect cleanup] [IntersectionObserver disconnect dan cleanup timeout/activeVideos dilakukan]
- [backend/internal/middleware/auth.go] [AuthRequired()] [Semua admin routes menggunakan middleware AuthRequired sesuai gin setup]
- [backend/internal/handler/upload_handler.go] [Upload validations] [Validasi ekstensi + batas ukuran + UUID untuk filename + defer src.Close]
- [backend/cmd/server/main.go] [CORS & ngrok header allowance] [AllowOriginFunc menerima subdomain ngrok; AllowCredentials=true dan AllowHeaders mencakup Authorization]

## 📋 RINGKASAN
- Total critical: 5
- Total warning: 12
- Estimasi waktu fix: ~240 menit

