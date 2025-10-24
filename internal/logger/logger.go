package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	logsDir     = "logs"
	logFilename = "llm_interactions.json"
)

// LogEntry represents a logged LLM interaction
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	VirtualKey string                 `json:"virtual_key"`
	Provider   string                 `json:"provider"`
	Method     string                 `json:"method"`
	Status     int                    `json:"status"`
	DurationMs int64                  `json:"duration_ms"`
	Request    map[string]interface{} `json:"request"`
	Response   map[string]interface{} `json:"response"`
}

// Logger handles structured logging of LLM interactions
type Logger struct {
	logFile *os.File
}

// New creates a new Logger instance
func New() (*Logger, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Open log file in append mode
	logPath := fmt.Sprintf("%s/%s", logsDir, logFilename)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{logFile: logFile}, nil
}

// LogInteraction logs an LLM interaction to both console and file
func (l *Logger) LogInteraction(entry LogEntry) {
	// Pretty-print JSON for both console and file
	prettyJSON, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		log.Printf("Error marshaling log entry: %v", err)
		return
	}

	// Log to console
	log.Printf("LLM Interaction Log:\n%s", string(prettyJSON))

	// Log to file without separator
	logOutput := string(prettyJSON) + "\n"

	if _, err := l.logFile.WriteString(logOutput); err != nil {
		log.Printf("Error writing to log file: %v", err)
	}
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}
