package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"time"
	"wgxDouYin/cmd/api/rpc"
	videoGrpc "wgxDouYin/grpc/video"
	"wgxDouYin/internal/response"
	"wgxDouYin/pkg/minio"
)

func PublishAction(c *gin.Context) {
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("标题不能为空").Error()))
		return
	}
	playUrl := c.PostForm("PlayUrl")
	if playUrl == "" {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("PlayUrl不能为空").Error()))
		return
	}
	req := videoGrpc.PublishActionRequest{
		TokenUserId: tokenUserId.(uint64),
		PlayUrl:     playUrl,
		Title:       title,
	}
	res, err := rpc.PublishAction(c, &req)
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.PublishAction{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
	})
}

func Feed(c *gin.Context) {
	latestTime := c.Query("latest_time")
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	var timestamp int64 = 0
	if latestTime != "" {
		timestamp, _ = strconv.ParseInt(latestTime, 10, 64)
	} else {
		timestamp = time.Now().UnixMilli()
	}
	req := &videoGrpc.FeedRequest{
		LatestTime:  timestamp,
		TokenUserId: tokenUserId.(uint64),
	}
	res, err := rpc.Feed(c, req)
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.Feed{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		VideoList: res.VideoList,
	})
}

func PublishList(c *gin.Context) {
	userId, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
		return
	}
	tokenUserId, exist := c.Get("token_user_id")
	if !exist {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Errorf("请求不合法").Error()))
		return
	}
	req := &videoGrpc.PublishListRequest{
		UserId:      userId,
		TokenUserId: tokenUserId.(uint64),
	}
	res, err := rpc.PublishList(c, req)
	if res == nil || err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(fmt.Sprintf("服务端请求错误:%v\n", err)))
		return
	}
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.ErrorResponse(res.StatusMsg))
		return
	}
	c.JSON(http.StatusOK, response.PublishList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		VideoList: res.VideoList,
	})
}

func PublishPostURL(c *gin.Context) {
	title := c.Query("title")
	fmt.Println(title)
	if title == "" {
		c.JSON(http.StatusOK, response.ErrorResponse("名称非法"))
		return
	}
	videoBucketName := minio.VideoBucketName
	PlayUrl := fmt.Sprintf("%s_%s_.mp4", uuid.New().String(), title)
	UploadUrl, err := minio.GetUploadURL(videoBucketName, PlayUrl)
	if err != nil {
		c.JSON(http.StatusOK, response.ErrorResponse(err.Error()))
	}
	c.JSON(http.StatusOK, response.PublishPostURL{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  "ok",
		},
		UploadUrl: UploadUrl,
		PlayUrl:   PlayUrl,
	})
}
