---
phase: 10-go-query-optimization
plan: 02
subsystem: database
tags: [gorm, raw-sql, query-optimization, statistics, sqlite, parallel-tests]

requires:
  - phase: 10-go-query-optimization
    provides: Joins-based User repository optimizations and raw SQL patterns from Plan 01
provides:
  - Consolidated GetStatistics using single raw SQL query (4 queries -> 1)
  - Named shared-memory SQLite pattern for parallel test stability
affects: [statistic-controller, dashboard-endpoint]

tech-stack:
  added: []
  patterns: [raw SQL subquery counts for multi-table statistics, named shared-memory SQLite for parallel subtests]

key-files:
  created: []
  modified: [internal/usecase/statistic_usecase.go, internal/usecase/statistic_usecase_admin_test.go, internal/usecase/statistic_usecase_user_test.go, internal/usecase/statistic_usecase_errors_test.go]

key-decisions:
  - "Used named shared-memory SQLite DBs (file:stat_<testname>?mode=memory&cache=shared) for parallel subtest stability instead of anonymous :memory:"
  - "Slight over-fetching acceptable: always count all 4 tables even if user only has partial permissions, filter results after"

patterns-established:
  - "Named shared-memory SQLite: Use file:<unique>?mode=memory&cache=shared for tests with t.Parallel() subtests"
  - "Raw SQL subquery counts: SELECT (SELECT COUNT(*) FROM x) AS y pattern for multi-table statistics"

requirements-completed: [GORM-03]

duration: 6min
completed: 2026-02-17
---

# Phase 10 Plan 02: Statistic Usecase Query Optimization Summary

**Consolidated dashboard statistics from 4 separate COUNT queries to 1 raw SQL with subquery counts, with named shared-memory SQLite fix for parallel test stability**

## Performance

- **Duration:** 6 min
- **Started:** 2026-02-17T11:45:40Z
- **Completed:** 2026-02-17T11:51:14Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- GetStatistics now executes a single `db.Raw()` query with 4 correlated subquery COUNTs instead of up to 4 separate repository calls
- Permission-based filtering preserved: unpermitted counts return nil in the response
- All 16 statistic tests pass with real DB counts instead of mocked CountByUserID calls
- Fixed parallel subtest instability caused by SQLite `:memory:` connection isolation

## Task Commits

Each task was committed atomically:

1. **Task 1: Consolidate GetStatistics into single query** - `a40fbda` (feat)
2. **Task 2: Fix test files for consolidated query pattern** - `f9d644c` (fix)

## Files Created/Modified
- `internal/usecase/statistic_usecase.go` - Replaced 4 sequential repo calls with single `db.Raw()` subquery SQL, added early return for no permissions
- `internal/usecase/statistic_usecase_admin_test.go` - Updated `setupTestDB` to use named shared-memory SQLite (`file:stat_<test>?mode=memory&cache=shared`), added `domain.Project` and `domain.Modul` to AutoMigrate, added `seedProjectModulData` helper, updated test wrapper to match consolidated query pattern
- `internal/usecase/statistic_usecase_user_test.go` - Replaced mock `CountByUserID` expectations with real DB seeding via `seedProjectModulData`
- `internal/usecase/statistic_usecase_errors_test.go` - Replaced mock `CountByUserID` expectations with real DB seeding

## Decisions Made
- Used named shared-memory SQLite DBs (`file:stat_<testname>?mode=memory&cache=shared`) instead of anonymous `:memory:` to fix parallel subtest isolation — each goroutine/connection to `:memory:` creates an independent database, losing tables
- Accepted slight over-fetching: the raw SQL always counts all 4 tables regardless of permissions, then filters after — simpler code, negligible cost for COUNT on small tables

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed SQLite :memory: parallel subtest instability**
- **Found during:** Task 2 (test file updates)
- **Issue:** Tests `VariousUserIDs` and `MixedPermissions` used `t.Parallel()` subtests sharing an in-memory SQLite DB created with `sqlite.Open(":memory:")`. Each parallel goroutine's connection to `:memory:` creates an independent database, causing "no such table" errors
- **Fix:** Changed `setupTestDB()` to use unique named shared-memory databases: `fmt.Sprintf("file:stat_%s?mode=memory&cache=shared", t.Name())`
- **Files modified:** internal/usecase/statistic_usecase_admin_test.go
- **Verification:** All 16 statistic tests pass including parallel subtests
- **Committed in:** f9d644c (part of Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** Essential fix for test reliability. Named shared-memory is standard practice for SQLite parallel tests. No scope creep.

## Issues Encountered
None beyond the auto-fixed parallel test instability.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 10 complete (both plans done): User repository and statistic usecase queries fully optimized
- Pattern established: named shared-memory SQLite for any future tests with `t.Parallel()` subtests
- Ready for Phase 11 (next milestone phase)

## Self-Check: PASSED

All files verified present, all commit hashes found in git history.

---
*Phase: 10-go-query-optimization*
*Completed: 2026-02-17*
