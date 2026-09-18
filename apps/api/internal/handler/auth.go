package handler

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"flocal/internal/config"
	"flocal/internal/middleware"
	"flocal/internal/models"
	"flocal/internal/service"
	"flocal/internal/utils"
)

const oauthStateCookie = "flocal_oauth_state"
const oauthLinkStateCookie = "flocal_oauth_link_state"
const maxAvatarUploadBytes = 5 << 20

type RequestMagicLinkRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ConsumeMagicLinkRequest struct {
	Email string `json:"email" validate:"required,email"`
	Token string `json:"token" validate:"required"`
}

type SessionResponse struct {
	User      *models.User `json:"user"`
	Token     string       `json:"token"`
	CsrfToken string       `json:"csrfToken"`
	ExpiresAt time.Time    `json:"expiresAt"`
}

type AuthHandler struct {
	cfg          *config.Config
	svc          *service.AuthService
	emailLimiter *middleware.IPRateLimiter
	geo          *utils.GeoIP
}

func NewAuthHandler(
	cfg *config.Config,
	svc *service.AuthService,
	emailLimiter *middleware.IPRateLimiter,
	geo *utils.GeoIP,
) *AuthHandler {
	return &AuthHandler{
		cfg:          cfg,
		svc:          svc,
		emailLimiter: emailLimiter,
		geo:          geo,
	}
}

type UpdateProfileRequest struct {
	Name                   string  `json:"name" validate:"required"`
	Username               *string `json:"username"`
	Image                  *string `json:"image"`
	DefaultPrepTimeSeconds int     `json:"defaultPrepTimeSeconds"`
	Timezone               *string `json:"timezone"`
}

type RequestEmailChangeRequest struct {
	NewEmail string `json:"newEmail" validate:"required,email"`
}

type ConsumeEmailChangeRequest struct {
	NewEmail string `json:"newEmail" validate:"required,email"`
	Token    string `json:"token" validate:"required"`
}

type SessionSummary struct {
	*models.Session
	Current bool `json:"current"`
}

func (h *AuthHandler) RegisterRoutes(
	rg *gin.RouterGroup,
	authValidator middleware.SessionValidator,
	authLimiter gin.HandlerFunc,
	cookies middleware.CookieConfig,
) {
	rg.GET("/google", authLimiter, h.GoogleLogin)
	rg.GET("/google/callback", authLimiter, h.GoogleCallback)

	rg.GET("/github", authLimiter, h.GitHubLogin)
	rg.GET("/github/callback", authLimiter, h.GitHubCallback)

	rg.POST("/magic-link", authLimiter, h.RequestMagicLink)
	rg.GET("/magic-link/verify", h.MagicLinkLanding)
	rg.POST("/magic-link/consume", authLimiter, h.ConsumeMagicLink)

	rg.GET("/accounts", middleware.RequireAuth(authValidator, cookies), h.ListLinkedAccounts)
	rg.GET("/google/link", middleware.RequireAuth(authValidator, cookies), h.GoogleLink)
	rg.GET("/github/link", middleware.RequireAuth(authValidator, cookies), h.GitHubLink)

	rg.POST("/logout", middleware.RequireAuth(authValidator, cookies), h.Logout)
	rg.GET("/me", middleware.RequireAuth(authValidator, cookies), h.Me)

	rg.PATCH("/me", middleware.RequireAuth(authValidator, cookies), h.UpdateProfile)
	rg.POST("/me/avatar", middleware.RequireAuth(authValidator, cookies), h.UploadAvatar)
	rg.DELETE("/me", middleware.RequireAuth(authValidator, cookies), h.DeleteAccount)

	rg.GET("/email/verify", h.EmailChangeLanding)
	rg.POST("/email/change", middleware.RequireAuth(authValidator, cookies), h.RequestEmailChange)
	rg.POST("/email/verify/confirm", middleware.RequireAuth(authValidator, cookies), h.ConsumeEmailChange)

	rg.GET("/sessions", middleware.RequireAuth(authValidator, cookies), h.ListSessions)
	rg.DELETE("/sessions/:id", middleware.RequireAuth(authValidator, cookies), h.RevokeSession)
}

func (h *AuthHandler) resolveCountryCode(c *gin.Context) *string {
	if h.geo == nil {
		return nil
	}

	ip := h.resolveClientIP(c)
	if ip == nil {
		return nil
	}

	code, ok := h.geo.CountryCode(ip)
	if !ok {
		return nil
	}

	return &code
}

func (h *AuthHandler) resolveClientIP(c *gin.Context) net.IP {
	raw := c.ClientIP()
	ip := net.ParseIP(raw)

	if ip == nil {
		return nil
	}

	if !h.cfg.IsProduction() && ip.IsLoopback() {
		return net.ParseIP("8.8.8.8")
	}

	return ip
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	if !h.svc.GoogleEnabled() {
		middleware.NotFound(c, "Google sign-in is not available.")
		return
	}

	state, err := utils.NewToken(24)
	if err != nil {
		middleware.Internal(c, "")
		return
	}

	h.setOAuthStateCookie(c, state)

	authURL, err := h.svc.GoogleAuthURL(state)
	if err != nil {
		middleware.Internal(c, "Could not start Google sign-in.")
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if userID, ok := h.verifyOAuthLinkState(c); ok {
		h.completeLink(c, "google", userID)
		return
	}

	if !h.verifyOAuthState(c) {
		h.redirectWithError(c, "invalid_state")
		return
	}

	code := c.Query("code")
	if code == "" {
		h.redirectWithError(c, "missing_code")
		return
	}

	issued, err := h.svc.CompleteGoogleLogin(
		c.Request.Context(),
		code,
		c.ClientIP(),
		c.Request.UserAgent(),
		h.resolveCountryCode(c),
	)
	if err != nil {
		if errors.Is(err, service.ErrEmailNotVerifiedByProvider) {
			h.redirectWithError(c, "email_not_verified")
			return
		}
		log.Printf("auth: google callback failed: %v", err)
		h.redirectWithError(c, "google_login_failed")
		return
	}

	if _, err := h.issueSessionCookies(c, issued.RawToken, issued.ExpiresAt); err != nil {
		log.Printf("auth: issue google session cookies failed: %v", err)
		h.redirectWithError(c, "session_failed")
		return
	}

	c.Redirect(
		http.StatusFound,
		strings.TrimSuffix(h.cfg.App.FrontendURL, "/")+"/",
	)
}

func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	if !h.svc.GitHubEnabled() {
		middleware.NotFound(c, "GitHub sign-in is not available.")
		return
	}

	state, err := utils.NewToken(24)
	if err != nil {
		middleware.Internal(c, "")
		return
	}

	h.setOAuthStateCookie(c, state)

	authURL, err := h.svc.GitHubAuthURL(state)
	if err != nil {
		middleware.Internal(c, "Could not start GitHub sign-in.")
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	if userID, ok := h.verifyOAuthLinkState(c); ok {
		h.completeLink(c, "github", userID)
		return
	}

	if !h.verifyOAuthState(c) {
		h.redirectWithError(c, "invalid_state")
		return
	}

	code := c.Query("code")
	if code == "" {
		h.redirectWithError(c, "missing_code")
		return
	}

	issued, err := h.svc.CompleteGitHubLogin(
		c.Request.Context(),
		code,
		c.ClientIP(),
		c.Request.UserAgent(),
		h.resolveCountryCode(c),
	)
	if err != nil {
		if errors.Is(err, service.ErrEmailNotVerifiedByProvider) {
			h.redirectWithError(c, "email_not_verified")
			return
		}
		log.Printf("auth: github callback failed: %v", err)
		h.redirectWithError(c, "github_login_failed")
		return
	}

	if _, err := h.issueSessionCookies(c, issued.RawToken, issued.ExpiresAt); err != nil {
		log.Printf("auth: issue github session cookies failed: %v", err)
		h.redirectWithError(c, "session_failed")
		return
	}

	c.Redirect(
		http.StatusFound,
		strings.TrimSuffix(h.cfg.App.FrontendURL, "/")+"/",
	)
}

func (h *AuthHandler) RequestMagicLink(c *gin.Context) {
	var req RequestMagicLinkRequest

	if !middleware.BindAndValidate(c, &req) {
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	if !h.emailLimiter.Allow(email) {
		middleware.RateLimited(
			c,
			"Too many sign-in requests for this email. Please wait a few minutes and try again.",
		)
		return
	}

	if err := h.svc.RequestMagicLink(
		c.Request.Context(),
		email,
	); err != nil {
		log.Printf("auth: request magic link failed: %v", err)
		middleware.Internal(c, "Could not send sign-in link. Please try again.")
		return
	}

	middleware.OK(c, gin.H{
		"message": "Check your email for a sign-in link.",
	})
}

func (h *AuthHandler) MagicLinkLanding(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	token := strings.TrimSpace(c.Query("token"))

	if email == "" || token == "" {
		h.redirectWithError(c, "missing_token")
		return
	}

	dest := fmt.Sprintf(
		"%s/verify?email=%s&token=%s",
		strings.TrimSuffix(h.cfg.App.FrontendURL, "/"),
		url.QueryEscape(email),
		url.QueryEscape(token),
	)

	c.Redirect(http.StatusFound, dest)
}

func (h *AuthHandler) ConsumeMagicLink(c *gin.Context) {
	var req ConsumeMagicLinkRequest

	if !middleware.BindAndValidate(c, &req) {
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	issued, err := h.svc.ConsumeMagicLink(
		c.Request.Context(),
		email,
		req.Token,
		c.ClientIP(),
		c.Request.UserAgent(),
		h.resolveCountryCode(c),
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrExpiredLink) {
			middleware.Fail(
				c,
				http.StatusBadRequest,
				middleware.CodeBadRequest,
				"This sign-in link is invalid or has expired.",
			)
			return
		}

		log.Printf("auth: consume magic link failed: %v", err)
		middleware.Internal(c, "Could not complete sign-in.")
		return
	}

	csrfToken, err := h.issueSessionCookies(c, issued.RawToken, issued.ExpiresAt)
	if err != nil {
		log.Printf("auth: issue magic link session cookies failed: %v", err)
		middleware.Internal(c, "Could not complete sign-in.")
		return
	}

	middleware.OK(c, SessionResponse{
		User:      issued.User,
		Token:     issued.RawToken,
		CsrfToken: csrfToken,
		ExpiresAt: issued.ExpiresAt,
	})
}

func (h *AuthHandler) ListLinkedAccounts(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	accounts, err := h.svc.ListLinkedAccounts(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("auth: list linked accounts failed: %v", err)
		middleware.Internal(c, "Could not load linked accounts.")
		return
	}

	middleware.OK(c, accounts)
}

func (h *AuthHandler) GoogleLink(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	if !h.svc.GoogleEnabled() {
		middleware.NotFound(c, "Google sign-in is not available.")
		return
	}

	state, err := utils.NewToken(24)
	if err != nil {
		middleware.Internal(c, "")
		return
	}

	h.setOAuthLinkStateCookie(c, user.ID, state)

	authURL, err := h.svc.GoogleAuthURL(state)
	if err != nil {
		middleware.Internal(c, "Could not start Google sign-in.")
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) GitHubLink(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	if !h.svc.GitHubEnabled() {
		middleware.NotFound(c, "GitHub sign-in is not available.")
		return
	}

	state, err := utils.NewToken(24)
	if err != nil {
		middleware.Internal(c, "")
		return
	}

	h.setOAuthLinkStateCookie(c, user.ID, state)

	authURL, err := h.svc.GitHubAuthURL(state)
	if err != nil {
		middleware.Internal(c, "Could not start GitHub sign-in.")
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	session := middleware.CurrentSession(c)

	resp := SessionResponse{
		User: user,
	}

	if session != nil {
		resp.ExpiresAt = session.ExpiresAt
	}

	middleware.OK(c, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, _ := middleware.ExtractToken(c)

	if err := h.svc.Logout(c.Request.Context(), token); err != nil {
		log.Printf("auth: logout failed: %v", err)
	}

	h.clearSessionCookie(c)

	middleware.OK(c, gin.H{
		"message": "Signed out.",
	})
}

func (h *AuthHandler) issueSessionCookies(
	c *gin.Context,
	rawToken string,
	expiresAt time.Time,
) (string, error) {
	csrfToken, err := utils.NewToken(32)
	if err != nil {
		return "", fmt.Errorf("generate csrf token: %w", err)
	}

	middleware.SetSessionCookie(
		c,
		h.cfg.Auth.CookieDomain,
		h.cfg.Auth.CookieSecure,
		rawToken,
		expiresAt,
	)

	middleware.SetCSRFCookie(
		c,
		h.cfg.Auth.CookieDomain,
		h.cfg.Auth.CookieSecure,
		csrfToken,
		expiresAt,
	)

	return csrfToken, nil
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	middleware.ClearSessionCookie(c, h.cfg.Auth.CookieDomain, h.cfg.Auth.CookieSecure)
	middleware.ClearCSRFCookie(c, h.cfg.Auth.CookieDomain, h.cfg.Auth.CookieSecure)
}

func (h *AuthHandler) setOAuthStateCookie(c *gin.Context, state string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, state, 300, "/", h.cfg.Auth.CookieDomain, h.cfg.Auth.CookieSecure, true)
}

func (h *AuthHandler) verifyOAuthState(c *gin.Context) bool {
	cookieState, err := c.Cookie(oauthStateCookie)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, "", -1, "/", h.cfg.Auth.CookieDomain, h.cfg.Auth.CookieSecure, true)
	if err != nil || cookieState == "" {
		return false
	}
	queryState := c.Query("state")
	if queryState == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieState), []byte(queryState)) == 1
}

func (h *AuthHandler) setOAuthLinkStateCookie(c *gin.Context, userID, state string) {
	sig := utils.HMACSHA256Hex("oauth-link."+userID+"."+state, h.cfg.Auth.SessionHashKey)
	value := userID + "." + state + "." + sig
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthLinkStateCookie, value, 300, "/", h.cfg.Auth.CookieDomain, h.cfg.Auth.CookieSecure, true)
}

func (h *AuthHandler) clearOAuthLinkStateCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthLinkStateCookie, "", -1, "/", h.cfg.Auth.CookieDomain, h.cfg.Auth.CookieSecure, true)
}

func (h *AuthHandler) verifyOAuthLinkState(c *gin.Context) (string, bool) {
	cookieVal, err := c.Cookie(oauthLinkStateCookie)
	if err != nil || cookieVal == "" {
		return "", false
	}

	h.clearOAuthLinkStateCookie(c)

	parts := strings.SplitN(cookieVal, ".", 3)
	if len(parts) != 3 {
		return "", false
	}

	userID, state, sig := parts[0], parts[1], parts[2]

	expectedSig := utils.HMACSHA256Hex("oauth-link."+userID+"."+state, h.cfg.Auth.SessionHashKey)
	if subtle.ConstantTimeCompare([]byte(sig), []byte(expectedSig)) != 1 {
		return "", false
	}

	queryState := c.Query("state")
	if queryState == "" || subtle.ConstantTimeCompare([]byte(state), []byte(queryState)) != 1 {
		return "", false
	}

	return userID, true
}

func (h *AuthHandler) completeLink(c *gin.Context, provider, userID string) {
	code := c.Query("code")
	if code == "" {
		h.redirectWithLinkError(c, "missing_code")
		return
	}

	var err error
	switch provider {
	case "google":
		_, err = h.svc.LinkGoogleAccount(c.Request.Context(), userID, code)
	case "github":
		_, err = h.svc.LinkGitHubAccount(c.Request.Context(), userID, code)
	}

	if err != nil {
		log.Printf("auth: link %s account failed: %v", provider, err)
		if errors.Is(err, service.ErrProviderAlreadyLinked) {
			h.redirectWithLinkError(c, "already_linked")
			return
		}
		h.redirectWithLinkError(c, "link_failed")
		return
	}

	dest := fmt.Sprintf(
		"%s/profile?linked=%s",
		strings.TrimSuffix(h.cfg.App.FrontendURL, "/"),
		provider,
	)
	c.Redirect(http.StatusFound, dest)
}

func (h *AuthHandler) redirectWithLinkError(c *gin.Context, code string) {
	dest := fmt.Sprintf(
		"%s/profile?linkError=%s",
		strings.TrimSuffix(h.cfg.App.FrontendURL, "/"),
		url.QueryEscape(code),
	)
	c.Redirect(http.StatusFound, dest)
}

func (h *AuthHandler) redirectWithError(c *gin.Context, code string) {
	dest := fmt.Sprintf(
		"%s/verify?error=%s",
		strings.TrimSuffix(h.cfg.App.FrontendURL, "/"),
		url.QueryEscape(code),
	)

	c.Redirect(http.StatusFound, dest)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req UpdateProfileRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	updated, err := h.svc.UpdateProfile(c.Request.Context(), user.ID, service.UpdateProfileInput{
		Name:                   req.Name,
		Username:               req.Username,
		Image:                  req.Image,
		DefaultPrepTimeSeconds: req.DefaultPrepTimeSeconds,
		Timezone:               req.Timezone,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			middleware.Fail(c, http.StatusConflict, middleware.CodeBadRequest, "That username is already taken.")
		default:
			log.Printf("auth: update profile failed: %v", err)
			middleware.Internal(c, "Could not update profile.")
		}
		return
	}

	middleware.OK(c, updated)
}

func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarUploadBytes)

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "No avatar file provided.")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		middleware.Internal(c, "Could not read avatar file.")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Avatar file is too large or unreadable.")
		return
	}

	contentType := fileHeader.Header.Get("Content-Type")

	updated, err := h.svc.UploadAvatar(c.Request.Context(), user.ID, data, contentType)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAvatarTooLarge):
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Avatar must be under 5MB.")
		case errors.Is(err, service.ErrAvatarInvalidType):
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Avatar must be a JPEG, PNG, or WEBP image.")
		default:
			log.Printf("auth: upload avatar failed: %v", err)
			middleware.Internal(c, "Could not upload avatar.")
		}
		return
	}

	middleware.OK(c, updated)
}

func (h *AuthHandler) EmailChangeLanding(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	token := strings.TrimSpace(c.Query("token"))

	if email == "" || token == "" {
		h.redirectWithError(c, "missing_token")
		return
	}

	dest := fmt.Sprintf(
		"%s/profile/settings/verify-email?email=%s&token=%s",
		strings.TrimSuffix(h.cfg.App.FrontendURL, "/"),
		url.QueryEscape(email),
		url.QueryEscape(token),
	)

	c.Redirect(http.StatusFound, dest)
}

func (h *AuthHandler) RequestEmailChange(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req RequestEmailChangeRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	if err := h.svc.RequestEmailChange(c.Request.Context(), user.ID, req.NewEmail); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			middleware.Fail(c, http.StatusConflict, middleware.CodeBadRequest, "That email is already in use.")
		default:
			log.Printf("auth: request email change failed: %v", err)
			middleware.Internal(c, "Could not send verification email.")
		}
		return
	}

	middleware.OK(c, gin.H{"message": "Check your new email for a verification link."})
}

func (h *AuthHandler) ConsumeEmailChange(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	var req ConsumeEmailChangeRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	updated, err := h.svc.ConsumeEmailChange(c.Request.Context(), user.ID, req.NewEmail, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOrExpiredLink):
			middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "This verification link is invalid or has expired.")
		case errors.Is(err, service.ErrEmailTaken):
			middleware.Fail(c, http.StatusConflict, middleware.CodeBadRequest, "That email is already in use.")
		default:
			log.Printf("auth: consume email change failed: %v", err)
			middleware.Internal(c, "Could not update email.")
		}
		return
	}

	middleware.OK(c, updated)
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteAccount(c.Request.Context(), user.ID); err != nil {
		log.Printf("auth: delete account failed: %v", err)
		middleware.Internal(c, "Could not delete account.")
		return
	}

	h.clearSessionCookie(c)

	middleware.OK(c, gin.H{"message": "Account deleted."})
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	current := middleware.CurrentSession(c)

	sessions, err := h.svc.ListSessions(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("auth: list sessions failed: %v", err)
		middleware.Internal(c, "Could not load sessions.")
		return
	}

	summaries := make([]SessionSummary, 0, len(sessions))
	for i := range sessions {
		summaries = append(summaries, SessionSummary{
			Session: &sessions[i],
			Current: current != nil && current.ID == sessions[i].ID,
		})
	}

	middleware.OK(c, summaries)
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	user, ok := middleware.RequireUser(c)
	if !ok {
		return
	}

	id := c.Param("id")

	current := middleware.CurrentSession(c)
	if current != nil && current.ID == id {
		middleware.Fail(c, http.StatusBadRequest, middleware.CodeBadRequest, "Use logout to sign out of your current session.")
		return
	}

	if err := h.svc.RevokeSession(c.Request.Context(), id, user.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFoundForUser):
			middleware.NotFound(c, "That session does not exist.")
		default:
			log.Printf("auth: revoke session failed: %v", err)
			middleware.Internal(c, "Could not revoke session.")
		}
		return
	}

	middleware.OK(c, gin.H{"message": "Session revoked."})
}
