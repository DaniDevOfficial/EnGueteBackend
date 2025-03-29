package meal

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterMealRoute(router *gin.Engine, db *sql.DB) {
	registerMealRoutes(router, db)
	registerPreferenceRoutes(router, db)
	registerMealUpdateRoutes(router, db)
	registerMealCooksRoutes(router, db)
}

func registerMealRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/meals", func(c *gin.Context) {
		GetMealById(c, db)
	})
	router.POST("/meals/", func(c *gin.Context) {
		CreateNewMeal(c, db)
	})
	router.DELETE("/meals/:mealId", func(c *gin.Context) {
		DeleteMeal(c, db)
	})
	router.POST("/meals/open/", func(c *gin.Context) {
		ChangeMealClosedFlag(c, db)
	})
	router.POST("/meals/fulfilled", func(c *gin.Context) {
		ChangeMealFulfilledFlag(c, db)
	})

}

func registerPreferenceRoutes(router *gin.Engine, db *sql.DB) {
	router.PUT("/meals/preferences", func(c *gin.Context) {
		UpdatePreference(c, db)
	})
}

func registerMealCooksRoutes(router *gin.Engine, db *sql.DB) {
	router.POST("/meals/cooks", func(context *gin.Context) {
		AddCookToMeal(context, db)
	})
	router.DELETE("/meals/cooks", func(c *gin.Context) {
		RemoveCookFromMeal(c, db)
	})

}

func registerMealUpdateRoutes(router *gin.Engine, db *sql.DB) {

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
