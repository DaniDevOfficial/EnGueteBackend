package group

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterGroupRoute(router *gin.Engine, db *sql.DB) {
	registerGroupRoutes(router, db)
}

func registerGroupRoutes(router *gin.Engine, db *sql.DB) {
	router.POST("/groups", func(c *gin.Context) {
		CreateNewGroup(c, db)
	})
	router.POST("/groups/invite/", func(c *gin.Context) {
		GenerateInviteLink(c, db)
	})
	router.GET("/groups/:id", func(c *gin.Context) {
		GetGroupById(c, db)
	})
}
