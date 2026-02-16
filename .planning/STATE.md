# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-15)

**Core value:** File storage that is reliable and resource-efficient on a 500MB RAM server -- upload, store, and download student files without failure.
**Current focus:** Phase 6 in progress — Polish & Verification

## Current Position

Phase: 6 of 6 (Polish & Verification)
Plan: 4 of 6 complete
Status: Phase 6 In Progress
Last activity: 2026-02-16 -- Plan 06-01 (Static Analysis & Code Cleanup) complete

Progress: [██████░░░░] 67% (4/6 plans)

## Performance Metrics

**Velocity:**
- Total plans completed: 27
- Average duration: ~9min
- Total execution time: ~4.5 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-foundation-rename | 5 | ~40min | ~8min |
| 02-memory-performance-tuning | 2 | ~12min | ~6min |
| 03-code-quality-standardization | 3 | ~25min | ~8min |
| 04-architecture-restructuring | 6 | ~55min | ~9min |
| 05-deep-architecture-improvements | 10 | ~219min | ~22min |

**Recent Trend:**
- Last 5 plans: 05-06, 05-07, 05-08, 05-09, 05-10
- Trend: Stable ~20-45min per plan

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Roadmap]: zerolog chosen over slog/zap (zero-allocation, official Fiber middleware)
- [Roadmap]: Module path `invento-service` (simple, not `github.com/...` -- internal module)
- [Roadmap]: GOMEMLIMIT=350MiB with GOGC=50 (conservative for stability)
- [Roadmap]: Keep hand-written mocks (no mockery/gomock migration)
- [Roadmap]: Fail-fast config validation (no silent defaults)

- [01-03]: All RBAC string literals replaced with typed constants for compile-time safety
- [01-03]: Test tokens confirmed as test fixtures (no real credentials); annotated with comments
- [01-05]: Used require.NoError for LoadConfig error checking in tests (fail-fast over assert)

- [02-01]: GOGC=100 (Go default) instead of 50 per user decision -- prioritizes speed over aggressive GC
- [02-01]: FiberReduceMemory defaults to false -- balanced approach per user discussion
- [02-01]: SkipDefaultTransaction NOT included -- user decided to keep GORM default

- [02-02]: Used *logger.Logger (pointer) for startMemoryMonitor since Logger is a struct
- [02-02]: FiberStreamRequestBody=false and FiberConcurrency=256 in test configs for stability

- [03-03]: Boot logger pattern (zerolog.New(os.Stderr)) for pre-config fatal logging in main.go
- [03-03]: ConnectDatabase accepts zerolog.Logger parameter instead of using global logger
- [03-03]: Global zerolog/log used in middleware/usecases where DI would over-complicate signatures
- [03-03]: zerolog.Nop() used in all test constructors to avoid log noise

- [04-01]: helper/tus_response.go imports httputil for Send* functions (delegation pattern)
- [04-01]: helper/middleware.go imports httputil for CookieHelper and response helpers
- [04-01]: TUS controllers keep helper import only (no httputil needed for SendTus* functions)
- [04-01]: Files using only httputil symbols have helper import removed entirely

- [04-04]: CasbinPermissionChecker interface stays in middleware package (accept-interfaces principle)
- [04-04]: rbac_middleware.go filename avoids confusion with internal/rbac/ package
- [04-04]: server.go keeps helper import for TUS store/queue/manager symbols

- [04-06]: ValidatePolijeEmail inlined as private function in auth_usecase.go (single consumer, no separate package)
- [04-06]: internal/helper/ fully deleted — god-package decomposition complete

- [05-01]: Response types (BaseResponse, SuccessResponse, ErrorResponse, etc.) migrated from domain/ to dto/ with copier mapper
- [05-01]: Route registration extracted from server.go into routes.go with routeDeps struct
- [05-01]: Centralized ErrorHandler uses errors.As for AppError + fiber.Error with dto.ErrorResponse format

- [05-02]: All domain-specific request/response types migrated to dto/ package (auth, user, role, project, modul, stat, health, TUS)
- [05-02]: Repository interfaces keep domain types; controllers and usecases use dto types

- [05-03]: context.Context on all Role/Permission/User domain repo/usecase interfaces
- [05-03]: RBAC helper interfaces updated with context.Context

- [05-04]: context.Context on all Project/Modul repo interfaces and all remaining usecase interfaces
- [05-04]: TUS usecases use context.Background() as temporary bridge (05-05 will fix)

- [05-05]: ALL 70 repository interface methods now accept context.Context -- project-wide context propagation complete
- [05-05]: Background cleanup goroutines use context.Background() to avoid request context cancellation
- [05-05]: domain.TusUploadMetadata/TusModulUploadMetadata used in repo tests (not dto types)

- [05-06]: Used _test.go suffix for mock files (no external packages import usecase mocks)
- [05-06]: Shared test helpers extracted to *_helpers_test.go (setupTestDB, assertRoleUsecaseAppError)
- [05-06]: Domain-specific mock naming: {domain}_mocks_test.go; test splits by concern: *_crud_test.go, *_status_test.go, etc.

- [05-07]: Split rbac_helper_test.go with BuildRoleDetailResponse in check file to keep setup under 500 lines
- [05-07]: TUS test split pattern: init (helpers, slot checks, initiate, info/status) vs chunk (handle chunk, completion, cancel)
- [05-08]: Controller subtests NOT parallelized - they share Fiber app instances per top-level test
- [05-08]: auth_integration_test.go NOT parallelized - uses file::memory:?cache=shared SQLite DSN
- [05-08]: TestVerify_ClockSkewTolerance NOT parallelized - modifies global jwt.TimeFunc
- [05-10]: Removed only blank lines and redundant comments to trim files — zero test cases deleted
- [05-08]: Fixed pre-existing mock bugs from 05-04 context propagation (args.Get index, missing mock.Anything)
- [Phase 05-09]: Thread ctx through health private helpers (getDatabaseStatus, getDetailedDatabaseStatus, getServicesStatus) for full context propagation; PingContext(ctx) for database checks

- [06-03]: TUS Upload tag for project endpoints, TUS Modul Upload for modul endpoints, Role Management for user role endpoints
- [06-03]: Modul IDs use string type (UUID) in Swagger @Param, project IDs use int

### Pending Todos

None yet.

### Blockers/Concerns

- Actual memory usage under load not yet measured (estimate is ~250MB baseline)
- SQLite vs PostgreSQL test divergence extent unknown

## Session Continuity

Last session: 2026-02-16
Stopped at: Completed 06-03-PLAN.md
Resume file: .planning/phases/06-polish-verification/06-03-SUMMARY.md
