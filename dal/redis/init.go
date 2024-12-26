package wgxRedis

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

var (
	config        = viper.Init("db")
	zapLogger     = zap.InitLogger()
	redisOnce     sync.Once
	redisHelper   *RedisHelper
	FavoriteMutex *redsync.Mutex
	RelationMutex *redsync.Mutex
)

type RedisHelper struct {
	*redis.Client
}

func GetRedisHelper() *RedisHelper {
	return redisHelper
}

func CreateRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", config.Viper.GetString("redis.host"),
			config.Viper.GetString("redis.port")),
		Password:     config.Viper.GetString("redis.password"),
		DB:           config.Viper.GetInt("redis.db"),
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
	return rdb
}

func InitRedisHelper(client *redis.Client) {
	redisOnce.Do(func() {
		redisHelper = new(RedisHelper)
		redisHelper.Client = client
	})
}

func init() {
	ctx := context.Background()
	rdb := CreateRedisClient()
	InitRedisHelper(rdb)
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		zapLogger.Errorln(err.Error())
		return
	}
	zapLogger.Info("redis service connection successful!")
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)
	FavoriteMutex = rs.NewMutex("mutex-favorite")
	RelationMutex = rs.NewMutex("mutex-relation")
}
