package wgxRedis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func TestRedisClusterConnection(t *testing.T) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6371", // Redis 地址
		Password: "1477364283",     // 密码（如无则留空）
		DB:       0,                // 默认 DB

		DialTimeout: 2 * time.Second,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("错误：%v\n", err)
	}
}

func TestRelationMoveToDB(t *testing.T) {
}

func TestAddSet(t *testing.T) {
	key := "dadabb"
	value := "testValue"
	if err := AddValueToKeySet(context.Background(), key, []string{value}); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestGetSet(t *testing.T) {
	key := fmt.Sprintf("following::%d", 3)
	sets, err := GetSet(context.Background(), key)
	if err != nil {
		t.Fatalf("GetSet err: %v", err)
	}
	fmt.Println(sets)
}

func TestSetCount(t *testing.T) {
	key := fmt.Sprintf("follower::3")
	count, err := GetSetCount(context.Background(), key)
	if err != nil {
		t.Fatalf("GetSet err: %v", err)
	}
	fmt.Println(count)
}

func TestIntersection(t *testing.T) {
	followingKey := fmt.Sprintf("following::%d", 1)
	followerKey := fmt.Sprintf("follower::%d", 1)
	res, err := GetSetIntersection(context.Background(), followingKey, followerKey)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(res)
	time.Sleep(2 * time.Second)
}

func TestExpire(t *testing.T) {
	expireTime := time.Now().Add(5 * time.Second)
	ctx := context.Background()
	err := SetKeyValue(ctx, "my_key", "my_value", expireTime)
	if err != nil {
		t.Fatalf("Error receiving message: %v", err)
		return
	}
	go func() {

	}()
	time.Sleep(10 * time.Second)
}

func TestIncrZSet(t *testing.T) {
	ctx := context.Background()
	videoRankName := ""
	videoId := "4"
	err := IncrNumInZSet(ctx, videoRankName, videoId, -1, FavoriteMutex)
	if err != nil {
		t.Fatalf("Error receiving message: %v", err)
	}
}
