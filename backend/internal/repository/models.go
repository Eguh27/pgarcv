package repository

import (
	"time"

	"gorm.io/gorm"
)

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
	// Legacy single-value fields (kept for backward compatibility)
	Category string `json:"category"`
	Genre    string `json:"genre"`
	// New taxonomy (multi-value)
	Categories    []Category `gorm:"many2many:video_categories;" json:"categories"`
	Genres        []Genre    `gorm:"many2many:video_genres;" json:"genres"`
	Views         int64      `gorm:"default:0" json:"views"`
	IsPublished   bool       `gorm:"default:false" json:"is_published"`
	AllowDownload bool       `gorm:"default:false" json:"allow_download"`
	IsEncrypted   bool       `gorm:"default:false" json:"is_encrypted"`
	HLSPath       string     `json:"-"`
	RawVideoPath  string     `json:"-"`
	// ✅ HLS processing status tracking
	HLSStatus   string `gorm:"default:'pending'" json:"hls_status"` // pending, processing, done, error
	HLSErrorMsg string `json:"hls_error_msg,omitempty"`             // error message if HLS failed
}

type Category struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
}

type Genre struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
}

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

type Ad struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AdCode     string    `gorm:"type:text" json:"ad_code"`
	DeviceType string    `gorm:"default:all" json:"device_type"` // mobile | desktop | all
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	Name       string    `json:"name"`
}
