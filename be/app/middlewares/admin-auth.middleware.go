package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"ielts-practice/pkg/exceptions"
)

const (
	// AdminTokenCookie is set by the Next.js server action after an admin signs in.
	AdminTokenCookie = "admin_token"
	adminIDKey       = "admin_id"
)

// AdminVerifier checks an admin token and that its account still exists (admin/ports/in.IAdminSession).
// ok is false for an invalid or expired token or a deleted account; err is for failures (database).
type AdminVerifier interface {
	VerifyAdminToken(ctx context.Context, raw string) (adminID string, ok bool, err error)
}

// AdminAuthMiddleware accepts the admin token from `Authorization: Bearer` or the `admin_token`
// cookie. Anything wrong → 401 ADMIN_UNAUTHORIZED. Learner tokens never pass (different secret).
func AdminAuthMiddleware(v AdminVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := ""
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			raw = strings.TrimPrefix(h, "Bearer ")
		} else if cookie, err := c.Cookie(AdminTokenCookie); err == nil {
			raw = cookie
		}
		unauthorized := exceptions.NewException(http.StatusUnauthorized, exceptions.CodeAdminUnauthorized, "Admin sign-in required", nil)
		if raw == "" {
			WriteError(c, unauthorized)
			return
		}
		id, ok, err := v.VerifyAdminToken(c.Request.Context(), raw)
		if err != nil {
			_ = c.Error(exceptions.ErrInternalServerError("Failed to verify admin session", err))
			c.Abort()
			return
		}
		if !ok {
			WriteError(c, unauthorized)
			return
		}
		c.Set(adminIDKey, id)
		c.Next()
	}
}

// AdminGuard is the chain for every /api/v1/admin/* route except login: the per-IP short-term
// rate limit, then the admin token.
func AdminGuard(v AdminVerifier) []gin.HandlerFunc {
	return []gin.HandlerFunc{ShortTermRateLimitMiddleware(), AdminAuthMiddleware(v)}
}

// MustGetAdminID returns the signed-in admin's ID. Only call it behind AdminAuthMiddleware.
func MustGetAdminID(c *gin.Context) string {
	id := c.GetString(adminIDKey)
	if id == "" {
		panic("admin_id not found in context: route is missing AdminAuthMiddleware")
	}
	return id
}
