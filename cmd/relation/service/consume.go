package service

import (
	"context"
	"encoding/json"
	"fmt"
	relationDao "wgxDouYin/cmd/relation/redisDAO"
	userDao "wgxDouYin/cmd/user/redisDao"
	"wgxDouYin/dal/db"
	redis "wgxDouYin/dal/redis"
	rabbitmq "wgxDouYin/pkg/rabbitMQ"
)

func consume() error {
	messages, err := RelationMQ.ConsumeSimple()
	ctx := context.Background()
	if err != nil {
		return err
	}
	for msg := range messages {
		rc := new(relationDao.RelationCache)
		if err = json.Unmarshal(msg.Body, &rc); err != nil {
			return err
		}
		fmt.Printf("==> Get new message: %v\n", rc)
		if inRank, err := relationDao.UpdateRelation(ctx, rc); err != nil {
			return err
		} else {
			if inRank {
				fmt.Printf("inRank, rc id:%v\n", rc.ToUserID)
				err = MoveUserInformToRedis(ctx, rc)
				if err != nil {
					return err
				}
			}
		}
		if !rabbitmq.GetServerAck("relation") {
			err := msg.Ack(false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func MoveUserInformToRedis(ctx context.Context, rc *relationDao.RelationCache) error {
	mutex := redis.RS.NewMutex(fmt.Sprintf("user_exist_lock::%d", rc.ToUserID))
	if err := mutex.LockContext(ctx); err != nil {
		return err
	}
	defer func() {
		if ok, err := mutex.UnlockContext(ctx); err != nil || !ok {
			logger.Error("unlock failed: ", err)
		}
	}()
	userInRedis, err := userDao.GetUserInform(ctx, rc.ToUserID)
	if err != nil {
		return err
	}
	if userInRedis == nil {
		user, err := db.GetUserByID(ctx, rc.ToUserID)
		if err != nil {
			return err
		}
		err = userDao.SetUserInform(ctx, &userDao.UserCache{
			UserID:    uint64(user.ID),
			UserName:  user.UserName,
			Following: user.FollowingCount,
			Follower:  user.FollowerCount,
			WorkCount: user.WorkCount,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
