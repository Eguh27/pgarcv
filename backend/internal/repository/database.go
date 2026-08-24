package repository

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
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

	// Auto migrate semua tabel
	db.AutoMigrate(&Video{}, &Category{}, &Genre{}, &Banner{}, &Ad{})

	// Seed data awal jika kosong
	seedData(db)

	DB = db
	if databaseURL != "" {
		log.Println("✅ Database PostgreSQL siap")
	} else {
		log.Println("✅ Database SQLite siap:", dbPath)
	}
	return db
}

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&Video{}).Count(&count)
	if count > 0 {
		return
	}

	videos := []Video{
		{Title: "Video Pertama", Subtitle: "Subtitle video ini", Description: "Deskripsi lengkap video pertama yang menarik.", ThumbnailURL: "/uploads/placeholder.jpg", VideoURL: "", Duration: 300, Category: "Umum", Genre: "Edu", IsPublished: true},
		{Title: "Tutorial Go Backend", Subtitle: "Belajar REST API", Description: "Tutorial lengkap membangun REST API dengan Go dan Gin framework.", ThumbnailURL: "/uploads/placeholder.jpg", VideoURL: "", Duration: 600, Category: "Teknologi", Genre: "Programming", IsPublished: true},
		{Title: "Next.js Dasar", Subtitle: "Frontend modern", Description: "Pengenalan Next.js App Router untuk pemula.", ThumbnailURL: "/uploads/placeholder.jpg", VideoURL: "", Duration: 450, Category: "Teknologi", Genre: "Programming", IsPublished: false},
	}
	db.Create(&videos)

	db.Create(&Banner{Title: "Banner Utama", ImageURL: "/uploads/placeholder.jpg", LinkURL: "#", IsActive: true, SortOrder: 1})
	db.Create(&Ad{Name: "Iklan Utama", AdCode: `<div style="background:#f0f0f0;padding:20px;text-align:center">Slot Iklan</div>`, DeviceType: "all", IsActive: true})

	log.Println("✅ Seed data berhasil")
}
