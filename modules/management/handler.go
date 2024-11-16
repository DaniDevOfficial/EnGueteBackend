package management

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterManagementRoute(router *gin.Engine, db *sql.DB) {
	registerUserManagementRoutes(router, db)
	registerRoleManagementRoutes(router, db)
}

func registerUserManagementRoutes(router *gin.Engine, db *sql.DB) {
	router.DELETE("management/user/kick", func(c *gin.Context) {
		KickUserFromGroup(c, db)
	})
	router.DELETE("management/user/ban", func(c *gin.Context) {
		BanUserFromGroup(c, db)
	})
	router.DELETE("management/user/unban", func(c *gin.Context) {
		UnbanUserFromGroup(c, db)
	})
}

func registerRoleManagementRoutes(router *gin.Engine, db *sql.DB) {
	router.POST("management/roles/add", func(c *gin.Context) {
		AddRoleToUser(c, db)
	})
	router.POST("management/roles/remove", func(c *gin.Context) {
		RemoveRoleFromUser(c, db)
	})
}
