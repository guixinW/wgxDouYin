package jwt

import (
	"crypto/ecdsa"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var (
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotValidYet = errors.New("token is not active yet")
	ErrTokenMalformed   = errors.New("that's not even a token")
	ErrTokenInvalid     = errors.New("couldn't handle this token")
)

type JWT struct {
	serverToPublicKey map[string]*ecdsa.PublicKey
	serverPrivateKey  *ecdsa.PrivateKey
}

// CustomClaims The member variables in CustomClaims need to be capitalized because jwt will deserialize them later.
type CustomClaims struct {
	UserId uint64
	jwt.RegisteredClaims
}

// NewJWT generates a JWT for the given serverName.
// It saves the private key and the corresponding public key associated with the serverName.
func NewJWT(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey, serverName string) *JWT {
	return &JWT{serverPrivateKey: privateKey, serverToPublicKey: map[string]*ecdsa.PublicKey{serverName: publicKey}}
}

// getServerPublicKey retrieves the public key associated with specified service.
func (j *JWT) getServerPublicKey(server string) (*ecdsa.PublicKey, error) {
	serverPublicKey, ok := j.serverToPublicKey[server]
	if !ok {
		return nil, errors.New("can't find server's public key")
	}
	return serverPublicKey, nil
}

// addServerPublicKey saves the public key associated with specified service.
func (j *JWT) addServerPublicKey(serverName string, serverPublicKey *ecdsa.PublicKey) error {
	if j == nil {
		return errors.New("JWT is nil object")
	}
	j.serverToPublicKey[serverName] = serverPublicKey
	return nil
}

// CreateToken creates a jwt by userid
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(j.serverPrivateKey)
}

// ParseToken parses a token by the public key of the corresponding service.
func (j *JWT) ParseToken(tokenString, serverName string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, ErrTokenInvalid
		}
		if serverPublicKey, err := j.getServerPublicKey(serverName); err != nil {
			return nil, err
		} else {
			return serverPublicKey, nil
		}
	})
	//fmt.Println(token)
	switch {
	case token == nil:
		return nil, ErrTokenInvalid
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

func TransferTimeToJwtTime(old time.Time) *jwt.NumericDate {
	return jwt.NewNumericDate(old)
}
