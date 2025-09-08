package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

var (
	// 전역 로거 인스턴스
	Logger *slog.Logger
)

// Config : 로거 설정 구조체
type Config struct {
	Level      string `json:"level"`      // debug, info, warn, error
	Format     string `json:"format"`     // json, text, pretty
	OutputPath string `json:"outputPath"` // stdout, stderr, 또는 파일 경로
}

// Init : 전역 로거 초기화
func Init(config Config) error {
	// 출력 경로 설정
	output, err := getOutput(config.OutputPath)
	if err != nil {
		return err
	}

	// 로그 레벨 설정
	level := getLevel(config.Level)

	// 핸들러 설정
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler

	switch config.Format {
	case "json":
		handler = NewCustomJSONHandler(output, opts)
	case "pretty":
		handler = NewPrettyHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	// 로거 생성 및 전역 설정
	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	return nil
}

// InitDefault : 기본 설정으로 로거 초기화
func InitDefault() error {
	return Init(Config{
		Level:      "info",
		Format:     "pretty",
		OutputPath: "stdout",
	})
}

// getOutput : 출력 경로에 따른 io.Writer 반환
func getOutput(outputPath string) (io.Writer, error) {
	if outputPath == "" {
		outputPath = "stdout"
	}

	switch outputPath {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
}

// getLevel : 레벨 문자열에 따른 slog.Level 반환
func getLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithFields : 구조화된 필드를 사용하여 로거 생성
func WithFields(attrs ...slog.Attr) *slog.Logger {
	ensureLogger()

	// slog.Attr를 []any로 변환
	args := make([]any, len(attrs)*2)
	for i, attr := range attrs {
		args[i*2] = attr.Key
		args[i*2+1] = attr.Value.Any()
	}
	return Logger.With(args...)
}

// WithContext : 컨텍스트 필드를 사용하여 로거 생성
func WithContext(operation string, attrs ...slog.Attr) *slog.Logger {
	contextAttrs := make([]slog.Attr, 0, len(attrs)+1)
	contextAttrs = append(contextAttrs, slog.String("operation", operation))
	contextAttrs = append(contextAttrs, attrs...)
	return WithFields(contextAttrs...)
}

// Info : 정보 로그 출력
func Info(msg string, attrs ...any) {
	if Logger != nil {
		Logger.Info(msg, attrs...)
	}
}

// Debug : 디버그 로그 출력
func Debug(msg string, attrs ...any) {
	if Logger != nil {
		Logger.Debug(msg, attrs...)
	}
}

// Warn : 경고 로그 출력
func Warn(msg string, attrs ...any) {
	if Logger != nil {
		Logger.Warn(msg, attrs...)
	}
}

// Error : 에러 로그 출력
func Error(msg string, attrs ...any) {
	if Logger != nil {
		Logger.Error(msg, attrs...)
	}
}

// Fatal : 치명적 에러 로그 출력 후 프로그램 종료
func Fatal(msg string, attrs ...any) {
	if Logger != nil {
		Logger.Error(msg, attrs...)
	}
	os.Exit(1)
}

// GetLogger : 전역 로거 인스턴스 반환
func GetLogger() *slog.Logger {
	ensureLogger()
	return Logger
}

// WithContextLogger : 컨텍스트를 사용하여 로거 생성
func WithContextLogger(ctx context.Context) *slog.Logger {
	ensureLogger()
	return Logger
}

// ensureLogger : 로거가 초기화되지 않은 경우 기본 설정으로 초기화
func ensureLogger() {
	if Logger == nil {
		_ = InitDefault()
	}
}

// 공통 속성을 위한 헬퍼 함수

// String : 문자열 속성 생성
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

// Int : 정수 속성 생성
func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

// Bool : 불린 속성 생성
func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}

// Any : 임의 타입 속성 생성
func Any(key string, value any) slog.Attr {
	return slog.Any(key, value)
}

// ErrorAttr : 에러 속성 생성
func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

// CustomJSONHandler : 커스텀 시간 포맷을 사용하는 JSON 핸들러
type CustomJSONHandler struct {
	opts   *slog.HandlerOptions
	writer io.Writer
}

// NewCustomJSONHandler : 새로운 CustomJSONHandler 생성
func NewCustomJSONHandler(w io.Writer, opts *slog.HandlerOptions) *CustomJSONHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &CustomJSONHandler{
		opts:   opts,
		writer: w,
	}
}

// Enabled : 로그 레벨이 활성화되어 있는지 확인
func (h *CustomJSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle : 로그 레코드 처리
func (h *CustomJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	// 기본 필드들로 시작
	logEntry := map[string]interface{}{
		"time":  r.Time.UTC().Format("2006-01-02T15:04:05Z"),
		"level": r.Level.String(),
		"msg":   r.Message,
	}

	// 속성들을 직접 추가 (중첩 없이)
	if r.NumAttrs() > 0 {
		r.Attrs(func(a slog.Attr) bool {
			logEntry[a.Key] = a.Value.Any()
			return true
		})
	}

	// JSON 인코딩
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return err
	}

	// 출력
	_, err = h.writer.Write(append(jsonData, '\n'))
	return err
}

// WithAttrs : 속성을 가진 새로운 핸들러 생성
func (h *CustomJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomJSONHandler{
		opts:   h.opts,
		writer: h.writer,
	}
}

// WithGroup : 그룹을 가진 새로운 핸들러 생성
func (h *CustomJSONHandler) WithGroup(name string) slog.Handler {
	return &CustomJSONHandler{
		opts:   h.opts,
		writer: h.writer,
	}
}

// PrettyHandler : 예쁜 형태의 로그 핸들러
type PrettyHandler struct {
	opts   *slog.HandlerOptions
	writer io.Writer
}

// NewPrettyHandler : 새로운 PrettyHandler 생성
func NewPrettyHandler(w io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &PrettyHandler{
		opts:   opts,
		writer: w,
	}
}

// Enabled : 로그 레벨이 활성화되어 있는지 확인
func (h *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle : 로그 레코드 처리
func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	// 시간 포맷팅 (시분초까지만)
	timeStr := r.Time.Format("2006-01-02 15:04:05")

	// 레벨 색상 및 아이콘
	levelStr := h.getLevelString(r.Level)

	// 메시지 포맷팅
	msg := fmt.Sprintf("%s %s %s", timeStr, levelStr, r.Message)

	// 속성들 포맷팅
	if r.NumAttrs() > 0 {
		attrs := make([]string, 0, r.NumAttrs())
		r.Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
			return true
		})
		msg += " " + strings.Join(attrs, " ")
	}

	// 새 줄 추가
	msg += "\n"

	// 출력
	_, err := h.writer.Write([]byte(msg))
	return err
}

// WithAttrs : 속성을 가진 새로운 핸들러 생성
func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		opts:   h.opts,
		writer: h.writer,
	}
}

// WithGroup : 그룹을 가진 새로운 핸들러 생성
func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		opts:   h.opts,
		writer: h.writer,
	}
}

// getLevelString : 레벨에 따른 문자열과 색상 반환
func (h *PrettyHandler) getLevelString(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "🔍 DEBUG"
	case slog.LevelInfo:
		return "ℹ️  INFO "
	case slog.LevelWarn:
		return "⚠️  WARN "
	case slog.LevelError:
		return "❌ ERROR"
	default:
		return fmt.Sprintf("📝 %-5s", level.String())
	}
}
