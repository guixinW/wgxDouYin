package wgxRedis

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron/v2"
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
	err := RelationMoveToDB()
	if err != nil {
		t.Fatalf("RelationMoveToDB err: %v", err)
	}
}

func TestGetSet(t *testing.T) {
	key := fmt.Sprintf("following::%d", 1)
	sets, err := getSet(context.Background(), key)
	if err != nil {
		t.Fatalf("getSet err: %v", err)
	}
	fmt.Println(sets)
}

func TestGocronToDB(t *testing.T) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	_, err = scheduler.NewJob(
		gocron.DurationJob(time.Second),
		gocron.NewTask(func() {
			fmt.Println("test")
		}))
	scheduler.Start()
	time.Sleep(3 * time.Second)
	err = scheduler.Shutdown()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	time.Sleep(10 * time.Second)
}
