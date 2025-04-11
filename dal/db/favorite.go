package db

import (
	"context"
	"fmt"
	"github.com/go-sql-driver/mysql"
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

const (
	FieldFavoriteCount = "favorite_count"
	FieldDislikeCount  = "dislike_count"
)

func UpdateVideoAndUser(tx *gorm.DB, newVideoRelation *FavoriteVideoRelation, incr int, field string) error {
	// 更新 Video 表
	res := tx.Model(&Video{}).Where("id = ?", newVideoRelation.VideoID).Update(field, gorm.Expr(fmt.Sprintf("%s + ?", field), incr))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return ErrorDatabase()
	}

	// 更新 User 表
	res = tx.Model(&User{}).Where("id = ?", newVideoRelation.UserID).Update(field, gorm.Expr(fmt.Sprintf("%s + ?", field), incr))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return ErrorDatabase()
	}
	return nil
}

func CreateVideoRelation(ctx context.Context, newVideoRelation *FavoriteVideoRelation) error {
	if !(newVideoRelation.ActionType == favorite.VideoActionType_LIKE || newVideoRelation.ActionType == favorite.VideoActionType_DISLIKE) {
		return errors.New("错误的操作类型")
	}
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existVideoRelation FavoriteVideoRelation
		//查找待插入的记录是否存在
		err := tx.Where("user_id = ? AND video_id = ?", newVideoRelation.UserID, newVideoRelation.VideoID).First(&existVideoRelation).Error

		//如果不存在则可以插入
		if errors.Is(err, gorm.ErrRecordNotFound) {
			createRes := tx.Create(newVideoRelation)
			if createRes.Error != nil {
				var mysqlErr *mysql.MySQLError
				//video外键冲突
				if errors.As(createRes.Error, &mysqlErr) {
					if mysqlErr.Number == 1452 {
						return errors.New("喜欢或者不喜欢的视频不存在")
					}
				}
				return createRes.Error
			}
			switch newVideoRelation.ActionType {
			case favorite.VideoActionType_LIKE:
				err := UpdateVideoAndUser(tx, newVideoRelation, 1, FieldFavoriteCount)
				if err != nil {
					return err
				}
			case favorite.VideoActionType_DISLIKE:
				fmt.Println("dislike")
				err := UpdateVideoAndUser(tx, newVideoRelation, 1, FieldDislikeCount)
				if err != nil {
					return err
				}
			}
		} else if err != nil {
			return err
		} else if existVideoRelation.ActionType != newVideoRelation.ActionType {
			//如果存在且类型发生改变则修改VideoActionType
			err := tx.Model(&FavoriteVideoRelation{}).Where("user_id = ? AND video_id = ?", newVideoRelation.UserID, newVideoRelation.VideoID).Update("action_type", newVideoRelation.ActionType).Error
			if err != nil {
				return err
			}
			err = tx.Where("user_id = ? AND video_id = ?", newVideoRelation.UserID, newVideoRelation.VideoID).First(&newVideoRelation).Error
			if err != nil {
				return err
			}
			if newVideoRelation.ActionType == favorite.VideoActionType_LIKE {
				err := UpdateVideoAndUser(tx, newVideoRelation, 1, FieldFavoriteCount)
				if err != nil {
					return err
				}
				err = UpdateVideoAndUser(tx, newVideoRelation, -1, FieldDislikeCount)
				if err != nil {
					return err
				}
			}
			if newVideoRelation.ActionType == favorite.VideoActionType_DISLIKE {
				err := UpdateVideoAndUser(tx, newVideoRelation, 1, FieldDislikeCount)
				if err != nil {
					return err
				}
				err = UpdateVideoAndUser(tx, newVideoRelation, -1, FieldFavoriteCount)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

func DelVideoRelation(ctx context.Context, delVideoRelation *FavoriteVideoRelation) error {
	if !(delVideoRelation.ActionType == favorite.VideoActionType_CANCEL_LIKE || delVideoRelation.ActionType == favorite.VideoActionType_LIKE) {
		return errors.New("错误的操作类型")
	}

	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//查询是否存在
		var existVideoRelation FavoriteVideoRelation
		err := tx.Where("user_id = ? AND video_id = ?", delVideoRelation.UserID, delVideoRelation.VideoID).First(&existVideoRelation).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if delVideoRelation.ActionType == favorite.VideoActionType_CANCEL_LIKE {
				return errors.New("无法取消一个未点赞的视频")
			}
			if delVideoRelation.ActionType == favorite.VideoActionType_DISLIKE {
				return errors.New("无法取消未不喜欢的视频")
			}
		}
		if err != nil {
			return err
		}
		if existVideoRelation.ActionType == favorite.VideoActionType_LIKE &&
			delVideoRelation.ActionType == favorite.VideoActionType_CANCEL_LIKE {
			err = tx.Unscoped().Delete(delVideoRelation).Error
			if err != nil {
				return err
			}
		} else if existVideoRelation.ActionType == favorite.VideoActionType_DISLIKE &&
			delVideoRelation.ActionType == favorite.VideoActionType_CANCEL_DISLIKE {
			err = tx.Unscoped().Delete(delVideoRelation).Error
			if err != nil {
				return err
			}
		} else {
			return errors.New("无法取消未点赞、未不喜欢")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
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
