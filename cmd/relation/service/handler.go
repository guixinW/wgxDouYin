package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wgxDouYin/cmd/relation/redisDAO"
	"wgxDouYin/dal/db"
	"wgxDouYin/grpc/relation"
	"wgxDouYin/grpc/user"
)

type RelationServiceImpl struct {
	relation.UnimplementedRelationServiceServer
}

// RelationAction 关注以及取关操作
func (s *RelationServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (resp *relation.RelationActionResponse, err error) {
	//token既然已经针对该ID进行签发，则tokenUserId一定存在，无需验证；只需验证被操作方是否存在
	//过程：
	//1.插入MySQL，发送消息到消息队列给Redis消费关系信息；
	//2.消费关系信息，然后利用redis查询被操作方是否为热点用户，如果是更新其热点用户排行榜的信息
	//3.如果不是，则更新其Redis中的最近24小时关注量，再判断该关注量是否超过热点用户阈值，如果超过则加入热点用户排行榜
	//4.定时任务对照补偿，修正因为发送队列失败造成的部分数据错误，查询热点用户是否有遗漏
	if req.TokenUserId == req.ToUserId {
		logger.Errorf("操作非法：用户无法成为自己的粉丝：%d", req.TokenUserId)
		res := &relation.RelationActionResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：用户无法成为自己的粉丝",
		}
		return res, nil
	}
	relationActionRecord := db.FollowRelation{UserID: req.TokenUserId, ToUserID: req.ToUserId}
	switch req.ActionType {
	case relation.RelationActionType_FOLLOW:
		if err := db.CreateRelation(ctx, &relationActionRecord); err != nil {
			res := &relation.RelationActionResponse{
				StatusCode: -1,
				StatusMsg:  err.Error(),
			}
			return res, nil
		}
	case relation.RelationActionType_UN_FOLLOW:
		if err := db.DelRelationByUserID(ctx, &relationActionRecord); err != nil {
			res := &relation.RelationActionResponse{
				StatusCode: -1,
				StatusMsg:  err.Error(),
			}
			return res, nil
		}
	case relation.RelationActionType_WRONG_TYPE:
		res := &relation.RelationActionResponse{
			StatusCode: -1,
			StatusMsg:  "错误的用户操作类型",
		}
		return res, nil
	}
	res := &relation.RelationActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	if time.Now().Sub(relationActionRecord.CreatedAt) >= 24*time.Hour {
		return res, nil
	}
	relationCache := &redisDAO.RelationCache{
		UserID:     req.TokenUserId,
		ToUserID:   req.ToUserId,
		ActionType: req.ActionType,
		CreatedAt:  relationActionRecord.CreatedAt,
		UpdatedAt:  relationActionRecord.UpdatedAt,
	}
	relationCacheByte, err := json.Marshal(relationCache)
	if err != nil {
		logger.Errorln(err)
		return res, nil
	}
	if err := RelationMQ.PublishSimple(ctx, relationCacheByte); err != nil {
		return res, nil
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
	userIds, err := redisDAO.GetFollowingIDs(ctx, req.UserId)
	if err != nil {
		logger.Errorf("用户%v获取关注者列表失败", req.UserId)
		res := &relation.RelationFollowListResponse{
			StatusCode: -1,
			StatusMsg:  "获取列表失败",
		}
		return res, nil
	}
	userList := make([]*user.User, 0, len(userIds))
	for _, userId := range userIds {
		userFollowerCount, followerErr := redisDAO.GetFollowerCount(ctx, userId)
		userFollowingCount, followingErr := redisDAO.GetFollowingCount(ctx, userId)
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
func (s *RelationServiceImpl) RelationFollowerList(ctx context.Context, req *relation.RelationFollowerListRequest) (*relation.RelationFollowerListResponse, error) {
	if req.TokenUserId != req.UserId {
		logger.Errorf("操作非法：用户%v尝试获取用户%v的粉丝列表", req.TokenUserId, req.UserId)
		res := &relation.RelationFollowerListResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：你无法获取其他用户的粉丝列表",
		}
		return res, nil
	}
	userIds, err := redisDAO.GetFollowerIDs(ctx, req.UserId)
	if err != nil {
		logger.Errorf("用户%v获取关注者列表失败", req.UserId)
		res := &relation.RelationFollowerListResponse{
			StatusCode: -1,
			StatusMsg:  "获取列表失败",
		}
		return res, nil
	}
	userList := make([]*user.User, 0, len(userIds))
	for _, userId := range userIds {
		userFollowerCount, followerErr := redisDAO.GetFollowerCount(ctx, userId)
		userFollowingCount, followingErr := redisDAO.GetFollowingCount(ctx, userId)
		if followingErr != nil || followerErr != nil {
			res := &relation.RelationFollowerListResponse{
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
	res := &relation.RelationFollowerListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserList:   userList,
	}
	return res, nil
}

// RelationFriendList 获取用户的好友
func (s *RelationServiceImpl) RelationFriendList(ctx context.Context, req *relation.RelationFriendListRequest) (*relation.RelationFriendListResponse, error) {
	if req.TokenUserId != req.UserId {
		logger.Errorf("操作非法：用户%v尝试获取用户%v的好友列表", req.TokenUserId, req.UserId)
		res := &relation.RelationFriendListResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法：你无法获取其他用户的粉丝列表",
		}
		return res, nil
	}
	friendIds, err := redisDAO.GetFriends(ctx, req.UserId)
	if err != nil {
		logger.Errorf("用户%v获取关注者列表失败", req.UserId)
		res := &relation.RelationFriendListResponse{
			StatusCode: -1,
			StatusMsg:  "获取列表失败",
		}
		return res, nil
	}
	fmt.Println(friendIds)
	friendList := make([]*relation.FriendUser, 0, len(friendIds))
	for _, friendId := range friendIds {
		friendFollowerCount, followerErr := redisDAO.GetFollowerCount(ctx, friendId)
		friendFollowingCount, followingErr := redisDAO.GetFollowingCount(ctx, friendId)
		if followingErr != nil || followerErr != nil {
			res := &relation.RelationFriendListResponse{
				StatusCode: -1,
				StatusMsg:  "获取用户详细信息失败",
			}
			return res, nil
		}
		fmt.Println(friendId, friendFollowerCount, friendFollowerCount)
		friendList = append(friendList, &relation.FriendUser{
			User: &user.User{
				Id:             friendId,
				IsFollow:       true,
				FollowerCount:  friendFollowerCount,
				FollowingCount: friendFollowingCount,
			},
		})
	}
	res := &relation.RelationFriendListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserList:   friendList,
	}
	return res, nil
}
