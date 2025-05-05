package service

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
	"wgxDouYin/dal/db"
	"wgxDouYin/grpc/user"
	"wgxDouYin/internal/tool"
	myJwt "wgxDouYin/pkg/jwt"
)

const (
	AccessTokenExpireTime  = 60 * 60 * 24 * 7
	RefreshTokenExpireTime = 60 * 60 * 24 * 7
)

type UserServiceImpl struct {
	user.UnimplementedUserServiceServer
}

func (s *UserServiceImpl) UserRegister(ctx context.Context, req *user.UserRegisterRequest) (resp *user.UserRegisterResponse, err error) {
	usr, err := db.GetUserNameIdAndPasswordByName(ctx, req.Username)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Errorln(err.Error())
			res := &user.UserRegisterResponse{
				StatusCode: -1,
				StatusMsg:  "注册失败，getUserByName服务器内部错误",
			}
			return res, err
		}
	}
	if usr != nil {
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

func (s *UserServiceImpl) RefreshAccessToken(ctx context.Context, req *user.AccessTokenRequest) (resp *user.AccessTokenResponse, err error) {
	//通过签发的token解析的userID不可能不存在
	userDeviceAndRefreshToken, err := db.GetRefreshTokenAndDeviceIdByUserId(ctx, req.UserId)
	if err != nil {
		res := &user.AccessTokenResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误",
		}
		return res, nil
	}
	if tool.GenerateHashOfLength64(req.RefreshToken) != userDeviceAndRefreshToken.RefreshHashedToken ||
		req.DeviceId != userDeviceAndRefreshToken.DeviceId {
		res := &user.AccessTokenResponse{
			StatusCode: -1,
			StatusMsg:  "该账号存在安全风险",
		}
		return res, nil
	}
	accessClaims := myJwt.CustomClaims{
		UserId: req.UserId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpireTime * time.Second)),
			Issuer:    "Login",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, err := myJwt.CreateToken(KeyManager.GetPrivateKey(), accessClaims)
	if err != nil {
		logger.Errorf("发生错误:%v", err.Error())
		res := &user.AccessTokenResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：token 创建失败",
		}
		return res, nil
	}
	res := &user.AccessTokenResponse{
		AccessToken: accessToken,
		StatusCode:  0,
		StatusMsg:   "success",
	}
	return res, nil
}

// Login 登陆
func (s *UserServiceImpl) Login(ctx context.Context, req *user.UserLoginRequest) (resp *user.UserLoginResponse, err error) {
	//1.查询用户是否存在
	usr, err := db.GetUserNameIdAndPasswordByName(ctx, req.Username)
	if err != nil {
		var res user.UserLoginResponse
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res = user.UserLoginResponse{
				StatusCode: -1,
				StatusMsg:  "登陆失败：不存在该用户",
			}
		} else {
			res = user.UserLoginResponse{
				StatusCode: -1,
				StatusMsg:  "登陆失败：服务器内部错误",
			}
		}
		return &res, err
	}
	//2.比较密码是否正确
	if tool.PasswordCompare(req.Password, usr.Password) == false {
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "用户名或密码错误",
		}
		return res, nil
	}

	//3.签发token
	refreshClaims := myJwt.CustomClaims{
		UserId: uint64(usr.ID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpireTime * time.Second)),
			Issuer:    "Login",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessClaims := myJwt.CustomClaims{
		UserId: uint64(usr.ID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpireTime * time.Second)),
			Issuer:    "Login",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken, err := myJwt.CreateToken(KeyManager.GetPrivateKey(), refreshClaims)
	if err != nil {
		logger.Errorf("发生错误:%v", err.Error())
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：token 创建失败",
		}
		return res, nil
	}
	accessToken, err := myJwt.CreateToken(KeyManager.GetPrivateKey(), accessClaims)
	if err != nil {
		logger.Errorf("发生错误:%v", err.Error())
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：token 创建失败",
		}
		return res, nil
	}
	refreshTokenHashed := tool.GenerateHashOfLength64(refreshToken)
	if err := db.UpdateDeviceIdAndRefreshToken(ctx, usr, req.DeviceId, refreshTokenHashed); err != nil {
		logger.Errorf("发生错误:%v", err.Error())
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "登陆失败",
		}
		return res, nil
	}

	//4.返回token
	res := &user.UserLoginResponse{
		StatusCode:   0,
		StatusMsg:    "success",
		UserId:       uint64(usr.ID),
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}
	return res, nil
}

// UserInfo 接收一个包含tokenUserId、queryUserID的rpc请求。返回queryUserID的信息
// 由于tokenUserID能够确定是登陆用户发送的，因此无需查看该id的合法性
func (s *UserServiceImpl) UserInfo(ctx context.Context, req *user.UserInfoRequest) (resp *user.UserInfoResponse, err error) {
	//用户查看自身信息与查看其他用户信息需要区分开，
	//A.用户查看自身信息通过MySQL查询详细信息，包含favorite_count、dislike_count（其他用户查询则不包含）
	//B.用户查看其他用户信息通过Redis查询该用户是否为热点用户，如果是则直接返回Redis中的数据，如果不是则查询MySQL
	//查询步骤：
	//1.先通过Redis查看用户是否为热点用户，如果是则直接使用Redis
	//2.如果是
	if req.QueryUserId == req.TokenUserId {
		return querySelf(ctx, req.QueryUserId)
	}
	return queryOtherUser(ctx, req.QueryUserId)
}
