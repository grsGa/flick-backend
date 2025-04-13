package models

import (
	"time"
)

// Interaction 数据库交互模型
type Interaction struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	ContentID uint      `gorm:"index"`
	Type      string    `gorm:"type:varchar(50);index"`
	Value     int       `gorm:"default:0"`
	Data      string    `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 设置表名
func (Interaction) TableName() string {
	return "interactions"
} 