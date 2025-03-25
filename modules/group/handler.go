package group

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterGroupRoute(router *gin.Engine, db *sql.DB) {
	registerGroupRoutes(router, db)
	registerInviteTokenRoutes(router, db)
}

func registerGroupRoutes(router *gin.Engine, db *sql.DB) {
	router.POST("/groups", func(c *gin.Context) {
		CreateNewGroup(c, db)
	})
	router.GET("/groups/:groupId/members", func(c *gin.Context) {
		GetGroupMembers(c, db)
	})
	router.GET("/groups", func(c *gin.Context) {
		GetGroupById(c, db)
	})
	router.DELETE("/groups/:groupId/leave", func(c *gin.Context) {
		LeaveGroup(c, db)
	})
}

func registerInviteTokenRoutes(router *gin.Engine, db *sql.DB) {
	router.POST("/groups/invite/", func(c *gin.Context) {
		GenerateInviteLink(c, db)
	})
	router.DELETE("/groups/invite/:inviteToken", func(c *gin.Context) {
		VoidInviteToken(c, db)
	})
	router.POST("/groups/invite/join/:inviteToken", func(c *gin.Context) {
		JoinGroupWithInviteToken(c, db)
	})
}
