# ⚡ EKSEKUSI PROYEK — VIDEO PLATFORM
> Panduan command + kode siap pakai. Kamu tinggal jalankan, agent handle sisanya.
> **~70% sudah disiapkan di sini, ~30% sisanya debugging & fitur baru via agent.**

---

## 🖥️ PRASYARAT — Install dulu sebelum mulai

```bash
# Cek versi (semua harus ada)
node --version      # >= 18.x
go version          # >= 1.21
git --version

# Kalau belum ada Go: https://go.dev/dl/
# Kalau belum ada Node: https://nodejs.org/
```

---

## 📦 STEP 1 — Buat Struktur Folder Proyek

```bash
# Buat root folder proyek (ganti "videoplatform" sesuai nama proyekmu)
mkdir videoplatform && cd videoplatform

# Buat struktur folder sekaligus
mkdir -p frontend backend/cmd/server backend/internal/{handler,service,repository,middleware} backend/pkg/{storage,response} backend/migrations backend/config uploads

# Init git
git init
echo "node_modules/\n.next/\n.env\n.env.local\nbackend/uploads/\n*.db\n*.exe\ndist/" > .gitignore

echo "✅ Struktur folder siap"
```

---

## 📦 STEP 2 — Setup Frontend (Next.js)

```bash
cd frontend

# Buat project Next.js (pilih: TypeScript ✅, Tailwind ✅, App Router ✅, src/ ❌)
npx create-next-app@latest . --typescript --tailwind --app --no-src-dir --import-alias "@/*"

# Install dependensi tambahan
npm install framer-motion next-themes clsx
npm install @radix-ui/react-dialog @radix-ui/react-dropdown-menu
npm install lucide-react
npm install swr
npm install -D @types/node
```

### 📝 Edit `next.config.ts` — ganti seluruh isinya:

```typescript
// frontend/next.config.ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    remotePatterns: [
      {
        protocol: "http",
        hostname: "localhost",
        port: "8080",
        pathname: "/uploads/**",
      },
    ],
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:8080/api/:path*",
      },
    ];
  },
};

export default nextConfig;
```

### 📝 Edit `app/layout.tsx` — ganti seluruh isinya:

```typescript
// frontend/app/layout.tsx
import type { Metadata } from "next";
import { Poppins, Montserrat } from "next/font/google";
import { ThemeProvider } from "@/components/providers/ThemeProvider";
import "./globals.css";

const poppins = Poppins({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
  variable: "--font-poppins",
});

const montserrat = Montserrat({
  subsets: ["latin"],
  weight: ["400", "500", "600"],
  variable: "--font-montserrat",
});

export const metadata: Metadata = {
  title: "VideoSite",
  description: "Platform video terbaik",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="id" suppressHydrationWarning>
      <body className={`${poppins.variable} ${montserrat.variable}`}>
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          {children}
        </ThemeProvider>
      </body>
    </html>
  );
}
```

### 📝 Edit `app/globals.css` — ganti seluruh isinya:

```css
/* frontend/app/globals.css */
@import "tailwindcss";

:root {
  --font-poppins: 'Poppins', sans-serif;
  --font-montserrat: 'Montserrat', sans-serif;

  /* Light mode tokens */
  --bg-primary: #ffffff;
  --bg-secondary: #f4f4f5;
  --bg-card: #ffffff;
  --surface: #e4e4e7;
  --text-primary: #09090b;
  --text-secondary: #71717a;
  --text-muted: #a1a1aa;
  --accent: #e11d48;
  --accent-hover: #be123c;
  --border: #e4e4e7;
  --overlay: rgba(0, 0, 0, 0.5);
  --shadow: 0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.07), 0 2px 4px rgba(0,0,0,0.06);
}

.dark {
  --bg-primary: #09090b;
  --bg-secondary: #18181b;
  --bg-card: #1c1c1f;
  --surface: #27272a;
  --text-primary: #fafafa;
  --text-secondary: #a1a1aa;
  --text-muted: #71717a;
  --accent: #f43f5e;
  --accent-hover: #e11d48;
  --border: #27272a;
  --overlay: rgba(0, 0, 0, 0.7);
  --shadow: 0 1px 3px rgba(0,0,0,0.4);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.3);
}

* { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: var(--font-montserrat);
  background-color: var(--bg-primary);
  color: var(--text-primary);
  transition: background-color 0.2s, color 0.2s;
}

h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-poppins);
}
```

### 📝 Buat file `components/providers/ThemeProvider.tsx`:

```bash
mkdir -p frontend/components/providers
```

```typescript
// frontend/components/providers/ThemeProvider.tsx
"use client";
import { ThemeProvider as NextThemesProvider } from "next-themes";

export function ThemeProvider({
  children,
  ...props
}: React.ComponentProps<typeof NextThemesProvider>) {
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>;
}
```

### 📝 Buat `components/ui/ThemeToggle.tsx`:

```bash
mkdir -p frontend/components/ui
```

```typescript
// frontend/components/ui/ThemeToggle.tsx
"use client";
import { useTheme } from "next-themes";
import { Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);
  if (!mounted) return null;

  return (
    <button
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      style={{
        background: "var(--surface)",
        border: "1px solid var(--border)",
        borderRadius: "8px",
        padding: "8px",
        cursor: "pointer",
        color: "var(--text-primary)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        transition: "all 0.2s",
      }}
      aria-label="Toggle theme"
    >
      {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
    </button>
  );
}
```

```bash
# Test frontend jalan
cd frontend && npm run dev
# Buka http://localhost:3000 — harus tampil halaman Next.js default
# Ctrl+C untuk stop
```

---

## 📦 STEP 3 — Setup Backend (Go)

```bash
cd ../backend

# Init Go module (ganti "videoplatform" sesuai nama proyekmu)
go mod init videoplatform

# Install dependensi
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors
go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get github.com/joho/godotenv
go get github.com/golang-jwt/jwt/v5
go get github.com/google/uuid
```

### 📝 Buat file `.env` di folder `backend/`:

```env
# backend/.env
PORT=8080
JWT_SECRET=ganti-dengan-secret-panjang-acak-minimal-32-karakter
DB_PATH=./videoplatform.db
UPLOAD_PATH=./uploads
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
```

### 📝 Buat `config/config.go`:

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
	DBPath        string
	UploadPath    string
	AdminUsername string
	AdminPassword string
}

func Load() *Config {
	godotenv.Load()
	return &Config{
		Port:          getEnv("PORT", "8080"),
		JWTSecret:     getEnv("JWT_SECRET", "secret-dev-only"),
		DBPath:        getEnv("DB_PATH", "./videoplatform.db"),
		UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

### 📝 Buat `internal/repository/models.go`:

```go
// backend/internal/repository/models.go
package repository

import (
	"gorm.io/gorm"
	"time"
)

type Video struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Title       string         `gorm:"not null" json:"title"`
	Subtitle    string         `json:"subtitle"`
	Description string         `gorm:"type:text" json:"description"`
	ThumbnailURL string        `json:"thumbnail_url"`
	VideoURL    string         `json:"video_url"`
	PreviewURL  string         `json:"preview_url"`
	Duration    int            `json:"duration"` // detik
	Views       int64          `gorm:"default:0" json:"views"`
	IsPublished bool           `gorm:"default:false" json:"is_published"`
}

type Banner struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	ImageURL  string    `json:"image_url"`
	LinkURL   string    `json:"link_url"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
}

type Ad struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AdCode     string    `gorm:"type:text" json:"ad_code"`
	DeviceType string    `gorm:"default:all" json:"device_type"` // mobile | desktop | all
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	Name       string    `json:"name"`
}
```

### 📝 Buat `internal/repository/database.go`:

```go
// backend/internal/repository/database.go
package repository

import (
	"log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}

	// Auto migrate semua tabel
	db.AutoMigrate(&Video{}, &Banner{}, &Ad{})

	// Seed data awal jika kosong
	seedData(db)

	DB = db
	log.Println("✅ Database siap:", dbPath)
	return db
}

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&Video{}).Count(&count)
	if count > 0 {
		return
	}

	videos := []Video{
		{Title: "Video Pertama", Subtitle: "Subtitle video ini", Description: "Deskripsi lengkap video pertama yang menarik.", ThumbnailURL: "/uploads/placeholder.jpg", VideoURL: "", Duration: 300, IsPublished: true},
		{Title: "Tutorial Go Backend", Subtitle: "Belajar REST API", Description: "Tutorial lengkap membangun REST API dengan Go dan Gin framework.", ThumbnailURL: "/uploads/placeholder.jpg", VideoURL: "", Duration: 600, IsPublished: true},
		{Title: "Next.js Dasar", Subtitle: "Frontend modern", Description: "Pengenalan Next.js App Router untuk pemula.", ThumbnailURL: "/uploads/placeholder.jpg", VideoURL: "", Duration: 450, IsPublished: false},
	}
	db.Create(&videos)

	db.Create(&Banner{Title: "Banner Utama", ImageURL: "/uploads/placeholder.jpg", LinkURL: "#", IsActive: true, SortOrder: 1})
	db.Create(&Ad{Name: "Iklan Utama", AdCode: `<div style="background:#f0f0f0;padding:20px;text-align:center">Slot Iklan</div>`, DeviceType: "all", IsActive: true})

	log.Println("✅ Seed data berhasil")
}
```

### 📝 Buat `pkg/response/response.go`:

```go
// backend/pkg/response/response.go
package response

import "github.com/gin-gonic/gin"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Response{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(201, Response{Success: true, Data: data})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, Response{Success: false, Message: msg})
}

func Paginated(c *gin.Context, data interface{}, total int64, page, limit int) {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	c.JSON(200, PaginatedResponse{
		Success: true, Data: data,
		Total: total, Page: page,
		Limit: limit, TotalPages: totalPages,
	})
}
```

### 📝 Buat `internal/middleware/auth.go`:

```go
// backend/internal/middleware/auth.go
package middleware

import (
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"videoplatform/pkg/response"
)

func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cek cookie dulu, lalu header
		tokenStr, err := c.Cookie("admin_token")
		if err != nil {
			auth := c.GetHeader("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				response.Error(c, 401, "Unauthorized")
				c.Abort()
				return
			}
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Error(c, 401, "Token tidak valid")
			c.Abort()
			return
		}
		c.Next()
	}
}
```

### 📝 Buat `internal/handler/video_handler.go`:

```go
// backend/internal/handler/video_handler.go
package handler

import (
	"math"
	"strconv"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"videoplatform/internal/repository"
	"videoplatform/pkg/response"
)

type VideoHandler struct{ db *gorm.DB }

func NewVideoHandler(db *gorm.DB) *VideoHandler { return &VideoHandler{db: db} }

func (h *VideoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	search := c.Query("search")
	if page < 1 { page = 1 }
	if limit < 1 || limit > 50 { limit = 12 }
	offset := (page - 1) * limit

	var videos []repository.Video
	var total int64
	q := h.db.Model(&repository.Video{}).Where("is_published = ?", true)
	if search != "" {
		q = q.Where("title LIKE ? OR subtitle LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	q.Count(&total)
	q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos)
	response.Paginated(c, videos, total, page, limit)
}

func (h *VideoHandler) Get(c *gin.Context) {
	var video repository.Video
	if err := h.db.First(&video, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	// Increment views
	h.db.Model(&video).UpdateColumn("views", gorm.Expr("views + ?", 1))
	video.Views++
	response.OK(c, video)
}

func (h *VideoHandler) Featured(c *gin.Context) {
	var videos []repository.Video
	h.db.Where("is_published = ?", true).Order("views DESC").Limit(6).Find(&videos)
	response.OK(c, videos)
}

// Admin handlers
func (h *VideoHandler) AdminList(c *gin.Context) {
	var videos []repository.Video
	var total int64
	h.db.Model(&repository.Video{}).Count(&total)
	h.db.Order("created_at DESC").Find(&videos)
	response.Paginated(c, videos, total, 1, int(math.Max(float64(total), 1)))
}

func (h *VideoHandler) Create(c *gin.Context) {
	var video repository.Video
	if err := c.ShouldBindJSON(&video); err != nil {
		response.Error(c, 400, "Data tidak valid: "+err.Error())
		return
	}
	h.db.Create(&video)
	response.Created(c, video)
}

func (h *VideoHandler) Update(c *gin.Context) {
	var video repository.Video
	if err := h.db.First(&video, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	if err := c.ShouldBindJSON(&video); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	h.db.Save(&video)
	response.OK(c, video)
}

func (h *VideoHandler) Delete(c *gin.Context) {
	if err := h.db.Delete(&repository.Video{}, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}
```

### 📝 Buat `internal/handler/auth_handler.go`:

```go
// backend/internal/handler/auth_handler.go
package handler

import (
	"time"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"videoplatform/pkg/response"
)

type AuthHandler struct {
	Username  string
	Password  string
	JWTSecret string
}

func NewAuthHandler(username, password, secret string) *AuthHandler {
	return &AuthHandler{Username: username, Password: password, JWTSecret: secret}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	if req.Username != h.Username || req.Password != h.Password {
		response.Error(c, 401, "Username atau password salah")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(h.JWTSecret))

	// Set cookie httpOnly
	c.SetCookie("admin_token", tokenStr, 86400, "/", "", false, true)
	response.OK(c, gin.H{"token": tokenStr})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("admin_token", "", -1, "/", "", false, true)
	response.OK(c, gin.H{"message": "Logout berhasil"})
}
```

### 📝 Buat `internal/handler/upload_handler.go`:

```go
// backend/internal/handler/upload_handler.go
package handler

import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"videoplatform/pkg/response"
)

type UploadHandler struct{ uploadPath string }

func NewUploadHandler(uploadPath string) *UploadHandler {
	return &UploadHandler{uploadPath: uploadPath}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, "File tidak ditemukan")
		return
	}

	// Validasi ekstensi
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".mp4": true, ".webm": true, ".mov": true}
	if !allowed[ext] {
		response.Error(c, 400, "Format file tidak didukung")
		return
	}

	// Batasi ukuran
	if file.Size > 500*1024*1024 { // 500MB
		response.Error(c, 400, "File terlalu besar (maks 500MB)")
		return
	}

	filename := uuid.New().String() + ext
	dst := filepath.Join(h.uploadPath, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Error(c, 500, "Gagal menyimpan file")
		return
	}

	url := fmt.Sprintf("/uploads/%s", filename)
	response.OK(c, gin.H{"url": url, "filename": filename})
}
```

### 📝 Buat `internal/handler/banner_handler.go`:

```go
// backend/internal/handler/banner_handler.go
package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"videoplatform/internal/repository"
	"videoplatform/pkg/response"
)

type BannerHandler struct{ db *gorm.DB }

func NewBannerHandler(db *gorm.DB) *BannerHandler { return &BannerHandler{db: db} }

func (h *BannerHandler) ListActive(c *gin.Context) {
	var banners []repository.Banner
	h.db.Where("is_active = ?", true).Order("sort_order ASC").Find(&banners)
	response.OK(c, banners)
}

func (h *BannerHandler) AdminList(c *gin.Context) {
	var banners []repository.Banner
	h.db.Order("sort_order ASC").Find(&banners)
	response.OK(c, banners)
}

func (h *BannerHandler) Create(c *gin.Context) {
	var banner repository.Banner
	if err := c.ShouldBindJSON(&banner); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	h.db.Create(&banner)
	response.Created(c, banner)
}

func (h *BannerHandler) Update(c *gin.Context) {
	var banner repository.Banner
	if err := h.db.First(&banner, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Banner tidak ditemukan")
		return
	}
	c.ShouldBindJSON(&banner)
	h.db.Save(&banner)
	response.OK(c, banner)
}

func (h *BannerHandler) Delete(c *gin.Context) {
	h.db.Delete(&repository.Banner{}, c.Param("id"))
	response.OK(c, gin.H{"deleted": true})
}
```

### 📝 Buat `internal/handler/ad_handler.go`:

```go
// backend/internal/handler/ad_handler.go
package handler

import (
	"strings"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"videoplatform/internal/repository"
	"videoplatform/pkg/response"
)

type AdHandler struct{ db *gorm.DB }

func NewAdHandler(db *gorm.DB) *AdHandler { return &AdHandler{db: db} }

func (h *AdHandler) Serve(c *gin.Context) {
	ua := strings.ToLower(c.GetHeader("User-Agent"))
	deviceType := "desktop"
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		deviceType = "mobile"
	}

	var ad repository.Ad
	result := h.db.Where("is_active = ? AND (device_type = ? OR device_type = 'all')", true, deviceType).
		Order("RANDOM()").First(&ad)
	if result.Error != nil {
		response.OK(c, nil)
		return
	}
	response.OK(c, ad)
}

func (h *AdHandler) AdminList(c *gin.Context) {
	var ads []repository.Ad
	h.db.Find(&ads)
	response.OK(c, ads)
}

func (h *AdHandler) Create(c *gin.Context) {
	var ad repository.Ad
	if err := c.ShouldBindJSON(&ad); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	h.db.Create(&ad)
	response.Created(c, ad)
}

func (h *AdHandler) Update(c *gin.Context) {
	var ad repository.Ad
	if err := h.db.First(&ad, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Iklan tidak ditemukan")
		return
	}
	c.ShouldBindJSON(&ad)
	h.db.Save(&ad)
	response.OK(c, ad)
}

func (h *AdHandler) Delete(c *gin.Context) {
	h.db.Delete(&repository.Ad{}, c.Param("id"))
	response.OK(c, gin.H{"deleted": true})
}
```

### 📝 Buat `cmd/server/main.go` — File utama backend:

```go
// backend/cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"videoplatform/config"
	"videoplatform/internal/handler"
	"videoplatform/internal/middleware"
	"videoplatform/internal/repository"
)

func main() {
	cfg := config.Load()

	// Pastikan folder uploads ada
	os.MkdirAll(cfg.UploadPath, 0755)

	// Init database
	db := repository.InitDB(cfg.DBPath)

	// Init handlers
	videoH := handler.NewVideoHandler(db)
	bannerH := handler.NewBannerHandler(db)
	adH := handler.NewAdHandler(db)
	authH := handler.NewAuthHandler(cfg.AdminUsername, cfg.AdminPassword, cfg.JWTSecret)
	uploadH := handler.NewUploadHandler(cfg.UploadPath)

	// Setup Gin
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Static files untuk uploads
	r.Static("/uploads", cfg.UploadPath)

	// === PUBLIC ROUTES ===
	api := r.Group("/api")
	{
		// Videos
		api.GET("/videos", videoH.List)
		api.GET("/videos/:id", videoH.Get)
		api.GET("/videos/featured", videoH.Featured)

		// Banners & Ads
		api.GET("/banners", bannerH.ListActive)
		api.GET("/ads/serve", adH.Serve)

		// Auth
		api.POST("/admin/login", authH.Login)
		api.POST("/admin/logout", authH.Logout)
	}

	// === ADMIN ROUTES (protected) ===
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		// Upload
		admin.POST("/upload", uploadH.Upload)

		// Videos CRUD
		admin.GET("/videos", videoH.AdminList)
		admin.POST("/videos", videoH.Create)
		admin.PUT("/videos/:id", videoH.Update)
		admin.DELETE("/videos/:id", videoH.Delete)

		// Banners CRUD
		admin.GET("/banners", bannerH.AdminList)
		admin.POST("/banners", bannerH.Create)
		admin.PUT("/banners/:id", bannerH.Update)
		admin.DELETE("/banners/:id", bannerH.Delete)

		// Ads CRUD
		admin.GET("/ads", adH.AdminList)
		admin.POST("/ads", adH.Create)
		admin.PUT("/ads/:id", adH.Update)
		admin.DELETE("/ads/:id", adH.Delete)

		// Favicon dummy biar tidak error 404
		r.GET("/favicon.ico", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	}

	log.Printf("🚀 Server jalan di http://localhost:%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
```

```bash
# Test backend jalan
cd backend && go run cmd/server/main.go
# Harus muncul: 🚀 Server jalan di http://localhost:8080
# Test di browser: http://localhost:8080/api/videos
# Ctrl+C untuk stop
```

---

## 📦 STEP 4 — Frontend: API Layer & VideoCard

### 📝 Buat `lib/api.ts`:

```bash
mkdir -p frontend/lib
```

```typescript
// frontend/lib/api.ts
const BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface Video {
  id: number;
  title: string;
  subtitle: string;
  description: string;
  thumbnail_url: string;
  video_url: string;
  preview_url: string;
  duration: number;
  views: number;
  is_published: boolean;
  created_at: string;
}

export interface Banner {
  id: number;
  title: string;
  image_url: string;
  link_url: string;
  is_active: boolean;
  sort_order: number;
}

export interface Ad {
  id: number;
  name: string;
  ad_code: string;
  device_type: string;
  is_active: boolean;
}

export interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`, { credentials: "include" });
  if (!res.ok) throw new Error(`API Error: ${res.status}`);
  const json = await res.json();
  return json.data ?? json;
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(body),
  });
  const json = await res.json();
  if (!json.success) throw new Error(json.message || "Error");
  return json.data;
}

async function put<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(body),
  });
  const json = await res.json();
  if (!json.success) throw new Error(json.message || "Error");
  return json.data;
}

async function del(path: string): Promise<void> {
  await fetch(`${BASE}${path}`, { method: "DELETE", credentials: "include" });
}

// Public APIs
export const api = {
  videos: {
    list: (page = 1, search = "") =>
      get<PaginatedResponse<Video>>(`/api/videos?page=${page}&search=${search}`),
    get: (id: number) => get<Video>(`/api/videos/${id}`),
    featured: () => get<Video[]>(`/api/videos/featured`),
  },
  banners: {
    list: () => get<Banner[]>(`/api/banners`),
  },
  ads: {
    serve: () => get<Ad | null>(`/api/ads/serve`),
  },
};

// Admin APIs
export const adminApi = {
  auth: {
    login: (username: string, password: string) =>
      post<{ token: string }>("/api/admin/login", { username, password }),
    logout: () => post("/api/admin/logout", {}),
  },
  videos: {
    list: () => get<{ data: Video[] }>(`/api/admin/videos`),
    create: (data: Partial<Video>) => post<Video>("/api/admin/videos", data),
    update: (id: number, data: Partial<Video>) => put<Video>(`/api/admin/videos/${id}`, data),
    delete: (id: number) => del(`/api/admin/videos/${id}`),
  },
  banners: {
    list: () => get<Banner[]>(`/api/admin/banners`),
    create: (data: Partial<Banner>) => post<Banner>("/api/admin/banners", data),
    update: (id: number, data: Partial<Banner>) => put<Banner>(`/api/admin/banners/${id}`, data),
    delete: (id: number) => del(`/api/admin/banners/${id}`),
  },
  upload: async (file: File): Promise<{ url: string }> => {
    const form = new FormData();
    form.append("file", file);
    const res = await fetch(`${BASE}/api/admin/upload`, {
      method: "POST",
      credentials: "include",
      body: form,
    });
    const json = await res.json();
    if (!json.success) throw new Error(json.message);
    return json.data;
  },
};

export function mediaUrl(path: string) {
  if (!path) return "/placeholder.jpg";
  if (path.startsWith("http")) return path;
  return `${BASE}${path}`;
}

export function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function formatViews(views: number): string {
  if (views >= 1_000_000) return `${(views / 1_000_000).toFixed(1)}Jt`;
  if (views >= 1_000) return `${(views / 1_000).toFixed(1)}Rb`;
  return views.toString();
}
```

### 📝 Buat `components/video/VideoCard.tsx`:

```bash
mkdir -p frontend/components/video
```

```typescript
// frontend/components/video/VideoCard.tsx
"use client";
import { useRef, useState, useEffect, useCallback } from "react";
import Image from "next/image";
import { Eye, Clock } from "lucide-react";
import type { Video } from "@/lib/api";
import { mediaUrl, formatDuration, formatViews } from "@/lib/api";

// Manager global: maks 3 preview aktif
const activeVideos = new Set<HTMLVideoElement>();
const MAX_ACTIVE = 3;

interface Props {
  video: Video;
  onClick?: (id: number) => void;
}

export function VideoCard({ video, onClick }: Props) {
  const [isHovered, setIsHovered] = useState(false);
  const [showPreview, setShowPreview] = useState(false);
  const [imgLoaded, setImgLoaded] = useState(false);
  const videoRef = useRef<HTMLVideoElement>(null);
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const cardRef = useRef<HTMLDivElement>(null);

  // IntersectionObserver: stop video saat keluar viewport
  useEffect(() => {
    const el = cardRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry.isIntersecting && videoRef.current) {
          videoRef.current.pause();
          activeVideos.delete(videoRef.current);
          setShowPreview(false);
        }
      },
      { threshold: 0.3 }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  const handleMouseEnter = useCallback(() => {
    setIsHovered(true);
    if (!video.preview_url) return;
    hoverTimer.current = setTimeout(() => {
      if (activeVideos.size >= MAX_ACTIVE) {
        // Stop video terlama
        const oldest = activeVideos.values().next().value;
        if (oldest) { oldest.pause(); activeVideos.delete(oldest); }
      }
      setShowPreview(true);
      if (videoRef.current) {
        videoRef.current.play().catch(() => {});
        activeVideos.add(videoRef.current);
      }
    }, 800);
  }, [video.preview_url]);

  const handleMouseLeave = useCallback(() => {
    setIsHovered(false);
    if (hoverTimer.current) clearTimeout(hoverTimer.current);
    setShowPreview(false);
    if (videoRef.current) {
      videoRef.current.pause();
      videoRef.current.currentTime = 0;
      activeVideos.delete(videoRef.current);
    }
  }, []);

  // Cleanup
  useEffect(() => {
    return () => {
      if (hoverTimer.current) clearTimeout(hoverTimer.current);
      if (videoRef.current) activeVideos.delete(videoRef.current);
    };
  }, []);

  return (
    <div
      ref={cardRef}
      onClick={() => onClick?.(video.id)}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      style={{
        cursor: "pointer",
        borderRadius: "12px",
        overflow: "hidden",
        background: "var(--bg-card)",
        border: "1px solid var(--border)",
        boxShadow: isHovered ? "var(--shadow-md)" : "var(--shadow)",
        transition: "transform 0.2s, box-shadow 0.2s",
        transform: isHovered ? "translateY(-2px)" : "translateY(0)",
      }}
    >
      {/* Thumbnail area */}
      <div style={{ position: "relative", aspectRatio: "16/9", background: "var(--surface)", overflow: "hidden" }}>
        {/* Thumbnail */}
        <Image
          src={mediaUrl(video.thumbnail_url)}
          alt={video.title}
          fill
          sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
          style={{
            objectFit: "cover",
            opacity: showPreview ? 0 : imgLoaded ? 1 : 0,
            transition: "opacity 0.3s",
          }}
          onLoad={() => setImgLoaded(true)}
        />

        {/* Preview video */}
        {video.preview_url && (
          <video
            ref={videoRef}
            src={mediaUrl(video.preview_url)}
            muted
            loop
            playsInline
            style={{
              position: "absolute", inset: 0,
              width: "100%", height: "100%",
              objectFit: "cover",
              opacity: showPreview ? 1 : 0,
              transition: "opacity 0.3s",
            }}
          />
        )}

        {/* Duration badge */}
        <span style={{
          position: "absolute", bottom: "8px", right: "8px",
          background: "rgba(0,0,0,0.8)", color: "#fff",
          borderRadius: "4px", padding: "2px 6px",
          fontSize: "12px", fontWeight: 600,
          fontFamily: "var(--font-poppins)",
        }}>
          {formatDuration(video.duration)}
        </span>
      </div>

      {/* Info area */}
      <div style={{ padding: "12px" }}>
        {/* Title */}
        <p style={{
          fontFamily: "var(--font-poppins)",
          fontWeight: 600,
          fontSize: "14px",
          color: "var(--text-primary)",
          lineHeight: 1.4,
          display: "-webkit-box",
          WebkitLineClamp: 2,
          WebkitBoxOrient: "vertical",
          overflow: "hidden",
          marginBottom: "4px",
        }}>
          {video.title}
        </p>

        {/* Subtitle */}
        <p style={{
          fontSize: "12px",
          color: "var(--text-secondary)",
          marginBottom: "8px",
          overflow: "hidden",
          textOverflow: "ellipsis",
          whiteSpace: "nowrap",
        }}>
          {video.subtitle}
        </p>

        {/* Deskripsi — muncul saat hover */}
        <div style={{
          overflow: "hidden",
          maxHeight: isHovered ? "60px" : "0",
          transition: "max-height 0.3s ease",
          marginBottom: isHovered ? "8px" : "0",
        }}>
          <p style={{
            fontSize: "11px",
            color: "var(--text-muted)",
            lineHeight: 1.5,
            display: "-webkit-box",
            WebkitLineClamp: 3,
            WebkitBoxOrient: "vertical",
            overflow: "hidden",
          }}>
            {video.description}
          </p>
        </div>

        {/* Meta: views & waktu */}
        <div style={{ display: "flex", gap: "12px", alignItems: "center" }}>
          <span style={{ display: "flex", alignItems: "center", gap: "4px", fontSize: "11px", color: "var(--text-muted)" }}>
            <Eye size={12} /> {formatViews(video.views)} penonton
          </span>
          <span style={{ display: "flex", alignItems: "center", gap: "4px", fontSize: "11px", color: "var(--text-muted)" }}>
            <Clock size={12} /> {new Date(video.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" })}
          </span>
        </div>
      </div>
    </div>
  );
}
```

### 📝 Buat `components/layout/Navbar.tsx`:

```bash
mkdir -p frontend/components/layout
```

```typescript
// frontend/components/layout/Navbar.tsx
"use client";
import Link from "next/link";
import Image from "next/image";
import { Search } from "lucide-react";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { ThemeToggle } from "@/components/ui/ThemeToggle";

export function Navbar() {
  const [search, setSearch] = useState("");
  const router = useRouter();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (search.trim()) router.push(`/search?q=${encodeURIComponent(search)}`);
  };

  return (
    <nav style={{
      position: "sticky", top: 0, zIndex: 100,
      background: "var(--bg-primary)",
      borderBottom: "1px solid var(--border)",
      backdropFilter: "blur(8px)",
    }}>
      <div style={{
        maxWidth: "1400px", margin: "0 auto",
        padding: "0 24px", height: "64px",
        display: "flex", alignItems: "center", gap: "16px",
      }}>
        {/* Logo */}
        <Link href="/" style={{ display: "flex", alignItems: "center", textDecoration: "none", flexShrink: 0 }}>
          {/* Ganti dengan komponen Logo.tsx kamu nanti */}
          <span style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "20px", color: "var(--accent)" }}>
            VideoSite
          </span>
        </Link>

        {/* Search bar */}
        <form onSubmit={handleSearch} style={{ flex: 1, maxWidth: "480px", display: "flex" }}>
          <div style={{ position: "relative", width: "100%" }}>
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari video..."
              style={{
                width: "100%", padding: "8px 16px 8px 40px",
                borderRadius: "24px",
                border: "1px solid var(--border)",
                background: "var(--bg-secondary)",
                color: "var(--text-primary)",
                fontSize: "14px", outline: "none",
                fontFamily: "var(--font-montserrat)",
              }}
            />
            <Search size={16} style={{ position: "absolute", left: "14px", top: "50%", transform: "translateY(-50%)", color: "var(--text-muted)" }} />
          </div>
        </form>

        <div style={{ marginLeft: "auto" }}>
          <ThemeToggle />
        </div>
      </div>
    </nav>
  );
}
```

### 📝 Buat `app/(public)/layout.tsx`:

```bash
mkdir -p "frontend/app/(public)"
```

```typescript
// frontend/app/(public)/layout.tsx
import { Navbar } from "@/components/layout/Navbar";

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <Navbar />
      <main style={{ maxWidth: "1400px", margin: "0 auto", padding: "24px" }}>
        {children}
      </main>
    </>
  );
}
```

### 📝 Buat `app/(public)/page.tsx` — Halaman utama:

```typescript
// frontend/app/(public)/page.tsx
import { api } from "@/lib/api";
import { VideoCard } from "@/components/video/VideoCard";
import { VideoGridClient } from "@/components/video/VideoGridClient";

export const revalidate = 60;

export default async function HomePage() {
  // Fetch data di server
  const [videosRes, banners] = await Promise.allSettled([
    api.videos.list(1),
    api.banners.list(),
  ]);

  const videos = videosRes.status === "fulfilled" ? videosRes.value.data : [];

  return (
    <div>
      {/* Banner section — nanti tambahkan slider */}
      <section style={{ marginBottom: "32px" }}>
        <div style={{
          borderRadius: "16px", overflow: "hidden",
          background: "var(--surface)",
          height: "240px",
          display: "flex", alignItems: "center", justifyContent: "center",
        }}>
          <p style={{ color: "var(--text-muted)", fontFamily: "var(--font-poppins)" }}>
            📸 Banner Slider — hubungkan ke API banners
          </p>
        </div>
      </section>

      {/* Slot iklan */}
      <div id="ad-slot-top" style={{ marginBottom: "24px" }} />

      {/* Video grid */}
      <section>
        <h2 style={{
          fontFamily: "var(--font-poppins)", fontWeight: 700,
          fontSize: "20px", marginBottom: "16px",
          color: "var(--text-primary)",
        }}>
          Video Terbaru
        </h2>
        <VideoGridClient initialVideos={videos} />
      </section>
    </div>
  );
}
```

### 📝 Buat `components/video/VideoGridClient.tsx`:

```typescript
// frontend/components/video/VideoGridClient.tsx
"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { VideoCard } from "./VideoCard";
import type { Video } from "@/lib/api";

export function VideoGridClient({ initialVideos }: { initialVideos: Video[] }) {
  const [videos] = useState(initialVideos);
  const router = useRouter();

  return (
    <div style={{
      display: "grid",
      gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))",
      gap: "20px",
    }}>
      {videos.map((v) => (
        <VideoCard
          key={v.id}
          video={v}
          onClick={(id) => router.push(`/watch/${id}`)}
        />
      ))}
      {videos.length === 0 && (
        <p style={{ color: "var(--text-muted)", gridColumn: "1/-1", textAlign: "center", padding: "48px 0" }}>
          Belum ada video yang dipublikasikan.
        </p>
      )}
    </div>
  );
}
```

---

## 📦 STEP 5 — Halaman Watch & Admin Login

### 📝 Buat `app/(public)/watch/[id]/page.tsx`:

```bash
mkdir -p "frontend/app/(public)/watch/[id]"
```

```typescript
// frontend/app/(public)/watch/[id]/page.tsx
import { api, mediaUrl, formatViews } from "@/lib/api";
import { notFound } from "next/navigation";
import { VideoCard } from "@/components/video/VideoCard";

export default async function WatchPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  
  let video, relatedRes;
  try {
    [video, relatedRes] = await Promise.all([
      api.videos.get(Number(id)),
      api.videos.list(1),
    ]);
  } catch {
    notFound();
  }

  const related = relatedRes.data.filter((v) => v.id !== video.id).slice(0, 8);

  return (
    <div style={{ display: "grid", gridTemplateColumns: "1fr 360px", gap: "24px" }}>
      {/* Main content */}
      <div>
        {/* Player */}
        <div style={{ aspectRatio: "16/9", background: "#000", borderRadius: "12px", overflow: "hidden", marginBottom: "16px" }}>
          {video.video_url ? (
            <video
              src={mediaUrl(video.video_url)}
              controls
              autoPlay
              style={{ width: "100%", height: "100%" }}
            />
          ) : (
            <div style={{ display: "flex", alignItems: "center", justifyContent: "center", height: "100%", color: "#888" }}>
              Video belum tersedia
            </div>
          )}
        </div>

        {/* Info */}
        <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px", marginBottom: "4px" }}>
          {video.title}
        </h1>
        <p style={{ fontSize: "13px", color: "var(--text-secondary)", marginBottom: "8px" }}>{video.subtitle}</p>
        <p style={{ fontSize: "13px", color: "var(--text-muted)", marginBottom: "16px" }}>
          {formatViews(video.views)} penonton · {new Date(video.created_at).toLocaleDateString("id-ID", { day: "numeric", month: "long", year: "numeric" })}
        </p>

        {/* Deskripsi */}
        <div style={{ background: "var(--bg-secondary)", borderRadius: "12px", padding: "16px" }}>
          <p style={{ fontSize: "14px", lineHeight: 1.7, color: "var(--text-primary)", whiteSpace: "pre-wrap" }}>
            {video.description || "Tidak ada deskripsi."}
          </p>
        </div>
      </div>

      {/* Sidebar */}
      <div>
        <h3 style={{ fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "16px", marginBottom: "12px" }}>
          Video Lainnya
        </h3>
        <div style={{ display: "flex", flexDirection: "column", gap: "12px" }}>
          {related.map((v) => (
            <VideoCard key={v.id} video={v} onClick={(id) => window.location.href = `/watch/${id}`} />
          ))}
        </div>
      </div>
    </div>
  );
}
```

### 📝 Buat halaman admin login:

```bash
mkdir -p "frontend/app/(admin)/login"
mkdir -p "frontend/app/(admin)"
```

```typescript
// frontend/app/(admin)/login/page.tsx
"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { adminApi } from "@/lib/api";

export default function AdminLoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true); setError("");
    try {
      await adminApi.auth.login(username, password);
      router.push("/admin/dashboard");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Login gagal");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", background: "var(--bg-secondary)" }}>
      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "16px", padding: "40px", width: "100%", maxWidth: "400px", boxShadow: "var(--shadow-md)" }}>
        <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "24px", marginBottom: "8px", textAlign: "center" }}>Admin Panel</h1>
        <p style={{ textAlign: "center", color: "var(--text-secondary)", marginBottom: "32px", fontSize: "14px" }}>Masuk untuk mengelola konten</p>

        {error && (
          <div style={{ background: "#fee2e2", border: "1px solid #fca5a5", borderRadius: "8px", padding: "12px", marginBottom: "16px", color: "#dc2626", fontSize: "14px" }}>
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          <div>
            <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Username</label>
            <input value={username} onChange={(e) => setUsername(e.target.value)} required
              style={{ width: "100%", padding: "10px 14px", borderRadius: "8px", border: "1px solid var(--border)", background: "var(--bg-secondary)", color: "var(--text-primary)", fontSize: "14px", outline: "none", fontFamily: "var(--font-montserrat)" }} />
          </div>
          <div>
            <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Password</label>
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required
              style={{ width: "100%", padding: "10px 14px", borderRadius: "8px", border: "1px solid var(--border)", background: "var(--bg-secondary)", color: "var(--text-primary)", fontSize: "14px", outline: "none", fontFamily: "var(--font-montserrat)" }} />
          </div>
          <button type="submit" disabled={loading}
            style={{ padding: "12px", borderRadius: "8px", background: "var(--accent)", color: "#fff", border: "none", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "15px", cursor: loading ? "not-allowed" : "pointer", opacity: loading ? 0.7 : 1, transition: "opacity 0.2s" }}>
            {loading ? "Masuk..." : "Masuk"}
          </button>
        </form>
      </div>
    </div>
  );
}
```

---

## 📦 STEP 6 — Jalankan Kedua Server

```bash
# Terminal 1 — Backend
cd backend && go run cmd/server/main.go

# Terminal 2 — Frontend
cd frontend && npm run dev
```

**Cek endpoint:**
```bash
# Harus return JSON daftar video
curl http://localhost:8080/api/videos

# Harus return banner aktif
curl http://localhost:8080/api/banners

# Test login admin
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

**Cek di browser:**
- `http://localhost:3000` → Homepage
- `http://localhost:3000/admin/login` → Admin login

---

## 📦 STEP 7 — Serahkan Sisanya ke Agent (30%)

Setelah STEP 1–6 berjalan, bagian berikut **minta ke agent** dengan prompt yang sudah disediakan:

### 🤖 Prompt untuk Agent: Admin Dashboard

```
Buatkan halaman admin dashboard lengkap di:
- app/(admin)/dashboard/page.tsx  ← stat cards (total video, views, dll)
- app/(admin)/videos/page.tsx     ← tabel video + tombol tambah/edit/hapus
- app/(admin)/videos/new/page.tsx ← form tambah video dengan drag & drop upload
- app/(admin)/layout.tsx          ← sidebar navigasi admin

Gunakan adminApi dari @/lib/api yang sudah ada.
Semua styling pakai CSS variables (var(--bg-card), dll), jangan hardcode warna.
Font: var(--font-poppins) untuk heading, var(--font-montserrat) untuk body.
```

### 🤖 Prompt untuk Agent: Banner Slider

```
Buatkan komponen BannerSlider.tsx di components/video/ dengan:
- Auto-slide setiap 5 detik
- Tombol prev/next
- Dot indicator
- Fade atau slide transition smooth
- Data dari prop: banners: Banner[]  (type dari @/lib/api)
- Fallback jika tidak ada banner
Integrasikan ke app/(public)/page.tsx menggantikan placeholder yang ada.
```

### 🤖 Prompt untuk Agent: Sistem Iklan

```
Buatkan komponen AdSlot.tsx di components/ui/ dengan:
- Cek sessionStorage key "ad_shown"
- Jika belum: fetch GET /api/ads/serve, render ad_code via dangerouslySetInnerHTML, set sessionStorage
- Jika sudah: render null (tidak tampil)
- Props: position ("top" | "sidebar" | "bottom")
Tempatkan di homepage (setelah banner) dan watch page (bawah player).
```

### 🤖 Prompt untuk Agent: Halaman Search

```
Buatkan halaman pencarian di app/(public)/search/page.tsx:
- Ambil query param ?q= dari URL
- Fetch GET /api/videos?search={q}
- Tampilkan grid VideoCard hasil pencarian
- Loading skeleton saat fetch
- Pesan "Tidak ada hasil untuk X" jika kosong
```

### 🤖 Prompt untuk Agent: Middleware Admin Auth

```
Buatkan middleware Next.js di frontend/middleware.ts:
- Proteksi semua route /admin/* kecuali /admin/login
- Cek cookie "admin_token" — jika tidak ada, redirect ke /admin/login
- Jika sudah login dan akses /admin/login, redirect ke /admin/dashboard
```

---

## 🗂️ CHECKLIST PROGRES

```
[ ] STEP 1 — Folder struktur & git init
[ ] STEP 2 — Next.js setup, tema, font, ThemeToggle
[ ] STEP 3 — Go backend, database, semua handler, main.go
[ ] STEP 4 — api.ts, VideoCard, Navbar, homepage
[ ] STEP 5 — Watch page, Admin login page
[ ] STEP 6 — Kedua server jalan, endpoint test OK
[ ] STEP 7 — Agent: Admin dashboard
[ ] STEP 7 — Agent: Banner slider
[ ] STEP 7 — Agent: Sistem iklan
[ ] STEP 7 — Agent: Halaman search
[ ] STEP 7 — Agent: Middleware auth
```

---

## ⚠️ TROUBLESHOOTING UMUM

| Error | Solusi |
|---|---|
| `go: module not found` | Pastikan kamu di folder `backend/` saat `go mod init` |
| `sqlite3: cgo not enabled` | Install GCC: `sudo apt install gcc` atau pakai `go-sqlite3` pure Go fork |
| `CORS error` di browser | Cek `AllowOrigins` di main.go sesuai port frontend |
| `next/image` domain error | Tambahkan domain backend di `next.config.ts` → `remotePatterns` |
| Cookie tidak terkirim | Pastikan `credentials: "include"` di semua fetch, dan `AllowCredentials: true` di CORS Go |
| Port 8080 sudah dipakai | Ganti `PORT=8081` di `.env` backend dan update `next.config.ts` |

---

> **🔑 Kunci sukses:** Selesaikan STEP 1–6 dulu sampai kedua server jalan dan homepage tampil.
> Baru serahkan STEP 7 ke agent satu per satu — jangan sekaligus biar mudah debug.
