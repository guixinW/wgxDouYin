package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"wgxDouYin/cmd/api/rpc"
	relationGrpc "wgxDouYin/grpc/relation"
	"wgxDouYin/internal/response"
)

func RelationAction(c *gin.Context) {
	tid, err := strconv.ParseUint(c.PostForm("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.RelationAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "to_user_id 不合法",
			},
		})
		return
	}
	actionType, err := strconv.ParseInt(c.PostForm("action_type"), 10, 64)
	if err != nil || (actionType != 1 && actionType != 2) {
		c.JSON(http.StatusOK, response.RelationAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "action_type 不合法",
			},
		})
		return
	}
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
	req := &relationGrpc.RelationActionRequest{
		TokenUserId: userId.(uint64),
		ToUserId:    tid,
		ActionType:  relationGrpc.ActionType_FOLLOW,
	}
	res, err := rpc.RelationAction(c, req)
	if res == nil {
		c.JSON(http.StatusOK, response.Login{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  fmt.Sprintf("server request error:%v\n", err),
			},
		})
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.FollowList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.RelationAction{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
	})
}

func FriendList(c *gin.Context) {
}

func FollowList(c *gin.Context) {

}

func FollowerList(c *gin.Context) {

}
