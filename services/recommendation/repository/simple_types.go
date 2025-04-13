package repository

import (
	"time"
)

// SimpleUserActivityRecord 简化的用户活动记录
type SimpleUserActivityRecord struct {
	ID           string    
	UserID       string    
	ActivityType string    
	TargetType   string    
	TargetID     string    
	Duration     int       
	Weight       float64   
	CreatedAt    time.Time 
}

// SimpleRecommendationItem 简化的推荐项
type SimpleRecommendationItem struct {
	ID        string    
	UserID    string    
	ItemType  string    
	ItemID    string    
	Score     float64   
	ModelID   string    
	Reason    string    
	ExpiresAt time.Time 
	IsClicked bool      
	IsViewed  bool      
	CreatedAt time.Time 
}

// SimpleUserFeed 简化的用户Feed
type SimpleUserFeed struct {
	ID          string    
	UserID      string    
	Items       []string  
	LastUpdated time.Time 
	ExpiresAt   time.Time 
	CreatedAt   time.Time 
} 