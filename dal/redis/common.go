package wgxRedis

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"time"
)

const syncTime = 5

func getKeys(ctx context.Context, keyPattern string) ([]string, error) {
	keys, err := GetRedisHelper().Keys(ctx, keyPattern).Result()
	if err != nil {
		return nil, ErrorWrap(err, "getKeys")
	}
	return keys, nil
}

func deleteKey(ctx context.Context, key string, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	if err != nil {
		return ErrorWrap(err, "deleteKey")
	}
	err = GetRedisHelper().Del(ctx, key).Err()
	if err != nil {
		return ErrorWrap(err, "deleteKey")
	}
	_, err = mutex.UnlockContext(ctx)
	if err != nil {
		return ErrorWrap(err, "deleteKey")
	}
	return nil
}

func setKey(ctx context.Context, key string, value string, expireTime time.Duration, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	if err != nil {
		return ErrorWrap(err, "setKey")
	}
	_, err = GetRedisHelper().Set(ctx, key, value, expireTime).Result()
	if err != nil {
		return ErrorWrap(err, "setKey")
	}
	_, err = mutex.UnlockContext(ctx)
	if err != nil {
		return ErrorWrap(err, "setKey")
	}
	return nil
}

func GoCronRelation() {
	ticker := time.NewTicker(syncTime * time.Second)
	for {
		select {
		case <-ticker.C:
			fmt.Println("test")
		}
	}
}
