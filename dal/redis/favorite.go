package wgxRedis

import (
	"context"
	"fmt"
	"strings"
	"time"
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
	CreatedAt  time.Time                `json:"created_at" redis:"created_at"`
	ActionType favorite.VideoActionType `json:"action_type" redis:"action_type"`
}

func UpdateFavorite(ctx context.Context, favoriteCache *FavoriteCache) error {
	keyFavorite := fmt.Sprintf("video::%d::user::%d", favoriteCache.VideoID, favoriteCache.UserID)
	valueFavorite := fmt.Sprintf("%d::%d", favoriteCache.CreatedAt.UnixMilli(), favoriteCache.ActionType)
	expireTime := favoriteCache.CreatedAt.Add(10 * time.Second)
	keyExisted, err := isKeyExist(ctx, keyFavorite)
	if err != nil {
		return ErrorWrap(err, "UpdateFavorite KeyExist error")
	}
	if !keyExisted {
		err := setKeyValue(ctx, keyFavorite, valueFavorite, expireTime, RelationMutex)
		if err != nil {
			return ErrorWrap(err, "UpdateFavorite set read key error")
		}
	} else {
		existFavoriteValue, err := getKeyValue(ctx, keyFavorite)
		if err != nil {
			return ErrorWrap(err, "UpdateFavorite get key error")
		}
		existFavoriteValueSplit := strings.Split(existFavoriteValue, "::")
		existFavoriteCreatedAt, err := time.Parse("2006-01-02 15:04:05.000", existFavoriteValueSplit[0])
		if err != nil {
			return ErrorWrap(err, "UpdateFavorite parse redis time error")
		}
		existFavoriteActionType := StrToVideoActionType(existFavoriteValueSplit[1])
		if existFavoriteActionType == favoriteCache.ActionType {
			return nil
		}
		if favoriteCache.CreatedAt.After(existFavoriteCreatedAt) {
			err := setKeyValue(ctx, keyFavorite, valueFavorite, expireTime, RelationMutex)
			if err != nil {
				return ErrorWrap(err, "UpdateRelation")
			}
		}
	}
	return nil
}
