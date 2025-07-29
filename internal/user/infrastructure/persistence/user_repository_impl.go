package persistence

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"wgxDouYin/dal/db"
	"wgxDouYin/internal/user/domain/model"
	"wgxDouYin/internal/user/domain/repository"
)

// GormUserRepository 是 UserRepository 的 GORM 实现
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository 创建一个新的 GormUserRepository
func NewGormUserRepository(db *gorm.DB) repository.UserRepository {
	return &GormUserRepository{db: db}
}

// toDomainUser 将 db.User 转换为 model.User
func toDomainUser(user *db.User) *model.User {
	if user == nil {
		return nil
	}
	return &model.User{
		ID:            user.ID,
		Username:      user.Username,
		Password:      user.Password,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
		IsFollow:      user.IsFollow,
	}
}

// fromDomainUser 将 model.User 转换为 db.User
func fromDomainUser(user *model.User) *db.User {
	if user == nil {
		return nil
	}
	return &db.User{
		ID:            user.ID,
		Username:      user.Username,
		Password:      user.Password,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
		IsFollow:      user.IsFollow,
	}
}

func (r *GormUserRepository) Create(ctx context.Context, user *model.User) error {
	dbUser := fromDomainUser(user)
	return r.db.WithContext(ctx).Create(dbUser).Error
}

func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var dbUser db.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(&dbUser), nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var dbUser db.User
	err := r.db.WithContext(ctx).First(&dbUser, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(&dbUser), nil
}

func (r *GormUserRepository) FindUsersByIDs(ctx context.Context, userIDs []int64) ([]*model.User, error) {
	var dbUsers []*db.User
	if err := r.db.WithContext(ctx).Where("id IN ?", userIDs).Find(&dbUsers).Error; err != nil {
		return nil, err
	}
	domainUsers := make([]*model.User, len(dbUsers))
	for i, u := range dbUsers {
		domainUsers[i] = toDomainUser(u)
	}
	return domainUsers, nil
}
