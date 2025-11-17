package main

import (
	"context"
	"fmt"
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

	// Create model
	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create custom tool
	customTool, err := functiontool.New(functiontool.Config{
		Name:        "AddVariants",
		Description: "Generate three different component variants using HTML and Tailwind CSS and write them to demo.html",
	}, AddVariants)
	if err != nil {
		log.Fatalf("Failed to create custom tool: %v", err)
	}

	// Create agent
	agent, err := llmagent.New(llmagent.Config{
		Name:        "ui_component_agent",
		Model:       model,
		Description: "Generate tailwind components with three variants",
		Instruction: `You MUST use the AddVariants tool for every component request. 
		When the user describes a component, you MUST FIRST GENERATE three distinct variants of the component using **HTML and Tailwind CSS classes**, each variant separated by a clear HTML comment (e.g., ).
		Then, you MUST call AddVariants with the ENTIRE GENERATED HTML for the three variants as the componentDescription parameter.
		Never generate HTML outside of the tool call argument.`,
		Tools: []tool.Tool{customTool},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Create session service
	sessionService := session.InMemoryService()

	// Create runner
	agentRunner, err := runner.New(runner.Config{
		Agent:          agent,
		AppName:        "ui_component_agent",
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	// Serve index.html on root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		email := r.Header.Get("X-Forwarded-Email")
		user := r.Header.Get("X-Forwarded-User")

		log.Printf("Page access - User: %s, Email: %s", user, email)

		http.ServeFile(w, r, "index.html")
	})

	// Handle component generation
	http.HandleFunc("/generate-component", func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		input := r.FormValue("user_input")
		if input == "" {
			http.Error(w, "Component name is required", http.StatusBadRequest)
			return
		}

		// Create session
		userID := "user123"
		appName := "ui_component_agent"
		sessResp, err := sessionService.Create(ctx, &session.CreateRequest{
			AppName: appName,
			UserID:  userID,
		})
		if err != nil {
			fmt.Printf("Error creating session: %v\n", err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Run agent
		msg := &genai.Content{
			Role: "user",
			Parts: []*genai.Part{
				{Text: input},
			},
		}
		events := agentRunner.Run(ctx, userID, sessResp.Session.ID(), msg, adkagent.RunConfig{})

		// Process events and log any errors
		for _, err := range events {
			if err != nil {
				fmt.Printf("ERROR in event stream: %v\n", err)
				log.Printf("Error in event stream: %v", err)
				continue
			}

		}

		// Read the updated demo.html and return it
		htmlContent, err := os.ReadFile("demo.html")
		if err != nil {
			fmt.Printf("Error reading demo.html: %v\n", err)
			http.Error(w, "Error reading generated HTML", http.StatusInternalServerError)
			return
		}

		fmt.Println("HTML Content: ", string(htmlContent))

		// Set content type to HTML
		w.Header().Set("Content-Type", "text/html")
		w.Write(htmlContent)
	})

	fmt.Println("Server running on http://localhost:8000")
	fmt.Println("Open http://localhost:8000 in your browser")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

type AddVariantsParams struct {
	ComponentDescription string `json:"componentDescription" jsonschema:"A detailed description of the UI component to generate, including its purpose, style, and any specific features or behaviors required"`
}

type AddVariantsResult struct {
	Result string `json:"result"`
}

func AddVariants(ctx tool.Context, args AddVariantsParams) AddVariantsResult {
	llmContent := args.ComponentDescription
	fmt.Println("LLM Content: ", llmContent)

	// Write the complete HTML to demo.html
	err := os.WriteFile("demo.html", []byte(llmContent), 0644)
	if err != nil {
		log.Printf("ERROR: Failed to write HTML file: %v", err)
		return AddVariantsResult{
			Result: "Error writing HTML file: " + err.Error(),
		}
	}

	return AddVariantsResult{
		Result: "Successfully created demo.html with components",
	}
}
