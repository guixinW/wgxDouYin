package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"wgxDouYin/cmd/api/rpc"
	favoriteGrpc "wgxDouYin/grpc/favorite"
	"wgxDouYin/internal/response"
	"wgxDouYin/internal/tool"
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
	postFormActionType := c.PostForm("video_action_type")
	actionType := tool.StrToVideoActionType(postFormActionType)
	if actionType == favoriteGrpc.VideoActionType_WRONG_TYPE {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("操作类型%v不合法", postFormActionType).Error()))
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
