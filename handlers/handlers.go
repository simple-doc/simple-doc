package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"

	"path/filepath"
	"regexp"
	"unicode"

	"docgen/internal/db"
	"docgen/internal/markdown"

	"golang.org/x/text/unicode/norm"
)

// TemplateSection mirrors the template data model from the old generator.
type TemplateSection struct {
	ID           string
	Name         string
	Title        string
	Description  string
	Icon         string
	BasePath     string
	RequiredRole string
	Disabled     bool
	RowID        *string
	IsEditor     bool
}

type TemplateRow struct {
	ID          string
	Title       string
	Description string
	Sections    []TemplateSection
}

type TemplatePage struct {
	Title      string
	Slug       string
	Content    template.HTML
	IsActive   bool
	Children   []TemplatePage
	IsChild    bool
	ParentSlug string
}

type SiteData struct {
	SiteTitle     string
	Badge         string
	ThemeCSS      template.HTML
	Pages         []TemplatePage
	Current       TemplatePage
	Section       TemplateSection
	HomePath      string
	UserFirstname string
	IsEditor      bool
	PreviewMode   bool
	PreviewRoles  string
}

type EditData struct {
	SiteTitle     string
	Badge         string
	ThemeCSS      template.HTML
	Pages         []TemplatePage
	Section       TemplateSection
	HomePath      string
	PageTitle     string
	ContentMD     string
	Slug          string
	Version       int
	Images        []db.ImageMeta
	UserFirstname string
	IsEditor      bool
	Error         string
}

type EditSectionData struct {
	SiteTitle     string
	ThemeCSS      template.HTML
	HomePath      string
	SectionID     string
	SectionName   string
	Title         string
	Description   string
	Icon          string
	Version       int
	UserFirstname string
	IsEditor      bool
	Roles         []db.Role
	RequiredRole  string
	Pages         []TemplatePage
}

type HomeData struct {
	SiteTitle         string
	ThemeCSS          template.HTML
	Sections          []TemplateSection
	Badge             string
	Heading           string
	Description       string
	Footer            string
	UserFirstname     string
	UserLastname      string
	IsEditor          bool
	IsAdmin           bool
	Roles             []db.Role
	Rows              []TemplateRow
	UngroupedSections []TemplateSection
	HasRows           bool
	RowIDParam        string
	PreviewMode       bool
	PreviewRoles      string
	ShowPreviewBtn    bool
	PreviewAllRoles   []db.Role
	PreviewUsers      []db.UserWithRoles
}

type RowFormData struct {
	SiteTitle     string
	ThemeCSS      template.HTML
	HomePath      string
	Title         string
	Description   string
	UserFirstname string
	IsEditor      bool
	RowID         string
	Version       int
	IsNew         bool
}

type EditHomeData struct {
	SiteTitle     string
	ThemeCSS      template.HTML
	HomePath      string
	Badge         string
	Heading       string
	Description   string
	Footer        string
	Theme         string
	AccentColor   string
	Version       int
	UserFirstname string
	IsEditor      bool
	HasFavicon    bool
}


var nonAlphanumDash = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// sanitizeFilename normalizes a filename for safe use in URLs and markdown.
// It strips diacritics, replaces spaces and special characters with hyphens,
// and lowercases the result.
func sanitizeFilename(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	// Normalize unicode and strip diacritical marks
	var clean []rune
	for _, r := range norm.NFD.String(base) {
		if !unicode.Is(unicode.Mn, r) {
			clean = append(clean, r)
		}
	}
	base = string(clean)

	// Replace non-alphanumeric characters with hyphens
	base = nonAlphanumDash.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	base = strings.ToLower(base)

	if base == "" {
		base = "image"
	}

	return base + strings.ToLower(ext)
}

// FormatBytes returns a human-readable byte size string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

type Handlers struct {
	DB             *db.Queries
	Tmpl           *template.Template
	TemplatesFS    fs.FS
	FuncMap        template.FuncMap
	DefaultFavicon []byte
	faviconV       atomic.Int64
}

// FaviconVersionFunc returns a template.FuncMap with a "faviconVersion"
// function that returns the current favicon cache-bust version.
func (h *Handlers) FaviconVersionFunc() template.FuncMap {
	return template.FuncMap{
		"faviconVersion": func() string {
			return strconv.FormatInt(h.faviconV.Load(), 10)
		},
	}
}

// InitFaviconVersion loads the current site_settings version into the
// in-memory counter used for favicon cache busting.
func (h *Handlers) InitFaviconVersion(ctx context.Context) {
	settings, _ := h.DB.GetSiteSettings(ctx)
	h.faviconV.Store(int64(settings.Version))
}

func (h *Handlers) bumpFaviconVersion() {
	h.faviconV.Add(1)
}

// tmpl returns the template set. When TemplatesFS is set (dev mode),
// it re-parses templates from disk on every call so edits appear without
// a server restart. Otherwise it returns the cached Tmpl.
func (h *Handlers) tmpl() *template.Template {
	if h.TemplatesFS != nil {
		t, err := template.New("").Funcs(h.FuncMap).ParseFS(h.TemplatesFS, "*.html")
		if err != nil {
			slog.Error("tmpl re-parse", "error", err)
			return h.Tmpl
		}
		return t
	}
	return h.Tmpl
}

type ErrorData struct {
	SiteTitle string
	ThemeCSS  template.HTML
	Code      int
	Title     string
	Message   string
}

func (h *Handlers) renderError(w http.ResponseWriter, r *http.Request, code int, title, message string) {
	siteTitle, _, themeCSS := h.siteSettings(r.Context())
	w.WriteHeader(code)
	data := ErrorData{
		SiteTitle: siteTitle,
		ThemeCSS:  themeCSS,
		Code:      code,
		Title:     title,
		Message:   message,
	}
	if err := h.tmpl().ExecuteTemplate(w, "error.html", data); err != nil {
		slog.Error("renderError template", "error", err)
		http.Error(w, title, code)
	}
}

func (h *Handlers) notFound(w http.ResponseWriter, r *http.Request) {
	h.renderError(w, r, http.StatusNotFound, "Page Not Found",
		"The page you're looking for doesn't exist or has been moved.")
}

func (h *Handlers) forbidden(w http.ResponseWriter, r *http.Request) {
	h.renderError(w, r, http.StatusForbidden, "Access Denied",
		"You don't have permission to access this page.")
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request) {
	h.renderError(w, r, http.StatusInternalServerError, "Something Went Wrong",
		"An unexpected error occurred. Please try again later.")
}

func (h *Handlers) siteSettings(ctx context.Context) (string, string, template.HTML) {
	settings, _ := h.DB.GetSiteSettings(ctx)
	return settings.SiteTitle, settings.Badge, ThemeCSS(settings.Theme, settings.AccentColor)
}

func userFirstname(ctx context.Context) string {
	if u := UserFromContext(ctx); u != nil {
		return u.Firstname
	}
	return ""
}

func userID(ctx context.Context) string {
	if u := UserFromContext(ctx); u != nil {
		return u.ID
	}
	return ""
}

// buildPageTree converts a flat list of pages into a tree with one level of nesting.
// Top-level pages (parent_slug == nil) are returned in order, with their children nested.
func buildPageTree(pages []db.Page, activeSlug string) []TemplatePage {
	// Separate top-level and children
	childrenMap := make(map[string][]db.Page) // parent_slug -> children
	var topLevel []db.Page
	for _, p := range pages {
		if p.ParentSlug != nil {
			childrenMap[*p.ParentSlug] = append(childrenMap[*p.ParentSlug], p)
		} else {
			topLevel = append(topLevel, p)
		}
	}

	var result []TemplatePage
	for _, p := range topLevel {
		tp := TemplatePage{
			Title:    p.Title,
			Slug:     p.Slug,
			IsActive: p.Slug == activeSlug,
		}
		if kids, ok := childrenMap[p.Slug]; ok {
			for _, c := range kids {
				tp.Children = append(tp.Children, TemplatePage{
					Title:      c.Title,
					Slug:       c.Slug,
					IsActive:   c.Slug == activeSlug,
					IsChild:    true,
					ParentSlug: p.Slug,
				})
			}
		}
		result = append(result, tp)
	}
	return result
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	sections, err := h.DB.ListSections(r.Context())
	if err != nil {
		h.serverError(w, r)
		slog.Error("Home", "error", err)
		return
	}

	settings, _ := h.DB.GetSiteSettings(r.Context())

	isEditor := h.isEditor(r.Context())

	var tplSections []TemplateSection
	for _, s := range sections {
		disabled := !h.canAccessSection(r.Context(), s.RequiredRole)
		tplSections = append(tplSections, TemplateSection{
			ID:           s.ID,
			Name:         s.Name,
			Title:        s.Title,
			Description:  s.Description,
			Icon:         s.Icon,
			BasePath:     "/" + s.Name + "/",
			RequiredRole: s.RequiredRole,
			Disabled:     disabled,
			RowID:        s.RowID,
			IsEditor:     isEditor,
		})
	}

	sectionRows, err := h.DB.ListSectionRows(r.Context())
	if err != nil {
		slog.Error("Home rows", "error", err)
	}

	hasRows := len(sectionRows) > 0

	var tplRows []TemplateRow
	var ungrouped []TemplateSection

	if hasRows {
		tplRows = make([]TemplateRow, len(sectionRows))
		rowIdx := make(map[string]int) // row ID -> index in tplRows
		for i, row := range sectionRows {
			tplRows[i] = TemplateRow{
				ID:          row.ID,
				Title:       row.Title,
				Description: row.Description,
			}
			rowIdx[row.ID] = i
		}
		for _, ts := range tplSections {
			if ts.RowID == nil {
				ungrouped = append(ungrouped, ts)
			} else if idx, ok := rowIdx[*ts.RowID]; ok {
				tplRows[idx].Sections = append(tplRows[idx].Sections, ts)
			} else {
				ungrouped = append(ungrouped, ts)
			}
		}
	}

	u := UserFromContext(r.Context())
	previewing := inPreviewMode(r.Context())
	var previewRolesStr string
	if previewing {
		roles := PreviewRolesFromContext(r.Context())
		previewRolesStr = strings.Join(roles, ", ")
		if previewRolesStr == "" {
			previewRolesStr = "(no custom roles)"
		}
	}

	// Determine if the real user is an editor (ignoring preview mode) for showing preview button
	realIsEditor := false
	if u != nil {
		adm, _ := h.DB.HasRole(r.Context(), u.ID, "admin")
		if adm {
			realIsEditor = true
		} else {
			ed, _ := h.DB.HasRole(r.Context(), u.ID, "editor")
			realIsEditor = ed
		}
	}

	data := HomeData{
		SiteTitle:         settings.SiteTitle,
		ThemeCSS:          ThemeCSS(settings.Theme, settings.AccentColor),
		Sections:          tplSections,
		Badge:             settings.Badge,
		Heading:           settings.Heading,
		Description:       settings.Description,
		Footer:            settings.Footer,
		UserFirstname:     u.Firstname,
		UserLastname:      u.Lastname,
		IsEditor:          isEditor,
		IsAdmin:           h.isAdmin(r.Context()),
		Rows:              tplRows,
		UngroupedSections: ungrouped,
		HasRows:           hasRows,
		PreviewMode:       previewing,
		PreviewRoles:      previewRolesStr,
	}

	// Populate modal data for preview button (only when real editor and not in preview)
	if realIsEditor && !previewing {
		data.ShowPreviewBtn = true
		if allRoles, err := h.DB.ListRoles(r.Context()); err == nil {
			data.PreviewAllRoles = allRoles
		}
		if users, err := h.DB.ListNonEditorUsers(r.Context()); err == nil {
			data.PreviewUsers = users
		}
	}

	if err := h.tmpl().ExecuteTemplate(w, "home.html", data); err != nil {
		slog.Error("Home template", "error", err)
	}
}

func (h *Handlers) Section(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	if !h.canAccessSection(r.Context(), section.RequiredRole) {
		h.forbidden(w, r)
		return
	}

	first, err := h.DB.GetFirstPage(r.Context(), section.ID)
	if err != nil {
		// Section exists but has no pages — show empty state
		title, badge, themeCSS := h.siteSettings(r.Context())
		previewing := inPreviewMode(r.Context())
		var previewRolesStr string
		if previewing {
			pr := PreviewRolesFromContext(r.Context())
			previewRolesStr = strings.Join(pr, ", ")
			if previewRolesStr == "" {
				previewRolesStr = "(no custom roles)"
			}
		}
		data := SiteData{
			SiteTitle: title,
			Badge:     badge,
			ThemeCSS:  themeCSS,
			Section: TemplateSection{
				ID:          section.ID,
				Name:        section.Name,
				Title:       section.Title,
				Description: section.Description,
				Icon:        section.Icon,
				BasePath:    "/" + section.Name + "/",
			},
			HomePath:      "/",
			UserFirstname: userFirstname(r.Context()),
			IsEditor:      h.isEditor(r.Context()),
			PreviewMode:   previewing,
			PreviewRoles:  previewRolesStr,
		}
		if err := h.tmpl().ExecuteTemplate(w, "empty-section.html", data); err != nil {
			slog.Error("Section empty template", "error", err)
		}
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", section.Name, first.Slug), http.StatusFound)
}

func (h *Handlers) Page(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")
	slug := r.PathValue("slug")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	if !h.canAccessSection(r.Context(), section.RequiredRole) {
		h.forbidden(w, r)
		return
	}

	page, err := h.DB.GetPage(r.Context(), section.ID, slug)
	if err != nil {
		h.notFound(w, r)
		return
	}

	allPages, err := h.DB.ListPagesBySection(r.Context(), section.ID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("Page", "error", err)
		return
	}

	htmlBytes, err := markdown.Render([]byte(page.ContentMD))
	if err != nil {
		h.serverError(w, r)
		slog.Error("Page render", "error", err)
		return
	}

	// Rewrite image paths from static/images/ to /images/
	htmlStr := strings.ReplaceAll(string(htmlBytes), "static/images/", "/images/")

	navPages := buildPageTree(allPages, slug)

	pageTitle, pageBadge, pageThemeCSS := h.siteSettings(r.Context())
	previewing := inPreviewMode(r.Context())
	var previewRolesStr string
	if previewing {
		pr := PreviewRolesFromContext(r.Context())
		previewRolesStr = strings.Join(pr, ", ")
		if previewRolesStr == "" {
			previewRolesStr = "(no custom roles)"
		}
	}
	data := SiteData{
		SiteTitle: pageTitle,
		Badge:     pageBadge,
		ThemeCSS:  pageThemeCSS,
		Pages:     navPages,
		Current: TemplatePage{
			Title:   page.Title,
			Slug:    page.Slug,
			Content: template.HTML(htmlStr),
		},
		Section: TemplateSection{
			ID:       section.ID,
			Name:     section.Name,
			Title:    section.Title,
			BasePath: "/" + section.Name + "/",
		},
		HomePath:      "/",
		UserFirstname: userFirstname(r.Context()),
		IsEditor:      h.isEditor(r.Context()),
		PreviewMode:   previewing,
		PreviewRoles:  previewRolesStr,
	}

	if err := h.tmpl().ExecuteTemplate(w, "page.html", data); err != nil {
		slog.Error("Page template", "error", err)
	}
}

func (h *Handlers) EditPage(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")
	slug := r.PathValue("slug")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	page, err := h.DB.GetPage(r.Context(), section.ID, slug)
	if err != nil {
		h.notFound(w, r)
		return
	}

	allPages, err := h.DB.ListPagesBySection(r.Context(), section.ID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("EditPage", "error", err)
		return
	}

	navPages := buildPageTree(allPages, slug)

	imageMetas, err := h.DB.ListImageMetasBySection(r.Context(), section.ID)
	if err != nil {
		slog.Error("EditPage images", "error", err)
	}

	editTitle, editBadge, editThemeCSS := h.siteSettings(r.Context())
	data := EditData{
		SiteTitle: editTitle,
		Badge:     editBadge,
		ThemeCSS:  editThemeCSS,
		Pages:     navPages,
		Section: TemplateSection{
			ID:       section.ID,
			Name:     section.Name,
			Title:    section.Title,
			BasePath: "/" + section.Name + "/",
		},
		HomePath:      "/",
		PageTitle:     page.Title,
		ContentMD:     page.ContentMD,
		Slug:          page.Slug,
		Version:       page.Version,
		Images:        imageMetas,
		UserFirstname: userFirstname(r.Context()),
		Error:         r.URL.Query().Get("error"),
	}

	if err := h.tmpl().ExecuteTemplate(w, "edit.html", data); err != nil {
		slog.Error("EditPage template", "error", err)
	}
}

func (h *Handlers) SavePage(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")
	slug := r.PathValue("slug")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	contentMD := r.FormValue("content_md")

	if title == "" || contentMD == "" {
		http.Error(w, "title and content are required", http.StatusBadRequest)
		return
	}

	changedBy := userID(r.Context())
	updated, err := h.DB.UpdatePage(r.Context(), section.ID, slug, title, contentMD, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("SavePage", "error", err)
		return
	}

	if err := h.DB.SavePageHistory(r.Context(), updated, changedBy); err != nil {
		slog.Error("SavePage history", "error", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", section.Name, slug), http.StatusSeeOther)
}

func (h *Handlers) PreviewPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	contentMD := r.FormValue("content_md")

	htmlBytes, err := markdown.Render([]byte(contentMD))
	if err != nil {
		h.serverError(w, r)
		slog.Error("PreviewPage", "error", err)
		return
	}

	htmlStr := strings.ReplaceAll(string(htmlBytes), "static/images/", "/images/")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlStr))
}

func (h *Handlers) Image(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")

	img, err := h.DB.GetImage(r.Context(), filename)
	if err != nil {
		h.notFound(w, r)
		return
	}

	hash := sha256.Sum256(img.Data)
	etag := fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:16]))
	w.Header().Set("Content-Type", img.ContentType)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("ETag", etag)

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Write(img.Data)
}

func (h *Handlers) UploadImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.serverError(w, r)
		slog.Error("UploadImage read", "error", err)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	filename := sanitizeFilename(header.Filename)
	sectionID := r.FormValue("section_id")

	changedBy := userID(r.Context())

	// Upsert: update if filename already exists, otherwise create
	var img db.Image
	_, err = h.DB.GetImage(r.Context(), filename)
	if err == nil {
		img, err = h.DB.UpdateImage(r.Context(), filename, contentType, data, changedBy)
	} else {
		img, err = h.DB.CreateImage(r.Context(), filename, contentType, data, sectionID, changedBy)
	}
	if err != nil {
		h.serverError(w, r)
		slog.Error("UploadImage", "error", err)
		return
	}

	if err := h.DB.SaveImageHistory(r.Context(), img, changedBy); err != nil {
		slog.Error("UploadImage history", "error", err)
	}

	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect+"#images", http.StatusSeeOther)
}

func (h *Handlers) UpdateImageHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.serverError(w, r)
		slog.Error("UpdateImage read", "error", err)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	changedBy := userID(r.Context())
	img, err := h.DB.UpdateImage(r.Context(), filename, contentType, data, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("UpdateImage update", "error", err)
		return
	}

	if err := h.DB.SaveImageHistory(r.Context(), img, changedBy); err != nil {
		slog.Error("UpdateImage history", "error", err)
	}

	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect+"#images", http.StatusSeeOther)
}

func (h *Handlers) NewPageForm(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	allPages, err := h.DB.ListPagesBySection(r.Context(), section.ID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("NewPageForm", "error", err)
		return
	}

	navPages := buildPageTree(allPages, "")

	npTitle, npBadge, npThemeCSS := h.siteSettings(r.Context())
	data := EditData{
		SiteTitle: npTitle,
		Badge:     npBadge,
		ThemeCSS:  npThemeCSS,
		Pages:     navPages,
		Section: TemplateSection{
			ID:       section.ID,
			Name:     section.Name,
			Title:    section.Title,
			BasePath: "/" + section.Name + "/",
		},
		HomePath:      "/",
		UserFirstname: userFirstname(r.Context()),
	}

	if err := h.tmpl().ExecuteTemplate(w, "new-page.html", data); err != nil {
		slog.Error("NewPageForm template", "error", err)
	}
}

func (h *Handlers) CreatePage(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	slug := r.FormValue("slug")
	title := r.FormValue("title")
	contentMD := r.FormValue("content_md")

	if slug == "" || title == "" {
		http.Error(w, "slug and title are required", http.StatusBadRequest)
		return
	}

	// Auto-calculate sort_order
	pages, err := h.DB.ListPagesBySection(r.Context(), section.ID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("CreatePage list", "error", err)
		return
	}
	sortOrder := len(pages)

	changedBy := userID(r.Context())
	page, err := h.DB.CreatePage(r.Context(), section.ID, slug, title, contentMD, sortOrder, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("CreatePage", "error", err)
		return
	}

	if err := h.DB.SavePageHistory(r.Context(), page, changedBy); err != nil {
		slog.Error("CreatePage history", "error", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", section.Name, slug), http.StatusSeeOther)
}

func (h *Handlers) NewSectionForm(w http.ResponseWriter, r *http.Request) {
	nsTitle, _, nsThemeCSS := h.siteSettings(r.Context())
	roles, _ := h.DB.ListRoles(r.Context())
	data := HomeData{
		SiteTitle:     nsTitle,
		ThemeCSS:      nsThemeCSS,
		UserFirstname: userFirstname(r.Context()),
		Roles:         roles,
		RowIDParam:    r.URL.Query().Get("row_id"),
	}

	if err := h.tmpl().ExecuteTemplate(w, "new-section.html", data); err != nil {
		slog.Error("NewSectionForm template", "error", err)
	}
}

func (h *Handlers) CreateSection(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	title := r.FormValue("title")
	description := r.FormValue("description")
	icon := r.FormValue("icon")
	requiredRole := r.FormValue("required_role")
	rowIDStr := r.FormValue("row_id")

	if name == "" || title == "" {
		http.Error(w, "name and title are required", http.StatusBadRequest)
		return
	}

	if icon == "" {
		icon = "document"
	}

	var rowID *string
	if rowIDStr != "" {
		rowID = &rowIDStr
	}

	// Auto-calculate sort_order from count of existing sections
	sections, err := h.DB.ListSections(r.Context())
	if err != nil {
		h.serverError(w, r)
		slog.Error("CreateSection list", "error", err)
		return
	}
	sortOrder := len(sections)

	changedBy := userID(r.Context())
	section, err := h.DB.CreateSection(r.Context(), name, title, description, icon, sortOrder, requiredRole, changedBy, rowID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("CreateSection", "error", err)
		return
	}

	if err := h.DB.SaveSectionHistory(r.Context(), section, changedBy); err != nil {
		slog.Error("CreateSection history", "error", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) EditSectionForm(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	roles, _ := h.DB.ListRoles(r.Context())

	allPages, _ := h.DB.ListPagesBySection(r.Context(), section.ID)
	tplPages := buildPageTree(allPages, "")

	esTitle, _, esThemeCSS := h.siteSettings(r.Context())
	data := EditSectionData{
		SiteTitle:     esTitle,
		ThemeCSS:      esThemeCSS,
		HomePath:      "/",
		SectionID:     section.ID,
		SectionName:   section.Name,
		Title:         section.Title,
		Description:   section.Description,
		Icon:          section.Icon,
		Version:       section.Version,
		UserFirstname: userFirstname(r.Context()),
		Roles:         roles,
		RequiredRole:  section.RequiredRole,
		Pages:         tplPages,
	}

	if err := h.tmpl().ExecuteTemplate(w, "edit-section.html", data); err != nil {
		slog.Error("EditSectionForm template", "error", err)
	}
}

func (h *Handlers) UpdateSection(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	icon := r.FormValue("icon")
	requiredRole := r.FormValue("required_role")

	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	if icon == "" {
		icon = "document"
	}

	changedBy := userID(r.Context())
	updated, err := h.DB.UpdateSection(r.Context(), section.ID, title, description, icon, requiredRole, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("UpdateSection", "error", err)
		return
	}

	if err := h.DB.SaveSectionHistory(r.Context(), updated, changedBy); err != nil {
		slog.Error("UpdateSection history", "error", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) DeleteSection(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	changedBy := userID(r.Context())
	if err := h.DB.SoftDeleteSection(r.Context(), section.ID, changedBy); err != nil {
		h.serverError(w, r)
		slog.Error("DeleteSection", "error", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) DeletePage(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")
	slug := r.PathValue("slug")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		h.notFound(w, r)
		return
	}

	_, err = h.DB.GetPage(r.Context(), section.ID, slug)
	if err != nil {
		h.notFound(w, r)
		return
	}

	changedBy := userID(r.Context())

	// Promote any children to top-level before deleting the parent
	if err := h.DB.PromoteChildren(r.Context(), section.ID, slug, changedBy); err != nil {
		slog.Error("DeletePage promote children", "error", err)
	}

	if err := h.DB.SoftDeletePage(r.Context(), section.ID, slug, changedBy); err != nil {
		h.serverError(w, r)
		slog.Error("DeletePage", "error", err)
		return
	}

	http.Redirect(w, r, "/"+section.Name+"/", http.StatusSeeOther)
}

func (h *Handlers) DeleteImage(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")

	if err := h.DB.DeleteImage(r.Context(), filename); err != nil {
		h.serverError(w, r)
		slog.Error("DeleteImage", "error", err)
		return
	}

	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect+"#images", http.StatusSeeOther)
}

func (h *Handlers) RenameImage(w http.ResponseWriter, r *http.Request) {
	oldFilename := r.PathValue("filename")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	newFilename := sanitizeFilename(r.FormValue("new_filename"))
	if newFilename == "" || newFilename == oldFilename {
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			redirect = "/"
		}
		http.Redirect(w, r, redirect+"#images", http.StatusSeeOther)
		return
	}

	// Check if the new filename is already taken
	if _, err := h.DB.GetImage(r.Context(), newFilename); err == nil {
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			redirect = "/"
		}
		sep := "?"
		if strings.Contains(redirect, "?") {
			sep = "&"
		}
		http.Redirect(w, r, redirect+sep+"error="+url.QueryEscape(fmt.Sprintf("An image with the filename %q already exists.", newFilename))+"#images", http.StatusSeeOther)
		return
	}

	changedBy := userID(r.Context())

	// Save history before rename (with old filename)
	oldImg, err := h.DB.GetImage(r.Context(), oldFilename)
	if err != nil {
		h.notFound(w, r)
		return
	}
	if err := h.DB.SaveImageHistory(r.Context(), oldImg, changedBy); err != nil {
		slog.Error("RenameImage history", "error", err)
	}

	if _, err := h.DB.RenameImage(r.Context(), oldFilename, newFilename, changedBy); err != nil {
		h.serverError(w, r)
		slog.Error("RenameImage", "error", err)
		return
	}

	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect+"#images", http.StatusSeeOther)
}

func (h *Handlers) EditHomeForm(w http.ResponseWriter, r *http.Request) {
	settings, _ := h.DB.GetSiteSettings(r.Context())

	data := EditHomeData{
		SiteTitle:     settings.SiteTitle,
		ThemeCSS:      ThemeCSS(settings.Theme, settings.AccentColor),
		HomePath:      "/",
		Badge:         settings.Badge,
		Heading:       settings.Heading,
		Description:   settings.Description,
		Footer:        settings.Footer,
		Theme:         settings.Theme,
		AccentColor:   settings.AccentColor,
		Version:       settings.Version,
		UserFirstname: userFirstname(r.Context()),
		HasFavicon:    settings.HasFavicon,
	}

	if err := h.tmpl().ExecuteTemplate(w, "edit-home.html", data); err != nil {
		slog.Error("EditHomeForm template", "error", err)
	}
}

func (h *Handlers) UpdateHome(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	siteTitle := r.FormValue("site_title")
	badge := r.FormValue("badge")
	heading := r.FormValue("heading")
	description := r.FormValue("description")
	footer := r.FormValue("footer")
	theme := r.FormValue("theme")
	accentColor := r.FormValue("accent_color")

	if siteTitle == "" || heading == "" {
		http.Error(w, "site title and heading are required", http.StatusBadRequest)
		return
	}

	if !ValidTheme(theme) {
		theme = "midnight"
	}
	if !ValidAccent(accentColor) {
		accentColor = "blue"
	}

	changedBy := userID(r.Context())
	settings, err := h.DB.UpdateSiteSettings(r.Context(), siteTitle, badge, heading, description, footer, theme, accentColor, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("UpdateHome", "error", err)
		return
	}

	if err := h.DB.SaveSiteSettingsHistory(r.Context(), settings, changedBy); err != nil {
		slog.Error("UpdateHome history", "error", err)
	}

	// Handle favicon: reset takes priority over upload
	if r.FormValue("reset_favicon") == "1" {
		if err := h.DB.DeleteFavicon(r.Context(), changedBy); err != nil {
			slog.Error("UpdateHome reset favicon", "error", err)
		} else {
			h.bumpFaviconVersion()
		}
	} else if file, header, err := r.FormFile("favicon"); err == nil {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			slog.Error("UpdateHome read favicon", "error", err)
		} else {
			contentType := header.Header.Get("Content-Type")
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			if err := h.DB.UpdateFavicon(r.Context(), data, contentType, changedBy); err != nil {
				slog.Error("UpdateHome save favicon", "error", err)
			} else {
				h.bumpFaviconVersion()
			}
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- Section Row handlers ---

func (h *Handlers) NewRowForm(w http.ResponseWriter, r *http.Request) {
	siteTitle, _, themeCSS := h.siteSettings(r.Context())
	data := RowFormData{
		SiteTitle:     siteTitle,
		ThemeCSS:      themeCSS,
		HomePath:      "/",
		UserFirstname: userFirstname(r.Context()),
		IsNew:         true,
	}
	if err := h.tmpl().ExecuteTemplate(w, "row-form.html", data); err != nil {
		slog.Error("NewRowForm template", "error", err)
	}
}

func (h *Handlers) CreateRow(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	existingRows, err := h.DB.ListSectionRows(r.Context())
	if err != nil {
		h.serverError(w, r)
		slog.Error("CreateRow list", "error", err)
		return
	}
	sortOrder := len(existingRows)

	changedBy := userID(r.Context())
	row, err := h.DB.CreateSectionRow(r.Context(), title, description, sortOrder, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("CreateRow", "error", err)
		return
	}

	if err := h.DB.SaveSectionRowHistory(r.Context(), row, changedBy); err != nil {
		slog.Error("CreateRow history", "error", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) EditRowForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	row, err := h.DB.GetSectionRow(r.Context(), id)
	if err != nil {
		h.notFound(w, r)
		return
	}

	siteTitle, _, themeCSS := h.siteSettings(r.Context())
	data := RowFormData{
		SiteTitle:     siteTitle,
		ThemeCSS:      themeCSS,
		HomePath:      "/",
		Title:         row.Title,
		Description:   row.Description,
		RowID:         row.ID,
		Version:       row.Version,
		UserFirstname: userFirstname(r.Context()),
		IsNew:         false,
	}
	if err := h.tmpl().ExecuteTemplate(w, "row-form.html", data); err != nil {
		slog.Error("EditRowForm template", "error", err)
	}
}

func (h *Handlers) UpdateRow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	changedBy := userID(r.Context())
	row, err := h.DB.UpdateSectionRow(r.Context(), id, title, description, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("UpdateRow", "error", err)
		return
	}

	if err := h.DB.SaveSectionRowHistory(r.Context(), row, changedBy); err != nil {
		slog.Error("UpdateRow history", "error", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) DeleteRow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	changedBy := userID(r.Context())
	if err := h.DB.SoftDeleteSectionRow(r.Context(), id, changedBy); err != nil {
		h.serverError(w, r)
		slog.Error("DeleteRow", "error", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type ReorderRequest struct {
	Rows []struct {
		ID        string   `json:"id"`
		SortOrder int      `json:"sort_order"`
		Sections  []string `json:"sections"`
	} `json:"rows"`
}

func (h *Handlers) Reorder(w http.ResponseWriter, r *http.Request) {
	var req ReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	var sectionItems []db.ReorderItem
	var rowItems []db.ReorderRowItem

	for _, row := range req.Rows {
		if row.ID != "" && row.ID != "0" {
			rowItems = append(rowItems, db.ReorderRowItem{
				RowID:     row.ID,
				SortOrder: row.SortOrder,
			})
		}
		for i, sectionID := range row.Sections {
			item := db.ReorderItem{
				SectionID: sectionID,
				SortOrder: i,
			}
			if row.ID != "" && row.ID != "0" {
				rid := row.ID
				item.RowID = &rid
			}
			sectionItems = append(sectionItems, item)
		}
	}

	changedBy := userID(r.Context())
	if err := h.DB.ReorderSectionsAndRows(r.Context(), sectionItems, rowItems, changedBy); err != nil {
		slog.Error("Reorder", "error", err)
		http.Error(w, "reorder failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *Handlers) ReorderPages(w http.ResponseWriter, r *http.Request) {
	sectionName := r.PathValue("section")

	section, err := h.DB.GetSectionByName(r.Context(), sectionName)
	if err != nil {
		http.Error(w, "section not found", http.StatusNotFound)
		return
	}

	var req struct {
		Pages []db.PageOrderItem `json:"pages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	changedBy := userID(r.Context())
	if err := h.DB.ReorderPages(r.Context(), section.ID, req.Pages, changedBy); err != nil {
		slog.Error("ReorderPages", "error", err)
		http.Error(w, "reorder failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h *Handlers) StartPreview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	token := sessionTokenFromContext(r.Context())
	if token == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var roles string
	if userIDVal := r.FormValue("user_id"); userIDVal != "" {
		userRoles, err := h.DB.GetUserRoles(r.Context(), userIDVal)
		if err != nil {
			slog.Error("StartPreview GetUserRoles", "error", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		roles = strings.Join(userRoles, ",")
	} else {
		roles = strings.Join(r.Form["roles"], ",")
	}

	if err := h.DB.SetSessionPreviewRoles(r.Context(), token, roles); err != nil {
		slog.Error("StartPreview SetSessionPreviewRoles", "error", err)
		h.serverError(w, r)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) StopPreview(w http.ResponseWriter, r *http.Request) {
	token := sessionTokenFromContext(r.Context())
	if token == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := h.DB.ClearSessionPreviewRoles(r.Context(), token); err != nil {
		slog.Error("StopPreview ClearSessionPreviewRoles", "error", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Favicon(w http.ResponseWriter, r *http.Request) {
	data, contentType, err := h.DB.GetFavicon(r.Context())
	if err == nil && len(data) > 0 {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-cache")
		hash := sha256.Sum256(data)
		etag := fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:16]))
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Write(data)
		return
	}

	if h.DefaultFavicon == nil {
		http.NotFound(w, r)
		return
	}
	hash := sha256.Sum256(h.DefaultFavicon)
	etag := fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:16]))
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Write(h.DefaultFavicon)
}

