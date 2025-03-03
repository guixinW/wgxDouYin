package wgxRedis

import (
	"context"
	"github.com/go-co-op/gocron/v2"
	"strconv"
	"strings"
	"time"
	"wgxDouYin/dal/db"
	"wgxDouYin/pkg/zap"
)

func RelationMoveToDB() error {
	logger := zap.InitLogger()
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
		relation, err := db.GetRelationByUserIDs(ctx, uint64(userId), uint64(toUserId))
		if err != nil {
			logger.Errorln(err)
			return err
		} else if relation == nil && actionTypeStr == "0" {
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
		} else if relation != nil && actionTypeStr == "1" {
			err = db.CreateRelation(ctx, uint64(userId), uint64(toUserId))
			if err != nil {
				logger.Errorln(err)
			}
			err = deleteKey(ctx, key, RelationMutex)
			if err != nil {
				logger.Errorln(err)
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
		gocron.NewTask(RelationMoveToDB()),
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
		gocron.NewTask(RelationMoveToDB()),
	)
	if err != nil {
		logger.Errorln(err)
	}
	scheduler.Start()
}
