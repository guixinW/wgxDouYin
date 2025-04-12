package tool

import "wgxDouYin/grpc/favorite"

func StrToVideoActionType(str string) favorite.VideoActionType {
	if str == "0" {
		return favorite.VideoActionType_LIKE
	} else if str == "1" {
		return favorite.VideoActionType_DISLIKE
	} else if str == "2" {
		return favorite.VideoActionType_CANCEL_LIKE
	} else if str == "3" {
		return favorite.VideoActionType_CANCEL_DISLIKE
	}
	return favorite.VideoActionType_WRONG_TYPE
}
