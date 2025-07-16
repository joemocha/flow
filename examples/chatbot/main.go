package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	flow "github.com/joemocha/flow"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// setupOpenAIClient initializes the OpenAI client with OpenRouter configuration
func setupOpenAIClient() (openai.Client, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return openai.Client{}, fmt.Errorf("OPENROUTER_API_KEY environment variable is required")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
	)

	return client, nil
}

// initializeConversation sets up the initial conversation history with system message
func initializeConversation() []openai.ChatCompletionMessageParamUnion {
	return []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a helpful AI assistant. Be concise and friendly in your responses."),
	}
}

// createChatNode creates and configures the chat node with execution logic
func createChatNode(client openai.Client, state *flow.SharedState) *flow.Node {
	chatNode := flow.NewNode()
	chatNode.SetParams(map[string]interface{}{
		"retries": 2,
	})

	chatNode.SetExecFunc(func(prep interface{}) (interface{}, error) {
		userInput := prep.(string)

		// Get conversation history from state (client is passed as parameter)
		history := state.Get("conversation_history").([]openai.ChatCompletionMessageParamUnion)

		// Add user message to conversation history
		userMessage := openai.UserMessage(userInput)
		history = append(history, userMessage)

		// Create chat completion request
		ctx := context.Background()
		completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Messages: history,
			Model:    "moonshotai/kimi-k2:free", // Using OpenRouter model format
		})
		if err != nil {
			return "", fmt.Errorf("failed to get response from OpenRouter: %w", err)
		}

		// Extract response
		response := completion.Choices[0].Message.Content

		// Add assistant response to conversation history
		assistantMessage := completion.Choices[0].Message.ToParam()
		history = append(history, assistantMessage)
		state.Set("conversation_history", history)

		return response, nil
	})

	return chatNode
}

func main() {
	fmt.Println("OpenRouter.ai Chatbot Example using Flow")
	fmt.Println("Type 'quit' to exit")

	// Setup OpenAI client
	client, err := setupOpenAIClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Setup shared state with conversation history
	state := flow.NewSharedState()
	state.Set("conversation_history", initializeConversation())

	// Create chatbot node with retry capability
	chatNode := createChatNode(client, state)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if userInput == "" {
			continue
		}

		// Store user input in state
		state.Set("current_input", userInput)

		// Set prep function to get current input
		chatNode.SetPrepFunc(func(shared *flow.SharedState) interface{} {
			return shared.Get("current_input")
		})

		// Run the chatbot node
		response := chatNode.Run(state)

		fmt.Printf("Bot: %s\n", response)
	}
}
