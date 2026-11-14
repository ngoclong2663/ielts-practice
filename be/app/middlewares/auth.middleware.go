package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"ielts-practice/pkg/exceptions"
	"ielts-practice/pkg/tokens"
)

const (
	AccessTokenCookie = "access_token"
	userIDKey         = "user_id"
)

// AuthMiddleware accepts the access token from the `access_token` cookie (browser requests
// proxied by Next.js) or from `Authorization: Bearer` (server-side calls from Next.js).
func AuthMiddleware(jwt *tokens.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := ""
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			raw = strings.TrimPrefix(h, "Bearer ")
		} else if cookie, err := c.Cookie(AccessTokenCookie); err == nil {
			raw = cookie
		}

		if raw == "" {
			WriteError(c, exceptions.ErrUnauthorized("Authentication required", nil))
			return
		}

		userID, err := jwt.Verify(raw)
		if err != nil {
			if errors.Is(err, tokens.ErrTokenExpired) {
				WriteError(c, exceptions.NewException(http.StatusUnauthorized, exceptions.CodeTokenExpired, "Access token expired", err))
				return
			}
			WriteError(c, exceptions.ErrUnauthorized("Invalid access token", err))
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

// MustGetUserID returns the authenticated user's ID. Only call it behind AuthMiddleware.
func MustGetUserID(c *gin.Context) string {
	id := c.GetString(userIDKey)
	if id == "" {
		panic("user_id not found in context: route is missing AuthMiddleware")
	}
	return id
}
