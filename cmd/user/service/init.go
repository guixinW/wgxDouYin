package service

import (
	"crypto/ecdsa"
	"wgxDouYin/pkg/jwt"
)

var (
	KeyManager *jwt.KeyManager
)

func Init(privateKey *ecdsa.PrivateKey, serviceName string) {
	KeyManager = jwt.NewJWT(privateKey, &privateKey.PublicKey, serviceName)
}
