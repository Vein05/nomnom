package nomnom

import (
	"context"
	"fmt"
	"log"
	"os"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// QueryOpts contains options for the query
type QueryOpts struct {
	Model string
}

// SendQueryToLLM sends a query to an LLM API to generate new file names
func SendQueryToLLM(client *deepseek.Client, query contentprocessors.Query, opts QueryOpts) error {
	// read prompt from prompt.txt
	prompt, err := os.ReadFile("prompt.txt")
	if err != nil {
		return fmt.Errorf("error reading prompt: %v", err)
	}

	// Iterate through the folders
	for i := range query.Folders {
		folder := &query.Folders[i]
		fmt.Println("Working on Folder Name:", folder.Name)

		// Iterate through the files in the folder
		for j := range folder.FileList {
			file := &folder.FileList[j]
			fmt.Println("Working on File Name:", file.Name)

			// Create a chat completion request
			request := &deepseek.ChatCompletionRequest{
				Model: opts.Model,
				Messages: []deepseek.ChatCompletionMessage{
					{Role: deepseek.ChatMessageRoleSystem, Content: string(prompt)},
					{Role: deepseek.ChatMessageRoleUser, Content: file.Context},
				},
			}

			// Send the request and handle the response
			ctx := context.Background()
			response, err := client.CreateChatCompletion(ctx, request)
			if err != nil {
				return fmt.Errorf("error creating chat completion: %v", err)
			}

			file.NewName = response.Choices[0].Message.Content
		}
	}
	return nil
}

// SendQuery sends a query to the deepseek API to generate new file names
func SendQueryWithDeepSeek(config configutils.Config, query contentprocessors.Query) {
	// Set up the Deepseek client
	client := deepseek.NewClient(os.Getenv("DEEPSEEK_API_KEY"))

	opts := QueryOpts{
		Model: deepseek.DeepSeekChat,
	}

	if err := SendQueryToLLM(client, query, opts); err != nil {
		log.Fatalf("error: %v", err)
	}
}
