// Package repository bertanggung jawab membuka koneksi database, menjalankan
// migrasi, dan mengisi data awal. Query spesifik per domain tetap dipegang
// handler lewat *gorm.DB yang di-inject.
package repository

import (
	"fmt"
	"log"

	"pgarcv/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New membuka koneksi ke database.
//
// databaseURL dipakai bila terisi (production, PostgreSQL); jika kosong
// sqlitePath yang dipakai (development, SQLite).
//
// Fungsi ini mengembalikan error, bukan memanggil log.Fatal, supaya dapat
// diuji dan Decisions oleh pemanggil.
func New(sqlitePath, databaseURL string) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	dialect := "SQLite"
	open := sqlite.Open(sqlitePath)
	if databaseURL != "" {
		dialect = "PostgreSQL"
		open = postgres.Open(databaseURL)
	}

	db, err := gorm.Open(open, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi database %s: %w", dialect, err)
	}

	if err := db.AutoMigrate(
		&model.Video{},
		&model.Category{},
		&model.Genre{},
		&model.Banner{},
		&model.Ad{},
	); err != nil {
		return nil, fmt.Errorf("gagal migrasi schema: %w", err)
	}

	log.Printf("✅ Database %s siap", dialect)
	return db, nil
}
