# Roadmap: Invento-Service

## Milestones

- ✅ **v1.0 Comprehensive Refactoring & Optimization** — Phases 1-8 (shipped 2026-02-17) — [archive](milestones/v1.0-ROADMAP.md)
- ✅ **v1.1 Performance & Security Fixes** — Phases 9-11 (shipped 2026-02-17) — [archive](milestones/v1.1-ROADMAP.md)
- 🚧 **v1.2 User Management & Import** — Phases 12-14 (in progress)

## Phases

<details>
<summary>✅ v1.0 Comprehensive Refactoring & Optimization (Phases 1-8) — SHIPPED 2026-02-17</summary>

- [x] Phase 1: Foundation & Rename (5/5 plans) — completed 2026-02-15
- [x] Phase 2: Memory & Performance Tuning (2/2 plans) — completed 2026-02-16
- [x] Phase 3: Code Quality Standardization (3/3 plans) — completed 2026-02-16
- [x] Phase 4: Architecture Restructuring (6/6 plans) — completed 2026-02-16
- [x] Phase 5: Deep Architecture Improvements (10/10 plans) — completed 2026-02-16
- [x] Phase 6: Polish & Verification (6/6 plans) — completed 2026-02-16
- [x] Phase 7: Swagger & Logger Integration Fixes (2/2 plans) — completed 2026-02-17
- [x] Phase 8: File Size Enforcement & Verification (3/3 plans) — completed 2026-02-17

</details>

<details>
<summary>✅ v1.1 Performance & Security Fixes (Phases 9-11) — SHIPPED 2026-02-17</summary>

- [x] Phase 9: RLS Policy Migration (2/2 plans) — completed 2026-02-17
- [x] Phase 10: Go Query Optimization (2/2 plans) — completed 2026-02-17
- [x] Phase 11: Verification & Validation (1/1 plan) — completed 2026-02-17

</details>

### 🚧 v1.2 User Management & Import (In Progress)

**Milestone Goal:** Add user management features (manual + bulk Excel import) with email confirmation flow that differentiates admin-created vs self-registered users.

- [x] **Phase 12: Auth Confirmation Flow** - Change login to reject unconfirmed users and register to require email confirmation
- [ ] **Phase 13: Manual User Creation** - Add endpoint for authorized users to create individual accounts with role assignment
- [ ] **Phase 14: Excel Import & Template** - Bulk user import from Excel with validation, skip logic, and detailed reporting

## Phase Details

### Phase 12: Auth Confirmation Flow
**Goal**: Login rejects unconfirmed users (triggering confirmation resend), and self-registration requires email confirmation before access
**Depends on**: Nothing (modifies existing auth behavior)
**Requirements**: AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06
**Success Criteria** (what must be TRUE):
  1. Self-registered user receives confirmation email immediately after signup and cannot login until confirmed
  2. Self-registered user is automatically assigned Mahasiswa role
  3. Self-registration requires @student.polije.ac.id email domain (Mahasiswa role)
  4. Unconfirmed user attempting login is rejected with clear Indonesian message indicating confirmation email has been sent
  5. First login attempt by unconfirmed user triggers Supabase resend API to send confirmation email
**Plans:** 2 plans
Plans:
- [x] 12-01-PLAN.md — Error codes + Domain/Service layer (ResendConfirmation, Register autoConfirm)
- [x] 12-02-PLAN.md — Usecase + Controller (student-only registration, login confirmation detection)

### Phase 13: Manual User Creation
**Goal**: Authorized users can create individual user accounts with role assignment; created users are unconfirmed until they confirm via email
**Depends on**: Phase 12
**Requirements**: AUTH-01, USER-01, USER-02, USER-03, USER-04, USER-05
**Success Criteria** (what must be TRUE):
  1. Authorized user (with correct RBAC permission) can create a new user by providing email, name, optional password, and role
  2. When password is omitted, system generates a secure random password and returns it in the API response
  3. Created user exists as unconfirmed in Supabase Auth with no confirmation email sent at creation time
  4. Created user has local DB record with assigned role and Casbin RBAC policy synchronized
  5. Creating user with Mahasiswa role requires @student.polije.ac.id email domain — request rejected with clear error if domain doesn't match
**Plans**: 2 plans
Plans:
- [x] 13-01-PLAN.md — Foundation: DTOs, AuthService.AdminCreateUser, Supabase implementation, Usecase business logic (validation, password gen, DB+Casbin sync, rollback)
- [ ] 13-02-PLAN.md — HTTP Layer: CreateUser controller handler with Swagger annotations, route registration with RBAC, Swagger docs regeneration

### Phase 14: Excel Import & Template
**Goal**: Authorized users can bulk import users from Excel with comprehensive validation, skip logic, and detailed reporting
**Depends on**: Phase 13
**Requirements**: IMPORT-01, IMPORT-02, IMPORT-03, IMPORT-04, IMPORT-05, IMPORT-06, IMPORT-07, IMPORT-08, IMPORT-09, IMPORT-10, IMPORT-11, USER-05
**Success Criteria** (what must be TRUE):
  1. Authorized user can download an Excel template with correct headers (Email, Nama, Password, Jenis Kelamin, Role), format guide, and example data
  2. Authorized user can upload .xlsx file and system processes all rows synchronously, creating unconfirmed users
  3. Rows with invalid emails, missing required fields (Email/Nama), or duplicate emails (in DB or within batch) are silently skipped
  4. Rows with Mahasiswa role but non-@student.polije.ac.id email domain are skipped with reason in report
  5. Empty Password fields get auto-generated passwords; empty Role fields use the default role from API request parameter
  6. Response includes detailed report: total rows, success count, skipped count, per-skip details (row number + reason), and generated passwords for auto-generated users
**Plans**: 2 plans

Plans:
- [ ] 14-01-PLAN.md — Excel template generation, import DTOs, and template download endpoint
- [ ] 14-02-PLAN.md — Bulk import logic with validation, skip logic, and upload endpoint

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation & Rename | v1.0 | 5/5 | Complete | 2026-02-15 |
| 2. Memory & Performance Tuning | v1.0 | 2/2 | Complete | 2026-02-16 |
| 3. Code Quality Standardization | v1.0 | 3/3 | Complete | 2026-02-16 |
| 4. Architecture Restructuring | v1.0 | 6/6 | Complete | 2026-02-16 |
| 5. Deep Architecture Improvements | v1.0 | 10/10 | Complete | 2026-02-16 |
| 6. Polish & Verification | v1.0 | 6/6 | Complete | 2026-02-16 |
| 7. Swagger & Logger Integration Fixes | v1.0 | 2/2 | Complete | 2026-02-17 |
| 8. File Size Enforcement & Verification | v1.0 | 3/3 | Complete | 2026-02-17 |
| 9. RLS Policy Migration | v1.1 | 2/2 | Complete | 2026-02-17 |
| 10. Go Query Optimization | v1.1 | 2/2 | Complete | 2026-02-17 |
| 11. Verification & Validation | v1.1 | 1/1 | Complete | 2026-02-17 |
| 12. Auth Confirmation Flow | v1.2 | 2/2 | Complete | 2026-02-20 |
| 13. Manual User Creation | v1.2 | 1/2 | In progress | - |
| 14. Excel Import & Template | v1.2 | 0/2 | Planning complete | - |
