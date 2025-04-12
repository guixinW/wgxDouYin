package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"wgxDouYin/cmd/api/rpc"
	commentGrpc "wgxDouYin/grpc/comment"
	"wgxDouYin/internal/response"
	"wgxDouYin/internal/tool"
)

func CommentAction(c *gin.Context) {
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
	postActionType := c.PostForm("action_type")
	actionType := tool.StrToCommentActionType(c.PostForm("action_type"))
	if actionType == commentGrpc.CommentActionType_WRONG_TYPE {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("action_type %v 不合法", postActionType).Error()))
		return
	}

	req := new(commentGrpc.CommentActionRequest)
	req.VideoId = videoId
	req.TokenUserId = tokenUserId.(uint64)
	req.ActionType = actionType

	switch req.ActionType {
	case commentGrpc.CommentActionType_COMMENT:
		req.CommentText = c.PostForm("comment_text")
		if req.CommentText == "" {
			c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("comment text不能为空").Error()))
			return
		}
	case commentGrpc.CommentActionType_DELETE_COMMENT:
		req.CommentId, err = strconv.ParseUint(c.PostForm("comment_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("comment_id %v 错误", req.CommentId).Error()))
			return
		}
	}
	res, err := rpc.CommentAction(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求错误%v", err).Error()))
		return
	}
	c.JSON(http.StatusOK, response.Base{
		StatusCode: 0,
		StatusMsg:  res.StatusMsg,
	})
}

func CommentList(c *gin.Context) {
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	videoId, err := strconv.ParseUint(c.Query("video_id"), 10, 64)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("解析video_id失败").Error()))
		return
	}

	req := new(commentGrpc.CommentListRequest)
	req.VideoId = videoId
	req.TokenUserId = tokenUserId.(uint64)
	res, err := rpc.CommentList(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求错误%v", err).Error()))
		return
	}
	c.JSON(http.StatusOK, response.CommentList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		CommentList: res.CommentList,
	})
}
