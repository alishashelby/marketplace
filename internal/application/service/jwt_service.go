package service

import (
	"errors"
	"fmt"
	"github.com/alishashelby/marketplace/internal/domain/entity"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strconv"
	"time"
)

type ctxKey string

const (
	UserIDKey   ctxKey = "id"
	UserKey     ctxKey = "user"
	UsernameKey ctxKey = "username"
	IssuedAtKey ctxKey = "iat"
	ExpiryKey   ctxKey = "exp"

	dotEnvJWTSecret     = "JWT_SECRET"
	dotEnvJWTExpiration = "JWT_TTL"
)

var (
	errorLoadingSecret  = errors.New("error loading JWT_SECRET environment variable")
	errorLoadingTTL     = errors.New("error loading JWT_TTL environment variable")
	errorParsingTTL     = errors.New("error parsing JWT_TTL environment variable")
	errorNotPositiveTTL = errors.New("JWT_TTL environment variable should be positive")
)

type JWTService struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTService() (*JWTService, error) {
	secret := []byte(os.Getenv(dotEnvJWTSecret))
	if secret == nil {
		return nil, errorLoadingSecret
	}

	ttl := os.Getenv(dotEnvJWTExpiration)
	if ttl == "" {
		return nil, errorLoadingTTL
	}

	ttlInSeconds, err := strconv.Atoi(ttl)
	if err != nil {
		return nil, errorParsingTTL
	}
	if ttlInSeconds < 0 {
		return nil, errorNotPositiveTTL
	}

	return &JWTService{
		secret: secret,
		ttl:    time.Duration(ttlInSeconds) * time.Second,
	}, nil
}

func (s *JWTService) GenerateToken(user *entity.User) (*string, error) {
	claims := jwt.MapClaims{
		string(UserKey): map[string]interface{}{
			string(UsernameKey): user.Username,
			string(UserIDKey):   user.ID.String(),
		},
		string(IssuedAtKey): time.Now().Unix(),
		string(ExpiryKey):   time.Now().Add(s.ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	return &signedToken, nil
}

func (s *JWTService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	return payload, nil
}
