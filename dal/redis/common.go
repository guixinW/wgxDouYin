package wgxRedis

import (
	"context"
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

func getKeyValue(ctx context.Context, key string) (string, error) {
	value, err := GetRedisHelper().Get(ctx, key).Result()
	if err != nil {
		return "", ErrorWrap(err, "getKey")
	}
	return value, nil
}

func getSet(ctx context.Context, key string) ([]string, error) {
	results, err := GetRedisHelper().SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return results, nil
}

func getSetCount(ctx context.Context, key string) (uint64, error) {
	count, err := GetRedisHelper().SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return uint64(count), nil
}

func getSetIntersection(ctx context.Context, key ...string) ([]string, error) {
	result, err := GetRedisHelper().SInter(ctx, key...).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func addKeyToSet(ctx context.Context, key string, value []string, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	if err != nil {
		return ErrorWrap(err, "add key")
	}
	err = GetRedisHelper().SAdd(ctx, key, value).Err()
	if err != nil {
		return err
	}
	_, err = mutex.UnlockContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func delKeyFormSet(ctx context.Context, key string, value []string, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	if err != nil {
		return ErrorWrap(err, "delete key")
	}
	err = GetRedisHelper().SRem(ctx, key, value).Err()
	if err != nil {
		return err
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
