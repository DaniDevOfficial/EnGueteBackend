package meal

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterMealRoute(router *gin.Engine, db *sql.DB) {
	registerMealRoutes(router, db)
}

func registerMealRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/meals/:id", func(c *gin.Context) {
		// GetMealById(c, db)
	})
	router.POST("/meals/", func(c *gin.Context) {
		CreateNewMeal(c, db)
	})
	router.POST("/meals/cooks", func(context *gin.Context) {
		AddCookToMeal(context, db)
	})
	router.DELETE("/meals/cooks", func(c *gin.Context) {
		RemoveCookFromMeal(c, db)
	})
	router.PUT("/meals/title", func(c *gin.Context) {
		UpdateMealTitle(c, db)
	})
	router.PUT("/meals/type", func(c *gin.Context) {
		UpdateMealType(c, db)
	})
	router.PUT("/meals/note", func(c *gin.Context) {
		UpdateMealNotes(c, db)
	})
	router.PUT("/meals/scheduledAt", func(c *gin.Context) {
		UpdateMealScheduledAt(c, db)
	})
}
