package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"wgxDouYin/pkg/viper"
	zap2 "wgxDouYin/pkg/zap"
)

var (
	config = viper.Init("rabbitmq")
	logger *zap.SugaredLogger
	conn   *amqp.Connection
)

func init() {
	logger = zap2.InitLogger()
}
