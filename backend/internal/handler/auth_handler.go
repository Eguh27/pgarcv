package handler

import (
	"crypto/subtle"
	"net/http"
	"time"

	"pgarcv/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Masa berlaku token admin dan cookie session.
const (
	tokenTTL     = 24 * time.Hour
	cookieMaxAge = 86400 // detik, setara tokenTTL
)

// AuthHandler melayani login dan logout admin.
type AuthHandler struct {
	username  string
	password  string
	jwtSecret string
	secure    bool
}

// NewAuthHandler membuat AuthHandler.
func NewAuthHandler(username, password, jwtSecret string, secure bool) *AuthHandler {
	return &AuthHandler{username: username, password: password, jwtSecret: jwtSecret, secure: secure}
}

// Login memverifikasi kredensial admin lalu menerbitkan token JWT dalam cookie
// httpOnly.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Data tidak valid")
		return
	}

	// Perbandingan constant-time agar waktu respons tidak membocorkan informasi
	// tentang mana antara username dan password yang salah.
	userOK := subtle.ConstantTimeCompare([]byte(req.Username), []byte(h.username)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(req.Password), []byte(h.password)) == 1
	if !userOK || !passOK {
		response.Error(c, http.StatusUnauthorized, "Username atau password salah")
		return
	}

	tokenStr, err := h.issueToken(req.Username)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membuat token")
		return
	}

	h.setAuthCookie(c, tokenStr, cookieMaxAge)
	response.OK(c, gin.H{
		"token":   tokenStr,
		"message": "Login berhasil",
	})
}

// Logout menghapus cookie token admin.
func (h *AuthHandler) Logout(c *gin.Context) {
	// MaxAge negatif membuat cookie langsung kedaluwarsa di sisi browser.
	h.setAuthCookie(c, "", -1)
	response.OK(c, gin.H{"message": "Logout berhasil"})
}

// issueToken membuat token JWT bertanda tangan HS256 untuk satu username.
func (h *AuthHandler) issueToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(tokenTTL).Unix(),
	})
	return token.SignedString([]byte(h.jwtSecret))
}

// setAuthCookie menulis cookie admin_token dengan atribut keamanan yang seragam
// untuk login maupun logout.
func (h *AuthHandler) setAuthCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("admin_token", value, maxAge, "/", "", h.secure, true)
}
