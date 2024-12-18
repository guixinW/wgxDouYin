package service

import (
	"crypto/ecdsa"
	"wgxDouYin/pkg/keys"
)

var (
	KeyManager *keys.KeyManager
)

func Init(privateKey *ecdsa.PrivateKey, serviceName string) error {
	var err error
	KeyManager, err = keys.NewKeyManager(privateKey, serviceName)
	if err != nil {
		return err
	}
	return nil
}
