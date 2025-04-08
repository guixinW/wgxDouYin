package db

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
	"sync"
	"testing"
)

func TestCreateRelation(t *testing.T) {
	insertFollowRelation := FollowRelation{UserID: 1, ToUserID: 2}
	err := CreateRelation(context.Background(), &insertFollowRelation)
	if err != nil {
		t.Fatalf("CreateRelation err: %v", err)
	}
	fmt.Println(insertFollowRelation.CreatedAt)
}

func TestTestTabeleWriteAfterRead(t *testing.T) {
	test := Test{Id: 3, Category: "1", Name: "ok"}
	var ans []Test
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := CreateTest(context.Background(), &test)
		if err != nil {
			t.Errorf(err.Error())
		}
	}()
	go func() {
		defer wg.Done()
		ans, _ = ReadTest(context.Background())
	}()
	wg.Wait()
	fmt.Println(ans)
}

func TestTestTableRead(t *testing.T) {
	ans, _ := ReadTest(context.Background())
	fmt.Println(ans)
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
