package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pgarcv/internal/hls"
	"pgarcv/internal/response"
	"pgarcv/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadHandler menerima file umum (gambar atau video) untuk thumbnail dan aset
// yang tidak melalui pipeline HLS.
type UploadHandler struct {
	uploadPath string
	cloudinary *storage.CloudinaryStorage
	processor  *hls.Processor
}

// NewUploadHandler membuat UploadHandler.
//
// processor dipakai untuk membuat thumbnail video; boleh nil bila seluruh
// upload ditangani Cloudinary.
func NewUploadHandler(uploadPath string, cloudinary *storage.CloudinaryStorage, processor *hls.Processor) *UploadHandler {
	return &UploadHandler{uploadPath: uploadPath, cloudinary: cloudinary, processor: processor}
}

// Upload menyimpan satu file dan mengembalikan URL publiknya.
//
// Penyimpanan lokal memakai nama acak (UUID) supaya nama asli dari klien tidak
// pernah menyentuh path filesystem.
func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File tidak ditemukan")
		return
	}
	if file.Size == 0 {
		response.Error(c, http.StatusBadRequest, "File kosong")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	kind, allowed := allowedUploadKind(ext)
	if !allowed {
		response.Error(c, http.StatusBadRequest, "Format file tidak didukung")
		return
	}
	if file.Size > maxUploadBytes {
		response.Error(c, http.StatusBadRequest, "File terlalu besar (maks 500MB)")
		return
	}

	// Ekstensi hanya dicek, bukan dipercaya: isi file harus cocok dengannya.
	// FileHeader.Open() membuka ulang file dari penyimpanan sementara Gin.
	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "File tidak bisa dibaca")
		return
	}
	header := make([]byte, 512)
	n, _ := f.Read(header)
	f.Close()
	if !validUploadContentType(kind, http.DetectContentType(header[:n])) {
		response.Error(c, http.StatusBadRequest, "Tipe konten file tidak sesuai")
		return
	}

	if h.cloudinary != nil {
		h.uploadToCloudinary(c, file)
		return
	}
	h.uploadToDisk(c, file, ext, kind)
}

// uploadToCloudinary meneruskan file ke Cloudinary dan mengembalikan URL-nya.
func (h *UploadHandler) uploadToCloudinary(c *gin.Context, file *multipart.FileHeader) {
	f, err := file.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuka file")
		return
	}
	defer f.Close()

	url, err := h.cloudinary.Upload(f, file.Filename)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal upload: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"url":           url,
		"filename":      file.Filename,
		"thumbnail_url": "",
	})
}

// uploadToDisk menyimpan file di direktori uploads lokal dan, untuk video,
// menghasilkan thumbnail tambahan secara best-effort.
func (h *UploadHandler) uploadToDisk(c *gin.Context, file *multipart.FileHeader, ext, kind string) {
	filename := uuid.New().String() + ext
	dst := filepath.Join(h.uploadPath, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan file")
		return
	}

	response.OK(c, gin.H{
		"url":           fmt.Sprintf("/uploads/%s", filename),
		"filename":      filename,
		"thumbnail_url": h.generateThumbnail(kind, dst),
	})
}

// generateThumbnail membuat thumbnail untuk file video lokal.
//
// Thumbnail bersifat opsional: kegagalan tidak menggagalkan upload.
func (h *UploadHandler) generateThumbnail(kind, videoPath string) string {
	if kind != "video" || h.processor == nil {
		return ""
	}

	for _, second := range []int{3, 1} {
		thumbPath, err := h.processor.ExtractThumbnail(videoPath, second)
		if err != nil {
			continue
		}
		thumbName := uuid.New().String() + ".jpg"
		thumbDst := filepath.Join(h.uploadPath, thumbName)
		if err := os.Rename(thumbPath, thumbDst); err != nil {
			continue
		}
		return fmt.Sprintf("/uploads/%s", thumbName)
	}
	return ""
}

// allowedUploadKind memetakan ekstensi ke jenis file yang diizinkan.
func allowedUploadKind(ext string) (string, bool) {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return "image", true
	case ".mp4", ".webm", ".mov", ".avi":
		return "video", true
	default:
		return "", false
	}
}

// validUploadContentType memastikan tipe konten terdeteksi sesuai jenis file
// yang diklaim oleh ekstensinya.
func validUploadContentType(kind, contentType string) bool {
	if kind == "image" {
		return strings.HasPrefix(contentType, "image/")
	}
	if kind == "video" {
		return strings.HasPrefix(contentType, "video/") || contentType == "application/octet-stream"
	}
	return false
}
