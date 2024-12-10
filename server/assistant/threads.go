package assistant

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/gastrader/repotalk/types"
	"github.com/sashabaranov/go-openai"
)

func CreateThread(client *openai.Client) (types.ThreadID, error) {
	request := openai.ThreadRequest{}
	thread, err := client.CreateThread(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("could not create thread: %v", err)
	}
	return types.ThreadID(thread.ID), nil
}

func GetThread(client *openai.Client, id types.ThreadID) (openai.Thread, error) {
	thread, err := client.RetrieveThread(context.Background(), string(id))
	if err != nil {
		return openai.Thread{}, fmt.Errorf("could not fetch thread: %v", err)
	}
	return thread, nil
}

func RunThread(client *openai.Client, aid types.AsstID, tid types.ThreadID, msg string) string {
	userMessage := UserMsg(msg)

	//attach msg to thread
	_, err := client.CreateMessage(context.Background(), string(tid), userMessage)
	if err != nil {
		log.Fatal("could not attach message")
	}

	//create a run for thread
	run, err := client.CreateRun(context.Background(), string(tid), openai.RunRequest{
		AssistantID: string(aid),
	})
	if err != nil {
		log.Fatalf("could not create run: %v", err)
	}

	for {
		// Add a delay between checks
		fmt.Print("<")
		time.Sleep(time.Millisecond * 500)

		// Retrieve the updated run status
		updatedRun, err := client.RetrieveRun(context.Background(), string(tid), run.ID)
		if err != nil {
			log.Fatalf("could not create run: %v", err)
		}

		switch updatedRun.Status {
		case "completed":
			return GetFirstThreadMessage(*client, tid)
		case "queued", "in_progress":
		default:
			log.Fatal("unexpected run status:")
		}
		fmt.Print(">\n")
	}
}

func GetFirstThreadMessage(client openai.Client, tid types.ThreadID) string {
	limit := 1
	var order string = "desc"
	var after *string = nil
	var before *string = nil
	var run *string = nil
	list, err := client.ListMessage(context.Background(), string(tid), &limit, &order, after, before, run)
	if err != nil {
		log.Fatalf("could not retrieve messages: %v", err)
	}
	if len(list.Messages) == 0 {
		return "no message found"
	}
	firstMessage := list.Messages[0]
	text := GetContent(firstMessage)
	return text
}

func UserMsg(content string) openai.MessageRequest {
	return openai.MessageRequest{
		Role:    "user",
		Content: content,
	}
}

func GetContent(msg openai.Message) string {
	if len(msg.Content) > 0 {
		firstContent := msg.Content[0]
		if firstContent.Type == "text" && firstContent.Text != nil {
			return firstContent.Text.Value
		}
		if firstContent.ImageFile != nil {
			return "images not supported"
		}
	}
	return "no message found"
}

func RunThreadMsg(client *openai.Client, asstID types.AsstID, threadID types.ThreadID, msg string) (string, error) {
	userMsg := UserMsg(msg)

	_, err := client.CreateMessage(context.Background(), string(threadID), userMsg)
	if err != nil {
		return "", fmt.Errorf("could not attach message to thread: %v", err)
	}

	runRequest := openai.RunRequest{
		AssistantID: string(asstID),
	}
	run, err := client.CreateRun(context.Background(), string(threadID), runRequest)
	if err != nil {
		return "", fmt.Errorf("could not create run for thread: %v", err)
	}

	for {
		run, err := client.RetrieveRun(context.Background(), string(threadID), run.ID)
		if err != nil {

			return "", fmt.Errorf("error while retrieving run: %v", err)
		}

		switch run.Status {
		case "completed":
			return GetFirstThreadMessage(*client, threadID), nil
		case "queued", "in_progress":
		default:

			return "", fmt.Errorf("error while run: %v", run.Status)
		}
		time.Sleep(time.Millisecond * 300)
	}
}

func GetFilesHashMap(client *openai.Client, asstID string) (map[string]string, error) {
	fileIDByName := make(map[string]string)

	var limit *int = nil
	var order *string = nil
	var after *string = nil
	var before *string = nil

	asstFiles, err := client.ListAssistantFiles(context.Background(), string(asstID), limit, order, after, before)
	if err != nil {
		return nil, fmt.Errorf("error listing assistant files: %v", err)
	}

	asstFileIDs := make(map[string]struct{})
	for _, file := range asstFiles.AssistantFiles {
		asstFileIDs[file.ID] = struct{}{}
	}

	orgFiles, err := client.ListFiles(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error listing organization files: %v", err)
	}

	for _, file := range orgFiles.Files {
		if _, exists := asstFileIDs[file.ID]; exists {
			fileIDByName[file.FileName] = file.ID
		}
	}

	return fileIDByName, nil
}

func UploadFileByName(client *openai.Client, asstID string, filePath string, force bool) (string, bool, error) {
	fileName := filepath.Base(filePath)
	fileIDByName, err := GetFilesHashMap(client, asstID)

	if err != nil {
		return "", false, fmt.Errorf("error getting files hashmap: %v", err)
	}
	fileID, exists := fileIDByName[filePath]

	if !force && exists {
		fmt.Println("Existing file found.")
		return fileID, false, nil
	}

	if exists {
		fmt.Println("Deleting old file")

		if err := client.DeleteAssistantFile(context.Background(), asstID, fileID); err != nil {
			fmt.Printf("Can't remove assistant file '%s': %v\n", fileName, err)
		}

		if err := client.DeleteFile(context.Background(), fileID); err != nil {
			fmt.Printf("Can't delete file '%s': %v\n", filePath, err)
		}
	}

	oaFile, err := client.CreateFile(context.Background(), openai.FileRequest{
		FilePath: filePath,
		FileName: fileName,
		Purpose:  "assistants",
	})
	if err != nil {
		return "", false, fmt.Errorf("failed to upload file '%s': %v", filePath, err)
	}

	if _, err := client.CreateAssistantFile(context.Background(), asstID, openai.AssistantFileRequest{
		FileID: oaFile.ID,
	}); err != nil {
		return "", false, fmt.Errorf("failed to attach file '%s' to assistant: %v", filePath, err)
	}

	fmt.Printf("Uploaded and attached file '%s'\n", fileName)
	return oaFile.ID, true, nil
}
