package relatioDAO

import (
	"wgxDouYin/pkg/zap"
)

var (
	logger = zap.InitLogger()
)

func init() {
	go ListenExpireRelation()
}
