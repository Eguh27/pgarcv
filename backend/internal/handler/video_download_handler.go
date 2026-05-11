package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pgarcv/internal/repository"
	"pgarcv/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoDownloadHandler struct {
	db         *gorm.DB
	uploadPath string
}

func NewVideoDownloadHandler(db *gorm.DB, uploadPath string) *VideoDownloadHandler {
	return &VideoDownloadHandler{db: db, uploadPath: uploadPath}
}

type videoDownloadRequest struct {
	VideoURL string `json:"-"`
}

func (h *VideoDownloadHandler) Download(c *gin.Context) {
	var video repository.Video
	if err := h.db.First(&video, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}

	// Public guard
	if !video.IsPublished {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	if !video.AllowDownload {
		response.Error(c, 403, "Video tidak bisa diunduh")
		return
	}

	if video.VideoURL == "" {
		response.Error(c, 404, "Video file tidak tersedia")
		return
	}

	srcURL := video.VideoURL
	// HLS video — cek raw_video_path dulu
	if strings.HasPrefix(srcURL, "/api/hls/") {
		if video.RawVideoPath != "" {
			if _, err := os.Stat(video.RawVideoPath); err == nil {
				downloadName := sanitizeFilename(video.Title) + ".mp4"
				c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
				c.Header("Content-Type", "video/mp4")
				c.File(video.RawVideoPath)
				return
			}
		}
		response.Error(c, 403, "Video streaming tidak dapat diunduh langsung")
		return
	}

	// Local file: /uploads/<filename>
	if strings.HasPrefix(srcURL, "/uploads/") {
		filename := strings.TrimPrefix(srcURL, "/uploads/")
		filePath := filepath.Join(h.uploadPath, filename)
		f, err := os.Open(filePath)
		if err != nil {
			response.Error(c, 404, "File video tidak ditemukan")
			return
		}
		defer f.Close()

		// Basic headers
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename="+filepath.Base(filename))
		c.Status(http.StatusOK)
		_, _ = io.Copy(c.Writer, f)
		return
	}

	// Absolute URL: proxy stream
	if strings.HasPrefix(srcURL, "http://") || strings.HasPrefix(srcURL, "https://") {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(srcURL)
		if err != nil {
			response.Error(c, 500, "Gagal mengunduh video")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			response.Error(c, resp.StatusCode, "Gagal mengunduh video")
			return
		}

		ct := resp.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/octet-stream"
		}
		c.Header("Content-Type", ct)
		c.Header("Content-Disposition", "attachment")
		c.Status(http.StatusOK)
		_, _ = io.Copy(c.Writer, resp.Body)
		return
	}

	response.Error(c, 400, "URL video tidak valid")
}

func sanitizeFilename(name string) string {
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == ' ' {
			result += string(r)
		}
	}
	if result == "" {
		return "video"
	}
	return result
}
