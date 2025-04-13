package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TracerConfig 配置分布式追踪
type TracerConfig struct {
	// 服务名称
	ServiceName string
	// 服务版本
	ServiceVersion string
	// 环境名称
	Environment string
	// 追踪采样率 (0.0-1.0)
	SamplingRate float64
	// 导出器类型: "jaeger", "otlp"
	ExporterType string
	// Jaeger导出器配置
	JaegerEndpoint string
	// OTLP导出器配置
	OTLPEndpoint string
	// 是否开启调试模式
	Debug bool
}

// DefaultTracerConfig 返回默认追踪配置
func DefaultTracerConfig() *TracerConfig {
	return &TracerConfig{
		ServiceName:    "flick",
		ServiceVersion: "0.1.0",
		Environment:    "development",
		SamplingRate:   0.2,
		ExporterType:   "jaeger",
		JaegerEndpoint: "http://localhost:14268/api/traces",
		OTLPEndpoint:   "localhost:4317",
		Debug:          false,
	}
}

// InitTracer 初始化分布式追踪系统
func InitTracer(config *TracerConfig) (func(context.Context) error, error) {
	// 创建资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("无法创建资源: %w", err)
	}

	// 创建导出器
	var exporter sdktrace.SpanExporter

	switch config.ExporterType {
	case "jaeger":
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
		if err != nil {
			return nil, fmt.Errorf("无法创建Jaeger导出器: %w", err)
		}
	case "otlp":
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		conn, err := grpc.DialContext(ctx, config.OTLPEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			return nil, fmt.Errorf("无法连接到OTLP端点: %w", err)
		}
		otlpClient := otlptracegrpc.NewClient(otlptracegrpc.WithGRPCConn(conn))
		exporter, err = otlptrace.New(context.Background(), otlpClient)
		if err != nil {
			return nil, fmt.Errorf("无法创建OTLP导出器: %w", err)
		}
	default:
		return nil, fmt.Errorf("不支持的导出器类型: %s", config.ExporterType)
	}

	// 创建采样器
	var sampler sdktrace.Sampler
	if config.Debug {
		sampler = sdktrace.AlwaysSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(config.SamplingRate)
	}

	// 创建追踪提供者
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	// 设置全局追踪提供者
	otel.SetTracerProvider(tp)

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 返回关闭函数
	return tp.Shutdown, nil
}

// TraceContext 用于在上下文中存储和传递追踪信息
type TraceContext struct {
	TraceID string
	SpanID  string
	Parent  context.Context
}

// TracerMiddleware 为Gin创建追踪中间件
func TracerMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		// 从请求头提取追踪上下文
		ctx := c.Request.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		// 创建请求处理的Span
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		spanName := fmt.Sprintf("%s %s", c.Request.Method, path)
		ctx, span := tracer.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(c.Request.Method),
				semconv.HTTPURLKey.String(c.Request.URL.String()),
				semconv.HTTPTargetKey.String(c.Request.URL.Path),
				semconv.HTTPSchemeKey.String(c.Request.URL.Scheme),
				semconv.HTTPHostKey.String(c.Request.Host),
				semconv.NetPeerIPKey.String(c.ClientIP()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
			),
		)
		defer span.End()

		// 将span的信息添加到请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 设置追踪ID和请求ID
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()

		// 向响应头添加追踪ID
		c.Header("X-Trace-ID", traceID)

		// 在上下文中存储追踪信息
		c.Set("trace_id", traceID)
		c.Set("span_id", spanID)

		// 处理请求
		start := time.Now()
		c.Next()

		// 添加响应信息到span
		statusCode := c.Writer.Status()
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(statusCode),
			attribute.Int("http.response_size", c.Writer.Size()),
			attribute.Int64("http.duration_ms", time.Since(start).Milliseconds()),
		)

		// 记录错误信息
		if statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
		}

		// 这里可以添加日志记录代码
		// LogHTTPRequest(ctx, c.Request.Method, path, statusCode, time.Since(start))
	}
}

// StartSpan 启动一个新的跟踪Span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("flick").Start(ctx, name, opts...)
}

// StartDBSpan 启动数据库操作的Span
func StartDBSpan(ctx context.Context, operation, query string) (context.Context, trace.Span) {
	return StartSpan(ctx, "db."+operation,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", query),
			attribute.String("db.operation", operation),
		),
	)
}

// StartServiceSpan 启动服务调用的Span
func StartServiceSpan(ctx context.Context, service, method string) (context.Context, trace.Span) {
	return StartSpan(ctx, fmt.Sprintf("%s.%s", service, method),
		trace.WithAttributes(
			attribute.String("service.name", service),
			attribute.String("service.method", method),
		),
	)
}

// StartMQSpan 启动消息队列操作的Span
func StartMQSpan(ctx context.Context, system, queue, operation string) (context.Context, trace.Span) {
	return StartSpan(ctx, fmt.Sprintf("mq.%s.%s", system, operation),
		trace.WithAttributes(
			attribute.String("messaging.system", system),
			attribute.String("messaging.destination", queue),
			attribute.String("messaging.operation", operation),
		),
	)
}

// GetTraceIDFromContext 从上下文中获取追踪ID
func GetTraceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanIDFromContext 从上下文中获取Span ID
func GetSpanIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// WithSpanAttributes 向当前span添加属性
func WithSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError 记录错误到当前span
func RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// InjectTraceContext 注入追踪上下文到HTTP头
func InjectTraceContext(ctx context.Context, headers map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))
}

// ExtractTraceContext 从HTTP头提取追踪上下文
func ExtractTraceContext(ctx context.Context, headers map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(headers))
}

// TraceDBQuery 追踪数据库查询的执行
func TraceDBQuery(ctx context.Context, operation string, query string, args ...interface{}) func(error) {
	ctx, span := StartDBSpan(ctx, operation, query)
	start := time.Now()

	// 这里可以添加日志记录代码
	// LogDBQuery(ctx, query, args...)

	return func(err error) {
		duration := time.Since(start)
		span.SetAttributes(attribute.Int64("db.duration_ms", duration.Milliseconds()))
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}

// TraceMQOperation 追踪消息队列操作
func TraceMQOperation(ctx context.Context, system, queue, operation string, payload interface{}) func(error) {
	ctx, span := StartMQSpan(ctx, system, queue, operation)
	start := time.Now()

	return func(err error) {
		duration := time.Since(start)
		span.SetAttributes(attribute.Int64("messaging.duration_ms", duration.Milliseconds()))
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}

// AsyncSpan 创建异步操作的Span
type AsyncSpan struct {
	Span  trace.Span
	Start time.Time
}

// NewAsyncSpan 创建新的异步操作Span
func NewAsyncSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) *AsyncSpan {
	_, span := StartSpan(ctx, name, trace.WithAttributes(attrs...))
	return &AsyncSpan{
		Span:  span,
		Start: time.Now(),
	}
}

// End 结束异步Span
func (a *AsyncSpan) End(err error) {
	duration := time.Since(a.Start)
	a.Span.SetAttributes(attribute.Int64("duration_ms", duration.Milliseconds()))
	if err != nil {
		a.Span.RecordError(err)
		a.Span.SetStatus(codes.Error, err.Error())
	}
	// 使用空的上下文调用
	a.Span.End()
}
