package wgxRedis

import (
	"context"
	"errors"
	"github.com/go-redsync/redsync/v4"
	"time"
)

func getKeys(ctx context.Context, keyPattern string) ([]string, error) {
	keys, err := GetRedisHelper().Keys(ctx, keyPattern).Result()
	if err != nil {
		return nil, err
	}
	return keys, err
}

func deleteKey(ctx context.Context, key string, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	if err != nil {
		return errors.New("redis lock failed: " + err.Error())
	}
	err = GetRedisHelper().Del(ctx, key).Err()
	if err != nil {
		return errors.New("redis del failed: " + err.Error())
	}
	_, err = mutex.UnlockContext(ctx)
	if err != nil {
		return errors.New("redis unlock failed: " + err.Error())
	}
	return nil
}

func setKey(ctx context.Context, key string, value string, expireTime time.Duration, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	if err != nil {
		return errors.New("redis lock failed: " + err.Error())
	}
	_, err = GetRedisHelper().Set(ctx, key, value, expireTime).Result()
	if err != nil {
		return errors.New("redis set failed: " + err.Error())
	}
	_, err = mutex.UnlockContext(ctx)
	if err != nil {
		return errors.New("redis unlock failed: " + err.Error())
	}
	return nil
}
