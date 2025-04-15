// utils/token.go
package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("your_secret_key")

// GenerateToken 根据用户ID和角色生成一个 JWT，过期时间为 24 小时
func GenerateToken(userID int, role string) (string, error) {
	// 使用 jwt.MapClaims 构造 claims
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
