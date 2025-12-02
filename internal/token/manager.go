package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

var (
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type Manager interface {
	GenerateAccessToken(userID int) (string, error)
	ParseAccessToken(accessToken string) (*Claims, error)
	GenerateRefreshTokenPlain() (string, error)
	HashToken(plainToken string) (string, error)
}

type Claims struct {
	UserId int `json:"uid"`
	jwt.RegisteredClaims
}
type TokenManager struct {
	jwtSecret []byte
	accessTTL time.Duration
}

func NewManager(secret []byte, accessTTL time.Duration) TokenManager {
	return TokenManager{
		jwtSecret: secret,
		accessTTL: accessTTL,
	}
}

func (tm TokenManager) GenerateAccessToken(userID int) (string, error) {
	now := time.Now()
	claims := Claims{
		UserId: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.accessTTL)),
			Subject:   strconv.Itoa(userID),
			Issuer:    "auth-service",
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return t.SignedString(tm.jwtSecret)
}

func (tm TokenManager) ParseAccessToken(accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidAccessToken
			}
			return tm.jwtSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return nil, ErrInvalidAccessToken
	}

	return claims, nil
}

func (tm TokenManager) GenerateRefreshTokenPlain() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (tm TokenManager) HashToken(plainToken string) (string, error) {
	h := sha256.Sum256([]byte(plainToken))
	return hex.EncodeToString(h[:]), nil
}
