package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SessionName is the signed/encrypted session cookie name shared by all
// handlers and tests.
const SessionName = "oauth_session"

// RefreshTokenCookieName is the HttpOnly cookie name used by the web refresh
// flow.
const RefreshTokenCookieName = "refresh_token"

// RefreshTokenCookiePath scopes the web refresh cookie to the refresh endpoint
// so it is not sent with unrelated API requests.
const RefreshTokenCookiePath = "/api/oauth/refresh"

// SetRefreshTokenCookie writes the web refresh token as an HttpOnly SameSite=Lax
// cookie with the configured TTL and secure flag.
func SetRefreshTokenCookie(c *gin.Context, value string, ttl time.Duration, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		RefreshTokenCookieName,
		value,
		int(ttl.Seconds()),
		RefreshTokenCookiePath,
		"",
		secure,
		true,
	)
}

// ClearRefreshTokenCookie expires the web refresh-token cookie using the same
// name and path used when setting it.
func ClearRefreshTokenCookie(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		RefreshTokenCookieName,
		"",
		-1,
		RefreshTokenCookiePath,
		"",
		secure,
		true,
	)
}
