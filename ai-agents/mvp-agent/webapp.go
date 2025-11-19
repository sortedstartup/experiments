package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

// setupRoutes configures all the routes for the application
func setupRoutes(e *echo.Echo) {
	// Serve the index.html file at root
	e.GET("/", serveIndex)

	// Handle MVP generation
	e.POST("/generate-mvp", generateMVP)
}

// serveIndex serves the index.html file
func serveIndex(c echo.Context) error {
	return c.File("ui/index.html")
}

// generateMVP handles the MVP generation request
func generateMVP(c echo.Context) error {
	// Get user input from form
	userInput := c.FormValue("user_input")
	if userInput == "" {
		return c.HTML(http.StatusBadRequest, `<div class="text-red-500 p-4">Please provide a valid input</div>`)
	}

	// Create a temporary requirements file
	requirementsFile, err := createRequirementsFile(userInput)
	if err != nil {
		log.Printf("Error creating requirements file: %v", err)
		return c.HTML(http.StatusInternalServerError, `<div class="text-red-500 p-4">Error creating requirements file</div>`)
	}

	// Copy starter template to timestamped output directory
	outputDir, err := copyStarterTemplate()
	if err != nil {
		log.Printf("Error copying starter template: %v", err)
		return c.HTML(http.StatusInternalServerError, `<div class="text-red-500 p-4">Error copying starter template</div>`)
	}

	// Copy requirements file to output directory
	if err := copyPRDToOutput(requirementsFile, outputDir); err != nil {
		log.Printf("Error copying requirements file: %v", err)
		return c.HTML(http.StatusInternalServerError, `<div class="text-red-500 p-4">Error copying requirements file</div>`)
	}

	// Run the MVP generation agent
	result, err := runMVPAgent(outputDir)
	if err != nil {
		log.Printf("Error running MVP agent: %v", err)
		return c.HTML(http.StatusInternalServerError, fmt.Sprintf(`<div class="text-red-500 p-4">Error generating MVP: %v</div>`, err))
	}

	// Clean up temporary requirements file
	os.Remove(requirementsFile)

	// Return success response with output directory
	return c.HTML(http.StatusOK, fmt.Sprintf(`
		<div class="p-4 bg-green-100 border border-green-400 rounded">
			<h3 class="text-lg font-bold text-green-800">MVP Generated Successfully!</h3>
			<p class="text-green-700 mt-2">Output Directory: %s</p>
			<p class="text-green-700">Status: %s</p>
		</div>
	`, outputDir, result))
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

// runMVPAgent runs the MVP generation agent using the refactored function from agent.go
func runMVPAgent(outputDir string) (string, error) {
	ctx := context.Background()

	// Use the refactored function from agent.go
	if err := RunMVPAgentInDirectory(ctx, outputDir); err != nil {
		return "", err
	}

	return "Agent processing completed successfully", nil
}
