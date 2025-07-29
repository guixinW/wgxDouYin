package service

import "wgxDouYin/pkg/jwt"

// TokenService is the interface for token generation and validation.
type TokenService interface {
	CreateToken(claims jwt.CustomClaims) (string, error)
}
