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

// SendQuery sends a query to the deepseek API to generate new file names
func SendQuery(config configutils.Config, query contentprocessors.Query) {
	// Set up the Deepseek client
	client := deepseek.NewClient(os.Getenv("DEEPSEEK_API_KEY"))

	// Iterate through the folders
	for i := range query.Folders {
		folder := &query.Folders[i]
		fmt.Println("World on Folder Name:", folder.Name)

		// Iterate through the files in the folder
		for j := range folder.FileList {
			file := &folder.FileList[j]
			fmt.Println("Working on File Name:", file.Name)

			// Create a chat completion request
			request := &deepseek.ChatCompletionRequest{
				Model: deepseek.DeepSeekChat,
				Messages: []deepseek.ChatCompletionMessage{
					{Role: deepseek.ChatMessageRoleSystem, Content: "You are a desktop organizer that creates nice names for the files with their context. Only respond with the new name."},
					{Role: deepseek.ChatMessageRoleUser, Content: file.Context},
				},
			}

			// Send the request and handle the response
			ctx := context.Background()
			response, err := client.CreateChatCompletion(ctx, request)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			file.NewName = response.Choices[0].Message.Content
		}
	}

}
