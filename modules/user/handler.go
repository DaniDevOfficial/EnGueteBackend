package user

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoute(router *gin.Engine, db *sql.DB) {
	registerUserRoutes(router, db)
	registerAuthRoutes(router, db)
}

func registerUserRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/users/:uuid", func(c *gin.Context) {
		GetUserByUUID(c, db)
	})
	router.DELETE("/users/", func(c *gin.Context) {
		DeleteUserWithJWT(c, db)
	})
	router.PUT("/users/name/", func(c *gin.Context) {
		UpdateUsername(c, db)
	})
	router.PUT("/users/password/", func(c *gin.Context) {
		UpdateUserPassword(c, db)
	})
}

func registerAuthRoutes(router *gin.Engine, db *sql.DB) {

	router.POST("/auth/signup", func(c *gin.Context) {
		CreateNewUser(c, db)
	})
	router.POST("/auth/signin", func(c *gin.Context) {
		SignIn(c, db)
	})

}