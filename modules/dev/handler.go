package dev

import (
	"github.com/gin-gonic/gin"
)

func RegisterDevRoutes(router *gin.Engine) {
	registerTestRoutes(router)
}
func registerTestRoutes(router *gin.Engine) {

	router.POST("/test/jwtAuth", func(c *gin.Context) {
		CheckValidJWT(c)
	})
	router.GET("/test/getAllUsers", func(c *gin.Context) {
		CheckValidJWT(c)
	})
	router.POST("/test/uuid", func(c *gin.Context) {
		ValidUUIDCheck(c)
	})
}
