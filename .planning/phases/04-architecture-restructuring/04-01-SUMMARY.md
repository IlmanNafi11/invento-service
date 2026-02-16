---
phase: 04-architecture-restructuring
plan: 01
subsystem: api
tags: [go-packages, refactoring, clean-architecture, http-utilities]

# Dependency graph
requires: []
provides:
  - "internal/httputil/ package with response helpers, HTTP status constants, pagination, query parsing, cookie helper, validator"
  - "Pattern for extracting focused packages from helper god-package"
affects: [04-02, 04-03, 04-04, 04-05, 04-06]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "httputil leaf package with zero internal dependencies beyond domain/errors/version/config"
    - "Consumer files use httputil.Symbol for HTTP utilities, helper.Symbol for remaining helpers"

key-files:
  created:
    - "internal/httputil/response.go"
    - "internal/httputil/http_status.go"
    - "internal/httputil/pagination.go"
    - "internal/httputil/query_parser.go"
    - "internal/httputil/cookie_helper.go"
    - "internal/httputil/validator.go"
    - "internal/httputil/response_coverage_test.go"
    - "internal/httputil/http_status_test.go"
    - "internal/httputil/pagination_test.go"
    - "internal/httputil/query_parser_coverage_test.go"
    - "internal/httputil/cookie_helper_test.go"
    - "internal/httputil/validator_test.go"
    - "internal/httputil/response_test.go"
  modified:
    - "internal/controller/base/controller.go"
    - "internal/controller/http/auth_controller.go"
    - "internal/controller/http/auth_controller_test.go"
    - "internal/controller/http/health_controller.go"
    - "internal/controller/http/modul_controller.go"
    - "internal/controller/http/project_controller.go"
    - "internal/controller/http/role_controller.go"
    - "internal/controller/http/statistic_controller.go"
    - "internal/controller/http/tus_helpers.go"
    - "internal/controller/http/user_controller.go"
    - "internal/helper/middleware.go"
    - "internal/helper/middleware_test.go"
    - "internal/helper/tus_response.go"
    - "internal/integration_test.go"
    - "internal/middleware/integration_test.go"
    - "internal/middleware/validation.go"
    - "internal/app/server.go"
    - "internal/usecase/role_usecase.go"
    - "internal/usecase/user_usecase.go"

key-decisions:
  - "helper/tus_response.go imports httputil for Send* functions rather than duplicating them"
  - "helper/middleware.go imports httputil for CookieHelper and response helpers"
  - "TUS controllers keep helper import for SendTus* functions, no httputil needed"
  - "Files using only httputil symbols have helper import removed entirely"

patterns-established:
  - "Extract-and-redirect: move files to new package, update package declaration, redirect all consumers"
  - "Dual-import pattern: files needing both httputil and helper symbols import both packages"

# Metrics
duration: 15min
completed: 2026-02-16
---

# Phase 04 Plan 01: Extract httputil Package Summary

**Extracted 6 HTTP utility files (response, status, pagination, query parser, cookie helper, validator) and 7 test files from helper god-package into new internal/httputil/ leaf package with all 20 consumer files updated**

## Performance

- **Duration:** 15 min
- **Started:** 2026-02-16T00:20:13Z
- **Completed:** 2026-02-16T00:35:00Z
- **Tasks:** 2
- **Files modified:** 33 (13 created, 20 modified)

## Accomplishments
- New `internal/httputil/` package with 6 source files and 7 test files, all compiling and passing independently
- All 20 consumer files across controllers, middleware, usecases, and integration tests updated to import from httputil
- `go build ./...`, `go test ./...` (non-network tests), and `go vet ./...` all pass clean
- helper/middleware.go and helper/tus_response.go within the helper package correctly import httputil for extracted symbols

## Task Commits

Each task was committed atomically:

1. **Task 1: Move httputil source and test files to new package** - `a4032d4` (refactor)
2. **Task 2: Update all consumer import paths from helper to httputil** - `8edbe11` (fix)

**Plan metadata:** [pending] (docs: complete plan)

## Files Created/Modified
- `internal/httputil/response.go` - HTTP response helpers (SendSuccess, SendError, SendAppError, etc.)
- `internal/httputil/http_status.go` - HTTP status constants and DefaultMessages map
- `internal/httputil/pagination.go` - PaginationParams, CalculatePagination, CalculateOffset
- `internal/httputil/query_parser.go` - ParsePaginationQuery, ParseSearchQuery, ParseFilterQuery, ParseIntQuery, ParseBoolQuery
- `internal/httputil/cookie_helper.go` - CookieHelper struct for httpOnly auth cookies
- `internal/httputil/validator.go` - ValidateStruct with Indonesian error messages
- `internal/httputil/*_test.go` - All corresponding test files (7 test files)
- `internal/controller/base/controller.go` - Dual import: httputil for Send*/Status*/Validate, helper for CasbinEnforcer
- `internal/helper/middleware.go` - Now imports httputil for CookieHelper and response functions
- `internal/helper/tus_response.go` - Now imports httputil for Send*Response delegations

## Decisions Made
- **helper/tus_response.go delegates to httputil**: TUS-specific response wrappers in helper now call `httputil.Send*Response` instead of removed same-package functions
- **helper/middleware.go imports httputil**: The middleware uses CookieHelper and response functions, both now in httputil
- **TUS controllers don't need httputil**: They only use `helper.SendTus*` functions which remain in helper
- **Files with only httputil symbols drop helper import entirely**: project_controller.go, user_controller.go, all middleware files, integration tests

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed helper/middleware.go internal references**
- **Found during:** Task 2 (consumer import updates)
- **Issue:** helper/middleware.go is within the helper package and called SendErrorResponse, SendUnauthorizedResponse, etc. directly — these were extracted to httputil
- **Fix:** Added `httputil` import to middleware.go and prefixed all extracted function calls with `httputil.`
- **Files modified:** internal/helper/middleware.go
- **Committed in:** 8edbe11

**2. [Rule 3 - Blocking] Fixed helper/tus_response.go internal references**
- **Found during:** Task 2 (consumer import updates)
- **Issue:** tus_response.go delegates to SendBadRequestResponse, SendNotFoundResponse, etc. which were extracted
- **Fix:** Added `httputil` import and prefixed all delegated calls with `httputil.`
- **Files modified:** internal/helper/tus_response.go
- **Committed in:** 8edbe11

**3. [Rule 1 - Bug] Reverted incorrectly changed SendTus* references**
- **Found during:** Task 2 verification
- **Issue:** sed pattern `helper.Send` was too broad — it matched `helper.SendTus*` functions which should remain in helper
- **Fix:** Reverted `httputil.SendTus*` back to `helper.SendTus*` in TUS controller files
- **Files modified:** tus_controller.go, tus_modul_controller.go, tus_helpers.go
- **Committed in:** 8edbe11

**4. [Rule 3 - Blocking] Fixed helper/middleware_test.go references to CookieHelper**
- **Found during:** Task 2 (consumer import updates)
- **Issue:** middleware_test.go used `helper.CookieHelper` and `helper.AccessTokenCookieName` — both now in httputil
- **Fix:** Added httputil import, changed references to `httputil.CookieHelper`, `httputil.NewCookieHelper`, `httputil.AccessTokenCookieName`
- **Files modified:** internal/helper/middleware_test.go
- **Committed in:** 8edbe11

---

**Total deviations:** 4 auto-fixed (1 bug, 3 blocking)
**Impact on plan:** All auto-fixes necessary for compilation correctness. No scope creep.

## Issues Encountered
- sed pattern for replacing `helper.Send*` was too broad and caught `helper.SendTus*` TUS-specific functions — required a targeted revert pass

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- httputil package fully extracted and all consumers updated
- Established the extract-and-redirect pattern for subsequent extractions (plans 02-06)
- helper package is ~35% smaller (6 files removed), making next extractions cleaner

---
*Phase: 04-architecture-restructuring*
*Completed: 2026-02-16*
