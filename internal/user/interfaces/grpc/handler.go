package grpc

import (
	"context"
	"wgxDouYin/internal/user/application/service"
	"wgxDouYin/internal/user/domain/model"
	"wgxDouYin/pkg/jwt"
	pb "wgxDouYin/grpc/user"
)

// UserServerImpl 是 user gRPC 服务的实现
type UserServerImpl struct {
	pb.UnimplementedUserServer
	userService *service.UserService
}

// NewUserServerImpl 创建一个新的 UserServerImpl
func NewUserServerImpl(userService *service.UserService) *UserServerImpl {
	return &UserServerImpl{
		userService: userService,
	}
}

func (s *UserServerImpl) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := s.userService.Register(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.RegisterResponse{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		}, nil
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return &pb.RegisterResponse{
			StatusCode: -1,
			StatusMsg:  "failed to generate token",
		}, nil
	}

	return &pb.RegisterResponse{
		StatusCode: 0,
		StatusMsg:  "register success",
		UserId:     user.ID,
		Token:      token,
	}, nil
}

func (s *UserServerImpl) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.LoginResponse{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		}, nil
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return &pb.LoginResponse{
			StatusCode: -1,
			StatusMsg:  "failed to generate token",
		}, nil
	}

	return &pb.LoginResponse{
		StatusCode: 0,
		StatusMsg:  "login success",
		UserId:     user.ID,
		Token:      token,
	}, nil
}

func (s *UserServerImpl) UserInfo(ctx context.Context, req *pb.UserInfoRequest) (*pb.UserInfoResponse, error) {
	user, err := s.userService.GetUserInfo(ctx, req.UserId)
	if err != nil {
		return &pb.UserInfoResponse{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		}, nil
	}

	return &pb.UserInfoResponse{
		StatusCode: 0,
		StatusMsg:  "get user info success",
		User:       toProtoUser(user),
	}, nil
}

// toProtoUser 将 model.User 转换为 pb.User
func toProtoUser(user *model.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Id:            user.ID,
		Name:          user.Username,
		FollowCount:   user.FollowCount,
		FollowerCount: user.FollowerCount,
		IsFollow:      user.IsFollow,
	}
}
