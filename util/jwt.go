package util

// util
//the basic configuration of JWT
import (
	"time"

	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/NJUPT-SAST/sast-link-backend/config"
)

var jwtSigningKey = config.Config.GetString("jwt.signingKey")

type CustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken
// token expireTime : not set, do this with redis

func GenerateToken(username string) (string, error) {
	claims := CustomClaims{
		username,
		jwt.RegisteredClaims{
			// expires at 3 hours
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "sast",
			Subject:   "user",
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSigningKey)

	return token, err
}

func ParseToken(token string) (*CustomClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSigningKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*CustomClaims); ok && tokenClaims.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("token parse error")
	}
}

func RefreshToken(token string) (string, error) {
	claims, err := ParseToken(token)
	if err != nil {
		return "", err
	}

	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour * 3))
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tokenClaims.SignedString(jwtSigningKey)

	return token, err
}
