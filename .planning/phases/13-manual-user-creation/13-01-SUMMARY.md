---
phase: 13-manual-user-creation
plan: 01
subsystem: user-management
tags: [admin-create-user, supabase-auth, casbin-rbac, dto, usecase]

requires:
  - phase: 12
    provides: "AuthService interface, Supabase Auth integration, email confirmation flow"
provides:
  - "CreateUserRequest/CreateUserResponse DTOs with validation tags"
  - "AdminCreateUser on AuthService interface + Supabase Admin API implementation (email_confirm:false)"
  - "AdminCreateUser usecase with Mahasiswa domain validation, auto-password gen, DB+Casbin sync, rollback"
affects: [13-02-plan]

tech-stack:
  added: ["crypto/rand for password generation", "encoding/base64 for password encoding"]
  patterns: ["Rollback chain: Casbin fail → undo DB + Supabase; DB fail → undo Supabase"]

key-files:
  created: []
  modified:
    - "internal/dto/user.go"
    - "internal/domain/auth.go"
    - "internal/supabase/auth.go"
    - "internal/usecase/user_usecase.go"
    - "internal/app/server.go"
    - "internal/usecase/auth_mocks_test.go"
    - "internal/usecase/auth_usecase_login_test.go"
    - "internal/usecase/auth_integration_test.go"
    - "internal/middleware/auth_test.go"
    - "internal/usecase/user_usecase_profile_test.go"
    - "internal/usecase/user_usecase_files_test.go"
    - "internal/usecase/user_usecase_role_test.go"

key-decisions:
  - "AdminCreateUser sets email_confirm:false — user created as unconfirmed (lazy confirmation on first login)"
  - "Password auto-generation uses crypto/rand (16 bytes) + base64.RawURLEncoding — 22-char URL-safe password"
  - "Mahasiswa domain check uses case-insensitive comparison (strings.EqualFold for role, ToLower for email suffix)"
  - "authService injected into userUsecase (new constructor parameter) — wired via supabaseAuthService in server.go"

patterns-established:
  - "Multi-step creation with ordered rollback: Supabase → DB → Casbin, reverse on failure"

requirements-completed: [AUTH-01, USER-01, USER-02, USER-03, USER-04, USER-05]

duration: 20min
completed: 2026-02-20
---

# Phase 13 Plan 01: Foundation — DTOs, AuthService, Usecase Summary

**CreateUser DTOs, Supabase Admin API integration, and AdminCreateUser usecase with full business logic**

## Performance

- **Duration:** 20 min
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments
- Added CreateUserRequest (email, name, optional password, role_id) and CreateUserResponse DTOs with validation tags
- Added AdminCreateUser to AuthService interface — returns Supabase user ID
- Implemented Supabase Admin API call (POST /admin/users with email_confirm:false) following DeleteUser patterns
- Implemented AdminCreateUser usecase: Mahasiswa domain validation, auto-password generation (crypto/rand 16 bytes), Supabase Auth creation, local DB record, Casbin RBAC sync, rollback on failure
- Wired authService into NewUserUsecase constructor and updated server.go initialization
- Updated all AuthService mock implementations (4 mocks) and all NewUserUsecase test call sites (31 calls)

## Task Commits

Each task was committed atomically:

1. **Task 1: DTOs + AdminCreateUser AuthService interface + Supabase implementation** — `dcefc99` (feat)
2. **Task 2: AdminCreateUser usecase + server.go wiring + test updates** — `c90b127` (feat)

## Files Created/Modified
- `internal/dto/user.go` — Added CreateUserRequest and CreateUserResponse structs
- `internal/domain/auth.go` — Added AdminCreateUser(ctx, email, password) (string, error) to AuthService interface
- `internal/supabase/auth.go` — Implemented AdminCreateUser via POST /admin/users with email_confirm:false
- `internal/usecase/user_usecase.go` — Added authService field, updated constructor, added AdminCreateUser to interface and implemented with full business logic
- `internal/app/server.go` — Updated NewUserUsecase call to pass supabaseAuthService
- `internal/usecase/auth_mocks_test.go` — Added AdminCreateUser to MockAuthService
- `internal/usecase/auth_usecase_login_test.go` — Added AdminCreateUser to AuthUsecaseMockAuthService
- `internal/usecase/auth_integration_test.go` — Added AdminCreateUser to IntegrationMockAuthService
- `internal/middleware/auth_test.go` — Added AdminCreateUser to mockAuthService
- `internal/usecase/user_usecase_profile_test.go` — Updated 11 NewUserUsecase calls with authService param
- `internal/usecase/user_usecase_files_test.go` — Updated 9 NewUserUsecase calls with authService param
- `internal/usecase/user_usecase_role_test.go` — Updated 11 NewUserUsecase calls with authService param

## Decisions Made
- AdminCreateUser sets email_confirm:false so admin-created users are unconfirmed (lazy confirmation triggers on first login, per Phase 12 decision)
- Password auto-generation uses crypto/rand (16 bytes) + base64.RawURLEncoding producing 22-char URL-safe passwords
- Mahasiswa domain check uses case-insensitive comparison (EqualFold for role name, ToLower for email suffix) — consistent with Phase 12 fix (8fb2da1)
- authService injected as new parameter in NewUserUsecase constructor (before casbinEnforcer) — nil passed in existing tests since they don't test AdminCreateUser

## Deviations from Plan
None.

## Issues Encountered
None.

## User Setup Required
None — no external service configuration required.

## Next Plan Readiness
- AdminCreateUser usecase complete — ready for 13-02
- 13-02 will add: HTTP controller handler with Swagger annotations, route registration with RBAC, Swagger docs regeneration

---
*Phase: 13-manual-user-creation*
*Completed: 2026-02-20*
