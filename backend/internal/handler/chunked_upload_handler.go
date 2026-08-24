package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"pgarcv/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChunkedUploadHandler struct {
	uploadPath string
	// Track ongoing uploads: uploadID -> chunk metadata
	uploads sync.Map
}

type UploadSession struct {
	UploadID    string
	Filename    string
	TotalChunks int
	TotalSize   int64
	ChunksDir   string
	ChunksMask  map[int]bool // Track received chunks
	mu          sync.Mutex
}

func NewChunkedUploadHandler(uploadPath string) *ChunkedUploadHandler {
	return &ChunkedUploadHandler{uploadPath: uploadPath}
}

// InitiateChunkedUpload creates a new upload session
func (h *ChunkedUploadHandler) InitiateChunkedUpload(c *gin.Context) {
	var req struct {
		Filename    string `json:"filename" binding:"required"`
		FileSize    int64  `json:"file_size" binding:"required"`
		TotalChunks int    `json:"total_chunks" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid request: "+err.Error())
		return
	}

	if req.FileSize > 500*1024*1024 {
		response.Error(c, 400, "File terlalu besar (maks 500MB)")
		return
	}

	if req.TotalChunks < 1 || req.TotalChunks > 500 {
		response.Error(c, 400, "Total chunks harus 1-500")
		return
	}

	uploadID := uuid.New().String()
	chunksDir := filepath.Join(h.uploadPath, ".chunks", uploadID)
	os.MkdirAll(chunksDir, 0755)

	session := &UploadSession{
		UploadID:    uploadID,
		Filename:    req.Filename,
		TotalChunks: req.TotalChunks,
		TotalSize:   req.FileSize,
		ChunksDir:   chunksDir,
		ChunksMask:  make(map[int]bool),
	}

	h.uploads.Store(uploadID, session)

	response.Created(c, gin.H{
		"upload_id":    uploadID,
		"total_chunks": req.TotalChunks,
		"chunk_size":   5 * 1024 * 1024, // 5MB recommended
	})
}

// UploadChunk uploads a single chunk
func (h *ChunkedUploadHandler) UploadChunk(c *gin.Context) {
	uploadID := c.Param("upload_id")
	chunkIndex, err := strconv.Atoi(c.PostForm("chunk_index"))
	if err != nil {
		response.Error(c, 400, "Invalid chunk_index")
		return
	}

	// Get upload session
	val, ok := h.uploads.Load(uploadID)
	if !ok {
		response.Error(c, 404, "Upload session tidak ditemukan")
		return
	}
	session := val.(*UploadSession)

	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		response.Error(c, 400, "Chunk index out of range")
		return
	}

	// Get file from request
	file, err := c.FormFile("chunk")
	if err != nil {
		response.Error(c, 400, "Chunk file required")
		return
	}

	// Save chunk
	chunkPath := filepath.Join(session.ChunksDir, fmt.Sprintf("chunk_%d", chunkIndex))
	if err := c.SaveUploadedFile(file, chunkPath); err != nil {
		response.Error(c, 500, "Failed to save chunk")
		return
	}

	// Mark chunk as received
	session.mu.Lock()
	session.ChunksMask[chunkIndex] = true
	received := len(session.ChunksMask)
	session.mu.Unlock()

	response.OK(c, gin.H{
		"chunk_index":     chunkIndex,
		"chunks_received": received,
		"total_chunks":    session.TotalChunks,
		"progress_pct":    int(float64(received) / float64(session.TotalChunks) * 100),
	})
}

// CompleteChunkedUpload assembles chunks into final file
func (h *ChunkedUploadHandler) CompleteChunkedUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")

	val, ok := h.uploads.Load(uploadID)
	if !ok {
		response.Error(c, 404, "Upload session tidak ditemukan")
		return
	}
	session := val.(*UploadSession)

	// Verify all chunks received
	session.mu.Lock()
	if len(session.ChunksMask) != session.TotalChunks {
		session.mu.Unlock()
		received := len(session.ChunksMask)
		response.Error(c, 400, fmt.Sprintf("Missing chunks: %d/%d", received, session.TotalChunks))
		return
	}
	session.mu.Unlock()

	// Assemble file
	ext := filepath.Ext(session.Filename)
	finalFilename := uuid.New().String() + ext
	rawDir := filepath.Join(h.uploadPath, "raw")
	os.MkdirAll(rawDir, 0755)
	finalPath := filepath.Join(rawDir, finalFilename)

	finalFile, err := os.Create(finalPath)
	if err != nil {
		response.Error(c, 500, "Failed to create final file")
		return
	}
	defer finalFile.Close()

	// Assemble chunks in order
	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.ChunksDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			finalFile.Close()
			os.Remove(finalPath)
			response.Error(c, 500, fmt.Sprintf("Failed to read chunk %d", i))
			return
		}

		if _, err := io.Copy(finalFile, chunkFile); err != nil {
			chunkFile.Close()
			finalFile.Close()
			os.Remove(finalPath)
			response.Error(c, 500, fmt.Sprintf("Failed to assemble chunk %d", i))
			return
		}
		chunkFile.Close()
	}

	// Cleanup chunks directory
	os.RemoveAll(session.ChunksDir)

	// Remove upload session
	h.uploads.Delete(uploadID)

	// Get file size
	fileInfo, err := os.Stat(finalPath)
	if err != nil {
		response.Error(c, 500, "Failed to stat final file")
		return
	}

	response.Created(c, gin.H{
		"filename": finalFilename,
		"path":     finalPath,
		"size":     fileInfo.Size(),
	})
}

// GetUploadProgress returns current upload progress
func (h *ChunkedUploadHandler) GetUploadProgress(c *gin.Context) {
	uploadID := c.Param("upload_id")

	val, ok := h.uploads.Load(uploadID)
	if !ok {
		response.Error(c, 404, "Upload session tidak ditemukan")
		return
	}
	session := val.(*UploadSession)

	session.mu.Lock()
	received := len(session.ChunksMask)
	session.mu.Unlock()

	progress := int(float64(received) / float64(session.TotalChunks) * 100)

	response.OK(c, gin.H{
		"upload_id":       uploadID,
		"chunks_received": received,
		"total_chunks":    session.TotalChunks,
		"progress_pct":    progress,
		"status":          "in_progress",
	})
}

// AbortUpload cancels an upload session
func (h *ChunkedUploadHandler) AbortUpload(c *gin.Context) {
	uploadID := c.Param("upload_id")

	val, ok := h.uploads.Load(uploadID)
	if !ok {
		response.Error(c, 404, "Upload session tidak ditemukan")
		return
	}
	session := val.(*UploadSession)

	// Cleanup
	os.RemoveAll(session.ChunksDir)
	h.uploads.Delete(uploadID)

	response.OK(c, gin.H{"message": "Upload dibatalkan"})
}
