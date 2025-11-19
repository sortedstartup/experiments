package main

import (
	"bufio" // For GrepFile and SedTool
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"  // For GrepFile and SedTool
	"strings" // For SedTool (strings.Join)
	"time"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
)

// The agent instruction guides the LLM on its goal and tool usage strategy.
const agentInstruction = `
You are the **MVP Code Modification Agent**. Your task is to modify code files based on the user's MVP request.
First Read the prdvided MVP request and code to understand the required changes. 
Identify APIs, functions and UI Components to be added, changed, or removed.
Determine which files need to be modified in the Go backend (Echo framework) and HTML/HTMX frontend.
Then, use the provided tools to make precise modifications to the codebase.

**Context:**
- The user provides a working directory path containing Go backend (Echo framework) and HTML/HTMX frontend files
- Typical structure: [WORKING_DIR]/backend/main.go, [WORKING_DIR]/backend/webapp.go, [WORKING_DIR]/backend/ui/index.html

**Your Process:**
1. **Understand:** Use GrepFile to locate relevant code sections
2. **Modify:** Use SedTool for line-level changes or WriteFile for complete rewrites
3. **Verify:** Use ReadFile sparingly to confirm critical modifications
`

func main() {
	ctx := context.Background()

	// Check required environment variables
	if os.Getenv("GOOGLE_API_KEY") == "" {
		log.Fatalf("GOOGLE_API_KEY environment variable is required")
	}

	// Treat the command line argument as the MVP Request
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <MVP-Request-Description>", os.Args[0])
	}
	mvpRequest := os.Args[1]

	// Copy starter template to timestamped output directory
	outputDir, err := copyStarterTemplate()
	if err != nil {
		log.Fatalf("Failed to copy starter template: %v", err)
	}
	fmt.Printf("✅ Copied starter template to: %s\n", outputDir)

	// Create model
	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// --- Tool Definitions ---

	readTool, err := functiontool.New(functiontool.Config{
		Name:        "ReadFile",
		Description: "Reads the entire content of a specified file.",
	}, ReadFile)
	if err != nil {
		log.Fatalf("Failed to create ReadFile tool: %v", err)
	}

	writeTool, err := functiontool.New(functiontool.Config{
		Name:        "WriteFile",
		Description: "Overwrites a file with new content. Use this primarily for NEW files. Prefer SedTool for modifications.",
	}, WriteFile)
	if err != nil {
		log.Fatalf("Failed to create WriteFile tool: %v", err)
	}

	grepTool, err := functiontool.New(functiontool.Config{
		Name:        "GrepFile",
		Description: "Searches for lines matching a regular expression pattern within a file. Useful for finding the exact location of code to modify.",
	}, GrepFile)
	if err != nil {
		log.Fatalf("Failed to create GrepFile tool: %v", err)
	}

	sedTool, err := functiontool.New(functiontool.Config{
		Name:        "SedTool",
		Description: "Performs surgical, line-based modification (replacement or insertion) in a file. Use this for code modifications to save tokens.",
	}, SedTool)
	if err != nil {
		log.Fatalf("Failed to create SedTool tool: %v", err)
	}
	// --- End Tool Definitions ---

	// Create agent
	agent, err := llmagent.New(llmagent.Config{
		Name:        "mvp_agent",
		Model:       model,
		Description: "Agent that modifies a Go/HTMX starter template based on a user's MVP request.",
		Instruction: agentInstruction,
		Tools:       []tool.Tool{readTool, writeTool, grepTool, sedTool}, // New tool set
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create session service
	sessionService := session.InMemoryService()

	// Create runner
	agentRunner, err := runner.New(runner.Config{
		Agent:          agent,
		AppName:        "mvp_agent",
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	// Create session
	userID := "user123"
	appName := "mvp_agent"
	sessResp, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: appName,
		UserID:  userID,
	})
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}

	// Pass the outputDir and MVP request directly in the userMessage
	userMessage := fmt.Sprintf(
		`Here is MVP Request: %s and Working Directory: %s`,
		mvpRequest,
		outputDir,
	)

	// Run agent
	msg := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: userMessage},
		},
	}
	events := agentRunner.Run(ctx, userID, sessResp.Session.ID(), msg, adkagent.RunConfig{})

	fmt.Println("Agent is working...")

	for _, err := range events {
		if err != nil {
			// These are actually normal events, not errors
			fmt.Printf("Error in event stream: %+v\n", err)
		}
	}

	fmt.Println("Agent processing completed!")
}

// --- Tool Structs ---

type ReadFileParams struct {
	FilePath string `json:"filePath" jsonschema:"The path to the file to read."`
}

type ReadFileResult struct {
	Status  string `json:"status"`
	Content string `json:"content,omitempty"`
	Message string `json:"message"`
}

type GrepFileParams struct {
	FilePath string `json:"filePath" jsonschema:"The path to the file to search in."`
	Pattern  string `json:"pattern" jsonschema:"The regular expression pattern to search for (e.g., '^func main')."`
}

type GrepFileResult struct {
	Status  string   `json:"status"`
	Matches []string `json:"matches"`
	Message string   `json:"message"`
}

type SedToolParams struct {
	FilePath     string `json:"filePath" jsonschema:"The path to the file to modify."`
	Pattern      string `json:"pattern" jsonschema:"The regular expression to match the line(s) to be changed. Use the full line content for best results."`
	Replacement  string `json:"replacement" jsonschema:"The new content to replace the matched line(s) with. For insertions, match the line BEFORE the insertion point."`
	InsertBefore bool   `json:"insertBefore" jsonschema:"If true, the replacement content is inserted before the matched line(s). If false or omitted, the matched line is replaced."`
}

type SedToolResult struct {
	Status        string `json:"status"`
	LinesModified int    `json:"linesModified"`
	Message       string `json:"message"`
}

type WriteFileParams struct {
	FilePath string `json:"filePath" jsonschema:"The path to the file to write to. Use this primarily for new files."`
	Content  string `json:"content" jsonschema:"The content to write to the file."`
}

type WriteFileResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// --- Tool Implementations ---

func ReadFile(ctx tool.Context, args ReadFileParams) ReadFileResult {
	fmt.Println("Reading file: ", args.FilePath)
	content, err := os.ReadFile(args.FilePath)
	if err != nil {
		return ReadFileResult{Status: "error", Message: fmt.Sprintf("Error reading file %s: %v", args.FilePath, err)}
	}
	return ReadFileResult{Status: "success", Content: string(content), Message: fmt.Sprintf("Read file %s successfully.", args.FilePath)}
}

func GrepFile(ctx tool.Context, args GrepFileParams) GrepFileResult {
	fmt.Println("Grepping file: ", args.FilePath)
	file, err := os.Open(args.FilePath)
	if err != nil {
		return GrepFileResult{Status: "error", Message: fmt.Sprintf("Error opening file %s: %v", args.FilePath, err)}
	}
	defer file.Close()

	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return GrepFileResult{Status: "error", Message: fmt.Sprintf("Invalid regex pattern: %v", err)}
	}

	var matches []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return GrepFileResult{Status: "error", Message: fmt.Sprintf("Error scanning file: %v", err)}
	}

	if len(matches) == 0 {
		return GrepFileResult{Status: "success", Matches: matches, Message: "No matches found."}
	}
	return GrepFileResult{Status: "success", Matches: matches, Message: fmt.Sprintf("Found %d matches.", len(matches))}
}

func SedTool(ctx tool.Context, args SedToolParams) SedToolResult {
	fmt.Println("SedTool: ", args)
	// 1. Read all lines from the file
	input, err := os.ReadFile(args.FilePath)
	if err != nil {
		return SedToolResult{Status: "error", Message: fmt.Sprintf("Error reading file %s: %v", args.FilePath, err)}
	}
	// Use bytes.NewReader to read the content line-by-line
	scanner := bufio.NewScanner(strings.NewReader(string(input)))

	// 2. Compile the regex pattern
	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return SedToolResult{Status: "error", Message: fmt.Sprintf("Invalid regex pattern: %v", err)}
	}

	var outputLines []string
	linesModified := 0

	// 3. Process lines
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			linesModified++

			if args.InsertBefore {
				// Insert new content, then the original line
				outputLines = append(outputLines, args.Replacement)
				outputLines = append(outputLines, line)
			} else {
				// Replace the original line with new content
				outputLines = append(outputLines, args.Replacement)
			}
		} else {
			outputLines = append(outputLines, line)
		}
	}

	if linesModified == 0 {
		return SedToolResult{
			Status:        "warning",
			LinesModified: 0,
			Message:       fmt.Sprintf("Pattern not found. No modifications made to %s.", args.FilePath),
		}
	}

	// 4. Write modified content back to the file
	output := []byte(strings.Join(outputLines, "\n"))
	if err := os.WriteFile(args.FilePath, output, 0644); err != nil {
		return SedToolResult{Status: "error", Message: fmt.Sprintf("Error writing file %s: %v", args.FilePath, err)}
	}

	return SedToolResult{
		Status:        "success",
		LinesModified: linesModified,
		Message:       fmt.Sprintf("Successfully modified %d line(s) in %s.", linesModified, args.FilePath),
	}
}

func WriteFile(ctx tool.Context, args WriteFileParams) WriteFileResult {
	fmt.Println("Writing file: ", args.FilePath)
	// Ensure the output folder exists before writing
	dir := filepath.Dir(args.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return WriteFileResult{Status: "error", Message: fmt.Sprintf("Failed to create parent directory: %v", err)}
	}

	err := os.WriteFile(args.FilePath, []byte(args.Content), 0644)
	if err != nil {
		return WriteFileResult{Status: "error", Message: err.Error()}
	}
	return WriteFileResult{Status: "success", Message: fmt.Sprintf("Wrote content to file %s successfully.", args.FilePath)}
}

// copyStarterTemplate copies the starter template to a timestamped output directory
func copyStarterTemplate() (string, error) {
	// Create timestamp for unique folder name
	timestamp := time.Now().Format("20060102_150405")
	outputDirName := fmt.Sprintf("output_%s", timestamp)

	// Define paths
	sourceDir := "./data/starter-template"
	outputBaseDir := "./data/outputs"
	outputDir := filepath.Join(outputBaseDir, outputDirName)

	// Create outputs directory if it doesn't exist
	if err := os.MkdirAll(outputBaseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create outputs directory: %v", err)
	}

	// Copy the starter template
	if err := copyDir(sourceDir, outputDir); err != nil {
		return "", fmt.Errorf("failed to copy directory: %v", err)
	}

	return outputDir, nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Set file permissions
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return nil
}
