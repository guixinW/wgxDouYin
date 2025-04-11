package redisDAO

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	wgxRedis "wgxDouYin/dal/redis"
	"wgxDouYin/grpc/favorite"
)

type FavoriteCache struct {
	VideoID    uint64                   `json:"video_id" redis:"video_id"`
	UserID     uint64                   `json:"user_id" redis:"user_id"`
	CreatedAt  time.Time                `json:"created_at" redis:"created_at"`
	ActionType favorite.VideoActionType `json:"action_type" redis:"action_type"`
}

const videoRankName = "video_rank"

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

func UpdateFavorite(ctx context.Context, favoriteCache *FavoriteCache) error {
	keyFavorite := fmt.Sprintf("video::%d::user::%d", favoriteCache.VideoID, favoriteCache.UserID)
	valueFavorite := fmt.Sprintf("%d::%d", favoriteCache.CreatedAt.UnixMilli(), favoriteCache.ActionType)
	videoLikeSet := fmt.Sprintf("videoLike::%d", favoriteCache.VideoID)
	expireTime := favoriteCache.CreatedAt.Add(wgxRedis.ExpireTime)

	addAction := func() error {
		err := wgxRedis.IncrNumInZSet(ctx, videoRankName, fmt.Sprintf("%v", favoriteCache.VideoID), 1, wgxRedis.FavoriteMutex)
		if err != nil {
			return err
		}
		err = wgxRedis.AddValueToKeySet(ctx, videoLikeSet, []string{strconv.Itoa(int(favoriteCache.UserID))}, wgxRedis.FavoriteMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set following error")
		}
		return nil
	}

	delAction := func() error {
		err := wgxRedis.IncrNumInZSet(ctx, videoRankName, fmt.Sprintf("%v", favoriteCache.VideoID), -1, wgxRedis.FavoriteMutex)
		if err != nil {
			return err
		}
		err = wgxRedis.DelValueFormKeySet(ctx, videoLikeSet, []string{strconv.Itoa(int(favoriteCache.UserID))}, wgxRedis.FavoriteMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateRelation set follower error")
		}
		return nil
	}

	keyExisted, err := wgxRedis.IsKeyExist(ctx, keyFavorite)
	if err != nil {
		return wgxRedis.ErrorWrap(err, "UpdateFavorite KeyExist error")
	}
	if !keyExisted {
		err := wgxRedis.SetKeyValue(ctx, keyFavorite, valueFavorite, expireTime, wgxRedis.FavoriteMutex)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateFavorite set read key error")
		}
		if favoriteCache.ActionType == favorite.VideoActionType_LIKE {
			err = addAction()
			if err != nil {
				return err
			}
		}
	} else {
		existFavoriteValue, err := wgxRedis.GetKeyValue(ctx, keyFavorite)
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateFavorite get key error")
		}
		existFavoriteValueSplit := strings.Split(existFavoriteValue, "::")
		existFavoriteActionType := StrToVideoActionType(existFavoriteValueSplit[1])
		fmt.Println(existFavoriteActionType, favoriteCache.ActionType)
		if existFavoriteActionType == favoriteCache.ActionType {
			return nil
		}
		existFavoriteCreatedAt, err := wgxRedis.ParseMillisTimestamp(existFavoriteValueSplit[0])
		if err != nil {
			return wgxRedis.ErrorWrap(err, "UpdateFavorite parse redis time error")
		}
		if favoriteCache.CreatedAt.After(existFavoriteCreatedAt) {
			fmt.Println("update")
			err := wgxRedis.SetKeyValue(ctx, keyFavorite, valueFavorite, expireTime, wgxRedis.FavoriteMutex)
			if err != nil {
				return wgxRedis.ErrorWrap(err, "UpdateRelation")
			}
			if existFavoriteActionType == favorite.VideoActionType_LIKE {
				if favoriteCache.ActionType == favorite.VideoActionType_DISLIKE || favoriteCache.ActionType == favorite.VideoActionType_CANCEL_LIKE {
					err = delAction()
					if err != nil {
						return err
					}
				}
			}
			if favoriteCache.ActionType == favorite.VideoActionType_LIKE {
				if existFavoriteActionType == favorite.VideoActionType_DISLIKE || existFavoriteActionType == favorite.VideoActionType_CANCEL_LIKE {
					err = addAction()
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func ListenExpireFavorite() {
	ctx := context.Background()
	expireRelationSub := wgxRedis.GetRedisHelper().PSubscribe(ctx, "__keyevent@0__:expired")
	for {
		msg, err := expireRelationSub.ReceiveMessage(ctx)
		if err != nil {
			logger.Errorln("Error receiving message: %v", err)
		}
		fmt.Printf("expire msg %v\n", msg)
		key := strings.Split(msg.Payload, "::")
		if key[0] != "video" {
			continue
		}
		videoId := strings.Split(msg.Payload, "::")[1]
		userId := strings.Split(msg.Payload, "::")[3]
		videoLikeSet := fmt.Sprintf("videoLike::%v", videoId)

		//删除videoLike Set中的点赞用户userId，删除videoId对应的videoLike rank减1
		isExist, err := wgxRedis.IsValueExistInKeySet(ctx, videoLikeSet, userId)
		fmt.Println(isExist, err)
		if err != nil {
			logger.Errorf("userId %v is not in video set", userId)
		}
		if isExist {
			err := wgxRedis.DelValueFormKeySet(ctx, videoLikeSet, []string{userId}, wgxRedis.FavoriteMutex)
			if err != nil {
				logger.Errorf("delete userId %v on videoLikeSet failed", userId)
			}
			err = wgxRedis.IncrNumInZSet(ctx, videoRankName, videoId, -1, wgxRedis.FavoriteMutex)
			if err != nil {
				logger.Errorf("increase %v in videoLikeSet failed", videoId)
			}
		}
	}
}
