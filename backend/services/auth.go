package services

import (
	"errors"
	"time"

	"image-host/config"
	"image-host/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

var Auth = &AuthService{}

type Claims struct {
	UserID   uint   `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// HashPassword 生成密码哈希
func (s *AuthService) HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword 比对密码
func (s *AuthService) CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// GenerateToken 生成 JWT
func (s *AuthService) GenerateToken(u *models.User) (string, error) {
	secret := []byte(config.AppConfig.JWTSecret)
	exp := time.Now().Add(time.Duration(config.AppConfig.JWTExpireHours) * time.Hour)

	claims := &Claims{
		UserID:   u.ID,
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseToken 解析 JWT
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	secret := []byte(config.AppConfig.JWTSecret)

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
