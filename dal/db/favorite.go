package db

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type FavoriteVideoRelation struct {
	gorm.Model
	Video   Video  `gorm:"foreignkey:VideoID;" json:"video,omitempty"`
	VideoID uint64 `gorm:"index:idx_videoid;not null" json:"video_id"`
	User    User   `gorm:"foreignkey:UserID;" json:"user,omitempty"`
	UserID  uint64 `gorm:"index:idx_userid;not null" json:"user_id"`
}

func (FavoriteVideoRelation) TableName() string {
	return "user_favorite_videos"
}

func GetFavoriteVideoRelationByUserVideoID(ctx context.Context, userID uint64, videoID uint64) (*FavoriteVideoRelation, error) {
	FavoriteVideoRelation := new(FavoriteVideoRelation)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).First(&FavoriteVideoRelation, "user_id = ? and video_id = ?", userID, videoID).Error; err == nil {
		return FavoriteVideoRelation, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return nil, err
	}
}
