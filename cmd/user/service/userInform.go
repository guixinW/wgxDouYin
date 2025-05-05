package service

import (
	"context"
	"wgxDouYin/dal/db"
	"wgxDouYin/grpc/user"
)

func querySelf(ctx context.Context, userId uint64) (*user.UserInfoResponse, error) {
	usr, err := db.GetUserByID(ctx, userId)
	if err != nil {
		logger.Errorln(err.Error())
		res := &user.UserInfoResponse{
			StatusCode: -1,
			StatusMsg:  "获取用户信息失败：服务器内部错误",
		}
		return res, err
	} else if usr == nil {
		res := &user.UserInfoResponse{
			StatusCode: -1,
			StatusMsg:  "用户名不存在",
		}
		return res, nil
	} else {
		res := &user.UserInfoResponse{
			StatusCode: 0,
			StatusMsg:  "success",
			User: &user.User{
				Id:             uint64(usr.ID),
				Name:           usr.UserName,
				FollowerCount:  usr.FollowerCount,
				FollowingCount: usr.FollowingCount,
			},
		}
		return res, nil
	}
}

func queryOtherUser(ctx context.Context, otherUserId uint64) (*user.UserInfoResponse, error) {
	return nil, nil
}
