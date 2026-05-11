# EDIT_NEXT_STEPS_IMPLEMENTATION.md

## Lanjutan dari tahap backend saat ini
Backend sudah punya fields `Category` & `Genre` dan search memfilter keduanya.

## Target implementasi berikutnya
### Backend
1) Tambah endpoint publik
- `GET /api/videos/suggest?q=...` => return list video minimal (id,title,subtitle,category,genre,thumbnail_url)
- `GET /api/videos/:id/download` => stream file dari `video_url`

2) Tambah logic download
- Jika `video_url` diawali `/uploads/` dan file ada di `cfg.UploadPath` => stream lokal
- Jika `video_url` absolute http(s) => fetch dan stream

### Frontend
1) `frontend/lib/api.ts`
- tambah `category`, `genre` pada interface Video
- tambah `api.videos.suggest(q)`
- tambah `api.videos.download(id)`

2) Navbar autocomplete
- gunakan hook state + debounce kecil
- dropdown clickable => router.push(`/watch/${id}`)

3) Watch page tombol
- tombol Download: panggil endpoint download (pakai window.location atau fetch+blob)
- tombol Share: `navigator.share` jika tersedia, fallback copy link

### Catatan
- Pastikan perubahan tidak merusak existing search results page `/search` yang sudah ada.
- Autocomplete bisa hanya di Navbar (tidak harus mengubah SearchPage).

