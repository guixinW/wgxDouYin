package db

import (
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	Video      Video  `gorm:"foreignkey:VideoID" json:"video,omitempty"`
	VideoID    uint   `gorm:"index:idx_videoid;not null" json:"video_id"`
	User       User   `gorm:"foreignkey:UserID" json:"user,omitempty"`
	UserID     uint   `gorm:"index:idx_userid;not null" json:"user_id"`
	Content    string `gorm:"type:varchar(255);not null" json:"content"`
	LikeCount  uint   `gorm:"column:like_count;default:0;not null" json:"like_count,omitempty"`
	TeaseCount uint   `gorm:"column:tease_count;default:0;not null" json:"tease_count,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}
