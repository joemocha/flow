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

func main() {
	fmt.Println("OpenRouter.ai Chatbot Example using Flow")
	fmt.Println("Type 'quit' to exit")

	// Check for OpenRouter API key
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENROUTER_API_KEY environment variable is required")
		os.Exit(1)
	}

	// Initialize OpenAI client with OpenRouter configuration
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
	)

	state := flow.NewSharedState()
	state.Set("conversation_history", []openai.ChatCompletionMessageParamUnion{})
	state.Set("client", client)

	// Create chatbot node with retry capability
	chatNode := flow.NewNode()
	chatNode.SetParams(map[string]interface{}{
		"retries": 2,
	})

	chatNode.SetExecFunc(func(prep interface{}) (interface{}, error) {
		userInput := prep.(string)

		// Get client and conversation history from state
		client := state.Get("client").(openai.Client)
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
