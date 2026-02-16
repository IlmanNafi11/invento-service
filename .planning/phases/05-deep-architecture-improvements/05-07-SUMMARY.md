---
phase: 05-deep-architecture-improvements
plan: 07
subsystem: testing
tags: [go-test, test-splitting, file-organization, single-responsibility]

# Dependency graph
requires:
  - phase: 05-05
    provides: context.Context propagation complete across all interfaces
provides:
  - All test files in the project under 500 lines
  - 14 oversized test files (501-750 lines) split into 28 focused files
affects: [05-08]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Test file splitting by concern: *_init_test.go, *_chunk_test.go, *_crud_test.go, *_query_test.go"
    - "Shared test helpers and mocks accessible across split files in same package"

key-files:
  created:
    - internal/usecase/tus_upload_usecase_init_test.go
    - internal/usecase/tus_upload_usecase_chunk_test.go
    - internal/usecase/tus_modul_usecase_init_test.go
    - internal/usecase/tus_modul_usecase_chunk_test.go
    - internal/usecase/repo/role_permission_repo_crud_test.go
    - internal/usecase/repo/role_permission_repo_query_test.go
    - internal/httputil/cookie_helper_set_test.go
    - internal/httputil/cookie_helper_get_test.go
    - internal/integration_setup_test.go
    - internal/integration_api_test.go
    - internal/rbac/rbac_helper_setup_test.go
    - internal/rbac/rbac_helper_check_test.go
    - internal/controller/http/modul_controller_crud_test.go
    - internal/controller/http/modul_controller_download_test.go
  modified: []

key-decisions:
  - "Split rbac_helper_test.go with BuildRoleDetailResponse in check file to keep setup under 500 lines"
  - "Duplicate TestTusModulUsecase_HandleModulChunk_OffsetMismatch removed from init file (kept in chunk)"

patterns-established:
  - "TUS test split: init (helpers, slot checks, initiate, info/status) vs chunk (handle chunk, completion, cancel)"
  - "Controller test split: CRUD operations vs download/export operations"
  - "Repository test split: CRUD operations vs complex query/workflow tests"

# Metrics
duration: ~45min
completed: 2026-02-16
---

# Phase 5 Plan 7: Split Remaining Oversized Test Files Summary

**14 oversized test files (501-750 lines) split into 28 focused test files under 500 lines each, completing project-wide 500-line enforcement**

## Performance

- **Duration:** ~45 min (across multiple sessions)
- **Started:** 2026-02-16
- **Completed:** 2026-02-16T08:48:58Z
- **Tasks:** 2
- **Files modified:** 28 (14 originals deleted, 14 new files created per batch)

## Accomplishments
- Split all 14 remaining oversized test files into 28 focused files
- Every test file in the project is now under 500 lines
- All tests pass with zero regressions
- Clean unused import removal across all split files

## Task Commits

Each task was committed atomically:

1. **Task 1: Split Batch A (7 files)** - `578e053` (refactor)
2. **Task 2: Split Batch B (7 files)** - `62ef5cf` (refactor)

## Files Created/Modified

### Batch A (Task 1 - committed in prior session)
- `internal/helper/casbin_enforcer_test.go` - Casbin enforcer operation tests
- `internal/helper/casbin_policy_test.go` - Casbin policy management tests
- `internal/usecase/repo/test/tus_upload_repository_test.go` - TUS upload repo tests
- `internal/usecase/repo/test/tus_modul_upload_repository_test.go` - TUS modul upload repo tests
- `internal/errors/errors_apperror_test.go` - AppError type tests
- `internal/errors/errors_sentinel_test.go` - Sentinel error tests
- `internal/errors/errors_helpers_test.go` - Error helper tests
- `internal/middleware/integration_auth_test.go` - Auth middleware integration
- `internal/middleware/integration_rbac_test.go` - RBAC middleware integration
- `internal/middleware/integration_validation_test.go` - Validation middleware integration
- `internal/domain/project_validation_test.go` - Project validation tests
- `internal/domain/project_model_test.go` - Project model tests
- `internal/middleware/validation_struct_test.go` - Struct validation tests
- `internal/middleware/validation_rules_test.go` - Validation rules tests
- `internal/usecase/project_usecase_crud_test.go` - Project CRUD usecase tests
- `internal/usecase/project_usecase_download_test.go` - Project download usecase tests

### Batch B (Task 2)
- `internal/usecase/tus_upload_usecase_init_test.go` (258 lines) - TUS upload init/status tests
- `internal/usecase/tus_upload_usecase_chunk_test.go` (427 lines) - TUS upload chunk/cancel tests
- `internal/usecase/tus_modul_usecase_init_test.go` (299 lines) - TUS modul init/status tests
- `internal/usecase/tus_modul_usecase_chunk_test.go` (291 lines) - TUS modul chunk/update tests
- `internal/usecase/repo/role_permission_repo_crud_test.go` (295 lines) - Role permission CRUD tests
- `internal/usecase/repo/role_permission_repo_query_test.go` (289 lines) - Role permission query tests
- `internal/httputil/cookie_helper_set_test.go` (268 lines) - Cookie set operations
- `internal/httputil/cookie_helper_get_test.go` (378 lines) - Cookie get/clear/security tests
- `internal/integration_setup_test.go` (233 lines) - Middleware chain, error handling, request ID
- `internal/integration_api_test.go` (420 lines) - DTO validation, pagination, CRUD cycle
- `internal/rbac/rbac_helper_setup_test.go` (486 lines) - RBAC mocks, validation, set/remove permissions
- `internal/rbac/rbac_helper_check_test.go` (143 lines) - Permission check, save policy, build response
- `internal/controller/http/modul_controller_crud_test.go` (375 lines) - Modul CRUD controller tests
- `internal/controller/http/modul_controller_download_test.go` (144 lines) - Modul download tests

### Originals Deleted (Batch B)
- `internal/usecase/tus_upload_usecase_test.go` (667 lines)
- `internal/usecase/tus_modul_usecase_test.go` (574 lines)
- `internal/usecase/repo/role_permission_repository_coverage_test.go` (571 lines)
- `internal/httputil/cookie_helper_test.go` (635 lines)
- `internal/integration_test.go` (638 lines)
- `internal/rbac/rbac_helper_test.go` (620 lines)
- `internal/controller/http/modul_controller_test.go` (502 lines)

## Decisions Made
- Split rbac_helper_test.go with BuildRoleDetailResponse tests in check file (not setup) to keep setup under 500 lines
- Duplicate `TestTusModulUsecase_HandleModulChunk_OffsetMismatch` removed from init file since it belongs with chunk handling tests
- Unused imports (`bytes`, `fmt`, `time`, `errors`, `apperrors`) cleaned up during split to pass `go vet`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Duplicate test function across split files**
- **Found during:** Task 2 (tus_modul_usecase split)
- **Issue:** `TestTusModulUsecase_HandleModulChunk_OffsetMismatch` was present in both init and chunk files
- **Fix:** Removed from init file (kept in chunk where it logically belongs)
- **Files modified:** internal/usecase/tus_modul_usecase_init_test.go
- **Verification:** `go vet` passes, no duplicate declarations
- **Committed in:** 62ef5cf

**2. [Rule 1 - Bug] Extra closing brace in tus_upload_usecase_chunk_test.go**
- **Found during:** Task 2 (go vet)
- **Issue:** Extra `}` at end of file caused compilation error
- **Fix:** Removed extra closing brace
- **Files modified:** internal/usecase/tus_upload_usecase_chunk_test.go
- **Verification:** `go vet` and `go test` pass
- **Committed in:** 62ef5cf

**3. [Rule 3 - Blocking] Unused imports in split files**
- **Found during:** Task 2 (go vet)
- **Issue:** `bytes`, `fmt`, `time`, `errors`, `apperrors` imports unused after moving functions to other files
- **Fix:** Removed unused imports from tus_upload_usecase_init_test.go, tus_modul_usecase_init_test.go, rbac_helper_setup_test.go, modul_controller_crud_test.go
- **Files modified:** 4 files
- **Verification:** `go vet ./...` passes cleanly
- **Committed in:** 62ef5cf

---

**Total deviations:** 3 auto-fixed (2 bugs, 1 blocking)
**Impact on plan:** All auto-fixes necessary for correct compilation. No scope creep.

## Issues Encountered
None beyond the auto-fixed deviations above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All test files in the project are now under 500 lines
- Ready for Plan 05-08 (final phase plan)

---
*Phase: 05-deep-architecture-improvements*
*Completed: 2026-02-16*

## Self-Check: PASSED
- All 14 split files exist on disk
- All 7 original files deleted
- Both task commits (578e053, 62ef5cf) exist in git log
- SUMMARY.md created successfully
