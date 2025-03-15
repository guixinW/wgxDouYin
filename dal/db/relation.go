package db

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type FollowRelation struct {
	gorm.Model
	User     User `gorm:"foreignkey:UserID;" json:"user,omitempty"`
	UserID   uint `gorm:"index:idx_userid;not null" json:"user_id"`
	ToUser   User `gorm:"foreignkey:ToUserID;" json:"to_user,omitempty"`
	ToUserID uint `gorm:"index:idx_userid;index:idx_userid_to;not null" json:"to_user_id"`
}

func (FollowRelation) TableName() string {
	return "relations"
}

func GetRelationByUserIDs(ctx context.Context, userId uint64, toUserID uint64) (*FollowRelation, error) {
	relation := new(FollowRelation)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Where("user_id=? and to_user_id=?", userId, toUserID).First(&relation).Error; err == nil {
		return relation, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return nil, NewDatabaseErrorMessage(userId, "GetRelationByUserIDs", Find)
	}
}

func CreateRelation(ctx context.Context, userID uint64, toUserID uint64) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&FollowRelation{UserID: uint(userID), ToUserID: uint(toUserID)}).Error
		if err != nil {
			return err
		}
		res := tx.Model(&User{}).Where("id = ?", userID).
			Update("following_count",
				gorm.Expr("following_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(userID, "CreateRelation", Update)
		}
		res = tx.Model(&User{}).Where("id = ?", toUserID).
			Update("follower_count",
				gorm.Expr("follower_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(toUserID, "CreateRelation", Update)
		}
		return nil
	})
	return err
}

func DelRelationByUserID(ctx context.Context, userID uint64, toUserID uint64) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		relation := new(FollowRelation)
		if err := tx.Where("user_id = ? AND to_user_id=?", userID, toUserID).First(&relation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		err := tx.Unscoped().Delete(&relation).Error
		if err != nil {
			return err
		}
		res := tx.Model(&User{}).Where("id = ?", userID).Update("following_count", gorm.Expr("following_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(userID, "DelRelationByUserIDs", Delete)
		}
		res = tx.Model(&User{}).Where("id = ?", toUserID).Update("follower_count", gorm.Expr("follower_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(toUserID, "DelRelationByUserIDs", Delete)
		}
		return nil
	})
	return err
}

func GetFollowingListByUserID(ctx context.Context, userID int64) ([]*FollowRelation, error) {
	var RelationList []*FollowRelation
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Where("user_id = ?", userID).Find(&RelationList).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetFollowingListByUserID")
	}
	return RelationList, nil
}

func GetFollowerListByUserID(ctx context.Context, toUserID int64) ([]*FollowRelation, error) {
	var RelationList []*FollowRelation
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Where("to_user_id = ?", toUserID).Find(&RelationList).Error
	if err != nil {
		return nil, errors.Wrap(err, "GetFollowerListByUserID")
	}
	return RelationList, nil
}

func GetFriendList(ctx context.Context, userID int64) ([]*FollowRelation, error) {
	var FriendList []*FollowRelation
	err := GetDB().WithContext(ctx).
		Table("relations").
		Select("user_id, to_user_id, created_at").
		Where("user_id = ? AND to_user_id IN (SELECT user_id FROM relations r WHERE r.to_user_id = relations.user_id)", userID).
		Clauses(dbresolver.Read).
		Scan(&FriendList).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "GetFriendList")
	}
	return FriendList, nil
}
