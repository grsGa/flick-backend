package observability

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

// LogLevel 日志级别
type LogLevel string

// 定义日志级别常量
const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarn    LogLevel = "warn"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
	LogLevelDisable LogLevel = "disable"
)

// LoggingConfig 日志配置
type LoggingConfig struct {
	// 日志级别
	Level LogLevel
	// 是否启用JSON格式
	EnableJSON bool
	// 是否启用控制台颜色
	EnableConsoleColor bool
	// 是否启用文件日志
	EnableFileLogging bool
	// 日志文件路径
	LogFilePath string
	// 是否启用请求日志
	EnableRequestLogging bool
	// 是否在日志中包含请求体
	LogRequestBody bool
	// 是否在日志中包含响应体
	LogResponseBody bool
	// 要排除的路径
	ExcludePaths []string
	// 是否在日志中包含调用者信息
	IncludeCallerInfo bool
	// 是否在日志中包含堆栈跟踪
	IncludeStackTrace bool
}

// DefaultLoggingConfig 返回默认日志配置
func DefaultLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		Level:               LogLevelInfo,
		EnableJSON:          true,
		EnableConsoleColor:  true,
		EnableFileLogging:   false,
		LogFilePath:         "logs/app.log",
		EnableRequestLogging: true,
		LogRequestBody:      false,
		LogResponseBody:     false,
		ExcludePaths:        []string{"/health", "/metrics"},
		IncludeCallerInfo:   true,
		IncludeStackTrace:   false,
	}
}

// 转换LogLevel到zerolog级别
func logLevelToZerologLevel(level LogLevel) zerolog.Level {
	switch level {
	case LogLevelDebug:
		return zerolog.DebugLevel
	case LogLevelInfo:
		return zerolog.InfoLevel
	case LogLevelWarn:
		return zerolog.WarnLevel
	case LogLevelError:
		return zerolog.ErrorLevel
	case LogLevelFatal:
		return zerolog.FatalLevel
	case LogLevelDisable:
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

// InitLogging 初始化日志系统
func InitLogging(config *LoggingConfig) error {
	// 设置全局日志级别
	zerolog.SetGlobalLevel(logLevelToZerologLevel(config.Level))

	// 配置日志输出
	var writers []io.Writer

	// 添加控制台输出
	if config.EnableConsoleColor {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		writers = append(writers, consoleWriter)
	} else {
		writers = append(writers, os.Stdout)
	}

	// 添加文件输出
	if config.EnableFileLogging {
		if err := os.MkdirAll(filepath.Dir(config.LogFilePath), 0755); err != nil {
			return fmt.Errorf("无法创建日志目录: %v", err)
		}

		fileWriter, err := os.OpenFile(
			config.LogFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			return fmt.Errorf("无法打开日志文件: %v", err)
		}

		writers = append(writers, fileWriter)
	}

	// 设置多输出
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else {
		output = zerolog.MultiLevelWriter(writers...)
	}

	// 配置全局logger
	logger := zerolog.New(output).With().Timestamp()

	// 添加调用者信息
	if config.IncludeCallerInfo {
		logger = logger.Caller()
	}

	// 完成logger配置
	log.Logger = logger.Logger()

	log.Info().
		Str("level", string(config.Level)).
		Bool("json", config.EnableJSON).
		Bool("console_color", config.EnableConsoleColor).
		Bool("file_logging", config.EnableFileLogging).
		Str("file_path", config.LogFilePath).
		Bool("request_logging", config.EnableRequestLogging).
		Msg("日志系统初始化完成")

	return nil
}

// LoggerMiddleware 创建Gin的日志中间件
func LoggerMiddleware(config *LoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否应该跳过该路径
		path := c.Request.URL.Path
		for _, excludePath := range config.ExcludePaths {
			if path == excludePath {
				c.Next()
				return
			}
		}

		start := time.Now()
		requestID, _ := c.Get("RequestID")
		traceID, _ := c.Get("TraceID")

		// 记录请求开始
		logger := log.With().
			Str("request_id", fmt.Sprintf("%v", requestID)).
			Str("trace_id", fmt.Sprintf("%v", traceID)).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Logger()

		if config.LogRequestBody && c.Request.ContentLength > 0 {
			if c.Request.Body != nil {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				logger = logger.With().RawJSON("request_body", bodyBytes).Logger()
			}
		}

		logger.Info().Msg("收到请求")

		// 创建自定义响应写入器以捕获状态码和大小
		responseWriter := &responseWriter{
			ResponseWriter: c.Writer,
			statusCode: 200,
		}
		c.Writer = responseWriter

		// 添加当前请求的logger到上下文
		c.Set("logger", logger)

		// 处理请求
		c.Next()

		// 记录请求结束
		latency := time.Since(start)
		statusCode := responseWriter.statusCode
		responseSize := responseWriter.size

		endLogger := logger.With().
			Int("status", statusCode).
			Dur("latency", latency).
			Int("response_size", responseSize).
			Logger()

		// 根据状态码确定日志级别
		event := endLogger.Info()
		if statusCode >= 400 && statusCode < 500 {
			event = endLogger.Warn()
		} else if statusCode >= 500 {
			event = endLogger.Error()
		}

		if config.LogResponseBody && responseWriter.body != nil {
			event = event.RawJSON("response_body", responseWriter.body)
		}

		event.Msg("请求完成")
	}
}

// LoggingResponseWriter 用于捕获响应数据的自定义响应写入器
type LoggingResponseWriter struct {
	gin.ResponseWriter
	statusCode int
	size       int
	body       []byte
	logBody    bool
}

// 创建新的日志响应写入器
func newLoggingResponseWriter(w gin.ResponseWriter, logBody bool) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		logBody:        logBody,
	}
}

// WriteHeader 重写WriteHeader以捕获状态码
func (rw *LoggingResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write 重写Write以捕获响应大小和内容
func (rw *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	if rw.logBody {
		rw.body = append(rw.body, b...)
	}
	return size, err
}

// WithContext 在上下文中添加结构化日志记录器
func WithContext(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return &log.Logger
	}

	// 尝试从上下文获取现有logger
	if l, ok := ctx.Value("logger").(zerolog.Logger); ok {
		return &l
	}

	// 获取请求ID和跟踪ID
	var requestID, traceID string
	if id, ok := ctx.Value("RequestID").(string); ok {
		requestID = id
	}

	// 尝试从OpenTelemetry跟踪上下文获取跟踪ID
	if span := trace.SpanFromContext(ctx); span != nil && span.SpanContext().IsValid() {
		traceID = span.SpanContext().TraceID().String()
	}

	// 创建并返回具有上下文信息的新logger
	logger := log.With().
		Str("request_id", requestID).
		Str("trace_id", traceID).
		Logger()

	return &logger
}

// ContextWithLogger 在上下文中添加logger
func ContextWithLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}

// Logger 从上下文获取logger
func Logger(ctx context.Context) *zerolog.Logger {
	return WithContext(ctx)
}

// LogError 记录带有堆栈跟踪的错误
func LogError(ctx context.Context, err error, msg string, fields ...map[string]interface{}) {
	if err == nil {
		return
	}

	logger := WithContext(ctx)
	event := logger.Error().Err(err).Str("message", msg)

	// 添加额外字段
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = addField(event, k, v)
		}
	}

	// 添加堆栈跟踪
	_, file, line, ok := runtime.Caller(1)
	if ok {
		event = event.Str("file", file).Int("line", line)
	}

	event.Msg(msg)
}

// 添加字段到事件（根据字段类型添加）
func addField(event *zerolog.Event, key string, value interface{}) *zerolog.Event {
	switch v := value.(type) {
	case int:
		return event.Int(key, v)
	case int64:
		return event.Int64(key, v)
	case uint64:
		return event.Uint64(key, v)
	case float64:
		return event.Float64(key, v)
	case string:
		return event.Str(key, v)
	case bool:
		return event.Bool(key, v)
	case []byte:
		return event.Bytes(key, v)
	case time.Time:
		return event.Time(key, v)
	case time.Duration:
		return event.Dur(key, v)
	case error:
		return event.Err(v)
	default:
		return event.Interface(key, v)
	}
}

// 以下是辅助函数，用于不同日志级别的记录

// Debug 记录调试级别日志
func Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := WithContext(ctx)
	event := logger.Debug()
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = addField(event, k, v)
		}
	}
	event.Msg(msg)
}

// Info 记录信息级别日志
func Info(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := WithContext(ctx)
	event := logger.Info()
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = addField(event, k, v)
		}
	}
	event.Msg(msg)
}

// Warn 记录警告级别日志
func Warn(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := WithContext(ctx)
	event := logger.Warn()
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = addField(event, k, v)
		}
	}
	event.Msg(msg)
}

// Error 记录错误级别日志
func Error(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := WithContext(ctx)
	event := logger.Error()
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = addField(event, k, v)
		}
	}
	event.Msg(msg)
}

// Fatal 记录致命级别日志
func Fatal(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := WithContext(ctx)
	event := logger.Fatal()
	if len(fields) > 0 {
		for k, v := range fields[0] {
			event = addField(event, k, v)
		}
	}
	event.Msg(msg)
} 