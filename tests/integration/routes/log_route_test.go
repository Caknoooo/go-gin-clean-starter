package routes_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Caknoooo/go-gin-clean-starter/routes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerRoute_Integration(t *testing.T) {

	tempDir := t.TempDir()
	currentMonth := time.Now().Format("January")
	testMonth := "December"

	createTestLogFile(t, tempDir, currentMonth, "current_month_log1\n\ncurrent_month_log2\n")
	createTestLogFile(t, tempDir, testMonth, "december_log1\n\ndecember_log2\n")

	originalLogDir := routes.LOG_DIR
	routes.LOG_DIR = tempDir
	defer func() { routes.LOG_DIR = originalLogDir }()

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	tempHTMLFile := filepath.Join(tempDir, "logs.html")
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Logs - {{.Month}}</title>
	</head>
	<body>
		<h1>Logs for {{.Month}}</h1>
		{{if .Logs}}
		<ul>
			{{range .Logs}}
			<li>{{.}}</li>
			{{end}}
		</ul>
		{{else}}
		<p>No logs available.</p>
		{{end}}
	</body>
	</html>
	`
	err := os.WriteFile(tempHTMLFile, []byte(htmlContent), 0644)
	require.NoError(t, err, "Failed to create temporary HTML template")

	originalLogHTML := routes.LOG_HTML
	routes.LOG_HTML = tempHTMLFile
	defer func() { routes.LOG_HTML = originalLogHTML }()

	routes.LoggerRoute(router)

	t.Run(
		"GET /logs - current month", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/logs", nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), currentMonth)
			assert.Contains(t, w.Body.String(), "current_month_log1")
			assert.Contains(t, w.Body.String(), "current_month_log2")
			assert.True(
				t, strings.Index(w.Body.String(), "current_month_log2") <
					strings.Index(w.Body.String(), "current_month_log1"), "Logs should be reversed",
			)
		},
	)

	t.Run(
		"GET /logs/:month - specific month", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/logs/December", nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), testMonth)
			assert.Contains(t, w.Body.String(), "december_log1")
			assert.Contains(t, w.Body.String(), "december_log2")
			assert.True(
				t, strings.Index(w.Body.String(), "december_log2") <
					strings.Index(w.Body.String(), "december_log1"), "Logs should be reversed",
			)
		},
	)

	t.Run(
		"GET /logs/:month - non-existent month", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/logs/Nonexistent", nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "Nonexistent")
			assert.NotContains(t, w.Body.String(), "current_month_log1")
			assert.NotContains(t, w.Body.String(), "december_log1")
		},
	)

	t.Run(
		"GET /logs - with empty log file", func(t *testing.T) {

			emptyMonth := "January"
			createTestLogFile(t, tempDir, emptyMonth, "")

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/logs/January", nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), emptyMonth)
			assert.NotContains(t, w.Body.String(), "current_month_log1")
		},
	)

	t.Run(
		"GET /logs - with malformed log file", func(t *testing.T) {

			malformedMonth := "February"
			createTestLogFile(t, tempDir, malformedMonth, "log1log2log3")

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/logs/February", nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), malformedMonth)

			assert.Contains(t, w.Body.String(), "log1log2log3")
		},
	)
}

func createTestLogFile(t *testing.T, dir string, month string, content string) {
	t.Helper()
	logFileName := strings.ToLower(month) + "_query.log"
	logPath := filepath.Join(dir, logFileName)

	err := os.WriteFile(logPath, []byte(content), 0644)
	require.NoError(t, err, "Failed to create test log file")
}

func TestReverseSlice_Integration(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []string{"a"},
			expected: []string{"a"},
		},
		{
			name:     "multiple elements",
			input:    []string{"a", "b", "c"},
			expected: []string{"c", "b", "a"},
		},
		{
			name:     "real log entries",
			input:    []string{"log1", "log2", "log3"},
			expected: []string{"log3", "log2", "log1"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				result := routes.ReverseSlice(tt.input)
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}
