package observability

import (
	"context"
	"io"
	"os"
	"time"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

// LogConfig 日志配置
type LogConfig struct {
	// 日志级别
	Level string
	// 是否使用JSON格式
	UseJSON bool
	// 是否在控制台输出
	ConsoleOutput bool
	// 日志文件路径
	FilePath string
	// 服务名称
	ServiceName string
	// 是否包含调用者信息
	IncludeCaller bool
	// 是否启用堆栈跟踪
	EnableStackTrace bool
}

// DefaultLogConfig 默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:            "info",
		UseJSON:          true,
		ConsoleOutput:    true,
		FilePath:         "",
		ServiceName:      "flick",
		IncludeCaller:    true,
		EnableStackTrace: true,
	}
}

// InitLogger 初始化日志系统
func InitLogger(config *LogConfig) {
	// 设置时间格式
	zerolog.TimeFieldFormat = time.RFC3339Nano
	
	// 设置日志级别
	setLogLevel(config.Level)
	
	// 确定日志输出
	var writers []io.Writer
	
	// 添加控制台输出
	if config.ConsoleOutput {
		if !config.UseJSON {
			// 使用格式化输出
			consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
			writers = append(writers, consoleWriter)
		} else {
			writers = append(writers, os.Stdout)
		}
	}
	
	// 添加文件输出
	if config.FilePath != "" {
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			writers = append(writers, file)
		} else {
			log.Error().Err(err).Str("path", config.FilePath).Msg("无法打开日志文件")
		}
	}
	
	// 设置输出
	var output io.Writer
	if len(writers) > 1 {
		output = zerolog.MultiLevelWriter(writers...)
	} else if len(writers) == 1 {
		output = writers[0]
	} else {
		output = os.Stdout // 默认输出
	}
	
	// 创建日志记录器
	logger := zerolog.New(output).With().Timestamp()
	
	// 添加服务名称
	if config.ServiceName != "" {
		logger = logger.Str("service", config.ServiceName)
	}
	
	// 添加调用者信息
	if config.IncludeCaller {
		logger = logger.Caller()
	}
	
	// 完成日志记录器设置
	log.Logger = logger.Logger()
	
	log.Info().Msg("日志系统初始化完成")
}

// 设置日志级别
func setLogLevel(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// TraceLogger 带有跟踪上下文的日志记录器
type TraceLogger struct {
	ctx context.Context
}

// NewTraceLogger 创建带有跟踪上下文的日志记录器
func NewTraceLogger(ctx context.Context) *TraceLogger {
	return &TraceLogger{ctx: ctx}
}

// getTraceInfo 获取跟踪信息
func (l *TraceLogger) getTraceInfo() (string, string) {
	span := trace.SpanFromContext(l.ctx)
	if !span.SpanContext().IsValid() {
		return "", ""
	}
	
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()
	return traceID, spanID
}

// createEvent 创建日志事件并添加跟踪信息
func (l *TraceLogger) createEvent(level zerolog.Level) *zerolog.Event {
	event := log.WithLevel(level)
	
	traceID, spanID := l.getTraceInfo()
	if traceID != "" {
		event = event.Str("trace_id", traceID)
	}
	if spanID != "" {
		event = event.Str("span_id", spanID)
	}
	
	// 从上下文中提取其他相关信息
	if userID, ok := l.ctx.Value("user_id").(uint); ok {
		event = event.Uint("user_id", userID)
	}
	
	if reqID, ok := l.ctx.Value("requestId").(string); ok {
		event = event.Str("request_id", reqID)
	}
	
	return event
}

// Debug 记录调试级别日志
func (l *TraceLogger) Debug() *zerolog.Event {
	return l.createEvent(zerolog.DebugLevel)
}

// Info 记录信息级别日志
func (l *TraceLogger) Info() *zerolog.Event {
	return l.createEvent(zerolog.InfoLevel)
}

// Warn 记录警告级别日志
func (l *TraceLogger) Warn() *zerolog.Event {
	return l.createEvent(zerolog.WarnLevel)
}

// Error 记录错误级别日志
func (l *TraceLogger) Error() *zerolog.Event {
	return l.createEvent(zerolog.ErrorLevel)
}

// Fatal 记录致命级别日志
func (l *TraceLogger) Fatal() *zerolog.Event {
	return l.createEvent(zerolog.FatalLevel)
}

// ContextLogger 从上下文创建日志记录器
func ContextLogger(ctx context.Context) *TraceLogger {
	return NewTraceLogger(ctx)
}

// LoggerWithContext 创建带有额外上下文字段的记录器
func LoggerWithContext(ctx context.Context) zerolog.Logger {
	logger := log.With()
	
	// 从上下文中提取跟踪信息
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		logger = logger.Str("trace_id", span.SpanContext().TraceID().String())
		logger = logger.Str("span_id", span.SpanContext().SpanID().String())
	}
	
	// 提取请求ID
	if reqID, ok := ctx.Value("requestId").(string); ok {
		logger = logger.Str("request_id", reqID)
	}
	
	// 提取用户ID
	if userID, ok := ctx.Value("user_id").(uint); ok {
		logger = logger.Uint("user_id", userID)
	}
	
	return logger.Logger()
}

// FieldLogger 为特定组件创建带有固定字段的记录器
func FieldLogger(component string, fields map[string]interface{}) zerolog.Logger {
	logger := log.With().Str("component", component)
	
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			logger = logger.Str(k, val)
		case int:
			logger = logger.Int(k, val)
		case bool:
			logger = logger.Bool(k, val)
		case float64:
			logger = logger.Float64(k, val)
		}
	}
	
	return logger.Logger()
}

// LogDBQuery 记录数据库查询
func LogDBQuery(ctx context.Context, query string, args ...interface{}) {
	logger := ContextLogger(ctx)
	logger.Debug().
		Str("query", query).
		Interface("args", args).
		Msg("执行数据库查询")
}

// LogHTTPRequest 记录HTTP请求
func LogHTTPRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	logger := ContextLogger(ctx)
	logger.Info().
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Dur("duration", duration).
		Msg("HTTP请求")
}

// LogServiceCall 记录服务间调用
func LogServiceCall(ctx context.Context, service, method string, duration time.Duration, err error) {
	logger := ContextLogger(ctx)
	event := logger.Info().
		Str("service", service).
		Str("method", method).
		Dur("duration", duration)
	
	if err != nil {
		event.Err(err).Msg("服务调用失败")
	} else {
		event.Msg("服务调用成功")
	}
}

// LogBusinessEvent 记录业务事件
func LogBusinessEvent(ctx context.Context, eventType string, details map[string]interface{}) {
	logger := ContextLogger(ctx)
	event := logger.Info().Str("event_type", eventType)
	
	for k, v := range details {
		switch val := v.(type) {
		case string:
			event = event.Str(k, val)
		case int:
			event = event.Int(k, val)
		case bool:
			event = event.Bool(k, val)
		case float64:
			event = event.Float64(k, val)
		default:
			event = event.Interface(k, v)
		}
	}
	
	event.Msg("业务事件")
}

// LogSimpleError 记录简单错误
func LogSimpleError(ctx context.Context, err error, msg string) {
	logger := ContextLogger(ctx)
	logger.Error().Err(err).Msg(msg)
}

// LogFatal 记录致命错误
func LogFatal(ctx context.Context, err error, msg string) {
	logger := ContextLogger(ctx)
	logger.Fatal().Err(err).Msg(msg)
} 