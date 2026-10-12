package handler

import (
	"net/http"
	"strings"

	"pgarcv/internal/model"
	"pgarcv/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// suggestLimitDefault dan suggestLimitMax membatasi ukuran hasil saran
// pencarian yang dikembalikan ke klien.
const (
	suggestLimitDefault = 8
	suggestLimitMax     = 20
)

// VideoSuggestHandler melayani endpoint saran pencarian video.
type VideoSuggestHandler struct{ db *gorm.DB }

// NewVideoSuggestHandler membuat VideoSuggestHandler.
func NewVideoSuggestHandler(db *gorm.DB) *VideoSuggestHandler {
	return &VideoSuggestHandler{db: db}
}

// videoSuggestResponse adalah bentuk ringkas untuk dropdown saran. Field yang
// tidak ada di sini sengaja tidak ikut agar payload tetap kecil.
type videoSuggestResponse struct {
	ID           uint   `json:"id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	ThumbnailURL string `json:"thumbnail_url"`
	Category     string `json:"category"`
	Genre        string `json:"genre"`
}

// Suggest mengembalikan video published yang judulnya cocok dengan query.
func (h *VideoSuggestHandler) Suggest(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		response.Error(c, http.StatusBadRequest, "Query tidak boleh kosong")
		return
	}

	limit := queryInt(c, "limit", suggestLimitDefault)
	if limit < 1 || limit > suggestLimitMax {
		limit = suggestLimitDefault
	}

	like := "%" + q + "%"
	var videos []model.Video
	err := h.db.
		Where("is_published = ?", true).
		Where("title LIKE ? OR subtitle LIKE ? OR category LIKE ? OR genre LIKE ?", like, like, like, like).
		Order("created_at DESC").
		Limit(limit).
		Find(&videos).Error
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memuat saran")
		return
	}

	results := make([]videoSuggestResponse, 0, len(videos))
	for _, v := range videos {
		results = append(results, videoSuggestResponse{
			ID:           v.ID,
			Title:        v.Title,
			Subtitle:     v.Subtitle,
			ThumbnailURL: v.ThumbnailURL,
			Category:     v.Category,
			Genre:        v.Genre,
		})
	}
	response.OK(c, results)
}
