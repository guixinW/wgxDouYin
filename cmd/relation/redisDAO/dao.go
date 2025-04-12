package redisDAO

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
	wgxRedis "wgxDouYin/dal/redis"
	"wgxDouYin/grpc/relation"
	"wgxDouYin/internal/tool"
)

type RelationCache struct {
	UserID     uint64                      `json:"user_id" redis:"user_id"`
	ToUserID   uint64                      `json:"to_user_id" redis:"to_user_id"`
	CreatedAt  time.Time                   `json:"created_at" redis:"created_at"`
	ActionType relation.RelationActionType `json:"action_type" redis:"action_type"`
}

const followerRankName = "follower_rank"

func UpdateRelation(ctx context.Context, relationCache *RelationCache) error {
	keyRelation := fmt.Sprintf("user::%d::to_user::%d", relationCache.UserID, relationCache.ToUserID)
	valueRelation := fmt.Sprintf("%d::%d", relationCache.CreatedAt.UnixMilli(), relationCache.ActionType)
	follower := fmt.Sprintf("follower::%d", relationCache.ToUserID)
	following := fmt.Sprintf("following::%d", relationCache.UserID)
	expireTime := relationCache.CreatedAt.Add(wgxRedis.ExpireTime)

	addAction := func() error {
		err := wgxRedis.AddValueToKeySet(ctx, follower, []string{strconv.Itoa(int(relationCache.UserID))}, wgxRedis.RelationMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set follower error")
		}
		err = wgxRedis.AddValueToKeySet(ctx, following, []string{strconv.Itoa(int(relationCache.ToUserID))}, wgxRedis.RelationMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set following error")
		}
		err = wgxRedis.IncrNumInZSet(ctx, followerRankName, fmt.Sprintf("%v", relationCache.ToUserID), 1, wgxRedis.RelationMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set follower error")
		}
		return nil
	}
	deleteAction := func() error {
		err := wgxRedis.DelValueFormKeySet(ctx, follower, []string{strconv.Itoa(int(relationCache.UserID))}, wgxRedis.RelationMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set follower error")
		}
		err = wgxRedis.DelValueFormKeySet(ctx, following, []string{strconv.Itoa(int(relationCache.ToUserID))}, wgxRedis.RelationMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set following error")
		}
		err = wgxRedis.IncrNumInZSet(ctx, followerRankName, fmt.Sprintf("%v", relationCache.ToUserID), -1, wgxRedis.RelationMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set follower error")
		}
		return nil
	}

	keyExisted, err := wgxRedis.IsKeyExist(ctx, keyRelation)
	if err != nil {
		return wgxRedis.ErrorWrap(err, "UpdateRelation KeyExist error")
	}
	if !keyExisted {
		err := wgxRedis.SetKeyValue(ctx, keyRelation, valueRelation, expireTime, wgxRedis.RelationMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set read key error")
		}
		if relationCache.ActionType == relation.RelationActionType_FOLLOW {
			err = addAction()
			if err != nil {
				return err
			}
		}
	} else {
		existRelationValue, err := wgxRedis.GetKeyValue(ctx, keyRelation)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation get keyRelationRead error")
		}
		existRelationValueSplit := strings.Split(existRelationValue, "::")
		existRelationCreatedAt, err := wgxRedis.ParseMillisTimestamp(existRelationValueSplit[0])
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateFavorite relation redis time error")
		}
		existRelationActionType := tool.StrToRelationActionType(existRelationValueSplit[1])
		if existRelationActionType == relationCache.ActionType {
			return nil
		}

		//说明消息队列传入的消息为最新且action_type改变，需要根据action_type更新
		if relationCache.CreatedAt.After(existRelationCreatedAt) {
			err := wgxRedis.SetKeyValue(ctx, keyRelation, valueRelation, expireTime, wgxRedis.RelationMutex)
			if err != nil {
				return wgxRedis.ErrorWrap(err, "UpdateRelation")
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
	result, err := wgxRedis.GetSet(ctx, key)
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
	count, err := wgxRedis.GetSetCount(ctx, key)
	if err != nil {
		return 0, errors.Wrap(err, "GetFollowerCount error")
	}
	return count, nil
}

// GetFollowingIDs 根据userID获取关注者ID列表
func GetFollowingIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	key := fmt.Sprintf("following::%d", userID)
	result, err := wgxRedis.GetSet(ctx, key)
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
	count, err := wgxRedis.GetSetCount(ctx, key)
	if err != nil {
		return 0, wgxRedis.ErrorWrap(err, "get following count error")
	}
	return count, nil
}

// GetFriends 获取好友id
func GetFriends(ctx context.Context, userID uint64) ([]uint64, error) {
	followingKey := fmt.Sprintf("following::%d", userID)
	followerKey := fmt.Sprintf("follower::%d", userID)
	result, err := wgxRedis.GetSetIntersection(ctx, followingKey, followerKey)
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
	expireRelationSub := wgxRedis.GetRedisHelper().PSubscribe(ctx, "__keyevent@0__:expired")
	deleteExpireKeyInRelatedStruct := func(key, value string, isDeleteFans bool) error {
		isExist, err := wgxRedis.IsValueExistInKeySet(ctx, key, value)
		if err != nil {
			return err
		}
		if isExist {
			err := wgxRedis.DelValueFormKeySet(ctx, key, []string{value}, wgxRedis.RelationMutex)
			if err != nil {
				return err
			}
			if isDeleteFans {
				err := wgxRedis.IncrNumInZSet(ctx, followerRankName, value, -1, wgxRedis.RelationMutex)
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
		key := strings.Split(msg.Payload, "::")
		if key[0] != "user" {
			continue
		}
		fmt.Printf("expire msg %v\n", msg)
		toUserId := strings.Split(msg.Payload, "::")[3]
		userId := strings.Split(msg.Payload, "::")[1]
		followerSet := fmt.Sprintf("follower::%v", toUserId)
		followingSet := fmt.Sprintf("following::%v", userId)
		//删除toUserId followerSet中的粉丝userId
		err = deleteExpireKeyInRelatedStruct(followerSet, userId, false)
		if err != nil {
			logger.Errorln("Error deleting expired follower: %v", err)
		}
		//删除userId followingSet中的关注用户toUserId，并将toUserId热点用户榜的计数减1
		err = deleteExpireKeyInRelatedStruct(followingSet, toUserId, true)
		if err != nil {
			logger.Errorln("Error deleting expired following: %v", err)
		}
	}
}
