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
	SectionRows  []SectionRowExport  `json:"section_rows"`
	Sections     []SectionExport     `json:"sections"`
	Pages        []PageExport        `json:"pages"`
	Images       []ImageExport       `json:"images"`
	SiteSettings *SiteSettingsExport `json:"site_settings"`
}

type SectionRowExport struct {
	ID          int       `json:"id"`
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
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	SortOrder    int       `json:"sort_order"`
	Icon         string    `json:"icon"`
	RowID        *int      `json:"row_id,omitempty"`
	RequiredRole *string   `json:"required_role,omitempty"`
	Deleted      bool      `json:"deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PageExport struct {
	ID        int       `json:"id"`
	SectionID string    `json:"section_id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	ContentMD string    `json:"content_md"`
	SortOrder int       `json:"sort_order"`
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
		Version:    "1.0",
		ExportedAt: time.Now().UTC(),
	}

	deletedFilter := " WHERE deleted = false"
	if includeDeleted {
		deletedFilter = ""
	}

	// Export section_rows
	rows, err := pool.Query(ctx, `SELECT id, title, description, sort_order, version, deleted, created_at, updated_at FROM section_rows`+deletedFilter+` ORDER BY id`)
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
	rows, err = pool.Query(ctx, `SELECT id, title, description, sort_order, icon, row_id, required_role, deleted, created_at, updated_at FROM sections`+deletedFilter+` ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("query sections: %w", err)
	}
	for rows.Next() {
		var s SectionExport
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.SortOrder, &s.Icon, &s.RowID, &s.RequiredRole, &s.Deleted, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan section: %w", err)
		}
		bundle.Sections = append(bundle.Sections, s)
	}
	rows.Close()
	slog.Info("exported sections", "count", len(bundle.Sections))

	// Export pages
	rows, err = pool.Query(ctx, `SELECT id, section_id, slug, title, content_md, sort_order, deleted, created_at, updated_at FROM pages`+deletedFilter+` ORDER BY section_id, sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("query pages: %w", err)
	}
	for rows.Next() {
		var p PageExport
		if err := rows.Scan(&p.ID, &p.SectionID, &p.Slug, &p.Title, &p.ContentMD, &p.SortOrder, &p.Deleted, &p.CreatedAt, &p.UpdatedAt); err != nil {
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
	err = pool.QueryRow(ctx, `SELECT site_title, badge, heading, description, footer, theme, accent_color, version, updated_at FROM site_settings WHERE id = 1`).
		Scan(&ss.SiteTitle, &ss.Badge, &ss.Heading, &ss.Description, &ss.Footer, &ss.Theme, &ss.AccentColor, &ss.Version, &ss.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("query site_settings: %w", err)
	}
	bundle.SiteSettings = &ss
	slog.Info("exported site_settings")

	return bundle, nil
}

// Import writes the given ExportBundle into the database inside a transaction.
func Import(ctx context.Context, pool *pgxpool.Pool, bundle *ExportBundle) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Import section_rows
	for _, sr := range bundle.SectionRows {
		_, err := tx.Exec(ctx,
			`INSERT INTO section_rows (id, title, description, sort_order, version, deleted, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (id) DO UPDATE SET title=$2, description=$3, sort_order=$4, version=$5, deleted=$6, updated_at=$8`,
			sr.ID, sr.Title, sr.Description, sr.SortOrder, sr.Version, sr.Deleted, sr.CreatedAt, sr.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert section_row %d: %w", sr.ID, err)
		}
	}
	slog.Info("imported section_rows", "count", len(bundle.SectionRows))

	// Import sections
	for _, s := range bundle.Sections {
		_, err := tx.Exec(ctx,
			`INSERT INTO sections (id, title, description, sort_order, icon, row_id, required_role, deleted, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 ON CONFLICT (id) DO UPDATE SET title=$2, description=$3, sort_order=$4, icon=$5, row_id=$6, required_role=$7, deleted=$8, updated_at=$10`,
			s.ID, s.Title, s.Description, s.SortOrder, s.Icon, s.RowID, s.RequiredRole, s.Deleted, s.CreatedAt, s.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert section %s: %w", s.ID, err)
		}
	}
	slog.Info("imported sections", "count", len(bundle.Sections))

	// Import pages
	for _, p := range bundle.Pages {
		_, err := tx.Exec(ctx,
			`INSERT INTO pages (id, section_id, slug, title, content_md, sort_order, deleted, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (id) DO UPDATE SET section_id=$2, slug=$3, title=$4, content_md=$5, sort_order=$6, deleted=$7, updated_at=$9`,
			p.ID, p.SectionID, p.Slug, p.Title, p.ContentMD, p.SortOrder, p.Deleted, p.CreatedAt, p.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert page %d: %w", p.ID, err)
		}
	}
	slog.Info("imported pages", "count", len(bundle.Pages))

	// Import images
	for _, img := range bundle.Images {
		imgData, err := base64.StdEncoding.DecodeString(img.DataBase64)
		if err != nil {
			return fmt.Errorf("decode image base64 %s: %w", img.Filename, err)
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO images (filename, content_type, data, section_id, created_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (filename) DO UPDATE SET content_type=$2, data=$3, section_id=$4`,
			img.Filename, img.ContentType, imgData, img.SectionID, img.CreatedAt)
		if err != nil {
			return fmt.Errorf("upsert image %s: %w", img.Filename, err)
		}
	}
	slog.Info("imported images", "count", len(bundle.Images))

	// Import site_settings
	if bundle.SiteSettings != nil {
		ss := bundle.SiteSettings
		_, err := tx.Exec(ctx,
			`INSERT INTO site_settings (id, site_title, badge, heading, description, footer, theme, accent_color, version, updated_at)
			 VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (id) DO UPDATE SET site_title=$1, badge=$2, heading=$3, description=$4, footer=$5, theme=$6, accent_color=$7, version=$8, updated_at=$9`,
			ss.SiteTitle, ss.Badge, ss.Heading, ss.Description, ss.Footer, ss.Theme, ss.AccentColor, ss.Version, ss.UpdatedAt)
		if err != nil {
			return fmt.Errorf("upsert site_settings: %w", err)
		}
		slog.Info("imported site_settings")
	}

	// Reset SERIAL sequences
	sequences := []struct {
		table  string
		column string
	}{
		{"section_rows", "id"},
		{"pages", "id"},
		{"images", "id"},
	}
	for _, seq := range sequences {
		_, err := tx.Exec(ctx, fmt.Sprintf(
			`SELECT setval(pg_get_serial_sequence('%s', '%s'), COALESCE((SELECT MAX(%s) FROM %s), 0) + 1, false)`,
			seq.table, seq.column, seq.column, seq.table))
		if err != nil {
			return fmt.Errorf("reset sequence %s: %w", seq.table, err)
		}
	}
	slog.Info("reset serial sequences")

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("import complete",
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

	rowIDs := map[int]bool{}
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
			return fmt.Errorf("section %s references unknown row_id: %d", s.ID, *s.RowID)
		}
	}

	// Validate pages reference valid sections
	for _, p := range bundle.Pages {
		if !sectionIDs[p.SectionID] {
			return fmt.Errorf("page %d references unknown section_id: %s", p.ID, p.SectionID)
		}
	}

	// Validate images reference valid sections
	for _, img := range bundle.Images {
		if img.SectionID != nil && !sectionIDs[*img.SectionID] {
			return fmt.Errorf("image %s references unknown section_id: %s", img.Filename, *img.SectionID)
		}
	}

	return nil
}
