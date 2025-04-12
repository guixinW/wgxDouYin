package service

import (
	"crypto/ecdsa"
	"wgxDouYin/pkg/keys"
	"wgxDouYin/pkg/zap"
)

var (
	KeyManager *keys.KeyManager
	logger     = zap.InitLogger()
	//FavoriteMQ *rabbitmq.RabbitMQ
)

func init() {
	//var err error
	//FavoriteMQ, err = rabbitmq.DefaultRabbitMQInstance("favorite")
	//if err != nil {
	//	panic(err)
	//}
}

func Init(privateKey *ecdsa.PrivateKey, serviceName string) error {
	var err error
	KeyManager, err = keys.NewKeyManager(privateKey, serviceName)
	if err != nil {
		return err
	}
	//go func() {
	//	err := consume()
	//	if err != nil {
	//		logger.Errorf(err.Error())
	//	}
	//}()
	return nil
}
