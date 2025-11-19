package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

// Global channel for streaming logs
var logChannel chan string

// setupRoutes configures all the routes for the application
func setupRoutes(e *echo.Echo) {
	// Serve the index.html file at root
	e.GET("/", serveIndex)

	// Handle MVP generation

	// SSE endpoint for streaming logs
	e.GET("/generate-mvp", generateMVP)
}

// serveIndex serves the index.html file
func serveIndex(c echo.Context) error {
	return c.File("ui/index.html")
}

func generateMVP(c echo.Context) error {
	// Get user input from query params
	userInput := c.QueryParam("user_input")
	if userInput == "" {
		return c.String(http.StatusBadRequest, "Please provide user_input parameter")
	}

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")

	// Initialize log channel
	logChannel = make(chan string, 100)

	// Start MVP generation in goroutine
	go func() {
		defer close(logChannel)

		// Create requirements file
		requirementsFile, err := createRequirementsFile(userInput)
		if err != nil {
			fmt.Println("Error creating requirements file: %w", err)
			return
		}

		// Copy starter template
		outputDir, err := copyStarterTemplate()
		if err != nil {
			fmt.Println("Error copying starter template: %w", err)
			return
		}
		logChannel <- fmt.Sprintf("✅ Copied starter template to: %s", outputDir)

		// Copy requirements file
		if err := copyPRDToOutput(requirementsFile, outputDir); err != nil {
			fmt.Println("Error copying requirements file: %w", err)
			return
		}
		logChannel <- "Copied requirements file to output directory"

		// Run agent
		ctx := context.Background()
		if err := RunMVPAgent(ctx, outputDir, logChannel); err != nil {
			logChannel <- fmt.Sprintf("Error running MVP agent: %v", err)
			return
		}

		// Clean up
		os.Remove(requirementsFile)
		logChannel <- "MVP generation completed successfully!"
	}()

	// Stream logs to client
	for logMsg := range logChannel {
		fmt.Fprintf(c.Response(), "data: %s\n\n", logMsg)
		c.Response().Flush()
	}

	return nil
}

// createRequirementsFile creates a temporary file with user requirements
func createRequirementsFile(userInput string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "requirements_*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()

	// Write user input to the file
	_, err = tmpFile.WriteString(userInput)
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %v", err)
	}

	return tmpFile.Name(), nil
}
