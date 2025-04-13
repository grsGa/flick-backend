package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDKey 请求ID的上下文键
	RequestIDKey = "requestId"
	// RequestIDHeader 请求ID的HTTP头
	RequestIDHeader = "X-Request-ID"
	// TraceIDHeader 跟踪ID的HTTP头
	TraceIDHeader = "X-Trace-ID"
)

// RequestID 中间件生成唯一的请求ID并添加到上下文和响应头中
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取请求ID
		requestID := c.GetHeader(RequestIDHeader)

		// 如果请求头中没有请求ID，则生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 从请求头中获取跟踪ID
		traceID := c.GetHeader(TraceIDHeader)

		// 如果请求头中没有跟踪ID，则使用请求ID作为跟踪ID
		if traceID == "" {
			traceID = requestID
		}

		// 将请求ID和跟踪ID设置到上下文中
		c.Set(RequestIDKey, requestID)
		c.Set("traceId", traceID)

		// 将请求ID和跟踪ID添加到响应头中
		c.Header(RequestIDHeader, requestID)
		c.Header(TraceIDHeader, traceID)

		// 继续处理请求
		c.Next()
	}
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(c *gin.Context) string {
	requestID, exists := c.Get(RequestIDKey)
	if !exists {
		return ""
	}

	return requestID.(string)
}

// GetTraceID 从上下文中获取跟踪ID
func GetTraceID(c *gin.Context) string {
	traceID, exists := c.Get("traceId")
	if !exists {
		return ""
	}

	return traceID.(string)
}
