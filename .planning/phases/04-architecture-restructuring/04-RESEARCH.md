# Phase 4: Architecture Restructuring - Research

**Researched:** 2026-02-16
**Domain:** Go package decomposition / internal restructuring
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Keep roadmap-proposed names: `httputil`, `storage`, `rbac`, `middleware`, `upload`
- `httputil` and `middleware` remain separate packages (httputil = HTTP utilities/response helpers; middleware = Fiber middleware like auth, CORS, rate limiting)
- RBAC constants (from Phase 1 constants package) move into `internal/rbac/` alongside Casbin enforcer setup
- All TUS-related code (handler, hooks, configuration) goes into `internal/upload/`
- `TurnOffAutoMigrate` stays with `NewCasbinEnforcer` in `internal/rbac/` (tightly coupled)
- Extract one package at a time, each as its own plan with tests passing after each extraction
- Extraction order follows dependency graph (leaf packages first): `httputil` -> `storage` -> `rbac` -> `middleware` -> `upload`
- Clean break on each extraction -- update all import paths immediately, no re-export wrappers from helper
- Test files move alongside the code they test in the same extraction step
- Nothing stays in helper -- `internal/helper/` is deleted entirely after all extractions complete
- Dead code (functions with zero callers) is deleted during extraction, not preserved

### Claude's Discretion
- Handling orphan utility functions (inline vs move to consumer)
- Internal file organization within each new package
- Exact dependency mapping and extraction boundaries for edge cases

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope.
</user_constraints>

## Summary

This phase decomposes `internal/helper/` (27 source files, 25 test files, plus a `test/` subdirectory and an `uploads/` data directory) into five focused packages: `httputil`, `storage`, `rbac`, `middleware` (note: the existing `internal/middleware/` already exists and will absorb the auth/RBAC middleware functions), and `upload`. The helper package is a classic Go "god package" where unrelated concerns -- HTTP response formatting, Casbin RBAC enforcement, TUS upload protocol handling, file system operations, cookie management -- all share the same package namespace.

The primary risk is the **naming conflict**: `internal/middleware/` already exists with `logging.go`, `requestid.go`, and `validation.go`. The middleware functions being extracted from helper (`SupabaseAuthMiddleware`, `RBACMiddleware`, `TusProtocolMiddleware`) must be merged INTO this existing package rather than creating a new one. This is a critical detail the planner must account for.

The codebase uses Go 1.24.0 with GORM, Fiber v2, Casbin v2, and `go-playground/validator`. The helper package is imported by 50+ files across controllers, usecases, middleware, testing, and app initialization. Every extraction step will touch many files for import path updates.

**Primary recommendation:** Follow the locked extraction order strictly (httputil -> storage -> rbac -> middleware -> upload), run `go build ./...` and `go test ./...` after each extraction, and use `goimports` to fix import paths. The middleware extraction must merge into the existing `internal/middleware/` package, not create a new one.

## Standard Stack

### Core (Already in Project)
| Library | Version | Purpose | Relevant to Phase |
|---------|---------|---------|-------------------|
| Go | 1.24.0 | Language runtime | Package import rules, `go vet`, `goimports` |
| gofiber/fiber/v2 | v2.52.11 | HTTP framework | `fiber.Ctx` used in httputil, middleware, upload |
| casbin/casbin/v2 | v2.127.0 | RBAC enforcement | casbin.go, rbac_helper.go move to `internal/rbac/` |
| casbin/gorm-adapter/v3 | v3.37.0 | Casbin DB adapter | Used in `NewCasbinEnforcer`, moves with it |
| gorm.io/gorm | v1.31.0 | ORM | Used by Casbin adapter only in helper |
| go-playground/validator/v10 | v10.27.0 | Struct validation | validator.go in helper (orphan candidate) |

### Supporting (Go Toolchain)
| Tool | Purpose | When to Use |
|------|---------|-------------|
| `goimports` | Auto-fix import paths after moves | After every file move |
| `go build ./...` | Verify no compile errors / import cycles | After every extraction step |
| `go test ./...` | Verify test suite passes | After every extraction step |
| `go vet ./...` | Catch common issues | After every extraction step |
| `golangci-lint run` | Full lint check | After final cleanup |

### No New Dependencies
This phase introduces zero new libraries. It is purely an internal reorganization.

## Architecture Patterns

### Current Structure (Before)
```
internal/
  helper/             # GOD PACKAGE - 27 source files
    casbin.go           # CasbinEnforcer, CasbinEnforcerInterface
    rbac_helper.go      # RBACHelper, SetRolePermissions, etc.
    middleware.go       # SupabaseAuthMiddleware, RBACMiddleware, TusProtocolMiddleware
    response.go         # SendSuccessResponse, SendErrorResponse, Send*Response
    http_status.go      # HTTP status constants, StatusText, DefaultMessages
    cookie_helper.go    # CookieHelper
    pagination.go       # PaginationParams, CalculatePagination
    query_parser.go     # ParsePaginationQuery, ParseSearchQuery, etc.
    validator.go        # ValidateStruct (struct validation)
    file.go             # File utilities: GenerateUniqueIdentifier, SaveUploadedFile, etc.
    file_manager.go     # FileManager (config-aware file operations)
    path_resolver.go    # PathResolver (environment-aware path resolution)
    download_helper.go  # DownloadHelper
    logger.go           # Logger (simple wrapper)
    email.go            # ValidatePolijeEmail
    project_helper.go   # ProjectHelper
    modul_helper.go     # ModulHelper
    user_helper.go      # UserHelper
    tus_store.go        # TusStore, TusFileInfo
    tus_queue.go        # TusQueue
    tus_manager.go      # TusManager
    tus_cleanup.go      # TusCleanup
    tus_headers.go      # TUS header constants & parsers
    tus_metadata.go     # ParseTusMetadata
    tus_response.go     # SendTus*Response functions
  middleware/         # ALREADY EXISTS
    logging.go
    requestid.go
    validation.go       # Uses helper.SendBadRequestResponse, helper.ValidateStruct
```

### Target Structure (After)
```
internal/
  httputil/           # NEW - HTTP response helpers, status codes, pagination, query parsing
    response.go         # Send*Response functions
    http_status.go      # Status constants and helpers
    pagination.go       # PaginationParams, CalculatePagination, CalculateOffset
    query_parser.go     # ParsePaginationQuery, ParseSearchQuery, etc.
    cookie_helper.go    # CookieHelper
    validator.go        # ValidateStruct (the helper one, wrapping go-playground/validator)
  storage/            # NEW - File system operations
    file.go             # GenerateUniqueIdentifier, SaveUploadedFile, DeleteFile, MoveFile, etc.
    file_manager.go     # FileManager
    path_resolver.go    # PathResolver
    download_helper.go  # DownloadHelper
    project_helper.go   # ProjectHelper
    modul_helper.go     # ModulHelper
    user_helper.go      # UserHelper
  rbac/               # NEW - RBAC/Casbin enforcement
    casbin.go           # CasbinEnforcer, CasbinEnforcerInterface, NewCasbinEnforcer
    rbac_helper.go      # RBACHelper
    constants.go        # RBAC resource/action constants (moved from internal/constants/rbac.go)
  middleware/         # EXISTING - absorbs auth/RBAC middleware from helper
    logging.go          # (existing)
    requestid.go        # (existing)
    validation.go       # (existing - update imports from helper to httputil)
    auth.go             # SupabaseAuthMiddleware (moved from helper/middleware.go)
    rbac.go             # RBACMiddleware (moved from helper/middleware.go)
    tus.go              # TusProtocolMiddleware (moved from helper/middleware.go)
  upload/             # NEW - TUS upload protocol
    tus_store.go        # TusStore, TusFileInfo
    tus_queue.go        # TusQueue
    tus_manager.go      # TusManager
    tus_cleanup.go      # TusCleanup
    tus_headers.go      # TUS header constants & parsers
    tus_metadata.go     # ParseTusMetadata
    tus_response.go     # SendTus*Response functions
```

### Pattern: Incremental Leaf-First Extraction
**What:** Extract packages starting with those that have zero internal dependencies (leaf packages), then work up the dependency tree.
**When to use:** Always, for any god-package decomposition.
**Why:** Each extraction step must compile and pass tests. Leaf packages have no internal deps so they can be moved without changing anything in the source files themselves -- only import paths in consumers change.

### Pattern: Clean Break, No Re-exports
**What:** When moving code from package A to package B, immediately update ALL consumers to import from B. Do not leave re-export aliases in A.
**When to use:** This is a locked decision for this phase.
**Why:** Re-exports create a false impression that the old package still provides the functionality, leading to confusion and eventually needing another migration.

### Anti-Patterns to Avoid
- **Partial extraction with fallback imports:** Never leave `helper.SendErrorResponse` working while also having `httputil.SendErrorResponse`. Clean break on each step.
- **Moving tests separately from code:** Test files must move in the same commit as their source. Otherwise tests reference symbols that no longer exist in the package.
- **Circular dependency from middleware -> httputil -> middleware:** The middleware package already imports helper for response functions. After extraction, middleware will import httputil for responses. The httputil package must NEVER import middleware. This is naturally safe since httputil is a leaf package.

## Detailed Dependency Map

### File-to-Package Assignment

| Source File | Target Package | Internal Deps (within helper) | External Deps |
|-------------|---------------|-------------------------------|---------------|
| `response.go` | `httputil` | domain, errors, version | fiber |
| `http_status.go` | `httputil` | (none) | (none) |
| `pagination.go` | `httputil` | domain | math |
| `query_parser.go` | `httputil` | (none) | fiber |
| `cookie_helper.go` | `httputil` | config | fiber |
| `validator.go` | `httputil` | domain | go-playground/validator |
| `file.go` | `storage` | (none) | archive/zip, crypto/rand, os |
| `file_manager.go` | `storage` | `file.go` (GenerateRandomString) | config, os |
| `path_resolver.go` | `storage` | config | os |
| `download_helper.go` | `storage` | `path_resolver.go`, `logger.go`, `file.go` | domain, os |
| `project_helper.go` | `storage` | `file.go` (GetFileExtension) | config |
| `modul_helper.go` | `storage` | `file.go` (GetFileExtension) | config |
| `user_helper.go` | `storage` | `path_resolver.go`, `file.go` | config, domain |
| `casbin.go` | `rbac` | (none) | casbin, gorm-adapter, gorm |
| `rbac_helper.go` | `rbac` | `casbin.go` (CasbinEnforcerInterface) | domain |
| `middleware.go` | `middleware` | `response.go`, `cookie_helper.go`, `casbin.go` | domain, fiber, usecase/repo |
| `logger.go` | orphan | (none) | log, os |
| `email.go` | orphan | (none) | strings |
| `tus_store.go` | `upload` | `path_resolver.go` | os, sync, json |
| `tus_queue.go` | `upload` | (none) | sync |
| `tus_manager.go` | `upload` | `tus_store.go`, `tus_queue.go`, `file_manager.go` | config, domain, errors |
| `tus_cleanup.go` | `upload` | `tus_store.go` | domain, time |
| `tus_headers.go` | `upload` | (none) | fiber |
| `tus_metadata.go` | `upload` | (none) | encoding/base64 |
| `tus_response.go` | `upload` | `tus_headers.go`, `response.go` | fiber |

### Cross-Package Dependencies After Extraction

```
httputil     <- (leaf, no internal deps)
storage      <- (leaf, no internal deps)
rbac         <- (leaf, no internal deps -- casbin.go only uses external gorm/casbin)
middleware   <- httputil (for Send*Response)
             <- rbac (for CasbinPermissionChecker interface, CasbinEnforcerInterface)
             <- httputil (for CookieHelper via cookie_helper.go)
upload       <- storage (for PathResolver via tus_store.go, FileManager via tus_manager.go)
             <- httputil (for Send*Response via tus_response.go)
```

**Critical finding:** The `middleware.go` file currently imports `usecase/repo` for `UserRepository`. After extraction, the `internal/middleware/` package will import `internal/usecase/repo`. This is already the case pattern-wise (the existing middleware package does not import repo, but the code being moved in does). This must be verified as acceptable.

### Extraction Order Validation

The locked order `httputil -> storage -> rbac -> middleware -> upload` is correct based on the dependency graph:

1. **httputil** (leaf): No deps on other helper files. Can be extracted first. Only uses external packages (fiber, domain, errors, version).
2. **storage** (leaf): No deps on other helper files. `file.go`, `file_manager.go`, `path_resolver.go` are self-contained. `download_helper.go` uses `NewLogger()` from `logger.go` -- this is an orphan that should be inlined or replaced.
3. **rbac** (leaf): `casbin.go` only uses external casbin/gorm. `rbac_helper.go` uses `CasbinEnforcerInterface` from `casbin.go` which will be in the same `rbac` package.
4. **middleware** (depends on httputil, rbac): `middleware.go` uses `SendErrorResponse`, `SendUnauthorizedResponse`, etc. (from httputil) and `CasbinPermissionChecker` (from rbac). After httputil and rbac are extracted, these imports just change to point at the new packages.
5. **upload** (depends on storage, httputil): `tus_store.go` uses `PathResolver` (from storage). `tus_manager.go` uses `FileManager` (from storage). `tus_response.go` uses `Send*Response` (from httputil). All these are already extracted by this point.

## Consumer Impact Analysis

### Files That Import `internal/helper` (50+ files)

**High-impact consumers (use many helper symbols):**
| File | helper symbols used | Post-extraction imports |
|------|--------------------|-----------------------|
| `internal/app/server.go` | 15+ symbols (NewPathResolver, NewCasbinEnforcer, NewTusStore, NewTusQueue, NewTusManager, NewFileManager, NewCookieHelper, NewTusCleanup, SupabaseAuthMiddleware, RBACMiddleware, TusProtocolMiddleware, SendInternalServerErrorResponse) | httputil, storage, rbac, middleware, upload |
| `internal/controller/base/controller.go` | 12+ symbols (CasbinEnforcer type, Send*Response, StatusOK, StatusCreated, StatusBadRequest, StatusInternalServerError, GetDefaultMessage, ValidateStruct) | httputil, rbac |
| `internal/controller/http/tus_controller.go` | 10+ symbols (SendTus*Response, GetTusHeaders, SetTus*Headers, HeaderUploadMetadata, ParseTusMetadata) | upload, httputil |
| `internal/controller/http/tus_modul_controller.go` | 10+ symbols (same TUS response/header functions) | upload, httputil |
| `internal/controller/http/tus_helpers.go` | 8+ symbols (HeaderTusResumable, HeaderContentType, TusContentType, GetTusHeaders, ValidateChunkSize, SendTusErrorResponse, SendAppError, SendInternalServerErrorResponse) | upload, httputil |

**Medium-impact consumers (use a few helper symbols):**
- `internal/controller/http/auth_controller.go` -- CookieHelper type, Logger type, Send*Response -> httputil
- `internal/controller/http/project_controller.go` -- CasbinEnforcer type, SendAppError -> rbac, httputil
- `internal/controller/http/user_controller.go` -- SendAppError -> httputil
- `internal/controller/http/modul_controller.go` -- SendAppError -> httputil
- `internal/usecase/role_usecase.go` -- RBACHelper, CasbinEnforcerInterface -> rbac
- `internal/usecase/user_usecase.go` -- PathResolver, UserHelper -> storage
- `internal/usecase/project_usecase.go` -- FileManager -> storage
- `internal/usecase/tus_upload_usecase.go` -- TusManager, FileManager, PathResolver -> upload, storage
- `internal/usecase/tus_modul_usecase.go` -- TusManager, FileManager, ModulHelper -> upload, storage

**Test file consumers:**
- `internal/testing/casbin.go` -- CasbinEnforcerInterface -> rbac
- `internal/usecase/test/mock_casbin.go` -- CasbinEnforcerInterface -> rbac
- All `*_test.go` files in helper/ move alongside their source files

### Existing `internal/middleware/` Package Conflict

**CRITICAL:** The existing `internal/middleware/` package already contains:
- `logging.go` -- request logging middleware
- `requestid.go` -- request ID middleware
- `validation.go` -- **imports `internal/helper` for SendBadRequestResponse, SendValidationErrorResponse, ValidateStruct**

When middleware functions from `helper/middleware.go` are moved INTO this existing package:
1. `validation.go` import of `internal/helper` must change to `internal/httputil` (for response functions)
2. The new `auth.go`, `rbac.go`, `tus.go` files must be added to this package
3. The new files will import `internal/httputil` and `internal/rbac` -- no circular dependency risk

### Existing `internal/constants/rbac.go` Migration

The RBAC constants currently in `internal/constants/rbac.go`:
```go
const (
    ResourcePermission = "Permission"
    ResourceRole       = "Role"
    ResourceUser       = "User"
    ResourceProject    = "Project"
    ResourceModul      = "Modul"
    ActionRead     = "read"
    ActionCreate   = "create"
    ActionUpdate   = "update"
    ActionDelete   = "delete"
    ActionDownload = "download"
)
```
These move into `internal/rbac/constants.go` per the locked decision. The `internal/constants/` package may have other files; if `rbac.go` is the ONLY file, the entire `internal/constants/` directory can be deleted. Currently it has only `rbac.go` and `rbac_test.go`, so it can be deleted entirely.

All consumers using `constants.ResourcePermission`, `constants.ActionRead`, etc. (primarily `internal/app/server.go`) must update to `rbac.ResourcePermission`, `rbac.ActionRead`, etc.

## Orphan Functions Analysis (Claude's Discretion)

### logger.go -- Logger struct
**Current usage:** `download_helper.go` creates `NewLogger()` inline for error logging. Also used in `auth_controller.go` as `helper.NewLogger()`.
**Recommendation:** The project already has `internal/logger/` with a sophisticated structured logger. The helper Logger is a simple thin wrapper. **Inline the two call sites** -- `download_helper.go` can use Go's standard `log.Printf()` directly, and `auth_controller.go` should use the existing `internal/logger` package. Delete `logger.go`.
**Confidence:** HIGH -- `internal/logger/` is clearly the intended structured logging solution. The helper Logger is dead-walking code.

### email.go -- ValidatePolijeEmail
**Current usage:** Used in `internal/usecase/auth_usecase.go` (`helper.ValidatePolijeEmail()`).
**Recommendation:** **Move to the closest consumer.** Since it is only used in `auth_usecase.go`, it can be inlined as a private function in `internal/usecase/auth_usecase.go`, or placed as a utility in the `internal/usecase/` package. It has zero helper-internal dependencies (only stdlib `strings` and `errors`).
**Confidence:** HIGH -- single consumer, no dependencies.

### validator.go -- ValidateStruct
**Current usage:** Used in `internal/controller/base/controller.go` and `internal/middleware/validation.go`.
**Recommendation:** **Move to `httputil`** since it is an HTTP-layer validation helper that produces `domain.ValidationError` slices for HTTP responses. It pairs naturally with the response functions. The existing `internal/validator/` package handles custom validators (password_strength, file_type, etc.) -- it is a different concern.
**Confidence:** HIGH -- tight coupling with response layer.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Import path updates | Manual find-and-replace | `goimports -w .` after each extraction | Handles formatting, removes unused, adds missing |
| Circular dependency detection | Manual inspection | `go build ./...` (compiler catches import cycles) | Go compiler is authoritative |
| Dead code detection | Manual grep | `go vet ./...` and check for unused exports | Catches unreachable code |
| Test verification | Running individual tests | `go test ./...` | Full suite after each extraction |

**Key insight:** Go's compiler is the authoritative tool for detecting import cycles. If `go build ./...` passes, there are no circular dependencies. Do not build custom validation -- just compile.

## Common Pitfalls

### Pitfall 1: Middleware Package Name Collision
**What goes wrong:** Creating a NEW `internal/middleware/` package when one already exists causes a compile error, or accidentally creating it as a sub-package (e.g., `internal/middleware/auth/`).
**Why it happens:** The extraction plan says "extract to `internal/middleware/`" but doesn't emphasize that this package already exists.
**How to avoid:** Add new files (`auth.go`, `rbac.go`, `tus.go`) to the EXISTING `internal/middleware/` directory. Update the existing `validation.go` import paths in the same step.
**Warning signs:** Compilation errors about duplicate package names or unexpected package paths.

### Pitfall 2: Interface Ownership After RBAC Extraction
**What goes wrong:** After moving `CasbinEnforcerInterface` to `internal/rbac/`, code in `internal/testing/casbin.go` and `internal/usecase/test/mock_casbin.go` still references `helper.CasbinEnforcerInterface`, causing compile errors.
**Why it happens:** These files implement the interface and have compile-time checks (`var _ helper.CasbinEnforcerInterface = (*TestCasbinEnforcer)(nil)`).
**How to avoid:** Update ALL interface references to `rbac.CasbinEnforcerInterface` in the same extraction step. Search for `helper.CasbinEnforcerInterface` across the entire codebase.
**Warning signs:** `go build ./...` fails with type mismatch errors.

### Pitfall 3: CasbinPermissionChecker vs CasbinEnforcerInterface
**What goes wrong:** `middleware.go` defines `CasbinPermissionChecker` interface (for `RBACMiddleware`), separate from `CasbinEnforcerInterface` in `casbin.go`. After extraction, if `CasbinPermissionChecker` moves to middleware but middleware needs to import rbac for the concrete type, it can create confusion.
**Why it happens:** Two overlapping interfaces exist -- `CasbinPermissionChecker` (1 method) and `CasbinEnforcerInterface` (15+ methods).
**How to avoid:** Keep `CasbinPermissionChecker` in the middleware package (where `RBACMiddleware` lives). It is a minimal interface following Go's "accept interfaces, return structs" principle. The concrete `CasbinEnforcer` in `rbac` satisfies it automatically. No import from middleware -> rbac is needed for the interface -- only the caller (app/server.go) wires them together.
**Warning signs:** Import cycle between middleware and rbac.

### Pitfall 4: TUS Response Functions Depend on httputil
**What goes wrong:** `tus_response.go` calls `SendBadRequestResponse`, `SendNotFoundResponse`, `SendForbiddenResponse`, etc. After extraction, these are in `httputil`. The `upload` package would need to import `httputil`.
**Why it happens:** TUS response helpers are thin wrappers around general HTTP response helpers.
**How to avoid:** Accept this as a valid dependency: `upload -> httputil`. This is a leaf-to-leaf dependency, not circular. Verify with `go build ./...`.
**Warning signs:** None -- this is architecturally sound.

### Pitfall 5: download_helper.go Uses PathResolver AND Logger
**What goes wrong:** `download_helper.go` calls `NewLogger()` (orphan) and uses `PathResolver` (storage). If logger.go is deleted before download_helper.go is moved, compilation fails.
**Why it happens:** Ordering issues when orphan functions are deleted out of sequence.
**How to avoid:** When moving `download_helper.go` to `storage`, simultaneously replace the `NewLogger()` call with `log.Printf()` from stdlib, since the Logger is being deleted as an orphan.
**Warning signs:** Undefined reference to `NewLogger` during storage extraction.

### Pitfall 6: Forgetting to Move Constants from internal/constants/
**What goes wrong:** `internal/constants/rbac.go` constants are duplicated or left behind, causing two different import paths for the same constants.
**Why it happens:** The constants are not in `internal/helper/` -- they are in a separate package. Easy to forget during helper extraction.
**How to avoid:** Include the `internal/constants/rbac.go` -> `internal/rbac/constants.go` move as an explicit step in the `rbac` extraction plan. Delete `internal/constants/` entirely after the move since it only contains RBAC files.
**Warning signs:** Two packages exporting the same constant names.

### Pitfall 7: Exported CasbinEnforcer Struct Field in base/controller.go
**What goes wrong:** `base/controller.go` has `Casbin *helper.CasbinEnforcer` as an exported struct field. This is a concrete type reference, not an interface. After rbac extraction, it must become `Casbin *rbac.CasbinEnforcer`.
**Why it happens:** The field uses the concrete type, which forces the import path to follow the type.
**How to avoid:** Update the type reference during rbac extraction. Also update `project_controller.go` which uses `*helper.CasbinEnforcer` in its constructor.
**Warning signs:** Type not found errors during compilation.

## Code Examples

### Example 1: Moving a file and updating imports (httputil extraction)

**Before** (`internal/helper/response.go`):
```go
package helper

import (
    "invento-service/internal/domain"
    apperrors "invento-service/internal/errors"
    "invento-service/internal/version"
    "github.com/gofiber/fiber/v2"
)

func SendSuccessResponse(c *fiber.Ctx, code int, message string, data interface{}) error {
    // ...
}
```

**After** (`internal/httputil/response.go`):
```go
package httputil  // ONLY the package declaration changes

import (
    "invento-service/internal/domain"
    apperrors "invento-service/internal/errors"
    "invento-service/internal/version"
    "github.com/gofiber/fiber/v2"
)

func SendSuccessResponse(c *fiber.Ctx, code int, message string, data interface{}) error {
    // ... identical body
}
```

**Consumer update** (`internal/controller/base/controller.go`):
```go
// Before:
import "invento-service/internal/helper"
// helper.SendSuccessResponse(c, helper.StatusOK, message, data)

// After:
import "invento-service/internal/httputil"
// httputil.SendSuccessResponse(c, httputil.StatusOK, message, data)
```

### Example 2: Middleware merge (adding auth.go to existing middleware/)

**New file** (`internal/middleware/auth.go`):
```go
package middleware  // joins existing middleware package

import (
    "invento-service/internal/domain"
    "invento-service/internal/httputil"    // was helper, now httputil
    "invento-service/internal/usecase/repo"
    "strings"
    "github.com/gofiber/fiber/v2"
)

// CasbinPermissionChecker stays in middleware package (small interface)
type CasbinPermissionChecker interface {
    CheckPermission(roleName, resource, action string) (bool, error)
}

func SupabaseAuthMiddleware(authService domain.AuthService, userRepo repo.UserRepository, cookieHelper *httputil.CookieHelper) fiber.Handler {
    // ... body uses httputil.SendErrorResponse instead of SendErrorResponse
}
```

### Example 3: RBAC constants migration

**Before** (`internal/constants/rbac.go`):
```go
package constants

const (
    ResourcePermission = "Permission"
    // ...
)
```

**After** (`internal/rbac/constants.go`):
```go
package rbac

const (
    ResourcePermission = "Permission"
    // ...
)
```

**Consumer update** (`internal/app/server.go`):
```go
// Before:
import "invento-service/internal/constants"
// role.Get("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceRole, constants.ActionRead), ...)

// After:
import "invento-service/internal/rbac"
// role.Get("/", middleware.RBACMiddleware(casbinEnforcer, rbac.ResourceRole, rbac.ActionRead), ...)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single helper/ package | Domain-focused packages | This phase | Clearer responsibilities, testability |
| Constants in separate constants/ pkg | Constants co-located with domain (rbac/) | This phase | Better cohesion |
| Helper Logger wrapper | Structured logger in internal/logger/ | Phase 3 (already done) | Helper Logger is now dead code |

**Deprecated/outdated after this phase:**
- `internal/helper/` -- deleted entirely
- `internal/constants/` -- deleted entirely (contents moved to rbac/)
- `helper.Logger` -- replaced by `internal/logger/`

## File Count Summary

| Target Package | Source Files | Test Files | Total |
|---------------|-------------|------------|-------|
| `httputil` | 6 (response.go, http_status.go, pagination.go, query_parser.go, cookie_helper.go, validator.go) | ~8 (corresponding test files + coverage tests) | ~14 |
| `storage` | 7 (file.go, file_manager.go, path_resolver.go, download_helper.go, project_helper.go, modul_helper.go, user_helper.go) | ~7 (corresponding test files + coverage tests) | ~14 |
| `rbac` | 3 (casbin.go, rbac_helper.go, constants.go from internal/constants/) | ~3 (casbin_test.go, rbac_helper_test.go, rbac_test.go from constants/) | ~6 |
| `middleware` | 3 new files added (auth.go, rbac.go, tus.go from helper/middleware.go split) | 1 moved (middleware_test.go) | ~4 new |
| `upload` | 7 (tus_store.go, tus_queue.go, tus_manager.go, tus_cleanup.go, tus_headers.go, tus_metadata.go, tus_response.go) | ~7 (corresponding test files + coverage tests) | ~14 |
| Orphans (deleted/inlined) | 2 (logger.go, email.go) | ~2 (logger_test.go, email_test.go) | ~4 deleted |
| **Total** | **26 moved + 2 deleted = 28** | **~28** | **~56** |

Note: helper/test/ subdirectory contains additional test files (validator_test.go, response_test.go, tus_helper_test.go) that need to move to their respective new packages.

## Open Questions

1. **helper/test/ subdirectory test files**
   - What we know: `internal/helper/test/` contains `validator_test.go`, `response_test.go`, `tus_helper_test.go` -- these are integration/coverage tests for helper functions
   - What's unclear: Whether these test files belong in the new package's test directory or should be restructured
   - Recommendation: Move each test file to the package that owns the code it tests. `validator_test.go` -> httputil, `response_test.go` -> httputil, `tus_helper_test.go` -> upload. May need package declaration updates.

2. **middleware.go splitting strategy**
   - What we know: `middleware.go` contains 3 functions (SupabaseAuthMiddleware, RBACMiddleware, TusProtocolMiddleware) plus a `CasbinPermissionChecker` interface
   - What's unclear: Whether to split into 3 files (auth.go, rbac.go, tus.go) or keep as one file
   - Recommendation: Split into separate files for clarity. `auth.go` (SupabaseAuthMiddleware), `rbac.go` (RBACMiddleware + CasbinPermissionChecker), `tus.go` (TusProtocolMiddleware). This matches the existing middleware package convention of one concern per file.

3. **middleware_test.go split**
   - What we know: `middleware_test.go` is 23KB with tests for all 3 middleware functions
   - What's unclear: Whether to split into matching test files or keep as one
   - Recommendation: Split into `auth_test.go`, `rbac_test.go`, `tus_test.go` to match the source file split

## Sources

### Primary (HIGH confidence)
- Direct codebase analysis of all 27 source files in `internal/helper/`
- Direct codebase analysis of all 50+ consumer files importing `internal/helper`
- `go.mod` file confirming Go 1.24.0, fiber v2.52.11, casbin v2.127.0
- Existing `internal/middleware/`, `internal/constants/`, `internal/validator/`, `internal/logger/` packages analyzed

### Secondary (MEDIUM confidence)
- Go official documentation on import cycle detection (compiler-enforced, no external tools needed)
- Go best practices for package extraction (accept interfaces, return structs; leaf-first extraction)

### Tertiary (LOW confidence)
- None -- all findings are based on direct codebase analysis

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- directly verified from go.mod and source files
- Architecture: HIGH -- dependency map built from reading every source file
- Pitfalls: HIGH -- identified from actual cross-references in the codebase
- Extraction order: HIGH -- validated against actual dependency graph

**Research date:** 2026-02-16
**Valid until:** 2026-03-16 (stable -- this is internal restructuring, no external API changes)
