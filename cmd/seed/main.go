package main

import (
	"bufio"
	"bytes"
	"context"
	"io/fs"
	"log/slog"
	"os"
	"sort"
	"strings"

	"docgen"
	"docgen/config"
	"docgen/internal/db"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type sectionDef struct {
	ID          string
	Title       string
	Description string
	SortOrder   int
}

var sections = []sectionDef{
	{
		ID:          "space-weather-api",
		Title:       "Space Weather API",
		Description: "REST API for querying real-time and historical space weather data including solar flares, geomagnetic storms, and coronal mass ejections.",
		SortOrder:   0,
	},
	{
		ID:          "alert-system",
		Title:       "Alert System",
		Description: "Configure and manage alerts for space weather events with customizable thresholds, delivery channels, and escalation rules.",
		SortOrder:   1,
	},
	{
		ID:          "data-feeds",
		Title:       "Data Feeds",
		Description: "Real-time streaming and historical bulk data feeds for solar activity, magnetosphere readings, and aurora forecasts.",
		SortOrder:   2,
	},
}

func main() {
	config.InitLogging()
	ctx := context.Background()

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

	// Run migrations
	migrationsFS := docgen.ResolveFS(config.MigrationsDir(), docgen.EmbeddedMigrations())
	d, err := iofs.New(migrationsFS, ".")
	if err != nil {
		slog.Error("failed to create migration source", "error", err)
		os.Exit(1)
	}
	connStr := config.PostgreSQLConnString()
	m, err := migrate.NewWithSourceInstance("iofs", d, "pgx5://"+connStr[len("postgres://"):]+"&x-migrations-table=simpledoc_version")
	if err != nil {
		slog.Error("failed to initialize migrations", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	// Ensure site_settings row exists
	_, err = pool.Exec(ctx, `INSERT INTO site_settings (id) VALUES (1) ON CONFLICT DO NOTHING`)
	if err != nil {
		slog.Error("failed to ensure site_settings", "error", err)
		os.Exit(1)
	}

	// Ensure default roles exist
	_, err = pool.Exec(ctx,
		`INSERT INTO roles (name, description) VALUES
			('admin', 'Full access to all features'),
			('editor', 'Can edit content')
		 ON CONFLICT (name) DO NOTHING`)
	if err != nil {
		slog.Error("failed to ensure default roles", "error", err)
		os.Exit(1)
	}

	contentFS := docgen.ResolveFS(config.ContentDir(), docgen.EmbeddedContent())
	staticFS := docgen.ResolveFS(config.StaticDir(), docgen.EmbeddedStatic())

	// Upsert sections
	for _, s := range sections {
		_, err := pool.Exec(ctx,
			`INSERT INTO sections (id, title, description, sort_order)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (id) DO UPDATE SET title=$2, description=$3, sort_order=$4, updated_at=now()`,
			s.ID, s.Title, s.Description, s.SortOrder)
		if err != nil {
			slog.Error("failed to upsert section", "section", s.ID, "error", err)
			os.Exit(1)
		}
		slog.Info("section created", "id", s.ID)
	}

	// Upsert pages
	totalPages := 0
	for _, s := range sections {
		entries, err := fs.ReadDir(contentFS, s.ID)
		if err != nil {
			slog.Error("failed to read content dir", "section", s.ID, "error", err)
			os.Exit(1)
		}

		var filenames []string
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, "_") || !strings.HasSuffix(name, ".md") {
				continue
			}
			filenames = append(filenames, name)
		}
		sort.Strings(filenames)

		for i, name := range filenames {
			data, err := fs.ReadFile(contentFS, s.ID+"/"+name)
			if err != nil {
				slog.Error("failed to read page file", "file", name, "error", err)
				os.Exit(1)
			}

			title, body := parseFrontMatter(data)
			if title == "" {
				title = strings.TrimSuffix(name, ".md")
			}

			slug := strings.TrimSuffix(name, ".md")

			_, err = pool.Exec(ctx,
				`INSERT INTO pages (section_id, slug, title, content_md, sort_order)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (section_id, slug) WHERE deleted = false DO UPDATE SET title=$3, content_md=$4, sort_order=$5, parent_slug=NULL, updated_at=now()`,
				s.ID, slug, title, string(body), i)
			if err != nil {
				slog.Error("failed to upsert page", "section", s.ID, "slug", slug, "error", err)
				os.Exit(1)
			}
			slog.Info("page created", "section", s.ID, "slug", slug, "title", title)
			totalPages++
		}
	}

	// Build image -> section mapping by scanning markdown content
	imageSectionMap := map[string]string{}
	for _, s := range sections {
		entries, err := fs.ReadDir(contentFS, s.ID)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			md, err := fs.ReadFile(contentFS, s.ID+"/"+e.Name())
			if err != nil {
				continue
			}
			content := string(md)
			// Find image references like static/images/filename.ext
			for _, line := range strings.Split(content, "\n") {
				if idx := strings.Index(line, "static/images/"); idx >= 0 {
					rest := line[idx+len("static/images/"):]
					// Extract filename (until closing paren or whitespace)
					end := strings.IndexAny(rest, ") \t\n")
					if end > 0 {
						imageSectionMap[rest[:end]] = s.ID
					}
				}
			}
		}
	}

	// Upsert images
	imgEntries, err := fs.ReadDir(staticFS, "images")
	if err != nil {
		slog.Error("failed to read images dir", "error", err)
		os.Exit(1)
	}

	totalImages := 0
	for _, e := range imgEntries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		data, err := fs.ReadFile(staticFS, "images/"+name)
		if err != nil {
			slog.Error("failed to read image", "filename", name, "error", err)
			os.Exit(1)
		}

		contentType := "application/octet-stream"
		switch {
		case strings.HasSuffix(name, ".svg"):
			contentType = "image/svg+xml"
		case strings.HasSuffix(name, ".png"):
			contentType = "image/png"
		case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"):
			contentType = "image/jpeg"
		}

		sectionID, ok := imageSectionMap[name]
		if !ok {
			continue
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO images (filename, content_type, data, section_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (filename) DO UPDATE SET content_type=$2, data=$3, section_id=$4`,
			name, contentType, data, sectionID)
		if err != nil {
			slog.Error("failed to upsert image", "filename", name, "error", err)
			os.Exit(1)
		}
		slog.Info("image created", "filename", name, "section", sectionID)
		totalImages++
	}

	// Set required_role on sections
	sectionRoles := map[string]string{
		"space-weather-api": "space weather api",
		"alert-system":      "alert system",
		"data-feeds":        "data feeds",
	}
	for secID, role := range sectionRoles {
		_, err := pool.Exec(ctx,
			`UPDATE sections SET required_role = $2 WHERE id = $1`,
			secID, role)
		if err != nil {
			slog.Error("failed to set required_role", "section", secID, "error", err)
			os.Exit(1)
		}
		slog.Info("section role set", "section", secID, "role", role)
	}

	// Seed admin user
	queries := &db.Queries{Pool: pool}
	adminEmail := "admin@example.com"
	_, err = queries.GetUserByEmail(ctx, adminEmail)
	if err != nil {
		// User doesn't exist, create it
		hash, err := bcrypt.GenerateFromPassword([]byte("changeme"), 12)
		if err != nil {
			slog.Error("failed to hash password", "error", err)
			os.Exit(1)
		}
		user, err := queries.CreateUser(ctx, "Admin", "User", "", adminEmail, string(hash))
		if err != nil {
			slog.Error("failed to create admin user", "error", err)
			os.Exit(1)
		}
		if err := queries.AssignRole(ctx, user.ID, "admin"); err != nil {
			slog.Error("failed to assign admin role", "error", err)
			os.Exit(1)
		}
		if err := queries.AssignRole(ctx, user.ID, "editor"); err != nil {
			slog.Error("failed to assign editor role", "error", err)
			os.Exit(1)
		}
		slog.Info("admin user created", "email", adminEmail)
	} else {
		slog.Info("admin user already exists", "email", adminEmail)
	}

	// Seed partner roles
	partnerRoles := []struct{ Name, Desc string }{
		{"space weather api", "Access to Space Weather API documentation"},
		{"alert system", "Access to Alert System documentation"},
		{"data feeds", "Access to Data Feeds documentation"},
	}
	for _, r := range partnerRoles {
		_, err := pool.Exec(ctx,
			`INSERT INTO roles (name, description) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`,
			r.Name, r.Desc)
		if err != nil {
			slog.Error("failed to upsert role", "role", r.Name, "error", err)
			os.Exit(1)
		}
		slog.Info("role created", "name", r.Name)
	}

	// Seed editor user
	editorEmail := "editor@example.com"
	_, err = queries.GetUserByEmail(ctx, editorEmail)
	if err != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte("changeme"), 12)
		if err != nil {
			slog.Error("failed to hash password", "error", err)
			os.Exit(1)
		}
		u, err := queries.CreateUser(ctx, "Editor", "User", "", editorEmail, string(hash))
		if err != nil {
			slog.Error("failed to create editor user", "error", err)
			os.Exit(1)
		}
		if err := queries.AssignRole(ctx, u.ID, "editor"); err != nil {
			slog.Error("failed to assign editor role", "error", err)
			os.Exit(1)
		}
		for _, r := range partnerRoles {
			if err := queries.AssignRole(ctx, u.ID, r.Name); err != nil {
				slog.Error("failed to assign role to editor", "role", r.Name, "error", err)
				os.Exit(1)
			}
		}
		slog.Info("editor user created", "email", editorEmail)
	} else {
		slog.Info("editor user already exists", "email", editorEmail)
	}

	slog.Info("seed complete", "sections", len(sections), "pages", totalPages, "images", totalImages)
}

// parseFrontMatter extracts title from simple YAML-like front matter.
func parseFrontMatter(data []byte) (string, []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	title := ""
	inFrontMatter := false
	lineCount := 0
	frontMatterEnd := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if lineCount == 1 && strings.TrimSpace(line) == "---" {
			inFrontMatter = true
			frontMatterEnd = len("---\n")
			continue
		}

		if inFrontMatter {
			if strings.TrimSpace(line) == "---" {
				frontMatterEnd += len(line) + 1
				break
			}
			frontMatterEnd += len(line) + 1
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == "title" {
				title = strings.TrimSpace(parts[1])
			}
		}
	}

	if !inFrontMatter {
		return "", data
	}

	if frontMatterEnd > len(data) {
		frontMatterEnd = len(data)
	}
	return title, data[frontMatterEnd:]
}
