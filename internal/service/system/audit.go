// Package system 提供系统级业务服务。
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

// AuditLog 操作审计日志条目。
type AuditLog struct {
	Timestamp  time.Time `json:"timestamp"`  // 操作时间
	Username   string    `json:"username"`   // 操作人
	Method     string    `json:"method"`     // HTTP 方法（WebSocket 记为 "WS"）
	URI        string    `json:"uri"`        // 请求 URI
	Body       string    `json:"body"`       // 请求体（文件字段替换为占位符）
	IP         string    `json:"ip"`         // 客户端 IP
	StatusCode int       `json:"statusCode"` // 响应状态码
	Success    bool      `json:"success"`    // 是否成功
	Duration   int64     `json:"duration"`   // 耗时（毫秒）
}

const maxAuditBufferSize = 100 // 内存缓冲最大条数

// AuditService 审计日志业务服务
type AuditService struct {
	mu      sync.RWMutex
	buffer  []AuditLog
	ch      chan AuditLog
	file    *os.File
	dateKey string // 当前日志文件对应的日期（YYYY-MM-DD）
}

// NewAuditService 创建审计日志业务服务并自动初始化
func NewAuditService() *AuditService {
	s := &AuditService{
		buffer: make([]AuditLog, 0, maxAuditBufferSize),
		ch:     make(chan AuditLog, maxAuditBufferSize),
	}

	today := time.Now().Format("2006-01-02")
	s.loadFile(auditFilePath(today))
	s.openFile(today)

	go s.process()

	return s
}

// Add 将审计条目写入内存缓冲，并异步追加到当日日志文件。
func (s *AuditService) Add(entry AuditLog) {
	s.mu.Lock()
	if len(s.buffer) >= maxAuditBufferSize {
		// 重新分配以释放底层数组，避免内存泄漏
		newBuf := make([]AuditLog, maxAuditBufferSize-1, maxAuditBufferSize)
		copy(newBuf, s.buffer[1:])
		s.buffer = newBuf
	}
	s.buffer = append(s.buffer, entry)
	s.mu.Unlock()

	select {
	case s.ch <- entry:
	default:
		logman.Warn("Audit", "msg", "审计日志通道已满，丢弃文件写入")
	}
}

// GetLogs 返回内存缓冲中的审计日志，按时间倒序排列。
// username 非空时仅返回该用户的记录；limit <= 0 时返回全部。
func (s *AuditService) GetLogs(username string, limit int) []AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []AuditLog
	for i := len(s.buffer) - 1; i >= 0; i-- {
		entry := s.buffer[i]
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

// RecordRequest 根据请求类型记录审计日志，供中间件在 c.Next() 后调用。
// WebSocket 升级请求记录 "WS" 方法；其余记录方法、URI、请求体、状态码。
func (s *AuditService) RecordRequest(c *gin.Context, startTime time.Time, body string) {
	// WebSocket
	if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
		statusCode := c.Writer.Status()
		if statusCode == 0 || statusCode == http.StatusOK {
			statusCode = http.StatusSwitchingProtocols
		}
		s.Add(AuditLog{
			Timestamp:  startTime,
			Username:   c.GetString("username"),
			Method:     "WS",
			URI:        c.Request.RequestURI,
			IP:         c.ClientIP(),
			StatusCode: statusCode,
			Success:    statusCode == http.StatusSwitchingProtocols,
			Duration:   time.Since(startTime).Milliseconds(),
		})
		return
	}

	statusCode := c.Writer.Status()
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	s.Add(AuditLog{
		Timestamp:  startTime,
		Username:   c.GetString("username"),
		Method:     c.Request.Method,
		URI:        c.Request.RequestURI,
		Body:       body,
		IP:         c.ClientIP(),
		StatusCode: statusCode,
		Success:    statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices,
		Duration:   time.Since(startTime).Milliseconds(),
	})
}

// ReadRequestBody 读取请求体并回填，对不同 Content-Type 做差异化处理：
//   - application/octet-stream：整体忽略，返回占位符
//   - multipart/form-data：保留文本字段，文件字段替换为占位符
//   - 其他：读取全部内容并回填 Body
func (s *AuditService) ReadRequestBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}

	contentType := c.ContentType()

	switch {
	case strings.HasPrefix(contentType, "application/octet-stream"):
		return "[Binary Omitted]"

	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil || c.Request.MultipartForm == nil {
			return ""
		}
		fields := make(map[string]any)
		for k, vs := range c.Request.MultipartForm.Value {
			if len(vs) == 1 {
				fields[k] = vs[0]
			} else {
				fields[k] = vs
			}
		}
		for k := range c.Request.MultipartForm.File {
			fields[k] = "[File Omitted]"
		}
		data, err := json.Marshal(fields)
		if err != nil {
			return ""
		}
		return string(data)

	default:
		raw, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))
		return string(raw)
	}
}

// ---- 内部方法 ----------------------------------------------------------------

// process 从通道消费审计条目，按日期轮转文件并写入。
func (s *AuditService) process() {
	for entry := range s.ch {
		today := entry.Timestamp.Format("2006-01-02")
		if today != s.dateKey {
			s.openFile(today)
		}
		if s.file == nil {
			continue
		}
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		data = append(data, '\n')
		if _, err = s.file.Write(data); err != nil {
			logman.Warn("Audit", "msg", "写入审计日志文件失败", "err", err)
		}
	}
}

// loadFile 从指定路径读取日志文件，将最后 maxAuditBufferSize 条记录追加到内存缓冲。
func (s *AuditService) loadFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	var buf []AuditLog
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry AuditLog
		if json.Unmarshal(scanner.Bytes(), &entry) == nil {
			buf = append(buf, entry)
		}
	}

	if len(buf) > maxAuditBufferSize {
		buf = buf[len(buf)-maxAuditBufferSize:]
	}
	s.buffer = append(s.buffer, buf...)
}

// openFile 打开指定日期的日志文件（追加模式），关闭旧文件。
func (s *AuditService) openFile(date string) {
	path := auditFilePath(date)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		logman.Warn("Audit", "msg", "无法创建审计日志目录", "err", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logman.Warn("Audit", "msg", "无法打开审计日志文件", "path", path, "err", err)
		return
	}

	if s.file != nil {
		s.file.Close()
	}
	s.file = f
	s.dateKey = date
}

// auditFilePath 返回指定日期的日志文件绝对路径。
func auditFilePath(date string) string {
	return filepath.Join(config.RootDirectory, "audit", date+".log")
}
