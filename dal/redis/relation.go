package wgxRedis

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"wgxDouYin/grpc/relation"
)

func StrToRelationActionType(str string) relation.RelationActionType {
	if str == "0" {
		return relation.RelationActionType_FOLLOW
	} else if str == "1" {
		return relation.RelationActionType_UN_FOLLOW
	} else {
		return relation.RelationActionType_WRONG_TYPE
	}
}

type RelationCache struct {
	UserID     uint64                      `json:"user_id" redis:"user_id"`
	ToUserID   uint64                      `json:"to_user_id" redis:"to_user_id"`
	CreatedAt  uint64                      `json:"created_at" redis:"created_at"`
	ActionType relation.RelationActionType `json:"action_type" redis:"action_type"`
}

func ErrorWrap(err error, warpMessage string) error {
	return errors.Wrap(err, warpMessage)
}

func UpdateRelation(ctx context.Context, relationCache *RelationCache) error {
	keyRelation := fmt.Sprintf("user::%d::to_user::%d", relationCache.UserID, relationCache.ToUserID)
	valueRelation := fmt.Sprintf("%d::%d", relationCache.CreatedAt, relationCache.ActionType)
	follower := fmt.Sprintf("follower::%d", relationCache.ToUserID)
	following := fmt.Sprintf("following::%d", relationCache.UserID)
	keyExisted, err := isKeyExist(ctx, keyRelation)
	if err != nil {
		return ErrorWrap(err, "UpdateRelation KeyExist error")
	}
	if !keyExisted {
		err := setKey(ctx, keyRelation, valueRelation, 0, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set read key error")
		}
		if relationCache.ActionType == relation.RelationActionType_FOLLOW {
			err = addKeyToSet(ctx, follower, []string{strconv.Itoa(int(relationCache.UserID))}, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation set follower error")
			}
			err = addKeyToSet(ctx, following, []string{strconv.Itoa(int(relationCache.ToUserID))}, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation set following error")
			}
		}
	} else {
		value, err := getKeyValue(ctx, keyRelation)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation get keyRelationRead error")
		}
		valueSplit := strings.Split(value, "::")
		redisCreatedAt, redisActionType := valueSplit[0], valueSplit[1]
		if StrToRelationActionType(redisActionType) == relationCache.ActionType {
			return nil
		}
		//说明消息队列传入的消息为最新且action_type改变，需要根据action_type更新
		if strconv.Itoa(int(relationCache.CreatedAt)) > redisCreatedAt {
			err := setKey(ctx, keyRelation, valueRelation, 0, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation")
			}
			if StrToRelationActionType(redisActionType) == relation.RelationActionType_UN_FOLLOW {
				err = delKeyFormSet(ctx, follower, []string{strconv.Itoa(int(relationCache.UserID))}, RelationMutex)
				if err != nil {
					return ErrorWrap(err, "UpdateRelation set follower error")
				}
				err = delKeyFormSet(ctx, following, []string{strconv.Itoa(int(relationCache.ToUserID))}, RelationMutex)
				if err != nil {
					return ErrorWrap(err, "UpdateRelation set following error")
				}
			} else if StrToRelationActionType(redisActionType) == relation.RelationActionType_FOLLOW {
				err = addKeyToSet(ctx, follower, []string{strconv.Itoa(int(relationCache.UserID))}, RelationMutex)
				if err != nil {
					return ErrorWrap(err, "UpdateRelation set follower error")
				}
				err = addKeyToSet(ctx, following, []string{strconv.Itoa(int(relationCache.ToUserID))}, RelationMutex)
				if err != nil {
					return ErrorWrap(err, "UpdateRelation set following error")
				}
			}
		}
	}
	return nil
}

// GetFollowerIDs 根据userID获取其粉丝ID列表
func GetFollowerIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	key := fmt.Sprintf("follower::%d", userID)
	result, err := getSet(ctx, key)
	if err != nil {
		return nil, err
	}
	followers := make([]uint64, 0, len(result))
	for _, follower := range result {
		followerId, err := strconv.ParseUint(follower, 10, 64)
		if err != nil {
			return nil, err
		}
		followers = append(followers, followerId)
	}
	return followers, nil
}

// GetFollowerCount 根据userID获取其粉丝数量
func GetFollowerCount(ctx context.Context, userID uint64) (uint64, error) {
	key := fmt.Sprintf("follower::%d", userID)
	count, err := getSetCount(ctx, key)
	if err != nil {
		return 0, errors.Wrap(err, "GetFollowerCount error")
	}
	return count, nil
}

// GetFollowingIDs 根据userID获取关注者ID列表
func GetFollowingIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	key := fmt.Sprintf("following::%d", userID)
	result, err := getSet(ctx, key)
	if err != nil {
		return nil, err
	}
	following := make([]uint64, 0, len(result))
	for _, follower := range result {
		followingId, err := strconv.ParseUint(follower, 10, 64)
		if err != nil {
			return nil, err
		}
		following = append(following, followingId)
	}
	return following, nil
}

// GetFollowingCount 根据userID获取关注者数量
func GetFollowingCount(ctx context.Context, userID uint64) (uint64, error) {
	key := fmt.Sprintf("following::%d", userID)
	count, err := getSetCount(ctx, key)
	if err != nil {
		return 0, ErrorWrap(err, "get following count error")
	}
	return count, nil
}

// GetFriends 获取好友id
func GetFriends(ctx context.Context, userID uint64) ([]uint64, error) {
	followingKey := fmt.Sprintf("following::%d", userID)
	followerKey := fmt.Sprintf("follower::%d", userID)
	result, err := getSetIntersection(ctx, followingKey, followerKey)
	if err != nil {
		return nil, err
	}
	friendIdsInt := make([]uint64, 0, len(result))
	for _, friendIdStr := range result {
		friendId, err := strconv.ParseUint(friendIdStr, 10, 64)
		if err != nil {
			return nil, err
		}
		friendIdsInt = append(friendIdsInt, friendId)
	}
	return friendIdsInt, nil
}
