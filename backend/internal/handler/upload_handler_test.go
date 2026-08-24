package handler

import "testing"

func TestAllowedUploadKind(t *testing.T) {
	tests := []struct {
		ext  string
		kind string
		ok   bool
	}{
		{".jpg", "image", true},
		{".jpeg", "image", true},
		{".png", "image", true},
		{".gif", "image", true},
		{".mp4", "video", true},
		{".webm", "video", true},
		{".mov", "video", true},
		{".avi", "video", true},
		{".exe", "", false},
		{".sh", "", false},
		{".php", "", false},
		{"", "", false},
		{".MP4", "", false}, // harus sudah dilowercase oleh pemanggil
	}
	for _, tt := range tests {
		kind, ok := allowedUploadKind(tt.ext)
		if kind != tt.kind || ok != tt.ok {
			t.Errorf("allowedUploadKind(%q) = (%q,%v), want (%q,%v)",
				tt.ext, kind, ok, tt.kind, tt.ok)
		}
	}
}

func TestValidUploadContentType(t *testing.T) {
	imageOK := []string{"image/jpeg", "image/png", "image/gif"}
	for _, ct := range imageOK {
		if !validUploadContentType("image", ct) {
			t.Errorf("image + %s harus valid", ct)
		}
	}
	videoOK := []string{"video/mp4", "video/webm", "video/x-msvideo", "application/octet-stream"}
	for _, ct := range videoOK {
		if !validUploadContentType("video", ct) {
			t.Errorf("video + %s harus valid", ct)
		}
	}
	bad := []struct {
		kind, ct string
	}{
		{"image", "video/mp4"},
		{"image", "text/html"},
		{"video", "text/html"},
		{"video", "application/javascript"},
		{"lain", "anything"},
	}
	for _, b := range bad {
		if validUploadContentType(b.kind, b.ct) {
			t.Errorf("%s + %s harus ditolak", b.kind, b.ct)
		}
	}
}
