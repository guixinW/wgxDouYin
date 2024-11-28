package service

import (
	"crypto/ecdsa"
	"wgxDouYin/pkg/jwt"
)

var (
	JWT *jwt.KeyManager
)

func Init(privateKey *ecdsa.PrivateKey) {
	JWT = jwt.NewJWT(privateKey, &privateKey.PublicKey, "user")
}
