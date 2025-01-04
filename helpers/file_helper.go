package helpers

import (
	"os"
	"path/filepath"
	"strings"
)

func GetFilesInFolder(folderPath string) ([]string, error) {
	if strings.HasPrefix(folderPath, ".") {
		absoluteFolderPath, err := filepath.Abs(folderPath)
		if err != nil {
			return nil, err
		}
		folderPath = absoluteFolderPath
	}
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	to_return := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		to_return = append(to_return, file.Name())
	}
	return to_return, nil
}

func IsFileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}
