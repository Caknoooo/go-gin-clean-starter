package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func GetProjectRoot() (string, error) {
	// Get the path of the current file (database.go)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to get current file path")
	}

	// Start from the current file's directory and walk up until we find go.mod
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", fmt.Errorf("project root not found (could not locate go.mod)")
		}
		dir = parentDir
	}
}
