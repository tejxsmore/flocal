package middleware

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const CSRFCookieName = "flocal_csrf"
const CSRFHeaderName = "X-CSRF-Token"

func SetSessionCookie(c *gin.Context, domain string, secure bool, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookieName, token, maxAge, "/", domain, secure, true)
}

func ClearSessionCookie(c *gin.Context, domain string, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookieName, "", -1, "/", domain, secure, true)
}

func SetCSRFCookie(c *gin.Context, domain string, secure bool, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(CSRFCookieName, token, maxAge, "/", domain, secure, false)
}

func ClearCSRFCookie(c *gin.Context, domain string, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(CSRFCookieName, "", -1, "/", domain, secure, false)
}

func ValidCSRF(c *gin.Context) bool {
	cookieVal, err := c.Cookie(CSRFCookieName)
	if err != nil || cookieVal == "" {
		return false
	}
	headerVal := c.GetHeader(CSRFHeaderName)
	if headerVal == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookieVal), []byte(headerVal)) == 1
}
