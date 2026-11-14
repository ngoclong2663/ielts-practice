# Go / Gin Routing & Middlewares Reference

This guide covers generic validation middlewares, context retrieval helpers, and global error handling in Gin.

---

## 🛡️ Generic Validation Middleware (`app/middlewares/validate-input.middleware.go`)

Uses Go Generics to bind and validate incoming request payloads before they reach the presenter:

```go
// app/middlewares/validate-input.middleware.go
package middlewares

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/longtk26/template-frameworks.git/pkg"
)

func ValidateRequestBody[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var obj T
		if err := c.ShouldBindJSON(&obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errors": pkg.FormatValidationErrors(err)})
			c.Abort()
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
			c.JSON(http.StatusBadRequest, gin.H{"errors": pkg.FormatValidationErrors(err)})
			c.Abort()
			return
		}
		c.Set("query", obj)
		c.Next()
	}
}
```

---

## 📥 Context Retrieval Helpers (`pkg/handle-dto.go`)

Safely extracts the validated struct from `gin.Context`:

```go
// pkg/handle-dto.go
package pkg

import "github.com/gin-gonic/gin"

func MustGetBody[T any](c *gin.Context) T {
	value, exists := c.Get("body")
	if !exists {
		panic("body not found in context")
	}

	dto, ok := value.(T)
	if !ok {
		panic("invalid body type")
	}

	return dto
}

func GetQuery[T any](c *gin.Context) (T, bool) {
	value, exists := c.Get("query")
	if !exists {
		var zero T
		return zero, false
	}

	dto, ok := value.(T)
	if !ok {
		var zero T
		return zero, false
	}

	return dto, true
}
```

---

## 🚦 Attaching Routes and Presenters

### 1. In Router (`internal/<module>.router.go`)

```go
// app/modules/auth/internal/auth.route.go
package internal

import (
	"github.com/gin-gonic/gin"
	"github.com/longtk26/template-frameworks.git/app/middlewares"
	"github.com/longtk26/template-frameworks.git/app/modules/auth/dto"
)

type AuthRouter struct {
	presenter *AuthPresenter
}

func RegisterRoutes(router *gin.Engine, ar *AuthRouter) {
	ar.Register(router)
}

func NewAuthRouter(router *gin.Engine, presenter *AuthPresenter) *AuthRouter {
	return &AuthRouter{presenter: presenter}
}

func (ar *AuthRouter) Register(router *gin.Engine) {
	authGroup := router.Group("api/v1/auth")
	{
		authGroup.POST(
			"/signup",
			middlewares.ValidateRequestBody[dto.SignUpRequestDto](),
			ar.presenter.SignUp,
		)
	}
}
```

### 2. In Presenter (`internal/<module>.presenter.go`)

```go
// app/modules/auth/internal/auth.presenter.go
package internal

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/longtk26/template-frameworks.git/app/modules/auth/dto"
	"github.com/longtk26/template-frameworks.git/pkg"
)

type AuthPresenter struct {
	us *AuthUseCase
}

func NewAuthPresenter(us *AuthUseCase) *AuthPresenter {
	return &AuthPresenter{us: us}
}

func (p *AuthPresenter) SignUp(c *gin.Context) {
	req := pkg.MustGetBody[dto.SignUpRequestDto](c)

	response, err := p.us.SignUp(req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}
```
