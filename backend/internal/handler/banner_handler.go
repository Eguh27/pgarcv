package handler

import (
	"net/http"
	"strings"

	"pgarcv/internal/model"
	"pgarcv/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BannerHandler melayani endpoint banner.
type BannerHandler struct{ db *gorm.DB }

// NewBannerHandler membuat BannerHandler.
func NewBannerHandler(db *gorm.DB) *BannerHandler { return &BannerHandler{db: db} }

// bannerRequest adalah payload create/update banner (partial update).
type bannerRequest struct {
	Title     *string `json:"title"`
	ImageURL  *string `json:"image_url"`
	LinkURL   *string `json:"link_url"`
	IsActive  *bool   `json:"is_active"`
	SortOrder *int    `json:"sort_order"`
}

// ListActive mengembalikan banner aktif untuk tampilan publik.
func (h *BannerHandler) ListActive(c *gin.Context) {
	var banners []model.Banner
	err := h.db.Where("is_active = ?", true).
		Order("sort_order ASC").Find(&banners).Error
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memuat banner")
		return
	}
	response.OK(c, banners)
}

// AdminList mengembalikan seluruh banner termasuk yang nonaktif.
func (h *BannerHandler) AdminList(c *gin.Context) {
	var banners []model.Banner
	if err := h.db.Order("sort_order ASC").Find(&banners).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memuat banner")
		return
	}
	response.OK(c, banners)
}

// Create membuat banner baru.
func (h *BannerHandler) Create(c *gin.Context) {
	var req bannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	if req.Title == nil || strings.TrimSpace(*req.Title) == "" {
		response.Error(c, http.StatusBadRequest, "Judul wajib diisi")
		return
	}

	banner := model.Banner{Title: strings.TrimSpace(*req.Title), IsActive: true}
	applyBannerRequest(&banner, req)

	if err := h.db.Create(&banner).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan banner")
		return
	}
	response.Created(c, banner)
}

// Update mengubah field banner yang dikirim pada request.
func (h *BannerHandler) Update(c *gin.Context) {
	var banner model.Banner
	if err := h.db.First(&banner, c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Banner tidak ditemukan")
		return
	}

	var req bannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid")
		return
	}

	applyBannerRequest(&banner, req)

	if err := h.db.Save(&banner).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan banner")
		return
	}
	response.OK(c, banner)
}

// Delete menghapus banner.
func (h *BannerHandler) Delete(c *gin.Context) {
	result := h.db.Delete(&model.Banner{}, c.Param("id"))
	if result.Error != nil || result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "Banner tidak ditemukan")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

// applyBannerRequest menyalin field yang dikirim pada req ke banner.
func applyBannerRequest(banner *model.Banner, req bannerRequest) {
	if req.Title != nil {
		banner.Title = strings.TrimSpace(*req.Title)
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
