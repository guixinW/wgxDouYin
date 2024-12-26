package service

import (
	"crypto/ecdsa"
	"wgxDouYin/pkg/keys"
	"wgxDouYin/pkg/zap"
)

var (
	KeyManager *keys.KeyManager
	logger     = zap.InitLogger()
)

func Init(privateKey *ecdsa.PrivateKey, serviceName string) error {
	var err error
	KeyManager, err = keys.NewKeyManager(privateKey, serviceName)
	if err != nil {
		return err
	}
	return nil
}
