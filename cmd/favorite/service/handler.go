package service

import (
	"context"
	"encoding/json"
	"time"
	redisDAO "wgxDouYin/cmd/favorite/redisDAO"
	"wgxDouYin/dal/db"
	"wgxDouYin/grpc/favorite"
)

type FavoriteServiceImpl struct {
	favorite.UnimplementedFavoriteServiceServer
}

const limit = 30

func (s *FavoriteServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (*favorite.FavoriteActionResponse, error) {
	if req.TokenUserId == 0 || req.VideoId <= 0 {
		logger.Errorf("操作非法")
		res := &favorite.FavoriteActionResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法",
		}
		return res, nil
	}
	videoActionRecord := db.FavoriteVideoRelation{VideoID: req.VideoId, UserID: req.TokenUserId, ActionType: req.ActionType}
	switch req.ActionType {
	case favorite.VideoActionType_LIKE:
		if err := db.CreateVideoRelation(ctx, &videoActionRecord); err != nil {
			res := &favorite.FavoriteActionResponse{
				StatusCode: -1,
				StatusMsg:  err.Error(),
			}
			return res, nil
		}
	case favorite.VideoActionType_DISLIKE:
		if err := db.CreateVideoRelation(ctx, &videoActionRecord); err != nil {
			res := &favorite.FavoriteActionResponse{
				StatusCode: -1,
				StatusMsg:  err.Error(),
			}
			return res, nil
		}
	case favorite.VideoActionType_CANCEL_LIKE:
		if err := db.DelVideoRelation(ctx, &videoActionRecord); err != nil {
			res := &favorite.FavoriteActionResponse{
				StatusCode: -1,
				StatusMsg:  err.Error(),
			}
			return res, nil
		}
	case favorite.VideoActionType_CANCEL_DISLIKE:
		if err := db.DelVideoRelation(ctx, &videoActionRecord); err != nil {
			res := &favorite.FavoriteActionResponse{
				StatusCode: -1,
				StatusMsg:  err.Error(),
			}
			return res, nil
		}
	case favorite.VideoActionType_WRONG_TYPE:
		res := &favorite.FavoriteActionResponse{
			StatusCode: -1,
			StatusMsg:  "错误的用户操作类型",
		}
		return res, nil
	}
	res := &favorite.FavoriteActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	if time.Now().Sub(videoActionRecord.CreatedAt) >= 24*time.Hour {
		return res, nil
	}
	var latestVideoActionTime time.Time
	if videoActionRecord.CreatedAt.Before(videoActionRecord.UpdatedAt) {
		latestVideoActionTime = videoActionRecord.UpdatedAt
	} else {
		latestVideoActionTime = videoActionRecord.CreatedAt
	}
	favoriteCache := &redisDAO.FavoriteCache{
		VideoID:    req.VideoId,
		UserID:     req.TokenUserId,
		CreatedAt:  latestVideoActionTime,
		ActionType: req.ActionType,
	}
	favoriteCacheByte, err := json.Marshal(favoriteCache)
	if err != nil {
		logger.Errorln(err)
		return res, nil
	}
	if err := FavoriteMQ.PublishSimple(ctx, favoriteCacheByte); err != nil {
		logger.Errorln(err)
		return res, nil
	}
	return res, nil
}

func (s *FavoriteServiceImpl) FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (resp *favorite.FavoriteListResponse, err error) {
	return nil, nil
}
