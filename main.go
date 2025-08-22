package main

import (
	"enguete/modules/dev"
	"enguete/modules/group"
	"enguete/modules/management"
	"enguete/modules/meal"
	"enguete/modules/user"
	"enguete/util/db"
	"enguete/util/validator"
	"github.com/joho/godotenv"
	"os"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠️ No .env file found – continuing without it")
		}
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}
	dbConnection := db.InitDB(dbURL)

	validator.InitCustomValidators()

	router := gin.Default()
	router.Use(corsMiddleware())

	dev.RegisterDevRoutes(router, dbConnection)
	user.RegisterUserRoute(router, dbConnection)
	group.RegisterGroupRoute(router, dbConnection)
	meal.RegisterMealRoute(router, dbConnection)
	management.RegisterManagementRoute(router, dbConnection)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	log.Printf("🚀 Server is listening on http://localhost:%s/", port)
	log.Fatal(router.Run("0.0.0.0:" + port))
}

// corsMiddleware sets the CORS headers to allow all origins.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, RefreshToken")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			log.Println("Options request handled")
			return
		}

		log.Println("New Request Started")
		log.Printf("Method: %s, Path: %s\n", c.Request.Method, c.Request.URL.Path)

		c.Next()
	}
}
