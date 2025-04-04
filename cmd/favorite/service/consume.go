package service

import (
	"context"
	"encoding/json"
	"fmt"
	wgxRedis "wgxDouYin/dal/redis"
	rabbitmq "wgxDouYin/pkg/rabbitMQ"
)

func consume() error {
	messages, err := RelationMQ.ConsumeSimple()
	if err != nil {
		logger.Errorf("RelationMQ Err: %s", err.Error())
		return err
	}
	for msg := range messages {
		rc := new(wgxRedis.FavoriteCache)
		if err = json.Unmarshal(msg.Body, &rc); err != nil {
			fmt.Println("json unmarshal error:" + err.Error())
			logger.Errorf("RelationMQ Err: %s", err.Error())
			continue
		}
		fmt.Printf("==> Get new message: %v\n", rc)
		if err = wgxRedis.UpdateFavorite(context.Background(), rc); err != nil {
			fmt.Println("add to redis error:" + err.Error())
			logger.Errorf("RelationMQ Err: %s", err.Error())
			continue
		}
		if !rabbitmq.GetServerAck("relation") {
			err := msg.Ack(true)
			if err != nil {
				logger.Errorf("ack error: %s", err.Error())
				return err
			}
		}
	}
	return nil
}
