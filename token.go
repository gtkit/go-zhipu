// @Author xiaozhaofu 2023/8/15 11:02:00
package openai

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type ZpClaims struct {
	APIKey    string `json:"api_key"`
	Exp       int64  `json:"exp"`
	Timestamp int64  `json:"timestamp"`
}

// GetApiToken 生成一个token.
func GetAPIToken(apiKey string, duration time.Duration) (string, error) {
	apiKeyInfo := strings.Split(apiKey, ".")
	key, secret := apiKeyInfo[0], apiKeyInfo[1]

	claims := ZpClaims{
		key,
		time.Now().Add(duration).Unix(),
		time.Now().Unix(),
	}

	return CreateToken(claims, secret)
}

// CreateToken 生成一个token.
func CreateToken(claims ZpClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"api_key":   claims.APIKey,
		"exp":       claims.Exp,
		"timestamp": claims.Timestamp,
	})

	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"
	res, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return res, nil
}
