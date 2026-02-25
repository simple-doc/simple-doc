package config

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func PostgresConnString() string {
	return os.Getenv("POSTGRES_CONN_STRING")
}

func PostgreSQLConnString() string {
	if v := PostgresConnString(); v != "" {
		return v
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		PostgresUser(),
		PostgresPassword(),
		PostgresHost(),
		PostgresPort(),
		PostgresDB(),
	)
}

func PostgresHost() string {
	return env("POSTGRES_HOST", "localhost")
}

func PostgresPort() string {
	return env("POSTGRES_PORT", "5432")
}

func PostgresDB() string {
	return env("POSTGRES_DB", "postgres")
}

func PostgresUser() string {
	return env("POSTGRES_USER", "postgres")
}

func PostgresPassword() string {
	return env("POSTGRES_PASSWORD", "postgres")
}

func Port() string {
	return env("PORT", "8080")
}

func MigrationsDir() string {
	return env("MIGRATIONS_DIR", "migrations")
}

func TemplatesDir() string {
	return env("TEMPLATES_DIR", "templates")
}

func ContentDir() string {
	return env("CONTENT_DIR", "content")
}

func StaticDir() string {
	return env("STATIC_DIR", "static")
}

func SMTPHost() string {
	return env("SMTP_HOST", "localhost")
}

func SMTPPort() string {
	return env("SMTP_PORT", "25")
}

func SMTPUser() string {
	return env("SMTP_USER", "")
}

func SMTPPass() string {
	return env("SMTP_PASS", "")
}

func SMTPPass2() string {
	return env("SMTP_PASS", "")
}

func SMTPFrom() string {
	return env("SMTP_FROM", "noreply@example.com")
}

func BaseURL() string {
	return env("BASE_URL", "http://localhost:8080")
}

func LogLevel() string {
	return env("LOG_LEVEL", "info")
}

func LogFormat() string {
	return env("LOG_FORMAT", "text")
}

func LogFile() string {
	return env("LOG_FILE", "")
}

func newHandler(w io.Writer, level slog.Level, format string) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}
	if strings.ToLower(format) == "json" {
		return slog.NewJSONHandler(w, opts)
	}
	return slog.NewTextHandler(w, opts)
}

func InitLogging() {
	var consoleLevel slog.Level
	switch strings.ToLower(LogLevel()) {
	case "debug":
		consoleLevel = slog.LevelDebug
	case "warn":
		consoleLevel = slog.LevelWarn
	case "error":
		consoleLevel = slog.LevelError
	default:
		consoleLevel = slog.LevelInfo
	}

	format := LogFormat()

	if path := LogFile(); path != "" {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			slog.SetDefault(slog.New(newHandler(os.Stdout, consoleLevel, format)))
			slog.Error("failed to open log file, falling back to console only", "path", path, "error", err)
			return
		}

		consoleHandler := newHandler(os.Stdout, consoleLevel, format)
		fileHandler := newHandler(f, slog.LevelDebug, format)

		slog.SetDefault(slog.New(&multiHandler{handlers: []slog.Handler{consoleHandler, fileHandler}}))
		return
	}

	slog.SetDefault(slog.New(newHandler(os.Stdout, consoleLevel, format)))
}

// multiHandler fans out log records to multiple handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(_ context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(context.Background(), level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
