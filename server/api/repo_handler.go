package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gastrader/repotalk/assistant"
	"github.com/gastrader/repotalk/types"
	"github.com/gastrader/repotalk/utils"
	"github.com/sashabaranov/go-openai"
)

type RepoHandler struct {
	oaiClient   *openai.Client
	assistantID types.AsstID
}

func NewRepoHandler(oai *openai.Client, as types.AsstID) *RepoHandler {
	return &RepoHandler{
		oaiClient:   oai,
		assistantID: as,
	}
}

func (rh *RepoHandler) CrawlHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		return
	}

	var req types.CrawlRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.GithubURL == "" {
		http.Error(w, "GitHub URL is required", http.StatusBadRequest)
		return
	}

	username, reponame, err := parseGitHubURL(req.GithubURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid GitHub URL: %v", err), http.StatusBadRequest)
		return
	}

	repoDir := fmt.Sprintf("./repos/%s/%s", username, reponame)
	parentRepoDir := fmt.Sprintf("./repos/%s", username)
	bundleDir := fmt.Sprintf("./bundles/%s/%s/bundle.txt", username, reponame)

	if _, err := os.Stat(bundleDir); err == nil {
		fmt.Println("Bundled file already exists. Skipping git clone and bundling.")
	} else if os.IsNotExist(err) {

		err = os.MkdirAll(repoDir, os.ModePerm)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating directory: %v", err), http.StatusInternalServerError)
			return
		}

		err = cloneGitHubRepo(req.GithubURL, repoDir)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error cloning repository: %v", err), http.StatusInternalServerError)
			return
		}

		_, err = utils.EnsureDir(filepath.Dir(bundleDir))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error checking directory: %v", err), http.StatusInternalServerError)
			return
		}

		files, err := utils.ListFiles(repoDir)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error listing directory: %v", err), http.StatusInternalServerError)
			return
		}

		if len(files) == 0 {
			http.Error(w, fmt.Sprintf("No valid files: %v", err), http.StatusInternalServerError)
			return
		}

		err = utils.BundleToFile(files, bundleDir)
		if err != nil {
			log.Fatalf("Failed to bundle files: %v\n", err)
		}

		err = os.RemoveAll(parentRepoDir)
		if err != nil {
			log.Printf("Warning: Failed to delete directory '%s': %v\n", repoDir, err)
		}
	} else {
		// Handle other errors (e.g., permission issues)
		http.Error(w, fmt.Sprintf("Error checking for bundled file: %v", err), http.StatusInternalServerError)
		return
	}

	fileID, _, err := assistant.UploadFileByName(rh.oaiClient, string(rh.assistantID), bundleDir, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error uploading file: %v", err), http.StatusInternalServerError)
		return
	}

	threadID, err := assistant.CreateThread(rh.oaiClient)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating thread: %v", err), http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf("Uploaded file '%s'. Please analyze its contents.", filepath.Base(bundleDir))
	res, err := assistant.RunThreadMsg(rh.oaiClient, rh.assistantID, threadID, message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting thread: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Println("the res is:", res)

	response := types.CrawlResponse{
		Message:  "Crawl initiated successfully",
		URL:      req.GithubURL,
		Username: username,
		Reponame: reponame,
		Response: res,
		ThreadID: string(threadID),
		FileID:   fileID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func parseGitHubURL(githubURL string) (string, string, error) {
	// Example URL: https://github.com/username/reponame
	parts := strings.Split(githubURL, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL")
	}

	username := parts[len(parts)-2]
	repoName := parts[len(parts)-1]

	return username, repoName, nil
}

func cloneGitHubRepo(githubURL, repoDir string) error {
	cmd := exec.Command("git", "clone", githubURL, repoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("Error cloning repository: %v", err)
		return err
	}
	return nil
}

func (rh *RepoHandler) QueryHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.ThreadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	var threadID types.ThreadID
	if req.ThreadID == "" {
		newThreadID, err := assistant.CreateThread(rh.oaiClient)
		fmt.Println("creating new thread", newThreadID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating thread: %v", err), http.StatusInternalServerError)
			return
		}
		threadID = newThreadID
	} else {
		threadID = types.ThreadID(req.ThreadID)
	}

	res, err := assistant.RunThreadMsg(rh.oaiClient, rh.assistantID, threadID, req.Question)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error sending message to thread: %v", err), http.StatusInternalServerError)
		return
	}

	response := types.QueryResponse{
		Message:  "Query initiated successfully",
		Username: req.GithubUser,
		Reponame: req.RepoName,
		Response: res,
		ThreadID: string(threadID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
