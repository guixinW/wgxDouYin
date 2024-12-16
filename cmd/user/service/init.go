package service

import (
	"crypto/ecdsa"
	"fmt"
	"wgxDouYin/pkg/keys"
)

var (
	KeyManager *keys.KeyManager
)

func Init(privateKey *ecdsa.PrivateKey, serviceName string) error {
	var err error
	KeyManager, err = keys.NewKeyManager(privateKey, serviceName)
	fmt.Printf("user service key manage:%v\n", KeyManager)
	pub, _ := KeyManager.GetServerPublicKey(serviceName)
	fmt.Printf("user service public key %v\n", pub)
	if err != nil {
		return err
	}
	return nil
}
