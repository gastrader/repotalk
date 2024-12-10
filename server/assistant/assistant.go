package assistant

import (
	"context"
	"fmt"
	"log"

	"github.com/gastrader/repotalk/types"
	"github.com/sashabaranov/go-openai"
)

func CreateAssistant(client openai.Client, config types.AsstConfig) types.AsstID {
	AsstReq := openai.AssistantRequest{
		Model: config.Model,
		Name:  &config.Name,
		Tools: []openai.AssistantTool{
			{Type: "file_search"},
		},
	}
	AsstObj, err := client.CreateAssistant(context.Background(), AsstReq)
	if err != nil {
		log.Fatal("Could not create assistant: ", err)
	}

	return types.AsstID(AsstObj.ID)
}

func LoadOrCreate(client openai.Client, config types.AsstConfig, recreate bool) types.AsstID {
	existingAsst, err := findAsst(&client, config.Name)
	if err != nil {
		log.Fatalf("Error finding assistant: %v", err)
	}
	if existingAsst != nil {
		if recreate {
			deleted := DeleteAsst(&client, types.AsstID(existingAsst.ID))
			if deleted {
				fmt.Println("Assistant deleted.")
				asstID := CreateAssistant(client, config)
				fmt.Println("Created assistant.")
				return asstID
			}
			log.Fatal("error deleting asst")
		}
		fmt.Println("Assistant loaded")
		return types.AsstID(existingAsst.ID)
	}
	asstID := CreateAssistant(client, config)
	fmt.Println("Created assistant. ")
	return asstID
}

func findAsst(client *openai.Client, name string) (*openai.Assistant, error) {
	assistants, err := listAssistants(client)
	if err != nil {
		return nil, err
	}
	for _, asst := range assistants {
		if asst.Name != nil && *asst.Name == name {
			return &asst, nil
		}
	}
	return nil, nil
}

func listAssistants(client *openai.Client) ([]openai.Assistant, error) {
	var limit *int = nil
	var order, after, before *string
	oaiAssts, err := client.ListAssistants(context.Background(), limit, order, after, before)
	if err != nil {
		return nil, fmt.Errorf("failed to list assistants: %w", err)
	}
	return oaiAssts.Assistants, nil
}

func UploadInstructions(client *openai.Client, id types.AsstID, content string) {
	_, err := client.ModifyAssistant(context.Background(), string(id), openai.AssistantRequest{
		Instructions: &content,
	})
	if err != nil {
		log.Fatal("instructions could not be uploaded")
	}
}

func DeleteAsst(client *openai.Client, id types.AsstID) bool {

	res, err := client.DeleteAssistant(context.Background(), string(id))
	if err != nil {
		fmt.Println("error deleting assistant: ", id)
		return false
	}
	if !res.Deleted {
		fmt.Println("could not delete assistant: ", id)
		return false
	}
	return true
}
