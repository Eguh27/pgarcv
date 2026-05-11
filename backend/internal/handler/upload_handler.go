package handler

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pgarcv/pkg/response"
	"pgarcv/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	uploadPath string
	cloudinary *storage.CloudinaryStorage
}

func NewUploadHandler(uploadPath string, cloudinary *storage.CloudinaryStorage) *UploadHandler {
	return &UploadHandler{uploadPath: uploadPath, cloudinary: cloudinary}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, "File tidak ditemukan")
		return
	}
	if file.Size == 0 {
		response.Error(c, 400, "File kosong")
		return
	}

	// Validasi ekstensi
	ext := strings.ToLower(filepath.Ext(file.Filename))
	kind, allowed := allowedUploadKind(ext)
	if !allowed {
		response.Error(c, 400, "Format file tidak didukung")
		return
	}

	// Batasi ukuran
	if file.Size > 500*1024*1024 { // 500MB
		response.Error(c, 400, "File terlalu besar (maks 500MB)")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.Error(c, 400, "File tidak bisa dibaca")
		return
	}
	defer src.Close()

	header := make([]byte, 512)
	n, _ := src.Read(header)
	contentType := http.DetectContentType(header[:n])
	if !validUploadContentType(kind, contentType) {
		response.Error(c, 400, "Tipe konten file tidak sesuai")
		return
	}

	var url string
	filename := file.Filename
	var dst string

	// Kalau ada Cloudinary, upload ke sana
	if h.cloudinary != nil {
		f, err := file.Open()
		if err != nil {
			response.Error(c, 500, "Gagal membuka file")
			return
		}
		defer f.Close()

		cloudUrl, err := h.cloudinary.Upload(f, file.Filename)
		if err != nil {
			response.Error(c, 500, "Gagal upload: "+err.Error())
			return
		}
		url = cloudUrl
	} else {
		// Fallback: simpan lokal
		filename = uuid.New().String() + filepath.Ext(file.Filename)
		dst = filepath.Join(h.uploadPath, filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			response.Error(c, 500, "Gagal menyimpan file")
			return
		}
		url = fmt.Sprintf("/uploads/%s", filename)
	}

	// Auto thumbnail for video upload (best-effort, hanya untuk lokal upload).
	thumbURL := ""
	if kind == "video" && h.cloudinary == nil && dst != "" {
		base := strings.TrimSuffix(filename, ext)
		thumbName := base + ".jpg"
		thumbPath := filepath.Join(h.uploadPath, thumbName)
		if genThumb(dst, thumbPath) == nil {
			if _, err := os.Stat(thumbPath); err == nil {
				thumbURL = fmt.Sprintf("/uploads/%s", thumbName)
			}
		}
	}

	response.OK(c, gin.H{"url": url, "filename": filename, "thumbnail_url": thumbURL})
}

func genThumb(videoPath, outputPath string) error {
	// 1) try 3s
	cmd := exec.Command("ffmpeg", "-y", "-hide_banner", "-loglevel", "error", "-ss", "00:00:03", "-i", videoPath, "-frames:v", "1", "-q:v", "2", outputPath)
	if err := cmd.Run(); err == nil {
		return nil
	}
	// 2) fallback 1s
	cmd2 := exec.Command("ffmpeg", "-y", "-hide_banner", "-loglevel", "error", "-ss", "00:00:01", "-i", videoPath, "-frames:v", "1", "-q:v", "2", outputPath)
	if err := cmd2.Run(); err == nil {
		return nil
	}
	return cmd2.Err
}

func allowedUploadKind(ext string) (string, bool) {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image", true
	case ".mp4", ".webm", ".mov":
		return "video", true
	default:
		return "", false
	}
}

func validUploadContentType(kind, contentType string) bool {
	if kind == "image" {
		return strings.HasPrefix(contentType, "image/")
	}
	if kind == "video" {
		return strings.HasPrefix(contentType, "video/") || contentType == "application/octet-stream"
	}
	return false
}
