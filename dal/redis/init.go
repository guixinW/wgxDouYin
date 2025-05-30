package wgxRedis

import (
	"context"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
	"wgxDouYin/pkg/viper"
	"wgxDouYin/pkg/zap"
)

const KeyExpireTime = 24 * time.Hour
const LockExpireTime = 10 * time.Second

var (
	config         = viper.Init("db")
	logger         = zap.InitLogger()
	redisOnce      sync.Once
	redisHelper    *RedisHelper
	FavoriteMutex  *redsync.Mutex
	RelationMutex  *redsync.Mutex
	UserExistMutex *redsync.Mutex
	RS             *redsync.Redsync
)

type RedisHelper struct {
	//*redis.ClusterClient
	*redis.Client
}

func GetRedisHelper() *RedisHelper {
	return redisHelper
}

func CreateClusterClient(redisNames []string) *redis.ClusterClient {
	addr := []string{"127.0.0.1:6371", "127.0.0.1:6372", "127.0.0.1:6373"}
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addr,         // Redis 地址
		Password:     "1477364283", // 密码（如果有的话）
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	return rdb
}

func CreateClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6371", // Redis 地址
		Password:     "1477364283",     // 密码（如果有的话）
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	return rdb
}

//func InitRedisHelperByCluster(clusterClient *redis.ClusterClient) {
//	redisOnce.Do(func() {
//		redisHelper = new(RedisHelper)
//		redisHelper.Client = clusterClient
//	})
//}

func InitRedisHelperByClient(client *redis.Client) {
	redisOnce.Do(func() {
		redisHelper = new(RedisHelper)
		redisHelper.Client = client
	})
}

func init() {
	ctx := context.Background()
	cluster := CreateClient()
	InitRedisHelperByClient(cluster)
	if _, err := GetRedisHelper().Ping(ctx).Result(); err != nil {
		logger.Errorln(err.Error())
		return
	}
	pool := goredis.NewPool(cluster)
	RS = redsync.New(pool)
	UserExistMutex = RS.NewMutex("mutex-user-exist", redsync.WithExpiry(LockExpireTime))
	FavoriteMutex = RS.NewMutex("mutex-favorite", redsync.WithExpiry(LockExpireTime))
	RelationMutex = RS.NewMutex("mutex-relation", redsync.WithExpiry(LockExpireTime))
}
