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

// 审计日志全局状态。
var (
	auditLogBuffer  []AuditLog
	auditLogMutex   sync.RWMutex
	auditLogOnce    sync.Once
	auditLogChan    chan AuditLog
	auditLogFile    *os.File
	auditLogDateKey string // 当前日志文件对应的日期（YYYY-MM-DD）
)

// ---- 公开函数 ----------------------------------------------------------------

// AuditMiddleware 返回操作审计中间件。
// 记录所有非 GET 请求及 WebSocket 连接的方法、URI、请求体、状态码和耗时。
func AuditMiddleware() gin.HandlerFunc {
	initAuditLog()

	return func(c *gin.Context) {
		// WebSocket 升级请求：记录连接建立时间与持续时长
		if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			startTime := time.Now()
			c.Next()
			statusCode := c.Writer.Status()
			if statusCode == 0 {
				statusCode = http.StatusSwitchingProtocols
			}
			AddAuditLog(AuditLog{
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

		// GET 请求不记录审计（必须在 WebSocket 判断之后）
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		startTime := time.Now()
		body := readBody(c)
		c.Next()

		statusCode := c.Writer.Status()
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		AddAuditLog(AuditLog{
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
}

// AddAuditLog 将审计条目写入内存缓冲，并异步追加到当日日志文件。
func AddAuditLog(entry AuditLog) {
	auditLogMutex.Lock()
	if len(auditLogBuffer) >= maxAuditBufferSize {
		// 重新分配以释放底层数组，避免内存泄漏
		newBuf := make([]AuditLog, maxAuditBufferSize-1, maxAuditBufferSize)
		copy(newBuf, auditLogBuffer[1:])
		auditLogBuffer = newBuf
	}
	auditLogBuffer = append(auditLogBuffer, entry)
	auditLogMutex.Unlock()

	if auditLogChan != nil {
		select {
		case auditLogChan <- entry:
		default:
			logman.Warn("Audit", "msg", "审计日志通道已满，丢弃文件写入")
		}
	}
}

// GetAuditLogs 返回内存缓冲中的审计日志，按时间倒序排列。
// username 非空时仅返回该用户的记录；limit <= 0 时返回全部。
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

// ---- 内部函数 ----------------------------------------------------------------

// initAuditLog 初始化审计日志（仅执行一次）：
// 加载今日历史记录到内存缓冲，打开文件用于追加写入，并启动异步写入协程。
func initAuditLog() {
	auditLogOnce.Do(func() {
		auditLogBuffer = make([]AuditLog, 0, maxAuditBufferSize)
		auditLogChan = make(chan AuditLog, maxAuditBufferSize)

		today := time.Now().Format("2006-01-02")
		loadAuditFile(auditFilePath(today))
		openAuditFile(today)

		go processAuditLogs()
	})
}

// processAuditLogs 从通道消费审计条目，按日期轮转文件并写入。
func processAuditLogs() {
	for entry := range auditLogChan {
		today := entry.Timestamp.Format("2006-01-02")
		if today != auditLogDateKey {
			openAuditFile(today)
		}
		if auditLogFile == nil {
			continue
		}
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		data = append(data, '\n')
		if _, err = auditLogFile.Write(data); err != nil {
			logman.Warn("Audit", "msg", "写入审计日志文件失败", "err", err)
		}
	}
}

// readBody 读取请求体并回填，对不同 Content-Type 做差异化处理：
//   - application/octet-stream：整体忽略，返回占位符
//   - multipart/form-data：保留文本字段，文件字段替换为占位符
//   - 其他：读取全部内容并回填 Body
func readBody(c *gin.Context) string {
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

// loadAuditFile 从指定路径读取日志文件，将最后 maxAuditBufferSize 条记录追加到内存缓冲。
func loadAuditFile(path string) {
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
	auditLogBuffer = append(auditLogBuffer, buf...)
}

// openAuditFile 打开指定日期的日志文件（追加模式），关闭旧文件。
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

// auditFilePath 返回指定日期的日志文件绝对路径。
func auditFilePath(date string) string {
	return filepath.Join(config.RootDirectory, "audit", date+".log")
}
