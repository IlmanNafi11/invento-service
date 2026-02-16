# Roadmap: Invento-Service Refactoring & Optimization

## Overview

This roadmap transforms invento-service from a working-but-rough `fiber-boiler-plate` fork into a production-quality, memory-efficient backend service. The 6 phases are sequenced by dependency: module rename first (all imports depend on it), then config-only memory wins, then code quality standardization, then the high-risk helper package decomposition, then deep architectural changes that touch every interface, and finally verification and polish after all structural changes are complete.

## Milestone 1: Comprehensive Refactoring & Optimization

## Phases

- [x] **Phase 1: Foundation & Rename** - Establish correct module identity, eliminate crashes, and set safety baselines
- [x] **Phase 2: Memory & Performance Tuning** - Optimize runtime configuration for the 500MB RAM constraint (completed 2026-02-16)
- [x] **Phase 3: Code Quality Standardization** - Standardize logging, error handling, and response patterns (completed 2026-02-16)
- [x] **Phase 4: Architecture Restructuring** - Decompose the helper god-package into focused, single-responsibility packages (completed 2026-02-16)
- [x] **Phase 5: Deep Architecture Improvements** - Propagate context through all layers and enforce file size limits (completed 2026-02-16)
- [x] **Phase 6: Polish & Verification** - Remove dead code, verify memory under load, and finalize documentation (completed 2026-02-16)
- [ ] **Phase 7: Swagger & Logger Integration Fixes** - Fix Swagger route mismatches and replace global zerolog with injected loggers
- [ ] **Phase 8: File Size Enforcement & Verification** - Split/trim files exceeding 500-line limit and create formal Phase 6 verification

## Phase Details

### Phase 1: Foundation & Rename
**Goal**: The service runs under its correct identity (`invento-service`), never crashes on startup failures, and has baseline tooling for safe refactoring.
**Depends on**: Nothing (first phase)
**Requirements**: REN-01, SAF-01, CFG-01 (partial: constants package), SEC-01

**Success Criteria** (what must be TRUE):
  1. `go build ./...` and `go test ./...` pass with zero references to `fiber-boiler-plate` anywhere in the codebase
  2. The service starts and shuts down gracefully under all configuration scenarios (missing env vars cause clear error messages, not panics)
  3. Swagger UI loads and all endpoints reflect the `invento-service` module path
  4. No hardcoded secrets exist in test files -- all secrets come from environment or config

**Scope:**
- Atomic module rename: `sed` all 292 import occurrences across 116 files, `go mod tidy`, `swag init`
- Grep verification: zero remaining `fiber-boiler-plate` references
- Replace 4 `panic()` calls with error returns + graceful shutdown (`signal.NotifyContext` + `app.ShutdownWithContext`)
- Config validation at startup (fail-fast on missing critical env vars like `SUPABASE_URL`)
- Create constants package for RBAC magic strings and other repeated literals
- Remove hardcoded secrets from test files, replace with env/config references
- Set `GOMEMLIMIT=350MiB` and `GOGC=50` as baseline (Dockerfile + .env.example)
- Configure `golangci-lint` v2+ and `gofumpt` for automated quality checks
- Add Taskfile v3 for common commands (`task build`, `task test`, `task lint`)

**Risk**: Module rename leaves partial old references, breaking compilation. Mitigated by atomic sed + grep verification + full build/test pass.

**Plans:** 5/5 plans complete

Plans:
- [x] 01-01-PLAN.md -- Atomic module rename from fiber-boiler-plate to invento-service
- [x] 01-02-PLAN.md -- Replace panic/log.Fatal with error returns, add config validation and graceful shutdown
- [x] 01-03-PLAN.md -- Create RBAC constants package and remove hardcoded secrets from tests
- [x] 01-04-PLAN.md -- Set up developer tooling (golangci-lint, Taskfile, gofumpt) and memory baselines
- [x] 01-05-PLAN.md -- Gap closure: fix LoadConfig() signature in test files, rename README heading

### Phase 2: Memory & Performance Tuning
**Goal**: The service runs within a measurable memory budget, with profiling tools available to validate all subsequent changes stay within the 500MB constraint.
**Depends on**: Phase 1 (needs correct module path and GOMEMLIMIT baseline)
**Requirements**: MEM-01

**Success Criteria** (what must be TRUE):
  1. `go tool pprof` heap profile shows baseline memory usage, accessible via localhost-only pprof endpoint
  2. Fiber is configured with `ReduceMemoryUsage: true`, `StreamRequestBody: true`, and `Concurrency: 1024`
  3. Database connection pool uses `MaxOpenConns=10` and `MaxIdleConns=3` (verified via GORM config)
  4. GORM runs with `SkipDefaultTransaction=true` for non-transactional queries

**Scope:**
- Fiber config: `ReduceMemoryUsage: true`, `StreamRequestBody: true`, `Concurrency: 1024`
- GORM config: `MaxOpenConns=10`, `MaxIdleConns=3`, `SkipDefaultTransaction=true`
- Add pprof endpoint bound to localhost only (not exposed externally)
- Establish memory profiling baseline under realistic load
- Verify PgBouncer compatibility (never enable `PrepareStmt: true`, keep `QueryExecModeSimpleProtocol`)

**Risk**: `ReduceMemoryUsage` adds ~10-15% CPU overhead. Mitigated by benchmarking request latency before/after.

**Plans:** 2/2 plans complete

Plans:
- [x] 02-01-PLAN.md -- Add PerformanceConfig struct, env var helpers, update Dockerfile GOGC and .env.example
- [x] 02-02-PLAN.md -- Wire config into database pool, Fiber server, pprof, memory monitor, and runtime settings

### Phase 3: Code Quality Standardization
**Goal**: All logging uses structured zerolog, all errors are handled consistently with proper wrapping, and all API responses follow a single format.
**Depends on**: Phase 2 (memory baseline established; need to verify zerolog does not regress memory)
**Requirements**: LOG-01, ERR-01, ERR-02, API-01

**Success Criteria** (what must be TRUE):
  1. Zero stdlib `log.*` calls remain in the codebase -- all logging uses zerolog with structured fields
  2. Every controller error response is followed by a `return` statement (no fall-through after error responses)
  3. All file operation errors (modul delete, cleanup) are checked and logged, not silently ignored
  4. Every API endpoint returns responses in a consistent JSON envelope format (verified by inspecting 3+ different endpoint responses)

**Scope:**
- Replace 35+ stdlib `log.*` calls with zerolog across 10+ files
- Add `fiberzerolog` middleware for automatic request/response logging
- Remove the custom `internal/logger` package (it allocates `map[string]interface{}` per call -- GC pressure)
- Audit every controller: ensure `return` after all error response calls
- Wrap errors with `fmt.Errorf("context: %w", err)` consistently
- Fix ignored errors in file operations (modul delete, cleanup paths)
- Extract `HandleError` pattern to a `BaseController` or shared error handler
- Standardize API response envelope across all endpoints
- Eliminate duplicate code (file size formatting appears 3 times)
- Repository error translation: `gorm.ErrRecordNotFound` mapped to domain-level errors

**Risk**: Changing error handling patterns across all controllers has a wide blast radius. Mitigated by running `go test ./...` after each file change, never batching.

**Plans:** 3/3 plans complete

Plans:
- [x] 03-01-PLAN.md -- Zerolog foundation, response envelope update, replace internal/logger with zerolog in server.go
- [x] 03-02-PLAN.md -- Error handling audit (return-after-error), error wrapping, ignored errors, file size dedup
- [x] 03-03-PLAN.md -- Complete logging migration across all remaining files, delete helper/logger.go, update test assertions

### Phase 4: Architecture Restructuring
**Goal**: The `internal/helper/` god-package is decomposed into focused packages with clear single responsibilities, and no circular dependencies exist.
**Depends on**: Phase 3 (standardized patterns make code cleaner to split; stable test suite required)
**Requirements**: ARC-01

**Success Criteria** (what must be TRUE):
  1. The `internal/helper/` directory contains only domain-specific utility functions (not TUS, RBAC, HTTP, storage, or middleware code)
  2. Each extracted package (`httputil`, `storage`, `rbac`, `middleware`, `upload`) builds independently with no circular imports
  3. `go vet ./...` and `golangci-lint run` pass with zero import cycle errors
  4. All existing tests pass without modification to test assertions (only import paths change)

**Scope:**
- Map all dependencies within `internal/helper/` (27 source files, 52 including tests)
- Extract in strict dependency order (leaf packages first):
  1. `internal/httputil/` -- HTTP utility functions (no internal deps)
  2. `internal/storage/` -- File operation utilities
  3. `internal/rbac/` -- Casbin enforcer setup and RBAC helpers
  4. `internal/middleware/` -- HTTP middleware (depends on httputil, rbac)
  5. `internal/upload/` -- TUS upload handling (depends on storage)
- Keep `TurnOffAutoMigrate` + `NewCasbinEnforcer` together during rbac extraction
- Clean remaining `internal/helper/` to domain-specific utils only
- Update all import paths across the codebase
- Verify Casbin integration test passes after move

**Risk**: Circular dependencies surface during extraction, requiring unexpected refactoring. Mitigated by mapping all deps before moving any code, extracting leaf packages first.

**Plans:** 6/6 plans complete

Plans:
- [x] 04-01-PLAN.md -- Extract httputil package (HTTP response helpers, status constants, pagination, query parsing, cookie, validator)
- [x] 04-02-PLAN.md -- Extract storage package (file operations, FileManager, PathResolver, domain helpers)
- [x] 04-03-PLAN.md -- Extract rbac package (Casbin enforcer, RBAC helpers, constants from internal/constants/)
- [x] 04-04-PLAN.md -- Extract middleware functions into existing internal/middleware/ package (auth, RBAC, TUS middleware)
- [x] 04-05-PLAN.md -- Extract upload package (TUS store, queue, manager, cleanup, headers, metadata, response)
- [x] 04-06-PLAN.md -- Final cleanup: inline orphan email.go, delete internal/helper/ entirely

### Phase 5: Deep Architecture Improvements
**Goal**: All layers accept `context.Context` for proper timeout/cancellation support, route registration is modular, and no source file exceeds 500 lines.
**Depends on**: Phase 4 (packages must be in final locations before changing every interface signature)
**Requirements**: ARC-02, TST-01

**Success Criteria** (what must be TRUE):
  1. Every repository, usecase, and controller interface method accepts `context.Context` as its first parameter
  2. Route registration is split into domain-specific files (not one monolithic `server.go`)
  3. No Go source file in the project exceeds 500 lines (verified by `wc -l`)
  4. Test files are organized by concern, large test files are split, and `go test ./... -count=1` passes

**Scope:**
- Add `context.Context` to all repository interface methods
- Add `context.Context` to all usecase interface methods
- Thread Fiber's `c.UserContext()` through controller -> usecase -> repository
- Split `server.go` route registration into domain-specific route files
- Request/response DTO separation (domain models vs API contracts)
- Centralized error-to-HTTP mapping middleware
- Enforce max 500 lines per file -- split any oversized files
- Split large test files into focused test files by concern
- Parallelize independent test suites with `t.Parallel()`

**Risk**: Context propagation changes every interface signature -- largest blast radius of any phase. Mitigated by changing one domain at a time (e.g., user first, then modul, then project) and running tests after each domain.

**Plans:** 10/10 plans complete

Plans:
- [x] 05-01-PLAN.md -- Foundation: install copier, migrate response types to dto, extract routes.go, centralized ErrorHandler
- [x] 05-02-PLAN.md -- DTO migration: move all domain-specific request/response types to dto/, create mapping functions
- [x] 05-03-PLAN.md -- Context propagation for Role/Permission and User domains
- [x] 05-04-PLAN.md -- Context propagation for Project, Modul, Auth, Statistic, and Health domains
- [x] 05-05-PLAN.md -- Context propagation for TUS Upload and TUS Modul Upload domains (completes context rollout)
- [x] 05-06-PLAN.md -- Split test_mocks.go and top 15 oversized test files (>750 lines)
- [x] 05-07-PLAN.md -- Split remaining 14 oversized test files (501-750 lines)
- [x] 05-08-PLAN.md -- Test parallelization (t.Parallel()) and final 500-line enforcement verification
- [x] 05-09-PLAN.md -- Gap closure: context.Context for AuthUsecase, HealthUsecase, StatisticUsecase interfaces
- [x] 05-10-PLAN.md -- Gap closure: trim 5 test files to under 500-line limit

### Phase 6: Polish & Verification
**Goal**: The codebase is clean, verified under load to stay within memory limits, and all documentation reflects the final state.
**Depends on**: Phase 5 (all structural changes must be complete before final verification)
**Requirements**: CLN-01, CFG-01 (remaining: audit for any remaining magic strings)

**Success Criteria** (what must be TRUE):
  1. Zero commented-out code blocks or unused functions remain (verified by `golangci-lint` with `deadcode` and `unused` linters enabled)
  2. Memory profiling under simulated load shows heap usage stays below 350MB
  3. `golangci-lint run` passes with the full linter configuration (no new warnings)
  4. Swagger documentation is up-to-date and all endpoints render correctly in the Swagger UI

**Scope:**
- Remove all unused/commented-out code across the codebase
- Final audit for remaining magic strings or hardcoded values
- Full `golangci-lint` pass with comprehensive linter set
- Memory profiling under simulated concurrent load (5+ concurrent uploads)
- Test coverage audit (identify any untested critical paths)
- Regenerate Swagger docs (`swag init`) and verify all endpoints
- Update `.env.example` with all new configuration options

**Risk**: Memory issues only surface under concurrent load, not in unit tests. Mitigated by running pprof heap profiles during load simulation, not just at rest.

**Plans:** 6/6 plans complete

Plans:
- [x] 06-01-PLAN.md -- Fix golangci-lint config, extract magic string constants, remove dead code and TODO comments
- [x] 06-02-PLAN.md -- Fix failing tests in rbac, usecase, and app packages
- [x] 06-03-PLAN.md -- Add Swagger annotations to all missing TUS and user management endpoints
- [x] 06-04-PLAN.md -- Comprehensive golangci-lint pass (zero warnings) and gofumpt formatting
- [x] 06-05-PLAN.md -- Regenerate Swagger docs, update .env.example, create memory verification checklist
- [x] 06-06-PLAN.md -- Test coverage audit: add tests for config (32.5%) and usecase/repo (65.5%) packages

### Phase 7: Swagger & Logger Integration Fixes
**Goal**: Swagger documentation accurately reflects actual routes, and all logging uses injected zerolog (no global logger bypass).
**Depends on**: Phase 6 (all structural changes complete)
**Requirements**: API-01 (residual), LOG-01 (residual)
**Gap Closure**: Closes gaps from v1 milestone audit (integration 3→4, integration 5→6)

**Success Criteria** (what must be TRUE):
  1. Swagger @Router annotations for ComprehensiveHealthCheck and GetApplicationStatus match actual registered routes
  2. Regenerated Swagger docs (`swag init`) reflect correct paths
  3. `internal/middleware/rbac_middleware.go` uses injected `zerolog.Logger` instead of global `log.Warn()`
  4. `internal/storage/download_helper.go` uses injected `zerolog.Logger` instead of global `log.Error()`
  5. `golangci-lint run` passes with zero warnings

**Scope:**
- Fix 2 Swagger @Router annotations in health_controller.go
- Regenerate Swagger docs with `swag init -g cmd/app/main.go -o docs/`
- Inject zerolog.Logger via struct field or function parameter in rbac_middleware.go
- Inject zerolog.Logger via struct field or function parameter in download_helper.go
- Run golangci-lint as final gate

**Risk**: Minimal -- targeted fixes to 4 files. Mitigated by running `go test ./...` after each change.

**Plans:** 0/0 plans complete

### Phase 8: File Size Enforcement & Verification
**Goal**: All source files comply with the 500-line limit, and Phase 6 has formal verification documentation.
**Depends on**: Phase 7 (swagger fixes must be done before final verification)
**Requirements**: ARC-01 (residual)
**Gap Closure**: Closes gaps from v1 milestone audit (500-line violations, missing verification)

**Success Criteria** (what must be TRUE):
  1. `internal/controller/http/tus_controller.go` is under 500 lines (currently 540)
  2. `config/integration_test.go` is under 500 lines (currently 501)
  3. `clients/go/invento-client/client.go` is under 500 lines (currently 665)
  4. `06-VERIFICATION.md` exists covering all Phase 6 success criteria
  5. `go test ./... -count=1` passes after all splits

**Scope:**
- Split or trim tus_controller.go (540 → <500 lines)
- Trim config/integration_test.go (501 → <500 lines)
- Split clients/go/invento-client/client.go (665 → <500 lines)
- Create formal 06-VERIFICATION.md documenting all Phase 6 criteria verification
- Final `go test ./...` pass

**Risk**: File splits may break imports or test coverage. Mitigated by splitting one file at a time with test verification.

**Plans:** 0/0 plans complete

## Requirements Coverage

Every active requirement from PROJECT.md is mapped to exactly one phase.

| Requirement | ID | Phase | Status |
|-------------|----|-------|--------|
| Rename Go module from `fiber-boiler-plate` to `invento-service` | REN-01 | Phase 1 | Complete |
| Fix inconsistent error handling (missing `return` after error responses) | ERR-01 | Phase 3 | Complete |
| Replace magic strings with constants/config | CFG-01 | Phase 1 + Phase 6 | Complete |
| Remove unused/commented-out code | CLN-01 | Phase 6 | Complete |
| Split large files (>500 lines) into smaller, modular, reusable utilities | ARC-01 | Phase 4 | Complete |
| Fix potential panics on initialization failures | SAF-01 | Phase 1 | Complete |
| Remove hardcoded secrets from test files | SEC-01 | Phase 1 | Complete |
| Fix ignored errors in file operations | ERR-02 | Phase 3 | Complete |
| Implement structured logging (zerolog) | LOG-01 | Phase 3 | Complete |
| Improve architecture consistency across all layers | ARC-02 | Phase 5 | Complete |
| Optimize memory usage for 500MB RAM constraint | MEM-01 | Phase 2 | Complete |
| Standardize API response format across all endpoints | API-01 | Phase 3 | Complete |
| Improve test organization | TST-01 | Phase 5 | Complete |

**Coverage: 13/13 active requirements mapped.**

Note: CFG-01 spans two phases -- constants package created in Phase 1, final audit for remaining magic strings in Phase 6. Primary ownership is Phase 1.

## Progress

**Execution Order:** 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 8

| Phase | Plans Complete | Status | Completed |
|-------|---------------|--------|-----------|
| 1. Foundation & Rename | 5/5 | Complete | 2026-02-15 |
| 2. Memory & Performance Tuning | 2/2 | Complete | 2026-02-16 |
| 3. Code Quality Standardization | 3/3 | Complete | 2026-02-16 |
| 4. Architecture Restructuring | 6/6 | Complete | 2026-02-16 |
| 5. Deep Architecture Improvements | 10/10 | Complete | 2026-02-16 |
| 6. Polish & Verification | 6/6 | Complete | 2026-02-16 |
| 7. Swagger & Logger Integration Fixes | 0/0 | Pending | — |
| 8. File Size Enforcement & Verification | 0/0 | Pending | — |
