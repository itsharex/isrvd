// Package system 提供系统级业务服务
package system

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rehiy/pango/logman"

	"isrvd/config"
)

// AuditLog 操作审计日志结构
type AuditLog struct {
	Timestamp  time.Time `json:"timestamp"`  // 操作时间
	Username   string    `json:"username"`   // 操作人
	Method     string    `json:"method"`     // HTTP 方法
	URI        string    `json:"uri"`        // 请求 URI
	Body       string    `json:"body"`       // 请求 Body
	IP         string    `json:"ip"`         // 客户端 IP
	StatusCode int       `json:"statusCode"` // 响应状态码
	Success    bool      `json:"success"`    // 是否成功（2xx）
	Duration   int64     `json:"duration"`   // 耗时（毫秒）
}

const maxAuditBufferSize = 100 // 内存缓冲最大条数

var (
	auditLogBuffer  []AuditLog
	auditLogMutex   sync.RWMutex
	auditLogFile    *os.File
	auditLogOnce    sync.Once
	auditLogDateKey string        // 当前日志文件对应的日期（YYYY-MM-DD）
	auditLogChan    chan AuditLog // 异步写入通道
)

// AuditMiddleware 操作审计中间件，记录所有非 GET 请求的 method/uri/body/状态/耗时
func AuditMiddleware() gin.HandlerFunc {
	initAuditLog() // 在注册中间件时自动完成初始化

	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		startTime := time.Now()

		// 读取并回填 body
		body := ""
		if c.Request.Body != nil {
			contentType := c.ContentType()
			// 忽略文件上传类型的 body
			if strings.HasPrefix(contentType, "multipart/form-data") || strings.HasPrefix(contentType, "application/octet-stream") {
				body = "[File Upload Omitted]"
			} else {
				// 读取全部 body
				raw, _ := io.ReadAll(c.Request.Body)
				body = string(raw)
				// 将读取的内容重新放回 Body，供后续 Handler 使用
				c.Request.Body = io.NopCloser(bytes.NewReader(raw))
			}
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

// AddAuditLog 添加审计日志：写入内存缓冲并异步追加到当日文件
func AddAuditLog(entry AuditLog) {
	auditLogMutex.Lock()
	// 内存缓冲超限时移除最旧记录
	if len(auditLogBuffer) >= maxAuditBufferSize {
		auditLogBuffer = auditLogBuffer[1:]
	}
	auditLogBuffer = append(auditLogBuffer, entry)
	auditLogMutex.Unlock()

	// 异步发送到通道写入文件
	if auditLogChan != nil {
		select {
		case auditLogChan <- entry:
		default:
			logman.Warn("Audit", "msg", "审计日志通道已满，丢弃文件写入")
		}
	}
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

// initAuditLog 初始化审计日志：从今日文件加载历史记录到内存缓冲，并打开文件用于追加写入
func initAuditLog() {
	auditLogOnce.Do(func() {
		auditLogBuffer = make([]AuditLog, 0, maxAuditBufferSize)
		auditLogChan = make(chan AuditLog, maxAuditBufferSize)

		// 加载今日日志到内存缓冲
		today := time.Now().Format("2006-01-02")
		loadAuditFile(auditFilePath(today))

		// 打开今日文件用于追加写入
		openAuditFile(today)

		// 启动异步写入协程
		go processAuditLogs()
	})
}

// processAuditLogs 异步处理审计日志的文件写入和轮转
func processAuditLogs() {
	for entry := range auditLogChan {
		today := entry.Timestamp.Format("2006-01-02")
		if today != auditLogDateKey {
			openAuditFile(today)
		}

		if auditLogFile != nil {
			if data, err := json.Marshal(entry); err == nil {
				data = append(data, '\n')
				if _, err = auditLogFile.Write(data); err != nil {
					logman.Warn("Audit", "msg", "写入审计日志文件失败", "err", err)
				}
			}
		}
	}
}

// loadAuditFile 从指定文件读取日志追加到内存缓冲
func loadAuditFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	var tempBuffer []AuditLog
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry AuditLog
		if json.Unmarshal(scanner.Bytes(), &entry) == nil {
			tempBuffer = append(tempBuffer, entry)
		}
	}

	// 只保留最后 maxAuditBufferSize 条记录
	if len(tempBuffer) > maxAuditBufferSize {
		tempBuffer = tempBuffer[len(tempBuffer)-maxAuditBufferSize:]
	}
	auditLogBuffer = append(auditLogBuffer, tempBuffer...)
}

// openAuditFile 打开指定日期的日志文件用于追加写入，关闭旧文件
func openAuditFile(date string) {
	path := auditFilePath(date)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		logman.Warn("Audit", "msg", "无法创建审计日志目录", "err", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logman.Warn("Audit", "msg", "无法打开审计日志文件", "path", path, "err", err)
		return
	}

	if auditLogFile != nil {
		auditLogFile.Close()
	}

	auditLogFile = f
	auditLogDateKey = date
}

// auditFilePath 返回指定日期的日志文件路径
func auditFilePath(date string) string {
	dir := filepath.Join(config.RootDirectory, "audit")
	return filepath.Join(dir, date+".log")
}
