# 🔬 HLS Video Delivery - Authorization Flow Analysis

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Request                              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                    ┌────────▼────────┐
                    │  Route Matching │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
   [PUBLIC]            [ADMIN]            [HLS DELIVERY]
   /api/*              /api/admin/*       /api/hls/*
   NO MIDDLEWARE       AuthRequired       NO MIDDLEWARE ❌
                       ✅                 
                                         
```

---

## 1. Request Flow Analysis

### Route Registration (backend/cmd/server/main.go, lines 80-92)

```go
api := r.Group("/api")
{
  // Videos (PUBLIC)
  api.GET("/videos", videoH.List)                    // ✅ has is_published check
  api.GET("/videos/:id", videoH.Get)                 // ✅ has is_published check
  
  // HLS (PUBLIC - NO MIDDLEWARE)
  api.GET("/hls/key/:id", hlsH.ServeKey)             // ✅ checks is_published
  api.GET("/hls/:id/index.m3u8", hlsH.ServeM3U8)    // ❌ NO CHECKS
  api.GET("/hls/:id/:segment", hlsH.ServeSegment)   // ❌ NO CHECKS
}

admin := r.Group("/api/admin")
admin.Use(middleware.AuthRequired(cfg.JWTSecret))    // Only ADMIN has middleware
{
  // Admin operations...
}
```

**Key Observation**: HLS endpoints are in PUBLIC group, NOT in admin group!

---

## 2. Authorization Flow - Method Comparison

### ✅ CORRECT: `ServeKey()` - backend/internal/handler/hls_handler.go

```go
func (h *HLSHandler) ServeKey(c *gin.Context) {
  // 1. Parse video ID
  videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
  if err != nil { return }
  
  // 2. QUERY database to get video object
  var video repository.Video
  if err := h.db.First(&video, videoID).Error; err != nil {
    response.Error(c, 404, "Video tidak ditemukan")
    return
  }
  
  // 3. ✅ CHECK is_published FLAG
  if !video.IsPublished {
    response.Error(c, 403, "Video tidak tersedia")  // Forbid access
    return
  }
  
  // 4. Serve encryption key (only for published videos)
  keyBytes, err := hex.DecodeString(h.encryptKey)
  c.Data(http.StatusOK, "application/octet-stream", keyBytes)
}
```

**Flow**:
- Authenticated? NO (public endpoint)
- Published? YES (required)
- Result: ✅ Public can access published videos' encryption keys

---

### ❌ BUGGY: `ServeM3U8()` - backend/internal/handler/hls_handler.go

```go
func (h *HLSHandler) ServeM3U8(c *gin.Context) {
  videoID := c.Param("id")
  m3u8Path := filepath.Join(h.outputBase, "hls",
    fmt.Sprintf("video_%s", videoID), "index.m3u8")
  
  // 1. Check if FILE exists (filesystem check only)
  if _, err := os.Stat(m3u8Path); os.IsNotExist(err) {
    response.Error(c, 404, "HLS tidak tersedia")
    return
  }
  
  // 2. ❌ NO DATABASE CHECK
  // 3. ❌ NO is_published CHECK
  // 4. ❌ NO AUTHENTICATION CHECK
  
  // 5. Serve file directly from disk
  c.File(m3u8Path)
}
```

**Flow**:
- Authenticated? NO way to check (no middleware)
- Published? NO check
- Result: ❌ ANYONE can access ANY video's M3U8 file if video ID is known

**Problem**: Uses only filesystem check (file exists?), bypasses database/business logic entirely!

---

### ❌ BUGGY: `ServeSegment()` - backend/internal/handler/hls_handler.go

```go
func (h *HLSHandler) ServeSegment(c *gin.Context) {
  videoID := c.Param("id")
  segment := c.Param("segment")
  
  // 1. Path traversal check (security against directory escape)
  if filepath.Base(segment) != segment {
    response.Error(c, 400, "Invalid segment")
    return
  }
  
  // 2. Construct segment path
  segPath := filepath.Join(h.outputBase, "hls",
    fmt.Sprintf("video_%s", videoID), segment)
  
  // 3. Check if FILE exists (filesystem check only)
  if _, err := os.Stat(segPath); os.IsNotExist(err) {
    response.Error(c, 404, "Segment tidak ditemukan")
    return
  }
  
  // 4. ❌ NO DATABASE CHECK
  // 5. ❌ NO is_published CHECK
  // 6. ❌ NO AUTHENTICATION CHECK
  
  // 7. Serve TS segment directly from disk
  c.File(segPath)
}
```

**Flow**:
- Authenticated? NO way to check
- Published? NO check
- Result: ❌ ANYONE can download `.ts` segments

---

## 3. Middleware Analysis

### AuthRequired Middleware - backend/internal/middleware/auth.go

```go
func AuthRequired(secret string) gin.HandlerFunc {
  return func(c *gin.Context) {
    // 1. Try cookie first
    tokenStr, err := c.Cookie("admin_token")
    if err != nil {
      // 2. Fall back to Authorization header
      auth := c.GetHeader("Authorization")
      if !strings.HasPrefix(auth, "Bearer ") {
        response.Error(c, 401, "Unauthorized")
        c.Abort()  // Stop request
        return
      }
      tokenStr = strings.TrimPrefix(auth, "Bearer ")
    }
    
    // 3. Validate JWT token
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
      return []byte(secret), nil
    }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
    
    if err != nil || !token.Valid {
      response.Error(c, 401, "Token tidak valid")
      c.Abort()
      return
    }
    
    // 4. ❌ NO CONTEXT SET - doesn't preserve auth info downstream!
    c.Next()
  }
}
```

**Critical Issue**: Middleware validates token but **does NOT set any context**:
- No `c.Set("user", ...)` 
- No `c.Set("authenticated", true)`
- Downstream handlers have no way to know who is making request!

---

## 4. Database Access Patterns

### VideoHandler (Video Metadata) - backend/internal/handler/video_handler.go

#### Public Access:
```go
func (h *VideoHandler) Get(c *gin.Context) {
  var video repository.Video
  if err := h.db.Preload("Categories").Preload("Genres")
    .Where("is_published = ?", true)  // ✅ FILTER for published
    .First(&video, c.Param("id"))
    .Error; err != nil {
    response.Error(c, 404, "Video tidak ditemukan")
    return
  }
  response.OK(c, video)
}
```
✅ Filters to published videos only

#### Admin Access:
```go
func (h *VideoHandler) AdminGet(c *gin.Context) {
  var video repository.Video
  if err := h.db.Preload("Categories").Preload("Genres")
    .First(&video, c.Param("id"))  // ✅ NO FILTER (gets all videos)
    .Error; err != nil {
    response.Error(c, 404, "Video tidak ditemukan")
    return
  }
  response.OK(c, video)
}
```
✅ Gets all videos (admin should see unpublished)

### HLSHandler (HLS Delivery):
```go
// ServeKey queries DB:
var video repository.Video
if err := h.db.First(&video, videoID).Error; err != nil { ... }

// ServeM3U8 & ServeSegment: ❌ NO DB QUERY AT ALL - just checks filesystem
```

---

## 5. Why Unpublished Videos Are Publicly Accessible

### Attack Scenario:

```
Attacker knows video_id = 3 (unpublished test video)

1. Request: GET /api/videos/3
   Response: 404 "Video tidak ditemukan" ✅
   (blocked by is_published check in VideoHandler.Get)

2. Request: GET /api/hls/3/index.m3u8
   Response: 200 + M3U8 content ❌
   (NO publication check in ServeM3U8)
   
3. From M3U8, attacker sees:
   #EXTM3U
   #EXT-X-VERSION:3
   #EXT-X-TARGETDURATION:10
   #EXTINF:10.0,
   seg_000.ts
   
4. Request: GET /api/hls/3/seg_000.ts
   Response: 200 + video segment ❌
   (NO publication check in ServeSegment)
   
5. Attacker can now:
   - Download entire M3U8 manifest
   - Download all .ts segments
   - Reconstruct full video
   - Share/distribute unpublished content
```

**Root Cause**: Different code paths = inconsistent security:
- `VideoHandler` filters by `is_published` ✅
- `HLSHandler.ServeKey` filters by `is_published` ✅
- `HLSHandler.ServeM3U8` does NOT filter ❌
- `HLSHandler.ServeSegment` does NOT filter ❌

---

## 6. Why Admin Also Can't Play Unpublished Videos

### Admin Workflow:

```
1. Admin uploads video → creates Video record with is_published=false
2. Admin wants to preview → requests GET /api/hls/1/index.m3u8
3. ServeM3U8 is PUBLIC endpoint (no auth middleware applied)
4. Even if admin has valid JWT token, endpoint doesn't check it
5. Endpoint only checks if FILE EXISTS (filesystem)
6. File exists (HLS was processed) → returns M3U8 ✅

Wait, that should work... Let me check the actual issue...
```

Actually, let me reconsider. The admin CAN access M3U8 file (no auth needed = anyone can access). 

**The real admin issue from audit is**: Video shows `video_url: ""` (empty) initially because HLS processing is async. So even if M3U8 is accessible, the player component receives empty URL.

**Two separate issues:**
1. **Published videos**: M3U8/segments are accessible to public (security bug)
2. **Admin unpublished video preview**: Works IF video_url is populated, but may fail if accessing during HLS processing (workflow bug)

---

## 7. Context & State Propagation

### Current State (Broken):

```
Request → AuthMiddleware → Handler
           ✓ Validates JWT   ? Sets context?
           ✓ Allows through  ✗ No - c.Set() not called
                             ✗ Downstream can't know user status
```

### What's Missing:

```go
// In AuthRequired middleware after token validation:
if claims, ok := token.Claims.(jwt.MapClaims); ok {
  c.Set("authenticated", true)
  c.Set("username", claims["username"])
  c.Set("exp", claims["exp"])
}

// Then in HLS handler:
if isAuth, exists := c.Get("authenticated"); exists && isAuth.(bool) {
  // Allow access to unpublished videos
  // Allow admin preview
} else {
  // Only allow published videos
}
```

---

## 8. Exact Changes Needed

### Change 1: Fix `ServeM3U8` - Add Authorization

**Current**:
```go
func (h *HLSHandler) ServeM3U8(c *gin.Context) {
  videoID := c.Param("id")
  m3u8Path := filepath.Join(h.outputBase, "hls",
    fmt.Sprintf("video_%s", videoID), "index.m3u8")
  if _, err := os.Stat(m3u8Path); os.IsNotExist(err) {
    response.Error(c, 404, "HLS tidak tersedia")
    return
  }
  c.File(m3u8Path)
}
```

**Needs**:
1. Parse `videoID` as uint (like ServeKey does)
2. Query database: `db.First(&video, videoID)`
3. Check if video exists
4. Check if `is_published` is true (for public users)
5. OR check if user is authenticated (for admin preview)
6. Only then serve M3U8

---

### Change 2: Fix `ServeSegment` - Add Authorization

**Same as ServeM3U8** - Add DB query + publication/auth check before serving segment

---

### Change 3: Enhance AuthRequired Middleware

**Add context propagation**:
```go
// After token validation succeeds
if claims, ok := token.Claims.(jwt.MapClaims); ok {
  c.Set("authenticated", true)
  if username, ok := claims["username"].(string); ok {
    c.Set("username", username)
  }
}
```

---

### Change 4: Create Helper Method in HLSHandler

**Add authorization check method**:
```go
func (h *HLSHandler) canAccessVideo(c *gin.Context, videoID uint) (bool, error) {
  var video repository.Video
  if err := h.db.First(&video, videoID).Error; err != nil {
    return false, err
  }
  
  // Published videos: allow everyone
  if video.IsPublished {
    return true, nil
  }
  
  // Unpublished videos: only allow authenticated users (admin)
  if auth, exists := c.Get("authenticated"); exists && auth.(bool) {
    return true, nil
  }
  
  return false, nil
}
```

---

## Summary Table

| Aspect | Current | Issue | Solution |
|--------|---------|-------|----------|
| **ServeM3U8** | Filesystem check only | No DB validation, no is_published check | Query DB + check is_published + check auth |
| **ServeSegment** | Filesystem check only | No DB validation, no is_published check | Query DB + check is_published + check auth |
| **ServeKey** | DB query + is_published check | ✅ Correct (but inconsistent) | Keep as reference for other methods |
| **AuthMiddleware** | Validates JWT | Doesn't set context for downstream handlers | Add `c.Set("authenticated", true)` after validation |
| **Route Protection** | HLS endpoints are PUBLIC | Should check per-method | Create helper method to centralize logic |

---

## Request Flow Diagram (Current vs Fixed)

### CURRENT (Broken):
```
GET /api/hls/1/index.m3u8
    ↓
[No Middleware]
    ↓
ServeM3U8()
    ├─ os.Stat(m3u8Path) → exists
    ├─ ❌ No DB check
    ├─ ❌ No is_published check  
    └─ Serve file ✗ SECURITY HOLE

GET /api/hls/1/seg_000.ts
    ↓
[No Middleware]
    ↓
ServeSegment()
    ├─ os.Stat(segPath) → exists
    ├─ ❌ No DB check
    ├─ ❌ No is_published check
    └─ Serve file ✗ SECURITY HOLE
```

### FIXED:
```
GET /api/hls/1/index.m3u8
    ↓
[Optional: Check Bearer token if provided]
    ↓
ServeM3U8()
    ├─ DB query: Find video (ID=1)
    ├─ Check is_published
    │  ├─ If true → Serve ✓
    │  ├─ If false AND user authenticated → Serve ✓
    │  └─ If false AND NOT authenticated → 403 Forbidden ✓
    └─ [Middleware should have set auth context]
```

