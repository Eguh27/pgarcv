package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pgarcv/internal/repository"
	"pgarcv/pkg/hls"
	"pgarcv/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadVideoHandler struct {
	db         *gorm.DB
	uploadPath string
	hlsProc    *hls.Processor
}

func NewUploadVideoHandler(db *gorm.DB, uploadPath string, hlsProc *hls.Processor) *UploadVideoHandler {
	return &UploadVideoHandler{db: db, uploadPath: uploadPath, hlsProc: hlsProc}
}

func (h *UploadVideoHandler) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, "File tidak ditemukan")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".mp4" && ext != ".webm" && ext != ".mov" && ext != ".avi" {
		response.Error(c, 400, "File harus video (mp4/webm/mov/avi)")
		return
	}
	if file.Size > 500*1024*1024 {
		response.Error(c, 400, "File terlalu besar (maks 500MB)")
		return
	}

	rawDir := filepath.Join(h.uploadPath, "raw")
	os.MkdirAll(rawDir, 0755)
	filename := uuid.New().String() + ext
	rawPath := filepath.Join(rawDir, filename)
	if err := c.SaveUploadedFile(file, rawPath); err != nil {
		response.Error(c, 500, "Gagal simpan file")
		return
	}

	duration, _ := h.hlsProc.GetDuration(rawPath)
	thumbSecond := 5
	if duration > 50 {
		thumbSecond = duration / 10
	}

	thumbURL := ""
	if thumbPath, err := h.hlsProc.ExtractThumbnail(rawPath, thumbSecond); err == nil {
		thumbFilename := uuid.New().String() + ".jpg"
		thumbDst := filepath.Join(h.uploadPath, thumbFilename)
		if os.Rename(thumbPath, thumbDst) == nil {
			thumbURL = "/uploads/" + thumbFilename
		}
	}

	response.OK(c, gin.H{
		"url":           "", // Will be set after HLS processing
		"raw_path":      rawPath,
		"thumbnail_url": thumbURL,
		"duration":      duration,
		"message":       "Upload selesai. HLS akan diproses otomatis setelah video disimpan.",
	})
}

// ProcessHLSAsync triggers HLS processing in background and updates status
func (h *UploadVideoHandler) ProcessHLSAsync(videoID uint, rawPath string) {
	go func(rawPath string, videoID uint) {
		errPrefix := fmt.Sprintf("[upload-video process-hls] video_id=%d raw=%s: ", videoID, rawPath)

		// Update status: processing
		h.db.Model(&repository.Video{}).Where("id = ?", videoID).Update("hls_status", "processing")

		hlsDir, err := h.hlsProc.ProcessHLS(videoID, rawPath)
		if err != nil {
			fmt.Printf("❌ HLS gagal %s%v\n", errPrefix, err)
			// Update status: error
			h.db.Model(&repository.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
				"hls_status":    "error",
				"hls_error_msg": err.Error(),
			})
			return
		}

		// Pastikan output index.m3u8 ada sebelum raw dihapus
		m3u8Path := filepath.Join(hlsDir, "index.m3u8")
		if _, statErr := os.Stat(m3u8Path); statErr != nil {
			fmt.Printf("❌ HLS output tidak ditemukan %s (stat error=%v), raw tidak dihapus\n", m3u8Path, statErr)
			h.db.Model(&repository.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
				"hls_status":    "error",
				"hls_error_msg": fmt.Sprintf("M3U8 file not created: %v", statErr),
			})
			return
		}

		// Harus match route server: /api/hls/:id/index.m3u8
		m3u8URL := fmt.Sprintf("/api/hls/%d/index.m3u8", videoID)
		if upErr := h.db.Model(&repository.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
			"video_url":     m3u8URL,
			"hls_path":      hlsDir,
			"is_encrypted":  h.hlsProc.EncryptKey != "",
			"hls_status":    "done",
			"hls_error_msg": "", // Clear error
		}).Error; upErr != nil {
			fmt.Printf("❌ Gagal update video setelah HLS %s%v, raw tidak dihapus\n", errPrefix, upErr)
			h.db.Model(&repository.Video{}).Where("id = ?", videoID).Updates(map[string]interface{}{
				"hls_status":    "error",
				"hls_error_msg": upErr.Error(),
			})
			return
		}

		// Simpan raw_video_path di DB supaya download masih bisa
		h.db.Model(&repository.Video{}).Where("id = ?", videoID).
			Update("raw_video_path", rawPath)
		fmt.Printf("✅ HLS selesai video %d (raw disimpan di %s)\n",
			videoID, rawPath)
	}(rawPath, videoID)
}

func (h *UploadVideoHandler) ProcessHLS(c *gin.Context) {
	var req struct {
		VideoID uint   `json:"video_id"`
		RawPath string `json:"raw_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.VideoID == 0 {
		response.Error(c, 400, "video_id dan raw_path wajib diisi")
		return
	}

	var video repository.Video
	if err := h.db.First(&video, req.VideoID).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}

	// ✅ Trigger async HLS processing with status tracking
	h.ProcessHLSAsync(req.VideoID, req.RawPath)

	response.OK(c, gin.H{"message": "HLS processing dimulai", "video_id": req.VideoID})
}

// GetHLSStatus returns the current HLS processing status
func (h *UploadVideoHandler) GetHLSStatus(c *gin.Context) {
	videoID := c.Param("id")
	var video repository.Video
	if err := h.db.Select("id", "hls_status", "hls_error_msg", "video_url").First(&video, videoID).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	response.OK(c, gin.H{
		"id":          video.ID,
		"hls_status":  video.HLSStatus,
		"hls_error":   video.HLSErrorMsg,
		"video_url":   video.VideoURL,
		"is_complete": video.HLSStatus == "done",
	})
}
