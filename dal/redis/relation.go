package wgxRedis

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
	"wgxDouYin/grpc/relation"
)

func parseMillisTimestamp(ts string) (time.Time, error) {
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, errors.Errorf("时间戳解析失败: %v", err)
	}
	return time.UnixMilli(tsInt), nil
}

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
	CreatedAt  time.Time                   `json:"created_at" redis:"created_at"`
	ActionType relation.RelationActionType `json:"action_type" redis:"action_type"`
}

func ErrorWrap(err error, warpMessage string) error {
	return errors.Wrap(err, warpMessage)
}

func UpdateHotUser(ctx context.Context, userId uint64, isAdd bool) error {
	if isAdd == true {
		err := IncrNumInZSet(ctx, "follower_rank", fmt.Sprintf("%v", userId), 1, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateHotUser")
		}
	} else {
		err := IncrNumInZSet(ctx, "follower_rank", fmt.Sprintf("%v", userId), -1, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateHotUser")
		}
	}
	return nil
}

func UpdateRelation(ctx context.Context, relationCache *RelationCache) error {
	keyRelation := fmt.Sprintf("user::%d::to_user::%d", relationCache.UserID, relationCache.ToUserID)
	valueRelation := fmt.Sprintf("%d::%d", relationCache.CreatedAt.UnixMilli(), relationCache.ActionType)
	follower := fmt.Sprintf("follower::%d", relationCache.ToUserID)
	following := fmt.Sprintf("following::%d", relationCache.UserID)
	expireTime := relationCache.CreatedAt.Add(1 * time.Minute)

	addAction := func() error {
		err := addValueToKeySet(ctx, follower, []string{strconv.Itoa(int(relationCache.UserID))}, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set follower error")
		}
		err = addValueToKeySet(ctx, following, []string{strconv.Itoa(int(relationCache.ToUserID))}, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set following error")
		}
		err = UpdateHotUser(ctx, relationCache.ToUserID, true)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set follower error")
		}
		return nil
	}
	deleteAction := func() error {
		err := delValueFormKeySet(ctx, follower, []string{strconv.Itoa(int(relationCache.UserID))}, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set follower error")
		}
		err = delValueFormKeySet(ctx, following, []string{strconv.Itoa(int(relationCache.ToUserID))}, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set following error")
		}
		err = UpdateHotUser(ctx, relationCache.ToUserID, false)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set follower error")
		}
		return nil
	}

	keyExisted, err := isKeyExist(ctx, keyRelation)
	if err != nil {
		return ErrorWrap(err, "UpdateRelation KeyExist error")
	}
	if !keyExisted {
		err := setKeyValue(ctx, keyRelation, valueRelation, expireTime, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set read key error")
		}
		if relationCache.ActionType == relation.RelationActionType_FOLLOW {
			err = addAction()
			if err != nil {
				return err
			}
		}
	} else {
		existRelationValue, err := getKeyValue(ctx, keyRelation)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation get keyRelationRead error")
		}
		existRelationValueSplit := strings.Split(existRelationValue, "::")
		existRelationCreatedAt, err := parseMillisTimestamp(existRelationValueSplit[0])
		existRelationActionType := StrToRelationActionType(existRelationValueSplit[1])

		if existRelationActionType == relationCache.ActionType {
			return nil
		}
		//说明消息队列传入的消息为最新且action_type改变，需要根据action_type更新
		if relationCache.CreatedAt.After(existRelationCreatedAt) {
			err := setKeyValue(ctx, keyRelation, valueRelation, expireTime, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation")
			}
			if existRelationActionType == relation.RelationActionType_UN_FOLLOW {
				err = deleteAction()
				if err != nil {
					return err
				}
			} else if existRelationActionType == relation.RelationActionType_FOLLOW {
				err = addAction()
				if err != nil {
					return err
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

func ListenExpireRelation() {
	ctx := context.Background()
	expireRelationSub := GetRedisHelper().PSubscribe(ctx, "__keyevent@0__:expired")
	deleteExpireKeyInRelatedStruct := func(key, value string, isDeleteFans bool) error {
		isExist, err := isValueExistInKeySet(ctx, key, value)
		if err != nil {
			return err
		}
		if isExist {
			err := delValueFormKeySet(ctx, key, []string{value}, RelationMutex)
			if err != nil {
				return err
			}
			if isDeleteFans {
				err := IncrNumInZSet(ctx, "follower_rank", value, -1, RelationMutex)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	for {
		msg, err := expireRelationSub.ReceiveMessage(ctx)
		if err != nil {
			logger.Errorln("Error receiving message: %v", err)
		}
		toUserId := strings.Split(msg.Payload, "::")[3]
		userId := strings.Split(msg.Payload, "::")[1]
		followerSet := fmt.Sprintf("follower::%v", toUserId)
		followingSet := fmt.Sprintf("following::%v", userId)
		err = deleteExpireKeyInRelatedStruct(followerSet, userId, false)
		if err != nil {
			logger.Errorln("Error deleting expired follower: %v", err)
		}
		err = deleteExpireKeyInRelatedStruct(followingSet, toUserId, true)
		if err != nil {
			logger.Errorln("Error deleting expired following: %v", err)
		}
	}
}
