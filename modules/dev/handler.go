package dev

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterDevRoutes(router *gin.Engine, db *sql.DB) {
	registerTestRoutes(router, db)
}
func registerTestRoutes(router *gin.Engine, db *sql.DB) {

	router.POST("/test/jwtAuth", func(c *gin.Context) {
		CheckValidJWT(c)
	})
	router.GET("/test/getAllUsers", func(c *gin.Context) {
		CheckValidJWT(c)
	})
}
