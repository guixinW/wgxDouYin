package db

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
	"testing"
	"time"
	"wgxDouYin/grpc/favorite"
)

func TestCreateRelation(t *testing.T) {
	insertFollowRelation := FollowRelation{UserID: 1, ToUserID: 100}
	err := CreateRelation(context.Background(), &insertFollowRelation)
	if err != nil {
		t.Fatalf("CreateRelation err: %v", err)
	}
	fmt.Println(insertFollowRelation.CreatedAt)
}

func TestMySQLTransaction(t *testing.T) {
	addSum := func(sum *int) {
		*sum += 1
	}
	sum := 0
	err := GetDB().Clauses(dbresolver.Write).WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {

		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "to_user_id"}},
			DoNothing: true,
		}).Create(&FollowRelation{UserID: 1, ToUserID: 3}).Error
		if err != nil {
			return err
		}
		addSum(&sum)
		err = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "to_user_id"}},
			DoNothing: true,
		}).Create(&FollowRelation{UserID: 11, ToUserID: 12}).Error
		if err != nil {
			return err
		}
		return nil
	})
	fmt.Println(sum)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestFavoriteAction(t *testing.T) {
	ctx := context.Background()
	var userId uint64
	var videoId uint64
	userId = 5
	videoId = 4
	originUser, _ := GetUserByID(ctx, userId)
	newVideoRelation := FavoriteVideoRelation{VideoID: videoId, UserID: userId, ActionType: favorite.VideoActionType_LIKE}
	err := CreateVideoRelation(ctx, &newVideoRelation)
	if err != nil {
		t.Fatalf(err.Error())
	}
	time.Sleep(1 * time.Second)
	afterActionUser, _ := GetUserByID(ctx, userId)
	if originUser.FavoriteCount != afterActionUser.FavoriteCount-1 {
		t.Fatalf("video action 操作失败, originUser:%v, afterActionUser:%v\n", originUser.FavoriteCount, afterActionUser.FavoriteCount)
	}
}

func TestDislikeAction(t *testing.T) {
	ctx := context.Background()
	var userId uint64
	var videoId uint64
	userId = 5
	videoId = 4
	originUser, _ := GetUserByID(ctx, userId)
	newVideoRelation := FavoriteVideoRelation{VideoID: videoId, UserID: userId, ActionType: favorite.VideoActionType_DISLIKE}
	err := CreateVideoRelation(ctx, &newVideoRelation)
	if err != nil {
		t.Fatalf(err.Error())
	}
	time.Sleep(1 * time.Second)
	afterActionUser, _ := GetUserByID(ctx, userId)
	if originUser.DislikeCount != afterActionUser.DislikeCount-1 {
		t.Fatalf("video action 操作失败, originUser:%v, afterActionUser:%v\n", originUser.DislikeCount, afterActionUser.DislikeCount)
	}
}
