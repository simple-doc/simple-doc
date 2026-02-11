package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"docgen/config"
	"docgen/internal/db"

	"golang.org/x/crypto/bcrypt"
)

type AdminNavItem struct {
	Title    string
	Path     string
	IsActive bool
}

type AdminData struct {
	SiteTitle     string
	ThemeCSS      template.HTML
	NavItems      []AdminNavItem
	UserFirstname string
	IsEditor      bool
}

type AdminUsersData struct {
	AdminData
	Users []db.UserWithRoles
}

type AdminUserFormData struct {
	AdminData
	FormUser  db.User
	UserRoles []string
	AllRoles  []db.Role
	IsNew     bool
	ResetSent bool
}

type AdminRolesData struct {
	AdminData
	Roles []db.Role
}

type AdminRoleFormData struct {
	AdminData
	FormRole db.Role
	IsNew    bool
}

func adminNav(active string) []AdminNavItem {
	return []AdminNavItem{
		{Title: "Users", Path: "/admin/users", IsActive: active == "users"},
		{Title: "Roles", Path: "/admin/roles", IsActive: active == "roles"},
	}
}

// RequireAdmin wraps an http.HandlerFunc and returns 403 unless the user
// has the "admin" role.
func (h *Handlers) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := UserFromContext(r.Context())
		if u == nil {
			h.forbidden(w, r)
			return
		}
		isAdmin, _ := h.DB.HasRole(r.Context(), u.ID, "admin")
		if !isAdmin {
			h.forbidden(w, r)
			return
		}
		next(w, r)
	}
}

func (h *Handlers) adminData(r *http.Request, active string) AdminData {
	title, themeCSS := h.siteSettings(r.Context())
	return AdminData{
		SiteTitle:     title,
		ThemeCSS:      themeCSS,
		NavItems:      adminNav(active),
		UserFirstname: userFirstname(r.Context()),
		IsEditor:      true,
	}
}

// AdminIndex redirects to /admin/users.
func (h *Handlers) AdminIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/users", http.StatusFound)
}

// AdminUsers lists all users.
func (h *Handlers) AdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.DB.ListUsers(r.Context())
	if err != nil {
		h.serverError(w, r)
		slog.Error("AdminUsers", "error", err)
		return
	}

	data := AdminUsersData{
		AdminData: h.adminData(r, "users"),
		Users:     users,
	}

	if err := h.Tmpl.ExecuteTemplate(w, "admin-users.html", data); err != nil {
		slog.Error("AdminUsers template", "error", err)
	}
}

// AdminNewUserForm renders the create user form.
func (h *Handlers) AdminNewUserForm(w http.ResponseWriter, r *http.Request) {
	allRoles, _ := h.DB.ListAllRoles(r.Context())

	data := AdminUserFormData{
		AdminData: h.adminData(r, "users"),
		AllRoles:  allRoles,
		IsNew:     true,
	}

	if err := h.Tmpl.ExecuteTemplate(w, "admin-user-form.html", data); err != nil {
		slog.Error("AdminNewUserForm template", "error", err)
	}
}

// AdminCreateUser creates a new user.
func (h *Handlers) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	firstname := r.FormValue("firstname")
	lastname := r.FormValue("lastname")
	company := r.FormValue("company")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if firstname == "" || lastname == "" || email == "" || password == "" {
		http.Error(w, "firstname, lastname, email, and password are required", http.StatusBadRequest)
		return
	}

	if len(password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.serverError(w, r)
		slog.Error("AdminCreateUser bcrypt", "error", err)
		return
	}

	user, err := h.DB.CreateUser(r.Context(), firstname, lastname, company, email, string(hash))
	if err != nil {
		h.serverError(w, r)
		slog.Error("AdminCreateUser", "error", err)
		return
	}

	// Assign roles
	roleNames := r.Form["roles"]
	if len(roleNames) > 0 {
		if err := h.DB.SetUserRoles(r.Context(), user.ID, roleNames); err != nil {
			slog.Error("AdminCreateUser roles", "error", err)
		}
	}

	// Save history
	changedBy := userID(r.Context())
	version, _ := h.DB.GetUserVersion(r.Context(), user.ID)
	if err := h.DB.SaveUserHistory(r.Context(), user.ID, version, user.Firstname, user.Lastname, user.Company, user.Email, strings.Join(roleNames, ","), changedBy); err != nil {
		slog.Error("AdminCreateUser history", "error", err)
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// AdminEditUserForm renders the edit user form.
func (h *Handlers) AdminEditUserForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user, err := h.DB.GetUserByID(r.Context(), id)
	if err != nil {
		h.notFound(w, r)
		return
	}

	userRoles, _ := h.DB.GetUserRoles(r.Context(), id)
	allRoles, _ := h.DB.ListAllRoles(r.Context())

	data := AdminUserFormData{
		AdminData: h.adminData(r, "users"),
		FormUser:  user,
		UserRoles: userRoles,
		AllRoles:  allRoles,
		IsNew:     false,
		ResetSent: r.URL.Query().Get("reset_sent") == "1",
	}

	if err := h.Tmpl.ExecuteTemplate(w, "admin-user-form.html", data); err != nil {
		slog.Error("AdminEditUserForm template", "error", err)
	}
}

// AdminUpdateUser updates an existing user.
func (h *Handlers) AdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	firstname := r.FormValue("firstname")
	lastname := r.FormValue("lastname")
	company := r.FormValue("company")
	email := r.FormValue("email")

	if firstname == "" || lastname == "" || email == "" {
		http.Error(w, "firstname, lastname, and email are required", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password != "" && len(password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	user, err := h.DB.UpdateUser(r.Context(), id, firstname, lastname, company, email)
	if err != nil {
		h.serverError(w, r)
		slog.Error("AdminUpdateUser", "error", err)
		return
	}

	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			h.serverError(w, r)
			slog.Error("AdminUpdateUser bcrypt", "error", err)
			return
		}
		if err := h.DB.UpdateUserPassword(r.Context(), id, string(hash)); err != nil {
			h.serverError(w, r)
			slog.Error("AdminUpdateUser password", "error", err)
			return
		}
		// Invalidate any pending reset tokens
		if err := h.DB.DeletePasswordResetTokensForUser(r.Context(), id); err != nil {
			slog.Error("AdminUpdateUser delete reset tokens", "error", err)
		}
	}

	// Sync roles
	roleNames := r.Form["roles"]
	if err := h.DB.SetUserRoles(r.Context(), id, roleNames); err != nil {
		slog.Error("AdminUpdateUser roles", "error", err)
	}

	// Save history
	changedBy := userID(r.Context())
	version, _ := h.DB.GetUserVersion(r.Context(), user.ID)
	if err := h.DB.SaveUserHistory(r.Context(), user.ID, version, user.Firstname, user.Lastname, user.Company, user.Email, strings.Join(roleNames, ","), changedBy); err != nil {
		slog.Error("AdminUpdateUser history", "error", err)
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// AdminRoles lists all roles.
func (h *Handlers) AdminRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.DB.ListAllRoles(r.Context())
	if err != nil {
		h.serverError(w, r)
		slog.Error("AdminRoles", "error", err)
		return
	}

	data := AdminRolesData{
		AdminData: h.adminData(r, "roles"),
		Roles:     roles,
	}

	if err := h.Tmpl.ExecuteTemplate(w, "admin-roles.html", data); err != nil {
		slog.Error("AdminRoles template", "error", err)
	}
}

// AdminNewRoleForm renders the create role form.
func (h *Handlers) AdminNewRoleForm(w http.ResponseWriter, r *http.Request) {
	data := AdminRoleFormData{
		AdminData: h.adminData(r, "roles"),
		IsNew:     true,
	}

	if err := h.Tmpl.ExecuteTemplate(w, "admin-role-form.html", data); err != nil {
		slog.Error("AdminNewRoleForm template", "error", err)
	}
}

// AdminCreateRole creates a new role.
func (h *Handlers) AdminCreateRole(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	role, err := h.DB.CreateRole(r.Context(), name, description)
	if err != nil {
		h.serverError(w, r)
		slog.Error("AdminCreateRole", "error", err)
		return
	}

	changedBy := userID(r.Context())
	version, _ := h.DB.GetRoleVersion(r.Context(), role.ID)
	if err := h.DB.SaveRoleHistory(r.Context(), role.ID, version, role.Name, role.Description, changedBy); err != nil {
		slog.Error("AdminCreateRole history", "error", err)
	}

	http.Redirect(w, r, "/admin/roles", http.StatusSeeOther)
}

// AdminEditRoleForm renders the edit role form.
func (h *Handlers) AdminEditRoleForm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	role, err := h.DB.GetRole(r.Context(), id)
	if err != nil {
		h.notFound(w, r)
		return
	}

	data := AdminRoleFormData{
		AdminData: h.adminData(r, "roles"),
		FormRole:  role,
		IsNew:     false,
	}

	if err := h.Tmpl.ExecuteTemplate(w, "admin-role-form.html", data); err != nil {
		slog.Error("AdminEditRoleForm template", "error", err)
	}
}

// AdminUpdateRole updates an existing role.
func (h *Handlers) AdminUpdateRole(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	role, err := h.DB.UpdateRole(r.Context(), id, name, description)
	if err != nil {
		h.serverError(w, r)
		slog.Error("AdminUpdateRole", "error", err)
		return
	}

	changedBy := userID(r.Context())
	version, _ := h.DB.GetRoleVersion(r.Context(), role.ID)
	if err := h.DB.SaveRoleHistory(r.Context(), role.ID, version, role.Name, role.Description, changedBy); err != nil {
		slog.Error("AdminUpdateRole history", "error", err)
	}

	http.Redirect(w, r, "/admin/roles", http.StatusSeeOther)
}

// AdminSendResetPassword generates a reset token and emails the user.
func (h *Handlers) AdminSendResetPassword(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user, err := h.DB.GetUserByID(r.Context(), id)
	if err != nil {
		h.notFound(w, r)
		return
	}

	// Invalidate any existing tokens for this user
	if err := h.DB.DeletePasswordResetTokensForUser(r.Context(), id); err != nil {
		slog.Error("AdminSendResetPassword delete tokens", "error", err)
	}

	token, err := generateToken()
	if err != nil {
		h.serverError(w, r)
		slog.Error("AdminSendResetPassword token", "error", err)
		return
	}

	expiresAt := time.Now().Add(48 * time.Hour)
	if _, err := h.DB.CreatePasswordResetToken(r.Context(), id, token, expiresAt); err != nil {
		h.serverError(w, r)
		slog.Error("AdminSendResetPassword create token", "error", err)
		return
	}

	resetURL := config.BaseURL() + "/reset-password?token=" + token

	settings, _ := h.DB.GetSiteSettings(r.Context())
	siteTitle := settings.SiteTitle

	subject := fmt.Sprintf("[%s] Reset your password", siteTitle)
	body := fmt.Sprintf("Hello %s,\r\n\r\n"+
		"An administrator of %s has requested a password reset for your account.\r\n\r\n"+
		"Click the link below to set a new password:\r\n%s\r\n\r\n"+
		"This link expires in 48 hours.\r\n\r\n"+
		"If you did not expect this email, you can safely ignore it.\r\n",
		user.Firstname, siteTitle, resetURL)

	if err := sendEmail(user.Email, subject, body); err != nil {
		h.serverError(w, r)
		slog.Error("AdminSendResetPassword email", "error", err)
		return
	}

	http.Redirect(w, r, "/admin/users/"+id+"/edit?reset_sent=1", http.StatusSeeOther)
}

func sendEmail(to, subject, body string) error {
	from := config.SMTPFrom()
	host := config.SMTPHost()
	port := config.SMTPPort()
	addr := host + ":" + port

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" + body

	var auth smtp.Auth
	if user := config.SMTPUser(); user != "" {
		auth = smtp.PlainAuth("", user, config.SMTPPass(), host)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
