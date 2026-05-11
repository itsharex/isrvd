package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse 通用API响应结构
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Payload any    `json:"payload,omitempty"`
}

// respondSuccess 返回成功响应
func respondSuccess(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Payload: data,
	})
}

// respondError 返回错误响应
func respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
	})
}
