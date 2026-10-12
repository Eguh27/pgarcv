package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pgarcv/internal/model"
	"pgarcv/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// downloadProxyTimeout membatasi durasi pengambilan file dari origin eksternal.
const downloadProxyTimeout = 30 * time.Second

// VideoDownloadHandler menyajikan file video untuk diunduh.
type VideoDownloadHandler struct {
	db         *gorm.DB
	uploadPath string
}

// NewVideoDownloadHandler membuat VideoDownloadHandler.
func NewVideoDownloadHandler(db *gorm.DB, uploadPath string) *VideoDownloadHandler {
	return &VideoDownloadHandler{db: db, uploadPath: uploadPath}
}

// Download streamed video sebagai attachment.
//
// Endpoint ini publik, jadi urutan penjagaannya penting: video harus published
// dan harus mengizinkan unduhan sebelum sumber file disentuh.
func (h *VideoDownloadHandler) Download(c *gin.Context) {
	var video model.Video
	if err := h.db.First(&video, c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Video tidak ditemukan")
		return
	}

	// Sembunyikan keberadaan video draft di depan pengguna publik.
	if !video.IsPublished {
		response.Error(c, http.StatusNotFound, "Video tidak ditemukan")
		return
	}
	if !video.AllowDownload {
		response.Error(c, http.StatusForbidden, "Video tidak bisa diunduh")
		return
	}
	if video.VideoURL == "" {
		response.Error(c, http.StatusNotFound, "Video file tidak tersedia")
		return
	}

	switch {
	case strings.HasPrefix(video.VideoURL, "/api/hls/"):
		h.downloadRaw(c, &video)
	case strings.HasPrefix(video.VideoURL, "/uploads/"):
		h.downloadLocal(c, &video)
	case strings.HasPrefix(video.VideoURL, "http://"), strings.HasPrefix(video.VideoURL, "https://"):
		h.downloadRemote(c, &video)
	default:
		response.Error(c, http.StatusBadRequest, "URL video tidak valid")
	}
}

// downloadRaw melayani video yang diputar lewat HLS memakai file sumber asli.
func (h *VideoDownloadHandler) downloadRaw(c *gin.Context, video *model.Video) {
	if video.RawVideoPath == "" {
		response.Error(c, http.StatusForbidden, "Video streaming tidak dapat diunduh langsung")
		return
	}
	if _, err := os.Stat(video.RawVideoPath); err != nil {
		response.Error(c, http.StatusForbidden, "Video streaming tidak dapat diunduh langsung")
		return
	}

	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, sanitizeFilename(video.Title)+".mp4"))
	c.File(video.RawVideoPath)
}

// downloadLocal melayani file yang tersimpan di direktori uploads.
func (h *VideoDownloadHandler) downloadLocal(c *gin.Context, video *model.Video) {
	filename := strings.TrimPrefix(video.VideoURL, "/uploads/")

	// Guard traversal. Video.URL dikelola admin, tapi tetap tidak boleh
	// diizinkan keluar dari uploadPath hanya karena nilainya string bebas.
	if filename == "" || strings.Contains(filename, "/") || strings.Contains(filename, `\`) ||
		strings.HasPrefix(filename, ".") || filepath.Base(filename) != filename {
		response.Error(c, http.StatusBadRequest, "URL video tidak valid")
		return
	}

	filePath := filepath.Join(h.uploadPath, filename)
	f, err := os.Open(filePath)
	if err != nil {
		response.Error(c, http.StatusNotFound, "File video tidak ditemukan")
		return
	}
	defer f.Close()

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, f)
}

// downloadRelay meneruskan file dari origin eksternal sebagai attachment.
func (h *VideoDownloadHandler) downloadRemote(c *gin.Context, video *model.Video) {
	client := &http.Client{Timeout: downloadProxyTimeout}
	resp, err := client.Get(video.VideoURL)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "Gagal mengunduh video")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		response.Error(c, http.StatusBadGateway, "Gagal mengunduh video")
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, resp.Body)
}

// sanitizeFilename membuat nama file aman dari judul video: hanya alphanumeric,
// spasi, tanda hubung, dan underscore yang dipertahankan.
func sanitizeFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == ' ':
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "video"
	}
	return b.String()
}
