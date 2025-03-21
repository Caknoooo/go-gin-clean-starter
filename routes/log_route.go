package routes

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	LOG_DIR = "./config/logs/query_log"

	LOG_HTML = "logs.html"
)

func LoggerRoute(router *gin.Engine) {
	router.LoadHTMLGlob(LOG_HTML)

	router.GET("/logs/:month", Logger)
	router.GET("/logs", Logger)
}

func Logger(c *gin.Context) {
	month := c.Param("month")
	if month == "" {
		month = time.Now().Format("January")
	}

	logFileName := strings.ToLower(month) + "_query.log"
	logFile := filepath.Join(LOG_DIR, logFileName)

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		c.HTML(http.StatusOK, LOG_HTML, gin.H{
			"Month": month,
			"Logs":  nil,
		})
		return
	}

	file, err := os.Open(logFile)
	if err != nil {
		c.HTML(http.StatusInternalServerError, LOG_HTML, gin.H{
			"Month": month,
			"Logs":  nil,
		})
		return
	}
	defer file.Close()

	var logs []string
	scanner := bufio.NewScanner(file)
	var block strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			if block.Len() > 0 {
				logs = append(logs, block.String())
				block.Reset()
			}
		} else {
			block.WriteString(line + "\n")
		}
	}

	if block.Len() > 0 {
		logs = append(logs, block.String())
	}

	if err := scanner.Err(); err != nil {
		c.HTML(http.StatusInternalServerError, LOG_HTML, gin.H{
			"Month": month,
			"Logs":  nil,
		})
		return
	}

	reverseLogs := reverseSlice(logs)

	c.HTML(http.StatusOK, LOG_HTML, gin.H{
		"Month": month,
		"Logs":  reverseLogs,
	})
}

func reverseSlice(input []string) []string {
	length := len(input)
	reversed := make([]string, length)
	for i, v := range input {
		reversed[length-1-i] = v
	}
	return reversed
}
