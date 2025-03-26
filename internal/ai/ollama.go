package nomnom

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"

	api "github.com/ollama/ollama/api"
)

func SendQueryWithOllama(config configutils.Config, query contentprocessors.Query) (result contentprocessors.Query, err error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("🤖 Using Ollama model: %s", config.AI.Model)

	// Process each folder's files
	for i := range query.Folders {
		folder := &query.Folders[i]
		log.Printf("📁 Processing folder: %s", folder.Name)

		// Process each file in the folder
		for j := range folder.FileList {
			file := &folder.FileList[j]

			messages := []api.Message{
				{
					Role:    "system",
					Content: query.Prompt,
				},
				{
					Role:    "user",
					Content: file.Context,
				},
			}

			var newName string
			resp := func(response api.ChatResponse) error {
				newName = removeThink(response.Message.Content)
				newName = checkAndAddExtension(newName, file.Name)
				return nil
			}

			stream := false
			err := client.Chat(context.Background(), &api.ChatRequest{
				Model:    config.AI.Model,
				Messages: messages,
				Stream:   &stream,
			}, resp)

			if err != nil {
				log.Printf("❌ Failed to process file %s: %v", file.Name, err)
				continue
			}

			file.NewName = newName
			log.Printf("✅ Processed %s -> %s", file.Name, file.NewName)
		}
	}

	return query, nil
}
func removeThink(s string) string {
	// Remove everything between <Think> and </Think> tags
	startTag := "<think>"
	endTag := "</think>"
	result := s

	for {
		startIdx := strings.Index(result, startTag)
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(result[startIdx:], endTag)
		if endIdx == -1 {
			break
		}
		endIdx += startIdx + len(endTag)

		result = result[:startIdx] + result[endIdx:]
	}

	// Remove spaces and return
	s = strings.TrimSpace(result)
	s = strings.ReplaceAll(s, "  ", " ")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

func checkAndAddExtension(s string, file string) string {
	// Check if the string has an extension
	if strings.Contains(s, ".") {
		return s
	}
	extension := filepath.Ext(file)
	// Add a default extension
	return s + extension
}
