package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Section struct {
	ID           string
	Name         string
	Title        string
	Description  string
	Icon         string
	SortOrder    int
	Version      int
	RequiredRole string
	RowID        *string
}

type SectionRow struct {
	ID          string
	Title       string
	Description string
	SortOrder   int
	Version     int
}

type Page struct {
	ID         string
	SectionID  string
	Slug       string
	Title      string
	ContentMD  string
	SortOrder  int
	Version    int
	ParentSlug *string
}

type PageOrderItem struct {
	Slug     string   `json:"slug"`
	Children []string `json:"children"`
}

type PageHistory struct {
	ID        string
	PageID    string
	Version   int
	SectionID string
	Slug      string
	Title     string
	ContentMD string
	SortOrder int
	ChangedAt time.Time
}

type Image struct {
	ID          string
	Filename    string
	ContentType string
	Data        []byte
	SectionID   string
	CreatedAt   time.Time
	Version     int
}

type ImageMeta struct {
	ID          string
	Filename    string
	ContentType string
	Size        int64
	SectionID   string
	CreatedAt   time.Time
	Version     int
}

type ImageMetaWithSection struct {
	ImageMeta
	SectionTitle string
}

type User struct {
	ID        string
	Firstname string
	Lastname  string
	Company   string
	Email     string
	Password  string
	LastLogin *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Role struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Session struct {
	ID           string
	UserID       string
	Token        string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	PreviewRoles *string
}

type SiteSettings struct {
	SiteTitle   string
	Badge       string
	Heading     string
	Description string
	Footer      string
	Theme       string
	AccentColor string
	Version     int
}

type UserWithRoles struct {
	User
	Roles []string
}

type PasswordResetToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Queries struct {
	Pool *pgxpool.Pool
}

func (q *Queries) ListSections(ctx context.Context) ([]Section, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT id, name, title, description, icon, sort_order, version, COALESCE(required_role, ''), row_id FROM sections WHERE deleted = false ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []Section
	for rows.Next() {
		var s Section
		if err := rows.Scan(&s.ID, &s.Name, &s.Title, &s.Description, &s.Icon, &s.SortOrder, &s.Version, &s.RequiredRole, &s.RowID); err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}
	return sections, rows.Err()
}

func (q *Queries) GetSection(ctx context.Context, id string) (Section, error) {
	var s Section
	err := q.Pool.QueryRow(ctx,
		`SELECT id, name, title, description, icon, sort_order, version, COALESCE(required_role, ''), row_id FROM sections WHERE id = $1 AND deleted = false`, id).
		Scan(&s.ID, &s.Name, &s.Title, &s.Description, &s.Icon, &s.SortOrder, &s.Version, &s.RequiredRole, &s.RowID)
	return s, err
}

func (q *Queries) GetSectionByName(ctx context.Context, name string) (Section, error) {
	var s Section
	err := q.Pool.QueryRow(ctx,
		`SELECT id, name, title, description, icon, sort_order, version, COALESCE(required_role, ''), row_id FROM sections WHERE name = $1 AND deleted = false`, name).
		Scan(&s.ID, &s.Name, &s.Title, &s.Description, &s.Icon, &s.SortOrder, &s.Version, &s.RequiredRole, &s.RowID)
	return s, err
}

func (q *Queries) ListPagesBySection(ctx context.Context, sectionID string) ([]Page, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT id, section_id, slug, title, content_md, sort_order, version, parent_slug
		 FROM pages WHERE section_id = $1 AND deleted = false ORDER BY sort_order`, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []Page
	for rows.Next() {
		var p Page
		if err := rows.Scan(&p.ID, &p.SectionID, &p.Slug, &p.Title, &p.ContentMD, &p.SortOrder, &p.Version, &p.ParentSlug); err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	return pages, rows.Err()
}

func (q *Queries) GetPage(ctx context.Context, sectionID, slug string) (Page, error) {
	var p Page
	err := q.Pool.QueryRow(ctx,
		`SELECT id, section_id, slug, title, content_md, sort_order, version, parent_slug
		 FROM pages WHERE section_id = $1 AND slug = $2 AND deleted = false`, sectionID, slug).
		Scan(&p.ID, &p.SectionID, &p.Slug, &p.Title, &p.ContentMD, &p.SortOrder, &p.Version, &p.ParentSlug)
	return p, err
}

func (q *Queries) GetFirstPage(ctx context.Context, sectionID string) (Page, error) {
	var p Page
	err := q.Pool.QueryRow(ctx,
		`SELECT id, section_id, slug, title, content_md, sort_order, version, parent_slug
		 FROM pages WHERE section_id = $1 AND deleted = false AND parent_slug IS NULL ORDER BY sort_order LIMIT 1`, sectionID).
		Scan(&p.ID, &p.SectionID, &p.Slug, &p.Title, &p.ContentMD, &p.SortOrder, &p.Version, &p.ParentSlug)
	return p, err
}

func (q *Queries) GetImage(ctx context.Context, filename string) (Image, error) {
	var img Image
	err := q.Pool.QueryRow(ctx,
		`SELECT id, filename, content_type, data, COALESCE(section_id, ''), created_at, version
		 FROM images WHERE filename = $1`, filename).
		Scan(&img.ID, &img.Filename, &img.ContentType, &img.Data, &img.SectionID, &img.CreatedAt, &img.Version)
	return img, err
}

func (q *Queries) ListImageMetasBySection(ctx context.Context, sectionID string) ([]ImageMeta, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT id, filename, content_type, length(data), COALESCE(section_id, ''), created_at, version
		 FROM images WHERE section_id = $1 ORDER BY filename`, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metas []ImageMeta
	for rows.Next() {
		var m ImageMeta
		if err := rows.Scan(&m.ID, &m.Filename, &m.ContentType, &m.Size, &m.SectionID, &m.CreatedAt, &m.Version); err != nil {
			return nil, err
		}
		metas = append(metas, m)
	}
	return metas, rows.Err()
}

func (q *Queries) ListAllImageMetas(ctx context.Context) ([]ImageMetaWithSection, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT i.id, i.filename, i.content_type, length(i.data), COALESCE(i.section_id, ''), i.created_at, i.version, COALESCE(s.title, '')
		 FROM images i LEFT JOIN sections s ON s.id = i.section_id ORDER BY i.filename`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metas []ImageMetaWithSection
	for rows.Next() {
		var m ImageMetaWithSection
		if err := rows.Scan(&m.ID, &m.Filename, &m.ContentType, &m.Size, &m.SectionID, &m.CreatedAt, &m.Version, &m.SectionTitle); err != nil {
			return nil, err
		}
		metas = append(metas, m)
	}
	return metas, rows.Err()
}

func (q *Queries) CreateImage(ctx context.Context, filename, contentType string, data []byte, sectionID, changedBy string) (Image, error) {
	var img Image
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO images (filename, content_type, data, section_id, changed_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, filename, content_type, data, COALESCE(section_id, ''), created_at, version`,
		filename, contentType, data, sectionID, changedBy).
		Scan(&img.ID, &img.Filename, &img.ContentType, &img.Data, &img.SectionID, &img.CreatedAt, &img.Version)
	return img, err
}

func (q *Queries) UpdateImage(ctx context.Context, filename, contentType string, data []byte, changedBy string) (Image, error) {
	var img Image
	err := q.Pool.QueryRow(ctx,
		`UPDATE images
		 SET content_type = $2, data = $3, version = version + 1, updated_at = now(), changed_by = $4
		 WHERE filename = $1
		 RETURNING id, filename, content_type, data, COALESCE(section_id, ''), created_at, version`,
		filename, contentType, data, changedBy).
		Scan(&img.ID, &img.Filename, &img.ContentType, &img.Data, &img.SectionID, &img.CreatedAt, &img.Version)
	return img, err
}

func (q *Queries) RenameImage(ctx context.Context, oldFilename, newFilename, changedBy string) (Image, error) {
	var img Image
	err := q.Pool.QueryRow(ctx,
		`UPDATE images
		 SET filename = $2, version = version + 1, updated_at = now(), changed_by = $3
		 WHERE filename = $1
		 RETURNING id, filename, content_type, data, COALESCE(section_id, ''), created_at, version`,
		oldFilename, newFilename, changedBy).
		Scan(&img.ID, &img.Filename, &img.ContentType, &img.Data, &img.SectionID, &img.CreatedAt, &img.Version)
	return img, err
}

func (q *Queries) DeleteImage(ctx context.Context, filename string) error {
	_, err := q.Pool.Exec(ctx,
		`DELETE FROM images WHERE filename = $1`, filename)
	return err
}

func (q *Queries) SaveImageHistory(ctx context.Context, img Image, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO images_history (image_id, version, filename, content_type, data, created_at, changed_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		img.ID, img.Version, img.Filename, img.ContentType, img.Data, img.CreatedAt, changedBy)
	return err
}

func (q *Queries) UpdatePage(ctx context.Context, sectionID, slug, title, contentMD, changedBy string) (Page, error) {
	var p Page
	err := q.Pool.QueryRow(ctx,
		`UPDATE pages
		 SET title = $3, content_md = $4, version = version + 1, updated_at = now(), changed_by = $5
		 WHERE section_id = $1 AND slug = $2
		 RETURNING id, section_id, slug, title, content_md, sort_order, version, parent_slug`,
		sectionID, slug, title, contentMD, changedBy).
		Scan(&p.ID, &p.SectionID, &p.Slug, &p.Title, &p.ContentMD, &p.SortOrder, &p.Version, &p.ParentSlug)
	return p, err
}

func (q *Queries) CreatePage(ctx context.Context, sectionID, slug, title, contentMD string, sortOrder int, changedBy string) (Page, error) {
	var p Page
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO pages (section_id, slug, title, content_md, sort_order, changed_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, section_id, slug, title, content_md, sort_order, version, parent_slug`,
		sectionID, slug, title, contentMD, sortOrder, changedBy).
		Scan(&p.ID, &p.SectionID, &p.Slug, &p.Title, &p.ContentMD, &p.SortOrder, &p.Version, &p.ParentSlug)
	return p, err
}

func (q *Queries) SavePageHistory(ctx context.Context, p Page, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO pages_history (page_id, version, section_id, slug, title, content_md, sort_order, changed_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		p.ID, p.Version, p.SectionID, p.Slug, p.Title, p.ContentMD, p.SortOrder, changedBy)
	return err
}

func (q *Queries) CreateSection(ctx context.Context, name, title, description, icon string, sortOrder int, requiredRole, changedBy string, rowID *string) (Section, error) {
	var s Section
	// If a soft-deleted section with this name exists, reactivate it
	err := q.Pool.QueryRow(ctx,
		`UPDATE sections
		 SET title = $2, description = $3, icon = $4, sort_order = $5, required_role = NULLIF($6, ''),
		     changed_by = $7, row_id = $8, deleted = false, version = version + 1, updated_at = now()
		 WHERE name = $1 AND deleted = true
		 RETURNING id, name, title, description, icon, sort_order, version, COALESCE(required_role, ''), row_id`,
		name, title, description, icon, sortOrder, requiredRole, changedBy, rowID).
		Scan(&s.ID, &s.Name, &s.Title, &s.Description, &s.Icon, &s.SortOrder, &s.Version, &s.RequiredRole, &s.RowID)
	if err == nil {
		return s, nil
	}
	// Otherwise insert fresh (id auto-generated)
	err = q.Pool.QueryRow(ctx,
		`INSERT INTO sections (name, title, description, icon, sort_order, required_role, changed_by, row_id)
		 VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8)
		 RETURNING id, name, title, description, icon, sort_order, version, COALESCE(required_role, ''), row_id`,
		name, title, description, icon, sortOrder, requiredRole, changedBy, rowID).
		Scan(&s.ID, &s.Name, &s.Title, &s.Description, &s.Icon, &s.SortOrder, &s.Version, &s.RequiredRole, &s.RowID)
	return s, err
}

func (q *Queries) UpdateSection(ctx context.Context, id, title, description, icon, requiredRole, changedBy string) (Section, error) {
	var s Section
	err := q.Pool.QueryRow(ctx,
		`UPDATE sections
		 SET title = $2, description = $3, icon = $4, required_role = NULLIF($5, ''),
		     version = version + 1, updated_at = now(), changed_by = $6
		 WHERE id = $1
		 RETURNING id, name, title, description, icon, sort_order, version, COALESCE(required_role, ''), row_id`,
		id, title, description, icon, requiredRole, changedBy).
		Scan(&s.ID, &s.Name, &s.Title, &s.Description, &s.Icon, &s.SortOrder, &s.Version, &s.RequiredRole, &s.RowID)
	return s, err
}

func (q *Queries) SaveSectionHistory(ctx context.Context, s Section, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO sections_history (section_id, version, title, description, icon, sort_order, required_role, changed_by, row_id)
		 VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9)`,
		s.ID, s.Version, s.Title, s.Description, s.Icon, s.SortOrder, s.RequiredRole, changedBy, s.RowID)
	return err
}

func (q *Queries) SoftDeleteSection(ctx context.Context, id, changedBy string) error {
	tx, err := q.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE pages SET deleted = true, version = version + 1, updated_at = now(), changed_by = $2
		 WHERE section_id = $1 AND deleted = false`, id, changedBy)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE sections SET deleted = true, version = version + 1, updated_at = now(), changed_by = $2
		 WHERE id = $1`, id, changedBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (q *Queries) SoftDeletePage(ctx context.Context, sectionID, slug, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE pages SET deleted = true, version = version + 1, updated_at = now(), changed_by = $3
		 WHERE section_id = $1 AND slug = $2`, sectionID, slug, changedBy)
	return err
}

func (q *Queries) GetSiteSettings(ctx context.Context) (SiteSettings, error) {
	var s SiteSettings
	err := q.Pool.QueryRow(ctx,
		`SELECT site_title, badge, heading, description, footer, theme, accent_color, version FROM site_settings WHERE singleton = TRUE`).
		Scan(&s.SiteTitle, &s.Badge, &s.Heading, &s.Description, &s.Footer, &s.Theme, &s.AccentColor, &s.Version)
	if err != nil {
		return SiteSettings{
			SiteTitle:   "SolarFlux Documentation",
			Badge:       "API Documentation",
			Heading:     "SolarFlux API Docs",
			Description: "Technical documentation for the SolarFlux space weather monitoring platform.",
			Footer:      "SolarFlux Platform",
			Theme:       "midnight",
			AccentColor: "blue",
			Version:     1,
		}, nil
	}
	if s.Theme == "" {
		s.Theme = "midnight"
	}
	if s.AccentColor == "" {
		s.AccentColor = "blue"
	}
	return s, nil
}

func (q *Queries) UpdateSiteSettings(ctx context.Context, siteTitle, badge, heading, description, footer, theme, accentColor, changedBy string) (SiteSettings, error) {
	var s SiteSettings
	err := q.Pool.QueryRow(ctx,
		`UPDATE site_settings
		 SET site_title = $1, badge = $2, heading = $3, description = $4, footer = $5,
		     theme = $6, accent_color = $7, changed_by = $8,
		     version = version + 1, updated_at = now()
		 WHERE singleton = TRUE
		 RETURNING site_title, badge, heading, description, footer, theme, accent_color, version`,
		siteTitle, badge, heading, description, footer, theme, accentColor, changedBy).
		Scan(&s.SiteTitle, &s.Badge, &s.Heading, &s.Description, &s.Footer, &s.Theme, &s.AccentColor, &s.Version)
	return s, err
}

func (q *Queries) SaveSiteSettingsHistory(ctx context.Context, s SiteSettings, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO site_settings_history (version, site_title, badge, heading, description, footer, theme, accent_color, changed_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		s.Version, s.SiteTitle, s.Badge, s.Heading, s.Description, s.Footer, s.Theme, s.AccentColor, changedBy)
	return err
}

// --- Auth queries ---

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := q.Pool.QueryRow(ctx,
		`SELECT id, firstname, lastname, company, email, password, last_login, created_at, updated_at
		 FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Firstname, &u.Lastname, &u.Company, &u.Email, &u.Password, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) GetUserByID(ctx context.Context, id string) (User, error) {
	var u User
	err := q.Pool.QueryRow(ctx,
		`SELECT id, firstname, lastname, company, email, password, last_login, created_at, updated_at
		 FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Firstname, &u.Lastname, &u.Company, &u.Email, &u.Password, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE users SET last_login = now() WHERE id = $1`, userID)
	return err
}

func (q *Queries) CreateSession(ctx context.Context, userID, token string, expiresAt time.Time) (Session, error) {
	var s Session
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO sessions (user_id, token, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token, expires_at, created_at`,
		userID, token, expiresAt).
		Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func (q *Queries) GetSessionByToken(ctx context.Context, token string) (Session, error) {
	var s Session
	err := q.Pool.QueryRow(ctx,
		`SELECT id, user_id, token, expires_at, created_at, preview_roles
		 FROM sessions WHERE token = $1 AND expires_at > now()`, token).
		Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.CreatedAt, &s.PreviewRoles)
	return s, err
}

func (q *Queries) SetSessionPreviewRoles(ctx context.Context, token, roles string) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE sessions SET preview_roles = $2 WHERE token = $1`, token, roles)
	return err
}

func (q *Queries) ClearSessionPreviewRoles(ctx context.Context, token string) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE sessions SET preview_roles = NULL WHERE token = $1`, token)
	return err
}

func (q *Queries) DeleteSession(ctx context.Context, token string) error {
	_, err := q.Pool.Exec(ctx,
		`DELETE FROM sessions WHERE token = $1`, token)
	return err
}

func (q *Queries) DeleteExpiredSessions(ctx context.Context) error {
	_, err := q.Pool.Exec(ctx,
		`DELETE FROM sessions WHERE expires_at <= now()`)
	return err
}

func (q *Queries) CreateUser(ctx context.Context, firstname, lastname, company, email, passwordHash string) (User, error) {
	var u User
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO users (firstname, lastname, company, email, password)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, firstname, lastname, company, email, password, last_login, created_at, updated_at`,
		firstname, lastname, company, email, passwordHash).
		Scan(&u.ID, &u.Firstname, &u.Lastname, &u.Company, &u.Email, &u.Password, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) AssignRole(ctx context.Context, userID, roleName string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id)
		 SELECT $1, id FROM roles WHERE name = $2
		 ON CONFLICT DO NOTHING`, userID, roleName)
	return err
}

func (q *Queries) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT r.name FROM roles r
		 JOIN user_roles ur ON ur.role_id = r.id
		 WHERE ur.user_id = $1 ORDER BY r.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}
	return roles, rows.Err()
}

func (q *Queries) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT id, name, description, created_at, updated_at FROM roles WHERE name NOT IN ('admin', 'editor', 'viewer') ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

func (q *Queries) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	var exists bool
	err := q.Pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = $1 AND r.name = $2
		)`, userID, roleName).Scan(&exists)
	return exists, err
}

// --- Admin queries ---

func (q *Queries) ListUsers(ctx context.Context) ([]UserWithRoles, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT id, firstname, lastname, company, email, password, last_login, created_at, updated_at
		 FROM users ORDER BY firstname, lastname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserWithRoles
	for rows.Next() {
		var u UserWithRoles
		if err := rows.Scan(&u.ID, &u.Firstname, &u.Lastname, &u.Company, &u.Email, &u.Password, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range users {
		roles, err := q.GetUserRoles(ctx, users[i].ID)
		if err != nil {
			return nil, err
		}
		users[i].Roles = roles
	}
	return users, nil
}

// ListNonEditorUsers returns all users that do not have the admin or editor role.
func (q *Queries) ListNonEditorUsers(ctx context.Context) ([]UserWithRoles, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT u.id, u.firstname, u.lastname, u.company, u.email, u.password, u.last_login, u.created_at, u.updated_at
		 FROM users u
		 WHERE u.id NOT IN (
		   SELECT ur.user_id FROM user_roles ur
		   JOIN roles r ON r.id = ur.role_id
		   WHERE r.name IN ('admin', 'editor')
		 )
		 ORDER BY u.firstname, u.lastname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserWithRoles
	for rows.Next() {
		var u UserWithRoles
		if err := rows.Scan(&u.ID, &u.Firstname, &u.Lastname, &u.Company, &u.Email, &u.Password, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range users {
		roles, err := q.GetUserRoles(ctx, users[i].ID)
		if err != nil {
			return nil, err
		}
		users[i].Roles = roles
	}
	return users, nil
}

func (q *Queries) UpdateUser(ctx context.Context, id, firstname, lastname, company, email string) (User, error) {
	var u User
	err := q.Pool.QueryRow(ctx,
		`UPDATE users
		 SET firstname = $2, lastname = $3, company = $4, email = $5,
		     version = version + 1, updated_at = now()
		 WHERE id = $1
		 RETURNING id, firstname, lastname, company, email, password, last_login, created_at, updated_at`,
		id, firstname, lastname, company, email).
		Scan(&u.ID, &u.Firstname, &u.Lastname, &u.Company, &u.Email, &u.Password, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE users SET password = $2, updated_at = now() WHERE id = $1`, id, passwordHash)
	return err
}

func (q *Queries) GetUserVersion(ctx context.Context, userID string) (int, error) {
	var v int
	err := q.Pool.QueryRow(ctx, `SELECT version FROM users WHERE id = $1`, userID).Scan(&v)
	return v, err
}

func (q *Queries) SaveUserHistory(ctx context.Context, userID string, version int, firstname, lastname, company, email, roles, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO users_history (user_id, version, firstname, lastname, company, email, roles, changed_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID, version, firstname, lastname, company, email, roles, changedBy)
	return err
}

func (q *Queries) SetUserRoles(ctx context.Context, userID string, roleNames []string) error {
	_, err := q.Pool.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	for _, name := range roleNames {
		if err := q.AssignRole(ctx, userID, name); err != nil {
			return err
		}
	}
	return nil
}

func (q *Queries) GetRole(ctx context.Context, id string) (Role, error) {
	var r Role
	err := q.Pool.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1`, id).
		Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (q *Queries) CreateRole(ctx context.Context, name, description string) (Role, error) {
	var r Role
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO roles (name, description)
		 VALUES ($1, $2)
		 RETURNING id, name, description, created_at, updated_at`,
		name, description).
		Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (q *Queries) UpdateRole(ctx context.Context, id, name, description string) (Role, error) {
	var r Role
	err := q.Pool.QueryRow(ctx,
		`UPDATE roles
		 SET name = $2, description = $3, version = version + 1, updated_at = now()
		 WHERE id = $1
		 RETURNING id, name, description, created_at, updated_at`,
		id, name, description).
		Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (q *Queries) GetRoleVersion(ctx context.Context, roleID string) (int, error) {
	var v int
	err := q.Pool.QueryRow(ctx, `SELECT version FROM roles WHERE id = $1`, roleID).Scan(&v)
	return v, err
}

func (q *Queries) SaveRoleHistory(ctx context.Context, roleID string, version int, name, description, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO roles_history (role_id, version, name, description, changed_by)
		 VALUES ($1, $2, $3, $4, $5)`,
		roleID, version, name, description, changedBy)
	return err
}

func (q *Queries) ListAllRoles(ctx context.Context) ([]Role, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT id, name, description, created_at, updated_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

// --- Password reset token queries ---

func (q *Queries) CreatePasswordResetToken(ctx context.Context, userID, token string, expiresAt time.Time) (PasswordResetToken, error) {
	var t PasswordResetToken
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO password_reset_tokens (user_id, token, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token, expires_at, created_at`,
		userID, token, expiresAt).
		Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt)
	return t, err
}

func (q *Queries) GetPasswordResetToken(ctx context.Context, token string) (PasswordResetToken, error) {
	var t PasswordResetToken
	err := q.Pool.QueryRow(ctx,
		`SELECT id, user_id, token, expires_at, created_at
		 FROM password_reset_tokens WHERE token = $1 AND expires_at > now()`, token).
		Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt)
	return t, err
}

func (q *Queries) DeletePasswordResetTokensForUser(ctx context.Context, userID string) error {
	_, err := q.Pool.Exec(ctx,
		`DELETE FROM password_reset_tokens WHERE user_id = $1`, userID)
	return err
}

func (q *Queries) DeletePasswordResetToken(ctx context.Context, token string) error {
	_, err := q.Pool.Exec(ctx,
		`DELETE FROM password_reset_tokens WHERE token = $1`, token)
	return err
}

// --- Section Row queries ---

func (q *Queries) ListSectionRows(ctx context.Context) ([]SectionRow, error) {
	rows, err := q.Pool.Query(ctx,
		`SELECT id, title, description, sort_order, version FROM section_rows WHERE deleted = false ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sectionRows []SectionRow
	for rows.Next() {
		var r SectionRow
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.SortOrder, &r.Version); err != nil {
			return nil, err
		}
		sectionRows = append(sectionRows, r)
	}
	return sectionRows, rows.Err()
}

func (q *Queries) GetSectionRow(ctx context.Context, id string) (SectionRow, error) {
	var r SectionRow
	err := q.Pool.QueryRow(ctx,
		`SELECT id, title, description, sort_order, version FROM section_rows WHERE id = $1 AND deleted = false`, id).
		Scan(&r.ID, &r.Title, &r.Description, &r.SortOrder, &r.Version)
	return r, err
}

func (q *Queries) CreateSectionRow(ctx context.Context, title, description string, sortOrder int, changedBy string) (SectionRow, error) {
	var r SectionRow
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO section_rows (title, description, sort_order, changed_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, title, description, sort_order, version`,
		title, description, sortOrder, changedBy).
		Scan(&r.ID, &r.Title, &r.Description, &r.SortOrder, &r.Version)
	return r, err
}

func (q *Queries) UpdateSectionRow(ctx context.Context, id string, title, description, changedBy string) (SectionRow, error) {
	var r SectionRow
	err := q.Pool.QueryRow(ctx,
		`UPDATE section_rows
		 SET title = $2, description = $3, version = version + 1, updated_at = now(), changed_by = $4
		 WHERE id = $1
		 RETURNING id, title, description, sort_order, version`,
		id, title, description, changedBy).
		Scan(&r.ID, &r.Title, &r.Description, &r.SortOrder, &r.Version)
	return r, err
}

func (q *Queries) SoftDeleteSectionRow(ctx context.Context, id string, changedBy string) error {
	tx, err := q.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE sections SET row_id = NULL, version = version + 1, updated_at = now(), changed_by = $2
		 WHERE row_id = $1 AND deleted = false`, id, changedBy)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE section_rows SET deleted = true, version = version + 1, updated_at = now(), changed_by = $2
		 WHERE id = $1`, id, changedBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (q *Queries) SaveSectionRowHistory(ctx context.Context, r SectionRow, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`INSERT INTO section_rows_history (row_id, version, title, description, sort_order, changed_by)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		r.ID, r.Version, r.Title, r.Description, r.SortOrder, changedBy)
	return err
}

type ReorderItem struct {
	SectionID string
	SortOrder int
	RowID     *string
}

type ReorderRowItem struct {
	RowID     string
	SortOrder int
}

func (q *Queries) ReorderPages(ctx context.Context, sectionID string, items []PageOrderItem, changedBy string) error {
	tx, err := q.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, item := range items {
		// Top-level page: set parent_slug = NULL
		_, err := tx.Exec(ctx,
			`UPDATE pages SET sort_order = $1, parent_slug = NULL, version = version + 1, updated_at = now(), changed_by = $4
			 WHERE section_id = $2 AND slug = $3 AND deleted = false`,
			i, sectionID, item.Slug, changedBy)
		if err != nil {
			return err
		}
		// Children of this page
		for j, childSlug := range item.Children {
			_, err := tx.Exec(ctx,
				`UPDATE pages SET sort_order = $1, parent_slug = $4, version = version + 1, updated_at = now(), changed_by = $5
				 WHERE section_id = $2 AND slug = $3 AND deleted = false`,
				j, sectionID, childSlug, item.Slug, changedBy)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (q *Queries) PromoteChildren(ctx context.Context, sectionID, parentSlug, changedBy string) error {
	_, err := q.Pool.Exec(ctx,
		`UPDATE pages SET parent_slug = NULL, version = version + 1, updated_at = now(), changed_by = $3
		 WHERE section_id = $1 AND parent_slug = $2 AND deleted = false`,
		sectionID, parentSlug, changedBy)
	return err
}

func (q *Queries) ReorderSectionsAndRows(ctx context.Context, sections []ReorderItem, sectionRows []ReorderRowItem, changedBy string) error {
	tx, err := q.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, s := range sections {
		_, err := tx.Exec(ctx,
			`UPDATE sections SET sort_order = $2, row_id = $3, version = version + 1, updated_at = now(), changed_by = $4
			 WHERE id = $1`,
			s.SectionID, s.SortOrder, s.RowID, changedBy)
		if err != nil {
			return err
		}
	}

	for _, r := range sectionRows {
		_, err := tx.Exec(ctx,
			`UPDATE section_rows SET sort_order = $2, version = version + 1, updated_at = now(), changed_by = $3
			 WHERE id = $1`,
			r.RowID, r.SortOrder, changedBy)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
