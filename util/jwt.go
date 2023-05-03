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

// GenerateToken
// token expireTime : not set, do this with redis
func GenerateToken(username string) (string, error) {
	claims := jwt.RegisteredClaims{
		// expires at 3 hours
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 3)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "sast",
		Subject:   username,
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSigningKey)

	return token, err
}

func ParseToken(token string) (*jwt.RegisteredClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSigningKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*jwt.RegisteredClaims); ok && tokenClaims.Valid {
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

func GetUsername(token string) (string, error) {
	claims, err := ParseToken(token)
	if err != nil {
		return "", err
	}
	username, claimsError := claims.GetSubject()
	if claimsError != nil {
		return "", claimsError
	}
	return username, nil
}
