// Command server adalah entrypoint backend pgarcv.
//
// Tanggung jawabnya sengaja dibatasi: memuat konfigurasi, merakit dependensi,
// menjalankan server, dan menangani shutdown dengan rapi. Seluruh definisi
// route ada di internal/router.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pgarcv/internal/config"
	"pgarcv/internal/handler"
	"pgarcv/internal/hls"
	"pgarcv/internal/repository"
	"pgarcv/internal/router"
	"pgarcv/internal/service"
	"pgarcv/internal/storage"
)

// shutdownTimeout adalah lamanya server menunggu request in-flight selesai saat
// shutdown.
const shutdownTimeout = 15 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatalf("❌ Server berhenti: %v", err)
	}
}

// run merakit dan menjalankan aplikasi. Dipisah dari main agar alur startup
// bisa diuji tanpa memanggil os.Exit.
func run() error {
	cfg := config.Load()
	log.Printf("⚙️  Konfigurasi: %s", cfg)

	if err := ensureRuntimeDirs(cfg); err != nil {
		return err
	}

	db, err := repository.New(cfg.DBPath, cfg.DatabaseURL)
	if err != nil {
		return err
	}

	if err := repository.Seed(db, cfg.UploadPath); err != nil {
		// Seed gagal tidak mematikan server: database mungkin sudah berisi data
		// dari lingkungan sebelumnya.
		log.Printf("⚠️  Seed data dilewati: %v", err)
	}

	// Context ini dicancel saat menerima SIGINT/SIGTERM dan dipakai seluruh
	// background worker sebagai sinyal berhenti.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	processor := hls.NewProcessor(cfg.HLSOutputPath, cfg.HLSEncryptKey, cfg.ServerURL)
	hlsJob := service.NewHLSJob(ctx, db, processor)
	sessions := service.NewSessionStore()

	handlers := router.Handlers{
		Video:         handler.NewVideoHandler(db),
		Banner:        handler.NewBannerHandler(db),
		Ad:            handler.NewAdHandler(db),
		VideoSuggest:  handler.NewVideoSuggestHandler(db),
		VideoDownload: handler.NewVideoDownloadHandler(db, cfg.UploadPath),
		Auth:          handler.NewAuthHandler(cfg.AdminUsername, cfg.AdminPassword, cfg.JWTSecret, cfg.CookieSecure),
		Upload:        handler.NewUploadHandler(cfg.UploadPath, newCloudinary(cfg), processor),
		UploadVideo:   handler.NewUploadVideoHandler(db, cfg.UploadPath, processor, hlsJob),
		HLS:           handler.NewHLSHandler(db, cfg.HLSOutputPath, cfg.HLSEncryptKey, cfg.ServerURL),
		ChunkedUpload: handler.NewChunkedUploadHandler(cfg.UploadPath, sessions),
	}

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router.New(cfg, handlers),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Worker background dihentikan lebih dulu agar tidak menulis ke database
	// setelah server sudah tidak melayani request.
	defer hlsJob.Shutdown()
	defer sessions.Close()

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("🚀 Server jalan di http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
		log.Println("🛑 Shutdown diminta, menutup server...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Println("✅ Server berhenti dengan bersih")
	return nil
}

// ensureRuntimeDirs membuat direktori yang dibutuhkan aplikasi saat runtime.
func ensureRuntimeDirs(cfg *config.Config) error {
	for _, dir := range []string{cfg.UploadPath, cfg.HLSOutputPath} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// newCloudinary menyiapkan client Cloudinary bila CLOUDINARY_URL diisi.
// Storage bersifat opsional: tanpa URL, file disimpan lokal.
func newCloudinary(cfg *config.Config) *storage.CloudinaryStorage {
	if cfg.CloudinaryURL == "" {
		return nil
	}
	cld, err := storage.NewCloudinaryStorage(cfg.CloudinaryURL)
	if err != nil {
		log.Printf("⚠️  Cloudinary tidak aktif: %v", err)
		return nil
	}
	log.Println("✅ Cloudinary siap")
	return cld
}
