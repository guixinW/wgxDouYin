package service

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"wgxDouYin/dal/db"
	user "wgxDouYin/grpc/user"
	"wgxDouYin/internal/tool"
	myJwt "wgxDouYin/pkg/jwt"
	"wgxDouYin/pkg/zap"
)

type UserServiceImpl struct {
	user.UnimplementedUserServiceServer
}

func (s *UserServiceImpl) UserRegister(ctx context.Context, req *user.UserRegisterRequest) (resp *user.UserRegisterResponse, err error) {
	usr, err := db.GetUserByName(ctx, req.Username)
	if err != nil {
		logger.Errorln(err.Error())
		res := &user.UserRegisterResponse{
			StatusCode: -1,
			StatusMsg:  "注册失败，getUserByName服务器内部错误",
		}
		return res, err
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
		UserId:     uint64(usr.ID),
		Token:      "",
	}
	return res, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, req *user.UserLoginRequest) (resp *user.UserLoginResponse, err error) {
	logger := zap.InitLogger()
	start := time.Now()
	usr, err := db.GetUserByName(ctx, req.Username)
	end := time.Now()
	fmt.Printf("数据库查询耗时：%v\n", end.Sub(start))
	if err != nil {
		logger.Errorln(err.Error())
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "登陆失败：服务器内部错误",
		}
		return res, err
	} else if usr == nil {
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "用户名不存在",
		}
		return res, nil
	}
	start = time.Now()
	if tool.PasswordCompare(req.Password, usr.Password) == false {
		logger.Errorf("%v尝试登录，但是密码%v错误", req.Username, req.Password)
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "用户名或密码错误",
		}
		return res, nil
	}
	end = time.Now()
	fmt.Printf("比较密钥耗时：%v\n", end.Sub(start))
	claims := myJwt.CustomClaims{
		UserId: uint64(usr.ID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			Issuer:    "Login",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	start = time.Now()
	token, err := myJwt.CreateToken(KeyManager.GetPrivateKey(), claims)
	end = time.Now()
	fmt.Printf("签发token耗时：%v\n", end.Sub(start))
	if err != nil {
		logger.Errorf("发生错误:%v", err.Error())
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：token 创建失败",
		}
		return res, nil
	}
	res := &user.UserLoginResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserId:     uint64(usr.ID),
		Token:      token,
	}
	return res, nil
}

func (s *UserServiceImpl) UserInfo(ctx context.Context, req *user.UserInfoRequest) (resp *user.UserInfoResponse, err error) {
	logger := zap.InitLogger()
	usr, err := db.GetUserByID(ctx, req.UserId)
	if err != nil {
		logger.Errorln(err.Error())
		res := &user.UserInfoResponse{
			StatusCode: -1,
			StatusMsg:  "获取用户信息失败：服务器内部错误",
		}
		return res, err
	} else if usr == nil {
		res := &user.UserInfoResponse{
			StatusCode: -1,
			StatusMsg:  "用户名不存在",
		}
		return res, nil
	} else {
		res := &user.UserInfoResponse{
			StatusCode: 0,
			StatusMsg:  "success",
			User: &user.User{
				Id:             uint64(usr.ID),
				Name:           usr.UserName,
				FollowerCount:  usr.FollowerCount,
				FollowingCount: usr.FollowingCount,
			},
		}
		return res, nil
	}
}
