# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-20)

**Core value:** File storage yang reliable dan resource-efficient pada server terbatas (500MB RAM) -- upload, simpan, dan download file modul/project mahasiswa tanpa gagal.
**Current focus:** v1.2.1 milestone ARCHIVED — ready for next milestone via /gsd-new-milestone

## Current Position

Phase: 15 of 15 (Tech Debt Cleanup) — COMPLETE
Plan: 15-02 complete (2/2)
Status: v1.2.1 tech debt cleanup complete
Last activity: 2026-02-20 — Completed plan 15-02: 10 test cases for AdminCreateUser & BulkImportUsers

Progress: [██████████] 100% (v1.2.1 — 2/2 plans)

## Performance Metrics

**Velocity:**
- v1.0 plans completed: 37 (Phases 1-8)
- v1.1 plans completed: 5 (Phases 9-11)
- v1.2 plans completed: 6 (Phases 12-14)
- v1.2.1 plans completed: 2 (Phase 15)
- Total plans: 50

## Accumulated Context

### Decisions

All v1.0 and v1.1 decisions archived in PROJECT.md Key Decisions table.

v1.2 decisions:
- Lazy email confirmation for admin-created users (send on first login, not at creation)
- Synchronous Excel import processing (no async/background jobs)
- Supabase resend API for confirmation emails (no custom email service)
- Mahasiswa role requires @student.polije.ac.id domain (manual, import, and self-registration)
- Self-registration: email confirmation required before login (no auto-confirm for anyone)
- ErrEmailNotConfirmed returns HTTP 403 with dedicated code (distinct from generic FORBIDDEN_ERROR)
- AutoConfirm defaults to false — callers must explicitly opt-in to auto-confirm
- RegisterResult in domain package (not usecase) so controller can reference without import cycles
- Register returns 200 (not 201) since user must confirm email before resource is usable
- Teacher email restriction at Register usecase level (not validatePolijeEmail) to preserve reuse for admin-created users
- ExcelHelper in new `internal/helper/` package (separate from `internal/httputil/`) for Excel-specific operations

v1.2.1 decisions:
- createSingleUser helper does NOT call SavePolicy — callers decide (per-user in AdminCreateUser vs batch in BulkImportUsers)
- Helper returns actual password used so callers decide what to expose
- Reuse existing MockAuthService from auth_mocks_test.go for admin/import tests (no duplicate mocks)
- Pass nil casbinEnforcer in usecase tests since constructor takes concrete struct pointer

### Pending Todos

None.

### Blockers/Concerns

- TUS upload memory test requires authenticated JWT tokens -- procedure documented, pending staging environment
- SQLite vs PostgreSQL test divergence: GORM query optimizations may need dual-database testing

### Quick Tasks Completed

| # | Description | Date | Commit | Directory |
|---|-------------|------|--------|-----------|
| 002 | Fix 500 error on GET /api/v1/user/:id/files endpoint | 2026-02-19 | 84739d5 | [002-fix-500-error-user-files-endpoint](./quick/002-fix-500-error-user-files-endpoint/) |
| 003 | Fix 500 error on POST /api/v1/user (tambah user) — wrap raw authService errors as AppError | 2026-02-20 | TBD | [3-fix-500-error-on-post-api-v1-user-tambah](./quick/3-fix-500-error-on-post-api-v1-user-tambah/) |

## Session Continuity

Last session: 2026-02-20
Stopped at: Plan 15-02 complete — Phase 15 and v1.2.1 milestone complete
Next action: `/gsd-new-milestone` — start next milestone (questioning → research → requirements → roadmap)
