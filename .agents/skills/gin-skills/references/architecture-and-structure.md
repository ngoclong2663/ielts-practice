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
│ Infrastructure (`infrastructure/persistence/...`)      │
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

### 4. Infrastructure Repository (`app/modules/<module>/infrastructure/persistence/`)

Each adapter technology has its own sub-package under `infrastructure/`: `persistence/` (GORM repositories),
and one folder per external service (`google/`, `payment/`, `mailer/`, `s3/`, …). Never put files directly in
`infrastructure/`.

```go
// app/modules/users/infrastructure/persistence/user.repository.go
package persistence

import (
	"context"
	"github.com/longtk26/template-frameworks.git/app/models"
	"github.com/longtk26/template-frameworks.git/app/modules/database"
	"github.com/longtk26/template-frameworks.git/app/modules/users/domain"
	"github.com/longtk26/template-frameworks.git/app/modules/users/domain/entity"
	"github.com/longtk26/template-frameworks.git/app/modules/users/internal/mapper"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var _ domain.IUserRepository = &UserRepository{}

type UserRepository struct {
	db         *gorm.DB
	userMapper *mapper.UserMapper
	logger     zerolog.Logger
}

func NewUserRepository(db *gorm.DB, userMapper *mapper.UserMapper, logger zerolog.Logger) domain.IUserRepository {
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

### 7. Ports: `ports/in` and `ports/out`

A module's interfaces to the outside world are split by direction. Repository interfaces stay in `domain/`.

| Folder | Package | Contains | Implemented by |
|---|---|---|---|
| `ports/in/` | `in` | What this module **offers** to other modules (e.g. `IUserService`, `IOrderService`) | the module's `<Module>UseCases` (or the adapter, for shared infrastructure modules, §8) |
| `ports/out/` | `out` | External services this module **needs** (e.g. `IGoogleIdentityProvider`, `IPaymentGateway`, `IMailer`), plus the errors that are part of that contract | an adapter in `infrastructure/<adapter>/` |

Data structs exchanged through a port (`ChargeRequest`, `EmailMessage`, …) are value objects and live in `domain/entity/`;
the port files hold only interfaces and contract errors.

Import ports with a module-prefixed alias, so the direction is visible and `in`/`out` never clash with variables:

```go
import (
	usersin "github.com/longtk26/template-frameworks.git/app/modules/users/ports/in" // offered by users
	authout "github.com/longtk26/template-frameworks.git/app/modules/auth/ports/out" // needed by auth
)
```

`module.go` binds an in-port to the use cases: `func(uc *usecases.UserUseCases) usersin.IUserService { return uc }`,
and provides out-port adapters directly: `google.NewGoogleIdentityProvider` returns `authout.IGoogleIdentityProvider`.

```go
// app/modules/users/ports/in/user-service.port.go
package in

type IUserService interface {
	CreateUser(ctx context.Context, userEntity *entity.UserEntity) (*entity.UserEntity, error)
}
```

### 8. Shared infrastructure modules (`database`, `storage`, `llm`)

Technical capabilities used by several feature modules are modules of their own, with no routes and no use cases,
e.g. `storage` (S3: `storage/ports/in.IObjectStorage`) or `llm` (language-model clients: `llm/ports/in.IJSONModel`).
Their in-port is implemented by the adapter itself. Feature modules depend on those in-ports, never on another
module's infrastructure. Feature-specific details (prompts, schemas, file layouts) stay in the feature module's own
adapter (`<module>/infrastructure/<adapter>/`), which calls the shared in-port.

### 9. The `domain/` folder

```text
domain/
├── entity/                 # entities AND value objects (package entity)
├── rules/                  # domain services: pure business rules (package rules)
└── <module>.repository.go  # repository interfaces (+ their query shapes); nothing else at the root
```

- **`entity/`**: plain structs, no database tags.
  - **Entities** have an identity (an ID): `User`, `Order`, `Product`.
  - **Value objects** are defined only by their values: `Money`, `Address`, `DateRange`, `OrderLine`.
  - They live together; each file starts with a two-line comment naming which types are entities and which are value objects.
  - Enumerations of entity fields (`OrderStatusPaid`) and the data sent through `ports/out` (`ChargeRequest`) are value objects too.
- **`rules/`**: domain services, i.e. business rules with no I/O.
  - Examples: price and discount calculation, a scheduling algorithm, input normalisation, eligibility checks, scoring.
  - Stateless functions, or small structs built once from config (e.g. `rules.NewScheduler(cfg)`).
  - Unit-tested with plain inputs (`rules/*_test.go`), no database and no mocks.
  - Not an application "service layer": that role belongs to the use cases. Never name these files `*.service.go`.
- **Root**: only `<module>.repository.go`.

**Use case or rule?**
- A **use case** (`internal/use-cases/`) is the flow of one user action: load, check ownership and limits, call ports, save in a transaction, return the response or error.
- A **rule** (`domain/rules/`) answers what is true or allowed given some data.

Put code in `rules/` when:
- it needs no database, API, config loader or logger;
- a domain expert would recognise it as a rule;
- it would be the same in a CLI or another front end;
- more than one use case needs it.

Warning signs:
- a `domain/` file importing `gorm`, `net/http`, `configs` or a logger;
- a use case containing a long calculation (a rule hiding in the flow);
- the same `if` written in two use cases.

If another module needs the same rule, extract it into `pkg/<name>` as a pure library. Do not make it a module.

**Not in `domain/`:**
- interfaces to external services (`ports/out`);
- HTTP or SDK code (`infrastructure/<adapter>/`).
