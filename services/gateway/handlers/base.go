package handlers

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"backend/services/gateway/config"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// BaseHandler 提供所有处理程序共享的基本功能
type BaseHandler struct {
	Config *config.Config
}

// ForwardRequest 将请求转发到目标微服务
func (h *BaseHandler) ForwardRequest(c *gin.Context, serviceName string, path string, method string, body io.Reader) (*http.Response, error) {
	// 获取服务配置
	serviceConfig, exists := h.Config.Services[strings.ToLower(serviceName)]
	if !exists {
		log.Error().Str("service", serviceName).Msg("未找到服务配置")
		return nil, fmt.Errorf("服务 '%s' 未配置", serviceName)
	}

	// 构建目标URL
	targetURL, err := url.Parse(serviceConfig.URL)
	if err != nil {
		log.Error().Err(err).Str("url", serviceConfig.URL).Msg("解析服务URL失败")
		return nil, fmt.Errorf("解析服务URL失败: %w", err)
	}

	// 组合完整路径
	targetURL.Path = path

	// 创建新请求
	req, err := http.NewRequest(method, targetURL.String(), body)
	if err != nil {
		log.Error().Err(err).Str("url", targetURL.String()).Str("method", method).Msg("创建请求失败")
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 复制原始请求的头部
	for key, values := range c.Request.Header {
		// 不复制Host和Connection头部
		if key != "Host" && key != "Connection" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// 如果有用户ID，添加到请求头
	if userID, exists := c.Get("user_id"); exists {
		req.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
	}

	// 如果有角色，添加到请求头
	if role, exists := c.Get("role"); exists {
		req.Header.Set("X-User-Role", fmt.Sprintf("%v", role))
	}

	// 设置内容类型
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("url", targetURL.String()).Msg("请求失败")
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	return resp, nil
}

// HandleRequest 处理请求并转发到微服务
func (h *BaseHandler) HandleRequest(c *gin.Context, serviceName string) {
	// 读取并保存请求体，因为后面可能还需要使用
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		// 恢复请求体，以便后续使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// 获取服务配置
	serviceConfig, exists := h.Config.Services[strings.ToLower(serviceName)]
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务未配置", "details": fmt.Sprintf("服务 '%s' 未配置", serviceName)})
		return
	}

	// 构建目标路径
	path := c.Request.URL.Path

	// 仅当服务URL不包含/api/v1前缀时才移除请求中的前缀
	// 检查服务URL是否已包含/api/v1
	if !strings.Contains(serviceConfig.URL, "/api/v1") && strings.HasPrefix(path, "/api/v1") {
		path = strings.TrimPrefix(path, "/api/v1")
	}

	// 包含查询参数
	if c.Request.URL.RawQuery != "" {
		path = path + "?" + c.Request.URL.RawQuery
	}

	// 转发请求
	resp, err := h.ForwardRequest(c, serviceName, path, c.Request.Method, bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务请求失败", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取响应失败", "details": err.Error()})
		return
	}

	// 复制响应头
	for k, values := range resp.Header {
		for _, v := range values {
			c.Writer.Header().Add(k, v)
		}
	}

	// 设置状态码
	c.Status(resp.StatusCode)

	// 写入响应体
	if len(respBody) > 0 {
		c.Writer.Write(respBody)
	}
}

// DecodeJSON 从请求中解码JSON数据
func (h *BaseHandler) DecodeJSON(c *gin.Context, v interface{}) error {
	return c.ShouldBindJSON(v)
}

// RespondWithJSON 以JSON格式响应
func (h *BaseHandler) RespondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

// RespondWithError 以JSON格式响应错误
func (h *BaseHandler) RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

// GetUserIDFromContext 从上下文获取用户ID
func (h *BaseHandler) GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", false
	}

	return userIDStr, true
}
