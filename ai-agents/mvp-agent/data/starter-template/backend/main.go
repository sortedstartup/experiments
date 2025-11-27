package main

import (
	"log"
	"os"
)

func main() {
	// Get port from environment or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize and start the web server
	app := NewApp()

	log.Printf("🚀 Server starting on http://localhost:%s", port)
	if err := app.Start(":" + port); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
