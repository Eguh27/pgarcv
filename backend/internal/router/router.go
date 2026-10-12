// Package router memuat seluruh definisi route HTTP dalam satu tempat, supaya
// cmd/server/main.go hanya bertugas merakit dependensi dan menjalankan server.
package router

import (
	"net/http"
	"strings"
	"time"

	"pgarcv/internal/config"
	"pgarcv/internal/handler"
	"pgarcv/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// maxMultipartMemory adalah memori maksimum yang dialokasikan Gin untuk
// multipart form sebelum menyalin sisanya ke disk.
const maxMultipartMemory = 64 << 20

// corsMaxAge adalah lamanya browser boleh menyimpan hasil preflight CORS.
const corsMaxAge = 12 * time.Hour

// Handlers adalah kumpulan handler yang dirangkai jadi satu engine.
type Handlers struct {
	Video         *handler.VideoHandler
	Banner        *handler.BannerHandler
	Ad            *handler.AdHandler
	VideoSuggest  *handler.VideoSuggestHandler
	VideoDownload *handler.VideoDownloadHandler
	Auth          *handler.AuthHandler
	Upload        *handler.UploadHandler
	UploadVideo   *handler.UploadVideoHandler
	HLS           *handler.HLSHandler
	ChunkedUpload *handler.ChunkedUploadHandler
}

// New menyusun engine Gin dengan seluruh route aplikasi.
func New(cfg *config.Config, h Handlers) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = maxMultipartMemory

	r.Use(newCORS(cfg))

	// Aset statis: file yang sudah diunggah pengguna.
	r.Static("/uploads", cfg.UploadPath)

	registerPublicRoutes(r, cfg, h)
	registerAdminRoutes(r, cfg, h)

	return r
}

// newCORS membangun middleware CORS dari whitelist origin eksplisit.
func newCORS(cfg *config.Config) gin.HandlerFunc {
	allowed := allowedOrigins(cfg)
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return allowed[origin]
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"ngrok-skip-browser-warning",
		},
		AllowCredentials: true,
		MaxAge:           corsMaxAge,
	})
}

// allowedOrigins menyusun set origin yang diizinkan.
//
// Origin dibaca dari CORS_ORIGIN (pisahkan dengan koma). Di luar production,
// localhost ditambahkan otomatis supaya pengembangan lokal tidak perlu
// mengonfigurasi ulang tiap kali port berubah.
func allowedOrigins(cfg *config.Config) map[string]bool {
	allowed := make(map[string]bool)
	for _, origin := range strings.Split(cfg.CORSOrigin, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			allowed[origin] = true
		}
	}
	if !cfg.IsProduction() {
		allowed["http://localhost:3000"] = true
		allowed["http://localhost:8080"] = true
	}
	return allowed
}

// registerPublicRoutes mendaftarkan endpoint yang boleh diakses tanpa autentikasi.
func registerPublicRoutes(r *gin.Engine, cfg *config.Config, h Handlers) {
	api := r.Group("/api")

	// Video
	api.GET("/videos", h.Video.List)
	api.GET("/videos/featured", h.Video.Featured)
	api.GET("/videos/suggest", h.VideoSuggest.Suggest)
	api.GET("/videos/:id", h.Video.Get)
	api.GET("/videos/:id/download", h.VideoDownload.Download)
	api.GET("/videos/:id/hls-status", h.UploadVideo.GetHLSStatus)
	api.GET("/categories", h.Video.ListCategories)
	api.GET("/genres", h.Video.ListGenres)

	// Banner & iklan
	api.GET("/banners", h.Banner.ListActive)
	api.GET("/ads/serve", h.Ad.Serve)

	// HLS. Endpoint publik untuk video published, tapi admin tetap boleh
	// mengakses preview video draft lewat OptionalAuth.
	hlsGroup := api.Group("/hls")
	hlsGroup.Use(middleware.OptionalAuth(cfg.JWTSecret))
	{
		hlsGroup.GET("/key/:id", h.HLS.ServeKey)
		hlsGroup.GET("/:id/index.m3u8", h.HLS.ServeM3U8)
		hlsGroup.GET("/:id/:segment", h.HLS.ServeSegment)
	}

	// Autentikasi
	api.POST("/admin/login", h.Auth.Login)
	api.POST("/admin/logout", h.Auth.Logout)

	// Favicon: tetap 204 agar browser tidak mencatat error di log.
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(http.StatusNoContent) })
}

// registerAdminRoutes mendaftarkan endpoint yang dilindungi AuthRequired.
func registerAdminRoutes(r *gin.Engine, cfg *config.Config, h Handlers) {
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(cfg.JWTSecret))

	// Upload
	admin.POST("/upload", h.Upload.Upload)
	admin.POST("/upload/video", h.UploadVideo.UploadVideo)
	admin.POST("/upload/process-hls", h.UploadVideo.ProcessHLS)

	// Upload chunked
	admin.POST("/upload/chunked/initiate", h.ChunkedUpload.InitiateChunkedUpload)
	admin.POST("/upload/chunked/:upload_id/chunk", h.ChunkedUpload.UploadChunk)
	admin.POST("/upload/chunked/:upload_id/complete", h.ChunkedUpload.CompleteChunkedUpload)
	admin.GET("/upload/chunked/:upload_id/progress", h.ChunkedUpload.GetUploadProgress)
	admin.DELETE("/upload/chunked/:upload_id", h.ChunkedUpload.AbortUpload)

	// Video CRUD
	admin.GET("/videos", h.Video.AdminList)
	admin.POST("/videos", h.Video.Create)
	admin.GET("/videos/:id", h.Video.AdminGet)
	admin.PUT("/videos/:id", h.Video.Update)
	admin.DELETE("/videos/:id", h.Video.Delete)

	admin.GET("/categories", h.Video.ListCategories)
	admin.GET("/genres", h.Video.ListGenres)

	// Banner CRUD
	admin.GET("/banners", h.Banner.AdminList)
	admin.POST("/banners", h.Banner.Create)
	admin.PUT("/banners/:id", h.Banner.Update)
	admin.DELETE("/banners/:id", h.Banner.Delete)

	// Iklan CRUD
	admin.GET("/ads", h.Ad.AdminList)
	admin.POST("/ads", h.Ad.Create)
	admin.PUT("/ads/:id", h.Ad.Update)
	admin.DELETE("/ads/:id", h.Ad.Delete)
}
