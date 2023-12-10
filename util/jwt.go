package util

// util
//the basic configuration of JWT
import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
)

var jwtSigningKey = config.Config.Sub("jwt").GetString("signing_key")

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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signingKey := []byte(jwtSigningKey)
	signToken, err := token.SignedString(signingKey)
	return signToken, err
}

// GenerateToken with expireTime
// identifier is something like `username-loginTicket` or `oauthIdentity-oauthLarkToken`
func GenerateTokenWithExp(ctx context.Context, identifier string, expireTime time.Duration) (string, error) {
	signingKey := []byte(jwtSigningKey)
	gen := NewJWTAccessGenerate("", signingKey, jwt.SigningMethodHS256)
	access, _, err := gen.Token(ctx, identifier, expireTime, false)
	return access, err
}

func ParseToken(token string) (*JWTAccessClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &JWTAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, result.AuthParseTokenFail
		}
		return []byte(jwtSigningKey), nil
	})
	if err != nil {
		return nil, result.AuthParseTokenFail.Wrap(err)
	}

	if claims, ok := tokenClaims.Claims.(*JWTAccessClaims); ok && tokenClaims.Valid {
		return claims, nil
	} else {
		return nil, result.AuthParseTokenFail.Wrap(err)
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

// TokenAudience get `Audience` field(information about user/oauth...) from claims
func TokenAudience(token string) (audience []string, err error){
	claims, err := ParseToken(token)
	if err != nil {
		return
	}
	if err = claims.Valid(); err != nil {
		return
	}

	return claims.GetAudience()
}

// IdentityFromToken return identity(now "username"/"union_id")
//
// flag: verify token type
func IdentityFromToken(token, flag string) (string, error) {
	audience, err := TokenAudience(token)
	identifier := strings.Split(audience[0], "-")
	if err != nil {
		return "", err
	}
	identity, err := identifier[0], nil
	if identifier[1] != "" && flag != "" && flag != identifier[1] {
		return "", result.TicketNotCorrect
	}
	return strings.ToLower(identity), err
}

// JWTAccessClaims jwt claims
type JWTAccessClaims struct {
	jwt.RegisteredClaims
}

// JWTAccessGenerate generate the jwt access token
type JWTAccessGenerate struct {
	SignedKeyID  string
	SignedKey    []byte
	SignedMethod jwt.SigningMethod
}

func (a *JWTAccessClaims) Valid() error {
	if time.Unix(a.ExpiresAt.Unix(), 0).Before(time.Now()) {
		return result.InvalidAccToken
	}
	return nil
}

// NewJWTAccessGenerate create to generate the jwt access token instance
func NewJWTAccessGenerate(kid string, key []byte, method jwt.SigningMethod) *JWTAccessGenerate {
	return &JWTAccessGenerate{
		SignedKeyID:  kid,
		SignedKey:    key,
		SignedMethod: method,
	}
}

// Token based on the UUID generate the jwt access token
func (a *JWTAccessGenerate) Token(ctx context.Context, username string, expireTime time.Duration, isGenRenfresh bool) (string, string, error) {
	claims := &JWTAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sast",
			Audience:  jwt.ClaimStrings{username},
		},
	}

	token := jwt.NewWithClaims(a.SignedMethod, claims)
	if a.SignedKeyID != "" {
		token.Header["kid"] = a.SignedKeyID
	}
	var key interface{}
	if a.isEs() {
		v, err := jwt.ParseECPrivateKeyFromPEM(a.SignedKey)
		if err != nil {
			return "", "", err
		}
		key = v
	} else if a.isRsOrPS() {
		v, err := jwt.ParseRSAPrivateKeyFromPEM(a.SignedKey)
		if err != nil {
			return "", "", err
		}
		key = v
	} else if a.isHs() {
		key = a.SignedKey
	} else if a.isEd() {
		v, err := jwt.ParseEdPrivateKeyFromPEM(a.SignedKey)
		if err != nil {
			return "", "", err
		}
		key = v
	} else {
		return "", "", errors.New("unsupported sign method")
	}
	accessToken, err := token.SignedString(key)
	if err != nil {
		return "", "", err
	}
	refresh := ""

	if isGenRenfresh {
		t := uuid.NewSHA1(uuid.Must(uuid.NewRandom()), []byte(accessToken)).String()
		refresh = base64.URLEncoding.EncodeToString([]byte(t))
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}

	return accessToken, refresh, nil
}

func (a *JWTAccessGenerate) isEs() bool {
	return strings.HasPrefix(a.SignedMethod.Alg(), "ES")
}

func (a *JWTAccessGenerate) isRsOrPS() bool {
	isRs := strings.HasPrefix(a.SignedMethod.Alg(), "RS")
	isPs := strings.HasPrefix(a.SignedMethod.Alg(), "PS")
	return isRs || isPs
}

func (a *JWTAccessGenerate) isHs() bool {
	return strings.HasPrefix(a.SignedMethod.Alg(), "HS")
}

func (a *JWTAccessGenerate) isEd() bool {
	return strings.HasPrefix(a.SignedMethod.Alg(), "Ed")
}
