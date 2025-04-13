package utils

import (
	"bytes"
	"github.com/google/uuid"
	"html/template"
	"strings"
)

// GenerateUUID 生成UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// Contains 检查切片是否包含某个值
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == "all" || s == item {
			return true
		}
	}
	return false
}

// RenderTemplate 渲染模板字符串
func RenderTemplate(templateContent string, data map[string]interface{}) string {
	if templateContent == "" {
		return ""
	}
	
	tmpl, err := template.New("notification").Parse(templateContent)
	if err != nil {
		return templateContent
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return templateContent
	}

	return buf.String()
}

// FormatToTitleCase 将字符串格式化为标题样式
func FormatToTitleCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// TruncateString 截断字符串，添加省略号
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
} 