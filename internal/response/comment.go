package response

import "wgxDouYin/grpc/comment"

type CommentAction struct {
	Base
	Comment *comment.Comment `json:"comment"`
}

type CommentList struct {
	Base
	CommentList []*comment.Comment `json:"comment_list"`
}
