package response

import user "wgxDouYin/grpc/user"

type RelationAction struct {
	Base
}

type FollowerList struct {
	Base
	UserList []*user.User `json:"user_list"`
}

type FollowList struct {
	Base
	UserList []*user.User `json:"user_list"`
}

type FriendList struct {
	Base
	UserList []*user.User `json:"user_list"`
}
