package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
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
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, config.PostgreSQLConnString())
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	// Run migrations
	migrationsFS := docgen.ResolveFS(config.MigrationsDir(), docgen.EmbeddedMigrations())
	d, err := iofs.New(migrationsFS, ".")
	if err != nil {
		log.Fatalf("migrate source: %v", err)
	}
	connStr := config.PostgreSQLConnString()
	m, err := migrate.NewWithSourceInstance("iofs", d, "pgx5://"+connStr[len("postgres://"):]+"&x-migrations-table=simpledoc_migrations")
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate up: %v", err)
	}
	log.Println("Migrations applied")

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
			log.Fatalf("upsert section %s: %v", s.ID, err)
		}
		fmt.Printf("Section: %s\n", s.ID)
	}

	// Upsert pages
	totalPages := 0
	for _, s := range sections {
		entries, err := fs.ReadDir(contentFS, s.ID)
		if err != nil {
			log.Fatalf("read dir %s: %v", s.ID, err)
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
				log.Fatalf("read %s: %v", name, err)
			}

			title, body := parseFrontMatter(data)
			if title == "" {
				title = strings.TrimSuffix(name, ".md")
			}

			slug := strings.TrimSuffix(name, ".md")

			_, err = pool.Exec(ctx,
				`INSERT INTO pages (section_id, slug, title, content_md, sort_order)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (section_id, slug) WHERE deleted = false DO UPDATE SET title=$3, content_md=$4, sort_order=$5, updated_at=now()`,
				s.ID, slug, title, string(body), i)
			if err != nil {
				log.Fatalf("upsert page %s/%s: %v", s.ID, slug, err)
			}
			fmt.Printf("  Page: %s/%s (%s)\n", s.ID, slug, title)
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
		log.Fatalf("read images dir: %v", err)
	}

	totalImages := 0
	for _, e := range imgEntries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		data, err := fs.ReadFile(staticFS, "images/"+name)
		if err != nil {
			log.Fatalf("read image %s: %v", name, err)
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

		sectionID := imageSectionMap[name]

		_, err = pool.Exec(ctx,
			`INSERT INTO images (filename, content_type, data, section_id)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (filename) DO UPDATE SET content_type=$2, data=$3, section_id=$4`,
			name, contentType, data, sectionID)
		if err != nil {
			log.Fatalf("upsert image %s: %v", name, err)
		}
		fmt.Printf("  Image: %s (section: %s)\n", name, sectionID)
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
			log.Fatalf("set required_role on %s: %v", secID, err)
		}
		fmt.Printf("Section %s: required_role = %s\n", secID, role)
	}

	// Seed admin user
	queries := &db.Queries{Pool: pool}
	adminEmail := "admin@example.com"
	_, err = queries.GetUserByEmail(ctx, adminEmail)
	if err != nil {
		// User doesn't exist, create it
		hash, err := bcrypt.GenerateFromPassword([]byte("changeme"), 12)
		if err != nil {
			log.Fatalf("bcrypt hash: %v", err)
		}
		user, err := queries.CreateUser(ctx, "Admin", "User", "", adminEmail, string(hash))
		if err != nil {
			log.Fatalf("create admin user: %v", err)
		}
		if err := queries.AssignRole(ctx, user.ID, "admin"); err != nil {
			log.Fatalf("assign admin role: %v", err)
		}
		if err := queries.AssignRole(ctx, user.ID, "editor"); err != nil {
			log.Fatalf("assign editor role: %v", err)
		}
		fmt.Printf("Admin user created: %s (password: changeme)\n", adminEmail)
	} else {
		fmt.Printf("Admin user already exists: %s\n", adminEmail)
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
			log.Fatalf("upsert role %s: %v", r.Name, err)
		}
		fmt.Printf("Role: %s\n", r.Name)
	}

	// Seed editor user
	editorEmail := "editor@example.com"
	_, err = queries.GetUserByEmail(ctx, editorEmail)
	if err != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte("changeme"), 12)
		if err != nil {
			log.Fatalf("bcrypt hash: %v", err)
		}
		u, err := queries.CreateUser(ctx, "Editor", "User", "", editorEmail, string(hash))
		if err != nil {
			log.Fatalf("create editor user: %v", err)
		}
		if err := queries.AssignRole(ctx, u.ID, "editor"); err != nil {
			log.Fatalf("assign editor role: %v", err)
		}
		for _, r := range partnerRoles {
			if err := queries.AssignRole(ctx, u.ID, r.Name); err != nil {
				log.Fatalf("assign role %s to editor: %v", r.Name, err)
			}
		}
		fmt.Printf("Editor user created: %s (password: changeme)\n", editorEmail)
	} else {
		fmt.Printf("Editor user already exists: %s\n", editorEmail)
	}

	fmt.Printf("\nSeed complete: %d sections, %d pages, %d images\n",
		len(sections), totalPages, totalImages)
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
