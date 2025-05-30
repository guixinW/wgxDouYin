package service

import (
	"crypto/ecdsa"
	"wgxDouYin/pkg/keys"
	"wgxDouYin/pkg/rabbitMQ"
	"wgxDouYin/pkg/zap"
)

var (
	KeyManager *keys.KeyManager
	logger     = zap.InitLogger()
	RelationMQ *rabbitmq.RabbitMQ
)

func init() {
	var err error
	RelationMQ, err = rabbitmq.NewRabbitMQInstance("relation")
	if err != nil {
		logger.Errorln(err)
	}
}

func Init(privateKey *ecdsa.PrivateKey, serviceName string) error {
	var err error
	KeyManager, err = keys.NewKeyManager(privateKey, serviceName)
	if err != nil {
		return err
	}
	go func() {
		err := consume()
		if err != nil {
			logger.Errorf(err.Error())
		}
	}()
	return nil
}
