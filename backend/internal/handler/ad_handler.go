package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pgarcv/internal/repository"
	"pgarcv/pkg/response"
)

type AdHandler struct{ db *gorm.DB }

type adRequest struct {
	Name       *string `json:"name"`
	AdCode     *string `json:"ad_code"`
	DeviceType *string `json:"device_type"`
	IsActive   *bool   `json:"is_active"`
}

func NewAdHandler(db *gorm.DB) *AdHandler { return &AdHandler{db: db} }

func (h *AdHandler) Serve(c *gin.Context) {
	ua := strings.ToLower(c.GetHeader("User-Agent"))
	deviceType := "desktop"
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		deviceType = "mobile"
	}

	var ad repository.Ad
	result := h.db.Where("is_active = ? AND (device_type = ? OR device_type = 'all')", true, deviceType).
		Order("RANDOM()").First(&ad)
	if result.Error != nil {
		response.OK(c, nil)
		return
	}
	response.OK(c, ad)
}

func (h *AdHandler) AdminList(c *gin.Context) {
	var ads []repository.Ad
	h.db.Order("created_at DESC").Find(&ads)
	response.OK(c, ads)
}

func (h *AdHandler) Create(c *gin.Context) {
	var req adRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	if req.Name == nil || *req.Name == "" {
		response.Error(c, 400, "Nama wajib diisi")
		return
	}

	ad := repository.Ad{Name: *req.Name, DeviceType: "all", IsActive: true}
	applyAdRequest(&ad, req)
	if !validDeviceType(ad.DeviceType) {
		response.Error(c, 400, "Target device tidak valid")
		return
	}
	h.db.Create(&ad)
	response.Created(c, ad)
}

func (h *AdHandler) Update(c *gin.Context) {
	var ad repository.Ad
	if err := h.db.First(&ad, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Iklan tidak ditemukan")
		return
	}

	var req adRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	applyAdRequest(&ad, req)
	if !validDeviceType(ad.DeviceType) {
		response.Error(c, 400, "Target device tidak valid")
		return
	}
	h.db.Save(&ad)
	response.OK(c, ad)
}

func (h *AdHandler) Delete(c *gin.Context) {
	result := h.db.Delete(&repository.Ad{}, c.Param("id"))
	if result.Error != nil || result.RowsAffected == 0 {
		response.Error(c, 404, "Iklan tidak ditemukan")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func applyAdRequest(ad *repository.Ad, req adRequest) {
	if req.Name != nil {
		ad.Name = *req.Name
	}
	if req.AdCode != nil {
		ad.AdCode = *req.AdCode
	}
	if req.DeviceType != nil {
		ad.DeviceType = *req.DeviceType
	}
	if req.IsActive != nil {
		ad.IsActive = *req.IsActive
	}
}

func validDeviceType(deviceType string) bool {
	return deviceType == "all" || deviceType == "mobile" || deviceType == "desktop"
}
