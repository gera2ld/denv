package filehandler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func (d *FileHandler) ListFiles(prefix string, baseDir string) ([]string, error) {
	baseDir = filepath.Join(d.RootDir, baseDir)

	var files []string
	err := filepath.Walk(filepath.Join(d.RootDir, prefix), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		path, err = filepath.Rel(baseDir, path)
		if err != nil || strings.HasPrefix(path, "..") {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		files = append(files, path)
		return nil
	})
	if err != nil && d.Debug {
		log.Printf("Error listing files: %s\n", err)
	}
	return files, err
}
