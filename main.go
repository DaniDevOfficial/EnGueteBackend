package main

import (
	_ "enguete/docs"
	"enguete/util/auth"
	"fmt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// @title EnGuete API
// @version 1.0
// @description This is the API for EnGuete application.
// @host localhost:8000
// @BasePath /
// @schemes http
func main() {
	// dbConnection := db.InitDB()

	router := gin.Default()
	router.GET("/documentation/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Use(corsMiddleware())
	/*
		user.RegisterUserRoute(router, dbConnection)
		dev.RegisterDevRoutes(router, dbConnection)
		group.RegisterGroupRoute(router, dbConnection)
	*/
	fmt.Println("🚀 Server is listening on http://localhost:8000/")
	log.Fatal(router.Run("localhost:8000"))
}

// corsMiddleware sets the CORS headers to allow all origins.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		log.Println("New Request Started")
		jwt, err := auth.GetJWTPayloadFromHeader(c)
		if err == nil {
			log.Printf("Request made by: %v\n", jwt)
		} else {
			log.Println("Request made by: Anonymous")
		}
		log.Printf("Method: %s, Path: %s\n", c.Request.Method, c.Request.URL.Path)

		c.Next()
	}
}
