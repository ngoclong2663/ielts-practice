---
name: gin-skills
description: Comprehensive guide for building Go backend applications using Gin. Covers project bootstrap from template repo branch (feat/go-gin), Clean Layered Architecture, Uber Fx Dependency Injection container, GORM ORM, Zerolog structured logging, generic validation middlewares (ValidateRequestBody, MustGetBody), RateLimitMiddleware, context-propagated database transactions, custom AppError exceptions, and SQL migrations.
---

# Go / Gin Backend Development Skill Guide

This skill guides agents when creating, bootstrapping, refactoring, or reviewing backend applications built with **Go** and **Gin** in this project.

---

## ⚡ Project Initialization & Branch Cloning Guide

To start a new Go / Gin backend project without writing boilerplate code from scratch, **always clone the template branch directly**:

```bash
# 1. Clone the dedicated Go Gin branch from template repo
git clone -b feat/go-gin --single-branch https://github.com/longtk26/template-frameworks.git <project_name>

# 2. Enter project directory & initialize fresh git repository
cd <project_name>
rm -rf .git
git init

# 3. Configure environment variables
cp .env.example .env  # or create .env with DB credentials and PORT

# 4. Download and tidy Go dependencies
go mod tidy

# 5. Apply SQL migrations and start server
make mgup
make run
```

---

## 🛡️ Mandatory Core Capabilities & Verification Requirements

Every Go / Gin backend created or maintained must verify and ensure the following **6 Core Pillars**:

1. **DTO Validation**:
   - Define DTO structs with `binding` tags (e.g. `binding:"required,email"`) in `app/modules/<module>/dto/`.
   - Routes attach `middlewares.ValidateRequestBody[T]()` or `middlewares.ValidateQueryParams[T]()`.
   - Handlers safely extract validated data via `pkg.MustGetBody[T](c)`.
2. **Rate Limiting**:
   - Rate limiting is attached globally or per route group via `middlewares.RateLimitMiddleware()` (or `ShortTermRateLimitMiddleware()` / `MediumTermRateLimitMiddleware()` in `app/middlewares/rate-limit.middleware.go` powered by `golang.org/x/time/rate`).
3. **Transaction Handling**:
   - Atomic multi-step writes run inside `transaction.Execute(ctx, func(txCtx context.Context) error { ... })`.
   - Repositories extract the active transaction connection from context using `database.GetDBFromContext(ctx, r.db)`.
4. **Database Migrations (SQL)**:
   - Raw SQL migration pairs stored in `migrations/<timestamp>_<name>.up.sql` and `.down.sql`.
   - Applied via `./scripts/migrate-up.sh` (or `make mgup`) and rolled back via `./scripts/migrate-down.sh` (or `make mgdown`).
5. **Centralized Error Handling**:
   - Custom `*exceptions.AppError` instances constructed using `exceptions.ErrBadRequest(...)`, `exceptions.ErrNotFound(...)`, `exceptions.ErrForbidden(...)`, and `exceptions.ErrInternalServerError(...)`.
   - Handlers pass errors to `_ = c.Error(err)` and return; `middlewares.ErrorMiddleware()` formats and writes standardized JSON responses.
6. **Structured Logging**:
   - `zerolog.Logger` initialized in `configs.InitLoggerConfig()` and injected via Uber Fx into use cases and repositories for structured, contextual logging.

---

## 🧭 Modular Reference Guides

For deep-dive explanations and code implementations, read the specialized reference documents:

| Topic / Task | Reference Guide | Summary / Focus |
|---|---|---|
| **Architecture & Structure** | [references/architecture-and-structure.md](references/architecture-and-structure.md) | Module directory structure (`domain/`, `infrastructure/`, `internal/`, `ports/`, `dto/`), mappers, separation |
| **Uber Fx Dependency Injection** | [references/uber-fx-dependency-injection.md](references/uber-fx-dependency-injection.md) | `BuildContainer()`, `fx.Provide`, `fx.Invoke`, `fx.Lifecycle` startup/shutdown hooks, module options |
| **Database & Transactions** | [references/database-and-transactions.md](references/database-and-transactions.md) | GORM setup, `TransactionStruct.Execute`, context extraction with `GetDBFromContext(ctx, fallback)` |
| **Routing & Middlewares** | [references/routing-and-middlewares.md](references/routing-and-middlewares.md) | Generic `ValidateRequestBody[T]()`, `MustGetBody[T]()`, `ErrorMiddleware()`, `RateLimitMiddleware()` |
| **Error Handling & DTOs** | [references/error-handling-and-dtos.md](references/error-handling-and-dtos.md) | `AppError` exception constructors, `c.Error(err)` flow, validator formatting, struct tags |
| **Migrations & Scripts** | [references/migrations-and-scripts.md](references/migrations-and-scripts.md) | Raw SQL migrations (`up.sql`, `down.sql`), bash migration scripts, Makefile commands |

---

## ⚡ Core Rules Summary

1. **Clean Layered Module Architecture**:
   - **Domain Entities** (`app/modules/<module>/domain/entity/`): Core business entity structs.
   - **Domain Repository Interface** (`app/modules/<module>/domain/<name>.repository.go`): Interface `I<Name>Repository` defining database contracts with `context.Context`.
   - **Infrastructure Repository** (`app/modules/<module>/infrastructure/<name>.repository.go`): GORM implementation using `database.GetDBFromContext(ctx, r.db)` to participate in active transactions.
   - **Mappers** (`app/modules/<module>/internal/mapper/<name>.mapper.go`, package `mapper`): Explicit mapping functions converting between Entities, GORM Models, and DTOs. Own package so both `infrastructure/` and `use-cases/` can import it without a cycle.
   - **Use Cases** (`app/modules/<module>/internal/use-cases/`, package `usecases`): **one use case per file** (`create-deck.go`, `update-profile.go`, ...), each a method on the module's single `<Module>UseCases` struct. `use-cases.go` holds the struct, constructor, the `I<Module>UseCases` interface used by presenters, and shared helpers/errors. Orchestrates business transactions using `TransactionStruct.Execute` and repositories.
   - **No separate service layer.** Do not add `<module>.service.go`: business rules live only in use cases. When other modules need this module, the same `<Module>UseCases` struct implements the port interface (see below), so a rule is never written twice.
   - **Presenters / Handlers** (`app/modules/<module>/internal/<name>.presenter.go`): Gin HTTP handlers extracting validated DTOs with `pkg.MustGetBody[DTO](c)` and reporting errors via `_ = c.Error(err)`.
   - **Routers** (`app/modules/<module>/internal/<name>.router.go`): Attaches route groups and validation middlewares (`middlewares.ValidateRequestBody[DTO]()`).
   - **Public Ports** (`app/modules/<module>/ports/`): Narrow interfaces exposed to other modules for cross-module calls (e.g. `IUserService` used by `auth`), implemented by the module's `<Module>UseCases` struct and provided in `module.go` with `func(uc *usecases.UserUseCases) ports.IUserService { return uc }`.
2. **Dependency Injection with Uber Fx**:
   - Every module provides a `New<Module>Module() fx.Option` providing all routers, presenters, use cases, repositories, and mappers (plus port bindings), and invoking route registration.
   - Root container in `app/container.go` bundles all module options.
3. **Database Transactions**:
   - Wrap multi-step mutations inside `transaction.Execute(ctx, func(txCtx context.Context) error { ... })`.
   - Repositories extract the active transaction from context via `database.GetDBFromContext(ctx, r.db)`.
4. **Validation & Middlewares**:
   - Routes bind and validate JSON requests using `middlewares.ValidateRequestBody[T]()`.
   - Presenters read validated structs safely via `pkg.MustGetBody[T](c)` or `pkg.GetQuery[T](c)`.
5. **Central Error Handling**:
   - Return `exceptions.ErrNotFound(...)`, `exceptions.ErrBadRequest(...)`, `exceptions.ErrInternalServerError(...)`, or `exceptions.ErrForbidden(...)`.
   - Handlers emit `_ = c.Error(err)` and return; `ErrorMiddleware()` intercepts errors and writes JSON responses.

---

## 📁 Project File Structure

```text
.
├── cmd/
│   └── main.go                      # Entry point (calls app.NewApp().Run())
│
├── configs/
│   ├── config.go                    # Configuration loader (Viper / env)
│   ├── types.go                     # Config structs
│   └── logger.go                    # Zerolog initialization
│
├── app/
│   ├── app.go                       # Gin engine, HTTP server, fx lifecycle hooks
│   ├── container.go                 # Uber Fx container assembly (BuildContainer)
│   ├── middlewares/
│   │   ├── errors.middleware.go     # Central AppError handler
│   │   ├── rate-limit.middleware.go # Request rate limiter (rate.NewLimiter)
│   │   └── validate-input.middleware.go # Generic ValidateRequestBody[T]
│   ├── models/
│   │   └── user.model.go            # Shared GORM table models
│   │
│   └── modules/
│       ├── database/
│       │   ├── module.go            # GORM DB provider
│       │   └── transaction.go       # TransactionStruct & GetDBFromContext
│       │
│       ├── auth/                    # Auth module
│       │   ├── dto/
│       │   ├── internal/
│       │   └── module.go
│       │
│       └── <module>/                # Feature module (e.g., users)
│           ├── module.go            # Fx Module definition (Provides & Invokes)
│           ├── dto/                 # Request & Response DTOs
│           ├── domain/
│           │   ├── entity/          # Pure entity structs
│           │   └── <module>.repository.go # I<Module>Repository interface
│           ├── infrastructure/
│           │   └── <module>.repository.go # GORM implementation
│           ├── ports/               # Public cross-module interfaces (implemented by the use cases)
│           │   └── <module>-service.port.go
│           └── internal/            # Module-private components
│               ├── mapper/          # package mapper
│               │   └── <module>.mapper.go
│               ├── use-cases/       # package usecases: one use case per file
│               │   ├── use-cases.go # <Module>UseCases struct, constructor, interface, shared helpers
│               │   ├── create-<thing>.go
│               │   └── update-<thing>.go
│               ├── <module>.presenter.go
│               └── <module>.router.go
│
├── pkg/
│   ├── exceptions/
│   │   └── exception.go             # AppError & standard error constructors
│   ├── passwords/
│   │   └── passwords.go             # Bcrypt hashing helpers
│   ├── format-validator-error.go    # Validator error formatter
│   └── handle-dto.go                # MustGetBody[T] & GetQuery[T]
│
├── migrations/                      # Raw SQL migrations (.up.sql, .down.sql)
├── scripts/                         # Migration runner scripts
└── Makefile                         # Build and run commands
```

---

## 🛠️ Step-by-Step: Adding a New Feature Module

When creating a new feature module (e.g. `products`):

1. **Create SQL Migration**:
   - Create `migrations/<timestamp>_init_products_table.up.sql` and `.down.sql`.
   - Run migration script: `./scripts/migrate-up.sh` (or `make mgup`).
2. **Define GORM Model**:
   - Create `app/models/product.model.go` with GORM and JSON tags.
3. **Define Domain Entity & Repository Interface**:
   - Create `app/modules/products/domain/entity/product.entity.go`.
   - Create `app/modules/products/domain/product.repository.go` with `IProductRepository` interface.
4. **Implement Infrastructure Repository**:
   - Create `app/modules/products/infrastructure/product.repository.go`.
   - Extract DB using `database.GetDBFromContext(ctx, r.db)` and map entities via mapper.
5. **Create Mapper & DTOs**:
   - Create `app/modules/products/dto/create-product.dto.go` with `binding:"required"` tags.
   - Create `app/modules/products/internal/mapper/product.mapper.go` (package `mapper`).
6. **Implement Use Cases** (one file per use case):
   - Create `app/modules/products/internal/use-cases/use-cases.go` with the `ProductUseCases` struct (repository, `TransactionStruct`, mapper, logger), `NewProductUseCases`, the `IProductUseCases` interface and shared helpers.
   - Add one file per use case, e.g. `use-cases/create-product.go` containing only `func (uc *ProductUseCases) CreateProduct(...)`.
   - Never add a `product.service.go`; if other modules need products, define a port in `ports/` and let `ProductUseCases` implement it.
7. **Create Presenter & Router**:
   - Create `app/modules/products/internal/product.presenter.go` using `pkg.MustGetBody[DTO](c)`.
   - Create `app/modules/products/internal/product.router.go` attaching `middlewares.ValidateRequestBody[DTO]()`.
8. **Assemble Fx Module**:
   - Create `app/modules/products/module.go`:
     ```go
     func NewProductModule() fx.Option {
         return fx.Options(
             fx.Provide(
                 internal.NewProductRouter,
                 internal.NewProductPresenter,
                 usecases.NewProductUseCases,
                 infrastructure.NewProductRepository,
                 mapper.NewProductMapper,
             ),
             fx.Invoke(internal.RegisterRoutes),
         )
     }
     ```
   - Register `products.NewProductModule()` in `app/container.go`.

---

## ✅ PR & Review Checklist (6 Pillars Verified)

Before finishing tasks or opening a PR, verify:

- [ ] **Bootstrap**: Project bootstrapped from `feat/go-gin` template branch.
- [ ] **1. DTO Validation**: Routes apply `middlewares.ValidateRequestBody[DTO]()` or `ValidateQueryParams[DTO]()`; handlers use `pkg.MustGetBody[DTO](c)`.
- [ ] **2. Rate Limiting**: `middlewares.RateLimitMiddleware()` active on router engine or specific endpoint groups.
- [ ] **3. Transactions**: Multi-step operations run inside `transaction.Execute(...)`; repositories call `database.GetDBFromContext(ctx, r.db)`.
- [ ] **4. Migrations**: Migration pairs (`.up.sql`, `.down.sql`) created and applied via `./scripts/migrate-up.sh` (or `make mgup`).
- [ ] **5. Error Handling**: Errors constructed using `exceptions.Err*` functions and passed to `c.Error(err)` for `ErrorMiddleware()`.
- [ ] **6. Logger**: `zerolog.Logger` injected into use cases and repositories for structured logging.
- [ ] **Uber Fx Integration**: `New<Module>Module()` provided in `app/container.go`.
- [ ] **Repository Interfaces**: Interfaces live in `domain/` and accept `context.Context`.
