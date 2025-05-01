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
	router.POST("/test/uuid", func(c *gin.Context) {
		ValidUUIDCheck(c)
	})
	router.GET("/test/dbnil", func(c *gin.Context) {
		DBNilCheck(c, db)
	})
}
