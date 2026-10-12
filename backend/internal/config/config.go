// Package config membaca dan memvalidasi konfigurasi aplikasi dari environment.
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// EnvProduction adalah nilai ENV yang mengaktifkan mode production.
const EnvProduction = "production"

// minJWTSecretLength adalah panjang minimum JWT_SECRET di production.
const minJWTSecretLength = 32

// Config berisi seluruh konfigurasi runtime yang dibutuhkan aplikasi.
type Config struct {
	Port          string
	JWTSecret     string
	DBPath        string
	DatabaseURL   string
	UploadPath    string
	AdminUsername string
	AdminPassword string
	CookieSecure  bool
	CORSOrigin    string
	CloudinaryURL string
	HLSEncryptKey string
	HLSOutputPath string
	ServerURL     string
	Env           string
}

// IsProduction melapor apakah aplikasi berjalan dalam mode production.
func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

// Addr mengembalikan alamat listen dalam format ":port".
func (c *Config) Addr() string { return ":" + c.Port }

// Load membaca konfigurasi dari environment (dengan .env bila ada).
//
// Di production, JWT_SECRET dan ADMIN_PASSWORD wajib diisi dan JWT_SECRET harus
// cukup panjang; kegagalan adalah kondisi fatal sehingga proses dihentikan di
// sini. Di luar production, nilai default dipakai agar server tetap bisa
// dijalankan lokal, disertai peringatan di log.
func Load() *Config {
	// Kegagalan membaca .env bukan error fatal: environment bisa disuplai
	// langsung oleh container atau system service.
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️  File .env tidak ditemukan, memakai environment langsung")
	}

	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		DBPath:        getEnv("DB_PATH", "./videoplatform.db"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		CookieSecure:  getEnv("COOKIE_SECURE", "false") == "true",
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:3000"),
		CloudinaryURL: os.Getenv("CLOUDINARY_URL"),
		HLSEncryptKey: os.Getenv("HLS_ENCRYPT_KEY"),
		HLSOutputPath: getEnv("HLS_OUTPUT_PATH", "./hls_output"),
		ServerURL:     getEnv("SERVER_URL", "http://localhost:8080"),
		Env:           getEnv("ENV", "development"),
	}

	if cfg.IsProduction() {
		validateProduction(cfg)
		return cfg
	}

	applyDevelopmentDefaults(cfg)
	return cfg
}

// validateProduction menghentikan proses bila konfigurasi production tidak aman.
func validateProduction(cfg *Config) {
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		log.Fatal("❌ JWT_SECRET wajib diset di production")
	}
	if len(cfg.JWTSecret) < minJWTSecretLength {
		log.Fatalf("❌ JWT_SECRET harus minimal %d karakter", minJWTSecretLength)
	}
	if strings.TrimSpace(cfg.AdminPassword) == "" {
		log.Fatal("❌ ADMIN_PASSWORD wajib diset di production")
	}
	if cfg.AdminUsername == "admin" {
		log.Println("⚠️  ADMIN_USERNAME masih memakai nilai default 'admin'")
	}
}

// applyDevelopmentDefaults mengisi nilai development dan memperingatkan
// pemakaian kredensial default.
func applyDevelopmentDefaults(cfg *Config) {
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = devJWTSecret
		log.Println("⚠️  JWT_SECRET tidak diset, memakai default DEV")
	}
	if cfg.AdminPassword == "" {
		cfg.AdminPassword = devAdminPassword
		log.Println("⚠️  ADMIN_PASSWORD tidak diset, memakai default DEV")
	}
}

// Nilai fallback khusus pengembangan lokal. Tidak boleh dipakai di production.
const (
	devJWTSecret     = "dev-only-secret-tidak-untuk-production-32char"
	devAdminPassword = "admin123"
)

// getEnv mengembalikan nilai environment, atau fallback bila kosong.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// String mengembalikan ringkasan konfigurasi yang aman untuk di-log. Nilai
// rahasia tidak pernah ikut dicetak.
func (c *Config) String() string {
	dialect := "SQLite"
	if c.DatabaseURL != "" {
		dialect = "PostgreSQL"
	}
	return fmt.Sprintf(
		"env=%s port=%s db=%s cors=%q hls_encrypted=%t cloudinary=%t",
		c.Env, c.Port, dialect, c.CORSOrigin, c.HLSEncryptKey != "", c.CloudinaryURL != "",
	)
}
