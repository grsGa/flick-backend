package dto

import (
	"time"
)

// InteractionDTO 用于API交互的数据传输对象
type InteractionDTO struct {
	ID        uint      `json:"id,omitempty"`
	UserID    uint      `json:"user_id"`
	ContentID uint      `json:"content_id"`
	Type      string    `json:"type"`
	Value     int       `json:"value"`
	Data      string    `json:"data,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// InteractionCreateRequest 创建交互的请求
type InteractionCreateRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`
	ContentID uint   `json:"content_id" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Value     int    `json:"value"`
	Data      string `json:"data,omitempty"`
}

// InteractionUpdateRequest 更新交互的请求
type InteractionUpdateRequest struct {
	Value int    `json:"value"`
	Data  string `json:"data,omitempty"`
}

// InteractionResponse 交互的响应
type InteractionResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Data    InteractionDTO `json:"data,omitempty"`
}

// InteractionsListResponse 交互列表的响应
type InteractionsListResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Total   int             `json:"total"`
	Data    []InteractionDTO `json:"data"`
} 