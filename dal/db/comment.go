package db

import (
	"context"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Comment struct {
	gorm.Model
	Video      Video  `gorm:"foreignkey:VideoID" json:"video,omitempty"`
	VideoID    uint64 `gorm:"index:idx_videoid;not null" json:"video_id"`
	User       User   `gorm:"foreignkey:UserID" json:"user,omitempty"`
	UserID     uint64 `gorm:"index:idx_userid;not null" json:"user_id"`
	Content    string `gorm:"type:varchar(255);not null" json:"content"`
	LikeCount  uint64 `gorm:"column:like_count;default:0;not null" json:"like_count,omitempty"`
	TeaseCount uint64 `gorm:"column:tease_count;default:0;not null" json:"tease_count,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}

func CreateComment(ctx context.Context, addComment *Comment) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Create(addComment)
		//判断是否出现错误以及是否未正常插入
		if res.Error != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(res.Error, &mysqlErr) {
				if mysqlErr.Number == 1452 {
					return errors.New("评论的视频或者用户不存在")
				}
			}
			return res.Error
		}

		//将add comment增加1
		res = tx.Model(&Video{}).Where("id = ?", addComment.VideoID).Update("comment_count", gorm.Expr("comment_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(uint64(addComment.ID), "DelCommentByCommentId", Delete)
		}
		return nil
	})
	return err
}

func DelComment(ctx context.Context, deleteComment *Comment) error {
	if deleteComment.ID <= 0 {
		return errors.New("不存在的comment id")
	}

	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		//删除评论
		res := tx.Unscoped().Delete(deleteComment)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("不能删除不存在的评论")
		}

		//将video表中的comment_count减1
		res = tx.Model(&Video{}).Where("id = ?", deleteComment.VideoID).Update("comment_count", gorm.Expr("comment_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return NewDatabaseErrorMessage(uint64(deleteComment.ID), "DelCommentByCommentId", Delete)
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func GetCommentsByVideoId(ctx context.Context, videoId uint64) ([]*Comment, error) {
	if videoId <= 0 {
		return nil, errors.New("不存在的video dd")
	}
	var commentList []*Comment
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Where("video_id = ?", videoId).Find(&commentList).Error
	fmt.Printf("created_at:%v\n", commentList[0].CreatedAt)
	if err != nil {
		return nil, err
	}
	fmt.Printf("commentList:%v\n", commentList)
	return commentList, nil
}
