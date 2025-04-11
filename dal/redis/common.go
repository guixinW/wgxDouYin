package wgxRedis

import (
	"context"
	"github.com/go-redsync/redsync/v4"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

const syncTime = 5

func ErrorWrap(err error, warpMessage string) error {
	return errors.Wrap(err, warpMessage)
}

func ParseMillisTimestamp(ts string) (time.Time, error) {
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Time{}, errors.Errorf("时间戳解析失败: %v", err)
	}
	return time.UnixMilli(tsInt), nil
}

func GetKeys(ctx context.Context, keyPattern string) ([]string, error) {
	keys, err := GetRedisHelper().Keys(ctx, keyPattern).Result()
	if err != nil {
		return nil, ErrorWrap(err, "GetKeys")
	}
	return keys, nil
}

func DeleteKey(ctx context.Context, key string, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	defer func(mutex *redsync.Mutex, ctx context.Context) {
		_, err := mutex.UnlockContext(ctx)
		if err != nil {
			logger.Errorln(err)
		}
	}(mutex, ctx)
	if err != nil {
		return ErrorWrap(err, "DeleteKey")
	}
	err = GetRedisHelper().Del(ctx, key).Err()
	if err != nil {
		return ErrorWrap(err, "DeleteKey")
	}
	return nil
}

func GetKeyValue(ctx context.Context, key string) (string, error) {
	value, err := GetRedisHelper().Get(ctx, key).Result()
	if err != nil {
		return "", ErrorWrap(err, "getKey")
	}
	return value, nil
}

func GetSet(ctx context.Context, key string) ([]string, error) {
	results, err := GetRedisHelper().SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return results, nil
}

func GetSetCount(ctx context.Context, key string) (uint64, error) {
	count, err := GetRedisHelper().SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return uint64(count), nil
}

func GetSetIntersection(ctx context.Context, key ...string) ([]string, error) {
	result, err := GetRedisHelper().SInter(ctx, key...).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func AddValueToKeySet(ctx context.Context, key string, value []string, mutex *redsync.Mutex) error {
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

func DelValueFormKeySet(ctx context.Context, key string, value []string, mutex *redsync.Mutex) error {
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

func IsValueExistInKeySet(ctx context.Context, key string, value string) (bool, error) {
	isExist, err := GetRedisHelper().SIsMember(ctx, key, value).Result()
	if err != nil {
		return false, err
	}
	return isExist, nil
}

func SetKeyValue(ctx context.Context, key string, value string, expireTime time.Time, mutex *redsync.Mutex) error {
	err := mutex.LockContext(ctx)
	defer func(mutex *redsync.Mutex, ctx context.Context) {
		_, err := mutex.UnlockContext(ctx)
		if err != nil {
			logger.Errorln(err)
		}
	}(mutex, ctx)
	if err != nil {
		return ErrorWrap(err, "SetKeyValue")
	}
	_, err = GetRedisHelper().Set(ctx, key, value, time.Until(expireTime)).Result()
	if err != nil {
		return ErrorWrap(err, "SetKeyValue")
	}
	return nil
}

func IsKeyExist(ctx context.Context, key string) (bool, error) {
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

func RunUsingLua(ctx context.Context, keys []string, args []interface{}, scriptStr string, mutex *redsync.Mutex) error {
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
