package portability

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Export bundle types

type ExportBundle struct {
	Version      string              `json:"version"`
	ExportedAt   time.Time           `json:"exported_at"`
	Roles        []RoleExport        `json:"roles"`
	SectionRows  []SectionRowExport  `json:"section_rows"`
	Sections     []SectionExport     `json:"sections"`
	Pages        []PageExport        `json:"pages"`
	Images       []ImageExport       `json:"images"`
	SiteSettings *SiteSettingsExport `json:"site_settings"`
}

type RoleExport struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SectionRowExport struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	Version     int       `json:"version"`
	Deleted     bool      `json:"deleted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SectionExport struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	SortOrder    int       `json:"sort_order"`
	Icon         string    `json:"icon"`
	RowID        *string   `json:"row_id,omitempty"`
	RequiredRole *string   `json:"required_role,omitempty"`
	Deleted      bool      `json:"deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PageExport struct {
	ID         string    `json:"id"`
	SectionID  string    `json:"section_id"`
	Slug       string    `json:"slug"`
	Title      string    `json:"title"`
	ContentMD  string    `json:"content_md"`
	SortOrder  int       `json:"sort_order"`
	ParentSlug *string   `json:"parent_slug,omitempty"`
	Deleted    bool      `json:"deleted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ImageExport struct {
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	DataBase64  string    `json:"data_base64"`
	SectionID   *string   `json:"section_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type SiteSettingsExport struct {
	SiteTitle   string    `json:"site_title"`
	Badge       string    `json:"badge"`
	Heading     string    `json:"heading"`
	Description string    `json:"description"`
	Footer      string    `json:"footer"`
	Theme       string    `json:"theme"`
	AccentColor string    `json:"accent_color"`
	Version     int       `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Export reads site data from the database and returns an ExportBundle.
func Export(ctx context.Context, pool *pgxpool.Pool, includeDeleted bool) (*ExportBundle, error) {
	bundle := &ExportBundle{
		Version:    "2.0",
		ExportedAt: time.Now().UTC(),
	}

	deletedFilter := " WHERE deleted = false"
	if includeDeleted {
		deletedFilter = ""
	}

	// Export roles
	rows, err := pool.Query(ctx, `SELECT id, name, description, created_at, updated_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	for rows.Next() {
		var r RoleExport
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		bundle.Roles = append(bundle.Roles, r)
	}
	rows.Close()
	slog.Info("exported roles", "count", len(bundle.Roles))

	// Export section_rows
	rows, err = pool.Query(ctx, `SELECT id, title, description, sort_order, version, deleted, created_at, updated_at FROM section_rows`+deletedFilter+` ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query section_rows: %w", err)
	}
	for rows.Next() {
		var sr SectionRowExport
		if err := rows.Scan(&sr.ID, &sr.Title, &sr.Description, &sr.SortOrder, &sr.Version, &sr.Deleted, &sr.CreatedAt, &sr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan section_row: %w", err)
		}
		bundle.SectionRows = append(bundle.SectionRows, sr)
	}
	rows.Close()
	slog.Info("exported section_rows", "count", len(bundle.SectionRows))

	// Export sections
	rows, err = pool.Query(ctx, `SELECT id, name, title, description, sort_order, icon, row_id, required_role, deleted, created_at, updated_at FROM sections`+deletedFilter+` ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("query sections: %w", err)
	}
	for rows.Next() {
		var s SectionExport
		if err := rows.Scan(&s.ID, &s.Name, &s.Title, &s.Description, &s.SortOrder, &s.Icon, &s.RowID, &s.RequiredRole, &s.Deleted, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan section: %w", err)
		}
		bundle.Sections = append(bundle.Sections, s)
	}
	rows.Close()
	slog.Info("exported sections", "count", len(bundle.Sections))

	// Export pages
	rows, err = pool.Query(ctx, `SELECT id, section_id, slug, title, content_md, sort_order, parent_slug, deleted, created_at, updated_at FROM pages`+deletedFilter+` ORDER BY section_id, sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("query pages: %w", err)
	}
	for rows.Next() {
		var p PageExport
		if err := rows.Scan(&p.ID, &p.SectionID, &p.Slug, &p.Title, &p.ContentMD, &p.SortOrder, &p.ParentSlug, &p.Deleted, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan page: %w", err)
		}
		bundle.Pages = append(bundle.Pages, p)
	}
	rows.Close()
	slog.Info("exported pages", "count", len(bundle.Pages))

	// Export images
	rows, err = pool.Query(ctx, `SELECT filename, content_type, data, section_id, created_at FROM images ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query images: %w", err)
	}
	for rows.Next() {
		var img ImageExport
		var data []byte
		if err := rows.Scan(&img.Filename, &img.ContentType, &data, &img.SectionID, &img.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan image: %w", err)
		}
		img.DataBase64 = base64.StdEncoding.EncodeToString(data)
		bundle.Images = append(bundle.Images, img)
	}
	rows.Close()
	slog.Info("exported images", "count", len(bundle.Images))

	// Export site_settings
	var ss SiteSettingsExport
	err = pool.QueryRow(ctx, `SELECT site_title, badge, heading, description, footer, theme, accent_color, version, updated_at FROM site_settings WHERE singleton = TRUE`).
		Scan(&ss.SiteTitle, &ss.Badge, &ss.Heading, &ss.Description, &ss.Footer, &ss.Theme, &ss.AccentColor, &ss.Version, &ss.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("query site_settings: %w", err)
	}
	bundle.SiteSettings = &ss
	slog.Info("exported site_settings")

	return bundle, nil
}

// Import writes the given ExportBundle into the database inside a transaction.
// When clean is true, all existing content is deleted before importing (history is preserved).
func Import(ctx context.Context, pool *pgxpool.Pool, bundle *ExportBundle, clean bool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if clean {
		slog.Info("clean import: deleting existing content")
		cleanQueries := []struct {
			label string
			query string
		}{
			{"pages", "DELETE FROM pages"},
			{"images", "DELETE FROM images"},
			{"sections", "DELETE FROM sections"},
			{"section_rows", "DELETE FROM section_rows"},
			{"site_settings", "DELETE FROM site_settings"},
			{"roles", "DELETE FROM roles WHERE name NOT IN ('admin', 'editor')"},
		}
		for _, q := range cleanQueries {
			if _, err := tx.Exec(ctx, q.query); err != nil {
				return fmt.Errorf("clean delete %s: %w", q.label, err)
			}
			slog.Info("clean import: deleted", "table", q.label)
		}
	}

	// Import roles
	for _, r := range bundle.Roles {
		_, err := tx.Exec(ctx,
			`INSERT INTO roles (id, name, description, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (name) DO UPDATE SET description=$3, updated_at=$5`,
			r.ID, r.Name, r.Description, r.CreatedAt, r.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert role %s: %w", r.Name, err)
		}
	}
	slog.Info("imported roles", "count", len(bundle.Roles))

	// Import section_rows
	for _, sr := range bundle.SectionRows {
		_, err := tx.Exec(ctx,
			`INSERT INTO section_rows (id, title, description, sort_order, version, deleted, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (id) DO UPDATE SET title=$2, description=$3, sort_order=$4, version=$5, deleted=$6, updated_at=$8`,
			sr.ID, sr.Title, sr.Description, sr.SortOrder, sr.Version, sr.Deleted, sr.CreatedAt, sr.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert section_row %s: %w", sr.ID, err)
		}
	}
	slog.Info("imported section_rows", "count", len(bundle.SectionRows))

	// Import sections — use name for conflict resolution, RETURNING id to remap pages/images
	sectionNameToID := make(map[string]string) // name -> new DB id
	for _, s := range bundle.Sections {
		// Backward compat: old exports used slug as ID and had no name field
		name := s.Name
		if name == "" {
			name = s.ID
		}
		var newID string
		err := tx.QueryRow(ctx,
			`INSERT INTO sections (name, title, description, sort_order, icon, row_id, required_role, deleted, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (name) WHERE deleted = false DO UPDATE SET title=$2, description=$3, sort_order=$4, icon=$5, row_id=$6, required_role=$7, deleted=$8, updated_at=$10
			 RETURNING id`,
			name, s.Title, s.Description, s.SortOrder, s.Icon, s.RowID, s.RequiredRole, s.Deleted, s.CreatedAt, s.UpdatedAt).
			Scan(&newID)
		if err != nil {
			return fmt.Errorf("upsert section %s: %w", name, err)
		}
		sectionNameToID[name] = newID
	}
	slog.Info("imported sections", "count", len(bundle.Sections))

	// Build export section ID -> name map for remapping pages/images
	exportIDToName := make(map[string]string)
	for _, s := range bundle.Sections {
		name := s.Name
		if name == "" {
			name = s.ID
		}
		exportIDToName[s.ID] = name
	}

	// Import pages — remap section_id through exportIDToName -> sectionNameToID
	for _, p := range bundle.Pages {
		name := exportIDToName[p.SectionID]
		newSectionID := sectionNameToID[name]
		if newSectionID == "" {
			return fmt.Errorf("page %s references unknown section_id: %s", p.ID, p.SectionID)
		}
		// Remove any existing page with same section_id+slug but different id to avoid unique constraint violation
		if _, err := tx.Exec(ctx, `DELETE FROM pages WHERE section_id = $1 AND slug = $2 AND id != $3`, newSectionID, p.Slug, p.ID); err != nil {
			return fmt.Errorf("clean conflicting page %s/%s: %w", newSectionID, p.Slug, err)
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO pages (id, section_id, slug, title, content_md, sort_order, parent_slug, deleted, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (id) DO UPDATE SET section_id=$2, slug=$3, title=$4, content_md=$5, sort_order=$6, parent_slug=$7, deleted=$8, updated_at=$10`,
			p.ID, newSectionID, p.Slug, p.Title, p.ContentMD, p.SortOrder, p.ParentSlug, p.Deleted, p.CreatedAt, p.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert page %s: %w", p.ID, err)
		}
	}
	slog.Info("imported pages", "count", len(bundle.Pages))

	// Import images — remap section_id
	for _, img := range bundle.Images {
		imgData, err := base64.StdEncoding.DecodeString(img.DataBase64)
		if err != nil {
			return fmt.Errorf("decode image base64 %s: %w", img.Filename, err)
		}
		var sectionID *string
		if img.SectionID != nil {
			if name, ok := exportIDToName[*img.SectionID]; ok {
				if id, ok := sectionNameToID[name]; ok {
					sectionID = &id
				}
			}
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO images (filename, content_type, data, section_id, created_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (filename) DO UPDATE SET content_type=$2, data=$3, section_id=$4`,
			img.Filename, img.ContentType, imgData, sectionID, img.CreatedAt)
		if err != nil {
			return fmt.Errorf("upsert image %s: %w", img.Filename, err)
		}
	}
	slog.Info("imported images", "count", len(bundle.Images))

	// Import site_settings
	if bundle.SiteSettings != nil {
		ss := bundle.SiteSettings
		_, err := tx.Exec(ctx,
			`INSERT INTO site_settings (singleton, site_title, badge, heading, description, footer, theme, accent_color, version, updated_at)
			 VALUES (TRUE, $1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (singleton) DO UPDATE SET site_title=$1, badge=$2, heading=$3, description=$4, footer=$5, theme=$6, accent_color=$7, version=$8, updated_at=$9`,
			ss.SiteTitle, ss.Badge, ss.Heading, ss.Description, ss.Footer, ss.Theme, ss.AccentColor, ss.Version, ss.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert site_settings: %w", err)
		}
		slog.Info("imported site_settings")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("import complete",
		"roles", len(bundle.Roles),
		"section_rows", len(bundle.SectionRows),
		"sections", len(bundle.Sections),
		"pages", len(bundle.Pages),
		"images", len(bundle.Images),
	)

	return nil
}

// Validate checks FK reference integrity within the bundle.
func Validate(bundle *ExportBundle) error {
	if bundle.Version == "" {
		return fmt.Errorf("missing version field")
	}

	// Backfill name from ID for old exports
	for i := range bundle.Sections {
		if bundle.Sections[i].Name == "" {
			bundle.Sections[i].Name = bundle.Sections[i].ID
		}
	}

	rowIDs := map[string]bool{}
	for _, sr := range bundle.SectionRows {
		rowIDs[sr.ID] = true
	}

	sectionIDs := map[string]bool{}
	for _, s := range bundle.Sections {
		sectionIDs[s.ID] = true
	}

	// Validate sections reference valid row_ids
	for _, s := range bundle.Sections {
		if s.RowID != nil && !rowIDs[*s.RowID] {
			return fmt.Errorf("section %s references unknown row_id: %s", s.ID, *s.RowID)
		}
	}

	// Validate pages reference valid sections
	for _, p := range bundle.Pages {
		if !sectionIDs[p.SectionID] {
			return fmt.Errorf("page %s references unknown section_id: %s", p.ID, p.SectionID)
		}
	}

	// Null out image section_ids that reference missing sections
	for i := range bundle.Images {
		if bundle.Images[i].SectionID != nil && !sectionIDs[*bundle.Images[i].SectionID] {
			bundle.Images[i].SectionID = nil
		}
	}

	return nil
}
