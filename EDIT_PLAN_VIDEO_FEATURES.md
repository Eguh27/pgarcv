# Edit Plan - Tambah kategori, genre, download/share, autocomplete, perbaikan mobile card, dan thumbnail otomatis

## Tahap A (Backend)
1) `backend/internal/repository/models.go`
   - tambah field pada `Video`:
     - `Category string`
     - `Genre string`
2) `backend/internal/handler/video_handler.go`
   - extend `videoRequest` dengan `Category`, `Genre`
   - extend `applyVideoRequest`
   - update search `List()` agar filter pakai `title/subtitle/category/genre`
3) `backend/cmd/server/main.go`
   - tambah route publik:
     - `GET /api/videos/suggest?q=...` (return top matches)
     - `GET /api/videos/:id/download`

## Tahap B (Frontend)
4) `frontend/lib/api.ts`
   - extend interface `Video` dengan `category`, `genre`
   - tambah fungsi request untuk suggest & download
5) Admin form
   - `frontend/components/admin/VideoForm.tsx`
     - tambah input `category` & `genre`
6) Autocomplete search
   - `frontend/components/layout/Navbar.tsx`
     - dropdown suggestions berdasarkan endpoint `/api/videos/suggest`
7) Watch page actions
   - `frontend/app/(public)/watch/[id]/page.tsx`
     - tombol Download
     - tombol Share (navigator.share + fallback copy)
8) Perbaikan mobile card
   - `frontend/components/video/VideoCard.tsx`
   - `frontend/components/video/VideoGridClient.tsx`
9) Thumbnail otomatis dari frame video (jika admin kosong)
   - `frontend/components/admin/VideoForm.tsx`
     - capture frame via canvas (browser-side) lalu upload image

## Follow-up testing
- Jalankan backend: `go run ./cmd/server`
- Jalankan frontend dev: `npm run dev`
- Cek:
  - admin create/edit video memunculkan category & genre
  - search autocomplete bekerja
  - watch page ada download & share
  - mobile card ukuran match

