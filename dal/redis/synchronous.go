package wgxRedis

import (
	"github.com/go-co-op/gocron/v2"
	"time"
	"wgxDouYin/pkg/zap"
)

func RelationSyncWithDB() error {
	return nil
}

func FavoriteSyncWithDB() error {
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
		gocron.NewTask(RelationSyncWithDB),
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
		gocron.NewTask(FavoriteSyncWithDB),
	)
	if err != nil {
		logger.Errorln(err)
	}
	scheduler.Start()
}
