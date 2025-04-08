package wgxRedis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	ctx := context.Background()
	addr := "127.0.0.1:6379"
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,         // Redis 地址
		Password:     "1477364283", // 密码（如果有的话）
		DB:           0,            // 使用默认 DB
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("无法连接到 Redis 实例：%s，错误：%v\n", addr, err)
	}
}

func TestSentinelPing(t *testing.T) {
	sentinel := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName: "mymaster",
		SentinelAddrs: []string{"127.0.0.1:26379",
			"127.0.0.1:26380",
			"127.0.0.1:26381"},
		Password:         "1477364283",
		SentinelPassword: "1477364283",
		DB:               0,
		DialTimeout:      2 * time.Second,
		ReadTimeout:      2 * time.Second,
		WriteTimeout:     2 * time.Second,
	})
	pong, err := sentinel.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("ping err: %v", err)
	}
	fmt.Println(pong)
}

func TestSentinelWrite(t *testing.T) {
	sentinel := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName: "mymaster",
		SentinelAddrs: []string{"127.0.0.1:26379",
			"127.0.0.1:26380",
			"127.0.0.1:26381"},
		Password:         "1477364283",
		SentinelPassword: "1477364283",
		DB:               0,
		DialTimeout:      2 * time.Second,
		ReadTimeout:      2 * time.Second,
		WriteTimeout:     2 * time.Second,
	})

	err := sentinel.Set(context.Background(), "key", "value", 0).Err()
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}
	fmt.Println("Key set successfully")
}

func TestRelationMoveToDB(t *testing.T) {
}

func TestGetSet(t *testing.T) {
	key := fmt.Sprintf("following::%d", 1)
	sets, err := getSet(context.Background(), key)
	if err != nil {
		t.Fatalf("getSet err: %v", err)
	}
	fmt.Println(sets)
}

func TestSetCount(t *testing.T) {
	key := fmt.Sprintf("follower::3")
	count, err := getSetCount(context.Background(), key)
	if err != nil {
		t.Fatalf("getSet err: %v", err)
	}
	fmt.Println(count)
}

func TestIntersection(t *testing.T) {
	followingKey := fmt.Sprintf("following::%d", 1)
	followerKey := fmt.Sprintf("follower::%d", 1)
	res, err := getSetIntersection(context.Background(), followingKey, followerKey)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(res)
	time.Sleep(2 * time.Second)
}

func TestExpire(t *testing.T) {
	expireTime := time.Now().Add(5 * time.Second)
	ctx := context.Background()
	err := setKeyValue(ctx, "my_key", "my_value", expireTime, RelationMutex)
	if err != nil {
		t.Fatalf("Error receiving message: %v", err)
		return
	}
	go func() {

	}()
	time.Sleep(10 * time.Second)
}
