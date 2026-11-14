package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"ielts-practice/app/modules/storage/ports"
)

type HealthPresenter struct {
	db      *gorm.DB
	storage ports.IObjectStorage
}

func NewHealthPresenter(db *gorm.DB, storage ports.IObjectStorage) *HealthPresenter {
	return &HealthPresenter{db: db, storage: storage}
}

// Check: GET /api/v1/health. Used by Railway health checks.
// storage is "not_configured" (not an error) until media storage is set up.
func (p *HealthPresenter) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "ok"
	if sqlDB, err := p.db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
		dbStatus = "down"
	}

	storageStatus := "not_configured"
	if p.storage.Configured() {
		storageStatus = "ok"
		if err := p.storage.Ping(ctx); err != nil {
			storageStatus = "down"
		}
	}

	healthy := dbStatus == "ok" && storageStatus != "down"
	status, label := http.StatusOK, "ok"
	if !healthy {
		status, label = http.StatusServiceUnavailable, "degraded"
	}
	c.JSON(status, gin.H{"status": label, "db": dbStatus, "storage": storageStatus})
}

func RegisterRoutes(router *gin.Engine, p *HealthPresenter) {
	router.GET("/api/v1/health", p.Check)
}

func NewHealthModule() fx.Option {
	return fx.Options(
		fx.Provide(NewHealthPresenter),
		fx.Invoke(RegisterRoutes),
	)
}
