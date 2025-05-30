package db

import (
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"time"
)

type FollowRelation struct {
	gorm.Model
	User     User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	UserID   uint64 `gorm:"uniqueIndex:unique_user_relation;not null" json:"user_id"`
	ToUser   User   `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
	ToUserID uint64 `gorm:"uniqueIndex:unique_user_relation;not null" json:"to_user_id"`
}

func (f *FollowRelation) TableName() string {
	return "relations"
}

func (f *FollowRelation) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Model(f).Update("updated_at", time.Now())
	return nil
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

func CreateRelation(ctx context.Context, followRelation *FollowRelation) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Unscoped().Where("user_id = ? AND to_user_id = ?", followRelation.UserID, followRelation.ToUserID).First(followRelation)
		if res.Error == nil {
			if followRelation.DeletedAt.Valid {
				updateRes := tx.Model(followRelation).Unscoped().Where("id = ?", followRelation.ID).Updates(map[string]interface{}{
					"deleted_at": nil,
					"updated_at": time.Now(),
				})
				if updateRes.Error != nil {
					return updateRes.Error
				}
			} else {
				return errors.New("已经关注该用户")
			}
		} else if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			createRes := tx.Create(followRelation)
			if createRes.Error != nil {
				return createRes.Error
			}
		} else {
			return res.Error
		}

		//更改被关注人的follower_count以及关注人的following_count
		res = tx.Model(&User{}).Where("id = ?", followRelation.UserID).Update("following_count", gorm.Expr("following_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(followRelation.UserID, "CreateRelation", Update)
		}
		res = tx.Model(&User{}).Where("id = ?", followRelation.ToUserID).Update("follower_count", gorm.Expr("follower_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(followRelation.ToUserID, "CreateRelation", Update)
		}
		return nil
	})
	return err
}

func DelRelationByUserID(ctx context.Context, followRelation *FollowRelation) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("user_id=? and to_user_id=?", followRelation.UserID, followRelation.ToUserID).First(followRelation).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("不能取关你未关注的人")
			}
			return err
		}
		res := tx.Delete(followRelation)
		if res.Error != nil {
			return res.Error
		}
		res = tx.Model(&User{}).Where("id = ?", followRelation.UserID).Update("following_count", gorm.Expr("following_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(followRelation.UserID, "DelRelationByUserIDs", Delete)
		}
		res = tx.Model(&User{}).Where("id = ?", followRelation.ToUserID).Update("follower_count", gorm.Expr("follower_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(followRelation.ToUserID, "DelRelationByUserIDs", Delete)
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
