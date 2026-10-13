package handler

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"pgarcv/internal/middleware"
	"pgarcv/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// itoa memformat ID uint sebagai string untuk path parameter.
func itoa(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

// setupTestDB membuat database sementara dengan tabel yang dipakai test.
//
// Memakai file di t.TempDir(), bukan ":memory:", karena mode memory pada SQLite
// memberi SETIAP koneksi di pool database yang berbeda yang membuat test
// bisa gagal dengan "no such table" secara tidak-deterministik.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "handler-test.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Video{}, &model.Category{}, &model.Genre{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

// seedVideo menyisipkan satu video dan mengembalikan ID-nya.
func seedVideo(t *testing.T, db *gorm.DB, published bool) model.Video {
	t.Helper()
	video := model.Video{Title: "test", IsPublished: published}
	if err := db.Create(&video).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	return video
}

// newCtx membuat context request dengan flag admin sesuai isAdmin.
func newCtx(isAdmin bool) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if isAdmin {
		c.Set(middleware.ContextIsAdmin, true)
	}
	return c
}

// TestCanAccessVideo memverifikasi aturan akses: video published terbuka untuk
// semua orang, video draft hanya untuk admin.
func TestCanAccessVideo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	published := seedVideo(t, db, true)
	draft := seedVideo(t, db, false)
	h := NewHLSHandler(db, t.TempDir(), "", "http://localhost:8080")

	t.Run("video publik bisa diakses tanpa auth", func(t *testing.T) {
		v, allowed := h.canAccessVideo(newCtx(false), published.ID)
		if !allowed || v == nil || v.ID != published.ID {
			t.Fatalf("allowed=%v v=%v", allowed, v)
		}
	})

	t.Run("draft ditolak tanpa auth", func(t *testing.T) {
		v, allowed := h.canAccessVideo(newCtx(false), draft.ID)
		if allowed || v == nil {
			t.Fatalf("allowed=%v harusnya false (video tetap dikembalikan utk pesan)", allowed)
		}
	})

	t.Run("draft boleh untuk admin terautentikasi", func(t *testing.T) {
		v, allowed := h.canAccessVideo(newCtx(true), draft.ID)
		if !allowed || v == nil {
			t.Fatalf("allowed=%v v=%v", allowed, v)
		}
	})

	t.Run("video tidak ada mengembalikan nil,false", func(t *testing.T) {
		v, allowed := h.canAccessVideo(newCtx(true), 99999)
		if allowed || v != nil {
			t.Fatalf("allowed=%v v=%v, want false,nil", allowed, v)
		}
	})
}

// TestCanAccessVideoUsesMiddlewareContextKey adalah guard regresi: flag admin
// harus dibaca dari kunci yang sama dengan yang ditulis middleware.OptionalAuth.
//
// Versi sebelumnya membaca kunci "authenticated" yang tidak pernah di-set,
// sehingga preview video draft oleh admin selalu ditolak.
func TestCanAccessVideoUsesMiddlewareContextKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	draft := seedVideo(t, db, false)
	h := NewHLSHandler(db, t.TempDir(), "", "http://localhost:8080")

	// Simulasikan persis apa yang ditulis middleware.OptionalAuth.
	c := newCtx(false)
	c.Set(middleware.ContextIsAdmin, true)

	if _, allowed := h.canAccessVideo(c, draft.ID); !allowed {
		t.Fatal("draft harus boleh diakses dengan ContextIsAdmin")
	}

	// Kunci lama yang tidak lagi dipakai harus tidak memberi akses.
	cLegacy := newCtx(false)
	cLegacy.Set("authenticated", true)
	if _, allowed := h.canAccessVideo(cLegacy, draft.ID); allowed {
		t.Fatal(`kunci "authenticated" tidak lagi boleh memberi akses`)
	}
}

// TestServeKey memeriksa guard referer dan aturan publish pada endpoint kunci.
func TestServeKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	published := seedVideo(t, db, true)
	draft := seedVideo(t, db, false)

	const testHexKey = "0123456789abcdef0123456789abcdef" // 16 byte AES-128
	h := NewHLSHandler(db, t.TempDir(), testHexKey, "http://localhost:8080")

	serveKeyRequest := func(referer, videoID string) *httptest.ResponseRecorder {
		t.Helper()
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

	t.Run("referer asing ditolak", func(t *testing.T) {
		w := serveKeyRequest("https://evil.example.com/watch", "1")
		if w.Code != http.StatusForbidden {
			t.Fatalf("got %d want 403", w.Code)
		}
	})

	t.Run("referer IP LAN yang terdaftar presenting key", func(t *testing.T) {
		// Origin LAN harus diterima: hanya SERVER_URL + localhost yang
		// membuat video tidak bisa diputar dari HP.
		lan := NewHLSHandler(db, t.TempDir(), testHexKey, "http://localhost:8080",
			"http://localhost:3000,http://192.168.100.78:3000")
		r := gin.New()
		r.GET("/api/hls/key/:id", lan.ServeKey)
		req := httptest.NewRequest(http.MethodGet, "/api/hls/key/1", nil)
		req.Header.Set("Referer", "http://192.168.100.78:3000/watch/1")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("got %d want 200 untuk origin LAN terdaftar", w.Code)
		}
	})

	t.Run("origin LAN tidak terdaftar tetap ditolak", func(t *testing.T) {
		w := serveKeyRequest("http://10.0.0.9:3000/watch", "1")
		if w.Code != http.StatusForbidden {
			t.Fatalf("got %d want 403", w.Code)
		}
	})

	t.Run("referer server sendiri menyajikan key", func(t *testing.T) {
		w := serveKeyRequest("http://localhost:8080/watch/1", "1")
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

	t.Run("video draft ditolak tanpa konteks admin", func(t *testing.T) {
		// Request ini tidak melewati OptionalAuth, jadi tidak ada ContextIsAdmin.
		w := serveKeyRequest("http://localhost:8080/watch/2", itoa(draft.ID))
		if w.Code != http.StatusForbidden {
			t.Fatalf("got %d want 403", w.Code)
		}
	})

	t.Run("video draft menyajikan key untuk admin", func(t *testing.T) {
		r := gin.New()
		r.GET("/api/hls/key/:id", h.ServeKey)
		req := httptest.NewRequest(http.MethodGet, "/api/hls/key/"+itoa(draft.ID), nil)
		req.Header.Set("Referer", "http://localhost:8080/watch/1")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set(middleware.ContextIsAdmin, true)
		c.Params = gin.Params{{Key: "id", Value: itoa(draft.ID)}}
		h.ServeKey(c)
		if w.Code != http.StatusOK {
			t.Fatalf("got %d want 200 untuk admin", w.Code)
		}
	})

	t.Run("id bukan angka ditolak", func(t *testing.T) {
		w := serveKeyRequest("http://localhost:8080/x", "abc")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got %d want 400", w.Code)
		}
	})

	t.Run("video tidak ada ditolak", func(t *testing.T) {
		w := serveKeyRequest("http://localhost:8080/x", "99999")
		if w.Code != http.StatusNotFound {
			t.Fatalf("got %d want 404", w.Code)
		}
	})

	_ = published
}

// TestServeSegmentRejectsTraversal memastikan parameter segment tidak bisa
// keluar dari direktori output video.
func TestServeSegmentRejectsTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	published := seedVideo(t, db, true)
	h := NewHLSHandler(db, t.TempDir(), "", "http://localhost:8080")

	for _, seg := range []string{"../../secret.ts", "..", "sub/dir.ts", ".hidden"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Params = gin.Params{
			{Key: "id", Value: itoa(published.ID)},
			{Key: "segment", Value: seg},
		}
		h.ServeSegment(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("segment %q diterima: got %d want 400", seg, w.Code)
		}
	}
}

// TestVideoDownloadRejectsTraversal memastikan video_url lokal tidak bisa
// menunjuk ke file di luar uploadPath.
func TestVideoDownloadRejectsTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	video := model.Video{
		Title:         "jahat",
		VideoURL:      "/uploads/../../etc/passwd",
		IsPublished:   true,
		AllowDownload: true,
	}
	if err := db.Create(&video).Error; err != nil {
		t.Fatal(err)
	}

	h := NewVideoDownloadHandler(db, t.TempDir())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: itoa(video.ID)}}

	h.Download(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("traversal diterima: got %d want 400", w.Code)
	}
}

// TestSanitizeFilename memverifikasi judul video disanitasi sebelum dipakai
// sebagai nama file pada header Content-Disposition.
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Video Baru 2024", "Video Baru 2024"},
		{`../etc/passwd`, "etcpasswd"},
		{`a"; drop`, "a drop"},
		{"", "video"},
		{"***", "video"},
	}
	for _, tt := range tests {
		if got := sanitizeFilename(tt.in); got != tt.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// --- relativizeKeyURI ---

func TestRelativizeKeyURI(t *testing.T) {
	t.Run("URI absolut diganti path relatif", func(t *testing.T) {
		in := []byte("#EXTM3U\n" +
			"#EXT-X-KEY:METHOD=AES-128,URI=\"http://localhost:8080/api/hls/key/4\",IV=0x00\n" +
			"#EXTINF:10.0,\nseg_000.ts\n")
		out := string(relativizeKeyURI(in, 4))
		if !strings.Contains(out, `URI="/api/hls/key/4"`) {
			t.Fatalf("URI belum di-rewrite: %s", out)
		}
		if strings.Contains(out, "localhost") {
			t.Fatalf("masih ada host absolut: %s", out)
		}
	})

	t.Run("METHOD dan IV tetap utuh", func(t *testing.T) {
		in := []byte(`#EXT-X-KEY:METHOD=AES-128,URI="http://example.com/k",IV=0xDEADBEEF`)
		out := string(relativizeKeyURI(in, 7))
		if !strings.Contains(out, "METHOD=AES-128") || !strings.Contains(out, "IV=0xDEADBEEF") {
			t.Fatalf("atribut lain ikut berubah: %s", out)
		}
	})

	t.Run("manifest tanpa EXT-X-KEY tidak diubah", func(t *testing.T) {
		in := []byte("#EXTM3U\n#EXTINF:10.0,\nseg_000.ts\n")
		if string(relativizeKeyURI(in, 1)) != string(in) {
			t.Fatal("manifest tanpa key ikut berubah")
		}
	})

	t.Run("semua key di-multi-variant di-rewrite", func(t *testing.T) {
		in := []byte(`#EXT-X-KEY:METHOD=AES-128,URI="http://a/k1"` + "\n" +
			`#EXT-X-KEY:METHOD=AES-128,URI="http://a/k2"`)
		out := string(relativizeKeyURI(in, 9))
		if strings.Contains(out, "http://") {
			t.Fatalf("masih ada URI absolut: %s", out)
		}
		if strings.Count(out, `URI="/api/hls/key/9"`) != 2 {
			t.Fatalf("harus dua key ter-rewrite: %s", out)
		}
	})
}
