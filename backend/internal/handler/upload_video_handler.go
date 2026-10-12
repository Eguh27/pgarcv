package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pgarcv/internal/hls"
	"pgarcv/internal/model"
	"pgarcv/internal/response"
	"pgarcv/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// maxUploadBytes adalah batas ukuran file untuk seluruh jalur upload.
const maxUploadBytes = 500 * 1024 * 1024

// videoExtensions adalah ekstensi yang diterima untuk video.
var videoExtensions = map[string]bool{
	".mp4":  true,
	".webm": true,
	".mov":  true,
	".avi":  true,
}

// UploadVideoHandler menerima file video mentah dan memicu transcode HLS.
type UploadVideoHandler struct {
	db         *gorm.DB
	uploadPath string
	hlsJob     *service.HLSJob
	processor  *hls.Processor
}

// NewUploadVideoHandler membuat UploadVideoHandler.
func NewUploadVideoHandler(db *gorm.DB, uploadPath string, processor *hls.Processor, hlsJob *service.HLSJob) *UploadVideoHandler {
	return &UploadVideoHandler{db: db, uploadPath: uploadPath, processor: processor, hlsJob: hlsJob}
}

// UploadVideo menyimpan file video mentah, membuat thumbnail, dan mengembalikan
// jalur file untuk diproses HLS pada langkah berikutnya.
//
// Response sengaja tidak memuat url playable: video baru bisa diputar setelah
// transcode selesai dan Video.VideoURL terisi.
func (h *UploadVideoHandler) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File tidak ditemukan")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !videoExtensions[ext] {
		response.Error(c, http.StatusBadRequest, "File harus video (mp4/webm/mov/avi)")
		return
	}
	if file.Size > maxUploadBytes {
		response.Error(c, http.StatusBadRequest, "File terlalu besar (maks 500MB)")
		return
	}

	rawDir := filepath.Join(h.uploadPath, "raw")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyiapkan direktori upload")
		return
	}

	rawPath := filepath.Join(rawDir, uuid.New().String()+ext)
	if err := c.SaveUploadedFile(file, rawPath); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan file")
		return
	}

	duration, err := h.processor.GetDuration(rawPath)
	if err != nil {
		// Durasi hanya untuk choosing titik thumbnail; kegagalan tidak fatal.
		duration = 0
	}

	thumbURL := ""
	if thumbPath, err := h.processor.ExtractThumbnail(rawPath, thumbnailSecond(duration)); err == nil {
		if url, ok := h.publishThumbnail(thumbPath); ok {
			thumbURL = url
		}
	}

	response.OK(c, gin.H{
		"url":           "", // diisi setelah transcode HLS selesai
		"raw_path":      rawPath,
		"thumbnail_url": thumbURL,
		"duration":      duration,
		"message":       "Upload selesai. HLS akan diproses otomatis setelah video disimpan.",
	})
}

// thumbnailSecond memilih detik untuk diambilnya frame thumbnail: sepuluh
// persen dari durasi untuk video panjang, default 5 detik untuk video pendek.
//
// Titik ambil tidak boleh melewati durasi video — ffmpeg akan gagal seek dan
// thumbnail tidak pernah terbentuk. Video pendek (<5 detik) karena itu diambil
// dari tengah durasinya, bukan dari detik ke-5.
func thumbnailSecond(duration int) int {
	if duration <= 0 {
		return 0
	}
	second := 5
	if duration > 50 {
		second = duration / 10
	}
	if second >= duration {
		second = duration / 2
	}
	if second < 0 {
		second = 0
	}
	return second
}

// publishThumbnail memindahkan file thumbnail ke direktori uploads yang
// dilayani secara statis, lalu mengembalikan URL publiknya.
func (h *UploadVideoHandler) publishThumbnail(thumbPath string) (string, bool) {
	thumbName := uuid.New().String() + ".jpg"
	thumbDst := filepath.Join(h.uploadPath, thumbName)
	if err := os.Rename(thumbPath, thumbDst); err != nil {
		return "", false
	}
	return fmt.Sprintf("/uploads/%s", thumbName), true
}

// processHLSRequest adalah payload endpoint trigger transcode.
type processHLSRequest struct {
	VideoID uint   `json:"video_id"`
	RawPath string `json:"raw_path"`
}

// ProcessHLS menjadwalkan transcode HLS untuk video yang baru di-upload.
//
// Transcode berjalan di latar belakang; klien memantau Video.HLSStatus lewat
// endpoint hls-status.
func (h *UploadVideoHandler) ProcessHLS(c *gin.Context) {
	var req processHLSRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.VideoID == 0 || req.RawPath == "" {
		response.Error(c, http.StatusBadRequest, "video_id dan raw_path wajib diisi")
		return
	}

	var video model.Video
	if err := h.db.First(&video, req.VideoID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Video tidak ditemukan")
		return
	}

	h.hlsJob.Start(req.VideoID, req.RawPath)

	response.OK(c, gin.H{"message": "HLS processing dimulai", "video_id": req.VideoID})
}

// GetHLSStatus mengembalikan status transcode HLS untuk polling klien.
func (h *UploadVideoHandler) GetHLSStatus(c *gin.Context) {
	var video model.Video
	err := h.db.
		Select("id", "hls_status", "hls_error_msg", "video_url").
		First(&video, c.Param("id")).Error
	if err != nil {
		response.Error(c, http.StatusNotFound, "Video tidak ditemukan")
		return
	}

	response.OK(c, gin.H{
		"id":          video.ID,
		"hls_status":  video.HLSStatus,
		"hls_error":   video.HLSErrorMsg,
		"video_url":   video.VideoURL,
		"is_complete": video.HLSStatus == model.HLSStatusDone,
	})
}
