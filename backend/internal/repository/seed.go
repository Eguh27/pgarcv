package repository

import (
	"log"

	"pgarcv/internal/model"

	"gorm.io/gorm"
)

// Seed mengisi data contoh pada database yang masih kosong.
//
// Pemanggilan yang berulang aman: seeding dilewati begitu ada satu saja baris
// Video di tabel.
func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Video{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	videos := []model.Video{
		{
			Title:        "Video Pertama",
			Subtitle:     "Subtitle video ini",
			Description:  "Deskripsi lengkap video pertama yang menarik.",
			ThumbnailURL: "/uploads/placeholder.jpg",
			Duration:     300,
			Category:     "Umum",
			Genre:        "Edu",
			IsPublished:  true,
		},
		{
			Title:        "Tutorial Go Backend",
			Subtitle:     "Belajar REST API",
			Description:  "Tutorial lengkap membangun REST API dengan Go dan Gin framework.",
			ThumbnailURL: "/uploads/placeholder.jpg",
			Duration:     600,
			Category:     "Teknologi",
			Genre:        "Programming",
			IsPublished:  true,
		},
		{
			Title:        "Next.js Dasar",
			Subtitle:     "Frontend modern",
			Description:  "Pengenalan Next.js App Router untuk pemula.",
			ThumbnailURL: "/uploads/placeholder.jpg",
			Duration:     450,
			Category:     "Teknologi",
			Genre:        "Programming",
			IsPublished:  false,
		},
	}

	if err := db.Create(&videos).Error; err != nil {
		return err
	}

	banner := model.Banner{
		Title:     "Banner Utama",
		ImageURL:  "/uploads/placeholder.jpg",
		LinkURL:   "#",
		IsActive:  true,
		SortOrder: 1,
	}
	if err := db.Create(&banner).Error; err != nil {
		return err
	}

	ad := model.Ad{
		Name:       "Iklan Utama",
		AdCode:     `<div style="background:#f0f0f0;padding:20px;text-align:center">Slot Iklan</div>`,
		DeviceType: model.DeviceTypeAll,
		IsActive:   true,
	}
	if err := db.Create(&ad).Error; err != nil {
		return err
	}

	log.Println("✅ Seed data berhasil")
	return nil
}
