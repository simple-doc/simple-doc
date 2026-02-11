package config

import (
	"fmt"
	"os"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func PostgreSQLConnString() string {
	if v := os.Getenv("POSTGRES_CONN_STRING"); v != "" {
		return v
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env("POSTGRES_USER", "postgres"),
		env("POSTGRES_PASSWORD", "postgres"),
		env("POSTGRES_HOST", "localhost"),
		env("POSTGRES_PORT", "5432"),
		env("POSTGRES_DB", "postgres"),
	)
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

func SMTPFrom() string {
	return env("SMTP_FROM", "noreply@example.com")
}

func BaseURL() string {
	return env("BASE_URL", "http://localhost:8080")
}
