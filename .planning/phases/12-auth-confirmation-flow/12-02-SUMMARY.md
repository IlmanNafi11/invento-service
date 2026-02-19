---
phase: 12-auth-confirmation-flow
plan: 02
subsystem: auth
tags: [supabase-auth, email-confirmation, registration, swagger, controller]

requires:
  - phase: 12-01
    provides: "ErrEmailNotConfirmed error code, ResendConfirmation method, AutoConfirm field"
provides:
  - "Student-only self-registration with email confirmation (no auto-confirm)"
  - "Login detection of unconfirmed users with resend trigger"
  - "RegisterMessageResponse DTO for message-only registration response"
  - "Updated Swagger docs reflecting confirmation flow"
affects: [13-manual-user-creation]

tech-stack:
  added: []
  patterns: ["RegisterResult domain type for multi-outcome usecase returns"]

key-files:
  created: []
  modified:
    - "internal/usecase/auth_usecase.go"
    - "internal/controller/http/auth_controller.go"
    - "internal/dto/auth.go"
    - "internal/domain/auth.go"
    - "docs/swagger.json"

key-decisions:
  - "RegisterResult in domain package — avoids import cycles for controller access"
  - "Register returns 200 (not 201) — user cannot use account until email confirmed"
  - "Login resend failure logged but not surfaced to user — they still get unconfirmed error"
  - "Teacher emails rejected at usecase level, not validatePolijeEmail — preserves future admin-created dosen flow"

patterns-established:
  - "Multi-outcome usecase return: domain struct with NeedsX bool + Message for flow control"

requirements-completed: [AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06]

duration: 15min
completed: 2026-02-20
---

# Phase 12 Plan 02: Usecase + Controller Summary

**Student-only self-registration with email confirmation, unconfirmed-user login detection with automatic resend, updated Swagger annotations**

## Performance

- **Duration:** 15 min
- **Started:** 2026-02-20T01:25:00Z
- **Completed:** 2026-02-20T01:40:00Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments
- Restricted self-registration to @student.polije.ac.id (Mahasiswa role auto-assigned)
- Registration returns message-only 200 response (no tokens/cookies — confirmation required)
- Login detects unconfirmed users, triggers Supabase resend, returns 403 with Indonesian message
- Swagger docs updated: Register 200 with RegisterMessageResponse, Login 403 for unconfirmed

## Task Commits

Each task was committed atomically:

1. **Task 1: Auth usecase — student-only registration, login confirmation detection** - `b48c443` (feat)
2. **Task 2: DTO, controller, Swagger, tests** - `458e70e` (feat)

## Files Created/Modified
- `internal/domain/auth.go` — Added RegisterResult struct
- `internal/usecase/auth_usecase.go` — Updated Register (student-only, AutoConfirm=false, returns RegisterResult), Login (detects unconfirmed + resend)
- `internal/dto/auth.go` — Added RegisterMessageResponse DTO
- `internal/controller/http/auth_controller.go` — Updated Register handler (200 + message, no cookies), Swagger annotations
- `internal/controller/http/auth_controller_test.go` — Updated mock and tests for new Register behavior
- `internal/usecase/auth_usecase_login_test.go` — Added teacher rejection and unconfirmed login tests
- `internal/usecase/auth_usecase_token_test.go` — Updated Register edge case tests
- `internal/usecase/auth_integration_test.go` — Updated integration Register tests
- `docs/docs.go` — Regenerated
- `docs/swagger.json` — Regenerated
- `docs/swagger.yaml` — Regenerated

## Decisions Made
- RegisterResult placed in domain package (not usecase) so controller can reference it without import cycles
- Register returns 200 (not 201) since no resource is immediately usable — user must confirm email first
- Login resend failure is logged but not surfaced to user — they still get the unconfirmed error message
- Teacher emails (@teacher.polije.ac.id) explicitly rejected at usecase level, not at validatePolijeEmail level (preserving it for future admin-created dosen users)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] User controller ParsePathID → ParsePathUUID**
- **Found during:** Task 2 (Swagger regeneration)
- **Issue:** User controller endpoints used ParsePathID (integer) but user IDs are UUIDs — Swagger showed `type: integer` for user ID params
- **Fix:** Changed to ParsePathUUID, updated Swagger annotations and all related tests
- **Files modified:** internal/controller/http/user_controller.go, user_controller_list_test.go, user_controller_role_test.go, internal/app/server_auth_test.go
- **Verification:** All tests pass, Swagger correctly shows `type: string` for user ID params
- **Committed in:** 458e70e (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Fix was necessary for correct Swagger documentation and consistent UUID handling. No scope creep.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Auth confirmation flow complete — Phase 12 done
- Ready for Phase 13: Manual User Creation (admin-created users with role assignment)
- Foundation ready: RegisterResult pattern, ErrEmailNotConfirmed handling, ResendConfirmation method all available for reuse

---
*Phase: 12-auth-confirmation-flow*
*Completed: 2026-02-20*
