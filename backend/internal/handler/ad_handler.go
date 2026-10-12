package handler

import (
	"math/rand"
	"net/http"
	"strings"

	"pgarcv/internal/model"
	"pgarcv/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdHandler melayani endpoint iklan.
type AdHandler struct{ db *gorm.DB }

// NewAdHandler membuat AdHandler.
func NewAdHandler(db *gorm.DB) *AdHandler { return &AdHandler{db: db} }

// adRequest adalah payload create/update iklan (partial update).
type adRequest struct {
	Name       *string `json:"name"`
	AdCode     *string `json:"ad_code"`
	DeviceType *string `json:"device_type"`
	IsActive   *bool   `json:"is_active"`
}

// Serve mengembalikan satu iklan aktif yang cocok dengan perangkat requester.
//
// outwardly respons untuk device tanpa iklan sengaja mengembalikan data null
// dengan status 200 supaya klien tidak perlu menangani error.
func (h *AdHandler) Serve(c *gin.Context) {
	deviceType := deviceTypeFromUserAgent(c.GetHeader("User-Agent"))

	var ads []model.Ad
	err := h.db.Where("is_active = ? AND (device_type = ? OR device_type = ?)",
		true, deviceType, model.DeviceTypeAll).
		Limit(20).
		Find(&ads).Error
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memuat iklan")
		return
	}
	if len(ads) == 0 {
		response.OK(c, nil)
		return
	}

	response.OK(c, ads[rand.Intn(len(ads))])
}

// deviceTypeFromUserAgent menebak perangkat requester dari User-Agent.
func deviceTypeFromUserAgent(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return model.DeviceTypeMobile
	}
	return model.DeviceTypeDesktop
}

// AdminList mengembalikan seluruh iklan termasuk yang nonaktif.
func (h *AdHandler) AdminList(c *gin.Context) {
	var ads []model.Ad
	if err := h.db.Order("created_at DESC").Find(&ads).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memuat iklan")
		return
	}
	response.OK(c, ads)
}

// Create membuat iklan baru.
func (h *AdHandler) Create(c *gin.Context) {
	var req adRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid")
		return
	}
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		response.Error(c, http.StatusBadRequest, "Nama wajib diisi")
		return
	}

	ad := model.Ad{Name: strings.TrimSpace(*req.Name), DeviceType: model.DeviceTypeAll, IsActive: true}
	applyAdRequest(&ad, req)
	if !model.ValidDeviceType(ad.DeviceType) {
		response.Error(c, http.StatusBadRequest, "Target device tidak valid")
		return
	}

	if err := h.db.Create(&ad).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan iklan")
		return
	}
	response.Created(c, ad)
}

// Update mengubah field iklan yang dikirim pada request.
func (h *AdHandler) Update(c *gin.Context) {
	var ad model.Ad
	if err := h.db.First(&ad, c.Param("id")).Error; err != nil {
		response.Error(c, http.StatusNotFound, "Iklan tidak ditemukan")
		return
	}

	var req adRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid")
		return
	}

	applyAdRequest(&ad, req)
	if !model.ValidDeviceType(ad.DeviceType) {
		response.Error(c, http.StatusBadRequest, "Target device tidak valid")
		return
	}

	if err := h.db.Save(&ad).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan iklan")
		return
	}
	response.OK(c, ad)
}

// Delete menghapus iklan.
func (h *AdHandler) Delete(c *gin.Context) {
	result := h.db.Delete(&model.Ad{}, c.Param("id"))
	if result.Error != nil || result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, "Iklan tidak ditemukan")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

// applyAdRequest menyalin field yang dikirim pada req ke ad.
func applyAdRequest(ad *model.Ad, req adRequest) {
	if req.Name != nil {
		ad.Name = strings.TrimSpace(*req.Name)
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
