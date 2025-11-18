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

	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <transcript-file-path> <repository-name>", os.Args[0])
	}
	transcriptFilePath := os.Args[1]
	// NEW: Get repository name from command line
	repoName := os.Args[2]

	// Verify file exists
	if _, err := os.Stat(transcriptFilePath); os.IsNotExist(err) {
		log.Fatalf("File not found: %s", transcriptFilePath)
	}

	// Create model
	model, err := gemini.NewModel(ctx, "gemini-2.0-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	transcriptTool, err := functiontool.New(functiontool.Config{
		Name:        "GenerateSystemPromptFromTranscript",
		Description: "Reads a meeting transcript file and generates a system prompt from its contents with GitHub action suggestions",
	}, GenerateSystemPromptFromTranscript)
	if err != nil {
		log.Fatalf("Failed to create GenerateSystemPromptFromTranscript tool: %v", err)
	}

	githubActionTool, err := functiontool.New(functiontool.Config{
		Name:        "GitHubMCPServerAction",
		Description: "Tool for interacting with GitHub via REST API to create, update, or close issues",
	}, GitHubMCPServerAction)
	if err != nil {
		log.Fatalf("Failed to create GitHubMCPServerAction tool: %v", err)
	}

	githubListTool, err := functiontool.New(functiontool.Config{
		Name:        "GitHubMCPServerListIssues",
		Description: "Lists existing GitHub issues for a repository. Use this to check for duplicates before creating new issues.",
	}, GitHubMCPServerListIssues)
	if err != nil {
		log.Fatalf("Failed to create GitHubMCPServerListIssues tool: %v", err)
	}

	agentInstruction := fmt.Sprintf(`You are a helpful agent that processes meeting transcripts and manages GitHub issues for the **%s** repository.
1. **Always** start by using **GenerateSystemPromptFromTranscript** to get the meeting content.
2. **Pre-check for updates**: Analyze the summary from the transcript. If the summary contains mentions of specific, existing tasks, issues, or ticket numbers (e.g., "Issue #12 discussed," "We need to clarify the scope of the dashboard ticket"), or if the discussion is clearly an elaboration on a prior topic, then proceed to step 3. Otherwise, if the points are entirely new, skip to step 5 (create new issues).
3. **If an update is suspected**: Use **GitHubMCPServerListIssues** to retrieve a list of existing **open** issues in '%s'.
4. Analyze the transcript summary and the list of existing issues.
5. **Crucially**: If a key point already corresponds to an open issue (check issue titles/bodies for similarity), use **GitHubMCPServerAction** with the **'update'** action to add more context or a mermaid diagram to the existing issue.
6. If a key point is entirely new and does not have an open issue, use **GitHubMCPServerAction** with the **'create'** action. Always create issues with a proper description and mermaid diagrams when applicable.
7. The repository name is always '%s'. Do not ask for confirmation; directly perform the necessary action.`, repoName, repoName, repoName)

	// Create agent
	agent, err := llmagent.New(llmagent.Config{
		Name:        "james_agent",
		Model:       model,
		Description: "Agent that processes meeting transcripts, generates actionable prompts, and interacts with GitHub MCP server to manage issues based on meeting discussions.",
		Instruction: agentInstruction,
		Tools:       []tool.Tool{transcriptTool, githubActionTool, githubListTool},
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
	userMessage := fmt.Sprintf("Please process the meeting transcript from file '%s'. Follow your instructions to check for existing issues before creating any new ones in the %s repository.", transcriptFilePath, repoName)

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
	fmt.Println("Checking if file exists : ", args.FilePath)
	if _, err := os.Stat(args.FilePath); os.IsNotExist(err) {
		return GenerateSystemPromptResult{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("File not found: %s", args.FilePath),
		}
	}
	transcript, err := os.ReadFile(args.FilePath)
	if err != nil {
		return GenerateSystemPromptResult{
			Status:       "error",
			ErrorMessage: err.Error(),
		}
	}
	prompt := fmt.Sprintf("System prompt generated from meeting transcript:\n%s\nSummarize the key points and suggest relevant GitHub actions (create/update/close issues) based on the discussion.", string(transcript))

	fmt.Println("Prompt generated: ", prompt)

	return GenerateSystemPromptResult{
		Status: "success",
		Prompt: prompt,
	}
}

// GitHubMCPServerAction tool structs and function (Unchanged, but now used conditionally by LLM)
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
	fmt.Println("Getting GitHub token : ")
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return GitHubActionResult{Status: "error", ErrorMessage: "GitHub token not found in environment variable GITHUB_TOKEN."}
	}
	repo, ok := args.IssueData["repo"].(string)
	if !ok || repo == "" {
		return GitHubActionResult{Status: "error", ErrorMessage: "Missing 'repo' in issue_data."}
	}

	var url string
	var payload map[string]interface{}
	var method string

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
			return GitHubActionResult{Status: "error", ErrorMessage: "Missing 'number' for update action."}
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
			return GitHubActionResult{Status: "error", ErrorMessage: "Missing 'number' for close action."}
		}
		url = fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", repo, issueNumber)
		method = "PATCH"
		payload = map[string]interface{}{"state": "closed"}
	default:
		return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Unknown action: %s", args.Action)}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Failed to marshal payload: %v", err)}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Failed to create request: %v", err)}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Failed to read response: %v", err)}
	}

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Failed to parse response: %v", err)}
		}
		return GitHubActionResult{Status: "success", Result: result}
	} else {
		return GitHubActionResult{Status: "error", ErrorMessage: string(body), Code: resp.StatusCode}
	}
}

// NEW: GitHubMCPServerListIssues tool structs and function
type GitHubListIssuesParams struct {
	Repo  string `json:"repo" jsonschema:"The repository name (e.g., owner/repo)"`
	State string `json:"state,omitempty" jsonschema:"The state of the issues (open, closed, or all). Use 'open' to check for duplicates."`
}

func GitHubMCPServerListIssues(ctx tool.Context, args GitHubListIssuesParams) GitHubActionResult {
	fmt.Println("Getting GitHub token for listing issues: ")
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: "GitHub token not found in environment variable GITHUB_TOKEN.",
		}
	}

	state := args.State
	if state == "" {
		state = "open"
	}

	// Build the URL for listing issues
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=%s", args.Repo, state)

	// Create HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Failed to create request: %v", err)}
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Request failed: %v", err)}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Failed to read response body: %v", err)}
	}

	// Check response status
	if resp.StatusCode == 200 {
		var issues []map[string]interface{}
		if err := json.Unmarshal(body, &issues); err != nil {
			return GitHubActionResult{Status: "error", ErrorMessage: fmt.Sprintf("Failed to parse response body: %v", err)}
		}

		// Prepare lightweight issue list with only title and number
		var lightweightIssues []map[string]interface{}

		for _, issue := range issues {
			title, titleOk := issue["title"].(string)
			numberFloat, numberOk := issue["number"].(float64)

			if titleOk && numberOk {
				lightweightIssues = append(lightweightIssues, map[string]interface{}{
					"title":  title,
					"number": int(numberFloat),
				})
			}
		}

		return GitHubActionResult{
			Status: "success",
			Result: map[string]interface{}{
				"message": "List of open issue titles and numbers for duplicate checking:",
				"issues":  lightweightIssues,
			},
		}
	} else {
		return GitHubActionResult{
			Status:       "error",
			ErrorMessage: string(body),
			Code:         resp.StatusCode,
		}
	}
}

// Helper function to safely get string values from map (Unchanged)
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}
