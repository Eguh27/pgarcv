package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

type CloudinaryStorage struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryStorage(cloudinaryURL string) (*CloudinaryStorage, error) {
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, fmt.Errorf("gagal init cloudinary: %w", err)
	}
	return &CloudinaryStorage{cld: cld}, nil
}

func (s *CloudinaryStorage) Upload(file multipart.File, filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	publicID := uuid.New().String()

	// Tentukan folder berdasarkan tipe file
	folder := "thumbnails"
	resourceType := "image"
	if ext == ".mp4" || ext == ".webm" || ext == ".mov" || ext == ".avi" {
		folder = "videos"
		resourceType = "video"
	}

	ctx := context.Background()
	result, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:     folder + "/" + publicID,
		ResourceType: resourceType,
	})
	if err != nil {
		return "", fmt.Errorf("gagal upload ke cloudinary: %w", err)
	}

	return result.SecureURL, nil
}
