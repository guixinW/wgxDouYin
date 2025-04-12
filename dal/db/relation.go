package db

import (
	"context"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

type FollowRelation struct {
	gorm.Model
	User     User   `gorm:"foreignkey:UserID;" json:"user,omitempty"`
	UserID   uint64 `gorm:"index:idx_userid;not null" json:"user_id"`
	ToUser   User   `gorm:"foreignkey:ToUserID;" json:"to_user,omitempty"`
	ToUserID uint64 `gorm:"index:idx_userid;index:idx_userid_to;not null" json:"to_user_id"`
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

func CreateRelation(ctx context.Context, followRelation *FollowRelation) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		createRes := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "to_user_id"}},
			DoNothing: true,
		}).Create(followRelation)

		//判断是否出现错误以及是否未正常插入
		if createRes.Error != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(createRes.Error, &mysqlErr) {
				if mysqlErr.Number == 1452 {
					return errors.New("关注的用户不存在")
				}
			}
			return createRes.Error
		}
		if createRes.RowsAffected == 0 {
			return errors.New("重复关注")
		}

		//更改被关注人的follower_count以及关注人的following_count
		res := tx.Model(&User{}).Where("id = ?", followRelation.UserID).
			Update("following_count",
				gorm.Expr("following_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(followRelation.UserID, "CreateRelation", Update)
		}
		res = tx.Model(&User{}).Where("id = ?", followRelation.ToUserID).
			Update("follower_count",
				gorm.Expr("follower_count + ?", 1))
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
		res := tx.Unscoped().Delete(followRelation)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("不能取关你未关注的人")
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
