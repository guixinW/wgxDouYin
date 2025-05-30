package relatioDAO

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
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
	UpdatedAt  time.Time                   `json:"updated_at" redis:"updated_at"`
	CreatedAt  time.Time                   `json:"created_at" redis:"created_at"`
	ActionType relation.RelationActionType `json:"action_type" redis:"action_type"`
}

const followerRankName = "follower_rank"

func IncrAndCheckTopRank(ctx context.Context, key string, member string, incr float64, topN int64) (bool, error) {
	luaScript := redis.NewScript(`
		local score = redis.call("ZINCRBY", KEYS[1], ARGV[2], ARGV[1])
		local rank = redis.call("ZREVRANK", KEYS[1], ARGV[1])
		if rank ~= false and rank <= tonumber(ARGV[3]) then
			return 1
		else
			return 0
		end
	`)
	result, err := luaScript.Run(ctx, wgxRedis.GetRedisHelper(), []string{key}, member, incr, topN-1).Result()
	if err != nil {
		return false, err
	}
	inTopN, ok := result.(int64)
	if !ok {
		return false, nil
	}
	return inTopN == 1, nil
}

func handleFollow(ctx context.Context, cache *RelationCache) (bool, error) {
	userIDStr := strconv.Itoa(int(cache.UserID))
	toUserIDStr := strconv.Itoa(int(cache.ToUserID))

	if err := wgxRedis.AddValueToKeySet(ctx, fmt.Sprintf("follower::%d", cache.ToUserID), []string{userIDStr}); err != nil {
		return false, wgxRedis.ErrorWrap(err, "add follower error")
	}
	if err := wgxRedis.AddValueToKeySet(ctx, fmt.Sprintf("following::%d", cache.UserID), []string{toUserIDStr}); err != nil {
		return false, wgxRedis.ErrorWrap(err, "add following error")
	}
	inRank, err := IncrAndCheckTopRank(ctx, followerRankName, fmt.Sprintf("%d", cache.ToUserID), 1, 5)
	return inRank, err
}

func handleUnfollow(ctx context.Context, cache *RelationCache) (bool, error) {
	userIDStr := strconv.Itoa(int(cache.UserID))
	toUserIDStr := strconv.Itoa(int(cache.ToUserID))
	if err := wgxRedis.DelValueFormKeySet(ctx, fmt.Sprintf("follower::%d", cache.ToUserID), []string{userIDStr}, wgxRedis.RelationMutex); err != nil {
		return false, wgxRedis.ErrorWrap(err, "del follower error")
	}
	if err := wgxRedis.DelValueFormKeySet(ctx, fmt.Sprintf("following::%d", cache.UserID), []string{toUserIDStr}, wgxRedis.RelationMutex); err != nil {
		return false, wgxRedis.ErrorWrap(err, "del following error")
	}
	inRank, err := IncrAndCheckTopRank(ctx, followerRankName, fmt.Sprintf("%d", cache.ToUserID), -1, 5)
	return inRank, err
}

func UpdateRelation(ctx context.Context, relationCache *RelationCache) (bool, error) {
	keyRelation := fmt.Sprintf("user::%d::to_user::%d", relationCache.UserID, relationCache.ToUserID)
	valueRelation := fmt.Sprintf("%d::%d", relationCache.UpdatedAt.UnixMilli(), relationCache.ActionType)
	expireTime := relationCache.CreatedAt.Add(wgxRedis.KeyExpireTime)
	keyExisted, err := wgxRedis.IsKeyExist(ctx, keyRelation)
	if err != nil {
		return false, wgxRedis.ErrorWrap(err, "UpdateRelation KeyExist error")
	}
	//如果keyRelation不存在，则直接添加
	if !keyExisted {
		err := wgxRedis.SetKeyValue(ctx, keyRelation, valueRelation, expireTime)
		if err != nil {
			return false, wgxRedis.ErrorWrap(err, "UpdateRelation set read key error")
		}
		if relationCache.ActionType == relation.RelationActionType_FOLLOW {
			return handleFollow(ctx, relationCache)
		}
	}

	//如果存在，则与现有的keyRelation比较，查看是否需要更新
	existRelationValue, err := wgxRedis.GetKeyValue(ctx, keyRelation)
	if err != nil {
		return false, wgxRedis.ErrorWrap(err, "UpdateRelation get keyRelationRead error")
	}
	existRelationValueSplit := strings.Split(existRelationValue, "::")
	existRelationUpdatedAt, err := wgxRedis.ParseMillisTimestamp(existRelationValueSplit[0])
	if err != nil {
		return false, wgxRedis.ErrorWrap(err, "UpdateFavorite relation redis time error")
	}
	existRelationActionType := tool.StrToRelationActionType(existRelationValueSplit[1])
	if existRelationActionType == relationCache.ActionType {
		return false, nil
	}

	//说明消息队列传入的消息为最新且action_type改变，需要根据action_type更新
	if relationCache.UpdatedAt.After(existRelationUpdatedAt) {
		err := wgxRedis.SetKeyValue(ctx, keyRelation, valueRelation, expireTime)
		if err != nil {
			return false, wgxRedis.ErrorWrap(err, "UpdateRelation")
		}
		//当前relation为UN_FOLLOW，需要改变为FOLLOW
		if existRelationActionType == relation.RelationActionType_UN_FOLLOW {
			return handleFollow(ctx, relationCache)
		}
		//当前relation为FOLLOW，需要改变为UN_FOLLOW
		if existRelationActionType == relation.RelationActionType_FOLLOW {
			return handleUnfollow(ctx, relationCache)
		}
	}
	return false, nil
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
