package keys

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
	myJwt "wgxDouYin/pkg/jwt"
)

func TestMutualConversion(t *testing.T) {
	privateKey, publicKey, err := CreateKeyPair()
	if privateKey == nil || publicKey == nil {
		t.Fatalf("Failed to create key pair")
	}
	if err != nil {
		t.Fatalf(err.Error())
	}
	publicStr, err := PublicKeyToPEM(publicKey)
	if err != nil {
		t.Fatalf(err.Error())
	}
	duplicationPublicKey, err := PEMToPublicKey(publicStr)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if duplicationPublicKey.Curve != publicKey.Curve {
		t.Fatalf("two key's curve is not equal.")
	}
	if duplicationPublicKey.X.Cmp(publicKey.X) != 0 && duplicationPublicKey.Y.Cmp(publicKey.Y) != 0 {
		t.Fatalf("two key's X and Y are not equal")
	}
}

func TestCreateKeyParse(t *testing.T) {
	privateKey, publicKey, _ := CreateKeyPair()
	originClaims := myJwt.CustomClaims{
		UserId: 1234,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "test",
			IssuedAt:  jwt.NewNumericDate(time.Now())},
	}
	tokenString, err := myJwt.CreateToken(privateKey, originClaims)
	if err != nil {
		t.Fatalf(err.Error())
	}
	pbkParseStr1, err := myJwt.ParseToken(publicKey, tokenString)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if originClaims.UserId != pbkParseStr1.UserId {
		t.Fatalf("orignal key parse error")
	}
	publicStr, err := PublicKeyToPEM(publicKey)
	if err != nil {
		t.Fatalf(err.Error())
	}
	duplicationPublicKey, err := PEMToPublicKey(publicStr)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if duplicationPublicKey == nil {
		t.Fatalf("PEM to public key failed, duplication public key is nil")
	}
	pbkParseStr2, err := myJwt.ParseToken(duplicationPublicKey, tokenString)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if pbkParseStr2 == nil {
		t.Fatalf("ParseToken failed, pbkParseStr2 is nil")
	}
	if originClaims.UserId != pbkParseStr2.UserId {
		t.Fatalf("orignal key parse error")
	}
}

func TestLoadKeyParse(t *testing.T) {
	privateKeyPath := fmt.Sprintf("../../cmd/user/keys/%v.pem", "wgxDouYinUserServer")
	privateKey, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	publicKey := &privateKey.PublicKey
	fmt.Printf("public key is %v\n", publicKey)
	originClaims := myJwt.CustomClaims{
		UserId: 1234,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "test",
			IssuedAt:  jwt.NewNumericDate(time.Now())},
	}
	tokenString, err := myJwt.CreateToken(privateKey, originClaims)
	if err != nil {
		t.Fatalf(err.Error())
	}
	pbkParseStr1, err := myJwt.ParseToken(publicKey, tokenString)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if originClaims.UserId != pbkParseStr1.UserId {
		t.Fatalf("orignal key parse error")
	}
	publicStr, err := PublicKeyToPEM(publicKey)
	if err != nil {
		t.Fatalf(err.Error())
	}
	duplicationPublicKey, err := PEMToPublicKey(publicStr)
	if err != nil {
		t.Fatalf(err.Error())
	}
	pbkParseStr2, err := myJwt.ParseToken(duplicationPublicKey, tokenString)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if originClaims.UserId != pbkParseStr2.UserId {
		t.Fatalf("orignal key parse error")
	}
}
