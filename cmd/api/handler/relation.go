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
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	var actionType relationGrpc.RelationActionType
	postActionType := c.PostForm("relation_action_type")
	if postActionType == "0" {
		actionType = relationGrpc.RelationActionType_FOLLOW
	} else if postActionType == "1" {
		actionType = relationGrpc.RelationActionType_UN_FOLLOW
	} else {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("action_type不合法").Error()))
		return
	}
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	req := &relationGrpc.RelationActionRequest{
		TokenUserId: tokenUserId.(uint64),
		ToUserId:    tid,
		ActionType:  actionType,
	}
	res, err := rpc.RelationAction(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
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
	userId, err := strconv.ParseUint(c.PostForm("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	req := &relationGrpc.RelationFriendListRequest{
		TokenUserId: tokenUserId.(uint64),
		UserId:      userId,
	}
	res, err := rpc.RelationFriendList(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.FriendList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		FriendList: res.UserList,
	})
}

func FollowingList(c *gin.Context) {
	userId, err := strconv.ParseUint(c.PostForm("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	req := &relationGrpc.RelationFollowListRequest{
		TokenUserId: tokenUserId.(uint64),
		UserId:      userId,
	}
	res, err := rpc.RelationFollowList(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.FollowList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserList: res.UserList,
	})
}

func FollowerList(c *gin.Context) {
	userId, err := strconv.ParseUint(c.PostForm("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	req := &relationGrpc.RelationFollowerListRequest{
		TokenUserId: tokenUserId.(uint64),
		UserId:      userId,
	}
	res, err := rpc.RelationFollowerList(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.FollowerList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserList: res.UserList,
	})
}
