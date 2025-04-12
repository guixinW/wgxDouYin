package tool

import (
	"wgxDouYin/grpc/comment"
)

func StrToCommentActionType(str string) comment.CommentActionType {
	if str == "0" {
		return comment.CommentActionType_COMMENT
	} else if str == "1" {
		return comment.CommentActionType_DELETE_COMMENT
	}
	return comment.CommentActionType_WRONG_TYPE
}
