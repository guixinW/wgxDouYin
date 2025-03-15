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

const ExpireTime = 5 * time.Second

var (
	config        = viper.Init("db")
	logger        = zap.InitLogger()
	redisOnce     sync.Once
	redisHelper   *RedisHelper
	FavoriteMutex *redsync.Mutex
	RelationMutex *redsync.Mutex
)

type RedisHelper struct {
	*redis.ClusterClient
}

func GetRedisHelper() *RedisHelper {
	return redisHelper
}

func CreateFailoverClusterClient(redisNames []string) *redis.ClusterClient {
	sentinelAddresses := make([]string, 0)
	for _, sentinelName := range redisNames {
		ip := config.Viper.GetString(fmt.Sprintf("redis.%v.host", sentinelName))
		port := config.Viper.GetString(fmt.Sprintf("redis.%v.port", sentinelName))
		sentinelAddresses = append(sentinelAddresses, fmt.Sprintf("%s:%s", ip, port))
	}
	password := config.Viper.GetString(fmt.Sprintf("redis.%v.password", redisNames[0]))
	db := config.Viper.GetInt(fmt.Sprintf("redis.%v.db", redisNames[0]))
	rdb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:       "mymaster",
		SentinelAddrs:    sentinelAddresses,
		Password:         password,
		SentinelPassword: password,
		DB:               db,
		DialTimeout:      2 * time.Second,
		ReadTimeout:      2 * time.Second,
		WriteTimeout:     2 * time.Second,
	})
	return rdb
}

func InitRedisHelper(clusterClient *redis.ClusterClient) {
	redisOnce.Do(func() {
		redisHelper = new(RedisHelper)
		redisHelper.ClusterClient = clusterClient
	})
}

func init() {
	ctx := context.Background()
	cluster := CreateFailoverClusterClient([]string{"sentinel1", "sentinel2", "sentinel3"})
	InitRedisHelper(cluster)
	if _, err := GetRedisHelper().Ping(ctx).Result(); err != nil {
		logger.Errorln(err.Error())
		return
	}
	pool := goredis.NewPool(cluster)
	rs := redsync.New(pool)
	FavoriteMutex = rs.NewMutex("mutex-favorite")
	RelationMutex = rs.NewMutex("mutex-relation")
	go SyncRelationToDB()
	go SyncFavoriteToDB()
}
