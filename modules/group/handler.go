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
}
