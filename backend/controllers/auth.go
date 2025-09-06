package controllers

import (
	"fmt"
	"net/http"
	"time"

	"image-host/config"
	"image-host/database"
	"image-host/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct{}

var Auth = &AuthController{}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type guestLoginRequest struct {
	Code string `json:"code"`
}

type changePwdRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (a *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credentials"})
		return
	}
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username or password incorrect"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username or password incorrect"})
		return
	}

	exp := time.Now().Add(time.Duration(config.AppConfig.JWTExpireHours) * time.Hour)
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(exp),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      claims.ExpiresAt.Unix(),
		"iat":      claims.IssuedAt.Unix(),
	})
	tokenStr, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":    tokenStr,
			"username": user.Username,
			"expires":  exp.Unix(),
		},
	})
}

func (a *AuthController) Me(c *gin.Context) {
	username := c.GetString("username")
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"username": username}})
}

// GuestLogin 游客码登录，返回用户名形如 guest:<id>
func (a *AuthController) GuestLogin(c *gin.Context) {
	var req guestLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code"})
		return
	}

	var gc models.GuestCode
	// 允许永久（ExpiresAt 为空）或未过期
	if err := database.DB.Where("code = ?", req.Code).First(&gc).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid code"})
		return
	}
	if gc.ExpiresAt != nil && time.Now().After(*gc.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Code expired"})
		return
	}

	// JWT：username = guest:<id>
	exp := time.Now().Add(time.Duration(config.AppConfig.JWTExpireHours) * time.Hour)
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(exp),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "guest:" + fmt.Sprintf("%d", gc.ID),
		"exp":      claims.ExpiresAt.Unix(),
		"iat":      claims.IssuedAt.Unix(),
	})
	tokenStr, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":    tokenStr,
			"username": "guest",
			"expires":  exp.Unix(),
		},
	})
}

func (a *AuthController) ChangePassword(c *gin.Context) {
	var req changePwdRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.OldPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}
	username := c.GetString("username")
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Old password incorrect"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set new password"})
		return
	}
	user.PasswordHash = string(hash)
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save new password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
