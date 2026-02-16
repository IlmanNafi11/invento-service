# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-15)

**Core value:** File storage that is reliable and resource-efficient on a 500MB RAM server -- upload, store, and download student files without failure.
**Current focus:** Phase 4 in progress — Architecture Restructuring

## Current Position

Phase: 4 of 6 (Architecture Restructuring)
Plan: 4 of 6 complete
Status: In Progress
Last activity: 2026-02-16 -- Plan 04-04 (extract middleware functions) complete

Progress: [██████░░░░] 67% (4/6 plans)

## Performance Metrics

**Velocity:**
- Total plans completed: 14
- Average duration: ~8min
- Total execution time: ~1.8 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-foundation-rename | 5 | ~40min | ~8min |
| 02-memory-performance-tuning | 2 | ~12min | ~6min |
| 03-code-quality-standardization | 3 | ~25min | ~8min |
| 04-architecture-restructuring | 4 | ~45min | ~11min |

**Recent Trend:**
- Last 5 plans: 03-03, 04-01, 04-02, 04-03, 04-04
- Trend: Stable ~8-15min per plan

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

### Pending Todos

None yet.

### Blockers/Concerns

- Actual memory usage under load not yet measured (estimate is ~250MB baseline)
- SQLite vs PostgreSQL test divergence extent unknown

## Session Continuity

Last session: 2026-02-16
Stopped at: Completed 04-04-PLAN.md
Resume file: None
