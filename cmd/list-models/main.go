package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Get API key from environment
	apiKey := os.Getenv("CHAT_BOX_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: CHAT_BOX_API_KEY not set in environment")
		fmt.Println("Usage: Set CHAT_BOX_API_KEY in your .env file or export it")
		os.Exit(1)
	}

	ctx := context.Background()

	// Create client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// println apikey
	fmt.Printf("Using CHAT_BOX_API_KEY: %s", apiKey)
	fmt.Println("📋 Available Google AI Models:")
	fmt.Println("================================")

	// List models
	iter := client.ListModels(ctx)
	count := 0
	for {
		model, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Error listing models: %v\n", err)
			os.Exit(1)
		}

		count++
		fmt.Printf("\n%d. Model: %s\n", count, model.Name)
		fmt.Printf("   Display Name: %s\n", model.DisplayName)
		fmt.Printf("   Description: %s\n", model.Description)

		// Show supported generation methods
		if len(model.SupportedGenerationMethods) > 0 {
			fmt.Printf("   Supported Methods: ")
			for i, method := range model.SupportedGenerationMethods {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(method)
			}
			fmt.Println()
		}

		// Show input/output token limits
		if model.InputTokenLimit > 0 {
			fmt.Printf("   Input Token Limit: %d\n", model.InputTokenLimit)
		}
		if model.OutputTokenLimit > 0 {
			fmt.Printf("   Output Token Limit: %d\n", model.OutputTokenLimit)
		}
	}

	if count == 0 {
		fmt.Println("No models found. Please check your API key.")
	} else {
		fmt.Printf("\n✅ Total models found: %d\n", count)
		fmt.Println("\nℹ️  To use a model, set CHAT_BOX_MODEL_NAME in your .env file")
		fmt.Println("   Example: CHAT_BOX_MODEL_NAME=gemini-1.5-flash")
	}
}
