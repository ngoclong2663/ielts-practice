package app

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"ielts-practice/app/middlewares"
	"ielts-practice/configs"
	"ielts-practice/pkg/tokens"
)

type App struct {
	fxApp *fx.App
}

func NewApp() *App {
	return &App{
		fxApp: BuildContainer(),
	}
}

func (a *App) Run() {
	a.fxApp.Run()
}

func NewGinEngine(cfg *configs.Config, logger zerolog.Logger) (*gin.Engine, error) {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	if err := engine.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
		return nil, err
	}
	registerJSONFieldNames()

	engine.Use(
		gin.Recovery(),
		middlewares.RequestIDMiddleware(),
		middlewares.RequestLoggerMiddleware(logger),
		cors.New(cors.Config{
			AllowOrigins:     []string{cfg.Server.FrontendURL},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"},
			ExposeHeaders:    []string{"X-Request-Id"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		middlewares.ErrorMiddleware(logger),
		middlewares.RateLimitMiddleware(),
	)
	return engine, nil
}

// registerJSONFieldNames makes validation errors report JSON field names (e.g. "display_name").
func registerJSONFieldNames() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "" {
				return f.Name
			}
			return name
		})
	}
}

func NewJWTManager(cfg *configs.Config) *tokens.JWTManager {
	return tokens.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL)
}

func NewHTTPServer(cfg *configs.Config, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func RunHTTPServer(lifecycle fx.Lifecycle, server *http.Server, logger zerolog.Logger) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info().Str("addr", server.Addr).Msg("HTTP server listening")
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Fatal().Err(err).Msg("HTTP server stopped")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
