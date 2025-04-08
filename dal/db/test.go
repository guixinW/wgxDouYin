package db

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Test struct {
	Id       uint64
	Category string
	Name     string
}

func (Test) TableName() string {
	return "test"
}

func CreateTest(ctx context.Context, test *Test) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(test).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func ReadTest(ctx context.Context) ([]Test, error) {
	var res []Test
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
