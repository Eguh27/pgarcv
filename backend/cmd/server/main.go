package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// "path/filepath"

	"pgarcv/config"
	"pgarcv/internal/handler"
	"pgarcv/internal/middleware"
	"pgarcv/internal/repository"
	"pgarcv/pkg/hls"
	"pgarcv/pkg/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Pastikan folder uploads ada
	os.MkdirAll(cfg.UploadPath, 0755)
	os.MkdirAll(cfg.HLSOutputPath, 0755)

	// Init database
	db := repository.InitDB(cfg.DBPath, cfg.DatabaseURL)

	// Init HLS processor
	hlsProcessor := hls.NewProcessor(cfg.HLSOutputPath, cfg.HLSEncryptKey, cfg.ServerURL)

	// Init handlers
	videoH := handler.NewVideoHandler(db)
	bannerH := handler.NewBannerHandler(db)
	adH := handler.NewAdHandler(db)
	videoSuggestH := handler.NewVideoSuggestHandler(db)
	videoDownloadH := handler.NewVideoDownloadHandler(db, cfg.UploadPath)
	authH := handler.NewAuthHandler(cfg.AdminUsername, cfg.AdminPassword, cfg.JWTSecret, cfg.CookieSecure)

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

	uploadH := handler.NewUploadHandler(cfg.UploadPath, cldStorage)
	uploadVideoH := handler.NewUploadVideoHandler(db, cfg.UploadPath, hlsProcessor)
	hlsH := handler.NewHLSHandler(db, cfg.HLSOutputPath, cfg.HLSEncryptKey, cfg.ServerURL)
	chunkedUploadH := handler.NewChunkedUploadHandler(cfg.UploadPath)

	// Setup Gin
	r := gin.Default()
	r.MaxMultipartMemory = 64 << 20

	// CORS — whitelist eksplisit via CORS_ORIGIN (pisahkan dengan koma).
	// Contoh: CORS_ORIGIN=http://localhost:3000,https://pgarcv.vercel.app
	allowedOrigins := make(map[string]bool)
	for _, o := range strings.Split(cfg.CORSOrigin, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowedOrigins[o] = true
		}
	}
	if cfg.Env != "production" {
		allowedOrigins["http://localhost:3000"] = true
		allowedOrigins["http://localhost:8080"] = true
	}

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return allowedOrigins[origin]
		},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"ngrok-skip-browser-warning",
		},
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
		api.GET("/videos/suggest", videoSuggestH.Suggest)
		api.GET("/videos/:id/download", videoDownloadH.Download)
		api.GET("/categories", videoH.ListCategories)
		api.GET("/genres", videoH.ListGenres)

		// Banners & Ads
		api.GET("/banners", bannerH.ListActive)
		api.GET("/ads/serve", adH.Serve)

		// HLS (with optional auth for admin preview)
		hls := api.Group("/hls")
		hls.Use(middleware.OptionalAuth(cfg.JWTSecret))
		{
			hls.GET("/key/:id", hlsH.ServeKey)
			hls.GET("/:id/index.m3u8", hlsH.ServeM3U8)
			hls.GET("/:id/:segment", hlsH.ServeSegment)
		}

		// ✅ HLS status endpoint (for polling)
		api.GET("/videos/:id/hls-status", uploadVideoH.GetHLSStatus)

		// Auth
		api.POST("/admin/login", authH.Login)
		api.POST("/admin/logout", authH.Logout)
	}

	// Favicon (public, no auth required)
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	// === ADMIN ROUTES (protected) ===
	admin := r.Group("/api/admin")
	admin.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		// Upload (traditional single-file)
		admin.POST("/upload", uploadH.Upload)
		admin.POST("/upload/video", uploadVideoH.UploadVideo)
		admin.POST("/upload/process-hls", uploadVideoH.ProcessHLS)

		// ✅ Chunked upload endpoints
		admin.POST("/upload/chunked/initiate", chunkedUploadH.InitiateChunkedUpload)
		admin.POST("/upload/chunked/:upload_id/chunk", chunkedUploadH.UploadChunk)
		admin.POST("/upload/chunked/:upload_id/complete", chunkedUploadH.CompleteChunkedUpload)
		admin.GET("/upload/chunked/:upload_id/progress", chunkedUploadH.GetUploadProgress)
		admin.DELETE("/upload/chunked/:upload_id", chunkedUploadH.AbortUpload)

		// Videos CRUD
		admin.GET("/videos", videoH.AdminList)
		admin.GET("/videos/:id", videoH.AdminGet)
		admin.POST("/videos", videoH.Create)
		admin.PUT("/videos/:id", videoH.Update)
		admin.DELETE("/videos/:id", videoH.Delete)

		admin.GET("/categories", videoH.ListCategories)
		admin.GET("/genres", videoH.ListGenres)

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
	}

	log.Printf("🚀 Server jalan di http://localhost:%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
