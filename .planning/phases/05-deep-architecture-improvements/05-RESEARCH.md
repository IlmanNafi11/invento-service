# Phase 5: Deep Architecture Improvements - Research

**Researched:** 2026-02-16
**Domain:** Go architecture patterns -- context propagation, DTO separation, route modularization, error-to-HTTP middleware, file size enforcement, test parallelization
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Route Organization:**
- Single routes file with grouped domain-specific register functions (e.g., `registerUserRoutes(api fiber.Router)`)
- Each register function receives the router group and applies its own middleware (auth, RBAC) per-group -- no global middleware application
- Route file location (stay in server.go vs separate routes.go) is at Claude's discretion based on server.go size after refactoring

**DTO Separation:**
- Dedicated `internal/dto/` package for all request/response structs, separate from `internal/domain/`
- Naming convention: Action + Entity + Type (e.g., `CreateUserRequest`, `CreateUserResponse`, `UpdateModulRequest`)
- Mapping between domain models and DTOs uses a mapping library (e.g., `jinzhu/copier`) -- user confirmed this is an intentional exception to the project's general "no external tools" philosophy
- Validation tag placement (on DTOs only vs both layers) at Claude's discretion

**File Size & Test Splitting:**
- Hard 500-line limit per Go source file -- no exceptions
- Files exceeding 500 lines are split by interface layer (e.g., separate repository, usecase, controller sub-files)
- Test file split strategy at Claude's discretion (mirror source or by test type)
- Aggressive test parallelization with `t.Parallel()` -- all tests that don't share mutable state should be parallelized
- Note for researcher: SQLite in tests may need DB isolation strategy (separate DB files per test or transaction rollback) to support aggressive parallelization

**Specific Ideas:**
- Register functions pattern: `registerUserRoutes(api fiber.Router)` -- standalone functions, not controller methods
- DTO naming must be fully spelled out: `CreateUserRequest`, not `UserCreateReq`
- Mapping library preference noted despite project convention of hand-written approaches -- user values reduced boilerplate here

### Claude's Discretion
- Context.Context rollout order across domains (user, modul, project) -- Claude decides based on dependency analysis
- Route file location (server.go vs routes.go)
- Validation tag placement on DTOs vs domain models
- Test file splitting approach (mirror source files vs by test type)
- Specific mapping library choice (copier or alternative)
- Centralized error-to-HTTP mapping middleware design

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

## Summary

Phase 5 is the highest blast-radius phase in the roadmap. It touches every interface signature (adding `context.Context`), restructures how domain models interact with controllers (DTO separation), reorganizes route registration (modular routes), adds centralized error-to-HTTP mapping middleware, enforces a hard 500-line limit on all source files, and parallelizes the entire test suite.

The codebase currently has 9 usecase interfaces (totaling ~66 methods), 8 repository interfaces (totaling ~70 methods), and 0 existing `context.Context` usage in any usecase or repository interface. The `domain.AuthService` interface already uses `context.Context` (6 of 7 methods), proving the pattern works in this codebase. The existing `internal/dto/` package already has `common.go` and `request.go` with pagination and error detail types, providing a foundation to build on.

The test suite has approximately 1349 test functions across 84 test files, with zero uses of `t.Parallel()` anywhere. The test infrastructure uses in-memory SQLite (`:memory:`) with GORM, which creates a natural isolation boundary since each `SetupTestDatabase()` call creates a separate in-memory database. The largest source files are `test_mocks.go` (603 lines), `tus_modul_usecase.go` (462 lines), and `tus_upload_usecase.go` (430 lines) -- all currently under 500 lines. However, several test files exceed 500 lines significantly (e.g., `tus_helper_test.go` at 2009 lines, `user_controller_test.go` at 1323 lines).

**Primary recommendation:** Roll out context propagation domain-by-domain starting with `user` (fewest cross-domain dependencies), extract route registration into a dedicated `routes.go` file (server.go is 330 lines, so extraction keeps both files well under 500), move all request/response structs from `domain/` to `dto/` using `jinzhu/copier` for mapping, add centralized error-to-HTTP mapping as Fiber middleware, and split all test files exceeding 500 lines.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/gofiber/fiber/v2` | v2.52.11 | HTTP framework | Already in use; `c.UserContext()` provides `context.Context` bridge |
| `gorm.io/gorm` | v1.31.0 | ORM with context support | Already in use; `db.WithContext(ctx)` is the standard pattern |
| `github.com/jinzhu/copier` | latest | Struct-to-struct mapping (domain <-> DTO) | User-approved exception to "no external tools" philosophy; same author as GORM |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/go-playground/validator/v10` | v10.27.0 | Struct validation tags | Already in use; validation tags go on DTOs only |
| `github.com/stretchr/testify` | v1.11.1 | Test assertions and mocks | Already in use; mock interfaces must be updated for context |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `jinzhu/copier` | `ulule/deepcopier` | copier is simpler API, same author as GORM, more widely used; deepcopier has tag-based mapping but more complex |
| `jinzhu/copier` | Hand-written mapping functions | More control but significantly more boilerplate; user explicitly chose library approach |
| `jinzhu/copier` | `mitchellh/mapstructure` | mapstructure is for map-to-struct, not struct-to-struct; wrong tool |

**Recommendation for Claude's Discretion (mapping library):** Use `jinzhu/copier`. It is the simplest API, created by the same author as GORM (Jinzhu), has 5k+ GitHub stars, supports field name matching, slice copying, and `copier.Option{IgnoreEmpty: true}`. No strong reason to deviate from the user's stated preference.

**Installation:**
```bash
go get github.com/jinzhu/copier
```

## Architecture Patterns

### Recommended Project Structure Changes
```
internal/
  dto/                  # Request/response DTOs (expanded from existing)
    common.go           # Shared types (ErrorDetail, pagination constants, time helpers)
    request.go          # Shared request types (PaginationRequest, IDParam)
    auth.go             # Auth DTOs (LoginRequest, RegisterRequest, AuthResponse, etc.)
    user.go             # User DTOs (UserListQueryParams, ProfileData, etc.)
    role.go             # Role DTOs (RoleCreateRequest, RoleDetailResponse, etc.)
    project.go          # Project DTOs (ProjectCreateRequest, ProjectListData, etc.)
    modul.go            # Modul DTOs (ModulUpdateRequest, ModulListData, etc.)
    tus.go              # TUS upload DTOs (TusUploadInitRequest, TusUploadResponse, etc.)
    statistic.go        # Statistic DTOs
    health.go           # Health check DTOs
    response.go         # API response envelope types (BaseResponse, SuccessResponse, ErrorResponse, PaginationData)
  domain/               # Pure domain models only (GORM entities with TableName methods)
    user.go             # User struct + TableName()
    role.go             # Role, Permission, RolePermission structs + no request/response types
    project.go          # Project struct + no request/response types
    modul.go            # Modul struct + BeforeCreate + TableName()
    tus_upload.go       # TusUpload struct + status constants (no request/response types)
    tus_modul_upload.go # TusModulUpload struct (no request/response types)
  app/
    server.go           # Fiber app setup, middleware, DI wiring
    routes.go           # Route registration functions (extracted from server.go)
```

### Pattern 1: Context Propagation Through All Layers

**What:** Every repository and usecase method receives `context.Context` as its first parameter. Controllers extract context from Fiber via `c.UserContext()` and pass it down.

**When to use:** Every database-touching and external-service-calling method.

**Example:**
```go
// Repository interface (internal/usecase/repo/interfaces.go)
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) error
    GetAll(ctx context.Context, search, filterRole string, page, limit int) ([]domain.UserListItem, int, error)
}

// Repository implementation (internal/usecase/repo/user_repository.go)
func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).Preload("Role").First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Usecase interface (internal/usecase/user_usecase.go)
type UserUsecase interface {
    GetProfile(ctx context.Context, userID string) (*dto.ProfileData, error)
}

// Usecase implementation
func (uc *userUsecase) GetProfile(ctx context.Context, userID string) (*dto.ProfileData, error) {
    user, jumlahProject, jumlahModul, err := uc.userRepo.GetProfileWithCounts(ctx, userID)
    // ...
}

// Controller
func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
    ctx := c.UserContext()
    userID := ctrl.GetAuthenticatedUserID(c)
    if userID == "" {
        return nil
    }
    result, err := ctrl.userUsecase.GetProfile(ctx, userID)
    // ...
}
```

### Pattern 2: DTO Separation with copier Mapping

**What:** Request/response structs live in `internal/dto/`, domain models live in `internal/domain/`. Mapping between them uses `jinzhu/copier` or simple assignment for trivial cases.

**When to use:** Whenever a controller receives or returns data that differs from the domain model. Domain models carry GORM tags and database concerns; DTOs carry JSON tags and validation tags.

**Example:**
```go
// internal/dto/project.go
package dto

type CreateProjectRequest struct {
    NamaProject string `form:"nama_project" validate:"required"`
    Semester    int    `form:"semester" validate:"required,min=1,max=8"`
}

type ProjectResponse struct {
    ID          uint      `json:"id"`
    NamaProject string    `json:"nama_project"`
    Kategori    string    `json:"kategori"`
    Semester    int       `json:"semester"`
    Ukuran      string    `json:"ukuran"`
    PathFile    string    `json:"path_file"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// Mapping: domain -> DTO (using copier for complex cases)
import "github.com/jinzhu/copier"

func ProjectToResponse(p *domain.Project) *ProjectResponse {
    var resp ProjectResponse
    copier.Copy(&resp, p)
    return &resp
}

// For trivial mappings, direct assignment is acceptable:
func ProjectToResponse(p *domain.Project) *ProjectResponse {
    return &ProjectResponse{
        ID:          p.ID,
        NamaProject: p.NamaProject,
        Kategori:    p.Kategori,
        Semester:    p.Semester,
        Ukuran:      p.Ukuran,
        PathFile:    p.PathFile,
        CreatedAt:   p.CreatedAt,
        UpdatedAt:   p.UpdatedAt,
    }
}
```

### Pattern 3: Modular Route Registration

**What:** Route registration is extracted from `server.go` into standalone functions in `routes.go`, one function per domain. Each function receives the API group and all controller/middleware dependencies it needs.

**When to use:** For all route groups.

**Example:**
```go
// internal/app/routes.go
package app

import (
    "invento-service/internal/controller/http"
    "invento-service/internal/helper"
    "github.com/gofiber/fiber/v2"
)

// registerAuthRoutes sets up authentication routes.
func registerAuthRoutes(
    api fiber.Router,
    authController *http.AuthController,
    supabaseAuthService domain.AuthService,
    userRepo repo.UserRepository,
    cookieHelper *httputil.CookieHelper,
) {
    auth := api.Group("/auth")
    auth.Post("/login", authController.Login)
    auth.Post("/register", authController.Register)
    auth.Post("/refresh", authController.RefreshToken)
    auth.Post("/reset-password", authController.RequestPasswordReset)

    protected := auth.Group("/", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
    protected.Post("logout", authController.Logout)
}

// registerUserRoutes sets up user management routes.
func registerUserRoutes(
    api fiber.Router,
    userController *http.UserController,
    authMiddleware fiber.Handler,
    casbinEnforcer helper.CasbinPermissionChecker,
) {
    user := api.Group("/user", authMiddleware)
    user.Get("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceUser, constants.ActionRead), userController.GetUserList)
    // ... more routes
}

// server.go becomes:
// api := app.Group("/api/v1")
// registerAuthRoutes(api, authController, supabaseAuthService, userRepo, cookieHelper)
// registerUserRoutes(api, userController, authMiddleware, casbinEnforcer)
// registerRoleRoutes(api, roleController, userController, authMiddleware, casbinEnforcer)
// registerProjectRoutes(api, projectController, tusController, authMiddleware, casbinEnforcer, cfg)
// registerModulRoutes(api, modulController, tusModulController, authMiddleware, casbinEnforcer, cfg)
// registerStatisticRoutes(api, statisticController, authMiddleware)
// registerMonitoringRoutes(api, healthController)
```

### Pattern 4: Centralized Error-to-HTTP Mapping Middleware

**What:** A Fiber middleware that intercepts `*apperrors.AppError` returns from handlers and automatically maps them to the correct HTTP response, eliminating the repetitive `errors.As(err, &appErr)` pattern in every controller method.

**When to use:** As a global middleware applied before routes, or as a Fiber ErrorHandler.

**Example -- ErrorHandler approach (recommended):**
```go
// internal/app/server.go -- replace current ErrorHandler
ErrorHandler: func(c *fiber.Ctx, err error) error {
    // Handle AppError types centrally
    var appErr *apperrors.AppError
    if errors.As(err, &appErr) {
        return httputil.SendAppError(c, appErr)
    }

    // Handle Fiber errors
    var fiberErr *fiber.Error
    if errors.As(err, &fiberErr) {
        if fiberErr.Code == fiber.StatusNotFound {
            return httputil.SendNotFoundResponse(c, "Endpoint tidak ditemukan")
        }
        return httputil.SendErrorResponse(c, fiberErr.Code, fiberErr.Message, nil)
    }

    // TUS protocol errors
    tusVersion := c.Get("Tus-Resumable")
    if tusVersion != "" && (c.Method() == "PATCH" || c.Method() == "HEAD" || c.Method() == "DELETE") {
        c.Set("Tus-Resumable", cfg.Upload.TusVersion)
        return c.SendStatus(fiber.StatusInternalServerError)
    }

    // Default: internal server error
    appLogger.Error().Str("path", c.Path()).Str("method", c.Method()).Err(err).Msg("unhandled error")
    return httputil.SendInternalServerErrorResponse(c)
},
```

**Controller simplification (before and after):**
```go
// BEFORE (current pattern -- repeated 30+ times across controllers)
result, err := ctrl.userUsecase.GetProfile(userID)
if err != nil {
    var appErr *apperrors.AppError
    if errors.As(err, &appErr) {
        return httputil.SendAppError(c, appErr)
    }
    return ctrl.SendInternalError(c)
}

// AFTER (controller returns error, ErrorHandler maps it)
result, err := ctrl.userUsecase.GetProfile(ctx, userID)
if err != nil {
    return err  // ErrorHandler handles *AppError -> HTTP response
}
```

**Important:** This is a significant simplification but requires careful rollout. Controllers that return errors for non-AppError types (e.g., validation parsing) still need local handling. The transition can be incremental: update ErrorHandler first, then simplify controllers one by one.

### Pattern 5: Test Parallelization with SQLite Isolation

**What:** Add `t.Parallel()` to all test functions that use isolated mock objects. For integration tests using real SQLite databases, each test gets its own in-memory database via a fresh `SetupTestDatabase()` call.

**When to use:** All unit tests (mock-based) can be parallelized immediately. Integration tests need per-test database isolation.

**Example:**
```go
// Unit test (mock-based) -- always safe to parallelize
func TestUserUsecase_GetProfile_Success(t *testing.T) {
    t.Parallel()
    mockUserRepo := new(MockUserRepository)
    // ... setup and assertions
}

// Integration test -- per-test DB isolation
func TestUserRepository_GetByID_Integration(t *testing.T) {
    t.Parallel()
    db, err := testing.SetupTestDatabase()  // Fresh in-memory SQLite per test
    require.NoError(t, err)
    defer testing.TeardownTestDatabase(db)
    // ... test with real DB
}
```

### Anti-Patterns to Avoid
- **Passing `*fiber.Ctx` to usecases or repositories:** Only `context.Context` should cross layer boundaries. Never leak HTTP concerns into business logic.
- **Storing `c.UserContext()` beyond request lifetime:** Fiber (fasthttp) reuses request objects. Extract the context at the start of the handler and pass it; do not store it on long-lived objects.
- **Mixing DTOs and domain models:** Never add `json` response tags to domain model structs; never add `gorm` tags to DTOs. Clean separation prevents accidental field exposure.
- **Global `context.Background()` in handlers:** Always use `c.UserContext()` in controllers -- it inherits any deadline or tracing data set by middleware.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Struct-to-struct copying | Manual field-by-field assignment for every DTO | `jinzhu/copier` | Reduces boilerplate, handles nested structs, maintains field name matching automatically |
| Error-to-HTTP mapping | Switch statement in every controller | Fiber `ErrorHandler` with `errors.As` | Single place for all error-to-HTTP translation, consistent behavior |
| Context propagation | Custom context keys and passing | Standard `context.Context` as first param | Go convention, supported natively by GORM and Fiber |
| Validation | Custom validation logic | `go-playground/validator` tags on DTOs | Already in use, well-tested, declarative |

**Key insight:** The repetitive `errors.As(err, &appErr)` pattern appears 30+ times across controllers. Centralizing this in the Fiber ErrorHandler eliminates the most common boilerplate in the controller layer while maintaining the same HTTP response behavior.

## Common Pitfalls

### Pitfall 1: Context Loss in Goroutines
**What goes wrong:** Background goroutines (like the zip cleanup in `ProjectUsecase.Download`) inherit a request context that gets cancelled when the HTTP request completes, causing operations to fail.
**Why it happens:** `c.UserContext()` is tied to the request lifecycle. Goroutines spawned from handlers that outlive the request see a cancelled context.
**How to avoid:** Use `context.Background()` or `context.WithoutCancel(ctx)` (Go 1.21+) for truly background operations that must outlive the request. The cleanup goroutines in `project_usecase.go` and `tus_cleanup.go` should NOT use the request context.
**Warning signs:** "context cancelled" errors in logs from background operations.

### Pitfall 2: Breaking Mock Interfaces During Context Addition
**What goes wrong:** Adding `context.Context` to repository and usecase interfaces requires updating every mock method signature. Missing a single mock causes compilation failures.
**Why it happens:** The `test_mocks.go` file has 603 lines of hand-written mocks. The `mock.Called()` argument lists must match exactly.
**How to avoid:** Process domain-by-domain. After updating the real interface for a domain, immediately update all corresponding mocks, then run `go build ./...` before moving to the next domain. Use `mock.Anything` for context arguments in most test setups: `mockRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)`.
**Warning signs:** Compilation errors mentioning argument count mismatch in mock calls.

### Pitfall 3: Circular Imports Between dto and domain
**What goes wrong:** DTO types reference domain types (e.g., a response DTO embeds a domain model), and domain types reference DTO types, creating an import cycle.
**Why it happens:** The current codebase has request/response structs in `domain/` that reference domain models directly. Moving them to `dto/` while keeping those references creates cycles.
**How to avoid:** DTOs should only reference primitive types and other DTO types. Mapping functions (in `dto/` or in a `mapper/` sub-package within `dto/`) convert between domain and DTO types. Domain models never import from `dto/`.
**Warning signs:** `import cycle not allowed` compilation error.

### Pitfall 4: DTO Migration Breaking Swagger Annotations
**What goes wrong:** Swagger `@Success` and `@Failure` annotations reference `domain.SuccessResponse`, `domain.ErrorResponse`, `domain.UserListData`, etc. Moving these to `dto/` requires updating every Swagger annotation.
**Why it happens:** Swagger type references are string-based and not checked by the compiler.
**How to avoid:** Update Swagger annotations in the same commit that moves the types. Run `swag init` after migration and verify the generated `docs/` output. Consider doing annotation updates as a dedicated sub-task.
**Warning signs:** `swag init` errors or missing types in Swagger UI.

### Pitfall 5: SQLite BUSY Errors with Test Parallelization
**What goes wrong:** Integration tests sharing a single SQLite database file experience `SQLITE_BUSY` or `database is locked` errors when run with `t.Parallel()`.
**Why it happens:** SQLite uses file-level locking. Concurrent writers conflict.
**How to avoid:** The existing pattern uses `:memory:` databases, which are inherently per-connection and isolated. Each `SetupTestDatabase()` call creates a fresh in-memory database. This is already safe for `t.Parallel()`. Just ensure no test functions share a database variable across parallel subtests without isolation.
**Warning signs:** Flaky tests with "database is locked" errors only when running `go test -parallel N`.

### Pitfall 6: copier Silent Field Mismatches
**What goes wrong:** `copier.Copy` silently ignores fields that don't match between source and destination structs. If a field is renamed in one struct but not the other, data is silently lost.
**Why it happens:** copier matches by field name. `NamaProject` in domain maps to `NamaProject` in DTO, but if DTO renames it to `ProjectName`, copier silently skips it.
**How to avoid:** Write unit tests for every mapping function that verify all expected fields are populated. Use `copier.Option{CaseSensitive: true}` to enforce exact name matching. For fields with different names between domain and DTO, use explicit assignment instead of copier.
**Warning signs:** API responses with missing/zero-value fields that were populated in the domain model.

## Code Examples

### Example 1: Complete Context Propagation for User Domain

```go
// 1. Repository interface (internal/usecase/repo/interfaces.go)
type UserRepository interface {
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    GetByID(ctx context.Context, id string) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) error
    GetAll(ctx context.Context, search, filterRole string, page, limit int) ([]domain.UserListItem, int, error)
    // ... all methods get ctx as first param
}

// 2. Repository implementation
func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    var user domain.User
    err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).Preload("Role").First(&user).Error
    return &user, err
}

// 3. Usecase interface
type UserUsecase interface {
    GetProfile(ctx context.Context, userID string) (*dto.ProfileData, error)
    // ... all methods get ctx as first param
}

// 4. Usecase implementation
func (uc *userUsecase) GetProfile(ctx context.Context, userID string) (*dto.ProfileData, error) {
    user, jumlahProject, jumlahModul, err := uc.userRepo.GetProfileWithCounts(ctx, userID)
    if err != nil {
        // ... error handling
    }
    return uc.userHelper.BuildProfileData(user, jumlahProject, jumlahModul), nil
}

// 5. Controller
func (ctrl *UserController) GetProfile(c *fiber.Ctx) error {
    ctx := c.UserContext()
    userID := ctrl.GetAuthenticatedUserID(c)
    if userID == "" {
        return nil
    }
    result, err := ctrl.userUsecase.GetProfile(ctx, userID)
    if err != nil {
        return err  // ErrorHandler maps AppError -> HTTP response
    }
    return ctrl.SendSuccess(c, result, "Profil user berhasil diambil")
}

// 6. Mock update
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

// 7. Test usage
mockUserRepo.On("GetByID", mock.Anything, "user-1").Return(user, nil)
```

### Example 2: DTO File Structure

```go
// internal/dto/user.go
package dto

import "time"

// --- Requests ---

type UserListQueryParams struct {
    Search     string `query:"search"`
    FilterRole string `query:"filter_role"`
    Page       int    `query:"page"`
    Limit      int    `query:"limit"`
}

type UpdateUserRoleRequest struct {
    Role string `json:"role" validate:"required"`
}

type UpdateProfileRequest struct {
    Name         string `form:"name" validate:"required,min=2,max=100"`
    JenisKelamin string `form:"jenis_kelamin" validate:"omitempty,oneof=Laki-laki Perempuan"`
}

type BulkAssignRoleRequest struct {
    UserIDs []string `json:"user_ids" validate:"required,min=1"`
}

// --- Responses ---

type ProfileData struct {
    Name          string    `json:"name"`
    Email         string    `json:"email"`
    JenisKelamin  *string   `json:"jenis_kelamin,omitempty"`
    FotoProfil    *string   `json:"foto_profil,omitempty"`
    Role          string    `json:"role"`
    CreatedAt     time.Time `json:"created_at"`
    JumlahProject int       `json:"jumlah_project"`
    JumlahModul   int       `json:"jumlah_modul"`
}

type UserListData struct {
    Items      []UserListItem `json:"items"`
    Pagination PaginationData `json:"pagination"`
}

// UserListItem is a DTO -- matches database query output shape
type UserListItem struct {
    ID         string    `json:"id"`
    Email      string    `json:"email"`
    Role       string    `json:"role"`
    DibuatPada time.Time `json:"dibuat_pada"`
}
```

### Example 3: Route Registration Function

```go
// internal/app/routes.go
package app

func registerProjectRoutes(
    api fiber.Router,
    projectController *http.ProjectController,
    tusController *http.TusController,
    authMiddleware fiber.Handler,
    casbinEnforcer helper.CasbinPermissionChecker,
    cfg *config.Config,
) {
    project := api.Group("/project", authMiddleware)
    project.Get("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), projectController.GetList)
    project.Get("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), projectController.GetByID)
    project.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionUpdate), projectController.UpdateMetadata)
    project.Post("/download", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), projectController.Download)
    project.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionDelete), projectController.Delete)

    // TUS upload routes for project
    tusUploadCheck := api.Group("/project/upload", authMiddleware)
    tusUploadCheck.Get("/check-slot", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), tusController.CheckUploadSlot)
    tusUploadCheck.Post("/reset-queue", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionCreate), tusController.ResetUploadQueue)

    tusUpload := api.Group("/project/upload", authMiddleware, helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject))
    tusUpload.Post("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionCreate), tusController.InitiateUpload)
    // ... more TUS routes
}
```

## Context Propagation: Dependency Analysis and Rollout Order

**Recommendation for Claude's Discretion (rollout order):**

Based on analysis of inter-domain dependencies:

| Order | Domain | Methods to Update | Why This Order |
|-------|--------|-------------------|----------------|
| 1 | **Role/Permission** | ~11 repo + 6 usecase | Fewest cross-domain dependencies, smallest blast radius. Only depends on its own repos. |
| 2 | **User** | ~12 repo + 10 usecase | Depends on Role repos (already updated). Most controller methods but straightforward. |
| 3 | **Project** | ~7 repo + 5 usecase | Depends on User repo (done). Has file operations that don't need context. |
| 4 | **Modul** | ~8 repo + 5 usecase | Depends on User repo (done). Similar to Project. |
| 5 | **TUS Upload** | ~14 repo + 12 usecase | Depends on Project repo (done). Most methods, but largely self-contained. |
| 6 | **TUS Modul Upload** | ~11 repo + 11 usecase | Depends on Modul repo (done). Parallel to TUS Upload. |
| 7 | **Statistic** | ~0 repo (uses existing) + 1 usecase | Uses repos already updated. Trivial. |
| 8 | **Health** | 0 repo + 4 usecase | Uses `*gorm.DB` directly, not repositories. Add `WithContext` to direct DB calls. |
| 9 | **Auth** | ~0 repo (uses existing) + 5 usecase | Already uses `context.Context` internally via `domain.AuthService`. Just thread from controller. |

## File Size Analysis and Split Plan

### Source Files Exceeding 500 Lines (Non-Test)
| File | Lines | Action |
|------|-------|--------|
| `clients/go/invento-client/client.go` | 665 | Out of scope (client library) |
| `internal/usecase/test_mocks.go` | 603 | Split by domain (user_mocks.go, role_mocks.go, etc.) |

All other source files are under 500 lines. The DTO migration and route extraction should keep new files well under the limit.

### Test Files Exceeding 500 Lines
| File | Lines | Split Strategy |
|------|-------|----------------|
| `internal/helper/test/tus_helper_test.go` | 2009 | Split by TUS operation (store, queue, manager, cleanup) |
| `internal/controller/http/user_controller_test.go` | 1323 | Split by handler (profile, user_list, role_management) |
| `internal/usecase/user_usecase_test.go` | 1261 | Split by operation (crud, profile, permissions, downloads) |
| `internal/controller/http/tus_controller_test.go` | 1198 | Split by TUS operation (upload, chunk, status, cancel) |
| `internal/httputil/validator_test.go` | 1043 | Split by validator type |
| `internal/domain/tus_upload_test.go` | 1005 | Split by test concern |
| `internal/app/server_test.go` | 975 | Split by route group tested |
| `internal/usecase/auth_usecase_test.go` | 918 | Split by auth operation |
| `internal/controller/http/project_controller_test.go` | 915 | Split by CRUD operation |
| `internal/usecase/statistic_usecase_test.go` | 900 | Split by statistic type |
| `internal/helper/file_test.go` | 878 | Split by file operation type |
| `internal/usecase/role_usecase_test.go` | 849 | Split by role CRUD operation |
| `internal/helper/middleware_test.go` | 787 | Split by middleware type (auth, RBAC, TUS) |
| `internal/domain/tus_modul_upload_test.go` | 780 | Split by test concern |
| `internal/usecase/modul_usecase_test.go` | 769 | Split by modul operation |
| `internal/helper/casbin_test.go` | 757 | Split by casbin operation |
| `internal/usecase/repo/test/tus_repository_test.go` | 754 | Split by entity (upload, modul_upload) |
| `internal/errors/errors_test.go` | 743 | Split by error type |
| `internal/middleware/integration_test.go` | 734 | Split by middleware tested |
| `internal/domain/project_test.go` | 731 | Split by test concern |
| `internal/middleware/validation_test.go` | 721 | Split by validation rule |
| `internal/usecase/project_usecase_test.go` | 717 | Split by project operation |
| `internal/usecase/tus_upload_usecase_test.go` | 662 | Split by TUS operation |
| `internal/integration_test.go` | 639 | Split by integration scope |
| `internal/httputil/cookie_helper_test.go` | 635 | Split by cookie operation |
| `internal/helper/rbac_helper_test.go` | 619 | Split by RBAC operation |
| `internal/usecase/tus_modul_usecase_test.go` | 572 | Split by TUS modul operation |
| `internal/usecase/repo/role_permission_repository_coverage_test.go` | 570 | Split by operation type |
| `internal/controller/http/modul_controller_test.go` | 501 | Just over limit; split by CRUD |

**Recommendation for Claude's Discretion (test split strategy):** Mirror source file naming. For example, `user_usecase_test.go` becomes `user_usecase_crud_test.go`, `user_usecase_profile_test.go`, `user_usecase_permissions_test.go`. This keeps test discoverability aligned with the code they test. Each split file should contain tests for logically related operations, targeting 200-400 lines per file.

## Validation Tag Placement

**Recommendation for Claude's Discretion:** Place validation tags on DTOs only, not on domain models.

**Rationale:**
1. Domain models already carry `gorm` tags for database concerns -- adding `validate` tags creates tag bloat and mixes concerns
2. Validation is an input boundary concern (HTTP layer), not a domain concern
3. The existing `domain.AuthService` interface already follows this pattern: `AuthServiceRegisterRequest` in `domain/auth.go` has validation tags because it serves as both a domain type and a DTO (this is the coupling we're breaking)
4. Repository-layer validation (e.g., NOT NULL constraints) is handled by GORM/database, not by `validator` tags
5. When validation rules differ between create and update operations (e.g., `required` on create vs `omitempty` on update), having validation only on DTOs avoids confusion about which rules apply

## Route File Location

**Recommendation for Claude's Discretion:** Extract to a separate `internal/app/routes.go` file.

**Rationale:**
1. `server.go` is currently 330 lines. After extracting routes (~100 lines of route registration), it drops to ~230 lines for DI wiring, middleware setup, and Fiber config -- a clean single-responsibility file
2. `routes.go` will be approximately 150-200 lines with all register functions -- well under the 500-line limit
3. The route registration functions are self-contained and don't share state with the rest of `server.go` beyond function arguments
4. Having routes in a separate file makes it easy to see the complete API surface at a glance

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No context propagation | `context.Context` as first param in all interface methods | Go best practice since Go 1.7 (2016), GORM added `WithContext` in v2 (2020) | Enables timeouts, cancellation, tracing |
| Request/response types in domain package | Dedicated DTO package separated from domain | Standard clean architecture pattern | Prevents domain model from leaking API concerns |
| Monolithic route registration | Domain-specific route functions | Fiber v2 supports `fiber.Router` interface for composability | Better readability, easier to test routes independently |
| Manual error-to-HTTP in every controller | Centralized ErrorHandler | Fiber v2 supports custom `ErrorHandler` in config | Eliminates boilerplate, consistent error responses |

**Deprecated/outdated:**
- `context.TODO()` as placeholder: Was acceptable during migration, but final code should use `c.UserContext()` from controllers and `context.Background()` only for truly non-request-scoped work
- `context.WithoutCancel` requires Go 1.21+ (project uses Go 1.24, so this is available)

## Open Questions

1. **Should `PaginationData` stay in `domain/` or move to `dto/`?**
   - What we know: `PaginationData` is currently in `domain/response.go` and is used by both `domain.*ListData` types and `httputil.CalculatePagination`. It's a response-only concern, not a domain model.
   - What's unclear: Moving it to `dto/` requires all `*ListData` types to also be in `dto/`, which is the plan. But it creates a transitive dependency: `httputil` -> `dto` for `PaginationData`.
   - Recommendation: Move `PaginationData` to `dto/response.go` along with `BaseResponse`, `SuccessResponse`, `ErrorResponse`, `ListData`, and `ValidationError`. Update `httputil` to reference `dto.PaginationData`. This is clean because `httputil` is already a response-layer concern.

2. **What about `domain.UserListItem` and similar "query result" types?**
   - What we know: Types like `UserListItem`, `ProjectListItem`, `ModulListItem` are used as GORM `Scan` targets in repositories. They are not domain entities but query projections.
   - What's unclear: Should they stay in `domain/` (since GORM scans into them) or move to `dto/` (since they exist purely for API responses)?
   - Recommendation: Move them to `dto/`. Repositories can return `[]dto.UserListItem` directly since these are read-only projections. The `dto` package has no GORM dependency -- GORM uses struct field matching for `Scan`, not GORM tags. Alternatively, keep them as repository-internal types if the dependency direction is a concern; but since `repo` already imports `domain`, replacing `domain` imports with `dto` imports is a neutral change.

3. **How to handle the `domain.AuthService` interface?**
   - What we know: `AuthService` is defined in `domain/auth.go` and already uses `context.Context`. It also includes request/response types (`AuthServiceRegisterRequest`, `AuthServiceResponse`).
   - What's unclear: Should `AuthService` stay in `domain/` as an interface (good) while its request/response types move to `dto/` (also good)?
   - Recommendation: Keep `AuthService` interface in `domain/` (it defines domain behavior). Move `AuthServiceRegisterRequest`, `AuthServiceResponse`, `AuthServiceUserInfo` to `dto/auth.go`. Update the interface to reference `dto` types. This requires `domain` to import `dto`, which is the correct dependency direction (domain defines behavior, dto defines shapes).

## Sources

### Primary (HIGH confidence)
- Codebase analysis: All interface counts, file sizes, dependency maps derived from direct file inspection
- Go standard library: `context.Context` convention (first parameter) is Go best practice per [Go Blog: Context](https://go.dev/blog/context)
- GORM official docs: `WithContext(ctx)` for context propagation, confirmed by codebase's GORM v1.31.0
- Fiber v2 docs: `c.UserContext()` returns `context.Context` for the current request

### Secondary (MEDIUM confidence)
- `jinzhu/copier` GitHub repository: API surface confirmed via search results and README; exact latest version should be verified via `go get`
- SQLite in-memory isolation: Confirmed that each `gorm.Open(sqlite.Open(":memory:"))` call creates an independent database, safe for parallel tests

### Tertiary (LOW confidence)
- None. All findings verified through codebase inspection or official documentation.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use except copier; copier is well-documented with stable API
- Architecture patterns: HIGH - Patterns derived from direct codebase analysis and Go conventions; context propagation is standard Go practice with confirmed GORM/Fiber support
- Pitfalls: HIGH - Pitfalls identified from actual codebase patterns (e.g., goroutine context, mock update count, circular imports) based on code inspection
- File size enforcement: HIGH - Exact line counts measured; split strategy based on actual file contents

**Research date:** 2026-02-16
**Valid until:** 2026-03-16 (30 days -- stable patterns, no fast-moving dependencies)
