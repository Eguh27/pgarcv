# 🧪 Testing Guide - Video Platform Fixes

## Quick Start

```bash
# 1. Compile backend
cd backend
go build ./cmd/server

# 2. Build frontend
cd ../frontend
npm run build

# 3. Run tests (see below)
```

---

## Test Scenarios

### SCENARIO 1: Security - Unpublished Video Protection

**Test Case 1.1: Public user tries to access unpublished video M3U8**

```bash
# Setup: Create a video with is_published=false and HLS processed

# Attempt download as public user (no auth token)
curl -i http://localhost:8080/api/hls/5/index.m3u8

# EXPECTED:
# HTTP/1.1 403 Forbidden
# Content-Type: application/json
# {"success":false,"code":403,"message":"Video tidak tersedia"}
```

**Test Case 1.2: Public user tries to access segment**

```bash
curl -i http://localhost:8080/api/hls/5/seg_000.ts

# EXPECTED: 403 Forbidden
```

**Test Case 1.3: Admin with valid token accesses unpublished video**

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r .data.token)

# Access with token
curl -i -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/hls/5/index.m3u8

# EXPECTED:
# HTTP/1.1 200 OK
# Content-Type: application/vnd.apple.mpegurl
# #EXTM3U
# ...
```

**Test Case 1.4: Published video accessible to public**

```bash
# Setup: Create video with is_published=true and HLS processed

curl -i http://localhost:8080/api/hls/1/index.m3u8

# EXPECTED: 200 OK + M3U8 content
```

---

### SCENARIO 2: Admin Experience - HLS Status Polling

**Test Case 2.1: Admin uploads video and polls HLS status**

```bash
# 1. Upload video (traditional)
curl -X POST http://localhost:8080/api/admin/upload/video \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@testvideo.mp4"

# Response:
# {"success":true,"data":{"raw_path":"/path/to/video.mp4","duration":300}}

# Store raw_path for next step

# 2. Create video record (auto-triggers HLS)
VIDEO_ID=$(curl -s -X POST http://localhost:8080/api/admin/videos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Video","video_url":"","raw_path":"/path/to/video.mp4", ...}' \
  | jq -r .data.id)

# 3. Poll HLS status
for i in {1..10}; do
  STATUS=$(curl -s http://localhost:8080/api/hls/status/$VIDEO_ID | jq .)
  echo "Attempt $i: $(echo $STATUS | jq -r .hls_status)"
  sleep 2
done

# EXPECTED sequence:
# Attempt 1: pending
# Attempt 2: processing
# Attempt 3: processing
# Attempt 4: done (or error)
```

**Test Case 2.2: Check video_url updates after HLS completion**

```bash
# Poll until done
while true; do
  STATUS=$(curl -s http://localhost:8080/api/hls/status/$VIDEO_ID)
  IS_COMPLETE=$(echo $STATUS | jq -r .is_complete)
  VIDEO_URL=$(echo $STATUS | jq -r .video_url)
  
  if [ "$IS_COMPLETE" = "true" ]; then
    echo "HLS Complete! video_url: $VIDEO_URL"
    break
  fi
  
  sleep 2
done

# EXPECTED:
# HLS Complete! video_url: /api/hls/{VIDEO_ID}/index.m3u8
```

**Test Case 2.3: Admin can play unpublished video**

```bash
# Get video
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/videos/$VIDEO_ID | jq .

# EXPECTED: video_url is populated
# {
#   "id": 5,
#   "video_url": "/api/hls/5/index.m3u8",
#   "hls_status": "done",
#   ...
# }

# Frontend loads M3U8
curl -i -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/hls/5/index.m3u8

# EXPECTED: 200 OK + M3U8 content
```

---

### SCENARIO 3: Upload Performance - Chunked Upload

**Test Case 3.1: Initiate chunked upload session**

```bash
UPLOAD_SESSION=$(curl -s -X POST http://localhost:8080/api/admin/upload/chunked/initiate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "filename": "largefile.mp4",
    "file_size": 10485760,
    "total_chunks": 3
  }')

UPLOAD_ID=$(echo $UPLOAD_SESSION | jq -r .data.upload_id)
CHUNK_SIZE=$(echo $UPLOAD_SESSION | jq -r .data.chunk_size)

echo "Upload ID: $UPLOAD_ID"
echo "Chunk Size: $CHUNK_SIZE"

# EXPECTED:
# Upload ID: abc123-def456-...
# Chunk Size: 5242880
```

**Test Case 3.2: Upload chunks in parallel**

```bash
# Create test file (10MB)
dd if=/dev/zero of=test_10mb.bin bs=1M count=10

# Split into chunks (5MB each = 2 chunks)
split -b 5242880 test_10mb.bin chunk_

# Upload chunks in parallel
for i in 0 1; do
  (
    echo "Uploading chunk $i..."
    curl -X POST http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID/chunk \
      -H "Authorization: Bearer $TOKEN" \
      -F "chunk=@chunk_$(printf '%02d' $i)" \
      -F "chunk_index=$i" | jq .
  ) &
done

wait

# EXPECTED:
# Each response shows progress_pct increasing
# {"success":true,"data":{"chunk_index":0,"chunks_received":1,"progress_pct":50}}
# {"success":true,"data":{"chunk_index":1,"chunks_received":2,"progress_pct":100}}
```

**Test Case 3.3: Check upload progress**

```bash
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID/progress | jq .

# EXPECTED:
# {
#   "success": true,
#   "data": {
#     "chunks_received": 2,
#     "total_chunks": 2,
#     "progress_pct": 100,
#     "status": "in_progress"
#   }
# }
```

**Test Case 3.4: Complete chunked upload**

```bash
FINAL_FILE=$(curl -s -X POST http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID/complete \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | jq .data)

echo $FINAL_FILE | jq .

# EXPECTED:
# {
#   "filename": "abc123-def456.mp4",
#   "path": "/path/to/uploads/raw/abc123-def456.mp4",
#   "size": 10485760
# }

# Verify file exists and size matches
ls -lh /path/to/uploads/raw/abc123-def456.mp4
```

**Test Case 3.5: Abort upload**

```bash
# Initiate another session
UPLOAD_ID2=$(curl -s -X POST http://localhost:8080/api/admin/upload/chunked/initiate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filename":"test2.mp4","file_size":5242880,"total_chunks":2}' \
  | jq -r .data.upload_id)

# Abort before completion
curl -X DELETE http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID2 \
  -H "Authorization: Bearer $TOKEN"

# EXPECTED:
# {"success":true,"data":{"message":"Upload dibatalkan"}}

# Verify chunks directory cleaned up
ls -la /path/to/uploads/.chunks/$UPLOAD_ID2/
# EXPECTED: Not found or empty
```

---

### SCENARIO 4: Full End-to-End Workflow

**Test Case 4.1: Admin uploads, processes, publishes, and public watches**

```bash
#!/bin/bash

# 1. Admin login
echo "1. Logging in..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r .data.token)

echo "   Token: $TOKEN"

# 2. Upload using chunked (10MB file)
echo "2. Starting chunked upload..."
UPLOAD_SESSION=$(curl -s -X POST http://localhost:8080/api/admin/upload/chunked/initiate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"filename":"e2e_test.mp4","file_size":10485760,"total_chunks":2}')

UPLOAD_ID=$(echo $UPLOAD_SESSION | jq -r .data.upload_id)
echo "   Upload ID: $UPLOAD_ID"

# Create and upload chunks
dd if=/dev/zero of=test_chunk0.bin bs=1M count=5
dd if=/dev/zero of=test_chunk1.bin bs=1M count=5

curl -s -X POST http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID/chunk \
  -H "Authorization: Bearer $TOKEN" \
  -F "chunk=@test_chunk0.bin" \
  -F "chunk_index=0" | jq .data.progress_pct

curl -s -X POST http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID/chunk \
  -H "Authorization: Bearer $TOKEN" \
  -F "chunk=@test_chunk1.bin" \
  -F "chunk_index=1" | jq .data.progress_pct

# 3. Complete upload
echo "3. Completing upload..."
UPLOAD_DATA=$(curl -s -X POST http://localhost:8080/api/admin/upload/chunked/$UPLOAD_ID/complete \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | jq .data)

RAW_PATH=$(echo $UPLOAD_DATA | jq -r .path)
echo "   Saved to: $RAW_PATH"

# 4. Create video (triggers auto HLS)
echo "4. Creating video (auto-triggers HLS)..."
VIDEO=$(curl -s -X POST http://localhost:8080/api/admin/videos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"title\": \"E2E Test Video\",
    \"subtitle\": \"End-to-end test\",
    \"description\": \"Testing full workflow\",
    \"thumbnail_url\": \"/uploads/placeholder.jpg\",
    \"video_url\": \"\",
    \"preview_url\": \"\",
    \"duration\": 300,
    \"is_published\": false,
    \"allow_download\": false,
    \"categories\": [],
    \"genres\": [],
    \"raw_path\": \"$RAW_PATH\"
  }" | jq .data)

VIDEO_ID=$(echo $VIDEO | jq -r .id)
echo "   Video ID: $VIDEO_ID"

# 5. Wait for HLS processing
echo "5. Waiting for HLS processing..."
for i in {1..30}; do
  STATUS=$(curl -s http://localhost:8080/api/hls/status/$VIDEO_ID | jq .data)
  HLS_STATUS=$(echo $STATUS | jq -r .hls_status)
  IS_COMPLETE=$(echo $STATUS | jq -r .is_complete)
  
  echo "   Attempt $i: $HLS_STATUS"
  
  if [ "$IS_COMPLETE" = "true" ]; then
    echo "   ✅ HLS Complete!"
    break
  fi
  
  sleep 2
done

# 6. Publish video
echo "6. Publishing video..."
curl -s -X PUT http://localhost:8080/api/admin/videos/$VIDEO_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"is_published": true}' | jq .data.is_published

# 7. Public user views video
echo "7. Public viewing video..."
curl -s http://localhost:8080/api/videos/$VIDEO_ID | jq .

# 8. Public accesses HLS
echo "8. Public accessing HLS..."
curl -i http://localhost:8080/api/hls/$VIDEO_ID/index.m3u8

echo ""
echo "✅ E2E Test Complete!"
```

**Expected Output**:
```
1. Logging in...
   Token: eyJ...
2. Starting chunked upload...
   Upload ID: abc123...
   Uploaded chunk 0: 50%
   Uploaded chunk 1: 100%
3. Completing upload...
   Saved to: /path/to/uploads/raw/abc123.mp4
4. Creating video (auto-triggers HLS)...
   Video ID: 5
5. Waiting for HLS processing...
   Attempt 1: pending
   Attempt 2: processing
   Attempt 3: processing
   Attempt 4: done
   ✅ HLS Complete!
6. Publishing video...
   true
7. Public viewing video...
   {id: 5, title: "E2E Test Video", ...}
8. Public accessing HLS...
   HTTP/1.1 200 OK
   #EXTM3U
   ...
✅ E2E Test Complete!
```

---

## 🐛 Troubleshooting

### Problem: "Upload session tidak ditemukan"
**Cause**: Upload session expired or invalid upload_id
**Solution**: Create new session with `initiate` endpoint

### Problem: "Missing chunks: X/Y"
**Cause**: Not all chunks uploaded before `complete`
**Solution**: Upload all chunks, check `progress` endpoint

### Problem: "HLS processing gagal"
**Cause**: ffmpeg error, invalid video file, or permissions
**Solution**: Check server logs, verify file is valid MP4

### Problem: "Video tidak tersedia" (403)
**Cause**: Video not published and no auth token
**Solution**: 
- Publish video, OR
- Add `Authorization: Bearer $TOKEN` header

### Problem: Chunked upload very slow
**Cause**: Uploading one at a time instead of parallel
**Solution**: Use background requests to upload multiple chunks simultaneously

---

## 📊 Performance Benchmarks

### Expected Upload Performance:
- Single 10MB file: ~5-10 seconds
- Single 100MB file: ~1-2 minutes (traditional)
- Single 100MB file: ~30-45 seconds (chunked, 2-3 parallel)
- 500MB file: ~2-3 minutes (chunked with good network)

### HLS Processing:
- Small video (< 100MB): ~30-60 seconds
- Medium video (100-300MB): ~2-5 minutes
- Large video (300-500MB): ~5-10 minutes

---

## ✅ Checklist Before Going Live

- [ ] Backend compiles without errors
- [ ] Frontend builds successfully
- [ ] Test Case 1.1 passes (security works)
- [ ] Test Case 1.4 passes (published accessible)
- [ ] Test Case 2.1 passes (HLS status polling)
- [ ] Test Case 3.4 passes (chunked upload)
- [ ] Test Case 4.1 passes (E2E workflow)
- [ ] Database migrations applied
- [ ] CORS settings correct
- [ ] File permissions on upload directory OK
- [ ] ffmpeg installed on server
- [ ] Disk space sufficient for uploads

---

**Last Updated**: 2026-05-11
**Status**: All tests documented and ready
