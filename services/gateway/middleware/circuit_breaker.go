package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	// Closed 关闭状态：允许请求通过
	Closed CircuitBreakerState = iota
	// Open 开启状态：阻止请求通过
	Open
	// HalfOpen 半开状态：允许部分请求通过，用于探测服务是否恢复
	HalfOpen
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu                   sync.RWMutex
	state                CircuitBreakerState
	failureThreshold     int           // 触发熔断的失败次数阈值
	successThreshold     int           // 恢复服务的成功次数阈值
	timeout              time.Duration // 熔断恢复的超时时间
	lastStateChangeTime  time.Time     // 最后一次状态变更时间
	failureCount         int           // 当前失败次数
	successCount         int           // 当前成功次数
	degradedResponseFunc func(*gin.Context)
}

// NewCircuitBreaker 创建新的熔断器
func NewCircuitBreaker(
	failureThreshold int,
	successThreshold int,
	timeout time.Duration,
	degradedResponseFunc func(*gin.Context),
) *CircuitBreaker {
	return &CircuitBreaker{
		state:                Closed,
		failureThreshold:     failureThreshold,
		successThreshold:     successThreshold,
		timeout:              timeout,
		lastStateChangeTime:  time.Now(),
		failureCount:         0,
		successCount:         0,
		degradedResponseFunc: degradedResponseFunc,
	}
}

// GetState 获取当前熔断器状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// 检查是否应该从Open切换到HalfOpen
	if cb.state == Open && time.Since(cb.lastStateChangeTime) >= cb.timeout {
		// 注意: 这里不能直接修改状态，因为我们持有读锁
		return HalfOpen
	}

	return cb.state
}

// OnSuccess 记录成功请求
func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// 检查是否应该从Open切换到HalfOpen
	if cb.state == Open && time.Since(cb.lastStateChangeTime) >= cb.timeout {
		cb.state = HalfOpen
		cb.lastStateChangeTime = time.Now()
		cb.successCount = 0
	}

	// 只在HalfOpen状态下计数成功次数
	if cb.state == HalfOpen {
		cb.successCount++

		// 达到成功阈值，恢复服务
		if cb.successCount >= cb.successThreshold {
			cb.state = Closed
			cb.lastStateChangeTime = time.Now()
			cb.failureCount = 0
			cb.successCount = 0
		}
	}

	// 关闭状态下重置失败计数
	if cb.state == Closed {
		cb.failureCount = 0
	}
}

// OnFailure 记录失败请求
func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// 针对不同状态计数
	switch cb.state {
	case Closed:
		cb.failureCount++
		// 达到失败阈值，触发熔断
		if cb.failureCount >= cb.failureThreshold {
			cb.state = Open
			cb.lastStateChangeTime = time.Now()
			cb.successCount = 0
		}
	case HalfOpen:
		// 半开状态下任何失败都会回到熔断状态
		cb.state = Open
		cb.lastStateChangeTime = time.Now()
		cb.successCount = 0
	}
}

// CircuitBreakerMiddleware 熔断器中间件
func CircuitBreakerMiddleware(name string, cb *CircuitBreaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查熔断器状态
		state := cb.GetState()

		// 如果熔断器开启，直接返回降级响应
		if state == Open {
			cb.degradedResponseFunc(c)
			return
		}

		// 保存原始ResponseWriter
		originalWriter := c.Writer

		// 创建代理ResponseWriter
		blw := &bodyLogWriter{
			ResponseWriter: originalWriter,
			statusCode:     http.StatusOK, // 默认状态码
		}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 根据状态码更新熔断器
		statusCode := blw.statusCode

		// 判断是否为失败状态码 (4xx和5xx)
		if statusCode >= 400 {
			cb.OnFailure()
		} else {
			cb.OnSuccess()
		}
	}
}

// CircuitBreakerMiddlewareByPathName 为不同路径创建熔断器的中间件
func CircuitBreakerMiddlewareByPathName() gin.HandlerFunc {
	circuitBreakers := make(map[string]*CircuitBreaker)
	var mu sync.RWMutex

	// 默认降级响应
	defaultDegradedResponse := func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "服务暂时不可用",
			"message": "请稍后再试",
		})
		c.Abort()
	}

	return func(c *gin.Context) {
		path := c.FullPath()

		// 获取或创建对应路径的熔断器
		mu.RLock()
		cb, exists := circuitBreakers[path]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			// 双重检查
			cb, exists = circuitBreakers[path]
			if !exists {
				// 为每个路径创建一个熔断器
				cb = NewCircuitBreaker(5, 3, 10*time.Second, defaultDegradedResponse)
				circuitBreakers[path] = cb
			}
			mu.Unlock()
		}

		// 检查熔断器状态
		state := cb.GetState()

		// 如果熔断器开启，直接返回降级响应
		if state == Open {
			defaultDegradedResponse(c)
			return
		}

		// 保存原始ResponseWriter
		originalWriter := c.Writer

		// 创建代理ResponseWriter
		blw := &bodyLogWriter{
			ResponseWriter: originalWriter,
			statusCode:     http.StatusOK, // 默认状态码
		}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 根据状态码更新熔断器
		statusCode := blw.statusCode

		// 判断是否为失败状态码 (4xx和5xx)
		if statusCode >= 400 {
			cb.OnFailure()
		} else {
			cb.OnSuccess()
		}
	}
}

// bodyLogWriter 是一个ResponseWriter代理，用于记录状态码
type bodyLogWriter struct {
	gin.ResponseWriter
	statusCode int
}

// WriteHeader 实现 ResponseWriter 接口
func (w *bodyLogWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
