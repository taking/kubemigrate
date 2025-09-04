package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	Logger *zap.Logger
)

// Config represents logger configuration
type Config struct {
	Level      string `json:"level"`      // debug, info, warn, error
	Format     string `json:"format"`     // json, console
	OutputPath string `json:"outputPath"` // stdout, stderr, or file path
}

// Init initializes the global logger
func Init(config Config) error {
	var zapConfig zap.Config

	// Set log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Set output path
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = "stdout"
	}

	// Configure based on format
	if config.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Override configuration
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.OutputPaths = []string{outputPath}
	zapConfig.ErrorOutputPaths = []string{"stderr"}
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create logger
	Logger, err = zapConfig.Build()
	if err != nil {
		return err
	}

	// Replace global logger
	zap.ReplaceGlobals(Logger)
	return nil
}

// InitDefault initializes logger with default configuration
func InitDefault() error {
	config := Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	}
	return Init(config)
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// WithFields creates a logger with structured fields
func WithFields(fields ...zap.Field) *zap.Logger {
	if Logger == nil {
		// Fallback to default logger if not initialized
		fallback, _ := zap.NewProduction()
		return fallback
	}
	return Logger.With(fields...)
}

// WithContext creates a logger with context fields
func WithContext(operation string, fields ...zap.Field) *zap.Logger {
	contextFields := []zap.Field{
		zap.String("operation", operation),
	}
	contextFields = append(contextFields, fields...)
	return WithFields(contextFields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
	os.Exit(1)
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Logger == nil {
		// Initialize with default config if not already initialized
		InitDefault()
	}
	return Logger
}
