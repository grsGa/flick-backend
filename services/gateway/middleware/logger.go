package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LoggerConfig 定义Logger中间件的配置
type LoggerConfig struct {
	// 跳过记录的路径
	SkipPaths []string
	// 是否记录请求体
	LogRequestBody bool
	// 是否记录响应体
	LogResponseBody bool
	// 最大记录体大小（字节）
	MaxBodySize int64
}

// DefaultLoggerConfig 返回默认的Logger配置
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		SkipPaths:       []string{"/health", "/metrics"},
		LogRequestBody:  true,
		LogResponseBody: true,
		MaxBodySize:     10240, // 10KB
	}
}

// responseBodyWriter 是一个自定义的响应写入器，用于捕获响应体
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 方法覆盖默认的写入方法，同时写入到原始响应和缓冲区
func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 方法实现 gin.ResponseWriter 接口
func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// WriteHeader 方法实现 gin.ResponseWriter 接口
func (w *responseBodyWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

// Status 返回当前请求的HTTP响应状态码
func (w *responseBodyWriter) Status() int {
	return w.ResponseWriter.Status()
}

// Size 返回已经写入响应体的字节数
func (w *responseBodyWriter) Size() int {
	return w.ResponseWriter.Size()
}

// Written 返回响应体是否已经写入
func (w *responseBodyWriter) Written() bool {
	return w.ResponseWriter.Written()
}

// WriteHeaderNow 强制立即写入HTTP头（状态码+头）
func (w *responseBodyWriter) WriteHeaderNow() {
	w.ResponseWriter.WriteHeaderNow()
}

// Pusher 返回http.Pusher用于服务器推送
func (w *responseBodyWriter) Pusher() http.Pusher {
	return w.ResponseWriter.Pusher()
}

// Hijack 实现http.Hijacker接口
func (w *responseBodyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush 实现http.Flusher接口
func (w *responseBodyWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

// Logger 中间件记录请求日志
func Logger(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过该路径
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if skipPath == path {
				c.Next()
				return
			}
		}

		// 开始时间
		startTime := time.Now()

		// 捕获请求体
		var requestBody string
		if config.LogRequestBody && c.Request.Body != nil && c.Request.ContentLength > 0 {
			// 限制读取的请求体大小
			maxSize := config.MaxBodySize
			if c.Request.ContentLength > 0 && c.Request.ContentLength < maxSize {
				maxSize = c.Request.ContentLength
			}

			// 读取请求体
			bodyBytes, _ := io.ReadAll(io.LimitReader(c.Request.Body, maxSize))
			if len(bodyBytes) > 0 {
				// 重置请求体，以便后续中间件可以继续读取
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				// 尝试将请求体格式化为JSON
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
					requestBody = prettyJSON.String()
				} else {
					requestBody = string(bodyBytes)
				}
			}
		}

		// 捕获响应体
		var respBodyWriter *responseBodyWriter
		if config.LogResponseBody {
			respBodyWriter = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:           &bytes.Buffer{},
			}
			c.Writer = respBodyWriter
		}

		// 处理请求
		c.Next()

		// 处理响应体
		var responseBody string
		if config.LogResponseBody && respBodyWriter != nil {
			// 获取响应体
			responseBodyBytes := respBodyWriter.body.Bytes()
			if len(responseBodyBytes) > 0 {
				// 尝试将响应体格式化为JSON
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, responseBodyBytes, "", "  "); err == nil {
					responseBody = prettyJSON.String()
				} else {
					responseBody = string(responseBodyBytes)
				}
			}
		}

		// 计算请求处理时间
		latency := time.Since(startTime)

		// 获取请求ID
		requestID := GetRequestID(c)

		// 创建日志条目
		logEntry := logrus.WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency.String(),
			"user_agent": c.Request.UserAgent(),
			"request_id": requestID,
			"error":      c.Errors.String(),
		})

		// 添加请求体和响应体到日志（如果启用）
		if config.LogRequestBody && requestBody != "" {
			logEntry = logEntry.WithField("request_body", requestBody)
		}

		if config.LogResponseBody && responseBody != "" {
			logEntry = logEntry.WithField("response_body", responseBody)
		}

		// 根据状态码记录不同级别的日志
		statusCode := c.Writer.Status()
		switch {
		case statusCode >= 500:
			logEntry.Error("服务器错误")
		case statusCode >= 400:
			logEntry.Warn("客户端错误")
		case statusCode >= 300:
			logEntry.Info("重定向")
		default:
			logEntry.Info("请求完成")
		}
	}
} 