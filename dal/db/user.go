package db

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type User struct {
	gorm.Model
	UserName           string  `gorm:"index:idx_username, unique;type:varchar(40);not null" json:"name,omitempty"`
	Password           string  `gorm:"type:varchar(256);not null" json:"password,omitempty"`
	FavoriteVideos     []Video `gorm:"many2many:user_favorite_videos" json:"favorite_videos,omitempty"`
	FollowingCount     uint64  `gorm:"default:0;not null" json:"follow_count,omitempty"`
	FollowerCount      uint64  `gorm:"default:0;not null" json:"follower_count,omitempty"`
	WorkCount          uint64  `gorm:"default:0;not null" json:"work_count,omitempty"`
	FavoriteCount      uint64  `gorm:"default:0;not null" json:"favorite_count,omitempty"`
	DislikeCount       uint64  `gorm:"default:0;not null" json:"dislike_count,omitempty"`
	RefreshHashedToken string  `gorm:"type:char(64)" json:"refresh_hashed_token,omitempty"`
	DeviceId           string  `gorm:"type:char(36)" json:"device_id,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// GetUserByID 根据用户id获取用户
func GetUserByID(ctx context.Context, userID uint64) (*User, error) {
	res := new(User)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).First(&res, userID).Error; err == nil {
		return res, nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else {
		return nil, err
	}
}

// GetUserByIDs 根据用户id列表获取用户列表
func GetUserByIDs(ctx context.Context, userIDs []uint64) ([]*User, error) {
	res := make([]*User, 0)
	if len(userIDs) == 0 {
		return res, nil
	}
	if err := GetDB().WithContext(ctx).Where("id in ?", userIDs).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// GetUserNameIdAndPasswordByName 根据用户名获取用户信息
func GetUserNameIdAndPasswordByName(ctx context.Context, userName string) (*User, error) {
	var res *User
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Select("id, user_name, password").Where("user_name = ?", userName).First(&res).Error; err == nil {
		return res, nil
	} else {
		return nil, err
	}
}

// CreateUser 创建用户
func CreateUser(ctx context.Context, user *User) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

// GetPasswordByUsername 根据用户名获取用户密码
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

func UpdateDeviceIdAndRefreshToken(ctx context.Context, user *User, newDeviceId, newRefreshToken string) error {
	user.DeviceId = newDeviceId
	user.RefreshHashedToken = newRefreshToken
	if err := GetDB().WithContext(ctx).Model(&User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"device_id":            newDeviceId,
		"refresh_hashed_token": newRefreshToken,
	}).Error; err != nil {
		return err
	}
	return nil
}

func GetRefreshTokenAndDeviceIdByUserId(ctx context.Context, userId uint64) (*User, error) {
	var res *User
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Select("refresh_hashed_token, device_id").Where("id = ?", userId).First(&res).Error; err == nil {
		return res, nil
	} else {
		return nil, err
	}
}
