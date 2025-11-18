package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	// Check required environment variables
	if os.Getenv("GOOGLE_API_KEY") == "" {
		log.Fatalf("GOOGLE_API_KEY environment variable is required")
	}
	if os.Getenv("GITHUB_TOKEN") == "" {
		log.Println("WARNING: GITHUB_TOKEN environment variable not set. GitHub operations will fail.")
	}

	// Create model
	model, err := gemini.NewModel(ctx, "gemini-2.0-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create custom tools
	transcriptTool, err := functiontool.New(functiontool.Config{
		Name:        "GenerateSystemPromptFromTranscript",
		Description: "Reads a meeting transcript file and generates a system prompt from its contents with GitHub action suggestions",
	}, GenerateSystemPromptFromTranscript)
	if err != nil {
		log.Fatalf("Failed to create GenerateSystemPromptFromTranscript tool: %v", err)
	}

	githubTool, err := functiontool.New(functiontool.Config{
		Name:        "GitHubMCPServerAction",
		Description: "Tool for interacting with GitHub via REST API to create, update, or close issues",
	}, GitHubMCPServerAction)
	if err != nil {
		log.Fatalf("Failed to create GitHubMCPServerAction tool: %v", err)
	}

	// Create agent
	agent, err := llmagent.New(llmagent.Config{
		Name:        "james_agent",
		Model:       model,
		Description: "Agent that processes meeting transcripts, generates actionable prompts, and interacts with GitHub MCP server to manage issues based on meeting discussions.",
		Instruction: `You are a helpful agent that reads meeting transcripts, summarizes key points, and performs GitHub issue actions (create, update, close) using the MCP server based on user input and meeting context. Please create issues with proper description and mermaid diagrams, do not create any issue without proper description. Repo name is sanskaraggarwal2025/Blogging, and please don't ask too much just directly create/update/close issue. Must use listed tools to create/update/close issue.`,
		// Instruction: `You are a helpful agent that reads meeting transcripts, summarizes key points, and performs GitHub issue actions (create, update, close) for the sanskaraggarwal2025/Blogging repository based on meeting context. If there is a clear action item, create the issue directly, even if the description is brief.`,
		Tools: []tool.Tool{transcriptTool, githubTool},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create session service
	sessionService := session.InMemoryService()

	// Create runner
	agentRunner, err := runner.New(runner.Config{
		Agent:          agent,
		AppName:        "james_agent",
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	// Get file path from command line arguments
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <transcript-file-path>", os.Args[0])
	}
	transcriptFilePath := os.Args[1]

	// Verify file exists
	if _, err := os.Stat(transcriptFilePath); os.IsNotExist(err) {
		log.Fatalf("File not found: %s", transcriptFilePath)
	}
	// Create session
	userID := "user123"
	appName := "james_agent"
	sessResp, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: appName,
		UserID:  userID,
	})
	if err != nil {
		fmt.Printf("Error creating session: %v\n", err)
		return
	}

	// Create message to process the transcript file
	userMessage := fmt.Sprintf("Please read the meeting transcript from file '%s', generate a system prompt with key points, and create appropriate GitHub issues for the sanskaraggarwal2025/Blogging repository based on the discussion. Include proper descriptions and mermaid diagrams where applicable.", transcriptFilePath)

	// Run agent
	msg := &genai.Content{
		Role: "user",
		Parts: []*genai.Part{
			{Text: userMessage},
		},
	}
	events := agentRunner.Run(ctx, userID, sessResp.Session.ID(), msg, adkagent.RunConfig{})

	// Process events and display results
	fmt.Printf("Processing transcript file: %s\n", transcriptFilePath)
	fmt.Println("Agent is working...")

	for _, err := range events {
		if err != nil {
			// These are actually normal events, not errors
			fmt.Printf("Error in event stream: %+v\n", err)
		}
	}

	fmt.Println("Agent processing completed!")

}

// GenerateSystemPromptFromTranscript tool structs and function
type GenerateSystemPromptParams struct {
	FilePath string `json:"filePath" jsonschema:"Path to the meeting transcript file to read and process"`
}

type GenerateSystemPromptResult struct {
	Status       string `json:"status"`
	Prompt       string `json:"prompt,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func GenerateSystemPromptFromTranscript(ctx tool.Context, args GenerateSystemPromptParams) GenerateSystemPromptResult {
	// Check if file exists
	fmt.Println("Checking if file exists sanskar: ", args.FilePath)
	if _, err := os.Stat(args.FilePath); os.IsNotExist(err) {
		return GenerateSystemPromptResult{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("File not found: %s", args.FilePath),
		}
	}

	// Read the transcript file
	transcript, err := os.ReadFile(args.FilePath)
	if err != nil {
		return GenerateSystemPromptResult{
			Status:       "error",
			ErrorMessage: err.Error(),
		}
	}

	// Generate the system prompt
	prompt := fmt.Sprintf("System prompt generated from meeting transcript:\n%s\nSummarize the key points and suggest relevant GitHub actions (create/update/close issues) based on the discussion.", string(transcript))

	fmt.Println("Prompt generated sanskar: ", prompt)

	return GenerateSystemPromptResult{
		Status: "success",
		Prompt: prompt,
	}
}

// GitHubMCPServerAction tool structs and function
type GitHubActionParams struct {
	Action    string                 `json:"action" jsonschema:"The action to perform: create, update, or close"`
	IssueData map[string]interface{} `json:"issueData" jsonschema:"Issue data containing repo, title, body, number, etc."`
}

type GitHubActionResult struct {
	Status       string                 `json:"status"`
	Result       map[string]interface{} `json:"result,omitempty"`
	ErrorMessage string                 `json:"errorMessage,omitempty"`
	Code         int                    `json:"code,omitempty"`
}

func GitHubMCPServerAction(ctx tool.Context, args GitHubActionParams) GitHubActionResult {
	// Get GitHub token from environment
	fmt.Println("Getting GitHub token sanskar: ")
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: "GitHub token not found in environment variable GITHUB_TOKEN.",
		}
	}

	// Extract repo from issue data
	repo, ok := args.IssueData["repo"].(string)
	if !ok || repo == "" {
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: "Missing 'repo' in issue_data.",
		}
	}

	var url string
	var payload map[string]interface{}
	var method string

	fmt.Println("Action: ", args.Action)
	fmt.Println("Repo: ", repo)

	switch args.Action {
	case "create":
		url = fmt.Sprintf("https://api.github.com/repos/%s/issues", repo)
		method = "POST"
		payload = map[string]interface{}{
			"title": getStringFromMap(args.IssueData, "title", "No title"),
			"body":  getStringFromMap(args.IssueData, "body", ""),
		}

	case "update":
		issueNumber := getStringFromMap(args.IssueData, "number", "")
		if issueNumber == "" {
			return GitHubActionResult{
				Status:       "error",
				ErrorMessage: "Missing 'number' for update action.",
			}
		}
		url = fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", repo, issueNumber)
		method = "PATCH"
		payload = make(map[string]interface{})
		if title, exists := args.IssueData["title"]; exists {
			payload["title"] = title
		}
		if body, exists := args.IssueData["body"]; exists {
			payload["body"] = body
		}

	case "close":
		issueNumber := getStringFromMap(args.IssueData, "number", "")
		if issueNumber == "" {
			return GitHubActionResult{
				Status:       "error",
				ErrorMessage: "Missing 'number' for close action.",
			}
		}
		url = fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", repo, issueNumber)
		method = "PATCH"
		payload = map[string]interface{}{
			"state": "closed",
		}

	default:
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("Unknown action: %s", args.Action),
		}
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("Failed to marshal payload: %v", err),
		}
	}

	// Create HTTP request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Failed to create request sanskar: ", err)
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	// Make HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to do request sanskar: ", err)
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read response body sanskar: ", err)
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("Failed to read response: %v", err),
		}
	}

	// Check response status
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println("Failed to parse response sanskar: ", err)
			return GitHubActionResult{
				Status:       "error",
				ErrorMessage: fmt.Sprintf("Failed to parse response: %v", err),
			}
		}
		return GitHubActionResult{
			Status: "success",
			Result: result,
		}
	} else {
		fmt.Println("Failed to check response status sanskar: ", resp.StatusCode)
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: string(body),
			Code:         resp.StatusCode,
		}
	}
}

// Helper function to safely get string values from map
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	fmt.Println("Getting string from map sanskar: ", m, key, defaultValue)
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}
