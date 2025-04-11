package redisDAO

import "wgxDouYin/pkg/zap"

var (
	logger = zap.InitLogger()
)

func init() {
	go ListenExpireFavorite()
}
