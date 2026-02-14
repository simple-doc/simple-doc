package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"docgen/internal/db"

	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const userContextKey contextKey = "user"
const previewRolesContextKey contextKey = "preview_roles"
const sessionTokenContextKey contextKey = "session_token"

const (
	sessionCookieName = "session_token"
	sessionDuration   = 24 * time.Hour
)

const challengeThreshold = 3

// challengeSecret is a random key generated at startup for HMAC-signing challenge answers.
var challengeSecret []byte

func init() {
	challengeSecret = make([]byte, 32)
	if _, err := rand.Read(challengeSecret); err != nil {
		panic("failed to generate challenge secret: " + err.Error())
	}
}

type failedLogin struct {
	Count    int
	LastFail time.Time
}

var (
	failedLogins   = make(map[string]*failedLogin)
	failedLoginsMu sync.Mutex
)

func getClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.SplitN(fwd, ",", 2)[0]
	}
	return strings.SplitN(r.RemoteAddr, ":", 2)[0]
}

func getFailCount(ip string) int {
	failedLoginsMu.Lock()
	defer failedLoginsMu.Unlock()
	fl, ok := failedLogins[ip]
	if !ok {
		return 0
	}
	// Reset after 15 minutes of no failures
	if time.Since(fl.LastFail) > 15*time.Minute {
		delete(failedLogins, ip)
		return 0
	}
	return fl.Count
}

func recordFail(ip string) int {
	failedLoginsMu.Lock()
	defer failedLoginsMu.Unlock()
	fl, ok := failedLogins[ip]
	if !ok || time.Since(fl.LastFail) > 15*time.Minute {
		failedLogins[ip] = &failedLogin{Count: 1, LastFail: time.Now()}
		return 1
	}
	fl.Count++
	fl.LastFail = time.Now()
	return fl.Count
}

func clearFails(ip string) {
	failedLoginsMu.Lock()
	defer failedLoginsMu.Unlock()
	delete(failedLogins, ip)
}

var numberWords = []string{"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
	"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen", "twenty"}

func randInt(max int64) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(max))
	return int(n.Int64())
}

func formatNum(n int, useWord bool) string {
	if useWord && n >= 1 && n <= 20 {
		return numberWords[n]
	}
	return strconv.Itoa(n)
}

type challenge struct {
	Question string
	Token    string
}

func generateChallenge() challenge {
	variant := randInt(4)
	var question string
	var answer int

	switch variant {
	case 0:
		// a × b + c  (e.g. "three × 4 + 2")
		a := randInt(8) + 2  // 2-9
		b := randInt(8) + 2  // 2-9
		c := randInt(9) + 1  // 1-9
		answer = int(a)*int(b) + int(c)
		useWord := randInt(2) == 0
		question = fmt.Sprintf("%s × %s + %s",
			formatNum(int(a), useWord),
			formatNum(int(b), !useWord),
			formatNum(int(c), randInt(2) == 0))
	case 1:
		// a × b - c  (ensure positive, e.g. "6 × seven - 3")
		a := randInt(8) + 2
		b := randInt(8) + 2
		product := int(a) * int(b)
		c := randInt(int64(product-1)) + 1
		answer = product - int(c)
		useWord := randInt(2) == 0
		question = fmt.Sprintf("%s × %s - %d",
			formatNum(int(a), useWord),
			formatNum(int(b), !useWord),
			int(c))
	case 2:
		// word-based addition: "twelve + fifteen"
		a := randInt(19) + 2  // 2-20
		b := randInt(19) + 2  // 2-20
		answer = int(a) + int(b)
		question = fmt.Sprintf("%s + %s",
			formatNum(int(a), true),
			formatNum(int(b), true))
	default:
		// a + b × c  (e.g. "5 + three × 4")
		a := randInt(15) + 1 // 1-15
		b := randInt(8) + 2  // 2-9
		c := randInt(8) + 2  // 2-9
		answer = int(a) + int(b)*int(c)
		useWord := randInt(2) == 0
		question = fmt.Sprintf("%s + %s × %s",
			formatNum(int(a), randInt(2) == 0),
			formatNum(int(b), useWord),
			formatNum(int(c), !useWord))
	}

	return challenge{Question: question, Token: signAnswer(answer)}
}

func signAnswer(answer int) string {
	mac := hmac.New(sha256.New, challengeSecret)
	mac.Write([]byte(strconv.Itoa(answer)))
	return hex.EncodeToString(mac.Sum(nil))
}

func verifyChallenge(userAnswer, token string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(userAnswer))
	if err != nil {
		return false
	}
	expected := signAnswer(n)
	return hmac.Equal([]byte(expected), []byte(token))
}

// UserFromContext extracts the authenticated user from the request context.
func UserFromContext(ctx context.Context) *db.User {
	u, _ := ctx.Value(userContextKey).(*db.User)
	return u
}

// PreviewRolesFromContext returns the preview roles if preview mode is active.
func PreviewRolesFromContext(ctx context.Context) []string {
	v, ok := ctx.Value(previewRolesContextKey).(string)
	if !ok {
		return nil
	}
	if v == "" {
		return []string{}
	}
	return strings.Split(v, ",")
}

// inPreviewMode returns true if the session is in preview mode.
func inPreviewMode(ctx context.Context) bool {
	return PreviewRolesFromContext(ctx) != nil
}

// sessionTokenFromContext returns the session token stored in context.
func sessionTokenFromContext(ctx context.Context) string {
	s, _ := ctx.Value(sessionTokenContextKey).(string)
	return s
}

func generateToken() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type LoginData struct {
	SiteTitle      string
	ThemeCSS       template.HTML
	Error          string
	ShowChallenge  bool
	ChallengeQ     string
	ChallengeToken string
}

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	title, themeCSS := h.siteSettings(r.Context())
	data := LoginData{
		SiteTitle: title,
		ThemeCSS:  themeCSS,
	}
	if err := h.Tmpl.ExecuteTemplate(w, "login.html", data); err != nil {
		slog.Error("LoginPage template", "error", err)
	}
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	ip := getClientIP(r)

	if email == "" || password == "" {
		h.renderLoginError(w, r, "Email and password are required")
		return
	}

	// If challenge is required, verify it first
	if getFailCount(ip) >= challengeThreshold {
		cAnswer := r.FormValue("challenge_answer")
		cToken := r.FormValue("challenge_token")
		if cAnswer == "" || cToken == "" || !verifyChallenge(cAnswer, cToken) {
			recordFail(ip)
			h.renderLoginError(w, r, "Incorrect answer to the security challenge")
			return
		}
	}

	user, err := h.DB.GetUserByEmail(r.Context(), email)
	if err != nil {
		recordFail(ip)
		h.renderLoginError(w, r, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		recordFail(ip)
		h.renderLoginError(w, r, "Invalid email or password")
		return
	}

	// Success — clear fail counter
	clearFails(ip)

	token, err := generateToken()
	if err != nil {
		h.serverError(w, r)
		slog.Error("Login generateToken", "error", err)
		return
	}

	expiresAt := time.Now().Add(sessionDuration)
	if _, err := h.DB.CreateSession(r.Context(), user.ID, token, expiresAt); err != nil {
		h.serverError(w, r)
		slog.Error("Login CreateSession", "error", err)
		return
	}

	if err := h.DB.UpdateLastLogin(r.Context(), user.ID); err != nil {
		slog.Error("Login UpdateLastLogin", "error", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		if err := h.DB.DeleteSession(r.Context(), cookie.Value); err != nil {
			slog.Error("Logout DeleteSession", "error", err)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handlers) renderLoginError(w http.ResponseWriter, r *http.Request, msg string) {
	title, themeCSS := h.siteSettings(r.Context())
	ip := getClientIP(r)
	data := LoginData{
		SiteTitle: title,
		ThemeCSS:  themeCSS,
		Error:     msg,
	}
	if getFailCount(ip) >= challengeThreshold {
		c := generateChallenge()
		data.ShowChallenge = true
		data.ChallengeQ = c.Question
		data.ChallengeToken = c.Token
	}
	w.WriteHeader(http.StatusUnauthorized)
	if err := h.Tmpl.ExecuteTemplate(w, "login.html", data); err != nil {
		slog.Error("renderLoginError template", "error", err)
	}
}

type ResetPasswordData struct {
	SiteTitle string
	ThemeCSS  template.HTML
	Token     string
	Error     string
	Success   bool
}

func (h *Handlers) ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.notFound(w, r)
		return
	}

	if _, err := h.DB.GetPasswordResetToken(r.Context(), token); err != nil {
		h.notFound(w, r)
		return
	}

	title, themeCSS := h.siteSettings(r.Context())
	data := ResetPasswordData{
		SiteTitle: title,
		ThemeCSS:  themeCSS,
		Token:     token,
	}
	if err := h.Tmpl.ExecuteTemplate(w, "reset-password.html", data); err != nil {
		slog.Error("ResetPasswordPage template", "error", err)
	}
}

func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	title, themeCSS := h.siteSettings(r.Context())

	if token == "" {
		h.notFound(w, r)
		return
	}

	rt, err := h.DB.GetPasswordResetToken(r.Context(), token)
	if err != nil {
		data := ResetPasswordData{SiteTitle: title, ThemeCSS: themeCSS, Token: token,
			Error: "This reset link has expired or is invalid"}
		w.WriteHeader(http.StatusBadRequest)
		h.Tmpl.ExecuteTemplate(w, "reset-password.html", data)
		return
	}

	if password == "" || len(password) < 8 {
		data := ResetPasswordData{SiteTitle: title, ThemeCSS: themeCSS, Token: token,
			Error: "Password must be at least 8 characters"}
		w.WriteHeader(http.StatusBadRequest)
		h.Tmpl.ExecuteTemplate(w, "reset-password.html", data)
		return
	}

	if password != confirm {
		data := ResetPasswordData{SiteTitle: title, ThemeCSS: themeCSS, Token: token,
			Error: "Passwords do not match"}
		w.WriteHeader(http.StatusBadRequest)
		h.Tmpl.ExecuteTemplate(w, "reset-password.html", data)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.serverError(w, r)
		slog.Error("ResetPassword bcrypt", "error", err)
		return
	}

	if err := h.DB.UpdateUserPassword(r.Context(), rt.UserID, string(hash)); err != nil {
		h.serverError(w, r)
		slog.Error("ResetPassword update", "error", err)
		return
	}

	// Invalidate all reset tokens for this user
	if err := h.DB.DeletePasswordResetTokensForUser(r.Context(), rt.UserID); err != nil {
		slog.Error("ResetPassword delete tokens", "error", err)
	}

	data := ResetPasswordData{SiteTitle: title, ThemeCSS: themeCSS, Success: true}
	if err := h.Tmpl.ExecuteTemplate(w, "reset-password.html", data); err != nil {
		slog.Error("ResetPassword template", "error", err)
	}
}

// canAccessSection checks whether the current user may view a section that
// requires the given role.  Empty requiredRole means no restriction.
func (h *Handlers) canAccessSection(ctx context.Context, requiredRole string) bool {
	if requiredRole == "" {
		return true
	}
	if inPreviewMode(ctx) {
		for _, r := range PreviewRolesFromContext(ctx) {
			if r == requiredRole {
				return true
			}
		}
		return false
	}
	u := UserFromContext(ctx)
	if u == nil {
		return false
	}
	isAdmin, _ := h.DB.HasRole(ctx, u.ID, "admin")
	if isAdmin {
		return true
	}
	has, _ := h.DB.HasRole(ctx, u.ID, requiredRole)
	return has
}

// RequireEditor wraps an http.HandlerFunc and returns 403 unless the user
// has the "editor" or "admin" role.
func (h *Handlers) RequireEditor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if inPreviewMode(r.Context()) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		u := UserFromContext(r.Context())
		if u == nil {
			h.forbidden(w, r)
			return
		}
		isAdmin, _ := h.DB.HasRole(r.Context(), u.ID, "admin")
		if isAdmin {
			next(w, r)
			return
		}
		isEd, _ := h.DB.HasRole(r.Context(), u.ID, "editor")
		if isEd {
			next(w, r)
			return
		}
		h.forbidden(w, r)
	}
}

func (h *Handlers) isAdmin(ctx context.Context) bool {
	if inPreviewMode(ctx) {
		return false
	}
	u := UserFromContext(ctx)
	if u == nil {
		return false
	}
	isAdmin, _ := h.DB.HasRole(ctx, u.ID, "admin")
	return isAdmin
}

func (h *Handlers) isEditor(ctx context.Context) bool {
	if inPreviewMode(ctx) {
		return false
	}
	u := UserFromContext(ctx)
	if u == nil {
		return false
	}
	isAdmin, _ := h.DB.HasRole(ctx, u.ID, "admin")
	if isAdmin {
		return true
	}
	isEd, _ := h.DB.HasRole(ctx, u.ID, "editor")
	return isEd
}

// RequireAuth wraps an http.Handler and enforces authentication on all routes
// except /login.
func (h *Handlers) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" || r.URL.Path == "/reset-password" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := h.DB.GetSessionByToken(r.Context(), cookie.Value)
		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := h.DB.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, &user)
		ctx = context.WithValue(ctx, sessionTokenContextKey, session.Token)
		if session.PreviewRoles != nil {
			ctx = context.WithValue(ctx, previewRolesContextKey, *session.PreviewRoles)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
