package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// InitTracerProvider TracerProvider初始化和配置OpenTelemetry跟踪提供程序
func InitTracerProvider(serviceName, endpoint string) (*tracesdk.TracerProvider, error) {
	// 创建OTLP HTTP导出器
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// 创建跟踪提供程序，配置导出器和资源属性
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	// 设置全局跟踪提供程序
	otel.SetTracerProvider(tp)

	// 设置全局传播器，用于在服务之间传递跟踪上下文
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// TracerMiddleware 创建OpenTelemetry跟踪中间件
func TracerMiddleware(serviceName string) gin.HandlerFunc {
	// 获取跟踪提供程序
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		// 从请求头提取传播的上下文
		ctx := c.Request.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		// 创建新的span
		spanName := c.FullPath()
		if spanName == "" {
			spanName = c.Request.URL.Path
		}

		opts := []trace.SpanStartOption{
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.host", c.Request.Host),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		ctx, span := tracer.Start(ctx, spanName, opts...)
		defer span.End()

		// 记录请求开始时间
		startTime := time.Now()

		// 将跟踪上下文传递到处理程序
		c.Request = c.Request.WithContext(ctx)

		// 保存traceID到gin上下文中
		traceID := span.SpanContext().TraceID().String()
		c.Set("traceID", traceID)
		c.Header("X-Trace-ID", traceID)

		// 处理请求
		c.Next()

		// 请求完成后记录属性
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int64("http.response_time_ms", duration.Milliseconds()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// 根据状态码设置span状态
		if statusCode >= 500 {
			span.SetAttributes(attribute.String("error.type", "server_error"))
		} else if statusCode >= 400 {
			span.SetAttributes(attribute.String("error.type", "client_error"))
		}

		// 记录错误信息
		for _, err := range c.Errors {
			span.RecordError(err.Err)
		}
	}
}

// PropagateTracingHeaders 将跟踪头部信息传播到下游服务的中间件
func PropagateTracingHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前跟踪上下文
		ctx := c.Request.Context()

		// 从上下文中获取当前活动的span并记录 traceID 到响应头
		span := trace.SpanFromContext(ctx)
		if span.SpanContext().HasTraceID() {
			c.Header("X-Trace-ID", span.SpanContext().TraceID().String())
		}

		// 注入跟踪信息到headers，稍后用于传递给下游服务
		carrier := make(propagation.MapCarrier)
		otel.GetTextMapPropagator().Inject(ctx, carrier)

		// 保存carrier到上下文，以便后续使用
		c.Set("tracing_headers", carrier)

		// 继续处理请求
		c.Next()
	}
}

// TraceID 从上下文中获取traceID的辅助函数
func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().HasTraceID() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}
