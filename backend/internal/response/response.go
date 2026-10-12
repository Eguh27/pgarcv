// Package response menyediakan pembungkus seragam untuk menulis JSON response
// dari handler.
package response

import "github.com/gin-gonic/gin"

// Response adalah bentuk payload sukses.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse adalah bentuk payload untuk daftar berpaginasi.
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// OK menulis 200 OK dengan data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Response{Success: true, Data: data})
}

// Created menulis 201 Created dengan data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(201, Response{Success: true, Data: data})
}

// Error menulis payload gagal dengan status code dan pesan.
func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, Response{Success: false, Message: msg})
}

// Paginated menulis 200 OK untuk daftar berpaginasi.
//
// limit dinormalisasi ke minimal 1 supaya TotalPages tidak pernah membagi
// dengan nol saat data kosong.
func Paginated(c *gin.Context, data interface{}, total int64, page, limit int) {
	if limit < 1 {
		limit = 1
	}
	if page < 1 {
		page = 1
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	c.JSON(200, PaginatedResponse{
		Success:    true,
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}
