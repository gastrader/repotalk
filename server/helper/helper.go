package buddy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gastrader/repotalk/assistant"
	"github.com/gastrader/repotalk/types"
	"github.com/gastrader/repotalk/utils"
	"github.com/sashabaranov/go-openai"
)

type Helper struct {
	Dir       string
	OaiClient *openai.Client
	AsstID    types.AsstID
	Config    types.AsstConfig
}

type Conv struct {
	Thread_ID types.ThreadID
}

func (h *Helper) DataDir() (string, error) {
	dataDir := filepath.Join(h.Dir, ".Helper")
	_, err := utils.EnsureDir(dataDir)
	if err != nil {
		return "", err
	}
	return dataDir, nil
}

func (h *Helper) DataFilesDir() (string, error) {
	dataDir, err := h.DataDir()
	if err != nil {
		return "", err
	}

	filesDir := filepath.Join(dataDir, "files")
	_, err = utils.EnsureDir(filesDir)
	if err != nil {
		return "", err
	}
	return filesDir, nil
}

func (h *Helper) LoadOrCreateConv(recreate bool) (*Conv, error) {
	dataDir, err := h.DataDir()
	if err != nil {
		return nil, err
	}

	convFile := filepath.Join(dataDir, "conv.json")

	if recreate {
		if _, err := os.Stat(convFile); err == nil {
			if err := os.Remove(convFile); err != nil {
				return nil, fmt.Errorf("failed to remove existing conversation file: %v", err)
			}
		}
	}

	conv := &Conv{}
	if err := utils.LoadFromJSON(convFile, conv); err == nil {
		_, err := assistant.GetThread(h.OaiClient, conv.Thread_ID)
		if err != nil {
			return nil, fmt.Errorf("cannot find thread_id for %v: %v", conv, err)
		}
		fmt.Println("Conversation loaded")
		return conv, nil
	} else {
		threadID, err := assistant.CreateThread(h.OaiClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create new thread: %v", err)
		}
		fmt.Println("Conversation created")
		conv.Thread_ID = threadID
		if err := utils.SaveToJSON(convFile, conv); err != nil {
			return nil, err
		}
		return conv, nil
	}
}

func (h *Helper) Chat(conv Conv, msg string) (string, error) {
	res, err := assistant.RunThreadMsg(h.OaiClient, h.AsstID, conv.Thread_ID, msg)
	if err != nil {
		return "", fmt.Errorf("failed to chat: %v", err)
	}
	return res, nil
}

func (h *Helper) UploadFiles(recreate bool) (int, error) {
	numUploaded := 0

	dataFilesDir, err := h.DataFilesDir() 
	if err != nil {
		return 0, err
	}

	files, err := utils.ListFiles(dataFilesDir)
	if err != nil {
		return 0, err
	}

	for _, file := range files {
		if !strings.Contains(file, ".Helper") {
			return 0, fmt.Errorf("error should not delete: '%s'", file)
		}
		if err := os.Remove(file); err != nil {
			return 0, err
		}
	}

	for _, bundle := range h.Config.FileBundles {
		srcDir := filepath.Join(h.Dir, bundle.SrcDir)

		if info, err := os.Stat(srcDir); err == nil && info.IsDir() {
			files, err := utils.ListFiles(srcDir)
			if err != nil {
				return 0, err
			}

			if len(files) > 0 {
				bundleFileName := fmt.Sprintf("%s-bundle-%s.%s", bundle.BundleName, h.AsstID, bundle.DstExt)
				bundleFile := filepath.Join(dataFilesDir, bundleFileName)
				utils.BundleToFile(files, bundleFile)
				forceReupload := recreate

				_, uploaded, err := assistant.UploadFileByName(h.OaiClient, string(h.AsstID), bundleFile, forceReupload)
				if err != nil {
					return 0, err
				}

				if uploaded {
					numUploaded++
				}
			}
		}
	}

	return numUploaded, nil
}