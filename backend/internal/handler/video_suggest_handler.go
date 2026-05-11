package handler

import (
	"net/http"
	"strconv"
	"strings"

	"pgarcv/internal/repository"
	"pgarcv/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoSuggestHandler struct{ db *gorm.DB }

func NewVideoSuggestHandler(db *gorm.DB) *VideoSuggestHandler { return &VideoSuggestHandler{db: db} }

type videoSuggestResponse struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Subtitle      string `json:"subtitle"`
	ThumbnailURL  string `json:"thumbnail_url"`
	Category      string `json:"category"`
	Genre         string `json:"genre"`
}

func (h *VideoSuggestHandler) Suggest(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	limitStr := c.DefaultQuery("limit", "8")
	limit64, _ := strconv.ParseInt(limitStr, 10, 32)
	limit := int(limit64)
	if limit < 1 || limit > 20 {
		limit = 8
	}

	if q == "" {
		response.Error(c, http.StatusBadRequest, "Query tidak boleh kosong")
		return
	}

	like := "%" + q + "%"
	var videos []repository.Video
	// Pastikan cocok dengan beberapa kolom untuk UX pencarian
	h.db.
		Where(
			"title LIKE ? OR subtitle LIKE ? OR category LIKE ? OR genre LIKE ?",
			like, like, like, like,
		).
		Where("is_published = ?", true).
		Order("created_at DESC").
		Limit(limit).
		Find(&videos)

	resp := make([]videoSuggestResponse, 0, len(videos))
	for _, v := range videos {
		resp = append(resp, videoSuggestResponse{
			ID:           v.ID,
			Title:        v.Title,
			Subtitle:     v.Subtitle,
			ThumbnailURL: v.ThumbnailURL,
			Category:     v.Category,
			Genre:        v.Genre,
		})
	}
	response.OK(c, resp)
}

