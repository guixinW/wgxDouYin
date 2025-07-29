package model

// User 领域模型
type User struct {
	ID            int64
	Username      string
	Password      string
	FollowCount   int64
	FollowerCount int64
	IsFollow      bool
}
