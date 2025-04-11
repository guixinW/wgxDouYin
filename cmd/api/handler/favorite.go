package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"wgxDouYin/cmd/api/rpc"
	favoriteGrpc "wgxDouYin/grpc/favorite"
	"wgxDouYin/internal/response"
)

func FavoriteAction(c *gin.Context) {
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	videoId, err := strconv.ParseUint(c.PostForm("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("解析video_id失败").Error()))
		return
	}
	var actionType favoriteGrpc.VideoActionType
	postActionType := c.PostForm("video_action_type")
	if postActionType == "0" {
		actionType = favoriteGrpc.VideoActionType_LIKE
	} else if postActionType == "1" {
		actionType = favoriteGrpc.VideoActionType_DISLIKE
	} else if postActionType == "2" {
		actionType = favoriteGrpc.VideoActionType_CANCEL_LIKE
	} else if postActionType == "3" {
		actionType = favoriteGrpc.VideoActionType_CANCEL_DISLIKE
	} else {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("操作类型不合法").Error()))
		return
	}
	req := &favoriteGrpc.FavoriteActionRequest{
		TokenUserId: tokenUserId.(uint64),
		VideoId:     videoId,
		ActionType:  actionType,
	}
	res, err := rpc.FavoriteAction(c, req)
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.FavoriteAction{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
	})
}

func FavoriteList(c *gin.Context) {
	res, err := rpc.FavoriteList(c, &favoriteGrpc.FavoriteListRequest{})
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.FavoriteAction{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
	})
}
