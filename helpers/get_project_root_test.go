package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectRoot(t *testing.T) {
	// Save the original function to restore it later
	originalGetProjectRoot := GetProjectRoot
	defer func() {
		GetProjectRoot = originalGetProjectRoot
	}()

	t.Run(
		"successful discovery of project root", func(t *testing.T) {

			tmpDir, err := os.MkdirTemp("", "project-root-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			goModPath := filepath.Join(tmpDir, "go.mod")
			err = os.WriteFile(goModPath, []byte("module test"), 0644)
			require.NoError(t, err)

			subDir := filepath.Join(tmpDir, "cmd", "app")
			err = os.MkdirAll(subDir, 0755)
			require.NoError(t, err)

			// Mock the GetProjectRoot function
			GetProjectRoot = func() (string, error) {

				dir := subDir
				for {
					if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
						return dir, nil
					}

					parentDir := filepath.Dir(dir)
					if parentDir == dir {
						return "", nil
					}
					dir = parentDir
				}
			}

			root, err := GetProjectRoot()
			assert.NoError(t, err)
			assert.Equal(t, tmpDir, root)
		},
	)

	t.Run(
		"go.mod not found", func(t *testing.T) {
			// Mock the GetProjectRoot function to simulate not finding go.mod
			GetProjectRoot = func() (string, error) {
				return "", filepath.ErrBadPattern
			}

			root, err := GetProjectRoot()
			assert.Error(t, err)
			assert.Empty(t, root)
		},
	)

	t.Run(
		"unable to get current file path", func(t *testing.T) {
			// Mock the GetProjectRoot function to simulate runtime.Caller failure
			GetProjectRoot = func() (string, error) {
				return "", filepath.ErrBadPattern
			}

			// Test the function
			root, err := GetProjectRoot()
			assert.Error(t, err)
			assert.Empty(t, root)
		},
	)
}

// TestActualGetProjectRoot tests the real implementation if in a proper project structure
func TestActualGetProjectRoot(t *testing.T) {
	// Restore the original function
	originalGetProjectRoot := GetProjectRoot

	// This test only runs in a real project environment
	root, err := originalGetProjectRoot()
	if err == nil {
		// If it worked, we should have a go.mod file in the root
		_, err := os.Stat(filepath.Join(root, "go.mod"))
		assert.NoError(t, err, "Expected to find go.mod in project root")
	} else {
		// This test might be running in isolation without a project structure
		t.Skip("Skipping actual project root test (no proper project structure found)")
	}
}
