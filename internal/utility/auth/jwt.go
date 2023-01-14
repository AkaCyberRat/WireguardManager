package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Scopes
const (
	ServerManager = "ServerManager"
	PeerManager   = "PeerManager"
)

const (
	HMACSecretKey = "secret"
)

type JwtClaims struct {
	Scopes []string `json:"scopes"`
	jwt.StandardClaims
}

func ValidateToken(signedToken string) (*JwtClaims, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtClaims{},
		func(token *jwt.Token) (interface{}, error) {

			if token.Method.Alg() != jwt.SigningMethodHS256.Name {
				return nil, fmt.Errorf("unsupported alg type")
			}

			return []byte(HMACSecretKey), nil
		},
	)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JwtClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, ErrTokenExpired
	}

	return claims, nil
}

var (
	ErrInvalidToken  = fmt.Errorf("invalid token")
	ErrTokenExpired  = fmt.Errorf("token expired")
	ErrInvalidClaims = fmt.Errorf("invalid claims")
)
