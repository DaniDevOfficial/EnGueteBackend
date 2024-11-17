package auth

import (
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
func GetJWTPayloadFromHeader(c *gin.Context) (jwt.JWTPayload, error) {
	jwtToken, err := GetJWTTokenFromHeader(c)
	var jwtData jwt.JWTPayload
	if err != nil {
		return jwtData, err
	}
	valid, err := jwt.VerifyToken(jwtToken)
	if !valid {
		return jwtData, fmt.Errorf("jwt token is not valid")
	}
	if err != nil {
		return jwtData, err
	}
	jwtData, err = jwt.DecodeBearer(jwtToken)
	if err != nil {
		return jwtData, err
	}
	return jwtData, err
}
