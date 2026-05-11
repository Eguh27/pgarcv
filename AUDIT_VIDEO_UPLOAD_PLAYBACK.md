# 🔍 AUDIT: Video Upload & Playback Issues

## Executive Summary
Tiga masalah kritis ditemukan:
1. **Upload Lambat/Berat** - No chunking, file upload ~500MB dalam 1 request
2. **Video Tidak Bisa Diplay** - HLS endpoints tidak check authorization
3. **Admin Juga Tidak Bisa Play** - Missing video.is_published checks

---

## 🔴 ISSUE #1: Upload Lambat (Berat)

### Root Cause
- **No Chunked Upload**: File dikirim dalam 1 request besar
- **No Resume**: Kalau gagal, harus ulang dari awal
- **Simulated Progress**: Progress bar hanya animasi, bukan real progress
- **Single Stream**: Tidak ada parallel/concurrent upload

### Affected Code
- Frontend: [frontend/components/admin/VideoForm.tsx](frontend/components/admin/VideoForm.tsx#L53-L75)
  ```typescript
  res = await adminApi.uploadVideo(file);  // Semuanya dalam 1 request
  ```
- Backend: [backend/internal/handler/upload_video_handler.go](backend/internal/handler/upload_video_handler.go#L28-L50)
  ```go
  if err := c.SaveUploadedFile(file, rawPath); err != nil {  // Semuanya langsung tersimpan
  ```

### Impact
- Timeout pada koneksi lambat/unstable
- Upload 100MB+ bisa memakan waktu 5-10 menit+
- User tidak tahu progress sebenarnya
- Network error = retry dari 0

### Solution
Implementasi chunked upload:
1. Backend: Buat endpoint untuk receive chunks + assembly
2. Frontend: Pecah file jadi 5-10MB chunks, upload parallel (2-3 chunks)
3. Add resume logic: Store chunk metadata, skip uploaded chunks
4. Real progress tracking: Report actual received bytes

---

## 🔴 ISSUE #2: Video Tidak Bisa Diplay (Public)

### Root Cause
**HLS endpoints tidak check `is_published` flag!**

#### Comparison:
- ✅ **ServeKey** (OK) - Checks `if !video.IsPublished`:
  ```go
  if !video.IsPublished {
    response.Error(c, 403, "Video tidak tersedia")
    return
  }
  ```

- ❌ **ServeM3U8** (BUG) - NO CHECK:
  ```go
  func (h *HLSHandler) ServeM3U8(c *gin.Context) {
    videoID := c.Param("id")
    m3u8Path := filepath.Join(h.outputBase, "hls",
      fmt.Sprintf("video_%s", videoID), "index.m3u8")
    // ❌ Tidak check is_published!
    // ❌ Tidak check authorization!
    c.File(m3u8Path)
  }
  ```

- ❌ **ServeSegment** (BUG) - NO CHECK:
  ```go
  func (h *HLSHandler) ServeSegment(c *gin.Context) {
    videoID := c.Param("id")
    segment := c.Param("segment")
    // ❌ Tidak check is_published!
    c.File(segPath)
  }
  ```

### Impact
- **Public users** dapat download M3U8 file langsung dari `/api/hls/1/index.m3u8`
- Dapat download semua `.ts` segments dari `/api/hls/1/seg_000.ts`, dll
- Unpublished videos bisa diakses & diunduh oleh siapa saja
- Admin/authenticated users juga tidak bisa akses unpublished videos

### Affected Files
- [backend/internal/handler/hls_handler.go](backend/internal/handler/hls_handler.go)

### Solution
Update `ServeM3U8` dan `ServeSegment` untuk:
1. Query video dari database
2. Check `is_published` untuk public users
3. Allow unpublished videos untuk authenticated admin
4. Return 403 jika tidak authorized

---

## 🔴 ISSUE #3: Admin Juga Tidak Bisa Main Video

### Root Cause
Kombinasi dari Issue #2 + workflow issue:

1. **No auth check di HLS endpoints** - HLS handler tidak bisa bedain public vs admin
2. **Video URL mungkin kosong** - HLS processing adalah async, video_url belum di-set saat UI render
3. **Manual two-step process**:
   - Upload video → dapat `raw_path`
   - Klik "Process HLS" manual → HLS di-process
   - Kalau tidak di-process, video_url kosong
   - Player coba load video dari URL kosong = fail

### Timeline Flow
```
1. Admin upload video (→ raw video tersimpan)
2. Response: raw_path = /path/to/video_uuid.mp4
3. Admin create video record dengan video_url masih kosong
4. Admin klik "Process HLS" button (manual!)
5. Backend: Async goroutine process HLS (background)
6. Goroutine: Update video_url = /api/hls/1/index.m3u8
7. Admin bisa play hanya kalau tunggu goroutine selesai
8. Kalau admin navigate away sebelum goroutine selesai = video_url masih kosong = can't play
```

### Affected Code
- Frontend: [frontend/app/admin/(authenticated)/videos/new/page.tsx](frontend/app/admin/(authenticated)/videos/new/page.tsx#L22)
  ```typescript
  if (data.rawPath) {
    try {
      await adminApi.processHLS(savedVideo.id, data.rawPath);  // Manual call
    }
  }
  ```
- Backend: [backend/internal/handler/upload_video_handler.go](backend/internal/handler/upload_video_handler.go#L89-L120)
  ```go
  go func(rawPath string, videoID uint) {  // Async goroutine, no status tracking
    // ... HLS processing happens here
  }(req.RawPath, req.VideoID)
  ```

### Impact
- Admin upload, create video, tapi video masih tidak bisa diplay
- No feedback when HLS processing complete
- Users might delete/move away before HLS ready
- No way to check HLS status

---

## 📋 Summary of Issues

| Issue | Severity | Impact | Affected Users |
|-------|----------|--------|-----------------|
| **No chunked upload** | High | Slow uploads, timeout on large files | Admin |
| **HLS endpoints missing auth** | Critical | Unpublished videos accessible to public | Everyone |
| **Admin can't play videos** | High | Can't verify videos after upload | Admin |
| **No HLS status tracking** | Medium | Don't know when video is ready | Admin |

---

## ✅ Fix Plan

### Phase 1: Security (URGENT)
- Add `is_published` checks ke `ServeM3U8` & `ServeSegment`
- Add auth middleware untuk allow admin to access unpublished videos
- Prevent unauthorized HLS access

### Phase 2: Upload Improvement
- Implement chunked upload (5-10MB chunks)
- Add resume capability
- Real progress tracking
- Backend: chunk assembly, verification

### Phase 3: UX/Admin Experience
- Make HLS processing automatic after upload
- Add status polling to check HLS completion
- Show loading state while HLS processing
- Automatic redirect to video list when ready

---

## 🔧 Recommended Fix Priority

1. **CRITICAL** (Do First): Fix HLS authorization - ISSUE #2
2. **HIGH** (Do Next): Fix admin video playback - ISSUE #3  
3. **HIGH** (Do Next): Implement chunked upload - ISSUE #1
