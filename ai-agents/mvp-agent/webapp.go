package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

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

		// Build the MVP
		if err := buildMVP(outputDir, logChannel); err != nil {
			fmt.Println("Error building MVP: %w", err)
			return
		}
		logChannel <- "MVP built successfully!"
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

func buildMVP(outputDir string, logChannel chan<- string) error {
	// Build directory
	logChannel <- fmt.Sprintf("Building MVP for all platforms")

	buildDir := filepath.Join(outputDir, "builds")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create builds directory: %v", err)
	}

	// Platforms to build for
	platforms := []struct {
		GOOS   string
		GOARCH string
	}{
		{"linux", "amd64"},
		{"darwin", "amd64"},
		{"windows", "amd64"},
	}

	var buildErrors []string

	// Build for each platform
	for _, platform := range platforms {
		logChannel <- fmt.Sprintf("Building for %s/%s...", platform.GOOS, platform.GOARCH)

		// Output filename
		output := fmt.Sprintf("mvp-%s-%s", platform.GOOS, platform.GOARCH)
		if platform.GOOS == "windows" {
			output += ".exe"
		}

		backendDir := filepath.Join(outputDir, "backend")
		absoluteBuildDir, _ := filepath.Abs(buildDir)
		outputPath := filepath.Join(absoluteBuildDir, output)

		// Build command
		cmd := exec.Command("go", "build", "-o", outputPath, ".")
		cmd.Dir = backendDir // Build from backend folder
		cmd.Env = append(os.Environ(),
			"GOOS="+platform.GOOS,
			"GOARCH="+platform.GOARCH,
		)

		// Log the command being run
		logChannel <- fmt.Sprintf("Running: go build -o %s . (from %s)", outputPath, backendDir)

		// Run build and capture output
		output_bytes, err := cmd.CombinedOutput()
		if err != nil {
			errorMsg := fmt.Sprintf(" Failed to build for %s/%s: %v\nOutput: %s", platform.GOOS, platform.GOARCH, err, string(output_bytes))
			logChannel <- errorMsg
			buildErrors = append(buildErrors, errorMsg)
			continue
		}

		logChannel <- fmt.Sprintf("Built: %s", output)
	}

	logChannel <- " All builds completed!"
	return nil
}
