package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var (
	Key     = "l08mCML6VZwXZ4Rk"
	Expires = 24
)

type User struct {
	UId      int64  `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(uid int64, username string) (string, error) {
	claims := User{
		uid,
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(Expires) * time.Hour)), // 过期时间24小时
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                         // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                         // 生效时间
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用HS256签名算法
	tokenString, err := token.SignedString([]byte(Key))

	return tokenString, err
}

func ParseToken(tokenString string) (*User, error) {
	if tokenString == "" {
		return nil, errors.New("token不能为空")
	}
	token, err := jwt.ParseWithClaims(tokenString, &User{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(Key), nil
	})

	if claims, ok := token.Claims.(*User); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
