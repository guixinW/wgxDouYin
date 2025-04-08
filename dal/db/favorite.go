package db

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"wgxDouYin/grpc/favorite"
)

type FavoriteVideoRelation struct {
	gorm.Model
	Video      Video                    `gorm:"foreignkey:VideoID;" json:"video,omitempty"`
	VideoID    uint64                   `gorm:"index:idx_videoid;not null" json:"video_id"`
	User       User                     `gorm:"foreignkey:UserID;" json:"user,omitempty"`
	UserID     uint64                   `gorm:"index:idx_userid;not null" json:"user_id"`
	ActionType favorite.VideoActionType `gorm:"not null" json:"action_type"`
}

func (FavoriteVideoRelation) TableName() string {
	return "user_favorite_videos"
}

func ErrorDatabase() error {
	return errors.New("Error Database")
}

func CreateVideoAction(ctx context.Context, userId uint64, videoId uint64, actionType favorite.VideoActionType) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&FavoriteVideoRelation{UserID: userId, VideoID: videoId, ActionType: actionType}).Error
		if err != nil {
			return err
		}
		if actionType == favorite.VideoActionType_LIKE {
			//修改video表中的favorite_count
			res := tx.Model(&Video{}).Where("id = ?", videoId).Update("favorite_count", gorm.Expr("favorite_count + ?", 1))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return ErrorDatabase()
			}

			//修改user表中的favorite_count
			res = tx.Model(&User{}).Where("id = ?", userId).Update("favorite_count", gorm.Expr("favorite_count + ?", 1))
			if res.Error != nil {
				return err
			}
			if res.RowsAffected != 1 {
				return ErrorDatabase()
			}
		}
		return nil
	})
	return err
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

func DelFavoriteByUserVideoID(ctx context.Context, userID uint64, videoID uint64) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		FavoriteVideoRelation := new(FavoriteVideoRelation)
		if err := tx.Where("user_id = ? and video_id = ?", userID, videoID).First(&FavoriteVideoRelation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		//删除点赞数据
		err := tx.Unscoped().Where("user_id = ? and video_id = ?", userID, videoID).Delete(&FavoriteVideoRelation).Error
		if err != nil {
			return err
		}
		if FavoriteVideoRelation.ActionType == favorite.VideoActionType_LIKE {
			//修改 video 表中的 favorite_count
			res := tx.Model(&Video{}).Where("id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count - ?", 1))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return ErrorDatabase()
			}

			//修改 user 表中的 favorite_count
			res = tx.Model(&User{}).Where("id = ?", userID).Update("favorite_count", gorm.Expr("favorite_count - ?", 1))
			if res.Error != nil {
				return err
			}
			if res.RowsAffected != 1 {
				return ErrorDatabase()
			}
		}
		return nil
	})
	return err
}
