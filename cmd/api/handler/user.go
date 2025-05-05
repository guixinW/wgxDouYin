package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"wgxDouYin/cmd/api/rpc"
	"wgxDouYin/cmd/user/service"
	userGrpc "wgxDouYin/grpc/user"
	"wgxDouYin/internal/response"
)

const (
	DeviceIdLength = 36
)

func UserRegister(c *gin.Context) {
	userName := c.PostForm("username")
	password := c.PostForm("password")
	if len(userName) == 0 || len(password) == 0 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("用户名或者密码不能为空"))
		return
	}
	if len(userName) > 32 || len(password) > 32 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("用户名或者密码长度不能大于32个字符"))
		return
	}
	req := &userGrpc.UserRegisterRequest{
		Username: userName,
		Password: password,
	}
	res, err := rpc.Register(c, req)
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.Register{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserID: res.UserId,
		Token:  res.Token,
	})
}

func UserLogin(c *gin.Context) {
	userName := c.PostForm("username")
	password := c.PostForm("password")
	deviceId := c.PostForm("device_id")
	if len(userName) == 0 || len(password) == 0 || len(deviceId) != DeviceIdLength {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("不合法的登陆请求"))
		return
	}
	req := &userGrpc.UserLoginRequest{
		Username: userName,
		Password: password,
		DeviceId: deviceId,
	}
	res, err := rpc.Login(c, req)
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Path:     "/",
		MaxAge:   service.RefreshTokenExpireTime,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	c.JSON(http.StatusOK, response.Login{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserID:      res.UserId,
		AccessToken: res.AccessToken,
	})
}

func UserInform(c *gin.Context) {
	queryUserId, err := strconv.ParseUint(c.Query("query_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("非法的query_user_id"))
	}
	tokenUserId, exist := c.Get("query_user_id")
	if !exist {
		c.JSON(http.StatusBadRequest, response.ErrorResponse("请求非法"))
		return
	}
	req := &userGrpc.UserInfoRequest{
		QueryUserId: queryUserId,
		TokenUserId: tokenUserId.(uint64),
	}
	res, err := rpc.UserInform(c, req)
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.UserInform{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		User: res.User,
	})
}

func RefreshToken(c *gin.Context) {
	userId, _ := c.Get("token_user_id")
	refreshToken, _ := c.Get("refresh_token")
	deviceId := c.Query("device_id")
	fmt.Printf("refresh, user_id:%v, refresh_token:%v, deviceId:%v\n", userId, refreshToken, deviceId)
	req := &userGrpc.AccessTokenRequest{
		UserId:       userId.(uint64),
		RefreshToken: refreshToken.(string),
		DeviceId:     deviceId,
	}
	res, err := rpc.AccessToken(c, req)
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	c.JSON(http.StatusOK, response.RefreshToken{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		AccessToken: res.AccessToken,
	})
}
