// Package middleware berisi middleware Gin untuk lintas-cutting concern
// seperti autentikasi.
package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"pgarcv/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Key yang ditulis middleware ke gin.Context.
const (
	ContextUsername = "username"
	ContextIsAdmin  = "is_admin"
)

// signingMethod mengunci algoritma penandatanganan ke HS256, sehingga token
// dengan algoritma "none" atau RS256 ditolak.
var signingMethod = jwt.SigningMethodHS256

// AuthRequired menolak request tanpa token JWT yang valid.
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			abortUnauthorized(c, "Unauthorized")
			return
		}

		username, err := parseToken(tokenStr, secret)
		if err != nil {
			abortUnauthorized(c, "Token tidak valid atau sudah expired")
			return
		}

		if username != "" {
			c.Set(ContextUsername, username)
		}
		c.Next()
	}
}

// OptionalAuth tidak menolak request tanpa token, tetapi menandai context
// sebagai admin bila token valid. Dipakai untuk endpoint yang bisa diakses
// publik maupun oleh admin (mis. preview HLS video draft).
func OptionalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tokenStr := extractToken(c); tokenStr != "" {
			if username, err := parseToken(tokenStr, secret); err == nil {
				if username != "" {
					c.Set(ContextUsername, username)
				}
				c.Set(ContextIsAdmin, true)
			}
		}
		c.Next()
	}
}

// parseToken memvalidasi tanda tangan dan masa berlaku token, lalu mengembalikan
// klaim username.
//
// Validasi eksplisit atas "exp" tetap dipertahankan (jwt v5 sudah memvalidasi
// expiry saat parse, tapi pengecekan eksplisit membuat intent jelas dan tidak
// bergantung pada perilaku library).
func parseToken(tokenStr, secret string) (username string, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != signingMethod.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{signingMethod.Alg()}))
	if err != nil || !token.Valid {
		return "", errors.New("token tidak valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("claims tidak valid")
	}

	exp, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		return "", errors.New("token sudah expired")
	}

	username, _ = claims["username"].(string)
	return username, nil
}

// abortUnauthorized menulis 401 lalu menghentikan rantai handler.
func abortUnauthorized(c *gin.Context, msg string) {
	response.Error(c, http.StatusUnauthorized, msg)
	c.Abort()
}

// extractToken membaca token dari cookie admin_token, lalu dari header
// Authorization: Bearer <token>.
func extractToken(c *gin.Context) string {
	if cookie, err := c.Cookie("admin_token"); err == nil && cookie != "" {
		return cookie
	}
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return ""
}
