package meal

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func RegisterMealRoute(router *gin.Engine, db *sql.DB) {
	registerMealRoutes(router, db)
	registerPreferenceRoutes(router, db)
	registerMealUpdateRoutes(router, db)
	registerSyncRoutes(router, db)
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

func registerSyncRoutes(router *gin.Engine, db *sql.DB) {
	router.GET("/sync/group/meals", func(c *gin.Context) {
		SyncGroupMeals(c, db)
	})
	router.GET("/sync/group/meal", func(c *gin.Context) {
		SyncMealInformation(c, db)
	})
}
