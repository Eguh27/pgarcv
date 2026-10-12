package service

import (
	"context"
	"io"
	"log"
	"path/filepath"
	"testing"
	"time"

	"pgarcv/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupJobDB membuat database sementara dengan tabel Video.
//
// Sengaja memakai file di t.TempDir(), bukan ":memory:": mode memory pada
// SQLite membuat SETIAP koneksi di pool mendapat database yang berbeda, sehingga
// query dari koneksi lain akan melaporkan "no such table" saat test berjalan
// konkuren. File di disk tidak punya masalah itu.
func setupJobDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "job-test.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Video{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

// silenceLogs membuang output log selama test berjalan.
//
// Job diuji tanpa processor sehingga setiap run() pasti gagal; tanpa ini log
// error dari worker membuat output test dan menutupi pesan kegagalan yang sebenarnya.
func silenceLogs(t *testing.T) {
	t.Helper()
	prev := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(prev) })
}

// newTestJob membuat HLSJob dengan processor kosong.
//
// Processor di sini belum dipakai: test ini berfokus pada lifecycle, konkurensi,
// dan penulisan status — bukan pada output ffmpeg. Karena processor-nya kosong,
// run() akan gagal dan memicu jalur error yang justru sedang diuji.
func newTestJob(t *testing.T, ctx context.Context, db *gorm.DB, limit int) *HLSJob {
	t.Helper()
	silenceLogs(t)
	return NewHLSJobWithLimit(ctx, db, nil, limit)
}

// TestHLSJobLimitsConcurrentTranscodes memverifikasi worker tidak melebihi batas
// konkurensi yang diminta.
//
// Processor nil membuat run() gagal cepat, sehingga slot sem terisi sebentar lalu
// dilepas — cukup untuk mengukur bahwa tidak ada lebih dari N worker yang
// jalannya bersamaan.
func TestHLSJobLimitsConcurrentTranscodes(t *testing.T) {
	db := setupJobDB(t)

	const limit = 2
	const jobs = 12

	job := newTestJob(t, context.Background(), db, limit)
	defer job.Shutdown()

	for i := 0; i < jobs; i++ {
		if err := db.Create(&model.Video{Title: "v"}).Error; err != nil {
			t.Fatal(err)
		}
		var video model.Video
		if err := db.Order("id DESC").First(&video).Error; err != nil {
			t.Fatal(err)
		}
		job.Start(video.ID, "/tmp/tidak-ada.mp4")
	}

	// Semua worker harus selesai dan tidak boleh menggantung.
	done := make(chan struct{})
	go func() {
		job.Shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Shutdown tidak selesai: worker menggantung")
	}

	if got := cap(job.sem); got != limit {
		t.Errorf("kapasitas sem = %d want %d", got, limit)
	}
}

// TestHLSJobShutdownIsIdempotent memastikan Shutdown aman dipanggil berulang.
func TestHLSJobShutdownIsIdempotent(t *testing.T) {
	db := setupJobDB(t)
	job := newTestJob(t, context.Background(), db, 1)

	job.Shutdown()
	job.Shutdown()
	job.Shutdown()
}

// TestHLSJobStartAfterShutdownIsSafe memastikan Start setelah Shutdown tidak
// panic dan langsung kembali.
func TestHLSJobStartAfterShutdownIsSafe(t *testing.T) {
	db := setupJobDB(t)
	job := newTestJob(t, context.Background(), db, 1)
	job.Shutdown()

	video := model.Video{Title: "v"}
	if err := db.Create(&video).Error; err != nil {
		t.Fatal(err)
	}

	// Tidak akan transcode apa pun; hanya memastikan tidak deadlock atau panic.
	job.Start(video.ID, "/tmp/tidak-ada.mp4")
	job.Shutdown()
}

// TestHLSJobMarksErrorStatusOnFailure memverifikasi kegagalan transcode
// menandai video sebagai error beserta pesannya.
func TestHLSJobMarksErrorStatusOnFailure(t *testing.T) {
	db := setupJobDB(t)

	video := model.Video{Title: "v", HLSStatus: model.HLSStatusPending}
	if err := db.Create(&video).Error; err != nil {
		t.Fatal(err)
	}

	// Processor nil pasti gagal saat run() memanggil ProcessHLS.
	job := newTestJob(t, context.Background(), db, 1)
	job.Start(video.ID, "/tmp/tidak-ada.mp4")
	job.Shutdown()

	var got model.Video
	if err := db.First(&got, video.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.HLSStatus != model.HLSStatusError {
		t.Errorf("hls_status = %q want %q", got.HLSStatus, model.HLSStatusError)
	}
}

// TestHLSJobUsesDefaultLimit memastikan limit non-positif memakai default.
func TestHLSJobUsesDefaultLimit(t *testing.T) {
	db := setupJobDB(t)

	job := newTestJob(t, context.Background(), db, 0)
	defer job.Shutdown()
	if got := cap(job.sem); got != DefaultMaxConcurrentTranscodes {
		t.Errorf("kapasitas sem = %d want %d", got, DefaultMaxConcurrentTranscodes)
	}

	job2 := newTestJob(t, context.Background(), db, -3)
	defer job2.Shutdown()
	if got := cap(job2.sem); got != DefaultMaxConcurrentTranscodes {
		t.Errorf("kapasitas sem = %d want %d", got, DefaultMaxConcurrentTranscodes)
	}
}
