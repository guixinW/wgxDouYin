package wgxRedis

import (
	"context"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
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
	defer func(mutex *redsync.Mutex, ctx context.Context) {
		_, err := mutex.UnlockContext(ctx)
		if err != nil {
			logger.Errorln(err)
		}
	}(mutex, ctx)
	if err != nil {
		return ErrorWrap(err, "deleteKey")
	}
	err = GetRedisHelper().Del(ctx, key).Err()
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

func addValueToKeySet(ctx context.Context, key string, value []string, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	defer func(mutex *redsync.Mutex, ctx context.Context) {
		_, err := mutex.UnlockContext(ctx)
		if err != nil {
			logger.Errorln(err)
		}
	}(mutex, ctx)
	if err != nil {
		return ErrorWrap(err, "add key")
	}
	err = GetRedisHelper().SAdd(ctx, key, value).Err()
	if err != nil {
		return err
	}
	return nil
}

func delValueFormKeySet(ctx context.Context, key string, value []string, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	defer func(mutex *redsync.Mutex, ctx context.Context) {
		_, err := mutex.UnlockContext(ctx)
		if err != nil {
			logger.Errorln(err)
		}
	}(mutex, ctx)
	if err != nil {
		return ErrorWrap(err, "delete key")
	}
	err = GetRedisHelper().SRem(ctx, key, value).Err()
	if err != nil {
		return err
	}
	return nil
}

func isValueExistInKeySet(ctx context.Context, key string, value string) (bool, error) {
	isExist, err := GetRedisHelper().SIsMember(ctx, key, value).Result()
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func setKeyValue(ctx context.Context, key string, value string, expireTime time.Time, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	defer func(mutex *redsync.Mutex, ctx context.Context) {
		_, err := mutex.UnlockContext(ctx)
		if err != nil {
			logger.Errorln(err)
		}
	}(mutex, ctx)
	if err != nil {
		return ErrorWrap(err, "setKeyValue")
	}
	_, err = GetRedisHelper().Set(ctx, key, value, time.Until(expireTime)).Result()
	if err != nil {
		return ErrorWrap(err, "setKeyValue")
	}
	return nil
}

func isKeyExist(ctx context.Context, key string) (bool, error) {
	keyExisted, err := GetRedisHelper().Exists(ctx, key).Result()
	return keyExisted == 1, err
}

func IncrNumInZSet(ctx context.Context, key string, value string, incrNum float64, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	defer func(mutex *redsync.Mutex, ctx context.Context) {
		_, err := mutex.UnlockContext(ctx)
		if err != nil {
			logger.Errorln(err)
		}
	}(mutex, ctx)
	if err != nil {
		return ErrorWrap(err, "IncrZSet get lock error")
	}
	err = GetRedisHelper().ZIncrBy(ctx, key, incrNum, value).Err()
	if err != nil {
		return ErrorWrap(err, "IncrZSet incr error")
	}
	return nil
}

func runUsingLua(ctx context.Context, keys []string, args []interface{}, scriptStr string, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	defer func(mutex *redsync.Mutex, ctx context.Context) {
		_, err := mutex.UnlockContext(ctx)
		if err != nil {
			logger.Errorln(err)
		}
	}(mutex, ctx)
	if err != nil {
		return ErrorWrap(err, "RunUsingLua")
	}
	script := redis.NewScript(scriptStr)
	_, err = script.Run(ctx, GetRedisHelper(), keys, args...).Result()
	if err != nil {
		return ErrorWrap(err, "RunUsingLua")
	}
	return nil
}
