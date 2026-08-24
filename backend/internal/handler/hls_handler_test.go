package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pgarcv/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&repository.Video{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedVideo(t *testing.T, db *gorm.DB, id uint, published bool) {
	t.Helper()
	if err := db.Create(&repository.Video{Title: "test", IsPublished: published}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	var v repository.Video
	db.Last(&v)
	if v.ID == 0 || (id != 0 && v.ID != id) {
		t.Fatalf("unexpected seeded id %d", v.ID)
	}
}

// --- canAccessVideoForUser ---

func TestCanAccessVideoForUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	pub := &repository.Video{Title: "publik", IsPublished: true}
	draft := &repository.Video{Title: "draft", IsPublished: false}
	if err := db.Create(pub).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(draft).Error; err != nil {
		t.Fatal(err)
	}

	h := NewHLSHandler(db, t.TempDir(), "", "http://localhost:8080")

	newCtx := func(authenticated bool) (*gin.Context, *httptest.ResponseRecorder) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		if authenticated {
			c.Set("authenticated", true)
		}
		return c, w
	}

	t.Run("video publik bisa diakses tanpa auth", func(t *testing.T) {
		c, _ := newCtx(false)
		v, allowed := h.canAccessVideoForUser(c, pub.ID)
		if !allowed || v == nil || v.ID != pub.ID {
			t.Fatalf("allowed=%v v=%v", allowed, v)
		}
	})

	t.Run("draft ditolak tanpa auth", func(t *testing.T) {
		c, _ := newCtx(false)
		v, allowed := h.canAccessVideoForUser(c, draft.ID)
		if allowed || v == nil {
			t.Fatalf("allowed=%v harusnya false (video tetap dikembalikan utk pesan)", allowed)
		}
	})

	t.Run("draft boleh untuk admin terautentikasi", func(t *testing.T) {
		c, _ := newCtx(true)
		v, allowed := h.canAccessVideoForUser(c, draft.ID)
		if !allowed || v == nil {
			t.Fatalf("allowed=%v v=%v", allowed, v)
		}
	})

	t.Run("video tidak ada → nil,false", func(t *testing.T) {
		c, _ := newCtx(true)
		v, allowed := h.canAccessVideoForUser(c, 99999)
		if allowed || v != nil {
			t.Fatalf("allowed=%v v=%v, want false,nil", allowed, v)
		}
	})
}

// --- ServeKey ---

const testHexKey = "0123456789abcdef0123456789abcdef" // 16 byte AES-128

func serveKeyRequest(t *testing.T, h *HLSHandler, referer string, videoID string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/hls/key/:id", h.ServeKey)
	req := httptest.NewRequest(http.MethodGet, "/api/hls/key/"+videoID, nil)
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestServeKey(t *testing.T) {
	db := setupTestDB(t)
	seedVideo(t, db, 1, true)   // ID auto increment mulai 1
	seedVideo(t, db, 2, false)

	var pub, draft repository.Video
	db.First(&pub, 1)
	db.Last(&draft)

	h := NewHLSHandler(db, t.TempDir(), testHexKey, "http://localhost:8080")

	t.Run("referer asing → 403", func(t *testing.T) {
		w := serveKeyRequest(t, h, "https://evil.example.com/watch", "1")
		if w.Code != http.StatusForbidden {
			t.Fatalf("got %d want 403", w.Code)
		}
	})

	t.Run("referer server sendiri → key disajikan", func(t *testing.T) {
		w := serveKeyRequest(t, h, "http://localhost:8080/watch/1", "1")
		if w.Code != http.StatusOK {
			t.Fatalf("got %d want 200", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); ct != "application/octet-stream" {
			t.Fatalf("content-type = %s", ct)
		}
		if len(w.Body.Bytes()) != 16 {
			t.Fatalf("key len = %d, want 16 byte", len(w.Body.Bytes()))
		}
	})

	t.Run("draft video → 403 walau referer valid", func(t *testing.T) {
		w := serveKeyRequest(t, h, "http://localhost:8080/watch/2", fmt.Sprintf("%d", draft.ID))
		// Draft memang diblok di ServeKey (belum pakai OptionalAuth context)
		if w.Code != http.StatusForbidden {
			t.Fatalf("got %d want 403", w.Code)
		}
	})

	t.Run("id bukan angka → 400", func(t *testing.T) {
		w := serveKeyRequest(t, h, "http://localhost:8080/x", "abc")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got %d want 400", w.Code)
		}
	})

	_ = pub
}

// --- ServeSegment traversal guard ---

func TestServeSegmentRejectsTraversal(t *testing.T) {
	db := setupTestDB(t)
	seedVideo(t, db, 1, true)

	h := NewHLSHandler(db, t.TempDir(), "", "http://localhost:8080")
	gin.SetMode(gin.TestMode)

	for _, seg := range []string{"../../secret.ts", "..", "sub/dir.ts"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "segment", Value: seg},
		}
		h.ServeSegment(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("segment %q diterima: got %d want 400", seg, w.Code)
		}
	}
}
