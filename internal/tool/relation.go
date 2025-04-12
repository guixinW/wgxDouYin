package tool

import "wgxDouYin/grpc/relation"

func StrToRelationActionType(str string) relation.RelationActionType {
	if str == "0" {
		return relation.RelationActionType_FOLLOW
	} else if str == "1" {
		return relation.RelationActionType_UN_FOLLOW
	} else {
		return relation.RelationActionType_WRONG_TYPE
	}
}
