package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"wgxDouYin/dal/db"
	wgxRedis "wgxDouYin/dal/redis"
	relation "wgxDouYin/grpc/relation"
	rabbitmq "wgxDouYin/pkg/rabbitMQ"
)

type RelationServiceImpl struct {
	relation.UnimplementedRelationServiceServer
}

func (s *RelationServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (resp *relation.RelationActionResponse, err error) {
	if req.TokenUserId == req.ToUserId {
		logger.Errorf("操作非法：用户无法成为自己的粉丝：%d", req.TokenUserId)
		res := &relation.RelationActionResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：用户无法成为自己的粉丝",
		}
		return res, nil
	}
	u1, _ := db.GetUserByID(ctx, req.TokenUserId)
	u2, _ := db.GetUserByID(ctx, req.ToUserId)
	if u1 == nil || u2 == nil {
		logger.Errorln("所请求的用户ID不存在")
		res := &relation.RelationActionResponse{
			StatusCode: -1,
			StatusMsg:  "所请求的用户ID不存在",
		}
		return res, nil
	}
	relationCache := &wgxRedis.RelationCache{
		UserID:     uint(req.TokenUserId),
		ToUserID:   uint(req.ToUserId),
		ActionType: uint(req.ActionType),
		CreatedAt:  uint(time.Now().UnixMilli()),
	}
	jsonRc, err := json.Marshal(relationCache)
	if err != nil {
		logger.Errorln("序列化Relation失败")
		res := &relation.RelationActionResponse{
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
			RelationMQ, err = rabbitmq.DefaultRabbitMQInstance("relation")
			if err != nil {
				logger.Errorf(err.Error())
			}
			go func() {
				err := consume()
				if err != nil {
					logger.Errorf(err.Error())
				}
			}()
			res := &relation.RelationActionResponse{
				StatusCode: 0,
				StatusMsg:  "success",
			}
			return res, nil
		}
		res := &relation.RelationActionResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：操作失败",
		}
		return res, nil
	}
	res := &relation.RelationActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	return res, nil
}

func (s *RelationServiceImpl) RelationFollowList(context.Context, *relation.RelationFollowListRequest) (*relation.RelationFollowListResponse, error) {
	return nil, nil
}

func (s *RelationServiceImpl) RelationFollowerList(context.Context, *relation.RelationFollowerListRequest) (*relation.RelationFollowerListResponse, error) {
	return nil, nil
}

func (s *RelationServiceImpl) RelationFriendList(context.Context, *relation.RelationFriendListRequest) (*relation.RelationFriendListResponse, error) {
	return nil, nil
}
