# Go / Gin Architecture & Layering Reference

This guide details the Clean / Layered module architecture and boundary separation used in Go / Gin backend applications.

---

## 🏗️ Layer Interaction Diagram

```text
HTTP Request (Client)
       │
       ▼
┌────────────────────────────────────────────────────────┐
│ Router (`internal/<module>.router.go`)                 │
│ - Defines route endpoints                              │
│ - Attaches ValidateRequestBody[DTO]() middleware       │
└───────────────────────────┬────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────┐
│ Presenter (`internal/<module>.presenter.go`)           │
│ - Reads validated payload via pkg.MustGetBody[DTO](c)  │
│ - Calls Use Case                                       │
│ - Emits errors with c.Error(err)                       │
│ - Returns JSON response on success                     │
└───────────────────────────┬────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────┐
│ Use Cases (`internal/use-cases/<use-case>.go`)         │
│ - Executes business operations                         │
│ - Wraps mutations in Transaction.Execute(...)          │
│ - Uses Mappers to convert DTOs <-> Entities            │
│ - Calls Repository interfaces (domain.IUserRepository) │
└──────────────┬──────────────────────────┬──────────────┘
               │                          │
               ▼                          ▼
┌───────────────────────────┐ ┌───────────────────────────┐
│ Domain Entities           │ │ Domain Repositories       │
│ (`domain/entity/`)        │ │ (`domain/<name>.repo.go`) │
│ - Pure business structs   │ │ - Context-aware interface │
└───────────────────────────┘ └─────────────┬─────────────┘
                                            │
                                            ▼
┌────────────────────────────────────────────────────────┐
│ Infrastructure (`infrastructure/<name>.repository.go`) │
│ - Implements IRepository interface                     │
│ - Resolves GORM DB via database.GetDBFromContext(ctx)  │
│ - Maps between GORM Models (app/models) and Entities   │
└────────────────────────────────────────────────────────┘
```

---

## 📦 Layer Implementations

### 1. Domain Entity (`app/modules/<module>/domain/entity/`)
Pure domain struct without database tags.

```go
// app/modules/users/domain/entity/user.entity.go
package entity

type UserEntity struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
	IsActive bool   `json:"is_active"`
}
```

### 2. GORM Table Model (`app/models/`)
Database persistence model with GORM tags.

```go
// app/models/user.model.go
package models

import "time"

type User struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username  string    `gorm:"type:varchar(255);not null" json:"username"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	IsActive  bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

### 3. Domain Repository Interface (`app/modules/<module>/domain/`)

```go
// app/modules/users/domain/user.repository.go
package domain

import (
	"context"
	"github.com/longtk26/template-frameworks.git/app/modules/users/domain/entity"
)

type IUserRepository interface {
	CreateUser(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error)
	GetUserByID(ctx context.Context, id string) (*entity.UserEntity, error)
	FindByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
}
```

### 4. Infrastructure Repository (`app/modules/<module>/infrastructure/`)

```go
// app/modules/users/infrastructure/user.repository.go
package infrastructure

import (
	"context"
	"github.com/longtk26/template-frameworks.git/app/models"
	"github.com/longtk26/template-frameworks.git/app/modules/database"
	"github.com/longtk26/template-frameworks.git/app/modules/users/domain"
	"github.com/longtk26/template-frameworks.git/app/modules/users/domain/entity"
	"github.com/longtk26/template-frameworks.git/app/modules/users/internal"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var _ domain.IUserRepository = &UserRepository{}

type UserRepository struct {
	db         *gorm.DB
	userMapper *internal.UserMapper
	logger     zerolog.Logger
}

func NewUserRepository(db *gorm.DB, userMapper *internal.UserMapper, logger zerolog.Logger) domain.IUserRepository {
	return &UserRepository{db: db, userMapper: userMapper, logger: logger}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error) {
	userModel := r.userMapper.MapUserEntityToModel(user)
	db := database.GetDBFromContext(ctx, r.db)

	if err := db.Create(&userModel).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*entity.UserEntity, error) {
	var user models.User
	db := database.GetDBFromContext(ctx, r.db)

	if err := db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return r.userMapper.MapUserModelToEntity(&user), nil
}
```

### 5. Mapper (`app/modules/<module>/internal/mapper/`, package `mapper`)

```go
// app/modules/users/internal/mapper/user.mapper.go
package mapper

import (
	"github.com/longtk26/template-frameworks.git/app/models"
	"github.com/longtk26/template-frameworks.git/app/modules/users/domain/entity"
)

type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

func (m *UserMapper) MapUserEntityToModel(e *entity.UserEntity) models.User {
	return models.User{
		ID:       e.ID,
		Username: e.Username,
		Email:    e.Email,
		Password: e.Password,
		IsActive: e.IsActive,
	}
}

func (m *UserMapper) MapUserModelToEntity(model *models.User) *entity.UserEntity {
	return &entity.UserEntity{
		ID:       model.ID,
		Username: model.Username,
		Email:    model.Email,
		IsActive: model.IsActive,
	}
}
```

### 6. Use Cases (`app/modules/<module>/internal/use-cases/`, package `usecases`)
One use case per file, all methods on the module's single `<Module>UseCases` struct:

```text
internal/use-cases/
├── use-cases.go           # UserUseCases struct, NewUserUseCases, IUserUseCases, shared helpers
├── get-me.go              # func (u *UserUseCases) GetMe(...)
├── update-profile.go      # func (u *UserUseCases) UpdateProfile(...)
└── upsert-google-user.go  # func (u *UserUseCases) UpsertGoogleUser(...)  (also part of the port)
```

There is **no service layer** (`<module>.service.go`). Business rules exist in exactly one place.

### 7. Cross-Module Public Ports (`app/modules/<module>/ports/`)
When other modules (e.g. `auth`) need to invoke user functionality, define a narrow public port in `ports/`. The module's `<Module>UseCases` struct implements it, and `module.go` binds it: `func(uc *usecases.UserUseCases) ports.IUserService { return uc }`.

```go
// app/modules/users/ports/user-service.port.go
package ports

import (
	"context"
	"github.com/longtk26/template-frameworks.git/app/modules/users/domain/entity"
)

type IUserService interface {
	CreateUser(ctx context.Context, userEntity *entity.UserEntity) (*entity.UserEntity, error)
}
```
