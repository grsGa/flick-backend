package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"backend/services/gateway/middleware"
)

// Service 服务定义
type Service struct {
	Name     string   // 服务名称
	Prefix   string   // API前缀
	Host     string   // 服务主机地址
	Versions []string // 支持的版本
	Active   bool     // 服务是否可用
}

// VersionConfig 版本配置
type VersionConfig struct {
	Status           string    // 状态：stable, beta, deprecated
	DeprecationDate  time.Time // 弃用日期
	WarningMessage   string    // 警告消息
	MinimumVersion   string    // 最低客户端版本
	RecommendVersion string    // 推荐客户端版本
}

// RouterConfig 路由器配置
type RouterConfig struct {
	// 服务列表
	Services []Service
	// 版本配置
	Versions map[string]VersionConfig
	// JWT配置
	JWTConfig *middleware.JWTConfig
	// RBAC配置
	RBACConfig *middleware.RBACConfig
	// CORS配置
	CORSConfig *cors.Config
}

// DefaultRouterConfig 默认配置
func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		Services: []Service{},
		Versions: map[string]VersionConfig{
			"v1": {
				Status:           "stable",
				MinimumVersion:   "1.0.0",
				RecommendVersion: "1.0.0",
			},
		},
		JWTConfig:  middleware.DefaultJWTConfig(),
		RBACConfig: middleware.NewRBACConfig(),
		CORSConfig: &cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
	}
}

// Router API网关路由器
type Router struct {
	Engine *gin.Engine   // Gin引擎
	Config *RouterConfig // 路由配置
}

// NewRouter 创建新的路由器
func NewRouter(config *RouterConfig) *Router {
	if config == nil {
		config = DefaultRouterConfig()
	}

	router := &Router{
		Engine: gin.Default(),
		Config: config,
	}

	// 应用全局中间件
	router.applyGlobalMiddleware()

	// 注册路由
	router.registerRoutes()

	return router
}

// 应用全局中间件
func (r *Router) applyGlobalMiddleware() {
	// CORS中间件
	if r.Config.CORSConfig != nil {
		r.Engine.Use(cors.New(*r.Config.CORSConfig))
	}

	// 请求ID中间件
	r.Engine.Use(middleware.RequestID())

	// 版本检查中间件
	r.Engine.Use(r.versionCheckMiddleware())

	// JWT认证中间件
	if r.Config.JWTConfig != nil {
		r.Engine.Use(middleware.JWTMiddleware(r.Config.JWTConfig))
	}

	// RBAC权限中间件
	if r.Config.RBACConfig != nil {
		r.Engine.Use(middleware.RBACMiddleware(r.Config.RBACConfig))
	}
}

// 注册所有路由
func (r *Router) registerRoutes() {
	// 健康检查
	r.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API信息
	r.Engine.GET("/api-info", func(c *gin.Context) {
		serviceInfo := make([]map[string]interface{}, 0)
		for _, svc := range r.Config.Services {
			if svc.Active {
				serviceInfo = append(serviceInfo, map[string]interface{}{
					"name":     svc.Name,
					"prefix":   svc.Prefix,
					"versions": svc.Versions,
				})
			}
		}

		versionInfo := make(map[string]interface{})
		for ver, config := range r.Config.Versions {
			versionInfo[ver] = map[string]interface{}{
				"status":            config.Status,
				"minimum_version":   config.MinimumVersion,
				"recommend_version": config.RecommendVersion,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"services": serviceInfo,
			"versions": versionInfo,
		})
	})

	// 注册API分组
	for _, service := range r.Config.Services {
		if !service.Active {
			continue
		}

		for _, version := range service.Versions {
			versionPrefix := "/" + version
			apiGroup := r.Engine.Group(versionPrefix + service.Prefix)

			// 在这里可以为每个API分组添加特定的中间件
			// 例如：apiGroup.Use(...)

			// 添加服务路由
			r.registerServiceRoutes(apiGroup, service, version)
		}
	}
}

// 注册服务路由
func (r *Router) registerServiceRoutes(group *gin.RouterGroup, service Service, version string) {
	// 此处应该根据服务配置动态添加路由
	// 这可能涉及到服务发现或配置文件

	// 服务健康检查
	group.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": service.Name,
			"version": version,
			"status":  "ok",
		})
	})

	// 示例路由 - 在实际应用中，这些会动态生成
	group.GET("/service-info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": service.Name,
			"host":    service.Host,
			"version": version,
		})
	})
}

// 版本检查中间件
func (r *Router) versionCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 检查路径是否包含版本信息
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			c.Next()
			return
		}

		// 假设版本在路径的第一部分，例如 /v1/users
		potentialVersion := parts[1]

		// 检查是否是已定义的版本
		versionConfig, exists := r.Config.Versions[potentialVersion]
		if !exists {
			c.Next()
			return
		}

		// 检查版本状态
		if versionConfig.Status == "deprecated" {
			// 添加弃用警告头
			c.Header("X-API-Deprecated", "true")
			c.Header("X-API-Deprecation-Date", versionConfig.DeprecationDate.Format(time.RFC3339))
			c.Header("X-API-Warning", versionConfig.WarningMessage)
			c.Header("X-API-Recommended-Version", versionConfig.RecommendVersion)
		}

		// 检查客户端版本
		clientVersion := c.GetHeader("X-Client-Version")
		if clientVersion != "" && versionConfig.MinimumVersion != "" {
			// 这里应该有版本比较逻辑
			// 如果客户端版本低于最低版本，可以添加警告或拒绝请求
		}

		c.Next()
	}
}

// AddService 添加服务
func (r *Router) AddService(service Service) {
	r.Config.Services = append(r.Config.Services, service)
}

// AddVersionConfig 添加版本配置
func (r *Router) AddVersionConfig(version string, config VersionConfig) {
	r.Config.Versions[version] = config
}

// Run 运行路由器
func (r *Router) Run(addr string) error {
	return r.Engine.Run(addr)
}
