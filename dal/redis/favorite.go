package wgxRedis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"wgxDouYin/grpc/favorite"
)

func StrToVideoActionType(str string) favorite.VideoActionType {
	if str == "0" {
		return favorite.VideoActionType_LIKE
	} else if str == "1" {
		return favorite.VideoActionType_DISLIKE
	} else if str == "2" {
		return favorite.VideoActionType_CANCEL_LIKE
	} else if str == "3" {
		return favorite.VideoActionType_CANCEL_DISLIKE
	}
	return favorite.VideoActionType_WRONG_TYPE
}

type FavoriteCache struct {
	VideoID    uint64                   `json:"video_id" redis:"video_id"`
	UserID     uint64                   `json:"user_id" redis:"user_id"`
	ActionType favorite.VideoActionType `json:"action_type" redis:"action_type"`
	CreatedAt  uint64                   `json:"created_at" redis:"created_at"`
}

func UpdateFavorite(ctx context.Context, favoriteCache *FavoriteCache) error {
	keyFavorite := fmt.Sprintf("video::%d::user::%d", favoriteCache.VideoID, favoriteCache.UserID)
	valueFavorite := fmt.Sprintf("%d::%d", favoriteCache.CreatedAt, favoriteCache.ActionType)
	keyExisted, err := isKeyExist(ctx, keyFavorite)
	if err != nil {
		return ErrorWrap(err, "UpdateFavorite KeyExist error")
	}
	if !keyExisted {
		err := setKey(ctx, keyFavorite, valueFavorite, 0, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation set read key error")
		}
	} else {
		value, err := getKeyValue(ctx, keyFavorite)
		if err != nil {
			return ErrorWrap(err, "UpdateRelation get keyRelationRead error")
		}
		valueSplit := strings.Split(value, "::")
		redisCreatedAtStr, redisActionTypeStr := valueSplit[0], valueSplit[1]
		if StrToVideoActionType(redisActionTypeStr) == favoriteCache.ActionType {
			return nil
		}
		redisCreatedAt, err := strconv.ParseUint(redisCreatedAtStr, 10, 64)
		if err != nil {
			return ErrorWrap(err, "UpdateFavorite ParseInt error")
		}
		if redisCreatedAt < favoriteCache.CreatedAt {
			err := setKey(ctx, keyFavorite, valueFavorite, 0, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation")
			}
		}
	}
	return nil
}
