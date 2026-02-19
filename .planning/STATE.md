# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-19)

**Core value:** File storage yang reliable dan resource-efficient pada server terbatas (500MB RAM) -- upload, simpan, dan download file modul/project mahasiswa tanpa gagal.
**Current focus:** v1.2 User Management & Import — Phase 13: Manual User Creation (IN PROGRESS)

## Current Position

Phase: 13 of 14 (Manual User Creation) — IN PROGRESS
Plan: 13-01 complete, 13-02 pending
Status: Plan 13-01 complete (foundation: DTOs, AuthService, Usecase), ready for 13-02 (HTTP layer)
Last activity: 2026-02-20 — Completed plan 13-01: CreateUser DTOs, AdminCreateUser on AuthService (Supabase Admin API with email_confirm:false), AdminCreateUser usecase with Mahasiswa domain validation, auto-password generation, DB+Casbin sync, rollback chain

Progress: [████░░░░░░] 50% (v1.2 — 3/6 plans)

## Performance Metrics

**Velocity:**
- v1.0 plans completed: 37 (Phases 1-8)
- v1.1 plans completed: 5 (Phases 9-11)
- v1.2 plans completed: 3 (Phases 12-13)
- Total plans: 45

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

### Pending Todos

None.

### Blockers/Concerns

- TUS upload memory test requires authenticated JWT tokens -- procedure documented, pending staging environment
- SQLite vs PostgreSQL test divergence: GORM query optimizations may need dual-database testing

### Quick Tasks Completed

| # | Description | Date | Commit | Directory |
|---|-------------|------|--------|-----------|
| 002 | Fix 500 error on GET /api/v1/user/:id/files endpoint | 2026-02-19 | 84739d5 | [002-fix-500-error-user-files-endpoint](./quick/002-fix-500-error-user-files-endpoint/) |

## Session Continuity

Last session: 2026-02-20
Stopped at: Phase 13, plan 13-01 complete
Next action: Plan and execute 13-02 (HTTP controller, route registration, Swagger docs)
