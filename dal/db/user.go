package db

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type User struct {
	gorm.Model
	UserName       string  `gorm:"index:idx_username, unique;type:varchar(40);not null" json:"name,omitempty"`
	Password       string  `gorm:"type:varchar(256);not null" json:"password,omitempty"`
	FavoriteVideos []Video `gorm:"many2many:user_favorite_videos" json:"favorite_videos,omitempty"`
	FollowingCount uint    `gorm:"default:0;not null" json:"follow_count,omitempty"`
	FollowerCount  uint    `gorm:"default:0;not null" json:"follower_count,omitempty"`
}

func (User) TableName() string {
	return "users"
}

func GetUserByID(ctx context.Context, userIDs []int64) ([]*User, error) {
	res := make([]*User, 0)
	if len(userIDs) == 0 {
		return res, nil
	}
	if err := GetDB().WithContext(ctx).Where("id in ?", userIDs).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func GetUserByName(ctx context.Context, userName string) (*User, error) {
	res := new(User)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Select("id, user_name, password").Where("user_name = ?", userName).First(&res).Error; err == nil {
		return res, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return nil, err
	}
}

func CreateUser(ctx context.Context, user *User) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

// GetPasswordByUsername
//
//	@Description: 根据用户名获取密码
//	@Date 2023-01-21 17:15:17
//	@param ctx 数据库操作上下文
//	@param userName 用户名
//	@return *User 用户
//	@return error
func GetPasswordByUsername(ctx context.Context, userName string) (*User, error) {
	user := new(User)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).
		Select("password").Where("user_name = ?", userName).
		First(&user).Error; err == nil {
		return user, nil
	} else {
		return nil, err
	}
}
