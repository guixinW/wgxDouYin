package service

import (
	"context"
	"wgxDouYin/dal/db"
	"wgxDouYin/grpc/user"
	pb "wgxDouYin/grpc/user"
	"wgxDouYin/internal/tool"
	"wgxDouYin/pkg/zap"
)

type UserServerImpl struct {
	pb.UnimplementedUserServiceServer
}

func (s *UserServerImpl) UserRegister(ctx context.Context, req *user.UserRegisterRequest) (resp *user.UserRegisterResponse, err error) {
	logger := zap.InitLogger()
	usr, err := db.GetUserByName(ctx, req.Username)
	if err != nil {
		logger.Errorln(err.Error())
		res := &user.UserRegisterResponse{
			StatusCode: -1,
			StatusMsg:  "注册失败，getUserByName服务器内部错误",
		}
		return res, nil
	} else if usr != nil {
		logger.Errorf("该用户名已存在:%s", usr.UserName)
		res := &user.UserRegisterResponse{
			StatusCode: -1,
			StatusMsg:  "该用户名已存在，请更换",
		}
		return res, nil
	}
	usr = &db.User{
		UserName: req.Username,
		Password: tool.PasswordEncrypt(req.Password),
	}
	if err := db.CreateUser(ctx, usr); err != nil {
		logger.Errorln(err.Error())
		res := &user.UserRegisterResponse{
			StatusCode: -1,
			StatusMsg:  "注册失败，CreateUser服务器内部错误",
		}
		return res, nil
	}
	res := &user.UserRegisterResponse{
		StatusCode: 0,
		StatusMsg:  "注册成功",
		UserId:     int64(usr.ID),
		Token:      "",
	}
	return res, nil
}

func (s *UserServerImpl) UserLogin(ctx context.Context, req *user.UserLoginRequest) (resp *user.UserLoginResponse, err error) {
	logger := zap.InitLogger()

	usr, err := db.GetUserByName(ctx, req.Username)
	if err != nil {
		logger.Errorln(err.Error())
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "登陆失败：服务器内部错误",
		}
		return res, err
	} else if usr == nil {
		logger.Errorln("用户名或密码错误")
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "用户名不存在",
		}
		return res, nil
	}
	return nil, nil
}
