package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"docgen/internal/db"
	"docgen/internal/markdown"
)

// TemplateSection mirrors the template data model from the old generator.
type TemplateSection struct {
	ID           string
	Title        string
	Description  string
	Icon         string
	BasePath     string
	RequiredRole string
	Disabled     bool
	RowID        *int
	IsEditor     bool
}

type TemplateRow struct {
	ID          int
	Title       string
	Description string
	Sections    []TemplateSection
}

type TemplatePage struct {
	Title    string
	Slug     string
	Content  template.HTML
	IsActive bool
}

type SiteData struct {
	SiteTitle     string
	ThemeCSS      template.HTML
	Pages         []TemplatePage
	Current       TemplatePage
	Section       TemplateSection
	HomePath      string
	UserFirstname string
	IsEditor      bool
}

type EditData struct {
	SiteTitle     string
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
}

type EditSectionData struct {
	SiteTitle     string
	ThemeCSS      template.HTML
	HomePath      string
	SectionID     string
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
}

type RowFormData struct {
	SiteTitle     string
	ThemeCSS      template.HTML
	HomePath      string
	Title         string
	Description   string
	UserFirstname string
	IsEditor      bool
	RowID         int
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
	DB   *db.Queries
	Tmpl *template.Template
}

type ErrorData struct {
	SiteTitle string
	ThemeCSS  template.HTML
	Code      int
	Title     string
	Message   string
}

func (h *Handlers) renderError(w http.ResponseWriter, r *http.Request, code int, title, message string) {
	siteTitle, themeCSS := h.siteSettings(r.Context())
	w.WriteHeader(code)
	data := ErrorData{
		SiteTitle: siteTitle,
		ThemeCSS:  themeCSS,
		Code:      code,
		Title:     title,
		Message:   message,
	}
	if err := h.Tmpl.ExecuteTemplate(w, "error.html", data); err != nil {
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

func (h *Handlers) siteSettings(ctx context.Context) (string, template.HTML) {
	settings, _ := h.DB.GetSiteSettings(ctx)
	return settings.SiteTitle, ThemeCSS(settings.Theme, settings.AccentColor)
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
			Title:        s.Title,
			Description:  s.Description,
			Icon:         s.Icon,
			BasePath:     "/" + s.ID + "/",
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
		rowIdx := make(map[int]int) // row ID -> index in tplRows
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
	}

	if err := h.Tmpl.ExecuteTemplate(w, "home.html", data); err != nil {
		slog.Error("Home template", "error", err)
	}
}

func (h *Handlers) Section(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("section")

	section, err := h.DB.GetSection(r.Context(), sectionID)
	if err != nil {
		h.notFound(w, r)
		return
	}

	if !h.canAccessSection(r.Context(), section.RequiredRole) {
		h.forbidden(w, r)
		return
	}

	first, err := h.DB.GetFirstPage(r.Context(), sectionID)
	if err != nil {
		// Section exists but has no pages — show empty state
		title, themeCSS := h.siteSettings(r.Context())
		data := SiteData{
			SiteTitle: title,
			ThemeCSS:  themeCSS,
			Section: TemplateSection{
				ID:          section.ID,
				Title:       section.Title,
				Description: section.Description,
				Icon:        section.Icon,
				BasePath:    "/" + section.ID + "/",
			},
			HomePath:      "/",
			UserFirstname: userFirstname(r.Context()),
			IsEditor:      h.isEditor(r.Context()),
		}
		if err := h.Tmpl.ExecuteTemplate(w, "empty-section.html", data); err != nil {
			slog.Error("Section empty template", "error", err)
		}
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", sectionID, first.Slug), http.StatusFound)
}

func (h *Handlers) Page(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("section")
	slug := r.PathValue("slug")

	page, err := h.DB.GetPage(r.Context(), sectionID, slug)
	if err != nil {
		h.notFound(w, r)
		return
	}

	section, err := h.DB.GetSection(r.Context(), sectionID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("Page", "error", err)
		return
	}

	if !h.canAccessSection(r.Context(), section.RequiredRole) {
		h.forbidden(w, r)
		return
	}

	allPages, err := h.DB.ListPagesBySection(r.Context(), sectionID)
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

	var navPages []TemplatePage
	for _, p := range allPages {
		navPages = append(navPages, TemplatePage{
			Title:    p.Title,
			Slug:     p.Slug,
			IsActive: p.Slug == slug,
		})
	}

	pageTitle, pageThemeCSS := h.siteSettings(r.Context())
	data := SiteData{
		SiteTitle: pageTitle,
		ThemeCSS:  pageThemeCSS,
		Pages:     navPages,
		Current: TemplatePage{
			Title:   page.Title,
			Slug:    page.Slug,
			Content: template.HTML(htmlStr),
		},
		Section: TemplateSection{
			ID:       section.ID,
			Title:    section.Title,
			BasePath: "/" + section.ID + "/",
		},
		HomePath:      "/",
		UserFirstname: userFirstname(r.Context()),
		IsEditor:      h.isEditor(r.Context()),
	}

	if err := h.Tmpl.ExecuteTemplate(w, "page.html", data); err != nil {
		slog.Error("Page template", "error", err)
	}
}

func (h *Handlers) EditPage(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("section")
	slug := r.PathValue("slug")

	page, err := h.DB.GetPage(r.Context(), sectionID, slug)
	if err != nil {
		h.notFound(w, r)
		return
	}

	section, err := h.DB.GetSection(r.Context(), sectionID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("EditPage", "error", err)
		return
	}

	allPages, err := h.DB.ListPagesBySection(r.Context(), sectionID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("EditPage", "error", err)
		return
	}

	var navPages []TemplatePage
	for _, p := range allPages {
		navPages = append(navPages, TemplatePage{
			Title:    p.Title,
			Slug:     p.Slug,
			IsActive: p.Slug == slug,
		})
	}

	imageMetas, err := h.DB.ListImageMetasBySection(r.Context(), sectionID)
	if err != nil {
		slog.Error("EditPage images", "error", err)
	}

	editTitle, editThemeCSS := h.siteSettings(r.Context())
	data := EditData{
		SiteTitle: editTitle,
		ThemeCSS:  editThemeCSS,
		Pages:     navPages,
		Section: TemplateSection{
			ID:       section.ID,
			Title:    section.Title,
			BasePath: "/" + section.ID + "/",
		},
		HomePath:      "/",
		PageTitle:     page.Title,
		ContentMD:     page.ContentMD,
		Slug:          page.Slug,
		Version:       page.Version,
		Images:        imageMetas,
		UserFirstname: userFirstname(r.Context()),
	}

	if err := h.Tmpl.ExecuteTemplate(w, "edit.html", data); err != nil {
		slog.Error("EditPage template", "error", err)
	}
}

func (h *Handlers) SavePage(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("section")
	slug := r.PathValue("slug")

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
	updated, err := h.DB.UpdatePage(r.Context(), sectionID, slug, title, contentMD, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("SavePage", "error", err)
		return
	}

	if err := h.DB.SavePageHistory(r.Context(), updated, changedBy); err != nil {
		slog.Error("SavePage history", "error", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", sectionID, slug), http.StatusSeeOther)
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

	etag := fmt.Sprintf(`"%s-%d"`, img.Filename, img.Version)
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

	sectionID := r.FormValue("section_id")

	changedBy := userID(r.Context())

	// Upsert: update if filename already exists, otherwise create
	var img db.Image
	_, err = h.DB.GetImage(r.Context(), header.Filename)
	if err == nil {
		img, err = h.DB.UpdateImage(r.Context(), header.Filename, contentType, data, changedBy)
	} else {
		img, err = h.DB.CreateImage(r.Context(), header.Filename, contentType, data, sectionID, changedBy)
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
	sectionID := r.PathValue("section")

	section, err := h.DB.GetSection(r.Context(), sectionID)
	if err != nil {
		h.notFound(w, r)
		return
	}

	allPages, err := h.DB.ListPagesBySection(r.Context(), sectionID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("NewPageForm", "error", err)
		return
	}

	var navPages []TemplatePage
	for _, p := range allPages {
		navPages = append(navPages, TemplatePage{
			Title: p.Title,
			Slug:  p.Slug,
		})
	}

	npTitle, npThemeCSS := h.siteSettings(r.Context())
	data := EditData{
		SiteTitle: npTitle,
		ThemeCSS:  npThemeCSS,
		Pages:     navPages,
		Section: TemplateSection{
			ID:       section.ID,
			Title:    section.Title,
			BasePath: "/" + section.ID + "/",
		},
		HomePath:      "/",
		UserFirstname: userFirstname(r.Context()),
	}

	if err := h.Tmpl.ExecuteTemplate(w, "new-page.html", data); err != nil {
		slog.Error("NewPageForm template", "error", err)
	}
}

func (h *Handlers) CreatePage(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("section")

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
	pages, err := h.DB.ListPagesBySection(r.Context(), sectionID)
	if err != nil {
		h.serverError(w, r)
		slog.Error("CreatePage list", "error", err)
		return
	}
	sortOrder := len(pages)

	changedBy := userID(r.Context())
	page, err := h.DB.CreatePage(r.Context(), sectionID, slug, title, contentMD, sortOrder, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("CreatePage", "error", err)
		return
	}

	if err := h.DB.SavePageHistory(r.Context(), page, changedBy); err != nil {
		slog.Error("CreatePage history", "error", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", sectionID, slug), http.StatusSeeOther)
}

func (h *Handlers) NewSectionForm(w http.ResponseWriter, r *http.Request) {
	nsTitle, nsThemeCSS := h.siteSettings(r.Context())
	roles, _ := h.DB.ListRoles(r.Context())
	data := HomeData{
		SiteTitle:     nsTitle,
		ThemeCSS:      nsThemeCSS,
		UserFirstname: userFirstname(r.Context()),
		Roles:         roles,
		RowIDParam:    r.URL.Query().Get("row_id"),
	}

	if err := h.Tmpl.ExecuteTemplate(w, "new-section.html", data); err != nil {
		slog.Error("NewSectionForm template", "error", err)
	}
}

func (h *Handlers) CreateSection(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	title := r.FormValue("title")
	description := r.FormValue("description")
	icon := r.FormValue("icon")
	requiredRole := r.FormValue("required_role")
	rowIDStr := r.FormValue("row_id")

	if id == "" || title == "" {
		http.Error(w, "id and title are required", http.StatusBadRequest)
		return
	}

	if icon == "" {
		icon = "document"
	}

	var rowID *int
	if rowIDStr != "" {
		if v, err := strconv.Atoi(rowIDStr); err == nil {
			rowID = &v
		}
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
	section, err := h.DB.CreateSection(r.Context(), id, title, description, icon, sortOrder, requiredRole, changedBy, rowID)
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
	sectionID := r.PathValue("section")

	section, err := h.DB.GetSection(r.Context(), sectionID)
	if err != nil {
		h.notFound(w, r)
		return
	}

	roles, _ := h.DB.ListRoles(r.Context())

	allPages, _ := h.DB.ListPagesBySection(r.Context(), sectionID)
	var tplPages []TemplatePage
	for _, p := range allPages {
		tplPages = append(tplPages, TemplatePage{Title: p.Title, Slug: p.Slug})
	}

	esTitle, esThemeCSS := h.siteSettings(r.Context())
	data := EditSectionData{
		SiteTitle:     esTitle,
		ThemeCSS:      esThemeCSS,
		HomePath:      "/",
		SectionID:     section.ID,
		Title:         section.Title,
		Description:   section.Description,
		Icon:          section.Icon,
		Version:       section.Version,
		UserFirstname: userFirstname(r.Context()),
		Roles:         roles,
		RequiredRole:  section.RequiredRole,
		Pages:         tplPages,
	}

	if err := h.Tmpl.ExecuteTemplate(w, "edit-section.html", data); err != nil {
		slog.Error("EditSectionForm template", "error", err)
	}
}

func (h *Handlers) UpdateSection(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("section")

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
	section, err := h.DB.UpdateSection(r.Context(), sectionID, title, description, icon, requiredRole, changedBy)
	if err != nil {
		h.serverError(w, r)
		slog.Error("UpdateSection", "error", err)
		return
	}

	if err := h.DB.SaveSectionHistory(r.Context(), section, changedBy); err != nil {
		slog.Error("UpdateSection history", "error", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) DeleteSection(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("section")

	_, err := h.DB.GetSection(r.Context(), sectionID)
	if err != nil {
		h.notFound(w, r)
		return
	}

	changedBy := userID(r.Context())
	if err := h.DB.SoftDeleteSection(r.Context(), sectionID, changedBy); err != nil {
		h.serverError(w, r)
		slog.Error("DeleteSection", "error", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) DeletePage(w http.ResponseWriter, r *http.Request) {
	sectionID := r.PathValue("section")
	slug := r.PathValue("slug")

	_, err := h.DB.GetPage(r.Context(), sectionID, slug)
	if err != nil {
		h.notFound(w, r)
		return
	}

	changedBy := userID(r.Context())
	if err := h.DB.SoftDeletePage(r.Context(), sectionID, slug, changedBy); err != nil {
		h.serverError(w, r)
		slog.Error("DeletePage", "error", err)
		return
	}

	http.Redirect(w, r, "/"+sectionID+"/", http.StatusSeeOther)
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
	}

	if err := h.Tmpl.ExecuteTemplate(w, "edit-home.html", data); err != nil {
		slog.Error("EditHomeForm template", "error", err)
	}
}

func (h *Handlers) UpdateHome(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
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

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- Section Row handlers ---

func (h *Handlers) NewRowForm(w http.ResponseWriter, r *http.Request) {
	siteTitle, themeCSS := h.siteSettings(r.Context())
	data := RowFormData{
		SiteTitle:     siteTitle,
		ThemeCSS:      themeCSS,
		HomePath:      "/",
		UserFirstname: userFirstname(r.Context()),
		IsNew:         true,
	}
	if err := h.Tmpl.ExecuteTemplate(w, "row-form.html", data); err != nil {
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
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.notFound(w, r)
		return
	}

	row, err := h.DB.GetSectionRow(r.Context(), id)
	if err != nil {
		h.notFound(w, r)
		return
	}

	siteTitle, themeCSS := h.siteSettings(r.Context())
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
	if err := h.Tmpl.ExecuteTemplate(w, "row-form.html", data); err != nil {
		slog.Error("EditRowForm template", "error", err)
	}
}

func (h *Handlers) UpdateRow(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
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
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.notFound(w, r)
		return
	}

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
		ID        int      `json:"id"`
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
		if row.ID > 0 {
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
			if row.ID > 0 {
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
