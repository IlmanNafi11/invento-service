# Roadmap: Invento-Service

## Milestones

- ✅ **v1.0 Comprehensive Refactoring & Optimization** — Phases 1-8 (shipped 2026-02-17) — [archive](milestones/v1.0-ROADMAP.md)
- 🚧 **v1.1 Performance & Security Fixes** — Phases 9-11 (in progress)

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

### 🚧 v1.1 Performance & Security Fixes (In Progress)

**Milestone Goal:** Fix all Supabase RLS performance issues, optimize Go queries, and verify correctness after changes.

- [ ] **Phase 9: RLS Policy Migration** - Fix all 108 Supabase linter warnings via SQL migrations
- [ ] **Phase 10: Go Query Optimization** - Eliminate redundant queries in GORM repositories
- [ ] **Phase 11: Verification & Validation** - Confirm zero linter warnings, passing tests, and correct access control

## Phase Details

### Phase 9: RLS Policy Migration
**Goal**: All Supabase RLS policies use optimal evaluation patterns, eliminating per-row auth function overhead and overlapping permissive policy warnings
**Depends on**: Nothing (first phase of v1.1; v1.0 complete)
**Requirements**: RLS-01, RLS-02, RLS-03
**Success Criteria** (what must be TRUE):
  1. All `auth.uid()`, `auth.jwt()`, and `auth.role()` calls in RLS policies are wrapped in `(SELECT ...)` subqueries, forcing PostgreSQL InitPlan caching (one evaluation per statement, not per row)
  2. Service role policies are scoped with `TO service_role` instead of defaulting to `{public}`, so they no longer appear as duplicate permissive policies alongside user policies
  3. User-facing policies are scoped with `TO authenticated`, preventing unintended access from `anon` or other roles
  4. All policy changes are applied within transactions to avoid DROP/CREATE windows that could cause request failures
**Plans**: 2 plans

Plans:
- [ ] 09-01-PLAN.md — CRUD table RLS migration (projects, moduls, tus_uploads, tus_modul_uploads)
- [ ] 09-02-PLAN.md — Special + admin table RLS migration (user_profiles, roles, permissions, casbin_rule, role_permissions)

### Phase 10: Go Query Optimization
**Goal**: GORM repository queries use minimal database round-trips, replacing N+1 patterns and multi-query operations with efficient single-query alternatives
**Depends on**: Phase 9
**Requirements**: GORM-01, GORM-02, GORM-03
**Success Criteria** (what must be TRUE):
  1. All User queries that load the Role association use `Joins("Role")` instead of `Preload("Role")`, reducing every authenticated request from 2 queries to 1
  2. `GetProfileWithCounts` returns user profile with module and project counts in a single query using LEFT JOIN and subquery counts, instead of 3-4 separate queries
  3. `GetStatistics` returns all dashboard counts in a single query with subquery counts, instead of up to 4 sequential count queries
**Plans**: 2 plans

Plans:
- [ ] 10-01-PLAN.md — User repository query optimization (Preload->Joins + GetProfileWithCounts consolidation)
- [ ] 10-02-PLAN.md — GetStatistics consolidation (single query + test updates)

### Phase 11: Verification & Validation
**Goal**: All changes from Phases 9-10 are verified correct -- zero linter warnings, all tests passing, and access control behaves as expected
**Depends on**: Phase 10
**Requirements**: VER-01, VER-02, VER-03
**Success Criteria** (what must be TRUE):
  1. Supabase database linter reports 0 warnings for both `auth_rls_initplan` and `multiple_permissive_policies` categories (down from 108 total)
  2. All Go tests pass (`go test ./...`) with zero failures after query optimization changes
  3. Service role connections bypass RLS as expected, and authenticated user connections are properly isolated to their own data by the updated policies
**Plans**: TBD

Plans:
- [ ] 11-01: TBD

## Progress

**Execution Order:** Phases execute in numeric order: 9 -> 10 -> 11

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
| 9. RLS Policy Migration | v1.1 | 0/2 | Planned | - |
| 10. Go Query Optimization | v1.1 | 0/2 | Planned | - |
| 11. Verification & Validation | v1.1 | 0/? | Not started | - |
