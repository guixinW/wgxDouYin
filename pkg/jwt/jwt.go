package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	ID int64
	jwt.Claims
}

var (
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotValidYet = errors.New("token is not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")
)
