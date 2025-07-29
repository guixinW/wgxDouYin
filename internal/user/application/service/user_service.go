package service

import (
	"context"
	"errors"
	"wgxDouYin/internal/user/domain/model"
	"wgxDouYin/internal/user/domain/repository"
)

var (
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotFound     = errors.New("user not found")
	ErrPassword         = errors.New("password error")
)

// UserService 应用服务
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建一个新的 UserService
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Register 处理用户注册
func (s *UserService) Register(ctx context.Context, username, password string) (*model.User, error) {
	// 检查用户是否已存在
	_, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil {
		return nil, ErrUserAlreadyExist
	}

	// 创建用户
	user := &model.User{
		Username: username,
		Password: password, // 实际项目中需要加密
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 处理用户登录
func (s *UserService) Login(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 实际项目中需要比较加密后的密码
	if user.Password != password {
		return nil, ErrPassword
	}

	return user, nil
}

// GetUserInfo 处理获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
