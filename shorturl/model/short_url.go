package model

import "time"

type ShortURL struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	ShortCode   string     `gorm:"type:varchar(16);uniqueIndex:idx_short_code;not null" json:"short_code"`
	OriginalURL string     `gorm:"type:varchar(2048);not null" json:"original_url"`
	UserID      uint64     `gorm:"index:idx_user_id;not null" json:"user_id"`
	VisitCount  int64      `gorm:"default:0" json:"visit_count"`
	ExpireAt    *time.Time `gorm:"index" json:"expire_at,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ShortURL) TableName() string {
	return "short_urls"
}

type VisitLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ShortCode string    `gorm:"type:varchar(16);index:idx_short_code;not null" json:"short_code"`
	IP        string    `gorm:"type:varchar(64)" json:"ip"`
	UserAgent string    `gorm:"type:varchar(512)" json:"user_agent"`
	Referer   string    `gorm:"type:varchar(1024)" json:"referer"`
	VisitedAt time.Time `gorm:"index:idx_visited_at;not null" json:"visited_at"`
}

func (VisitLog) TableName() string {
	return "visit_logs"
}
