package auth

import (
	"WireguardManager/pkg/slices"
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

//
// Roles
//

const (
	ServerManagerRole = "ServerManager"
	PeerManagerRole   = "PeerManager"
)

//
// Tool abstract
//

type AuthTool interface {
	IsEnabled() bool
	ValidateToken(signedToken string) (*JwtClaims, error)
	ValidateRoles(allowedRoles, receivedScopes []string) bool
}

type JwtClaims struct {
	jwt.StandardClaims
	Scopes []string `json:"scopes"`
}

//
// Errors
//

var (
	ErrInvalidToken  = fmt.Errorf("invalid token")
	ErrTokenExpired  = fmt.Errorf("token expired")
	ErrInvalidClaims = fmt.Errorf("invalid claims")
)

//
// Tool implemantation
//

type Tool struct {
	secretKeyHS256 *[]byte
	publicKeyRS256 *rsa.PublicKey
}

func NewJwtAuthTool() *Tool {
	tool := Tool{}
	return &tool
}

type KeysDeps struct {
	HS256SecretKey     *string
	RS256PublicKeyPath *string
}

func (t *Tool) LoadJwtKeys(deps KeysDeps) error {

	if deps.HS256SecretKey != nil {
		t.secretKeyHS256 = GetHmacSecretKey(*deps.HS256SecretKey)
	}

	if deps.RS256PublicKeyPath != nil {
		key, err := GetRsaPublicKey(*deps.RS256PublicKeyPath)
		if err != nil {
			return err
		}

		t.publicKeyRS256 = key
	}

	return nil
}

func (t *Tool) IsEnabled() bool {
	return !(t.publicKeyRS256 == nil && t.secretKeyHS256 == nil)
}

func (t *Tool) ValidateToken(signedToken string) (*JwtClaims, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtClaims{},
		func(token *jwt.Token) (interface{}, error) {

			if token.Method.Alg() == jwt.SigningMethodHS256.Name && t.secretKeyHS256 != nil {
				return *t.secretKeyHS256, nil
			}

			if token.Method.Alg() == jwt.SigningMethodRS256.Name && t.publicKeyRS256 != nil {
				return t.publicKeyRS256, nil
			}

			return nil, fmt.Errorf("unsupported alg type")
		},
	)
	if err != nil {
		logrus.Errorf("Jwt error: %s", err.Error())
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

func (t *Tool) ValidateRoles(allowedRoles, receivedRoles []string) bool {
	return slices.AnyCommon(allowedRoles, receivedRoles)
}

func GetHmacSecretKey(secretKey string) *[]byte {
	key := []byte(secretKey)
	return &key
}

func GetRsaPublicKey(filePath string) (*rsa.PublicKey, error) {
	pubKeyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}
