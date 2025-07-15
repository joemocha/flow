package main

import (
"bufio"
"fmt"
"os"
"strings"

flow "github.com/joemocha/flow"
)

func main() {
	fmt.Println("Simple Chatbot Example using Flow")
	fmt.Println("Type 'quit' to exit")

	state := flow.NewSharedState()
	state.Set("conversation_history", []string{})

	// Create chatbot node with retry capability
	chatNode := flow.NewNode()
	chatNode.SetParams(map[string]interface{}{
"retries": 2,
})

	chatNode.SetExecFunc(func(prep interface{}) (interface{}, error) {
userInput := prep.(string)

// Simple response logic
var response string
switch {
case strings.Contains(strings.ToLower(userInput), "hello"):
response = "Hello! How can I help you today?"
case strings.Contains(strings.ToLower(userInput), "how are you"):
response = "I'm doing great! Thanks for asking."
case strings.Contains(strings.ToLower(userInput), "weather"):
response = "I don't have access to weather data, but I hope it's nice where you are!"
case strings.Contains(strings.ToLower(userInput), "help"):
response = "I'm a simple chatbot. Try asking me about the weather or saying hello!"
default:
response = "That's interesting! Tell me more."
}

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
		
		// Add to conversation history
		history := state.Get("conversation_history").([]string)
		history = append(history, "User: "+userInput)
		state.Set("conversation_history", history)

		// Set prep function to get current input
		chatNode.SetPrepFunc(func(shared *flow.SharedState) interface{} {
return shared.Get("current_input")
})

		// Run the chatbot node
		response := chatNode.Run(state)
		
		// Add response to history
		history = state.Get("conversation_history").([]string)
		history = append(history, "Bot: "+response)
		state.Set("conversation_history", history)
		
		fmt.Printf("Bot: %s\n", response)
	}
}
