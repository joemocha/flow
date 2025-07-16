# OpenRouter.ai Chatbot Example

This example demonstrates how to build a chatbot using the Flow library with OpenRouter.ai integration via the OpenAI Go SDK.

## Features

- **OpenRouter.ai Integration**: Uses OpenRouter.ai API for AI-powered responses
- **Conversation History**: Maintains conversation context across messages
- **Retry Logic**: Built-in retry capability using Flow's adaptive node system
- **Error Handling**: Graceful error handling for API failures

## Prerequisites

1. **OpenRouter API Key**: You need an API key from [OpenRouter.ai](https://openrouter.ai/)
2. **Go 1.18+**: This example requires Go 1.18 or later

## Setup

1. **Set Environment Variable**:
   ```bash
   export OPENROUTER_API_KEY="your-api-key-here"
   ```

2. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

## Usage

Run the chatbot:

```bash
go run examples/chatbot/main.go
```

The chatbot will start and you can begin chatting:

```
OpenRouter.ai Chatbot Example using Flow
Type 'quit' to exit

You: Hello!
Bot: Hello! How can I help you today?

You: What's the weather like?
Bot: I don't have access to real-time weather data, but I'd be happy to help you with other questions!

You: quit
Goodbye!
```

## How It Works

### Flow Integration

The chatbot uses Flow's adaptive node system with the following features:

- **Retry Logic**: Configured with `retries: 2` for automatic retry on API failures
- **State Management**: Uses `SharedState` to maintain conversation history and client instance
- **Parameter-Driven Behavior**: Leverages Flow's parameter detection for retry patterns

### OpenRouter.ai Configuration

The chatbot is configured to use OpenRouter.ai with:

- **Base URL**: `https://openrouter.ai/api/v1`
- **Model**: `openai/gpt-3.5-turbo` (can be changed to other OpenRouter models)
- **Authentication**: Uses the `OPENROUTER_API_KEY` environment variable

### Code Structure

```go
// Initialize OpenAI client with OpenRouter configuration
client := openai.NewClient(
    option.WithAPIKey(apiKey),
    option.WithBaseURL("https://openrouter.ai/api/v1"),
)

// Create Flow node with retry capability
chatNode := flow.NewNode()
chatNode.SetParams(map[string]interface{}{
    "retries": 2,
})

// Set execution function for AI chat completion
chatNode.SetExecFunc(func(prep interface{}) (interface{}, error) {
    // Handle user input and get AI response
    // Maintain conversation history
    // Return response or error
})
```

## Available Models

You can change the model by modifying the `Model` parameter in the code. Some popular OpenRouter models include:

- `openai/gpt-3.5-turbo`
- `openai/gpt-4`
- `anthropic/claude-3-haiku`
- `meta-llama/llama-2-70b-chat`
- `google/palm-2-chat-bison`

## Error Handling

The chatbot includes comprehensive error handling:

- **API Key Validation**: Checks for required environment variable
- **Network Errors**: Automatic retry with exponential backoff
- **API Errors**: Graceful error messages for API failures
- **Input Validation**: Handles empty inputs and quit commands

## Customization

You can customize the chatbot by:

1. **Changing the Model**: Modify the `Model` parameter to use different AI models
2. **Adjusting Retry Logic**: Change the `retries` parameter for different retry behavior
3. **Adding System Messages**: Include system prompts for specific behavior
4. **Extending Features**: Add features like conversation saving, user profiles, etc.

## Dependencies

- `github.com/openai/openai-go`: OpenAI Go SDK for API integration
- `github.com/joemocha/flow`: Flow library for workflow orchestration
