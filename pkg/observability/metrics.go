package observability

import (
	"bufio"
	//"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"time"

	//"encoding/json"
	"github.com/gin-gonic/gin"
	//"github.com/gin-gonic/gin/json"
	//"github.com/gin-gonic/gin/binding"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	// 定义默认的标签
	defaultLabels = []string{"service", "endpoint", "method", "status"}

	// 注册指标收集器
	registry = prometheus.NewRegistry()

	// 请求计数器
	httpRequestsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// 请求持续时间
	httpRequestDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求持续时间(秒)",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// 请求大小
	httpRequestSize = promauto.With(registry).NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_request_size_bytes",
			Help:       "HTTP请求大小(字节)",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method", "path"},
	)

	// 响应大小
	httpResponseSize = promauto.With(registry).NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "http_response_size_bytes",
			Help:       "HTTP响应大小(字节)",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method", "path"},
	)

	// 当前活跃请求数
	httpRequestsInProgress = promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_progress",
			Help: "当前正在处理的HTTP请求数",
		},
		[]string{"method", "path"},
	)

	// 业务指标 - 用户活跃度
	userActiveTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_active_total",
			Help: "用户活跃事件计数",
		},
		[]string{"event_type"},
	)

	// 业务指标 - 内容发布数
	contentPublishedTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "content_published_total",
			Help: "内容发布计数",
		},
		[]string{"content_type"},
	)

	// 业务指标 - 互动数
	interactionTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "interaction_total",
			Help: "用户互动计数",
		},
		[]string{"interaction_type"},
	)

	// 数据库查询计数器
	dbQueryTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_total",
			Help: "数据库查询计数",
		},
		[]string{"operation"},
	)

	// 数据库查询持续时间
	dbQueryDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "数据库查询持续时间(秒)",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation"},
	)

	// 缓存命中率
	cacheHitTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hit_total",
			Help: "缓存命中计数",
		},
		[]string{"cache"},
	)

	// 缓存未命中率
	cacheMissTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_miss_total",
			Help: "缓存未命中计数",
		},
		[]string{"cache"},
	)

	// 系统指标 - Go运行时
	goroutinesGauge = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines",
			Help: "当前Goroutine数量",
		},
	)

	// 系统指标 - 内存使用
	memoryUsageGauge = promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "内存使用情况(字节)",
		},
		[]string{"type"},
	)

	// 业务指标定义
	userLoginTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_total",
			Help: "用户登录总次数",
		},
		[]string{"service", "status", "user_type"},
	)

	contentCreatedTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "content_created_total",
			Help: "创建的内容总数",
		},
		[]string{"service", "type"},
	)

	recommendationRequestsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "recommendation_requests_total",
			Help: "推荐请求总数",
		},
		[]string{"service", "model", "status"},
	)

	userInteractionsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_interactions_total",
			Help: "用户交互总数",
		},
		[]string{"service", "type"},
	)

	// 系统指标定义
	databaseOperationsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_operations_total",
			Help: "数据库操作总数",
		},
		[]string{"service", "operation", "status"},
	)

	databaseOperationDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_operation_duration_seconds",
			Help:    "数据库操作持续时间",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation"},
	)

	cacheHitRate = promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_hit_rate",
			Help: "缓存命中率",
		},
		[]string{"service", "cache"},
	)

	queueDepth = promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_depth",
			Help: "队列深度",
		},
		[]string{"service", "queue"},
	)

	asyncTasksTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "async_tasks_total",
			Help: "异步任务总数",
		},
		[]string{"service", "type", "status"},
	)

	// 声明和初始化指标变量
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求持续时间",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	requestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP请求大小",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP响应大小",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	inFlightRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_in_flight_requests",
			Help: "处理中的HTTP请求数",
		},
	)
)

// MetricsConfig 指标配置
type MetricsConfig struct {
	// 是否启用指标收集
	Enabled bool
	// 指标路径
	Path string
	// 是否收集详细的HTTP请求指标
	EnableHTTPMetrics bool
	// 是否收集详细的数据库指标
	EnableDBMetrics bool
	// 是否收集详细的业务指标
	EnableBusinessMetrics bool
	// 是否收集详细的系统指标
	EnableSystemMetrics bool
}

// DefaultMetricsConfig 返回默认指标配置
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:               true,
		Path:                  "/metrics",
		EnableHTTPMetrics:     true,
		EnableDBMetrics:       true,
		EnableBusinessMetrics: true,
		EnableSystemMetrics:   true,
	}
}

// InitMetrics 初始化指标系统
func InitMetrics(config *MetricsConfig) {
	// 如果禁用指标，则不进行任何操作
	if !config.Enabled {
		return
	}

	// 启动系统指标收集
	if config.EnableSystemMetrics {
		go collectSystemMetrics()
	}
}

// 收集系统指标
func collectSystemMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		collectRuntimeMetrics()
	}
}

// 收集运行时指标
func collectRuntimeMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	goroutinesGauge.Set(float64(runtime.NumGoroutine()))
	memoryUsageGauge.WithLabelValues("alloc").Set(float64(memStats.Alloc))
	memoryUsageGauge.WithLabelValues("sys").Set(float64(memStats.Sys))
	memoryUsageGauge.WithLabelValues("heap_alloc").Set(float64(memStats.HeapAlloc))
	memoryUsageGauge.WithLabelValues("heap_sys").Set(float64(memStats.HeapSys))
	memoryUsageGauge.WithLabelValues("heap_idle").Set(float64(memStats.HeapIdle))
	memoryUsageGauge.WithLabelValues("heap_released").Set(float64(memStats.HeapReleased))
}

// RegisterMetricsHandler 注册指标处理器到HTTP服务器
func RegisterMetricsHandler(router *gin.Engine, path string) {
	router.GET(path, gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})))
}

// MetricsMiddleware 创建Gin的指标中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 排除对metrics端点本身的请求
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		// 增加进行中请求计数
		httpRequestsInProgress.WithLabelValues(method, path).Inc()
		defer httpRequestsInProgress.WithLabelValues(method, path).Dec()

		// 记录请求大小
		if c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))
		}

		// 记录开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算持续时间
		duration := time.Since(start).Seconds()

		// 记录请求
		status := c.Writer.Status()
		httpRequestsTotal.WithLabelValues(method, path, string(rune(status))).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)

		// 记录响应大小
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
		}
	}
}

// RecordDBQuery 记录数据库查询指标
func RecordDBQuery(operation string, duration time.Duration) {
	dbQueryTotal.WithLabelValues(operation).Inc()
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordCacheHit 记录缓存命中
func RecordCacheHit(cacheName string) {
	cacheHitTotal.WithLabelValues(cacheName).Inc()
}

// RecordCacheMiss 记录缓存未命中
func RecordCacheMiss(cacheName string) {
	cacheMissTotal.WithLabelValues(cacheName).Inc()
}

// RecordUserActivity 记录用户活跃事件
func RecordUserActivity(eventType string) {
	userActiveTotal.WithLabelValues(eventType).Inc()
}

// RecordContentPublished 记录内容发布
func RecordContentPublished(contentType string) {
	contentPublishedTotal.WithLabelValues(contentType).Inc()
}

// RecordInteraction 记录用户互动
func RecordInteraction(interactionType string) {
	interactionTotal.WithLabelValues(interactionType).Inc()
}

// CustomCounter 创建自定义计数器
func CustomCounter(name, help string, labelNames ...string) *prometheus.CounterVec {
	counter := promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labelNames,
	)
	return counter
}

// CustomGauge 创建自定义仪表
func CustomGauge(name, help string, labelNames ...string) *prometheus.GaugeVec {
	gauge := promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labelNames,
	)
	return gauge
}

// CustomHistogram 创建自定义直方图
func CustomHistogram(name, help string, buckets []float64, labelNames ...string) *prometheus.HistogramVec {
	histogram := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labelNames,
	)
	return histogram
}

// CustomSummary 创建自定义摘要
func CustomSummary(name, help string, objectives map[float64]float64, labelNames ...string) *prometheus.SummaryVec {
	summary := promauto.With(registry).NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labelNames,
	)
	return summary
}

// ServiceUptime 服务运行时间指标
var (
	serviceStartTime = time.Now()
	serviceUptime    = promauto.With(registry).NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "service_uptime_seconds",
			Help: "服务运行时间(秒)",
		},
		func() float64 {
			return time.Since(serviceStartTime).Seconds()
		},
	)
)

// TraceMetricsMiddleware 为请求添加跟踪和指标中间件
func TraceMetricsMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		// 提取跟踪上下文
		ctx := c.Request.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(c.Request.Header))

		// 创建请求span
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		spanName := fmt.Sprintf("%s %s", method, path)
		ctx, span := tracer.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", method),
				attribute.String("http.url", c.Request.URL.String()),
				attribute.String("http.target", c.Request.URL.Path),
			),
		)
		defer span.End()

		// 设置上下文
		c.Request = c.Request.WithContext(ctx)

		// 记录指标
		httpRequestsInProgress.WithLabelValues(method, path).Inc()
		defer httpRequestsInProgress.WithLabelValues(method, path).Dec()

		if c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))
		}

		// 记录开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算持续时间
		duration := time.Since(start)

		// 记录指标
		status := c.Writer.Status()
		httpRequestsTotal.WithLabelValues(method, path, string(rune(status))).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())

		// 记录响应大小
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
		}

		// 更新span
		span.SetAttributes(
			attribute.Int("http.status_code", status),
			attribute.Int("http.response_size", responseSize),
			attribute.Int64("http.duration_ms", duration.Milliseconds()),
		)

		// 记录错误
		if status >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
		}
	}
}

// MonitorHandler 创建一个处理程序监控包装器
func MonitorHandler(handler http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrw := newResponseWriter(w)

		handler.ServeHTTP(wrw, r)

		duration := time.Since(start)
		status := wrw.statusCode

		// 记录请求指标
		httpRequestsTotal.WithLabelValues(r.Method, name, string(rune(status))).Inc()
		httpRequestDuration.WithLabelValues(r.Method, name).Observe(duration.Seconds())

		// 记录响应大小
		size := wrw.size
		if size > 0 {
			httpResponseSize.WithLabelValues(r.Method, name).Observe(float64(size))
		}
	})
}

// 定义自定义responseWriter以提取状态码和响应大小
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
	body       []byte
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	//TODO implement me
	panic("implement me")
}

func (rw *responseWriter) Flush() {
	//TODO implement me
	panic("implement me")
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	//TODO implement me
	panic("implement me")
}

func (rw *responseWriter) Status() int {
	//TODO implement me
	panic("implement me")
}

func (rw *responseWriter) Size() int {
	//TODO implement me
	panic("implement me")
}

func (rw *responseWriter) WriteString(s string) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (rw *responseWriter) Written() bool {
	//TODO implement me
	panic("implement me")
}

func (rw *responseWriter) WriteHeaderNow() {
	//TODO implement me
	panic("implement me")
}

func (rw *responseWriter) Pusher() http.Pusher {
	//TODO implement me
	panic("implement me")
}

// newResponseWriter 创建新的响应写入器
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader 重写WriteHeader以捕获状态码
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write 重写Write以捕获响应大小
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Unwrap 实现http.ResponseWriter接口
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// 初始化metrics
func init() {
	// 注册HTTP相关指标
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpRequestSize)
	prometheus.MustRegister(httpResponseSize)
	prometheus.MustRegister(httpRequestsInProgress)

	// 注册业务指标
	prometheus.MustRegister(userActiveTotal)
	prometheus.MustRegister(contentPublishedTotal)
	prometheus.MustRegister(interactionTotal)

	// 注册系统指标
	prometheus.MustRegister(dbQueryTotal)
	prometheus.MustRegister(dbQueryDuration)
	prometheus.MustRegister(cacheHitTotal)
	prometheus.MustRegister(cacheMissTotal)
	prometheus.MustRegister(goroutinesGauge)
	prometheus.MustRegister(memoryUsageGauge)

	log.Info().Msg("Prometheus metrics registered")
}

// DetailedMetricsMiddleware 创建用于收集详细HTTP请求指标的中间件
func DetailedMetricsMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 请求大小
		reqSize := float64(c.Request.ContentLength)

		// 增加活跃请求计数
		inFlightRequests.Inc()

		// 处理请求
		c.Next()

		// 减少活跃请求计数
		inFlightRequests.Dec()

		// 处理时间
		duration := time.Since(start).Seconds()

		// 状态码
		statusCode := strconv.Itoa(c.Writer.Status())

		// 响应大小
		respSize := float64(c.Writer.Size())

		// 端点
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		// 记录请求指标
		totalRequests.WithLabelValues(serviceName, endpoint, c.Request.Method, statusCode).Inc()
		requestDuration.WithLabelValues(serviceName, endpoint, c.Request.Method, statusCode).Observe(duration)

		// 记录请求和响应大小
		if reqSize > 0 {
			requestSize.WithLabelValues(serviceName, endpoint, c.Request.Method, statusCode).Observe(reqSize)
		}

		if respSize > 0 {
			responseSize.WithLabelValues(serviceName, endpoint, c.Request.Method, statusCode).Observe(respSize)
		}
	}
}

// SetupMetricsEndpoint 设置Prometheus指标暴露端点
func SetupMetricsEndpoint(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// RecordUserLogin 记录用户登录指标
func RecordUserLogin(serviceName, status, userType string) {
	userLoginTotal.WithLabelValues(serviceName, status, userType).Inc()
}

// RecordContentCreated 记录内容创建指标
func RecordContentCreated(serviceName, contentType string) {
	contentCreatedTotal.WithLabelValues(serviceName, contentType).Inc()
}

// RecordRecommendationRequest 记录推荐请求指标
func RecordRecommendationRequest(serviceName, modelID, status string) {
	recommendationRequestsTotal.WithLabelValues(serviceName, modelID, status).Inc()
}

// RecordUserInteraction 记录用户互动指标
func RecordUserInteraction(serviceName, interactionType string) {
	userInteractionsTotal.WithLabelValues(serviceName, interactionType).Inc()
}

// RecordDatabaseOperation 记录数据库操作指标
func RecordDatabaseOperation(serviceName, operation, status string) {
	databaseOperationsTotal.WithLabelValues(serviceName, operation, status).Inc()
}

// ObserveDatabaseOperationDuration 观察数据库操作耗时
func ObserveDatabaseOperationDuration(serviceName, operation string, duration time.Duration) {
	databaseOperationDuration.WithLabelValues(serviceName, operation).Observe(duration.Seconds())
}

// SetCacheHitRate 设置缓存命中率
func SetCacheHitRate(serviceName, cacheName string, rate float64) {
	cacheHitRate.WithLabelValues(serviceName, cacheName).Set(rate)
}

// SetQueueDepth 设置队列深度
func SetQueueDepth(serviceName, queueName string, depth float64) {
	queueDepth.WithLabelValues(serviceName, queueName).Set(depth)
}

// RecordAsyncTask 记录异步任务指标
func RecordAsyncTask(serviceName, taskType, status string) {
	asyncTasksTotal.WithLabelValues(serviceName, taskType, status).Inc()
}

// DatabaseTimer 用于测量数据库操作耗时的辅助结构
type DatabaseTimer struct {
	serviceName string
	operation   string
	startTime   time.Time
}

// NewDatabaseTimer 创建数据库操作计时器
func NewDatabaseTimer(serviceName, operation string) *DatabaseTimer {
	return &DatabaseTimer{
		serviceName: serviceName,
		operation:   operation,
		startTime:   time.Now(),
	}
}

// ObserveDuration 观察数据库操作耗时并记录
func (t *DatabaseTimer) ObserveDuration() {
	duration := time.Since(t.startTime)
	ObserveDatabaseOperationDuration(t.serviceName, t.operation, duration)
}

// RegisterCustomCounter 注册自定义计数器
func RegisterCustomCounter(name, help string, labels []string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	prometheus.MustRegister(counter)
	return counter
}

// RegisterCustomGauge 注册自定义仪表
func RegisterCustomGauge(name, help string, labels []string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	prometheus.MustRegister(gauge)
	return gauge
}

// RegisterCustomHistogram 注册自定义直方图
func RegisterCustomHistogram(name, help string, buckets []float64, labels []string) *prometheus.HistogramVec {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	prometheus.MustRegister(histogram)
	return histogram
}

// RegisterCustomSummary 注册自定义摘要
func RegisterCustomSummary(name, help string, objectives map[float64]float64, labels []string) *prometheus.SummaryVec {
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)
	prometheus.MustRegister(summary)
	return summary
}

// SetupPrometheusHandler 创建独立的Prometheus HTTP处理器
func SetupPrometheusHandler() http.Handler {
	return promhttp.Handler()
}
