package main

import (
	// For GrepFile and SedTool

	"context"
	"fmt"
	"log"
	"os"

	// For GrepFile and SedTool
	// For SedTool (strings.Join)

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

// RunMVPAgentInDirectory runs the MVP agent in the specified output directory
func RunMVPAgentInDirectory(ctx context.Context, outputDir string) error {
	// Check required environment variables
	if os.Getenv("GOOGLE_API_KEY") == "" {
		return fmt.Errorf("GOOGLE_API_KEY environment variable is required")
	}

	// Create model
	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		return fmt.Errorf("failed to create model: %v", err)
	}

	// --- Tool Definitions ---

	readTool, err := functiontool.New(functiontool.Config{
		Name:        "ReadFile",
		Description: "Reads the entire content of a specified file.",
	}, ReadFile)
	if err != nil {
		return fmt.Errorf("failed to create ReadFile tool: %v", err)
	}

	writeTool, err := functiontool.New(functiontool.Config{
		Name:        "WriteFile",
		Description: "Overwrites a file with new content. Use this primarily for NEW files. Prefer SedTool for modifications.",
	}, WriteFile)
	if err != nil {
		return fmt.Errorf("failed to create WriteFile tool: %v", err)
	}

	grepTool, err := functiontool.New(functiontool.Config{
		Name:        "GrepFile",
		Description: "Searches for lines matching a regular expression pattern within a file. Useful for finding the exact location of code to modify.",
	}, GrepFile)
	if err != nil {
		return fmt.Errorf("failed to create GrepFile tool: %v", err)
	}

	sedTool, err := functiontool.New(functiontool.Config{
		Name:        "SedTool",
		Description: "Performs surgical, line-based modification (replacement or insertion) in a file. Use this for code modifications to save tokens.",
	}, SedTool)
	if err != nil {
		return fmt.Errorf("failed to create SedTool tool: %v", err)
	}

	goBuildTool, err := functiontool.New(functiontool.Config{
		Name:        "GoBuild",
		Description: "Executes 'go build ./...' in the specified working directory and returns build logs and status. Use this to verify that your Go code compiles successfully.",
	}, GoBuild)
	if err != nil {
		return fmt.Errorf("failed to create GoBuild tool: %v", err)
	}

	insertInFileAtLineTool, err := functiontool.New(functiontool.Config{
		Name:        "InsertInFileAtLine",
		Description: "Inserts content at a specific line number in a file (1-based indexing). The content is inserted before the specified line number. Useful for adding new code at precise locations.",
	}, InsertInFileAtLine)
	if err != nil {
		return fmt.Errorf("failed to create InsertInFileAtLine tool: %v", err)
	}

	appendToFileTool, err := functiontool.New(functiontool.Config{
		Name:        "AppendToFile",
		Description: "Appends content to the end of a file. Automatically handles newlines. Creates the file if it doesn't exist. Useful for adding new functions or code blocks to the end of files.",
	}, AppendToFile)
	if err != nil {
		return fmt.Errorf("failed to create AppendToFile tool: %v", err)
	}

	renameFileTool, err := functiontool.New(functiontool.Config{
		Name:        "RenameFile",
		Description: "Renames or moves a file. Provide the current file path (oldPath) and the desired new path (newPath). Works for both simple renames and moving to different directories.",
	}, RenameFile)
	if err != nil {
		return fmt.Errorf("failed to create RenameFile tool: %v", err)
	}

	moveFileTool, err := functiontool.New(functiontool.Config{
		Name:        "MoveFile",
		Description: "Moves a file from one location to another. Creates parent directories if needed. Use this to relocate files to different directories.",
	}, MoveFile)
	if err != nil {
		return fmt.Errorf("failed to create MoveFile tool: %v", err)
	}

	listFilesTool, err := functiontool.New(functiontool.Config{
		Name:        "ListFiles",
		Description: "Lists all files and directories in a specified directory. Supports recursive listing to explore entire directory trees. Directories are marked with a trailing slash.",
	}, ListFiles)
	if err != nil {
		return fmt.Errorf("failed to create ListFiles tool: %v", err)
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
		return fmt.Errorf("failed to create agent: %v", err)
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
		return fmt.Errorf("failed to create runner: %v", err)
	}

	// Create session
	userID := "web_user"
	appName := "mvp_agent"
	sessResp, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: appName,
		UserID:  userID,
	})
	if err != nil {
		return fmt.Errorf("error creating session: %v", err)
	}

	// Create user message
	userMessage := fmt.Sprintf(
		`Create an MVP based on the requirements document in the code working directory: %s. This is also your working directory.`,
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

	log.Printf("Agent is working on directory: %s", outputDir)

	for _, err := range events {
		if err != nil {
			// These are actually normal events, not errors
			fmt.Printf("Error in event stream: %+v\n", err)
		}
	}

	fmt.Println("Agent processing completed!")
	return nil
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
