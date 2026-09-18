// auth.go
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"flocal/internal/models"
	"flocal/internal/service"
)

const (
	ctxKeyUser    = "auth.user"
	ctxKeySession = "auth.session"
)

const SessionCookieName = "flocal_session"

type SessionValidator interface {
	ValidateSession(ctx context.Context, rawToken string) (*models.User, *models.Session, error)
}

type CookieConfig struct {
	Domain string
	Secure bool
}

var ErrNoCredentials = errors.New("middleware: no session credentials on request")

func RequireAuth(validator SessionValidator, cookies CookieConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, fromCookie, err := extractTokenWithSource(c)
		if err != nil {
			Unauthorized(c, "You must be signed in to access this resource.")
			c.Abort()
			return
		}

		user, session, err := validator.ValidateSession(c.Request.Context(), token)
		if err != nil {
			handleValidationError(c, err, cookies)
			c.Abort()
			return
		}

		if fromCookie && isMutatingMethod(c.Request.Method) && !ValidCSRF(c) {
			Forbidden(c, "Missing or invalid CSRF token.")
			c.Abort()
			return
		}

		c.Set(ctxKeyUser, user)
		c.Set(ctxKeySession, session)
		c.Next()
	}
}

func OptionalAuth(validator SessionValidator, cookies CookieConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _, err := extractTokenWithSource(c)
		if err != nil {
			c.Next()
			return
		}

		user, session, err := validator.ValidateSession(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, service.ErrInvalidSession) || errors.Is(err, service.ErrAccountInactive) {
				ClearSessionCookie(c, cookies.Domain, cookies.Secure)
				ClearCSRFCookie(c, cookies.Domain, cookies.Secure)
			}
			c.Next()
			return
		}

		c.Set(ctxKeyUser, user)
		c.Set(ctxKeySession, session)
		c.Next()
	}
}

func handleValidationError(c *gin.Context, err error, cookies CookieConfig) {
	switch {
	case errors.Is(err, service.ErrAccountInactive):
		ClearSessionCookie(c, cookies.Domain, cookies.Secure)
		ClearCSRFCookie(c, cookies.Domain, cookies.Secure)
		Forbidden(c, "This account is no longer active.")
	case errors.Is(err, service.ErrInvalidSession):
		ClearSessionCookie(c, cookies.Domain, cookies.Secure)
		ClearCSRFCookie(c, cookies.Domain, cookies.Secure)
		Unauthorized(c, "Your session is invalid or has expired. Please sign in again.")
	default:
		Internal(c, "")
	}
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func extractTokenWithSource(c *gin.Context) (string, bool, error) {
	if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != "" {
		return cookie, true, nil
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && parts[1] != "" {
			return parts[1], false, nil
		}
	}

	return "", false, ErrNoCredentials
}

func ExtractToken(c *gin.Context) (string, error) {
	token, _, err := extractTokenWithSource(c)
	return token, err
}

func CurrentUser(c *gin.Context) *models.User {
	v, ok := c.Get(ctxKeyUser)
	if !ok {
		return nil
	}
	u, ok := v.(*models.User)
	if !ok {
		return nil
	}
	return u
}

func CurrentSession(c *gin.Context) *models.Session {
	v, ok := c.Get(ctxKeySession)
	if !ok {
		return nil
	}
	s, ok := v.(*models.Session)
	if !ok {
		return nil
	}
	return s
}

func RequireUser(c *gin.Context) (*models.User, bool) {
	u := CurrentUser(c)
	if u == nil {
		Fail(c, http.StatusUnauthorized, CodeUnauthorized, "Authentication context is missing.")
		c.Abort()
		return nil, false
	}
	return u, true
}
