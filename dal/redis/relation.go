package wgxRedis

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type RelationCache struct {
	UserID     uint `json:"user_id" redis:"user_id"`
	ToUserID   uint `json:"to_user_id" redis:"to_user_id"`
	ActionType uint `json:"action_type" redis:"action_type"`
	CreatedAt  uint `json:"created_at" redis:"created_at"`
}

func ErrorWrap(err error, warpMessage string) error {
	return errors.Wrap(err, warpMessage)
}

func UpdateRelation(ctx context.Context, relation *RelationCache) error {
	keyRelation := fmt.Sprintf("user::%d::to_user::%d", relation.UserID, relation.ToUserID)
	valueRelation := fmt.Sprintf("%d::%d", relation.CreatedAt, relation.ActionType)
	follower := fmt.Sprintf("follower::%d", relation.ToUserID)
	following := fmt.Sprintf("following::%d", relation.UserID)
	keyExisted, err := GetRedisHelper().Exists(ctx, keyRelation).Result()
	if err != nil {
		return ErrorWrap(err, "UpdateRelation KeyExist error")
	}
	fmt.Printf("keyExisted:%v\n", keyExisted)
	if keyExisted == 0 {
		fmt.Printf("add realtion to redis")
		err := setKey(ctx, keyRelation, valueRelation, 0, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set read key error")
		}
		if relation.ActionType == 0 {
			fmt.Printf("new relation add follower and following")
			err = addKeyToSet(ctx, follower, []string{strconv.Itoa(int(relation.UserID))}, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation set follower error")
			}
			err = addKeyToSet(ctx, following, []string{strconv.Itoa(int(relation.ToUserID))}, RelationMutex)
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
		if redisActionType == strconv.Itoa(int(relation.ActionType)) {
			return nil
		} else if strconv.Itoa(int(relation.CreatedAt)) > redisCreatedAt {
			//说明最新的relation操作为取消关注，则需从follower、following中删除;否则为关注，需要添加
			if redisActionType == "1" {
				err = delKeyFormSet(ctx, follower, []string{strconv.Itoa(int(relation.UserID))}, RelationMutex)
				if err != nil {
					return ErrorWrap(err, "UpdateRelation set follower error")
				}
				err = delKeyFormSet(ctx, following, []string{strconv.Itoa(int(relation.ToUserID))}, RelationMutex)
				if err != nil {
					return ErrorWrap(err, "UpdateRelation set following error")
				}
			} else {
				err = addKeyToSet(ctx, follower, []string{strconv.Itoa(int(relation.UserID))}, RelationMutex)
				if err != nil {
					return ErrorWrap(err, "UpdateRelation set follower error")
				}
				err = addKeyToSet(ctx, following, []string{strconv.Itoa(int(relation.ToUserID))}, RelationMutex)
				if err != nil {
					return ErrorWrap(err, "UpdateRelation set following error")
				}
			}
			err := setKey(ctx, keyRelation, valueRelation, 0, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation")
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

// GetFollowerIDs 根据userID获取其粉丝数量
func GetFollowerCount(ctx context.Context, userID uint64) (uint64, error) {
	key := fmt.Sprintf("follower::%d", userID)
	follower := 0
	if follower, err := getSetCount(ctx, key); follower < 0 || err != nil {
		return 0, ErrorWrap(err, "get follower count error")
	}
	return uint64(follower), nil
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

func GetFollowingCount(ctx context.Context, userID uint64) (uint64, error) {
	key := fmt.Sprintf("following::%d", userID)
	follower := 0
	if follower, err := getSetCount(ctx, key); follower < 0 || err != nil {
		return 0, ErrorWrap(err, "get following count error")
	}
	return uint64(follower), nil
}
