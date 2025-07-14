package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	Flow "github.com/joemocha/flow"
)

// OpenRouter API structures
type OpenRouterRequest struct {
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// LLMConfig holds configuration for LLM API calls
type LLMConfig struct {
	APIKey   string
	Model    string
	Endpoint string
}

// NewLLMConfig creates LLM configuration from environment and defaults
func NewLLMConfig() LLMConfig {
	return LLMConfig{
		APIKey:   os.Getenv("OPENROUTER_API_KEY"),
		Model:    "google/gemini-2.5-flash-lite-preview-06-17",
		Endpoint: "https://openrouter.ai/api/v1/chat/completions",
	}
}

// debugPrintState prints SharedState contents when debug mode is enabled
func debugPrintState(state *Flow.SharedState, label string, debugMode bool) {
	if !debugMode {
		return
	}

	fmt.Printf("\n--- DEBUG: %s ---\n", label)

	// Print conversation messages
	messages := state.Get("messages").([]map[string]interface{})
	fmt.Printf("Conversation History (%d messages):\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("  [%d] %s: %s\n", i, msg["role"], msg["content"])
	}

	// Print current input if exists
	if currentInput := state.Get("current_input"); currentInput != nil {
		fmt.Printf("Current Input: %s\n", currentInput)
	}

	fmt.Println("--- END DEBUG ---\n")
}

// CallLLM makes an API call to OpenRouter and returns the response
func CallLLM(messages []map[string]interface{}, config LLMConfig) (string, error) {
	// Prepare API request
	reqBody := OpenRouterRequest{
		Model:    config.Model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make API call to OpenRouter
	req, err := http.NewRequest("POST", config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var openRouterResp OpenRouterResponse
	err = json.Unmarshal(body, &openRouterResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices received")
	}

	return openRouterResp.Choices[0].Message.Content, nil
}

func main() {
	// Check for debug mode
	debugMode := os.Getenv("DEBUG") == "1"

	// Check for API key
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENROUTER_API_KEY environment variable is required")
		fmt.Println("Get your API key from: https://openrouter.ai/keys")
		os.Exit(1)
	}

	// Initialize shared state with conversation history
	state := Flow.NewSharedState()
	state.Set("messages", []map[string]interface{}{})

	// Create LLM configuration outside Node
	llmConfig := NewLLMConfig()

	// Create single adaptive node for conversation flow
	chatNode := Flow.NewNode()
	chatNode.SetParams(map[string]interface{}{
		"retry_max":   3,
		"retry_delay": time.Second * 2,
	})

	chatNode.SetPrepFunc(func(shared *Flow.SharedState) interface{} {
		return shared.Get("current_input")
	})

	chatNode.SetExecFunc(func(prep interface{}) (interface{}, error) {
		userInput := prep.(string)

		// Get conversation history from SharedState
		messages := state.Get("messages").([]map[string]interface{})

		// Add user message to conversation
		messages = append(messages, map[string]interface{}{
			"role":    "user",
			"content": userInput,
		})

		// Update conversation history in state
		state.Set("messages", messages)

		// Call external LLM function
		aiResponse, err := CallLLM(messages, llmConfig)
		if err != nil {
			return "", err
		}

		// Add AI response to conversation history
		messages = state.Get("messages").([]map[string]interface{})
		messages = append(messages, map[string]interface{}{
			"role":    "assistant",
			"content": aiResponse,
		})
		state.Set("messages", messages)

		return aiResponse, nil
	})

	// Welcome message
	fmt.Println("🤖 Chatbot powered by Google Gemini via OpenRouter")
	if debugMode {
		fmt.Println("DEBUG MODE: SharedState contents will be displayed")
	}
	fmt.Println("Type your message and press Enter. Type '/q' to quit.")
	fmt.Println("---")

	// Debug: Initial state
	debugPrintState(state, "Initial State", debugMode)

	// Create input scanner
	scanner := bufio.NewScanner(os.Stdin)

	// Main conversation loop (PocketFlow style)
	for {
		fmt.Print("You: ")

		// Read user input
		if !scanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(scanner.Text())

		// Check for quit command
		if userInput == "/q" {
			fmt.Println("Goodbye! 👋")
			break
		}

		// Skip empty input
		if userInput == "" {
			continue
		}

		// Store user input in state for the node to access
		state.Set("current_input", userInput)

		// Debug: Show state before API call
		debugPrintState(state, "Before API Call", debugMode)

		// Run chat node with user input (includes automatic retry)
		fmt.Print("Bot: ")
		response := chatNode.Run(state)

		// Debug: Show state after API response
		debugPrintState(state, "After API Response", debugMode)

		// Display AI response
		fmt.Println(response)
		fmt.Println()
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}