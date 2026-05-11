package handler

import (
	"log"
	"strconv"
	"strings"

	"pgarcv/internal/repository"
	"pgarcv/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VideoHandler struct{ db *gorm.DB }

type videoRequest struct {
	Title         *string   `json:"title"`
	Subtitle      *string   `json:"subtitle"`
	Description   *string   `json:"description"`
	ThumbnailURL  *string   `json:"thumbnail_url"`
	VideoURL      *string   `json:"video_url"`
	PreviewURL    *string   `json:"preview_url"`
	Duration      *int      `json:"duration"`
	Category      *string   `json:"category"`
	Genre         *string   `json:"genre"`
	Categories    *[]string `json:"categories"`
	Genres        *[]string `json:"genres"`
	IsPublished   *bool     `json:"is_published"`
	AllowDownload *bool     `json:"allow_download"`
}

func NewVideoHandler(db *gorm.DB) *VideoHandler { return &VideoHandler{db: db} }

func (h *VideoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	search := c.Query("search")
	genre := strings.TrimSpace(c.Query("genre"))
	category := strings.TrimSpace(c.Query("category"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}
	offset := (page - 1) * limit

	var videos []repository.Video
	var total int64
	q := h.db.Model(&repository.Video{}).Where("is_published = ?", true)
	if search != "" {
		q = q.Where(
			"title LIKE ? OR subtitle LIKE ? OR category LIKE ? OR genre LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
			"%"+search+"%",
		)
	}
	if genre != "" {
		q = q.Joins("JOIN video_genres ON video_genres.video_id = videos.id").
			Joins("JOIN genres ON genres.id = video_genres.genre_id").
			Where("genres.name = ?", genre)
	}
	if category != "" {
		q = q.Joins("JOIN video_categories ON video_categories.video_id = videos.id").
			Joins("JOIN categories ON categories.id = video_categories.category_id").
			Where("categories.name = ?", category)
	}
	q.Count(&total)
	q = q.Distinct()
	q.Preload("Categories").Preload("Genres").Order("created_at DESC").Offset(offset).Limit(limit).Find(&videos)
	response.Paginated(c, videos, total, page, limit)
}

func (h *VideoHandler) Get(c *gin.Context) {
	var video repository.Video
	if err := h.db.Preload("Categories").Preload("Genres").Where("is_published = ?", true).First(&video, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	h.db.Model(&video).UpdateColumn("views", gorm.Expr("views + ?", 1))
	video.Views++
	response.OK(c, video)
}

func (h *VideoHandler) Featured(c *gin.Context) {
	var videos []repository.Video
	h.db.Preload("Categories").Preload("Genres").Where("is_published = ?", true).Order("views DESC").Limit(6).Find(&videos)
	response.OK(c, videos)
}

func (h *VideoHandler) AdminList(c *gin.Context) {
	var videos []repository.Video
	var total int64
	h.db.Model(&repository.Video{}).Count(&total)
	h.db.Preload("Categories").Preload("Genres").Order("created_at DESC").Find(&videos)
	// limit fixed for admin list: since we always fetch without pagination,
	// use len(videos) so response's limit matches the returned items.
	response.Paginated(c, videos, total, 1, len(videos))

}

func (h *VideoHandler) AdminGet(c *gin.Context) {
	var video repository.Video
	if err := h.db.Preload("Categories").Preload("Genres").First(&video, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	response.OK(c, video)
}

func (h *VideoHandler) Create(c *gin.Context) {
	var req videoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid: "+err.Error())
		return
	}
	log.Printf("[video.create] allow_download=%v categories=%v genres=%v", ptrBool(req.AllowDownload), ptrSliceLen(req.Categories), ptrSliceLen(req.Genres))
	if req.Title == nil || *req.Title == "" {
		response.Error(c, 400, "Judul wajib diisi")
		return
	}

	video := repository.Video{Title: *req.Title}
	applyVideoRequest(&video, req)
	if err := h.db.Create(&video).Error; err != nil {
		response.Error(c, 500, "Gagal menyimpan video")
		return
	}
	if err := h.applyTaxonomy(&video, req); err != nil {
		response.Error(c, 500, "Gagal menyimpan genre/category")
		return
	}
	h.db.Preload("Categories").Preload("Genres").First(&video, video.ID)
	response.Created(c, video)
}

func (h *VideoHandler) Update(c *gin.Context) {
	var video repository.Video
	if err := h.db.First(&video, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}

	var req videoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	log.Printf("[video.update] id=%s allow_download=%v categories=%v genres=%v", c.Param("id"), ptrBool(req.AllowDownload), ptrSliceLen(req.Categories), ptrSliceLen(req.Genres))
	applyVideoRequest(&video, req)
	if err := h.db.Save(&video).Error; err != nil {
		response.Error(c, 500, "Gagal menyimpan video")
		return
	}
	if err := h.applyTaxonomy(&video, req); err != nil {
		response.Error(c, 500, "Gagal menyimpan genre/category")
		return
	}
	h.db.Preload("Categories").Preload("Genres").First(&video, video.ID)
	response.OK(c, video)
}

func (h *VideoHandler) ListCategories(c *gin.Context) {
	var items []repository.Category
	q := strings.TrimSpace(c.Query("q"))
	dbq := h.db.Model(&repository.Category{})
	if q != "" {
		dbq = dbq.Where("name LIKE ?", "%"+q+"%")
	}
	dbq.Order("name ASC").Limit(50).Find(&items)
	response.OK(c, items)
}

func (h *VideoHandler) ListGenres(c *gin.Context) {
	var items []repository.Genre
	q := strings.TrimSpace(c.Query("q"))
	dbq := h.db.Model(&repository.Genre{})
	if q != "" {
		dbq = dbq.Where("name LIKE ?", "%"+q+"%")
	}
	dbq.Order("name ASC").Limit(50).Find(&items)
	response.OK(c, items)
}

func (h *VideoHandler) Delete(c *gin.Context) {
	result := h.db.Delete(&repository.Video{}, c.Param("id"))
	if result.Error != nil || result.RowsAffected == 0 {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func applyVideoRequest(video *repository.Video, req videoRequest) {
	if req.Title != nil {
		video.Title = *req.Title
	}
	if req.Subtitle != nil {
		video.Subtitle = *req.Subtitle
	}
	if req.Description != nil {
		video.Description = *req.Description
	}
	if req.ThumbnailURL != nil {
		video.ThumbnailURL = *req.ThumbnailURL
	}
	if req.VideoURL != nil {
		video.VideoURL = *req.VideoURL
	}
	if req.PreviewURL != nil {
		video.PreviewURL = *req.PreviewURL
	}
	if req.Duration != nil {
		video.Duration = *req.Duration
	}
	if req.Category != nil {
		video.Category = *req.Category
	}
	if req.Genre != nil {
		video.Genre = *req.Genre
	}
	if req.IsPublished != nil {
		video.IsPublished = *req.IsPublished
	}
	if req.AllowDownload != nil {
		video.AllowDownload = *req.AllowDownload
	}
}

func normalizeName(s string) string {
	s = strings.TrimSpace(s)
	// collapse multiple spaces
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func ptrBool(b *bool) any {
	if b == nil {
		return nil
	}
	return *b
}
func ptrSliceLen(s *[]string) any {
	if s == nil {
		return nil
	}
	return len(*s)
}

func (h *VideoHandler) applyTaxonomy(video *repository.Video, req videoRequest) error {
	if req.Categories == nil && req.Genres == nil {
		return nil
	}

	if req.Categories != nil {
		var cats []repository.Category
		for _, raw := range *req.Categories {
			name := normalizeName(raw)
			if name == "" {
				continue
			}
			var cat repository.Category
			// case-insensitive lookup (SQLite)
			if err := h.db.Where("lower(name) = lower(?)", name).First(&cat).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					cat = repository.Category{Name: name}
					if err2 := h.db.Create(&cat).Error; err2 != nil {
						return err2
					}
				} else {
					return err
				}
			}
			cats = append(cats, cat)
		}
		if err := h.db.Model(video).Association("Categories").Replace(&cats); err != nil {
			return err
		}
		if len(cats) > 0 {
			video.Category = cats[0].Name
		} else {
			video.Category = ""
		}
	}

	if req.Genres != nil {
		var gens []repository.Genre
		for _, raw := range *req.Genres {
			name := normalizeName(raw)
			if name == "" {
				continue
			}
			var g repository.Genre
			if err := h.db.Where("lower(name) = lower(?)", name).First(&g).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					g = repository.Genre{Name: name}
					if err2 := h.db.Create(&g).Error; err2 != nil {
						return err2
					}
				} else {
					return err
				}
			}
			gens = append(gens, g)
		}
		if err := h.db.Model(video).Association("Genres").Replace(&gens); err != nil {
			return err
		}
		if len(gens) > 0 {
			video.Genre = gens[0].Name
		} else {
			video.Genre = ""
		}
	}

	// Keep legacy fields in sync if we touched taxonomy
	return h.db.Model(video).Updates(map[string]any{
		"category": video.Category,
		"genre":    video.Genre,
	}).Error
}
