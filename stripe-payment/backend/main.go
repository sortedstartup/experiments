package main

import (
	"log"
	"stripe-payment/api"
	"stripe-payment/db"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	log.Println(".env loaded successfully")

	// Init DB
	if err := db.InitDB("file:payment.db?cache=shared&_fk=1"); err != nil {
		log.Fatalf("DB init error: %v", err)
	}

	// Start router
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(func(c *gin.Context) {
		log.Printf("%s %s\n", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})
	api.RegisterStripeRoutes(r)

	// Start server
	r.Run(":8080")
}
