package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	wgxRedis "wgxDouYin/dal/redis"
	"wgxDouYin/grpc/favorite"
	rabbitmq "wgxDouYin/pkg/rabbitMQ"
)

type FavoriteServiceImpl struct {
	favorite.UnimplementedFavoriteServiceServer
}

const limit = 30

func (s *FavoriteServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (*favorite.FavoriteActionResponse, error) {
	if req.TokenUserId == 0 {
		logger.Errorf("操作非法：无法鉴别用户身份")
		res := &favorite.FavoriteActionResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：请登陆",
		}
		return res, nil
	}
	favoriteMessage := &wgxRedis.FavoriteCache{
		VideoID:    req.VideoId,
		UserID:     req.TokenUserId,
		ActionType: req.ActionType,
		CreatedAt:  uint64(time.Now().UnixMilli()),
	}
	jsonRc, err := json.Marshal(favoriteMessage)
	if err != nil {
		logger.Errorln("序列化Relation失败")
		res := &favorite.FavoriteActionResponse{
			StatusCode: -1,
			StatusMsg:  "内部错误",
		}
		return res, nil
	}
	if err = RelationMQ.PublishSimple(ctx, jsonRc); err != nil {
		logger.Errorf("消息队列发布错误：%v", err.Error())
		if strings.Contains(err.Error(), "连接断开") {
			go func() {
				err := RelationMQ.Destroy()
				if err != nil {
					logger.Errorln(err)
				}
			}()
			RelationMQ, err = rabbitmq.DefaultRabbitMQInstance("favorite")
			if err != nil {
				logger.Errorf(err.Error())
			}
			go func() {
				err := consume()
				if err != nil {
					logger.Errorf(err.Error())
				}
			}()
			res := &favorite.FavoriteActionResponse{
				StatusCode: 0,
				StatusMsg:  "success",
			}
			return res, nil
		}
		res := &favorite.FavoriteActionResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：操作失败",
		}
		return res, nil
	}
	res := &favorite.FavoriteActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	return res, nil
}

func (s *FavoriteServiceImpl) FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (resp *favorite.FavoriteListResponse, err error) {
	return nil, nil
}
