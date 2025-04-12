package service

import (
	"context"
	"fmt"
	"wgxDouYin/dal/db"
	"wgxDouYin/grpc/comment"
	"wgxDouYin/grpc/user"
)

type CommentServiceImpl struct {
	comment.UnimplementedCommentServiceServer
}

func (s *CommentServiceImpl) CommentAction(ctx context.Context, req *comment.CommentActionRequest) (resp *comment.CommentActionResponse, err error) {
	if req.CommentId <= 0 || req.VideoId <= 0 || req.TokenUserId <= 0 {
		resp = &comment.CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法",
		}
	}
	operateComment := db.Comment{VideoID: req.VideoId, UserID: req.TokenUserId, Content: req.CommentText}
	switch req.ActionType {
	case comment.CommentActionType_COMMENT:
		err := db.CreateComment(ctx, &operateComment)
		if err != nil {
			resp = &comment.CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  err.Error(),
			}
			return resp, nil
		}
	case comment.CommentActionType_DELETE_COMMENT:
		err := db.DelComment(ctx, &operateComment)
		if err != nil {
			resp = &comment.CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  err.Error(),
			}
			return resp, nil
		}
	case comment.CommentActionType_WRONG_TYPE:
		resp = &comment.CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "错误的操作类型",
		}
		return resp, nil
	}
	resp = &comment.CommentActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	return resp, nil
}

func (s *CommentServiceImpl) CommentList(ctx context.Context, req *comment.CommentListRequest) (resp *comment.CommentListResponse, err error) {
	if req.TokenUserId <= 0 || req.VideoId <= 0 {
		resp = &comment.CommentListResponse{
			StatusCode: -1,
			StatusMsg:  "操作非法",
		}
		return resp, nil
	}
	commentRecords, err := db.GetCommentsByVideoId(ctx, req.VideoId)
	if err != nil {
		fmt.Println(err.Error())
		resp = &comment.CommentListResponse{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		}
		return resp, nil
	}
	commentList := make([]*comment.Comment, 0, len(commentRecords))
	for _, commentRecord := range commentRecords {
		commenter, err := db.GetUserByID(ctx, commentRecord.UserID)
		if err != nil {
			logger.Errorln(err.Error())
			res := &comment.CommentListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：获取作者信息失败",
			}
			return res, nil
		}
		commentList = append(commentList, &comment.Comment{
			Id: uint64(commentRecord.ID),
			User: &user.User{
				Id:             uint64(commenter.ID),
				Name:           commenter.UserName,
				FollowingCount: commenter.FollowingCount,
				FollowerCount:  commenter.FollowerCount,
				TotalFavorite:  commenter.FavoriteCount,
				WorkCount:      commenter.WorkCount,
				FavoriteCount:  commenter.FavoriteCount,
			},
			Content:  commentRecord.Content,
			CreateAt: uint64(commentRecord.CreatedAt.UnixMilli()),
		})
	}
	resp = &comment.CommentListResponse{
		StatusCode:  0,
		StatusMsg:   "success",
		CommentList: commentList,
	}
	return resp, nil
}
