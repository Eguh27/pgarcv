package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"pgarcv/internal/response"
	"pgarcv/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Batas upload chunked.
const (
	// maxChunkedChunks membatasi jumlah chunk per upload.
	maxChunkedChunks = 500
	// recommendedChunkSize adalah ukuran chunk yang disarankan ke klien.
	recommendedChunkSize = 5 * 1024 * 1024
)

// ChunkedUploadHandler melayani upload video besar melalui potongan kecil.
//
// State sesi disimpan di service.SessionStore, bukan di dalam handler, supaya
// umurnya dikelola terpisah dari satu request dan direktori chunk yang
// tertinggal ikut dibersihkan.
type ChunkedUploadHandler struct {
	uploadPath string
	sessions   *service.SessionStore
}

// NewChunkedUploadHandler membuat ChunkedUploadHandler.
//
// Caller bertanggung jawab menutup sessions saat server shutdown.
func NewChunkedUploadHandler(uploadPath string, sessions *service.SessionStore) *ChunkedUploadHandler {
	return &ChunkedUploadHandler{uploadPath: uploadPath, sessions: sessions}
}

// initiateChunkedRequest adalah payload awal sesi upload.
type initiateChunkedRequest struct {
	Filename    string `json:"filename" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
	TotalChunks int    `json:"total_chunks" binding:"required"`
}

// InitiateChunkedUpload membuka sesi upload baru dan mengembalikan upload_id.
func (h *ChunkedUploadHandler) InitiateChunkedUpload(c *gin.Context) {
	var req initiateChunkedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	if req.FileSize <= 0 || req.FileSize > maxUploadBytes {
		response.Error(c, http.StatusBadRequest, "File terlalu besar (maks 500MB)")
		return
	}
	if req.TotalChunks < 1 || req.TotalChunks > maxChunkedChunks {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("Total chunks harus 1-%d", maxChunkedChunks))
		return
	}

	uploadID := uuid.New().String()
	if _, err := h.sessions.Put(uploadID, req.Filename, req.TotalChunks, req.FileSize, h.uploadPath); err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyiapkan sesi upload")
		return
	}

	response.Created(c, gin.H{
		"upload_id":    uploadID,
		"total_chunks": req.TotalChunks,
		"chunk_size":   recommendedChunkSize,
	})
}

// UploadChunk menyimpan satu potongan file.
func (h *ChunkedUploadHandler) UploadChunk(c *gin.Context) {
	session, ok := h.session(c)
	if !ok {
		return
	}

	chunkIndex, err := strconv.Atoi(c.PostForm("chunk_index"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid chunk_index")
		return
	}
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		response.Error(c, http.StatusBadRequest, "Chunk index out of range")
		return
	}

	file, err := c.FormFile("chunk")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Chunk file required")
		return
	}

	chunkPath := filepath.Join(session.ChunksDir, fmt.Sprintf("chunk_%d", chunkIndex))
	if err := c.SaveUploadedFile(file, chunkPath); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to save chunk")
		return
	}

	received := session.MarkReceived(chunkIndex)
	response.OK(c, gin.H{
		"chunk_index":     chunkIndex,
		"chunks_received": received,
		"total_chunks":    session.TotalChunks,
		"progress_pct":    progressPct(received, session.TotalChunks),
	})
}

// CompleteChunkedUpload menggabungkan seluruh chunk menjadi satu file final
// dan menutup sesi.
func (h *ChunkedUploadHandler) CompleteChunkedUpload(c *gin.Context) {
	session, ok := h.session(c)
	if !ok {
		return
	}

	if received := session.Received(); received != session.TotalChunks {
		response.Error(c, http.StatusBadRequest,
			fmt.Sprintf("Missing chunks: %d/%d", received, session.TotalChunks))
		return
	}

	rawDir := filepath.Join(h.uploadPath, "raw")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create final file")
		return
	}

	finalFilename := uuid.New().String() + filepath.Ext(session.Filename)
	finalPath := filepath.Join(rawDir, finalFilename)
	if err := assembleChunks(session, finalPath); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Sesi ditutup hanya setelah file final benar-benar tersusun.
	h.sessions.Remove(session.UploadID)

	fileInfo, err := os.Stat(finalPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to stat final file")
		return
	}

	response.Created(c, gin.H{
		"filename": finalFilename,
		"path":     finalPath,
		"size":     fileInfo.Size(),
	})
}

// GetUploadProgress mengembalikan progres upload yang sedang berjalan.
func (h *ChunkedUploadHandler) GetUploadProgress(c *gin.Context) {
	session, ok := h.session(c)
	if !ok {
		return
	}

	received := session.Received()
	response.OK(c, gin.H{
		"upload_id":       session.UploadID,
		"chunks_received": received,
		"total_chunks":    session.TotalChunks,
		"progress_pct":    progressPct(received, session.TotalChunks),
		"status":          "in_progress",
	})
}

// AbortUpload membatalkan sesi dan menghapus chunk yang sudah diterima.
func (h *ChunkedUploadHandler) AbortUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")
	if _, ok := h.sessions.Get(uploadID); !ok {
		response.Error(c, http.StatusNotFound, "Upload session tidak ditemukan")
		return
	}
	h.sessions.Remove(uploadID)
	response.OK(c, gin.H{"message": "Upload dibatalkan"})
}

// session mengambil sesi dari store, atau menulis error 404 dan mengembalikan
// false bila sesi tidak ada.
func (h *ChunkedUploadHandler) session(c *gin.Context) (*service.Session, bool) {
	uploadID := c.Param("upload_id")
	session, ok := h.sessions.Get(uploadID)
	if !ok {
		response.Error(c, http.StatusNotFound, "Upload session tidak ditemukan")
		return nil, false
	}
	return session, true
}

// assembleChunks menyalin seluruh chunk secara berurutan ke finalPath.
//
// File final yang gagal disusun dihapus agar tidak menyisakan file separuh jadi.
func assembleChunks(session *service.Session, finalPath string) error {
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("failed to create final file")
	}

	cleanup := func() {
		finalFile.Close()
		os.Remove(finalPath)
	}

	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.ChunksDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			cleanup()
			return fmt.Errorf("failed to read chunk %d", i)
		}

		_, copyErr := io.Copy(finalFile, chunkFile)
		chunkFile.Close()
		if copyErr != nil {
			cleanup()
			return fmt.Errorf("failed to assemble chunk %d", i)
		}
	}

	if err := finalFile.Close(); err != nil {
		os.Remove(finalPath)
		return fmt.Errorf("failed to finalize assembled file")
	}
	return nil
}

// progressPct menghitung persentase progres upload, aman dari pembagian nol.
func progressPct(received, total int) int {
	if total <= 0 {
		return 0
	}
	return received * 100 / total
}
