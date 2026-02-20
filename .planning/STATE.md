# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-19)

**Core value:** File storage yang reliable dan resource-efficient pada server terbatas (500MB RAM) -- upload, simpan, dan download file modul/project mahasiswa tanpa gagal.
**Current focus:** v1.2 User Management & Import — SHIPPED

## Current Position

Phase: 14 of 14 (Excel Import & Template) — COMPLETE
Plan: 14-02 complete (2/2)
Status: v1.2 milestone shipped — all phases (12-14) complete
Last activity: 2026-02-20 — Completed plan 14-02: BulkImportUsers business logic, FindByEmails batch query, ParseImportFile, ImportUsers HTTP handler, POST /user/import endpoint with RBAC

Progress: [██████████] 100% (v1.2 — 6/6 plans)

## Performance Metrics

**Velocity:**
- v1.0 plans completed: 37 (Phases 1-8)
- v1.1 plans completed: 5 (Phases 9-11)
- v1.2 plans completed: 6 (Phases 12-14)
- Total plans: 48

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
Stopped at: v1.2 milestone shipped — all 14 phases complete (48 plans total)
Next action: None — v1.2 complete. Begin v1.3 planning if needed.
