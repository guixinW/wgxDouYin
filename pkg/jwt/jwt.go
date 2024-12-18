// Package jwt jwt提供两个函数，用于对消息签名
package jwt

import (
	"crypto/ecdsa"
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotValidYet = errors.New("token is not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")
)

// CustomClaims The member variables in CustomClaims need to be capitalized because jwt will deserialize them later.
type CustomClaims struct {
	UserId uint64
	jwt.RegisteredClaims
}

// NewJWT generates a KeyManager for the given serverName.
// It saves the private key and the corresponding public key associated with the serverName.

// CreateToken creates a jwt by userid
func CreateToken(privateKey *ecdsa.PrivateKey, claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(privateKey)
}

// ParseToken parses a token by the public key of the corresponding service.
func ParseToken(publicKey *ecdsa.PublicKey, tokenString string) (*CustomClaims, error) {
	if publicKey == nil {
		return nil, ErrTokenInvalid
	}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, ErrTokenInvalid
		}
		return publicKey, nil
	})
	switch {
	case token == nil:
		return nil, jwt.ErrInvalidKey
	case errors.Is(err, jwt.ErrTokenMalformed):
		return nil, ErrTokenMalformed
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return nil, ErrTokenInvalid
	case errors.Is(err, jwt.ErrTokenExpired):
		return nil, ErrTokenExpired
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return nil, ErrTokenNotValidYet
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrTokenInvalid
}
