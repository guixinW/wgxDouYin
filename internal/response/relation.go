package response

import user "wgxDouYin/grpc/user"
import relation "wgxDouYin/grpc/relation"

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
	FriendList []*relation.FriendUser `json:"friend_list"`
}
