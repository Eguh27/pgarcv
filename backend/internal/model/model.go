// Package model berisi entitas domain yang dipetakan ke tabel database.
//
// Package ini sengaja dipisah dari repository: model hanya menyatakan bentuk
// data (struct + aturan mapping GORM), sedangkan repository memegang seluruh
// akses database. Handler mengimpor model, bukan repository, untuk membentuk
// request/response.
package model

import (
	"time"

	"gorm.io/gorm"
)

// Status nilai untuk Video.HLSStatus.
const (
	HLSStatusPending    = "pending"
	HLSStatusProcessing = "processing"
	HLSStatusDone       = "done"
	HLSStatusError      = "error"
)

// DeviceType untuk Ad.DeviceType.
const (
	DeviceTypeAll     = "all"
	DeviceTypeMobile  = "mobile"
	DeviceTypeDesktop = "desktop"
)

// Video adalah entitas utama platform: metadata, klasifikasi, dan status
// hasil transcode HLS.
type Video struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Title        string         `gorm:"not null" json:"title"`
	Subtitle     string         `json:"subtitle"`
	Description  string         `gorm:"type:text" json:"description"`
	ThumbnailURL string         `json:"thumbnail_url"`
	VideoURL     string         `json:"video_url"`
	PreviewURL   string         `json:"preview_url"`
	Duration     int            `json:"duration"` // detik
	// Category dan Genre adalah field lama (single value) yang dipertahankan
	// untuk backward compatibility dengan klien yang masih membacanya.
	Category string `json:"category"`
	Genre    string `json:"genre"`
	// Categories dan Genres adalah taksonomi baru (multi value).
	Categories    []Category `gorm:"many2many:video_categories;" json:"categories"`
	Genres        []Genre    `gorm:"many2many:video_genres;" json:"genres"`
	Views         int64      `gorm:"default:0" json:"views"`
	IsPublished   bool       `gorm:"default:false" json:"is_published"`
	AllowDownload bool       `gorm:"default:false" json:"allow_download"`
	IsEncrypted   bool       `gorm:"default:false" json:"is_encrypted"`
	// HLSPath dan RawVideoPath adalah path di disk — tidak pernah dikirim ke klien.
	HLSPath      string `json:"-"`
	RawVideoPath string `json:"-"`
	HLSStatus    string `gorm:"default:'pending'" json:"hls_status"` // pending|processing|done|error
	HLSErrorMsg  string `json:"hls_error_msg,omitempty"`
}

// Category adalah label klasifikasi video.
type Category struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
}

// Genre adalah label genre video.
type Genre struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
}

// Banner adalah banner promosi yang ditampilkan di beranda.
type Banner struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	ImageURL  string    `json:"image_url"`
	LinkURL   string    `json:"link_url"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
}

// Ad adalah slot iklan. DeviceType menentukan targeted platform.
type Ad struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AdCode     string    `gorm:"type:text" json:"ad_code"`
	DeviceType string    `gorm:"default:all" json:"device_type"` // mobile|desktop|all
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	Name       string    `json:"name"`
}

// ValidDeviceType melapor apakah deviceType adalah nilai yang dikenali.
func ValidDeviceType(deviceType string) bool {
	switch deviceType {
	case DeviceTypeAll, DeviceTypeMobile, DeviceTypeDesktop:
		return true
	default:
		return false
	}
}
