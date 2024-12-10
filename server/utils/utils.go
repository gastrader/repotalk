package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func EnsureDir(dir string) (bool, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755) 
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func ListFiles(dir string) ([]string, error) {
	var files []string

	allowedExtensions := map[string]bool{
		".go":   true,
		".ts":   true,
		".tsx":  true,
		".js":   true,
		".jsx":  true,
		".py":   true,
		".java": true,
		".rb":   true,
		".cpp":  true,
		".cs":   true,
		".zig":  true,
		".sh":   true,
		".html": true,
		".yaml": true,
		".toml": true,
		".c":    true,
		".kt":   true,
		".kts":  true,
		".php":  true,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if allowedExtensions[filepath.Ext(path)] {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func ReadToString(filePath string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func GetReader(filePath string) (*bufio.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	return bufio.NewReader(file), nil
}

func LoadFromJSON(filePath string, v interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %s", filePath)
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(v)
}

func SaveToJSON(filePath string, v interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot create file '%s': %v", filePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func BundleToFile(files []string, dstFilePath string) error {
	dstFile, err := os.Create(dstFilePath)
	if err != nil {
		return fmt.Errorf("cannot create file '%s': %v", dstFilePath, err)
	}
	defer dstFile.Close()

	writer := bufio.NewWriter(dstFile)
	for _, file := range files {
		if info, err := os.Stat(file); err != nil || info.IsDir() {
			return fmt.Errorf("cannot bundle '%s': it is not a file", file)
		}

		reader, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("cannot open file '%s': %v", file, err)
		}

		scanner := bufio.NewScanner(reader)
		_, err = fmt.Fprintf(writer, "\n// ==== file path: %s\n", filepath.ToSlash(file))
		if err != nil {
			reader.Close()
			return err
		}

		for scanner.Scan() {
			_, err = fmt.Fprintf(writer, "%s\n", scanner.Text())
			if err != nil {
				reader.Close()
				return err
			}
		}

		if err := scanner.Err(); err != nil {
			reader.Close()
			return fmt.Errorf("error reading file '%s': %v", file, err)
		}

		_, err = fmt.Fprint(writer, "\n\n")
		if err != nil {
			reader.Close()
			return err
		}

		reader.Close()
	}

	return writer.Flush()
}
