package handler

import (
	"net/http"
	"time"

	"pgarcv/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	Username  string
	Password  string
	JWTSecret string
	Secure    bool
}

func NewAuthHandler(username, password, secret string, secure bool) *AuthHandler {
	return &AuthHandler{Username: username, Password: password, JWTSecret: secret, Secure: secure}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Data tidak valid")
		return
	}
	if req.Username != h.Username || req.Password != h.Password {
		response.Error(c, 401, "Username atau password salah")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		response.Error(c, 500, "Gagal membuat token")
		return
	}

	// Set cookie httpOnly
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("admin_token", tokenStr, 86400, "/", "", h.Secure, true)
	response.OK(c, gin.H{
		"token":   tokenStr,
		"message": "Login berhasil",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("admin_token", "", -1, "/", "", h.Secure, true)
	response.OK(c, gin.H{"message": "Logout berhasil"})
}
