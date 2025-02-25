package wgxRedis

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
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
	errLock := RelationMutex.LockContext(ctx)
	defer func(RelationMutex *redsync.Mutex, ctx context.Context) {
		_, err := RelationMutex.UnlockContext(ctx)
		if err != nil {
			zapLogger.Errorln(err)
		}
	}(RelationMutex, ctx)
	if errLock != nil {
		return ErrorWrap(errLock, "UpdateRelation")
	}
	keyRelationRead := fmt.Sprintf("user::%d::to_user::%d::r", relation.UserID, relation.ToUserID)
	keyRelationWrite := fmt.Sprintf("user::%d::to_user::%d::w", relation.UserID, relation.ToUserID)
	valueRedis := fmt.Sprintf("%d::%d", relation.CreatedAt, relation.ActionType)
	readExisted, err := GetRedisHelper().Exists(ctx, keyRelationWrite).Result()
	if err != nil {
		return ErrorWrap(err, "UpdateRelation")
	}
	if readExisted == 0 {
		err := setKey(ctx, keyRelationRead, valueRedis, ExpireTime, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation")
		}
		err = setKey(ctx, keyRelationWrite, valueRedis, 0, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation")
		}
	} else {
		res, err := GetRedisHelper().Get(ctx, keyRelationRead).Result()
		if err != nil {
			return ErrorWrap(err, "UpdateRelation")
		}
		valueSplit := strings.Split(res, "::")
		redisCreatedAt, redisActionType := valueSplit[0], valueSplit[1]
		if redisActionType == strconv.Itoa(int(relation.ActionType)) {
			return nil
		} else if strconv.Itoa(int(relation.CreatedAt)) > redisCreatedAt {
			err := setKey(ctx, keyRelationRead, valueRedis, ExpireTime, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation")
			}
			err = setKey(ctx, keyRelationWrite, valueRedis, ExpireTime, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation")
			}
		}
	}
	return nil
}

// GetFollowerIDs 根据userID获取其粉丝ID列表
func GetFollowerIDs(ctx context.Context, userID uint64) (*[]uint64, error) {
	return nil, nil
}

// GetFollowingIDs 根据userID获取关注者ID列表
func GetFollowingIDs(ctx context.Context, userID uint64) (*[]uint64, error) {
	return nil, nil
}
