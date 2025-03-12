package main

import (
	"nomnom/cmd"
	aideepseek "nomnom/internal/ai"
	contentprocessors "nomnom/internal/content"
	configutils "nomnom/internal/utils"
	"os"
)

func main() {
	// Load configuration
	config := configutils.LoadConfig("config.json")

	// Create a query
	query, err := contentprocessors.NewQuery("", ".", "config.json", false, true, false)

	if err != nil {
		println("Error creating query:", err)
		os.Exit(1)
	}
	// populate the query with folders using ProcessDirectory function
	processedQuery, err := contentprocessors.ProcessDirectory(query.Dir)
	if err != nil {
		println("Error processing directory:", err)
		os.Exit(1)
	}
	query.Folders = processedQuery.Folders

	// Print some values from the config and query
	println("Output:", config.Output)
	println("Dir:", query.Dir)

	// Call the SendQuery function
	aideepseek.SendQueryWithOpenRouter(config, *query)

	cmd.Execute()
}
