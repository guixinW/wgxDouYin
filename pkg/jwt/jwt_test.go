package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
	"wgxDouYin/pkg/keys"
)

func TestParseJWT(t *testing.T) {
	privateKey, publicKey, _ := keys.CreateKeyPair()
	originClaims := CustomClaims{
		UserId: 1234,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "test",
			IssuedAt:  jwt.NewNumericDate(time.Now())},
	}
	start := time.Now()
	tokenString, err := CreateToken(privateKey, originClaims)
	end := time.Now().Sub(start)
	fmt.Printf("密钥签发耗时:%v\n", end)
	parseClaims, err := ParseToken(publicKey, tokenString)
	if err != nil {
		t.Fatalf("parse token string error %v", err)
	}
	if parseClaims.UserId != originClaims.UserId {
		t.Fatalf("parse error")
	}
}
