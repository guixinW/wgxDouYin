package response

import "wgxDouYin/grpc/user"

type Register struct {
	Base
	UserID uint64 `json:"user_id"`
	Token  string `json:"token"`
}

type Login struct {
	Base
	UserID uint64 `json:"user_id"`
	Token  string `json:"token"`
}

type UserInform struct {
	Base
	User *user.User `json:"user"`
}
