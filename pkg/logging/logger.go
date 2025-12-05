package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents logging severity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
}

// Logger manages application logging with rotation
type Logger struct {
	level      LogLevel
	file       *os.File
	logger     *log.Logger
	mu         sync.Mutex
	maxSize    int64 // Max size in bytes before rotation
	maxBackups int   // Number of old log files to keep
	logPath    string
	component  string // e.g., "broadcaster", "listener", "playback"
}

// Config holds logger configuration
type Config struct {
	Component  string   // Component name (broadcaster, listener, etc.)
	Level      LogLevel // Minimum log level
	MaxSize    int64    // Max file size in MB (default: 10)
	MaxBackups int      // Number of backups to keep (default: 3)
	LogDir     string   // Log directory (default: ~/.meshradio/logs)
}

// DefaultConfig returns sensible defaults
func DefaultConfig(component string) Config {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".meshradio", "logs")

	return Config{
		Component:  component,
		Level:      INFO, // Default to INFO level
		MaxSize:    10 * 1024 * 1024, // 10 MB
		MaxBackups: 3,
		LogDir:     logDir,
	}
}

// NewLogger creates a new logger instance
func NewLogger(cfg Config) (*Logger, error) {
	// Parse level from environment if set
	if envLevel := os.Getenv("MESHRADIO_LOG_LEVEL"); envLevel != "" {
		switch envLevel {
		case "DEBUG":
			cfg.Level = DEBUG
		case "INFO":
			cfg.Level = INFO
		case "WARN":
			cfg.Level = WARN
		case "ERROR":
			cfg.Level = ERROR
		}
	}

	// Override log dir from environment if set
	if envDir := os.Getenv("MESHRADIO_LOG_DIR"); envDir != "" {
		cfg.LogDir = envDir
	}

	// Create log directory
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logPath := filepath.Join(cfg.LogDir, fmt.Sprintf("%s.log", cfg.Component))
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer (file + stdout)
	writer := io.MultiWriter(file, os.Stdout)

	l := &Logger{
		level:      cfg.Level,
		file:       file,
		logger:     log.New(writer, "", 0), // We'll format timestamps ourselves
		maxSize:    cfg.MaxSize,
		maxBackups: cfg.MaxBackups,
		logPath:    logPath,
		component:  cfg.Component,
	}

	return l, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// rotate performs log rotation if needed
func (l *Logger) rotate() error {
	// Check file size
	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	if info.Size() < l.maxSize {
		return nil // No rotation needed
	}

	// Close current file
	l.file.Close()

	// Rotate old files
	for i := l.maxBackups - 1; i >= 0; i-- {
		var oldPath, newPath string
		if i == 0 {
			oldPath = l.logPath
			newPath = fmt.Sprintf("%s.1", l.logPath)
		} else {
			oldPath = fmt.Sprintf("%s.%d", l.logPath, i)
			newPath = fmt.Sprintf("%s.%d", l.logPath, i+1)
		}

		// Remove oldest if it exists
		if i == l.maxBackups-1 {
			os.Remove(newPath)
		}

		// Rename if file exists
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Open new file
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	l.file = file
	writer := io.MultiWriter(file, os.Stdout)
	l.logger.SetOutput(writer)

	return nil
}

// log writes a log message
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return // Skip if below threshold
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if rotation is needed
	l.rotate()

	// Format: 2024-12-05 10:30:45 [INFO] [broadcaster] message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]
	message := fmt.Sprintf(format, args...)

	l.logger.Printf("%s [%s] [%s] %s\n", timestamp, levelName, l.component, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// GetLogPath returns the path to the current log file
func (l *Logger) GetLogPath() string {
	return l.logPath
}

// SetLevel changes the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}
