package db

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"time"
)

type Video struct {
	gorm.Model
	Author        User   `gorm:"foreignkey:AuthorID" json:"author,omitempty"`
	AuthorID      uint64 `gorm:"index:idx_authorid;not null" json:"author_id,omitempty"`
	PlayUrl       string `gorm:"type:varchar(255);not null" json:"play_url,omitempty"`
	FavoriteCount uint64 `gorm:"default:0;not null" json:"favorite_count,omitempty"`
	DislikeCount  uint64 `gorm:"default:0;not null" json:"dislike_count,omitempty"`
	CommentCount  uint64 `gorm:"default:0;not null" json:"comment_count,omitempty"`
	Title         string `gorm:"type:varchar(50);not null" json:"title,omitempty"`
}

func (Video) TableName() string {
	return "videos"
}

// CreateVideo 创建视频
func CreateVideo(ctx context.Context, video *Video) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(video).Error
		if err != nil {
			return err
		}
		res := tx.Model(&User{}).Where("id = ?", video.AuthorID).Update("work_count", gorm.Expr("work_count + ?", 1))
		if res.Error != nil {
			return err
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("create video failed")
		}
		return nil
	})
	return err
}

// MGetVideos 获取视频
func MGetVideos(ctx context.Context, limit int, latestTime *int64) ([]*Video, error) {
	videos := make([]*Video, 0)
	if latestTime == nil || *latestTime == 0 {
		curTime := time.Now().UnixMilli()
		latestTime = &curTime
	}
	conn := GetDB().Clauses(dbresolver.Read).WithContext(ctx)
	if err := conn.Limit(limit).Order("created_at desc").Find(&videos, "created_at < ?", time.UnixMilli(*latestTime)).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

func GetVideosByUserID(ctx context.Context, authorId uint64) ([]*Video, error) {
	var pubList []*Video
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Model(&Video{}).Where(&Video{AuthorID: authorId}).Find(&pubList).Error
	if err != nil {
		return nil, err
	}
	return pubList, nil
}

func DeleteVideoByID(ctx context.Context, videoID uint, authorID uint64) error {
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().Delete(&Video{}, videoID).Error
		if err != nil {
			return err
		}
		res := tx.Model(&User{}).Where("id = ?", authorID).Update("work_count", gorm.Expr("work_count - ?", 1))
		if res.Error != nil {
			return err
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("delete video failed")
		}
		return nil
	})
	return err
}
