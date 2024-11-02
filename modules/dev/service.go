package dev

import (
	"database/sql"
	"enguete/util/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CheckValidJWT(c *gin.Context) {
	authHeader := c.GetHeader("bearer")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return
	}

	isValid, err := jwt.VerifyToken(authHeader)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "IDk "})
		return
	}
	if isValid {
		tokenStruct, err := jwt.DecodeBearer(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "IDk v2 "})
			return
		}
		c.JSON(http.StatusOK, tokenStruct)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "its a faulty jwt"})
	}
}

func GetAllUsers(c *gin.Context, db *sql.DB) {

}
