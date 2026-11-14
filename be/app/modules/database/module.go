package database

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"ielts-practice/configs"
)

func buildDSN(cfg *configs.Config) string {
	if cfg.Database.URL != "" {
		return cfg.Database.URL
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)
}

func InitDatabaseConnection(lc fx.Lifecycle, cfg *configs.Config, logger zerolog.Logger) (*gorm.DB, error) {
	logLevel := gormLogger.Warn
	if cfg.IsProduction() {
		logLevel = gormLogger.Error
	}

	db, err := gorm.Open(postgres.Open(buildDSN(cfg)), &gorm.Config{
		Logger:         gormLogger.Default.LogMode(logLevel),
		TranslateError: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := sqlDB.PingContext(ctx); err != nil {
				return fmt.Errorf("ping database: %w", err)
			}
			logger.Info().Msg("Connected to the database")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})
	return db, nil
}

func NewDatabaseModule() fx.Option {
	return fx.Options(
		fx.Provide(InitDatabaseConnection),
		fx.Provide(NewTransaction),
	)
}
