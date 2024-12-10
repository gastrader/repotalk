package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/gastrader/repotalk/api"
	"github.com/gastrader/repotalk/assistant"
	"github.com/gastrader/repotalk/types"
	"github.com/sashabaranov/go-openai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	asstCFG := types.AsstConfig{
		Name:  "repo_talk_01",
		Model: "gpt-3.5-turbo-1106",
	}
	asst := assistant.LoadOrCreate(*client, asstCFG, false)

	instructionsFile := "./instructions.md"
	content, err := os.ReadFile(instructionsFile)
	if err != nil {
		log.Fatalf("Error reading instructions file: %v", err)
	}

	// Upload the instructions to the assistant
	assistant.UploadInstructions(client, asst, string(content))

	repoHandler := api.NewRepoHandler(client, asst)
	http.HandleFunc("/api/v1/crawl", repoHandler.CrawlHandler)
	http.HandleFunc("/api/v1/query", repoHandler.QueryHandler)

	port := ":8080"
	fmt.Printf("Server is running on http://localhost%s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
