package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

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

func Load() *Config {
	godotenv.Load()

	jwtSecret := getEnv("JWT_SECRET", "")
	adminPassword := getEnv("ADMIN_PASSWORD", "")
	env := getEnv("ENV", "development")

	// Fail-fast: wajib set di production
	if env == "production" {
		if jwtSecret == "" {
			log.Fatal("❌ JWT_SECRET wajib diset di production")
		}
		if adminPassword == "" {
			log.Fatal("❌ ADMIN_PASSWORD wajib diset di production")
		}
		if len(jwtSecret) < 32 {
			log.Fatal("❌ JWT_SECRET harus minimal 32 karakter")
		}
	}

	// Development fallback (hanya untuk lokal)
	if jwtSecret == "" {
		jwtSecret = "dev-only-secret-tidak-untuk-production-32char"
		log.Println("⚠️  JWT_SECRET tidak diset, pakai default DEV")
	}
	if adminPassword == "" {
		adminPassword = "admin123"
		log.Println("⚠️  ADMIN_PASSWORD tidak diset, pakai default DEV")
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		JWTSecret:     jwtSecret,
		DBPath:        getEnv("DB_PATH", "./videoplatform.db"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: adminPassword,
		CookieSecure:  getEnv("COOKIE_SECURE", "false") == "true",
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:3000"),
		CloudinaryURL: getEnv("CLOUDINARY_URL", ""),
		HLSEncryptKey: getEnv("HLS_ENCRYPT_KEY", ""),
		HLSOutputPath: getEnv("HLS_OUTPUT_PATH", "./hls_output"),
		ServerURL:     getEnv("SERVER_URL", "http://localhost:8080"),
		Env:           env,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
