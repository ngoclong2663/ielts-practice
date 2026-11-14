# Go / Gin Error Handling & DTOs Reference

This guide explains the `AppError` exception structure, error translation middleware, and DTO validation rules in Go.

---

## 🛑 Application Error Structure (`pkg/exceptions/exception.go`)

Custom application errors encapsulate HTTP status code, internal application code, user message, and underlying root cause error:

```go
// pkg/exceptions/exception.go
package exceptions

import "net/http"

var (
	ErrBadRequest = func(message string, err error) *AppError {
		return NewException(http.StatusBadRequest, "BAD_REQUEST", message, err)
	}
	ErrInternalServerError = func(message string, err error) *AppError {
		return NewException(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, err)
	}
	ErrNotFound = func(message string, err error) *AppError {
		return NewException(http.StatusNotFound, "NOT_FOUND", message, err)
	}
	ErrForbidden = func(message string, err error) *AppError {
		return NewException(http.StatusForbidden, "FORBIDDEN", message, err)
	}
)

type AppError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Err     error  `json:"-"`
}

func NewException(status int, code string, message string, err error) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *AppError) Error() string {
	return e.Message
}
```

---

## 🌐 Central Error Middleware (`app/middlewares/errors.middleware.go`)

Intercepts any errors stored in `c.Errors` and converts `*exceptions.AppError` into standardized JSON responses:

```go
// app/middlewares/errors.middleware.go
package middlewares

import (
	"errors"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/longtk26/template-frameworks.git/pkg/exceptions"
)

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *exceptions.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Status, gin.H{"message": appErr.Message})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}
```

---

## 📋 DTO Structs with Validation Tags

Define DTOs with `binding` tags in `app/modules/<module>/dto/`:

```go
// app/modules/auth/dto/sign-up.dto.go
package dto

type SignUpRequestDto struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Username string `json:"username" binding:"required"`
}

type SignUpResponseDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
```

### Validator Error Formatting (`pkg/format-validator-error.go`)

Maps validator tags into user-friendly error messages:

```go
// pkg/format-validator-error.go
package pkg

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				errors[e.Field()] = fmt.Sprintf("%s is required", e.Field())
			case "email":
				errors[e.Field()] = "Invalid email format"
			case "min":
				errors[e.Field()] = fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
			default:
				errors[e.Field()] = fmt.Sprintf("%s is invalid", e.Field())
			}
		}
	}
	return errors
}
```
