package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
	"wgxDouYin/pkg/keys"
)

func TestLoadJWT(t *testing.T) {

}

func TestJWT(t *testing.T) {
	privateKey, publicKey, err := keys.CreateKeyPair()
	if err != nil {
		t.Fatalf("create key pairs err: %v", err)
	}
	registerServerName := "register"
	registerJWT := NewJWT(privateKey, publicKey, registerServerName)
	originClaims := CustomClaims{
		1234,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "test",
			IssuedAt:  jwt.NewNumericDate(time.Now())},
	}
	tokenString, err := registerJWT.CreateToken(originClaims)
	if err != nil {
		t.Fatalf("create token error %v", err)
	}
	otherPrivateKey, otherPublicKey, err := keys.CreateKeyPair()
	if err != nil {
		t.Fatalf("create key pairs err: %v", err)
	}
	loginServerName := "login"
	loginJWT := NewJWT(otherPrivateKey, otherPublicKey, loginServerName)
	err = loginJWT.addServerPublicKey(registerServerName, publicKey)
	if err != nil {
		t.Fatalf("add server public key error %v", err)
	}
	parseClaims, err := loginJWT.ParseToken(tokenString, registerServerName)
	if err != nil {
		t.Fatalf("parse token string error %v", err)
	}
	if parseClaims.UserId != originClaims.UserId {
		t.Fatalf("parse error")
	}
}
