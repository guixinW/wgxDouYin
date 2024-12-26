package wgxRedis

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestSetKey(t *testing.T) {
	testKey := "test key"
	testValue := "test value"
	expireTime := time.Duration(10) * time.Minute
	err := setKey(context.Background(), testKey, testValue, expireTime, FavoriteMutex)
	if err != nil {
		t.Error(err)
	}
	str, err := getKeys(context.Background(), testKey)
	if err != nil {
		t.Error(err)
	}
	if len(str) != 1 {
		t.Error(errors.New("wrong get"))
	}
	if str[0] != testKey {
		t.Error(fmt.Errorf("wrong get: %v", str))
	}
}
