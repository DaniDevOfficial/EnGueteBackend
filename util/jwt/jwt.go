package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("capybara") // TODO: add secret key via .env or some rotation

func CreateToken(userData JWTUser) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"UserId":   userData.UserId,
			"UserName": userData.Username,
			"Exp":      time.Now().Add(time.Hour * 24).Unix(),
		})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, nil
	}
	return true, nil
}

func DecodeBearer(tokenString string) (JWTPayload, error) {
	splitToken := strings.Split(tokenString, ".")
	if len(splitToken) != 3 {
		return JWTPayload{}, fmt.Errorf("invalid token format")
	}

	payloadSegment := splitToken[1]
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadSegment)
	if err != nil {
		return JWTPayload{}, fmt.Errorf("failed to decode payload: %v", err)
	}

	var payload JWTPayload
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return JWTPayload{}, fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	return payload, nil
}
