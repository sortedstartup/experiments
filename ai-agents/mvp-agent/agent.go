package main

import (
	"bufio" // For GrepFile and SedTool
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
You are a **MVP creator agent**.
You main job is to take requirements from the user and based on that create a working Minimum viable product.

A folder with a starter template with required files will be given to you.
Your main job is to modify the files in the starter template and modify the files to implement the features according to the requirements.

<starter_template>
It is a go lang web app with ui in index.html + tailwind + htmx

<file_structure>
	- backend/main.go
	- backend/webapp.go --> add your APIs here
	- backend/ui/index.html  --> tailwind + htmx
	no database is used, use in memory structures to store the data
</file_structure>

</starter_template>

Steps to follow for creating a working MVP from the users requirements
1. **Understand:** 
 - List all files in the working directory
 - First think and come up with a list of changes required to implement the users requirement for creating a working MVP
 - for the changes think what REST, APIs and UI components are needed.

2. **Modify:** Use SedTool for line-level changes or WriteFile for complete rewrites
 - Determine which files need to be modified in the Go backend (Echo framework) and HTML/HTMX frontend.
 - Use RenameFile or MoveFile if you need to reorganize files
 - Make sure your go code and ui code compiles

3. do a go build to verify your code builds and works

<coding_guidelines>
- Go backend uses echo framework
- We dont have a delete file tool, use rename file to soft delete a file
- NEVER create, write or edit go.sum, its NOT needed the build process will generate it
- You should never need to make changes to main.go, changes should be in webapp.go
- All go code that you generate must be in webapp.go
- Feel free to use go templates for returning direct html via APIs
- use HTMX to directly rendered html from the backend and display it as required
- use your judgement where you need a REST API and where you need direct HTML
- Make sure your code compiles
- Keep UI simple and minimal
- use simple colors in UI
</coding_guidelines>


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
	model, err := gemini.NewModel(ctx, "gemini-2.5-pro", &genai.ClientConfig{
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

	goBuildTool, err := functiontool.New(functiontool.Config{
		Name:        "GoBuild",
		Description: "Executes 'go build ./...' in the specified working directory and returns build logs and status. Use this to verify that your Go code compiles successfully.",
	}, GoBuild)
	if err != nil {
		log.Fatalf("Failed to create GoBuild tool: %v", err)
	}

	insertInFileAtLineTool, err := functiontool.New(functiontool.Config{
		Name:        "InsertInFileAtLine",
		Description: "Inserts content at a specific line number in a file (1-based indexing). The content is inserted before the specified line number. Useful for adding new code at precise locations.",
	}, InsertInFileAtLine)
	if err != nil {
		log.Fatalf("Failed to create InsertInFileAtLine tool: %v", err)
	}

	appendToFileTool, err := functiontool.New(functiontool.Config{
		Name:        "AppendToFile",
		Description: "Appends content to the end of a file. Automatically handles newlines. Creates the file if it doesn't exist. Useful for adding new functions or code blocks to the end of files.",
	}, AppendToFile)
	if err != nil {
		log.Fatalf("Failed to create AppendToFile tool: %v", err)
	}

	renameFileTool, err := functiontool.New(functiontool.Config{
		Name:        "RenameFile",
		Description: "Renames or moves a file. Provide the current file path (oldPath) and the desired new path (newPath). Works for both simple renames and moving to different directories.",
	}, RenameFile)
	if err != nil {
		log.Fatalf("Failed to create RenameFile tool: %v", err)
	}

	moveFileTool, err := functiontool.New(functiontool.Config{
		Name:        "MoveFile",
		Description: "Moves a file from one location to another. Creates parent directories if needed. Use this to relocate files to different directories.",
	}, MoveFile)
	if err != nil {
		log.Fatalf("Failed to create MoveFile tool: %v", err)
	}

	listFilesTool, err := functiontool.New(functiontool.Config{
		Name:        "ListFiles",
		Description: "Lists all files and directories in a specified directory. Supports recursive listing to explore entire directory trees. Directories are marked with a trailing slash.",
	}, ListFiles)
	if err != nil {
		log.Fatalf("Failed to create ListFiles tool: %v", err)
	}
	// --- End Tool Definitions ---

	// Create agent
	agent, err := llmagent.New(llmagent.Config{
		Name:        "mvp_agent",
		Model:       model,
		Description: "Agent that modifies a Go/HTMX starter template based on a user's MVP request.",
		Instruction: agentInstruction,
		Tools:       []tool.Tool{readTool, writeTool, grepTool, sedTool, goBuildTool, insertInFileAtLineTool, appendToFileTool, renameFileTool, moveFileTool, listFilesTool},
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
	fmt.Println("Output Directory: ", outputDir)
	fmt.Println("MVP Request: ", mvpRequest)
	userMessage := fmt.Sprintf(
		`Create a MVP based on this requirements documents: ./%s, Code Working Directory: %s, `,
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

type GoBuildParams struct {
	WorkingDir string `json:"workingDir" jsonschema:"The working directory where 'go build' should be executed."`
}

type GoBuildResult struct {
	Status     string `json:"status"`
	BuildLogs  string `json:"buildLogs"`
	Successful bool   `json:"successful"`
	Message    string `json:"message"`
}

type InsertInFileAtLineParams struct {
	FilePath   string `json:"filePath" jsonschema:"The path to the file to modify."`
	LineNumber int    `json:"lineNumber" jsonschema:"The line number where content should be inserted (1-based indexing). Content will be inserted before this line."`
	Content    string `json:"content" jsonschema:"The content to insert."`
}

type InsertInFileAtLineResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type AppendToFileParams struct {
	FilePath string `json:"filePath" jsonschema:"The path to the file to append to."`
	Content  string `json:"content" jsonschema:"The content to append to the end of the file."`
}

type AppendToFileResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type RenameFileParams struct {
	OldPath string `json:"oldPath" jsonschema:"The current path of the file to rename."`
	NewPath string `json:"newPath" jsonschema:"The new path/name for the file."`
}

type RenameFileResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type MoveFileParams struct {
	SourcePath      string `json:"sourcePath" jsonschema:"The path to the file to move."`
	DestinationPath string `json:"destinationPath" jsonschema:"The destination path where the file should be moved."`
}

type MoveFileResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ListFilesParams struct {
	Directory string `json:"directory" jsonschema:"The directory path to list files from."`
	Recursive bool   `json:"recursive" jsonschema:"If true, recursively lists all files in subdirectories. If false, only lists files in the specified directory."`
}

type ListFilesResult struct {
	Status  string   `json:"status"`
	Files   []string `json:"files"`
	Message string   `json:"message"`
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

func GoBuild(ctx tool.Context, args GoBuildParams) GoBuildResult {
	fmt.Println("Running go build in: ", args.WorkingDir)

	// Check if the working directory exists
	if _, err := os.Stat(args.WorkingDir); os.IsNotExist(err) {
		return GoBuildResult{
			Status:     "error",
			BuildLogs:  "",
			Successful: false,
			Message:    fmt.Sprintf("Working directory does not exist: %s", args.WorkingDir),
		}
	}

	// Create the go build command
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = args.WorkingDir
	fmt.Printf("Executing command: %s (in directory: %s)\n", cmd.String(), args.WorkingDir)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Combine stdout and stderr for the build logs
	buildLogs := stdout.String()
	if stderr.Len() > 0 {
		if buildLogs != "" {
			buildLogs += "\n"
		}
		buildLogs += stderr.String()
	}

	if err != nil {
		return GoBuildResult{
			Status:     "error",
			BuildLogs:  buildLogs,
			Successful: false,
			Message:    fmt.Sprintf("Build failed: %v", err),
		}
	}

	return GoBuildResult{
		Status:     "success",
		BuildLogs:  buildLogs,
		Successful: true,
		Message:    "Build completed successfully",
	}
}

func InsertInFileAtLine(ctx tool.Context, args InsertInFileAtLineParams) InsertInFileAtLineResult {
	fmt.Printf("Inserting content at line %d in file: %s\n", args.LineNumber, args.FilePath)

	// Read the file
	content, err := os.ReadFile(args.FilePath)
	if err != nil {
		return InsertInFileAtLineResult{
			Status:  "error",
			Message: fmt.Sprintf("Error reading file %s: %v", args.FilePath, err),
		}
	}

	// Split content into lines
	lines := strings.Split(string(content), "\n")

	// Validate line number (1-based indexing)
	if args.LineNumber < 1 || args.LineNumber > len(lines)+1 {
		return InsertInFileAtLineResult{
			Status:  "error",
			Message: fmt.Sprintf("Invalid line number %d. File has %d lines.", args.LineNumber, len(lines)),
		}
	}

	// Insert content at the specified line (convert to 0-based index)
	insertIndex := args.LineNumber - 1

	// Split new content into lines in case it's multi-line
	newLines := strings.Split(args.Content, "\n")

	// Create the result by inserting new lines
	result := make([]string, 0, len(lines)+len(newLines))
	result = append(result, lines[:insertIndex]...)
	result = append(result, newLines...)
	result = append(result, lines[insertIndex:]...)

	// Write back to file
	output := strings.Join(result, "\n")
	if err := os.WriteFile(args.FilePath, []byte(output), 0644); err != nil {
		return InsertInFileAtLineResult{
			Status:  "error",
			Message: fmt.Sprintf("Error writing file %s: %v", args.FilePath, err),
		}
	}

	return InsertInFileAtLineResult{
		Status:  "success",
		Message: fmt.Sprintf("Successfully inserted content at line %d in %s", args.LineNumber, args.FilePath),
	}
}

func AppendToFile(ctx tool.Context, args AppendToFileParams) AppendToFileResult {
	fmt.Println("Appending to file: ", args.FilePath)

	// Open file in append mode, create if doesn't exist
	file, err := os.OpenFile(args.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return AppendToFileResult{
			Status:  "error",
			Message: fmt.Sprintf("Error opening file %s: %v", args.FilePath, err),
		}
	}
	defer file.Close()

	// Check if file is empty or doesn't end with newline
	fileInfo, err := file.Stat()
	if err != nil {
		return AppendToFileResult{
			Status:  "error",
			Message: fmt.Sprintf("Error getting file info for %s: %v", args.FilePath, err),
		}
	}

	// If file is not empty, ensure there's a newline before appending
	if fileInfo.Size() > 0 {
		// Read last byte to check if it's a newline
		content, err := os.ReadFile(args.FilePath)
		if err == nil && len(content) > 0 && content[len(content)-1] != '\n' {
			// Add newline before appending
			if _, err := file.WriteString("\n"); err != nil {
				return AppendToFileResult{
					Status:  "error",
					Message: fmt.Sprintf("Error writing newline to %s: %v", args.FilePath, err),
				}
			}
		}
	}

	// Append content
	if _, err := file.WriteString(args.Content); err != nil {
		return AppendToFileResult{
			Status:  "error",
			Message: fmt.Sprintf("Error appending to file %s: %v", args.FilePath, err),
		}
	}

	// Ensure content ends with newline
	if len(args.Content) > 0 && args.Content[len(args.Content)-1] != '\n' {
		if _, err := file.WriteString("\n"); err != nil {
			return AppendToFileResult{
				Status:  "error",
				Message: fmt.Sprintf("Error writing final newline to %s: %v", args.FilePath, err),
			}
		}
	}

	return AppendToFileResult{
		Status:  "success",
		Message: fmt.Sprintf("Successfully appended content to %s", args.FilePath),
	}
}

func RenameFile(ctx tool.Context, args RenameFileParams) RenameFileResult {
	fmt.Printf("Renaming file from %s to %s\n", args.OldPath, args.NewPath)

	// Ensure destination directory exists if the path contains a directory
	destDir := filepath.Dir(args.NewPath)
	if destDir != "." && destDir != "" {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return RenameFileResult{
				Status:  "error",
				Message: fmt.Sprintf("Failed to create directory: %v", err),
			}
		}
	}

	// Rename/move the file
	err := os.Rename(args.OldPath, args.NewPath)
	if err != nil {
		return RenameFileResult{
			Status:  "error",
			Message: fmt.Sprintf("Failed to rename: %v", err),
		}
	}

	return RenameFileResult{
		Status:  "success",
		Message: fmt.Sprintf("Renamed %s to %s", args.OldPath, args.NewPath),
	}
}

func MoveFile(ctx tool.Context, args MoveFileParams) MoveFileResult {
	fmt.Printf("Moving file from %s to %s\n", args.SourcePath, args.DestinationPath)

	// Check if source file exists
	if _, err := os.Stat(args.SourcePath); os.IsNotExist(err) {
		return MoveFileResult{
			Status:  "error",
			Message: fmt.Sprintf("Source file does not exist: %s", args.SourcePath),
		}
	}

	// Check if destination already exists
	if _, err := os.Stat(args.DestinationPath); err == nil {
		return MoveFileResult{
			Status:  "error",
			Message: fmt.Sprintf("Destination file already exists: %s", args.DestinationPath),
		}
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(args.DestinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return MoveFileResult{
			Status:  "error",
			Message: fmt.Sprintf("Failed to create destination directory: %v", err),
		}
	}

	// Move the file (os.Rename works for both rename and move)
	if err := os.Rename(args.SourcePath, args.DestinationPath); err != nil {
		return MoveFileResult{
			Status:  "error",
			Message: fmt.Sprintf("Error moving file: %v", err),
		}
	}

	return MoveFileResult{
		Status:  "success",
		Message: fmt.Sprintf("Successfully moved %s to %s", args.SourcePath, args.DestinationPath),
	}
}

func ListFiles(ctx tool.Context, args ListFilesParams) ListFilesResult {
	fmt.Printf("Listing files in directory: %s (recursive: %t)\n", args.Directory, args.Recursive)

	// Check if directory exists
	dirInfo, err := os.Stat(args.Directory)
	if err != nil {
		if os.IsNotExist(err) {
			return ListFilesResult{
				Status:  "error",
				Files:   []string{},
				Message: fmt.Sprintf("Directory does not exist: %s", args.Directory),
			}
		}
		return ListFilesResult{
			Status:  "error",
			Files:   []string{},
			Message: fmt.Sprintf("Error accessing directory: %v", err),
		}
	}

	// Check if it's actually a directory
	if !dirInfo.IsDir() {
		return ListFilesResult{
			Status:  "error",
			Files:   []string{},
			Message: fmt.Sprintf("Path is not a directory: %s", args.Directory),
		}
	}

	var files []string

	if args.Recursive {
		// Recursively walk the directory tree
		err := filepath.Walk(args.Directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Include both files and directories in the listing
			// Use relative path from the base directory for cleaner output
			relPath, err := filepath.Rel(args.Directory, path)
			if err != nil {
				relPath = path
			}
			// Skip the root directory itself
			if relPath != "." {
				if info.IsDir() {
					files = append(files, relPath+"/")
				} else {
					files = append(files, relPath)
				}
			}
			return nil
		})
		if err != nil {
			return ListFilesResult{
				Status:  "error",
				Files:   []string{},
				Message: fmt.Sprintf("Error walking directory tree: %v", err),
			}
		}
	} else {
		// List only the immediate directory contents
		entries, err := os.ReadDir(args.Directory)
		if err != nil {
			return ListFilesResult{
				Status:  "error",
				Files:   []string{},
				Message: fmt.Sprintf("Error reading directory: %v", err),
			}
		}

		for _, entry := range entries {
			if entry.IsDir() {
				files = append(files, entry.Name()+"/")
			} else {
				files = append(files, entry.Name())
			}
		}
	}

	return ListFilesResult{
		Status:  "success",
		Files:   files,
		Message: fmt.Sprintf("Found %d items in %s", len(files), args.Directory),
	}
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
