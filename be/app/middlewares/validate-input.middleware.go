package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ielts-practice/pkg"
	"ielts-practice/pkg/exceptions"
)

func validationError(err error) *exceptions.AppError {
	appErr := exceptions.NewException(http.StatusBadRequest, exceptions.CodeValidationFailed, "Validation failed", err)
	appErr.Fields = pkg.FormatValidationErrors(err)
	return appErr
}

func ValidateRequestBody[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var obj T
		if err := c.ShouldBindJSON(&obj); err != nil {
			WriteError(c, validationError(err))
			return
		}
		c.Set("body", obj)
		c.Next()
	}
}

func ValidateQueryParams[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var obj T
		if err := c.ShouldBindQuery(&obj); err != nil {
			WriteError(c, validationError(err))
			return
		}
		c.Set("query", obj)
		c.Next()
	}
}

// ValidateUUIDParams rejects malformed IDs with 404 before they reach the database.
func ValidateUUIDParams(names ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, name := range names {
			if _, err := uuid.Parse(c.Param(name)); err != nil {
				WriteError(c, exceptions.ErrNotFound("Not found", err))
				return
			}
		}
		c.Next()
	}
}
