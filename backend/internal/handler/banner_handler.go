package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pgarcv/internal/repository"
	"pgarcv/pkg/response"
)

type BannerHandler struct{ db *gorm.DB }

type bannerRequest struct {
	Title     *string `json:"title"`
	ImageURL  *string `json:"image_url"`
	LinkURL   *string `json:"link_url"`
	IsActive  *bool   `json:"is_active"`
	SortOrder *int    `json:"sort_order"`
}

func NewBannerHandler(db *gorm.DB) *BannerHandler { return &BannerHandler{db: db} }

func (h *BannerHandler) ListActive(c *gin.Context) {
	var banners []repository.Banner
	h.db.Where("is_active = ?", true).Order("sort_order ASC").Find(&banners)
	response.OK(c, banners)
}

func (h *BannerHandler) AdminList(c *gin.Context) {
	var banners []repository.Banner
	h.db.Order("sort_order ASC").Find(&banners)
	response.OK(c, banners)
}

func (h *BannerHandler) Create(c *gin.Context) {
	var req bannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	if req.Title == nil || *req.Title == "" {
		response.Error(c, 400, "Judul wajib diisi")
		return
	}

	banner := repository.Banner{Title: *req.Title, IsActive: true}
	applyBannerRequest(&banner, req)
	h.db.Create(&banner)
	response.Created(c, banner)
}

func (h *BannerHandler) Update(c *gin.Context) {
	var banner repository.Banner
	if err := h.db.First(&banner, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Banner tidak ditemukan")
		return
	}

	var req bannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	applyBannerRequest(&banner, req)
	h.db.Save(&banner)
	response.OK(c, banner)
}

func (h *BannerHandler) Delete(c *gin.Context) {
	result := h.db.Delete(&repository.Banner{}, c.Param("id"))
	if result.Error != nil || result.RowsAffected == 0 {
		response.Error(c, 404, "Banner tidak ditemukan")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func applyBannerRequest(banner *repository.Banner, req bannerRequest) {
	if req.Title != nil {
		banner.Title = *req.Title
	}
	if req.ImageURL != nil {
		banner.ImageURL = *req.ImageURL
	}
	if req.LinkURL != nil {
		banner.LinkURL = *req.LinkURL
	}
	if req.IsActive != nil {
		banner.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		banner.SortOrder = *req.SortOrder
	}
}
