package helpers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/palantir/stacktrace"
)

func FilesInFolder(folderPath string) ([]string, error) {
	if strings.HasPrefix(folderPath, ".") {
		absoluteFolderPath, err := filepath.Abs(folderPath)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get files in folder: %s", folderPath)
		}
		folderPath = absoluteFolderPath
	}
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to read directory: %s", folderPath)
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
