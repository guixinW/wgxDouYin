package userDAO

import (
	"context"
	"fmt"
	"strconv"
	"time"
	wgxRedis "wgxDouYin/dal/redis"
)

type UserCache struct {
	UserID    uint64    `json:"user_id" redis:"user_id"`
	UserName  string    `json:"user_name" redis:"user_name"`
	Following uint64    `json:"following" redis:"following"`
	Follower  uint64    `json:"follower" redis:"follower"`
	WorkCount uint64    `json:"work_count" redis:"work_count"`
	UpdatedAt time.Time `json:"updated_at" redis:"updated_at"`
	CreatedAt time.Time `json:"created_at" redis:"created_at"`
}

func SetUserInform(ctx context.Context, user *UserCache) error {
	key := fmt.Sprintf("user_cache:%v", user.UserID)
	data := map[string]interface{}{
		"user_id":    user.UserID,
		"user_name":  user.UserName,
		"following":  user.Following,
		"follower":   user.Follower,
		"work_count": user.WorkCount,
		"updated_at": user.UpdatedAt.Format(time.RFC3339),
		"created_at": user.CreatedAt.Format(time.RFC3339),
	}
	return wgxRedis.GetRedisHelper().HSet(ctx, key, data).Err()
}

func GetUserInform(ctx context.Context, userId uint64) (*UserCache, error) {
	key := fmt.Sprintf("user_cache:%v", userId)
	result, err := wgxRedis.GetRedisHelper().HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}

	uid, _ := strconv.ParseUint(result["user_id"], 10, 64)
	following, _ := strconv.ParseUint(result["following"], 10, 64)
	follower, _ := strconv.ParseUint(result["follower"], 10, 64)
	workCount, _ := strconv.ParseUint(result["work_count"], 10, 64)
	createdAt, _ := time.Parse(time.RFC3339, result["created_at"])
	updatedAt, _ := time.Parse(time.RFC3339, result["updated_at"])

	return &UserCache{
		UserID:    uid,
		UserName:  result["user_name"],
		Following: following,
		Follower:  follower,
		WorkCount: workCount,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
