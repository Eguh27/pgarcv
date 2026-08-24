package handler

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pgarcv/internal/repository"
	"pgarcv/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HLSHandler struct {
	db         *gorm.DB
	outputBase string
	encryptKey string
	serverURL  string
}

func NewHLSHandler(db *gorm.DB, outputBase, encryptKey, serverURL string) *HLSHandler {
	return &HLSHandler{db: db, outputBase: outputBase, encryptKey: encryptKey, serverURL: serverURL}
}

// canAccessVideo checks if user can access a video
// - Published videos: anyone can access
// - Unpublished videos: only authenticated admins
func (h *HLSHandler) canAccessVideo(videoID uint) (video *repository.Video, err error, forbidden bool) {
	v := &repository.Video{}
	if err := h.db.First(v, videoID).Error; err != nil {
		return nil, err, false
	}
	return v, nil, false
}

func (h *HLSHandler) canAccessVideoForUser(c *gin.Context, videoID uint) (*repository.Video, bool) {
	video := &repository.Video{}
	if err := h.db.First(video, videoID).Error; err != nil {
		// Video not found
		return nil, false
	}

	// Published videos: anyone can access
	if video.IsPublished {
		return video, true
	}

	// Unpublished videos: only authenticated users (admins) can access
	if auth, exists := c.Get("authenticated"); exists && auth.(bool) {
		return video, true
	}

	// Unpublished and not authenticated
	return video, false
}

func (h *HLSHandler) ServeKey(c *gin.Context) {
	// Validasi referer — hanya izinkan dari domain sendiri
	referer := c.GetHeader("Referer")
	origin := c.GetHeader("Origin")

	allowed := false
	for _, check := range []string{referer, origin} {
		if check != "" && (strings.HasPrefix(check, h.serverURL) ||
			strings.HasPrefix(check, "http://localhost")) {
			allowed = true
			break
		}
	}
	if !allowed {
		c.Status(403)
		return
	}

	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "ID tidak valid")
		return
	}
	var video repository.Video
	if err := h.db.First(&video, videoID).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	if !video.IsPublished {
		response.Error(c, 403, "Video tidak tersedia")
		return
	}
	keyBytes, err := hex.DecodeString(h.encryptKey)
	if err != nil {
		response.Error(c, 500, "Key error")
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/octet-stream", keyBytes)
}

func (h *HLSHandler) ServeM3U8(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "ID tidak valid")
		return
	}

	// ✅ Check authorization: published OR authenticated
	video, allowed := h.canAccessVideoForUser(c, uint(videoID))
	if video == nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	if !allowed {
		response.Error(c, 403, "Video tidak tersedia")
		return
	}

	// ✅ Only then serve M3U8
	m3u8Path := filepath.Join(h.outputBase, "hls",
		fmt.Sprintf("video_%d", videoID), "index.m3u8")
	if _, err := os.Stat(m3u8Path); os.IsNotExist(err) {
		response.Error(c, 404, "HLS tidak tersedia")
		return
	}
	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.Header("Cache-Control", "no-cache")
	c.File(m3u8Path)
}

func (h *HLSHandler) ServeSegment(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "ID tidak valid")
		return
	}

	// ✅ Check authorization: published OR authenticated
	video, allowed := h.canAccessVideoForUser(c, uint(videoID))
	if video == nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	if !allowed {
		response.Error(c, 403, "Video tidak tersedia")
		return
	}

	segment := c.Param("segment")
	// Guard traversal: tolak dotfiles, "..", dan subpath.
	// (filepath.Base saja tidak cukup — Base("..") == "..")
	if segment == "" || strings.HasPrefix(segment, ".") || filepath.Base(segment) != segment {
		response.Error(c, 400, "Invalid segment")
		return
	}
	segPath := filepath.Join(h.outputBase, "hls",
		fmt.Sprintf("video_%d", videoID), segment)
	if _, err := os.Stat(segPath); os.IsNotExist(err) {
		response.Error(c, 404, "Segment tidak ditemukan")
		return
	}
	c.Header("Content-Type", "video/MP2T")
	c.File(segPath)
}
