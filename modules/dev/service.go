package dev

import (
	"database/sql"
	"enguete/util/jwt"
	"github.com/gin-gonic/gin"
	"log"
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

func ValidUUIDCheck(c *gin.Context) {
	var requestData struct {
		UUID string `json:"uuid" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "UUID is valid"})
}

func DBNilCheck(c *gin.Context, db *sql.DB) {
	query := `
SELECT user_id
FROM users u
    	WHERE $1::text IS NULL

`

	rows, err := db.Query(query, nil)
	if err != nil {
		log.Println("Error executing query:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		err := rows.Scan(
			&id,
		)
		if err != nil {
			log.Println("Error scanning row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}
		ids = append(ids, id)
	}
	c.JSON(http.StatusOK, gin.H{"message": "Query executed successfully", "ids": ids})
}
