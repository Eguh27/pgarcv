package handler

import (
	"strconv"
	"strings"

	"pgarcv/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// videoRequest adalah payload create/update video.
//
// Seluruh field berupa pointer supaya update bersifat partial: field yang tidak
// dikirim tidak menimpa nilai lama. Create mewajibkan Title.
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

// applyVideoRequest menyalin field yang dikirim pada req ke video.
func applyVideoRequest(video *model.Video, req videoRequest) {
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

// applyTaxonomy menyinkronkan relasi many2many Video↔Category dan Video↔Genre.
//
// Nama yang belum ada dibuat otomatis. Pencocokan bersifat case-insensitive
// supaya "Tutorial" dan "tutorial" tidak tercipta dua kali. Setelah relasi
// disimpan, field legacy Category/Genre disinkronkan ke elemen pertama agar
// klien lama yang masih membaca kolom tunggal tidak kosong.
func (h *VideoHandler) applyTaxonomy(video *model.Video, req videoRequest) error {
	if req.Categories == nil && req.Genres == nil {
		return nil
	}

	if req.Categories != nil {
		cats, err := h.resolveCategories(*req.Categories)
		if err != nil {
			return err
		}
		if err := h.db.Model(video).Association("Categories").Replace(&cats); err != nil {
			return err
		}
		video.Category = firstOrEmpty(categoryNames(cats))
	}

	if req.Genres != nil {
		gens, err := h.resolveGenres(*req.Genres)
		if err != nil {
			return err
		}
		if err := h.db.Model(video).Association("Genres").Replace(&gens); err != nil {
			return err
		}
		video.Genre = firstOrEmpty(genreNames(gens))
	}

	// Persistensi sinkronisasi kolom legacy, bukan hanya di memori.
	return h.db.Model(video).Updates(map[string]any{
		"category": video.Category,
		"genre":    video.Genre,
	}).Error
}

// resolveCategories mencari category berdasarkan nama, membuat yang belum ada.
func (h *VideoHandler) resolveCategories(names []string) ([]model.Category, error) {
	var out []model.Category
	for _, raw := range names {
		name := normalizeName(raw)
		if name == "" {
			continue
		}
		cat, err := h.findOrCreateCategory(name)
		if err != nil {
			return nil, err
		}
		out = append(out, *cat)
	}
	return out, nil
}

// resolveGenres mencari genre berdasarkan nama, membuat yang belum ada.
func (h *VideoHandler) resolveGenres(names []string) ([]model.Genre, error) {
	var out []model.Genre
	for _, raw := range names {
		name := normalizeName(raw)
		if name == "" {
			continue
		}
		genre, err := h.findOrCreateGenre(name)
		if err != nil {
			return nil, err
		}
		out = append(out, *genre)
	}
	return out, nil
}

// findOrCreateCategory mengembalikan category dengan nama tersebut, membuatnya
// bila belum ada. Pencocokan case-insensitive.
func (h *VideoHandler) findOrCreateCategory(name string) (*model.Category, error) {
	var cat model.Category
	err := h.db.Where("lower(name) = lower(?)", name).First(&cat).Error
	switch {
	case err == nil:
		return &cat, nil
	case err == gorm.ErrRecordNotFound:
		cat = model.Category{Name: name}
		if err := h.db.Create(&cat).Error; err != nil {
			return nil, err
		}
		return &cat, nil
	default:
		return nil, err
	}
}

// findOrCreateGenre mengembalikan genre dengan nama tersebut, membuatnya bila
// belum ada. Pencocokan case-insensitive.
func (h *VideoHandler) findOrCreateGenre(name string) (*model.Genre, error) {
	var genre model.Genre
	err := h.db.Where("lower(name) = lower(?)", name).First(&genre).Error
	switch {
	case err == nil:
		return &genre, nil
	case err == gorm.ErrRecordNotFound:
		genre = model.Genre{Name: name}
		if err := h.db.Create(&genre).Error; err != nil {
			return nil, err
		}
		return &genre, nil
	default:
		return nil, err
	}
}

func categoryNames(cats []model.Category) []string {
	names := make([]string, 0, len(cats))
	for _, c := range cats {
		names = append(names, c.Name)
	}
	return names
}

func genreNames(gens []model.Genre) []string {
	names := make([]string, 0, len(gens))
	for _, g := range gens {
		names = append(names, g.Name)
	}
	return names
}

func firstOrEmpty(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

// normalizeName merapikan nama taxonomy: trim dan runtuhkan spasi ganda.
func normalizeName(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// queryInt membaca parameter query sebagai integer, dengan fallback saat
// parameter hilang atau bukan angka.
func queryInt(c *gin.Context, key string, fallback int) int {
	raw := c.DefaultQuery(key, "")
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

// ptrBool membuka nilai pointer bool untuk keperluan logging.
func ptrBool(b *bool) any {
	if b == nil {
		return nil
	}
	return *b
}

// ptrSliceLen membuka jumlah elemen slice untuk keperluan logging.
func ptrSliceLen(s *[]string) any {
	if s == nil {
		return nil
	}
	return len(*s)
}
