package group

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterGroupRoute(router *gin.Engine, db *sql.DB) {
	registerGroupRoutes(router, db)
	registerInviteTokenRoutes(router, db)
	registerSyncRoutes(router, db)
}

func registerGroupRoutes(router *gin.Engine, db *sql.DB) {
	router.POST("/groups", func(c *gin.Context) {
		CreateNewGroup(c, db)
	})
	router.PUT("/groups/name", func(c *gin.Context) {
		UpdateGroupName(c, db)
	})
	router.GET("/groups/members", func(c *gin.Context) {
		GetGroupMembers(c, db)
	})
	router.GET("/groups", func(c *gin.Context) {
		GetGroupById(c, db)
	})
	router.GET("/groups/meals", func(c *gin.Context) {
		GetGroupMeals(c, db)
	})
	router.DELETE("/groups/", func(c *gin.Context) {
		DeleteGroup(c, db)
	})
	router.DELETE("/groups/leave", func(c *gin.Context) {
		LeaveGroup(c, db)
	})
}

func registerInviteTokenRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/groups/invite/", func(c *gin.Context) {
		GetAllInviteTokensInAGroup(c, db)
	})
	router.POST("/groups/invite/", func(c *gin.Context) {
		GenerateInviteLink(c, db)
	})
	router.DELETE("/groups/invite/", func(c *gin.Context) {
		VoidInviteToken(c, db)
	})
	router.POST("/groups/invite/join/", func(c *gin.Context) {
		JoinGroupWithInviteToken(c, db)
	})
}

func registerSyncRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/sync/groups", func(c *gin.Context) {
		SyncAllGroups(c, db)
	})
	router.GET("/sync/group", func(c *gin.Context) {
		SyncSpecificGroup(c, db)
	})
}
