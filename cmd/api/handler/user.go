package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"wgxDouYin/cmd/api/rpc"
	grpc "wgxDouYin/grpc/user"
	"wgxDouYin/internal/response"
)

func UserRegister(c *gin.Context) {
	userName := c.Query("username")
	password := c.Query("password")

	if len(userName) == 0 || len(password) == 0 {
		c.JSON(http.StatusBadRequest, response.Register{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "用户名或者密码不能为空",
			},
		})
		return
	}
	if len(userName) > 32 || len(password) > 32 {
		c.JSON(http.StatusBadRequest, response.Register{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "用户名或者密码长度不能大于32个字符",
			},
		})
		return
	}
	req := &grpc.UserRegisterRequest{
		Username: userName,
		Password: password,
	}
	res, err := rpc.Register(c, req)
	if err != nil {
		fmt.Printf("rpc register err:%v\n", err)
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.Register{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
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
	username := c.Query("username")
	password := c.Query("password")
	if len(username) == 0 || len(password) == 0 {
		c.JSON(http.StatusBadRequest, response.Login{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "用户名或者密码不能为空",
			},
		})
		return
	}
	req := &grpc.UserLoginRequest{
		Username: username,
		Password: password,
	}
	res, _ := rpc.Login(c, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.Login{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.Login{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserID: res.UserId,
		Token:  res.Token,
	})
}

func UserInform(c *gin.Context) {
	userId, exist := c.Get("UserID")
	if !exist {
		c.JSON(http.StatusBadRequest, response.UserInform{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "用户名或者密码不能为空",
			},
			User: nil,
		})
		return
	}
	id := userId.(uint64)
	req := &grpc.UserInfoRequest{
		UserId: id,
		Token:  "",
	}
	res, _ := rpc.UserInform(c, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.UserInform{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
			User: nil,
		})
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
