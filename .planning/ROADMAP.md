# Roadmap: Invento-Service Refactoring & Optimization

## Overview

This roadmap transforms invento-service from a working-but-rough `fiber-boiler-plate` fork into a production-quality, memory-efficient backend service. The 6 phases are sequenced by dependency: module rename first (all imports depend on it), then config-only memory wins, then code quality standardization, then the high-risk helper package decomposition, then deep architectural changes that touch every interface, and finally verification and polish after all structural changes are complete.

## Milestone 1: Comprehensive Refactoring & Optimization

## Phases

- [x] **Phase 1: Foundation & Rename** - Establish correct module identity, eliminate crashes, and set safety baselines
- [ ] **Phase 2: Memory & Performance Tuning** - Optimize runtime configuration for the 500MB RAM constraint
- [ ] **Phase 3: Code Quality Standardization** - Standardize logging, error handling, and response patterns
- [ ] **Phase 4: Architecture Restructuring** - Decompose the helper god-package into focused, single-responsibility packages
- [ ] **Phase 5: Deep Architecture Improvements** - Propagate context through all layers and enforce file size limits
- [ ] **Phase 6: Polish & Verification** - Remove dead code, verify memory under load, and finalize documentation

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

**Plans:** 5 plans

Plans:
- [ ] 01-01-PLAN.md -- Atomic module rename from fiber-boiler-plate to invento-service
- [ ] 01-02-PLAN.md -- Replace panic/log.Fatal with error returns, add config validation and graceful shutdown
- [ ] 01-03-PLAN.md -- Create RBAC constants package and remove hardcoded secrets from tests
- [ ] 01-04-PLAN.md -- Set up developer tooling (golangci-lint, Taskfile, gofumpt) and memory baselines
- [ ] 01-05-PLAN.md -- Gap closure: fix LoadConfig() signature in test files, rename README heading

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

**Plans:** 2 plans

Plans:
- [ ] 02-01-PLAN.md -- Add PerformanceConfig struct, env var helpers, update Dockerfile GOGC and .env.example
- [ ] 02-02-PLAN.md -- Wire config into database pool, Fiber server, pprof, memory monitor, and runtime settings

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

**Plans:** 3 plans

Plans:
- [ ] 03-01-PLAN.md -- Zerolog foundation, response envelope update, replace internal/logger with zerolog in server.go
- [ ] 03-02-PLAN.md -- Error handling audit (return-after-error), error wrapping, ignored errors, file size dedup
- [ ] 03-03-PLAN.md -- Complete logging migration across all remaining files, delete helper/logger.go, update test assertions

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

**Plans**: TBD

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

**Plans**: TBD

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
- Dockerfile optimization (verify multi-stage build, Alpine base, minimal final image)
- Update `.env.example` with all new configuration options

**Risk**: Memory issues only surface under concurrent load, not in unit tests. Mitigated by running pprof heap profiles during load simulation, not just at rest.

**Plans**: TBD

## Requirements Coverage

Every active requirement from PROJECT.md is mapped to exactly one phase.

| Requirement | ID | Phase | Status |
|-------------|----|-------|--------|
| Rename Go module from `fiber-boiler-plate` to `invento-service` | REN-01 | Phase 1 | Pending |
| Fix inconsistent error handling (missing `return` after error responses) | ERR-01 | Phase 3 | Pending |
| Replace magic strings with constants/config | CFG-01 | Phase 1 + Phase 6 | Pending |
| Remove unused/commented-out code | CLN-01 | Phase 6 | Pending |
| Split large files (>500 lines) into smaller, modular, reusable utilities | ARC-01 | Phase 4 | Pending |
| Fix potential panics on initialization failures | SAF-01 | Phase 1 | Pending |
| Remove hardcoded secrets from test files | SEC-01 | Phase 1 | Pending |
| Fix ignored errors in file operations | ERR-02 | Phase 3 | Pending |
| Implement structured logging (zerolog) | LOG-01 | Phase 3 | Pending |
| Improve architecture consistency across all layers | ARC-02 | Phase 5 | Pending |
| Optimize memory usage for 500MB RAM constraint | MEM-01 | Phase 2 | Pending |
| Standardize API response format across all endpoints | API-01 | Phase 3 | Pending |
| Improve test organization | TST-01 | Phase 5 | Pending |

**Coverage: 13/13 active requirements mapped.**

Note: CFG-01 spans two phases -- constants package created in Phase 1, final audit for remaining magic strings in Phase 6. Primary ownership is Phase 1.

## Progress

**Execution Order:** 1 -> 2 -> 3 -> 4 -> 5 -> 6

| Phase | Plans Complete | Status | Completed |
|-------|---------------|--------|-----------|
| 1. Foundation & Rename | 5/5 | ✓ Complete | 2026-02-15 |
| 2. Memory & Performance Tuning | 0/2 | Planned | - |
| 3. Code Quality Standardization | 0/3 | Planned | - |
| 4. Architecture Restructuring | 0/TBD | Not started | - |
| 5. Deep Architecture Improvements | 0/TBD | Not started | - |
| 6. Polish & Verification | 0/TBD | Not started | - |
