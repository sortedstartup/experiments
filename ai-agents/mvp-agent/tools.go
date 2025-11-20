package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"google.golang.org/adk/tool"
)

// Global log channel for tools
var toolLogChannel chan<- string

// Helper function to send logs
func toolLog(msg string) {
	if toolLogChannel != nil {
		select {
		case toolLogChannel <- msg:
		default: // Don't block if channel is full
		}
	}
	fmt.Println(msg) // Still print to console
}

func ReadFile(ctx tool.Context, args ReadFileParams) ReadFileResult {
	toolLog("Reading file: " + args.FilePath)
	content, err := os.ReadFile(args.FilePath)
	if err != nil {
		return ReadFileResult{Status: "error", Message: fmt.Sprintf("Error reading file %s: %v", args.FilePath, err)}
	}
	return ReadFileResult{Status: "success", Content: string(content), Message: fmt.Sprintf("Read file %s successfully.", args.FilePath)}
}

func GrepFile(ctx tool.Context, args GrepFileParams) GrepFileResult {
	toolLog("Grepping file: " + args.FilePath)
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
	toolLog(fmt.Sprintf("SedTool: %+v", args))
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
	toolLog("Writing file: " + args.FilePath)
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
	toolLog("Running go build in: " + args.WorkingDir)

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
	toolLog(fmt.Sprintf("Executing command: %s (in directory: %s)", cmd.String(), args.WorkingDir))

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
	toolLog(fmt.Sprintf("Inserting content at line %d in file: %s", args.LineNumber, args.FilePath))

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
	toolLog("Appending to file: " + args.FilePath)

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
	toolLog(fmt.Sprintf("Renaming file from %s to %s", args.OldPath, args.NewPath))

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
	toolLog(fmt.Sprintf("Moving file from %s to %s", args.SourcePath, args.DestinationPath))

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
	toolLog(fmt.Sprintf("Listing files in directory: %s (recursive: %t)", args.Directory, args.Recursive))

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
