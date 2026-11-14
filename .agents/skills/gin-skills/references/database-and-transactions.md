# Go / Gin Database & Transactions Reference

This guide covers GORM database initialization, context-propagated database transactions, and transaction execution helpers.

---

## 🗄️ Database Module Provider (`app/modules/database/module.go`)

```go
// app/modules/database/module.go
package database

import (
	"fmt"
	"github.com/longtk26/template-frameworks.git/configs"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabaseModule() fx.Option {
	return fx.Options(
		fx.Provide(
			NewGormDB,
			NewTransaction,
		),
	)
}

func NewGormDB(cfg *configs.Config, logger zerolog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
```

---

## 🔄 Context-Based Transaction Helper (`app/modules/database/transaction.go`)

`TransactionStruct` provides a clean way to execute a block of operations within an atomic transaction. It injects `*gorm.DB` (with `Begin()`) into `context.Context`.

```go
// app/modules/database/transaction.go
package database

import (
	"context"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type txContextKey struct{}

type TransactionStruct struct {
	db     *gorm.DB
	logger zerolog.Logger
}

func NewTransaction(db *gorm.DB, logger zerolog.Logger) *TransactionStruct {
	return &TransactionStruct{db: db, logger: logger}
}

func (ts *TransactionStruct) Execute(ctx context.Context, fc func(ctx context.Context) error) error {
	ts.logger.Info().Msg("Starting transaction")
	tx := ts.db.Begin()
	if err := tx.Error; err != nil {
		ts.logger.Error().Err(err).Msg("Failed to begin transaction")
		return err
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	defer func() {
		if r := recover(); r != nil {
			ts.logger.Error().Interface("panic", r).Msg("Panic occurred during transaction, rolling back")
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fc(txCtx); err != nil {
		tx.Rollback()
		ts.logger.Error().Err(err).Msg("Failed to execute transaction, rolling back")
		return err
	}

	ts.logger.Info().Msg("Committing transaction")
	return tx.Commit().Error
}

func GetDBFromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if ctx == nil {
		return fallback
	}

	tx, ok := ctx.Value(txContextKey{}).(*gorm.DB)
	if ok && tx != nil {
		return tx
	}

	return fallback
}
```

---

## 🎯 Usage in Use Cases & Repositories

### 1. In Repositories: Always call `GetDBFromContext(ctx, r.db)`

```go
func (r *UserRepository) CreateUser(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error) {
	userModel := r.userMapper.MapUserEntityToModel(user)
	db := database.GetDBFromContext(ctx, r.db)

	if err := db.Create(&userModel).Error; err != nil {
		return nil, err
	}
	return user, nil
}
```

### 2. In Use Cases: Wrap multi-step mutations with `Execute`

```go
func (uc *AuthUseCase) SignUp(payload dto.SignUpRequestDto) (dto.SignUpResponseDto, error) {
	// ... hash password and map entity ...
	err := uc.transaction.Execute(context.Background(), func(ctx context.Context) error {
		_, err := uc.userService.CreateUser(ctx, userEntity)
		if err != nil {
			return err
		}
		// any subsequent repository writes in the same ctx join this transaction!
		return nil
	})
	return responseDto, err
}
```
