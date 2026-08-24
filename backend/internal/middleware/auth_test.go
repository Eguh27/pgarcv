package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-minimal-32-karakter!!"

func makeToken(secret string, exp time.Time, username string) string {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      float64(exp.Unix()),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := t.SignedString([]byte(secret))
	return signed
}

func newRouter(mw gin.HandlerFunc, captured map[string]any) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw)
	r.GET("/protected", func(c *gin.Context) {
		if v, ok := c.Get("username"); ok {
			captured["username"] = v
		}
		if v, ok := c.Get("is_admin"); ok {
			captured["is_admin"] = v
		}
		c.Status(http.StatusOK)
	})
	return r
}

func doReq(r *gin.Engine, header string, cookie string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "admin_token", Value: cookie})
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthRequired(t *testing.T) {
	valid := makeToken(testSecret, time.Now().Add(time.Hour), "admin")
	expired := makeToken(testSecret, time.Now().Add(-time.Hour), "admin")
	otherSecret := makeToken("secret-lain-yang-beda-32-karakter!", time.Now().Add(time.Hour), "admin")

	tests := []struct {
		name   string
		header string
		cookie string
		want   int
	}{
		{"tanpa token", "", "", http.StatusUnauthorized},
		{"bearer valid", "Bearer " + valid, "", http.StatusOK},
		{"cookie valid", "", valid, http.StatusOK},
		{"bearer expired", "Bearer " + expired, "", http.StatusUnauthorized},
		{"signature salah", "Bearer " + otherSecret, "", http.StatusUnauthorized},
		{"format bukan bearer", "Token " + valid, "", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := map[string]any{}
			r := newRouter(AuthRequired(testSecret), captured)
			w := doReq(r, tt.header, tt.cookie)
			if w.Code != tt.want {
				t.Fatalf("got %d want %d", w.Code, tt.want)
			}
			if tt.want == http.StatusOK && captured["username"] != "admin" {
				t.Fatalf("username context = %v, want admin", captured["username"])
			}
		})
	}
}

func TestAuthRequiredRejectsNonHMAC(t *testing.T) {
	// Token RS256-like: gunakan alg none trick via unsigned token
	claims := jwt.MapClaims{"username": "admin", "exp": float64(time.Now().Add(time.Hour).Unix())}
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	r := newRouter(AuthRequired(testSecret), map[string]any{})
	if w := doReq(r, "Bearer "+signed, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("alg=none diterima, got %d want 401", w.Code)
	}
}

func TestOptionalAuth(t *testing.T) {
	valid := makeToken(testSecret, time.Now().Add(time.Hour), "admin")

	t.Run("tanpa token tetap lanjut", func(t *testing.T) {
		captured := map[string]any{}
		r := newRouter(OptionalAuth(testSecret), captured)
		if w := doReq(r, "", ""); w.Code != http.StatusOK {
			t.Fatalf("got %d want 200", w.Code)
		}
		if _, ok := captured["is_admin"]; ok {
			t.Fatal("is_admin tidak boleh diset tanpa token")
		}
	})

	t.Run("token invalid tetap lanjut tanpa is_admin", func(t *testing.T) {
		captured := map[string]any{}
		r := newRouter(OptionalAuth(testSecret), captured)
		if w := doReq(r, "Bearer invalid.token.here", ""); w.Code != http.StatusOK {
			t.Fatalf("got %d want 200", w.Code)
		}
		if _, ok := captured["is_admin"]; ok {
			t.Fatal("is_admin tidak boleh diset untuk token invalid")
		}
	})

	t.Run("token valid set is_admin", func(t *testing.T) {
		captured := map[string]any{}
		r := newRouter(OptionalAuth(testSecret), captured)
		if w := doReq(r, "Bearer "+valid, ""); w.Code != http.StatusOK {
			t.Fatalf("got %d want 200", w.Code)
		}
		if captured["is_admin"] != true || captured["username"] != "admin" {
			t.Fatalf("context = %v, want is_admin=true username=admin", captured)
		}
	})
}
