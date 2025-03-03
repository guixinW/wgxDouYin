package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"wgxDouYin/dal/db"
	wgxRedis "wgxDouYin/dal/redis"
	relation "wgxDouYin/grpc/relation"
	"wgxDouYin/grpc/user"
	rabbitmq "wgxDouYin/pkg/rabbitMQ"
)

type RelationServiceImpl struct {
	relation.UnimplementedRelationServiceServer
}

// RelationAction 关注以及取关操作
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

// RelationFollowList 获取用户的关注列表
func (s *RelationServiceImpl) RelationFollowList(ctx context.Context, req *relation.RelationFollowListRequest) (*relation.RelationFollowListResponse, error) {
	if req.TokenUserId != req.UserId {
		logger.Errorf("操作非法：用户%v尝试获取用户%v的关注类表", req.TokenUserId, req.UserId)
		res := &relation.RelationFollowListResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：你无法获取其他用户的关注列表",
		}
		return res, nil
	}
	fmt.Printf("user id:%v\n", req.UserId)
	userIds, err := wgxRedis.GetFollowingIDs(ctx, req.UserId)
	if err != nil {
		logger.Errorf("用户%v获取关注者列表失败", req.UserId)
		res := &relation.RelationFollowListResponse{
			StatusCode: -1,
			StatusMsg:  "获取列表失败",
		}
		return res, nil
	}
	fmt.Println(userIds)
	userList := make([]*user.User, 0, len(userIds))
	for _, userId := range userIds {
		userFollowerCount, followerErr := wgxRedis.GetFollowerCount(ctx, userId)
		userFollowingCount, followingErr := wgxRedis.GetFollowingCount(ctx, userId)
		if followingErr != nil || followerErr != nil {
			res := &relation.RelationFollowListResponse{
				StatusCode: -1,
				StatusMsg:  "获取用户详细信息失败",
			}
			return res, nil
		}
		userList = append(userList, &user.User{
			Id:             userId,
			FollowingCount: userFollowingCount,
			FollowerCount:  userFollowerCount,
		})
	}
	res := &relation.RelationFollowListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserList:   userList,
	}
	return res, nil
}

// RelationFollowerList 获取用户的粉丝列表
func (s *RelationServiceImpl) RelationFollowerList(context.Context, *relation.RelationFollowerListRequest) (*relation.RelationFollowerListResponse, error) {
	return nil, nil
}

// RelationFriendList 获取用户的好友
func (s *RelationServiceImpl) RelationFriendList(context.Context, *relation.RelationFriendListRequest) (*relation.RelationFriendListResponse, error) {
	return nil, nil
}
