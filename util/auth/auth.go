package auth

import (
	"database/sql"
	"enguete/util/jwt"
	"fmt"
	"github.com/gin-gonic/gin"
)

func GetJWTTokenFromHeader(c *gin.Context) (string, error) {
	jwtString := c.Request.Header.Get("Authorization")
	if jwtString == "" {
		return "", fmt.Errorf("missing authorization header")
	}
	return jwtString, nil
}

// GetJWTPayloadFromHeader extracts the JWT payload from the Authorization header of an HTTP request.
// It first retrieves the JWT token from the header, verifies the token, and then decodes the payload.
//
// Parameters:
//
//	r (*http.Request): The HTTP request containing the Authorization header with the JWT token.
//
// Returns:
//
//	(jwt.JWTPayload, error): Returns the decoded JWT payload if successful, otherwise returns an error.
func GetJWTPayloadFromHeader(c *gin.Context, db *sql.DB) (jwt.JWTPayload, error) {
	jwtToken, err := GetJWTTokenFromHeader(c)
	var jwtData jwt.JWTPayload
	if err != nil {
		jwtData, newJwtToken, err := CreateNewTokenWithRefreshToken(c, db)

		if err != nil {
			return jwtData, err
		}

		c.Header("Authorization", newJwtToken)
		return jwtData, err
	}

	valid, jwtData, err := jwt.VerifyToken(jwtToken)
	if err != nil {
		return jwtData, err
	}
	if !valid {
		jwtData, jwtToken, err = CreateNewTokenWithRefreshToken(c, db)

		if err != nil {
			return jwtData, err
		}

		c.Header("Authorization", jwtToken)
	}

	return jwtData, err
}

func GetRefreshTokenFromHeader(c *gin.Context) (string, error) {
	refreshToken := c.Request.Header.Get("RefreshToken")
	if refreshToken == "" {
		return "", fmt.Errorf("missing refresh token header")
	}

	return refreshToken, nil
}

func CreateNewTokenWithRefreshToken(c *gin.Context, db *sql.DB) (jwt.JWTPayload, string, error) {
	refreshToken, err := GetRefreshTokenFromHeader(c)
	var jwtData jwt.JWTPayload
	if err != nil {
		return jwtData, "", err
	}

	refreshTokenBody, err := jwt.VerifyRefreshToken(refreshToken, db)
	if err != nil {
		return jwtData, "", err
	}

	userData := jwt.JWTUser{
		UserId: refreshTokenBody.UserId,
	}

	jwtToken, err := jwt.CreateToken(userData)
	if err != nil {
		return jwtData, jwtToken, err
	}

	jwtData, err = jwt.DecodeBearer(jwtToken)
	return jwtData, jwtToken, err
}
