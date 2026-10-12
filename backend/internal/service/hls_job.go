// Package service memegang pekerjaan yang berumur panjang dan memiliki
// lifecycle sendiri: transcode HLS di latar belakang dan penyimpanan sesi
// upload chunked. Keduanya tidak cocok dengan handler HTTP karena statusnya
// hidup melewati satu request.
package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"pgarcv/internal/hls"
	"pgarcv/internal/model"

	"gorm.io/gorm"
)

// DefaultMaxConcurrentTranscodes membatasi jumlah transcode ffmpeg yang boleh
// berjalan bersamaan. Tanpa batas ini, banyak upload paralel akan menjalankan
// ffmpeg tanpa restraint dan menguras CPU serta memori.
const DefaultMaxConcurrentTranscodes = 2

// HLSJob menjalankan transcode HLS di luar jalur request.
//
// Job menurunkan context milik sendiri dari baseCtx, sehingga Shutdown()
// membatalkan worker yang sedang berjalan sekaligus mencegah job baru mulai.
type HLSJob struct {
	db        *gorm.DB
	processor *hls.Processor

	ctx    context.Context
	cancel context.CancelFunc

	// sem membatasi jumlah worker transcode aktif.
	sem chan struct{}

	// wg melacak worker yang berjalan supaya Shutdown() bisa menunggu.
	wg sync.WaitGroup
}

// NewHLSJob membuat HLSJob dengan batas konkurensi bawaan.
func NewHLSJob(baseCtx context.Context, db *gorm.DB, processor *hls.Processor) *HLSJob {
	return NewHLSJobWithLimit(baseCtx, db, processor, DefaultMaxConcurrentTranscodes)
}

// NewHLSJobWithLimit membuat HLSJob dengan batas konkurensi eksplisit.
// Nilai <= 0 akan memakai DefaultMaxConcurrentTranscodes.
func NewHLSJobWithLimit(baseCtx context.Context, db *gorm.DB, processor *hls.Processor, maxConcurrent int) *HLSJob {
	if maxConcurrent < 1 {
		maxConcurrent = DefaultMaxConcurrentTranscodes
	}
	ctx, cancel := context.WithCancel(baseCtx)
	return &HLSJob{
		db:        db,
		processor: processor,
		ctx:       ctx,
		cancel:    cancel,
		sem:       make(chan struct{}, maxConcurrent),
	}
}

// Start menjadwalkan transcode untuk satu video dan langsung kembali.
//
// Handler tidak menunggu transcode selesai; progress dipantau lewat status di
// database (Video.HLSStatus).
func (j *HLSJob) Start(videoID uint, rawPath string) {
	j.wg.Add(1)
	go func() {
		defer j.wg.Done()

		// Ambil slot; kalau job sudah di-shutdown, jangan mulai pekerjaan baru.
		select {
		case j.sem <- struct{}{}:
			defer func() { <-j.sem }()
		case <-j.ctx.Done():
			j.setError(videoID, fmt.Errorf("server shutting down: %w", j.ctx.Err()))
			return
		}

		j.run(videoID, rawPath)
	}()
}

// run melakukan transcode lalu menulis hasilnya ke database.
//
// Setiap worker memasang recover supaya panic di dalam tidak menjatuhkan
// seluruh proses server.
func (j *HLSJob) run(videoID uint, rawPath string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ panic saat transcode video_id=%d: %v", videoID, r)
			j.setError(videoID, fmt.Errorf("panic: %v", r))
		}
	}()

	j.setStatus(videoID, model.HLSStatusProcessing)

	hlsDir, err := j.processor.ProcessHLS(videoID, rawPath)
	if err != nil {
		log.Printf("❌ Transcode HLS gagal video_id=%d: %v", videoID, err)
		j.setError(videoID, err)
		return
	}

	// Pastikan index.m3u8 benar-benar ada sebelum video ditandai selesai.
	m3u8Path := filepath.Join(hlsDir, "index.m3u8")
	if _, statErr := os.Stat(m3u8Path); statErr != nil {
		log.Printf("❌ index.m3u8 tidak ditemukan video_id=%d: %v", videoID, statErr)
		j.setError(videoID, fmt.Errorf("index.m3u8 tidak terbentuk: %w", statErr))
		return
	}

	// Path harus cocok dengan route di router: /api/hls/:id/index.m3u8
	m3u8URL := fmt.Sprintf("/api/hls/%d/index.m3u8", videoID)
	err = j.db.Model(&model.Video{}).Where("id = ?", videoID).Updates(map[string]any{
		"video_url":     m3u8URL,
		"hls_path":      hlsDir,
		"is_encrypted":  j.processor.EncryptKey != "",
		"hls_status":    model.HLSStatusDone,
		"hls_error_msg": "",
		// Simpan path raw supaya fitur download tetap bisa melayani file asli.
		"raw_video_path": rawPath,
	}).Error
	if err != nil {
		log.Printf("❌ Gagal menyimpan hasil HLS video_id=%d: %v", videoID, err)
		j.setError(videoID, err)
		return
	}

	log.Printf("✅ HLS selesai video_id=%d (disimpan di %s)", videoID, hlsDir)
}

// setStatus menulis status tanpa menyentuh pesan error.
func (j *HLSJob) setStatus(videoID uint, status string) {
	if err := j.db.Model(&model.Video{}).Where("id = ?", videoID).
		Update("hls_status", status).Error; err != nil {
		log.Printf("❌ Gagal set hls_status=%s video_id=%d: %v", status, videoID, err)
	}
}

// setError menandai video gagal dan menyimpan pesan errornya.
func (j *HLSJob) setError(videoID uint, cause error) {
	if cause == nil {
		return
	}
	err := j.db.Model(&model.Video{}).Where("id = ?", videoID).Updates(map[string]any{
		"hls_status":    model.HLSStatusError,
		"hls_error_msg": cause.Error(),
	}).Error
	if err != nil {
		log.Printf("❌ Gagal set hls_status=error video_id=%d: %v", videoID, err)
	}
}

// Shutdown membatalkan worker yang sedang berjalan lalu menunggunya selesai.
// Aman dipanggil lebih dari sekali.
func (j *HLSJob) Shutdown() {
	j.cancel()
	j.wg.Wait()
}
