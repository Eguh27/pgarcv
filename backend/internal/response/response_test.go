package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// callPaginated menjalankan Paginated lalu mengembalikan status dan payload.
func callPaginated(t *testing.T, total int64, page, limit int) (int, PaginatedResponse) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	Paginated(c, []string{}, total, page, limit)

	var payload PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal %q: %v", w.Body.String(), err)
	}
	return w.Code, payload
}

// TestPaginatedTotalPagesLdividingByZero adalah guard regresi untuk divide-by-zero.
//
// Versi sebelumnya memanggil int(total)/limit tanpa proteksi, sehingga daftar kosong
// dengan limit=0 membuat server panic.
func TestPaginatedDoesNotDivideByZero(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		page      int
		limit     int
		wantLimit int
		wantPages int
	}{
		{"limit nol", 0, 1, 0, 1, 0},
		{"limit negatif", 0, 1, -5, 1, 0},
		{"tepat habis dibagi", 20, 2, 10, 10, 2},
		{"sisa pembagian", 21, 3, 10, 10, 3},
		{"nol total", 0, 1, 10, 10, 0},
		{"satu elemen", 1, 1, 10, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, payload := callPaginated(t, tt.total, tt.page, tt.limit)
			if code != http.StatusOK {
				t.Fatalf("status = %d want 200", code)
			}
			if payload.Limit != tt.wantLimit {
				t.Errorf("limit = %d want %d", payload.Limit, tt.wantLimit)
			}
			if payload.TotalPages != tt.wantPages {
				t.Errorf("total_pages = %d want %d", payload.TotalPages, tt.wantPages)
			}
		})
	}
}

// TestPaginatedClampsPage memastikan page minimal selalu 1.
func TestPaginatedClampsPage(t *testing.T) {
	_, payload := callPaginated(t, 100, -3, 10)
	if payload.Page != 1 {
		t.Errorf("page = %d want 1", payload.Page)
	}
}

// TestErrorOmitsData memastikan payload error tidak membawa field data.
func TestErrorOmitsData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	Error(c, http.StatusBadRequest, "gagal")

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d want 400", w.Code)
	}
	if raw["success"] != false {
		t.Errorf("success = %v want false", raw["success"])
	}
	if _, ok := raw["data"]; ok {
		t.Error("field data seharusnya tidak ada pada response error")
	}
}
