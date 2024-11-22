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
