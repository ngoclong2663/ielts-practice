# Uber Fx Dependency Injection Reference

This guide details how Uber Fx is configured and structured for dependency injection and application lifecycle management in Go.

---

## 📦 Container Composition (`app/container.go`)

The root container aggregates configuration providers, server factories, database providers, and feature modules:

```go
// app/container.go
package app

import (
	"github.com/longtk26/template-frameworks.git/app/modules/auth"
	"github.com/longtk26/template-frameworks.git/app/modules/database"
	"github.com/longtk26/template-frameworks.git/app/modules/users"
	"github.com/longtk26/template-frameworks.git/configs"
	"go.uber.org/fx"
)

func BuildContainer() *fx.App {
	return fx.New(
		fx.Provide(
			configs.LoadConfig,
			NewGinEngine,
			NewHTTPServer,
			configs.InitLoggerConfig,
		),
		auth.NewAuthModule(),
		users.NewUserModule(),
		database.NewDatabaseModule(),
		fx.Invoke(
			RunHTTPServer,
			configs.InitLoggerConfig,
		),
		fx.NopLogger,
	)
}
```

---

## 🧩 Feature Module Definition (`app/modules/<module>/module.go`)

Each module exports a constructor returning `fx.Option`:

```go
// app/modules/users/module.go
package users

import (
	"github.com/longtk26/template-frameworks.git/app/modules/users/infrastructure/persistence"
	usersInternal "github.com/longtk26/template-frameworks.git/app/modules/users/internal"
	"github.com/longtk26/template-frameworks.git/app/modules/users/internal/mapper"
	usecases "github.com/longtk26/template-frameworks.git/app/modules/users/internal/use-cases"
	usersin "github.com/longtk26/template-frameworks.git/app/modules/users/ports/in"
	"go.uber.org/fx"
)

func NewUserModule() fx.Option {
	return fx.Options(
		fx.Provide(
			usersInternal.NewUserRouter,
			usersInternal.NewUserPresenter,
			usecases.NewUserUseCases,
			// One struct, two views: presenters and other modules (port). No service layer.
			func(uc *usecases.UserUseCases) usecases.IUserUseCases { return uc },
			func(uc *usecases.UserUseCases) usersin.IUserService { return uc },
			persistence.NewUserRepository,
			mapper.NewUserMapper,
		),
		fx.Invoke(
			usersInternal.RegisterRoutes,
		),
	)
}
```

---

## ⏱️ Lifecycle Management (`app/app.go`)

Server lifecycle hooks are attached using `fx.Lifecycle`:

```go
// app/app.go
package app

import (
	"context"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/longtk26/template-frameworks.git/configs"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

func NewHTTPServer(config *configs.Config, engine *gin.Engine, logger zerolog.Logger) *http.Server {
	logger.Info().Msg(fmt.Sprintf("Starting server on port %s...", config.Server.Port))
	return &http.Server{
		Addr:    ":" + config.Server.Port,
		Handler: engine,
	}
}

func RunHTTPServer(lifecycle fx.Lifecycle, server *http.Server) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				_ = server.ListenAndServe()
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
```
