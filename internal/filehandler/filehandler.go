package filehandler

import (
	"log"
	"os"
	"path/filepath"
)

type FileHandler struct {
	RootDir string
	Debug   bool
}

func NewFileHandler(rootDir string, debug bool) *FileHandler {
	return &FileHandler{RootDir: rootDir, Debug: debug}
}

func (d *FileHandler) ReadFile(path string) (string, error) {
	filePath := filepath.Join(d.RootDir, path)
	if d.Debug {
		log.Printf("Reading file: %s\n", filePath)
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		if d.Debug {
			log.Printf("Error reading file: %s\n", err)
		}
		return "", err
	}
	return string(data), nil
}

func (d *FileHandler) WriteFile(path, content string) error {
	filePath := filepath.Join(d.RootDir, path)
	if d.Debug {
		log.Printf("Writing file: %s\n", filePath)
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil && d.Debug {
		log.Printf("Error writing file: %s\n", err)
	}
	return err
}

func (d *FileHandler) DeleteFile(path string) error {
	filePath := filepath.Join(d.RootDir, path)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(filePath)
}
