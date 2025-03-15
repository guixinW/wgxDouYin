package wgxRedis

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"strconv"
	"strings"
	"time"
	"wgxDouYin/dal/db"
	"wgxDouYin/grpc/favorite"
	"wgxDouYin/grpc/relation"
	"wgxDouYin/pkg/zap"
)

func RelationMoveToDB() error {
	fmt.Println("sync relation to db")
	ctx := context.Background()
	keys, err := getKeys(ctx, "user::*::to_user::*")
	if err != nil {
		logger.Errorln(err)
		return err
	}
	for _, key := range keys {
		value, err := getKeyValue(ctx, key)
		if err != nil {
			logger.Errorln(err)
			return err
		}
		valueSplit := strings.Split(value, "::")
		keySplit := strings.Split(key, "::")
		userIdStr, toUserIDStr := keySplit[1], keySplit[3]
		_, actionTypeStr := valueSplit[0], valueSplit[1]
		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		toUserId, err := strconv.ParseInt(toUserIDStr, 10, 64)
		if err != nil {
			logger.Errorln(err)
			return err
		}
		relationRecord, err := db.GetRelationByUserIDs(ctx, uint64(userId), uint64(toUserId))
		if err != nil {
			logger.Errorln(err)
			return err
		} else if relationRecord == nil && StrToRelationActionType(actionTypeStr) == relation.RelationActionType_FOLLOW {
			err = db.CreateRelation(ctx, uint64(userId), uint64(toUserId))
			if err != nil {
				logger.Errorln(err)
				return err
			}
			err = deleteKey(ctx, key, RelationMutex)
			if err != nil {
				logger.Errorln(err)
				return err
			}
		} else if relationRecord != nil && StrToRelationActionType(actionTypeStr) == relation.RelationActionType_UN_FOLLOW {
			err = db.DelRelationByUserID(ctx, uint64(userId), uint64(toUserId))
			if err != nil {
				logger.Errorln(err)
				return err
			}
			err = deleteKey(ctx, key, RelationMutex)
			if err != nil {
				logger.Errorln(err)
				return err
			}
		}
	}
	return nil
}

func FavoriteMoveToDB() error {
	fmt.Println("sync favorite to db")
	ctx := context.Background()
	keys, err := getKeys(ctx, "video::*::user::*")
	if err != nil {
		logger.Errorln(err)
		return err
	}
	for _, key := range keys {
		value, err := getKeyValue(ctx, key)
		if err != nil {
			logger.Errorln(err)
			return err
		}
		valueSplit := strings.Split(value, "::")
		keySplit := strings.Split(key, "::")
		videoIdStr, userIdStr := keySplit[1], keySplit[3]
		_, actionTypeStr := valueSplit[0], valueSplit[1]
		videoId, err := strconv.ParseUint(videoIdStr, 10, 64)
		userId, err := strconv.ParseUint(userIdStr, 10, 64)
		if err != nil {
			logger.Errorln(err)
			return err
		}
		favoriteRecord, err := db.GetFavoriteVideoRelationByUserVideoID(ctx, uint64(userId), uint64(videoId))
		//若favorite record不存在，则创建like或dislike记录
		if favoriteRecord == nil {
			if StrToVideoActionType(actionTypeStr) == favorite.VideoActionType_LIKE {
				err = db.CreateVideoFavorite(ctx, userId, videoId, favorite.VideoActionType_LIKE)
				if err != nil {
					logger.Errorln(err)
					return err
				}
			} else if StrToVideoActionType(actionTypeStr) == favorite.VideoActionType_DISLIKE {
				err = db.CreateVideoFavorite(ctx, userId, videoId, favorite.VideoActionType_DISLIKE)
				if err != nil {
					logger.Errorln(err)
					return err
				}
			}
			err = deleteKey(ctx, key, RelationMutex)
			if err != nil {
				logger.Errorln(err)
				return err
			}
		} else {
			if StrToVideoActionType(actionTypeStr) == favorite.VideoActionType_DISLIKE && favoriteRecord.ActionType == favorite.VideoActionType_LIKE {
				err = db.DelFavoriteByUserVideoID(ctx, userId, videoId)
				if err != nil {
					logger.Errorln(err)
					return err
				}
				err = db.CreateVideoFavorite(ctx, userId, videoId, favorite.VideoActionType_DISLIKE)
				if err != nil {
					logger.Errorln(err)
					return err
				}
				err = deleteKey(ctx, key, RelationMutex)
				if err != nil {
					logger.Errorln(err)
					return err
				}
			}
			if StrToVideoActionType(actionTypeStr) == favorite.VideoActionType_LIKE && favoriteRecord.ActionType == favorite.VideoActionType_DISLIKE {
				err = db.DelFavoriteByUserVideoID(ctx, userId, videoId)
				if err != nil {
					logger.Errorln(err)
					return err
				}
				err = db.CreateVideoFavorite(ctx, userId, videoId, favorite.VideoActionType_LIKE)
				if err != nil {
					logger.Errorln(err)
					return err
				}
				err = deleteKey(ctx, key, RelationMutex)
				if err != nil {
					logger.Errorln(err)
					return err
				}
			}
			if StrToVideoActionType(actionTypeStr) == favorite.VideoActionType_CANCEL_LIKE &&
				favoriteRecord.ActionType == favorite.VideoActionType_LIKE {
				err = db.DelFavoriteByUserVideoID(ctx, userId, videoId)
				if err != nil {
					logger.Errorln(err)
					return err
				}
				err = deleteKey(ctx, key, RelationMutex)
				if err != nil {
					logger.Errorln(err)
					return err
				}
			}
			if StrToVideoActionType(actionTypeStr) == favorite.VideoActionType_CANCEL_DISLIKE &&
				favoriteRecord.ActionType == favorite.VideoActionType_DISLIKE {
				err = db.DelFavoriteByUserVideoID(ctx, userId, videoId)
				if err != nil {
					logger.Errorln(err)
					return err
				}
				err = deleteKey(ctx, key, RelationMutex)
				if err != nil {
					logger.Errorln(err)
					return err
				}
			}
		}
	}
	return nil
}

func SyncRelationToDB() {
	logger := zap.InitLogger()
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.Errorln(err)
	}
	_, err = scheduler.NewJob(
		gocron.DurationJob(time.Second),
		gocron.NewTask(RelationMoveToDB),
	)
	if err != nil {
		logger.Errorln(err)
	}
	scheduler.Start()
}

func SyncFavoriteToDB() {
	logger := zap.InitLogger()
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		logger.Errorln(err)
	}
	_, err = scheduler.NewJob(
		gocron.DurationJob(time.Second),
		gocron.NewTask(FavoriteMoveToDB),
	)
	if err != nil {
		logger.Errorln(err)
	}
	scheduler.Start()
}
