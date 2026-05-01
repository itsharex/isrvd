// Package helper 提供操作审计日志工具函数
package helper

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rehiy/pango/logman"
)

// AuditLog 操作审计日志结构
type AuditLog struct {
	Timestamp  time.Time `json:"timestamp"`  // 操作时间
	Username   string    `json:"username"`   // 操作人
	Method     string    `json:"method"`     // HTTP 方法
	URI        string    `json:"uri"`        // 请求 URI
	Body       string    `json:"body"`       // 请求 Body（截断至 512 字节）
	IP         string    `json:"ip"`         // 客户端 IP
	StatusCode int       `json:"statusCode"` // 响应状态码
	Success    bool      `json:"success"`    // 是否成功（2xx）
	Duration   int64     `json:"duration"`   // 耗时（毫秒）
}

// auditLogBuffer 内存审计日志缓冲区
var (
	auditLogBuffer []AuditLog
	auditLogMutex  sync.RWMutex
	maxBufferSize  = 1000
)

func init() {
	auditLogBuffer = make([]AuditLog, 0, maxBufferSize)
}

// AddAuditLog 添加审计日志到缓冲区，超出上限时移除最旧记录
func AddAuditLog(log AuditLog) {
	auditLogMutex.Lock()
	defer auditLogMutex.Unlock()
	if len(auditLogBuffer) >= maxBufferSize {
		auditLogBuffer = auditLogBuffer[1:]
	}
	auditLogBuffer = append(auditLogBuffer, log)
}

// GetAuditLogs 获取审计日志，按时间倒序返回，支持按用户名过滤
func GetAuditLogs(username string, limit int) []AuditLog {
	auditLogMutex.RLock()
	defer auditLogMutex.RUnlock()

	var result []AuditLog
	for i := len(auditLogBuffer) - 1; i >= 0; i-- {
		entry := auditLogBuffer[i]
		if username != "" && entry.Username != username {
			continue
		}
		result = append(result, entry)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

// AuditMiddleware 操作审计中间件，记录所有非 GET 请求的 method/uri/body/状态/耗时
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		startTime := time.Now()

		// 读取并回填 body（最多 512 字节）
		body := ""
		if c.Request.Body != nil {
			raw, _ := io.ReadAll(io.LimitReader(c.Request.Body, 512))
			body = string(raw)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
		}

		c.Next()

		statusCode := c.Writer.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		entry := AuditLog{
			Timestamp:  startTime,
			Username:   c.GetString("username"),
			Method:     c.Request.Method,
			URI:        c.Request.RequestURI,
			Body:       body,
			IP:         c.ClientIP(),
			StatusCode: statusCode,
			Success:    statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices,
			Duration:   time.Since(startTime).Milliseconds(),
		}

		AddAuditLog(entry)

		logman.Info("Audit",
			"username", entry.Username,
			"method", entry.Method,
			"uri", entry.URI,
			"ip", entry.IP,
			"status", entry.StatusCode,
			"duration", entry.Duration,
		)
	}
}
