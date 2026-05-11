# 🚀 PANDUAN DEPLOY — Video Platform
> Frontend: Vercel | Backend: Render | Database: Neon (PostgreSQL) | Storage: Cloudinary
> Estimasi waktu: 2-3 jam pertama kali

---

## 📋 OVERVIEW ARSITEKTUR DEPLOY

```
GitHub (source code)
    ↓ auto deploy
    ├── Vercel          → frontend Next.js (namamu.vercel.app)
    ├── Render          → backend Go (namamu.onrender.com)  
    ├── Neon            → PostgreSQL database (gratis)
    └── Cloudinary      → file uploads foto & video (gratis 25GB)
```

**Soal domain:**
- Gratis: `namamu.vercel.app` + `namamu.onrender.com`
- Punya domain sendiri (misal beli di Niagahoster ~Rp50rb/tahun):
  attach ke Vercel → URL jadi `namasite.com` tanpa "vercel" sama sekali ✅

---

## 🗄️ STEP 1 — Migrasi Database: SQLite → PostgreSQL (Neon)

### 1.1 Daftar Neon
1. Buka https://neon.tech → Sign up gratis pakai GitHub
2. Create project → pilih region `Singapore` (terdekat)
3. Catat **Connection String**-nya:
   ```
   postgresql://user:password@ep-xxx.ap-southeast-1.aws.neon.tech/dbname?sslmode=require
   ```

### 1.2 Install driver PostgreSQL di backend Go

```bash
cd backend
go get gorm.io/driver/postgres
```

### 1.3 Edit `backend/config/config.go` — tambah DATABASE_URL

```go
// backend/config/config.go
package config

import (
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    Port          string
    JWTSecret     string
    DBPath        string        // untuk SQLite (dev)
    DatabaseURL   string        // untuk PostgreSQL (production)
    UploadPath    string
    AdminUsername string
    AdminPassword string
    CloudinaryURL string
    Env           string
}

func Load() *Config {
    godotenv.Load()
    return &Config{
        Port:          getEnv("PORT", "8080"),
        JWTSecret:     getEnv("JWT_SECRET", "secret-dev-only"),
        DBPath:        getEnv("DB_PATH", "./videoplatform.db"),
        DatabaseURL:   getEnv("DATABASE_URL", ""),
        UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
        AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
        AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
        CloudinaryURL: getEnv("CLOUDINARY_URL", ""),
        Env:           getEnv("ENV", "development"),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

### 1.4 Edit `backend/internal/repository/database.go` — support dua driver

```go
// backend/internal/repository/database.go
package repository

import (
    "log"
    "os"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dbPath, databaseURL string) *gorm.DB {
    var db *gorm.DB
    var err error

    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Warn),
    }

    if databaseURL != "" {
        // Production: PostgreSQL
        log.Println("🐘 Menggunakan PostgreSQL...")
        db, err = gorm.Open(postgres.Open(databaseURL), gormConfig)
    } else {
        // Development: SQLite
        log.Println("🗄️ Menggunakan SQLite...")
        db, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
    }

    if err != nil {
        log.Fatal("Gagal koneksi database:", err)
    }

    // Auto migrate
    db.AutoMigrate(&Video{}, &Banner{}, &Ad{})

    // Seed hanya jika database kosong
    seedData(db)

    DB = db
    log.Println("✅ Database siap")
    return db
}

func seedData(db *gorm.DB) {
    var count int64
    db.Model(&Video{}).Count(&count)
    if count > 0 {
        return
    }

    videos := []Video{
        {
            Title: "Video Pertama", Subtitle: "Subtitle video ini",
            Description: "Deskripsi lengkap video pertama.",
            ThumbnailURL: "", VideoURL: "", Duration: 300, IsPublished: true,
        },
    }
    db.Create(&videos)
    db.Create(&Banner{Title: "Banner Utama", ImageURL: "", LinkURL: "#", IsActive: true, SortOrder: 1})
    db.Create(&Ad{Name: "Iklan Utama", AdCode: `<div style="padding:20px;text-align:center;background:#f0f0f0">Slot Iklan</div>`, DeviceType: "all", IsActive: true})
    log.Println("✅ Seed data berhasil")
}
```

### 1.5 Update `backend/cmd/server/main.go` — pass DATABASE_URL

```go
// Cari baris InitDB dan ubah jadi:
db := repository.InitDB(cfg.DBPath, cfg.DatabaseURL)
```

---

## ☁️ STEP 2 — Setup File Storage: Cloudinary

Video dan foto tidak bisa disimpan di Render (filesystem tidak persistent).
Cloudinary gratis 25GB — cukup untuk ratusan video.

### 2.1 Daftar Cloudinary
1. Buka https://cloudinary.com → Sign up gratis
2. Catat: **Cloud Name**, **API Key**, **API Secret**
3. Format CLOUDINARY_URL: `cloudinary://API_KEY:API_SECRET@CLOUD_NAME`

### 2.2 Install Cloudinary SDK di Go

```bash
cd backend
go get github.com/cloudinary/cloudinary-go/v2
```

### 2.3 Buat `backend/pkg/storage/cloudinary.go`

```go
// backend/pkg/storage/cloudinary.go
package storage

import (
    "context"
    "fmt"
    "mime/multipart"
    "path/filepath"
    "strings"

    "github.com/cloudinary/cloudinary-go/v2"
    "github.com/cloudinary/cloudinary-go/v2/api/uploader"
    "github.com/google/uuid"
)

type CloudinaryStorage struct {
    cld *cloudinary.Cloudinary
}

func NewCloudinaryStorage(cloudinaryURL string) (*CloudinaryStorage, error) {
    cld, err := cloudinary.NewFromURL(cloudinaryURL)
    if err != nil {
        return nil, fmt.Errorf("gagal init cloudinary: %w", err)
    }
    return &CloudinaryStorage{cld: cld}, nil
}

func (s *CloudinaryStorage) Upload(file multipart.File, filename string) (string, error) {
    ext := strings.ToLower(filepath.Ext(filename))
    publicID := uuid.New().String()

    // Tentukan folder berdasarkan tipe file
    folder := "thumbnails"
    resourceType := "image"
    if ext == ".mp4" || ext == ".webm" || ext == ".mov" {
        folder = "videos"
        resourceType = "video"
    }

    ctx := context.Background()
    result, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
        PublicID:     folder + "/" + publicID,
        ResourceType: resourceType,
    })
    if err != nil {
        return "", fmt.Errorf("gagal upload ke cloudinary: %w", err)
    }

    return result.SecureURL, nil
}
```

### 2.4 Update `backend/internal/handler/upload_handler.go`

```go
// backend/internal/handler/upload_handler.go
package handler

import (
    "path/filepath"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "videoplatform/pkg/response"
    "videoplatform/pkg/storage"
)

type UploadHandler struct {
    uploadPath string
    cloudinary *storage.CloudinaryStorage
}

func NewUploadHandler(uploadPath string, cld *storage.CloudinaryStorage) *UploadHandler {
    return &UploadHandler{uploadPath: uploadPath, cloudinary: cld}
}

func (h *UploadHandler) Upload(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        response.Error(c, 400, "File tidak ditemukan")
        return
    }

    ext := strings.ToLower(filepath.Ext(file.Filename))
    allowed := map[string]bool{
        ".jpg": true, ".jpeg": true, ".png": true,
        ".gif": true, ".mp4": true, ".webm": true, ".mov": true,
    }
    if !allowed[ext] {
        response.Error(c, 400, "Format file tidak didukung")
        return
    }

    if file.Size > 500*1024*1024 {
        response.Error(c, 400, "File terlalu besar (maks 500MB)")
        return
    }

    // Kalau ada Cloudinary, upload ke sana
    if h.cloudinary != nil {
        f, err := file.Open()
        if err != nil {
            response.Error(c, 500, "Gagal membuka file")
            return
        }
        defer f.Close()

        url, err := h.cloudinary.Upload(f, file.Filename)
        if err != nil {
            response.Error(c, 500, "Gagal upload: "+err.Error())
            return
        }
        response.OK(c, gin.H{"url": url, "filename": file.Filename})
        return
    }

    // Fallback: simpan lokal
    filename := uuid.New().String() + ext
    dst := filepath.Join(h.uploadPath, filename)
    if err := c.SaveUploadedFile(file, dst); err != nil {
        response.Error(c, 500, "Gagal menyimpan file")
        return
    }
    response.OK(c, gin.H{"url": "/uploads/" + filename, "filename": filename})
}
```

### 2.5 Update `backend/cmd/server/main.go` — init Cloudinary

```go
// Tambahkan setelah cfg := config.Load()

import "videoplatform/pkg/storage"

// Init Cloudinary (opsional, hanya kalau CLOUDINARY_URL ada)
var cldStorage *storage.CloudinaryStorage
if cfg.CloudinaryURL != "" {
    var err error
    cldStorage, err = storage.NewCloudinaryStorage(cfg.CloudinaryURL)
    if err != nil {
        log.Printf("⚠️ Cloudinary tidak aktif: %v", err)
    } else {
        log.Println("✅ Cloudinary siap")
    }
}

// Update init upload handler
uploadH := handler.NewUploadHandler(cfg.UploadPath, cldStorage)
```

---

## 📤 STEP 3 — Push ke GitHub

```bash
# Di root proyek
git add .
git commit -m "feat: ready for deployment"
git push origin main
```

Kalau belum ada remote:
```bash
git remote add origin https://github.com/USERNAME/REPO_NAME.git
git branch -M main
git push -u origin main
```

---

## 🖥️ STEP 4 — Deploy Backend ke Render

### 4.1 Setup Render
1. Buka https://render.com → Sign up pakai GitHub
2. New → **Web Service**
3. Connect repository GitHub kamu
4. Konfigurasi:

```
Name:         videoplatform-backend
Region:       Singapore
Branch:       main
Root Dir:     backend
Runtime:      Go
Build Command: go build -o server ./cmd/server
Start Command: ./server
```

### 4.2 Environment Variables di Render
Tambahkan satu per satu di Settings → Environment:

```
PORT              = 8080
JWT_SECRET        = [random string panjang, minimal 32 karakter]
DATABASE_URL      = [connection string dari Neon]
CLOUDINARY_URL    = cloudinary://API_KEY:API_SECRET@CLOUD_NAME
ADMIN_USERNAME    = admin
ADMIN_PASSWORD    = [password kuat]
ENV               = production
```

### 4.3 Catat URL backend Render
Setelah deploy selesai, catat URL-nya:
```
https://videoplatform-backend.onrender.com
```

> ⚠️ **Cold Start:** Render free tier "tidur" setelah 15 menit tidak ada request.
> Request pertama butuh ~30 detik. Untuk mengurangi ini, bisa pakai 
> https://uptimerobot.com (gratis) untuk ping backend setiap 5 menit.

---

## 🌐 STEP 5 — Deploy Frontend ke Vercel

### 5.1 Setup Vercel
1. Buka https://vercel.com → Sign up pakai GitHub
2. New Project → Import repository
3. Konfigurasi:

```
Framework Preset:  Next.js          (auto-detect)
Root Directory:    frontend
Build Command:     npm run build    (auto)
Output Directory:  .next            (auto)
```

### 5.2 Environment Variables di Vercel
Settings → Environment Variables:

```
NEXT_PUBLIC_API_URL = https://videoplatform-backend.onrender.com
```

### 5.3 Deploy
Klik Deploy → tunggu 2-3 menit → dapat URL:
```
https://videoplatform-xxx.vercel.app
```

---

## 🔧 STEP 6 — Update CORS Backend untuk Vercel

Edit `backend/cmd/server/main.go` — update CORS:

```go
r.Use(cors.New(cors.Config{
    AllowOriginFunc: func(origin string) bool {
        allowed := []string{
            "http://localhost:3000",
            "https://videoplatform-xxx.vercel.app",  // URL Vercel kamu
            // Kalau sudah punya domain sendiri, tambahkan di sini
            // "https://namasite.com",
        }
        for _, a := range allowed {
            if origin == a {
                return true
            }
        }
        return strings.HasSuffix(origin, ".vercel.app")
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}))
```

Commit dan push → Render auto deploy.

---

## 🌍 STEP 7 — Custom Domain (Opsional, kalau sudah beli domain)

### Di Vercel:
1. Settings → Domains → Add Domain
2. Masukkan domain kamu: `namasite.com`
3. Vercel kasih DNS record → setting di registrar domain kamu
4. Tunggu propagasi 5-30 menit
5. URL jadi `https://namasite.com` — tidak ada "vercel" ✅

### Domain murah Indonesia:
- **Niagahoster**: `.com` ~Rp150rb/tahun, `.id` ~Rp50rb/tahun
- **Domainesia**: mirip harga
- **Pandi** (`.id`): langsung dari registry Indonesia

---

## ✅ CHECKLIST DEPLOY

```
[ ] STEP 1 — Daftar Neon, catat DATABASE_URL
[ ] STEP 1 — Edit database.go support PostgreSQL
[ ] STEP 2 — Daftar Cloudinary, catat CLOUDINARY_URL
[ ] STEP 2 — Update upload handler pakai Cloudinary
[ ] STEP 3 — Push ke GitHub
[ ] STEP 4 — Deploy backend ke Render + set env vars
[ ] STEP 4 — Catat URL backend Render
[ ] STEP 5 — Deploy frontend ke Vercel + set NEXT_PUBLIC_API_URL
[ ] STEP 6 — Update CORS backend dengan URL Vercel, push ulang
[ ] STEP 6 — Test login admin via URL Vercel
[ ] STEP 7 — (Opsional) Attach custom domain
```

---

## 🆘 TROUBLESHOOTING DEPLOY

| Error | Solusi |
|---|---|
| Render build gagal | Cek Root Directory sudah `backend`, bukan root proyek |
| `go: module not found` | Pastikan `go.mod` ada di folder `backend/` |
| Database connection refused | Cek DATABASE_URL format, pastikan `?sslmode=require` ada di akhir |
| Upload gagal di production | Pastikan CLOUDINARY_URL format benar: `cloudinary://key:secret@cloud` |
| CORS error di Vercel | Tambahkan URL Vercel ke AllowOrigin di backend, push ulang |
| Cold start Render 30 detik | Setup UptimeRobot ping backend setiap 5 menit |
| Image tidak muncul | Pastikan URL Cloudinary tersimpan di database, bukan path lokal |

---

## 💰 ESTIMASI BIAYA

| Service | Free Tier | Batas |
|---|---|---|
| Vercel | Gratis | 100GB bandwidth/bulan |
| Render | Gratis | 750 jam/bulan, cold start |
| Neon | Gratis | 0.5GB storage, 1 project |
| Cloudinary | Gratis | 25GB storage, 25GB bandwidth |
| Domain .com | ~Rp150rb/tahun | — |
| **Total** | **Gratis** (tanpa domain) | — |

> Kalau traffic sudah tinggi dan mau upgrade:
> - Vercel Pro: $20/bulan (lebih bandwidth, analytics)
> - Render Starter: $7/bulan (no cold start)
> - Neon Pro: $19/bulan (lebih storage)