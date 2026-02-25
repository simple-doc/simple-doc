package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"docgen"
	"docgen/config"
	"docgen/handlers"
	"docgen/internal/db"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

func mask(s string) string {
	if s == "" {
		return ""
	}
	return "***"
}

func maskConnString(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "***"
	}
	if _, hasPass := u.User.Password(); hasPass {
		u.User = url.UserPassword(u.User.Username(), "***")
	}
	return u.String()
}

func main() {
	config.InitLogging()
	ctx := context.Background()

	configAttrs := []any{
		"port", config.Port(),
		"base_url", config.BaseURL(),
	}
	if cs := config.PostgresConnString(); cs != "" {
		configAttrs = append(configAttrs, "postgres_conn_string", maskConnString(cs))
	} else {
		configAttrs = append(configAttrs,
			"postgres_host", config.PostgresHost(),
			"postgres_port", config.PostgresPort(),
			"postgres_db", config.PostgresDB(),
			"postgres_user", config.PostgresUser(),
			"postgres_password", mask(config.PostgresPassword()),
		)
	}
	configAttrs = append(configAttrs,
		"migrations_dir", config.MigrationsDir(),
		"templates_dir", config.TemplatesDir(),
		"content_dir", config.ContentDir(),
		"static_dir", config.StaticDir(),
		"smtp_host", config.SMTPHost(),
		"smtp_port", config.SMTPPort(),
		"smtp_user", config.SMTPUser(),
		"smtp_pass", mask(config.SMTPPass()),
		"smtp_from", config.SMTPFrom(),
		"log_level", config.LogLevel(),
		"log_format", config.LogFormat(),
		"log_file", config.LogFile(),
	)
	slog.Info("config", configAttrs...)

	// Connect to PostgreSQL
	pool, err := pgxpool.New(ctx, config.PostgreSQLConnString())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to PostgreSQL")

	// Run migrations
	migrationsFS := docgen.ResolveFS(config.MigrationsDir(), docgen.EmbeddedMigrations())
	d, err := iofs.New(migrationsFS, ".")
	if err != nil {
		slog.Error("failed to create migration source", "error", err)
		os.Exit(1)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, "pgx5://"+config.PostgreSQLConnString()[len("postgres://"):]+"&x-migrations-table=simpledoc_version")
	if err != nil {
		slog.Error("failed to initialize migrations", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("migrations: nothing to apply")
		} else {
			slog.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Info("migrations applied")
	}

	// Ensure site_settings row exists
	if _, err := pool.Exec(ctx, `INSERT INTO site_settings (singleton) VALUES (TRUE) ON CONFLICT DO NOTHING`); err != nil {
		slog.Error("failed to ensure site_settings", "error", err)
		os.Exit(1)
	}

	// Ensure default roles exist
	if _, err := pool.Exec(ctx,
		`INSERT INTO roles (name, description) VALUES
			('admin', 'Full access to all features'),
			('editor', 'Can edit content')
		 ON CONFLICT (name) DO NOTHING`); err != nil {
		slog.Error("failed to ensure default roles", "error", err)
		os.Exit(1)
	}

	// Parse templates with custom functions
	templatesFS := docgen.ResolveFS(config.TemplatesDir(), docgen.EmbeddedTemplates())
	funcMap := template.FuncMap{
		"formatBytes": handlers.FormatBytes,
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "*.html")
	if err != nil {
		slog.Error("failed to parse templates", "error", err)
		os.Exit(1)
	}

	h := &handlers.Handlers{
		DB:      &db.Queries{Pool: pool},
		Tmpl:    tmpl,
		FuncMap: funcMap,
	}

	// Enable template hot-reload when using local templates directory
	if _, err := os.Stat(config.TemplatesDir()); err == nil {
		h.TemplatesFS = templatesFS
		slog.Info("dev mode: templates will be re-parsed on each request")
	}

	// Session cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := h.DB.DeleteExpiredSessions(context.Background()); err != nil {
				slog.Error("session cleanup failed", "error", err)
			}
		}
	}()

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout)
	mux.HandleFunc("GET /reset-password", h.ResetPasswordPage)
	mux.HandleFunc("POST /reset-password", h.ResetPassword)
	mux.HandleFunc("GET /{$}", h.Home)
	mux.HandleFunc("GET /settings", h.RequireEditor(h.EditHomeForm))
	mux.HandleFunc("POST /settings", h.RequireEditor(h.UpdateHome))
	mux.HandleFunc("GET /sections/new", h.RequireEditor(h.NewSectionForm))
	mux.HandleFunc("POST /sections", h.RequireEditor(h.CreateSection))
	mux.HandleFunc("GET /images/{filename}", h.Image)
	mux.HandleFunc("POST /images/upload", h.RequireEditor(h.UploadImage))
	mux.HandleFunc("POST /images/{filename}/update", h.RequireEditor(h.UpdateImageHandler))
	mux.HandleFunc("POST /images/{filename}/rename", h.RequireEditor(h.RenameImage))
	mux.HandleFunc("POST /images/{filename}/delete", h.RequireEditor(h.DeleteImage))
	mux.HandleFunc("GET /rows/new", h.RequireEditor(h.NewRowForm))
	mux.HandleFunc("POST /rows/{$}", h.RequireEditor(h.CreateRow))
	mux.HandleFunc("GET /rows/{id}/edit", h.RequireEditor(h.EditRowForm))
	mux.HandleFunc("POST /rows/{id}", h.RequireEditor(h.UpdateRow))
	mux.HandleFunc("POST /rows/{id}/delete", h.RequireEditor(h.DeleteRow))
	mux.HandleFunc("POST /preview", h.RequireEditor(h.StartPreview))
	mux.HandleFunc("POST /preview/stop", h.StopPreview)
	mux.HandleFunc("POST /api/reorder", h.RequireEditor(h.Reorder))
	mux.HandleFunc("POST /api/{section}/reorder-pages", h.RequireEditor(h.ReorderPages))
	mux.HandleFunc("GET /sections/{section}/edit", h.RequireEditor(h.EditSectionForm))
	mux.HandleFunc("POST /sections/{section}/delete", h.RequireEditor(h.DeleteSection))
	mux.HandleFunc("POST /sections/{section}", h.RequireEditor(h.UpdateSection))
	mux.HandleFunc("GET /sections/{section}/pages/new", h.RequireEditor(h.NewPageForm))
	mux.HandleFunc("POST /sections/{section}/pages/new", h.RequireEditor(h.CreatePage))
	// Admin routes
	mux.HandleFunc("GET /admin/{$}", h.RequireAdmin(h.AdminIndex))
	mux.HandleFunc("GET /admin/users", h.RequireAdmin(h.AdminUsers))
	mux.HandleFunc("GET /admin/users/new", h.RequireAdmin(h.AdminNewUserForm))
	mux.HandleFunc("POST /admin/users", h.RequireAdmin(h.AdminCreateUser))
	mux.HandleFunc("GET /admin/users/{id}/edit", h.RequireAdmin(h.AdminEditUserForm))
	mux.HandleFunc("POST /admin/users/{id}/update", h.RequireAdmin(h.AdminUpdateUser))
	mux.HandleFunc("POST /admin/users/{id}/reset-password", h.RequireAdmin(h.AdminSendResetPassword))
	mux.HandleFunc("GET /admin/roles", h.RequireAdmin(h.AdminRoles))
	mux.HandleFunc("GET /admin/roles/new", h.RequireAdmin(h.AdminNewRoleForm))
	mux.HandleFunc("POST /admin/roles", h.RequireAdmin(h.AdminCreateRole))
	mux.HandleFunc("GET /admin/roles/{id}/edit", h.RequireAdmin(h.AdminEditRoleForm))
	mux.HandleFunc("POST /admin/roles/{id}/update", h.RequireAdmin(h.AdminUpdateRole))
	mux.HandleFunc("GET /admin/images", h.RequireAdmin(h.AdminImages))
	mux.HandleFunc("GET /admin/data", h.RequireAdmin(h.AdminDataPage))
	mux.HandleFunc("GET /admin/data/export", h.RequireAdmin(h.AdminExport))
	mux.HandleFunc("POST /admin/data/import", h.RequireAdmin(h.AdminImport))

	mux.HandleFunc("GET /{section}/{slug}/edit", h.RequireEditor(h.EditPage))
	mux.HandleFunc("POST /{section}/{slug}/preview", h.PreviewPage)
	mux.HandleFunc("POST /{section}/{slug}/delete", h.RequireEditor(h.DeletePage))
	mux.HandleFunc("POST /{section}/{slug}", h.RequireEditor(h.SavePage))
	mux.HandleFunc("GET /{section}/{slug}", h.Page)
	mux.HandleFunc("GET /{section}/{$}", h.Section)

	addr := ":" + config.Port()
	slog.Info("HTTP server started", "addr", addr)
	if err := http.ListenAndServe(addr, h.RequireAuth(mux)); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
