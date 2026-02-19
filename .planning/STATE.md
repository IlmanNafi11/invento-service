# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-19)

**Core value:** File storage yang reliable dan resource-efficient pada server terbatas (500MB RAM) -- upload, simpan, dan download file modul/project mahasiswa tanpa gagal.
**Current focus:** v1.2 User Management & Import — Phase 12: Auth Confirmation Flow (COMPLETE)

## Current Position

Phase: 12 of 14 (Auth Confirmation Flow) — COMPLETE
Plan: 12-01 complete, 12-02 complete
Status: Phase 12 complete, ready for Phase 13
Last activity: 2026-02-20 — Completed plan 12-02 (usecase + controller — student-only registration, login confirmation detection)

Progress: [███░░░░░░░] 33% (v1.2 — 2/6 plans)

## Performance Metrics

**Velocity:**
- v1.0 plans completed: 37 (Phases 1-8)
- v1.1 plans completed: 5 (Phases 9-11)
- v1.2 plans completed: 2 (Phase 12)
- Total plans: 44

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
Stopped at: Completed phase 12 (auth confirmation flow — both plans)
Next action: Plan and execute Phase 13 (manual user creation)
