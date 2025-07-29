package repository

import (
	"context"
	"wgxDouYin/internal/user/domain/model"
)

// UserRepository 定义了用户仓库的接口
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByID(ctx context.Context, id int64) (*model.User, error)
	FindUsersByIDs(ctx context.Context, userIDs []int64) ([]*model.User, error)
}
