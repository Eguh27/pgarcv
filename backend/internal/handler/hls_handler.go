package handler

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pgarcv/internal/middleware"
	"pgarcv/internal/model"
	"pgarcv/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HLSHandler menyajikan playlist, segmen, dan kunci enkripsi HLS.
type HLSHandler struct {
	db         *gorm.DB
	outputBase string
	encryptKey string
	serverURL  string
}

// NewHLSHandler membuat HLSHandler.
func NewHLSHandler(db *gorm.DB, outputBase, encryptKey, serverURL string) *HLSHandler {
	return &HLSHandler{db: db, outputBase: outputBase, encryptKey: encryptKey, serverURL: serverURL}
}

// videoDir adalah direktori output HLS untuk satu video.
func (h *HLSHandler) videoDir(videoID uint) string {
	return filepath.Join(h.outputBase, "hls", fmt.Sprintf("video_%d", videoID))
}

// canAccessVideo memutuskan apakah request diizinkan mengakses HLS sebuah video.
//
// Aturannya: video published boleh diakses siapa pun; video draft hanya boleh
// diakses admin yang terautentikasi (ditandai middleware.OptionalAuth lewat
// ContextIsAdmin). Video tidak ada menghasilkan video == nil.
func (h *HLSHandler) canAccessVideo(c *gin.Context, videoID uint) (*model.Video, bool) {
	var video model.Video
	if err := h.db.First(&video, videoID).Error; err != nil {
		return nil, false
	}

	if video.IsPublished {
		return &video, true
	}

	if isAdmin, ok := c.Get(middleware.ContextIsAdmin); ok && isAdmin == true {
		return &video, true
	}
	return &video, false
}

// authorize mengambil video dan menolak request yang tidak berhak.
//
// Menyatukan pengecekan "video ada" dan "video boleh diakses" supaya tiap
// endpoint tidak mengulang logikanya.
func (h *HLSHandler) authorize(c *gin.Context) (*model.Video, bool) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID tidak valid")
		return nil, false
	}

	video, allowed := h.canAccessVideo(c, uint(videoID))
	if video == nil {
		response.Error(c, http.StatusNotFound, "Video tidak ditemukan")
		return nil, false
	}
	if !allowed {
		response.Error(c, http.StatusForbidden, "Video tidak tersedia")
		return nil, false
	}
	return video, true
}

// allowedOrigin melapor apakah referer atau origin request berasal dari server
// sendiri atau dari localhost.
func (h *HLSHandler) allowedOrigin(referer, origin string) bool {
	for _, candidate := range []string{referer, origin} {
		if candidate == "" {
			continue
		}
		if strings.HasPrefix(candidate, h.serverURL) || strings.HasPrefix(candidate, "http://localhost") {
			return true
		}
	}
	return false
}

// ServeKey menyajikan kunci AES-128 untuk decryption HLS.
//
// Kunci hanya diberikan pada request yang dianggap berasal dari pemutar di
// origin sendiri, supaya URL kunci tidak bisa dipakai di situs pihak ketiga.
func (h *HLSHandler) ServeKey(c *gin.Context) {
	if !h.allowedOrigin(c.GetHeader("Referer"), c.GetHeader("Origin")) {
		c.Status(http.StatusForbidden)
		return
	}

	if _, ok := h.authorize(c); !ok {
		return
	}

	keyBytes, err := hex.DecodeString(h.encryptKey)
	if err != nil || len(keyBytes) != 16 {
		response.Error(c, http.StatusInternalServerError, "Kunci enkripsi tidak valid")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/octet-stream", keyBytes)
}

// ServeM3U8 menyajikan playlist HLS.
func (h *HLSHandler) ServeM3U8(c *gin.Context) {
	video, ok := h.authorize(c)
	if !ok {
		return
	}

	videoID := video.ID
	m3u8Path := filepath.Join(h.videoDir(videoID), "index.m3u8")
	if _, err := os.Stat(m3u8Path); err != nil {
		response.Error(c, http.StatusNotFound, "HLS tidak tersedia")
		return
	}

	// http.ServeFile hanya menetapkan Content-Type bila belum ada, jadi header
	// eksplisit di sini memang yang dipakai klien.
	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.Header("Cache-Control", "no-cache")
	c.File(m3u8Path)
}

// ServeSegment menyajikan satu segmen HLS.
func (h *HLSHandler) ServeSegment(c *gin.Context) {
	video, ok := h.authorize(c)
	if !ok {
		return
	}

	segment := c.Param("segment")
	// Guard traversal: tolak nama kosong, path relatif, dan apa pun yang bukan
	// nama file dasar. filepath.Base saja tidak cukup karena Base("..") == "..".
	if segment == "" || strings.HasPrefix(segment, ".") || filepath.Base(segment) != segment {
		response.Error(c, http.StatusBadRequest, "Segment tidak valid")
		return
	}

	segPath := filepath.Join(h.videoDir(video.ID), segment)
	if _, err := os.Stat(segPath); err != nil {
		response.Error(c, http.StatusNotFound, "Segment tidak ditemukan")
		return
	}

	c.Header("Content-Type", "video/MP2T")
	c.File(segPath)
}
