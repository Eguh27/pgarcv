package handler

import (
	"log"
	"strings"

	"pgarcv/internal/model"
	"pgarcv/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// pageSizeDefault dan pageSizeMax membatasi ukuran halaman daftar video publik.
const (
	pageSizeDefault = 12
	pageSizeMax     = 50
)

// VideoHandler melayani endpoint video: listing publik, featured, dan CRUD admin.
type VideoHandler struct{ db *gorm.DB }

// NewVideoHandler membuat VideoHandler.
func NewVideoHandler(db *gorm.DB) *VideoHandler { return &VideoHandler{db: db} }

// List mengembalikan daftar video yang published dengan filter dan paginasi.
//
// Filter genre/category memakai JOIN ke tabel pivot karena Video menyimpan
// relasi many2many; kolom Category/Genre yang legacy tidak dipakai sebagai
// sumber kebenaran.
func (h *VideoHandler) List(c *gin.Context) {
	page := queryInt(c, "page", 1)
	limit := queryInt(c, "limit", pageSizeDefault)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > pageSizeMax {
		limit = pageSizeDefault
	}

	search := strings.TrimSpace(c.Query("search"))
	genre := strings.TrimSpace(c.Query("genre"))
	category := strings.TrimSpace(c.Query("category"))

	base := h.db.Model(&model.Video{}).Where("is_published = ?", true)

	if search != "" {
		like := "%" + search + "%"
		base = base.Where(
			"title LIKE ? OR subtitle LIKE ? OR category LIKE ? OR genre LIKE ?",
			like, like, like, like,
		)
	}
	if genre != "" {
		base = base.
			Joins("JOIN video_genres ON video_genres.video_id = videos.id").
			Joins("JOIN genres ON genres.id = video_genres.genre_id").
			Where("genres.name = ?", genre)
	}
	if category != "" {
		base = base.
			Joins("JOIN video_categories ON video_categories.video_id = videos.id").
			Joins("JOIN categories ON categories.id = video_categories.category_id").
			Where("categories.name = ?", category)
	}

	// Count harus memakai sesi terpisah: GORM menyimpan state query pada
	// statement yang sama, sehingga Count lalu Find pada objek yang sama bisa
	// saling menimpa kondisinya.
	var total int64
	if err := base.Distinct().Session(&gorm.Session{}).Count(&total).Error; err != nil {
		response.Error(c, 500, "Gagal menghitung video")
		return
	}

	var videos []model.Video
	err := base.Distinct().
		Preload("Categories").Preload("Genres").
		Order("videos.created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&videos).Error
	if err != nil {
		response.Error(c, 500, "Gagal memuat video")
		return
	}

	response.Paginated(c, videos, total, page, limit)
}

// Get mengembalikan detail satu video yang published sambil menaikkan penghitung
// tayangan.
func (h *VideoHandler) Get(c *gin.Context) {
	var video model.Video
	err := h.db.Preload("Categories").Preload("Genres").
		Where("is_published = ?", true).
		First(&video, c.Param("id")).Error
	if err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}

	if err := h.db.Model(&video).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error; err != nil {
		// Penghitung tayangan bersifat kosmetik: kegagalan tidak boleh
		// membatalkan response.
		log.Printf("[video.get] gagal menambah views id=%s: %v", c.Param("id"), err)
	} else {
		video.Views++
	}

	response.OK(c, video)
}

// Featured mengembalikan video published yang paling banyak ditonton.
func (h *VideoHandler) Featured(c *gin.Context) {
	var videos []model.Video
	err := h.db.Preload("Categories").Preload("Genres").
		Where("is_published = ?", true).
		Order("views DESC").Limit(6).
		Find(&videos).Error
	if err != nil {
		response.Error(c, 500, "Gagal memuat video unggulan")
		return
	}
	response.OK(c, videos)
}

// AdminList mengembalikan seluruh video termasuk yang draft.
//
// Endpoint ini sengaja tidak dipaginasi: panel admin membutuhkan daftar penuh
// agar bisa difilter di sisi klien. Batas keras jumlah baris mencegah response
// membengkak tanpa batas.
func (h *VideoHandler) AdminList(c *gin.Context) {
	var videos []model.Video
	if err := h.db.Preload("Categories").Preload("Genres").
		Order("created_at DESC").Limit(adminListMaxRows).
		Find(&videos).Error; err != nil {
		response.Error(c, 500, "Gagal memuat video")
		return
	}

	var total int64
	if err := h.db.Model(&model.Video{}).Count(&total).Error; err != nil {
		response.Error(c, 500, "Gagal menghitung video")
		return
	}

	response.Paginated(c, videos, total, 1, adminListMaxRows)
}

// adminListMaxRows adalah batas keras baris untuk daftar admin.
const adminListMaxRows = 500

// AdminGet mengembalikan satu video tanpa memfilter status published.
func (h *VideoHandler) AdminGet(c *gin.Context) {
	var video model.Video
	err := h.db.Preload("Categories").Preload("Genres").First(&video, c.Param("id")).Error
	if err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	response.OK(c, video)
}

// Create membuat video baru beserta taksonominya.
func (h *VideoHandler) Create(c *gin.Context) {
	var req videoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid: "+err.Error())
		return
	}
	if req.Title == nil || strings.TrimSpace(*req.Title) == "" {
		response.Error(c, 400, "Judul wajib diisi")
		return
	}

	log.Printf("[video.create] allow_download=%v categories=%v genres=%v",
		ptrBool(req.AllowDownload), ptrSliceLen(req.Categories), ptrSliceLen(req.Genres))

	video := model.Video{Title: strings.TrimSpace(*req.Title)}
	applyVideoRequest(&video, req)

	if err := h.db.Create(&video).Error; err != nil {
		response.Error(c, 500, "Gagal menyimpan video")
		return
	}
	if err := h.applyTaxonomy(&video, req); err != nil {
		response.Error(c, 500, "Gagal menyimpan genre/category")
		return
	}
	if err := h.db.Preload("Categories").Preload("Genres").First(&video, video.ID).Error; err != nil {
		response.Error(c, 500, "Gagal memuat video setelah disimpan")
		return
	}

	response.Created(c, video)
}

// Update mengubah field video yang dikirim pada request.
func (h *VideoHandler) Update(c *gin.Context) {
	var video model.Video
	if err := h.db.First(&video, c.Param("id")).Error; err != nil {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}

	var req videoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			response.Error(c, 400, "Judul tidak boleh kosong")
			return
		}
		req.Title = &trimmed
	}

	log.Printf("[video.update] id=%s allow_download=%v categories=%v genres=%v",
		c.Param("id"), ptrBool(req.AllowDownload), ptrSliceLen(req.Categories), ptrSliceLen(req.Genres))

	applyVideoRequest(&video, req)
	if err := h.db.Save(&video).Error; err != nil {
		response.Error(c, 500, "Gagal menyimpan video")
		return
	}
	if err := h.applyTaxonomy(&video, req); err != nil {
		response.Error(c, 500, "Gagal menyimpan genre/category")
		return
	}
	if err := h.db.Preload("Categories").Preload("Genres").First(&video, video.ID).Error; err != nil {
		response.Error(c, 500, "Gagal memuat video setelah disimpan")
		return
	}

	response.OK(c, video)
}

// Delete menghapus video (soft delete lewat gorm.DeletedAt).
func (h *VideoHandler) Delete(c *gin.Context) {
	result := h.db.Delete(&model.Video{}, c.Param("id"))
	if result.Error != nil || result.RowsAffected == 0 {
		response.Error(c, 404, "Video tidak ditemukan")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

// ListCategories mengembalikan daftar category untuk form admin.
func (h *VideoHandler) ListCategories(c *gin.Context) {
	var items []model.Category
	q := h.db.Model(&model.Category{})
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		q = q.Where("name LIKE ?", "%"+search+"%")
	}
	if err := q.Order("name ASC").Limit(taxonomyListMaxRows).Find(&items).Error; err != nil {
		response.Error(c, 500, "Gagal memuat category")
		return
	}
	response.OK(c, items)
}

// ListGenres mengembalikan daftar genre untuk form admin.
func (h *VideoHandler) ListGenres(c *gin.Context) {
	var items []model.Genre
	q := h.db.Model(&model.Genre{})
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		q = q.Where("name LIKE ?", "%"+search+"%")
	}
	if err := q.Order("name ASC").Limit(taxonomyListMaxRows).Find(&items).Error; err != nil {
		response.Error(c, 500, "Gagal memuat genre")
		return
	}
	response.OK(c, items)
}

// taxonomyListMaxRows adalah batas baris untuk daftar category/genre.
const taxonomyListMaxRows = 50
