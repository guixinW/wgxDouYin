package service

import (
	"context"
	"fmt"
	"time"
	"wgxDouYin/dal/db"
	"wgxDouYin/grpc/user"
	"wgxDouYin/grpc/video"
	"wgxDouYin/pkg/minio"
	"wgxDouYin/pkg/viper"
)

type VideoServiceImpl struct {
	video.UnimplementedVideoServiceServer
}

const limit = 30

func (s *VideoServiceImpl) Feed(ctx context.Context, req *video.FeedRequest) (*video.FeedResponse, error) {
	nextTime := time.Now().UnixMilli()
	if req.TokenUserId == 0 {
		logger.Errorf("操作非法：无法鉴别用户身份")
		res := &video.FeedResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：请登陆",
		}
		return res, nil
	}
	videoRecords, err := db.MGetVideos(ctx, limit, &req.LatestTime)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.FeedResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：请登陆",
		}
		return res, nil
	}
	videoLists := make([]*video.Video, 0, len(videoRecords))
	for _, videoRecord := range videoRecords {
		author, err := db.GetUserByID(ctx, videoRecord.AuthorID)
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：获取作者信息失败",
			}
			return res, nil
		}
		relation, err := db.GetRelationByUserIDs(ctx, req.TokenUserId, uint64(author.ID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：获取用户关系失败",
			}
			return res, nil
		}
		favorite, err := db.GetFavoriteVideoRelationByUserVideoID(ctx, req.TokenUserId, uint64(videoRecord.ID))
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：视频获取失败",
			}
			return res, nil
		}
		playUrl, err := minio.GetFileTemporaryURL(minio.VideoBucketName, videoRecord.PlayUrl)
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：视频获取失败",
			}
			return res, nil
		}
		fmt.Printf("created_at:%v\n", uint64(videoRecord.CreatedAt.Unix()))
		videoLists = append(videoLists, &video.Video{
			Author: &user.User{
				Id:             uint64(author.ID),
				Name:           author.UserName,
				FollowingCount: author.FollowingCount,
				FollowerCount:  author.FollowerCount,
				IsFollow:       relation != nil,
				TotalFavorite:  author.FavoriteCount,
				WorkCount:      author.WorkCount,
				FavoriteCount:  author.FavoriteCount,
			},
			PlayUrl:       playUrl,
			FavoriteCount: videoRecord.FavoriteCount,
			IsFavorite:    favorite == nil,
			CreateAt:      uint64(videoRecord.CreatedAt.Unix()),
			ShareCount:    0,
			CommentCount:  0,
			Title:         videoRecord.Title,
		})
	}
	if len(videoLists) != 0 {
		nextTime = videoRecords[len(videoRecords)-1].UpdatedAt.UnixMilli()
	}
	res := &video.FeedResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		VideoList:  videoLists,
		NextTime:   nextTime,
	}
	return res, nil
}

func (s *VideoServiceImpl) PublishAction(ctx context.Context, req *video.PublishActionRequest) (resp *video.PublishActionResponse, err error) {
	if len(req.Title) == 0 || len(req.Title) > 32 {
		logger.Errorf("标题不能为空且不能超过32个字符：%d", len(req.Title))
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "标题不能为空且不能超过32个字符",
		}
		return res, nil
	}
	maxSize := viper.Init("video").Viper.GetInt("video.maxSizeLimit")
	size := len(req.Data) / (1024 * 1024)
	if size > maxSize {
		logger.Errorln("视频文件过大")
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  fmt.Sprintf("该视频文件大于%dMB，上传受限", maxSize),
		}
		return res, nil
	}
	createTimestamp := time.Now().UnixMilli()
	videoTitle := fmt.Sprintf("%d_%s_%d.mp4", req.TokenUserId, req.Title, createTimestamp)
	v := &db.Video{
		Title:    req.Title,
		PlayUrl:  videoTitle,
		AuthorID: req.TokenUserId,
	}
	err = db.CreateVideo(ctx, v)
	fmt.Printf("v id:%v\n", v.ID)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "视频发布失败，服务器内部错误",
		}
		return res, nil
	}
	go func() {
		err := VideoPublish(req.Data, videoTitle)
		if err != nil {
			e := db.DeleteVideoByID(ctx, v.ID, req.TokenUserId)
			if e != nil {
				logger.Errorf("视频记录删除失败：%s", err.Error())
			}
		}
	}()
	return nil, nil
}

func (s *VideoServiceImpl) PublishList(ctx context.Context, req *video.PublishListRequest) (resp *video.PublishListResponse, err error) {
	if req.TokenUserId != req.UserId {
		logger.Errorf("用户%v越权访问", req.UserId)
		res := &video.PublishListResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法",
		}
		return res, nil
	}
	videoRecords, err := db.GetVideosByUserID(ctx, req.UserId)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.PublishListResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：请登陆",
		}
		return res, nil
	}
	videoLists := make([]*video.Video, 0, len(videoRecords))
	for _, videoRecord := range videoRecords {
		author, err := db.GetUserByID(ctx, videoRecord.AuthorID)
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：获取作者信息失败",
			}
			return res, nil
		}
		relation, err := db.GetRelationByUserIDs(ctx, req.TokenUserId, uint64(author.ID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：获取用户关系失败",
			}
			return res, nil
		}
		favorite, err := db.GetFavoriteVideoRelationByUserVideoID(ctx, req.TokenUserId, uint64(videoRecord.ID))
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：视频获取失败",
			}
			return res, nil
		}
		playUrl, err := minio.GetFileTemporaryURL(minio.VideoBucketName, videoRecord.PlayUrl)
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：视频获取失败",
			}
			return res, nil
		}
		fmt.Printf("created_at:%v\n", uint64(videoRecord.CreatedAt.Unix()))
		videoLists = append(videoLists, &video.Video{
			Author: &user.User{
				Id:             uint64(author.ID),
				Name:           author.UserName,
				FollowingCount: author.FollowingCount,
				FollowerCount:  author.FollowerCount,
				IsFollow:       relation != nil,
				TotalFavorite:  author.FavoriteCount,
				WorkCount:      author.WorkCount,
				FavoriteCount:  author.FavoriteCount,
			},
			PlayUrl:       playUrl,
			FavoriteCount: videoRecord.FavoriteCount,
			IsFavorite:    favorite == nil,
			CreateAt:      uint64(videoRecord.CreatedAt.Unix()),
			ShareCount:    0,
			CommentCount:  0,
			Title:         videoRecord.Title,
		})
	}
	res := &video.PublishListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		VideoList:  videoLists,
	}
	return res, nil
}
