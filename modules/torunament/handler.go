package tournament

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterTorunamentRoute(router *gin.Engine, db *sql.DB) {
	registerTournamentCRUDRoutes(router, db)
}

func registerTournamentCRUDRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/tournament/:uuid", func(c *gin.Context) {
		// GetTournamentByUUID(c, db)
	})
	router.DELETE("/tournament/:uuid", func(c *gin.Context) {
		// DeleteUserWithJWT(c, db)
	})
	router.PUT("/tournament/general/:uuid", func(c *gin.Context) {
		// UpdateUsername(c, db)
	})

}
