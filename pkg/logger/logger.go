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

// Config : logger configuration
type Config struct {
	Level      string `json:"level"`      // debug, info, warn, error
	Format     string `json:"format"`     // json, console
	OutputPath string `json:"outputPath"` // stdout, stderr, or file path
}

// Init : 전역 로거 초기화
func Init(config Config) error {
	var zapConfig zap.Config

	// 로그 레벨 설정
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 출력 경로 설정
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = "stdout"
	}

	// 포맷에 따라 설정
	if config.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// 설정 오버라이드
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.OutputPaths = []string{outputPath}
	zapConfig.ErrorOutputPaths = []string{"stderr"}
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 로거 생성
	Logger, err = zapConfig.Build()
	if err != nil {
		return err
	}

	// 전역 로거 교체
	zap.ReplaceGlobals(Logger)
	return nil
}

// InitDefault : 기본 설정으로 로거 초기화
func InitDefault() error {
	config := Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "stdout",
	}
	return Init(config)
}

// Sync : 버퍼링된 로그 전송
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// WithFields : 구조화된 필드를 사용하여 로거 생성
func WithFields(fields ...zap.Field) *zap.Logger {
	if Logger == nil {
		// 초기화되지 않은 경우 기본 로거 사용
		fallback, _ := zap.NewProduction()
		return fallback
	}
	return Logger.With(fields...)
}

// WithContext : 컨텍스트 필드를 사용하여 로거 생성
func WithContext(operation string, fields ...zap.Field) *zap.Logger {
	contextFields := []zap.Field{
		zap.String("operation", operation),
	}
	contextFields = append(contextFields, fields...)
	return WithFields(contextFields...)
}

// Info : 정보 로그 출력
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Debug : 디버그 로그 출력
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Warn : 경고 로그 출력
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error : 에러 로그 출력
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal : 치명적 에러 로그 출력
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
	os.Exit(1)
}

// GetLogger : 전역 로거 인스턴스 반환
func GetLogger() *zap.Logger {
	if Logger == nil {
		// 초기화되지 않은 경우 기본 설정으로 초기화
		InitDefault()
	}
	return Logger
}
