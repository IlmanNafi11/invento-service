---
phase: 06-polish-verification
plan: 03
subsystem: api-documentation
tags: [swagger, openapi, tus, annotations]
dependency_graph:
  requires: []
  provides: [complete-swagger-coverage]
  affects: [docs/swagger.json, docs/swagger.yaml]
tech_stack:
  added: []
  patterns: [swagger-annotations, tus-protocol-headers]
key_files:
  created: []
  modified:
    - internal/controller/http/tus_controller.go
    - internal/controller/http/tus_modul_controller.go
    - internal/controller/http/user_controller.go
decisions:
  - "TUS Upload tag for project endpoints, TUS Modul Upload tag for modul endpoints"
  - "Role Management tag for GetUsersForRole and BulkAssignRole (matches route grouping)"
  - "Modul IDs use string type in @Param (UUID), project IDs use int"
metrics:
  duration: ~12min
  completed: 2026-02-16
---

# Phase 06 Plan 03: Add Missing Swagger Annotations Summary

**One-liner:** Complete Swagger annotation coverage for 25 previously undocumented TUS upload and role management endpoints across 3 controller files.

## What Was Done

### Task 1: TUS Controller Swagger Annotations (44fc421)

Added complete Swagger annotations to all 23 handler methods across both TUS controllers:

**tus_controller.go (12 methods):**
- `CheckUploadSlot` — GET `/project/upload/check-slot`
- `ResetUploadQueue` — POST `/project/upload/reset-queue`
- `InitiateUpload` — POST `/project/upload/`
- `InitiateProjectUpdateUpload` — POST `/project/{id}/upload`
- `UploadChunk` — PATCH `/project/upload/{id}`
- `UploadProjectUpdateChunk` — PATCH `/project/{id}/update/{upload_id}`
- `GetUploadStatus` — HEAD `/project/upload/{id}`
- `GetProjectUpdateUploadStatus` — HEAD `/project/{id}/update/{upload_id}`
- `GetUploadInfo` — GET `/project/upload/{id}`
- `GetProjectUpdateUploadInfo` — GET `/project/{id}/update/{upload_id}`
- `CancelUpload` — DELETE `/project/upload/{id}`
- `CancelProjectUpdateUpload` — DELETE `/project/{id}/update/{upload_id}`

**tus_modul_controller.go (11 methods):**
- `CheckUploadSlot` — GET `/modul/upload/check-slot`
- `InitiateUpload` — POST `/modul/upload/`
- `InitiateModulUpdateUpload` — POST `/modul/{id}/upload`
- `UploadChunk` — PATCH `/modul/upload/{upload_id}`
- `UploadModulUpdateChunk` — PATCH `/modul/{id}/update/{upload_id}`
- `GetUploadStatus` — HEAD `/modul/upload/{upload_id}`
- `GetModulUpdateUploadStatus` — HEAD `/modul/{id}/update/{upload_id}`
- `GetUploadInfo` — GET `/modul/upload/{upload_id}`
- `GetModulUpdateUploadInfo` — GET `/modul/{id}/update/{upload_id}`
- `CancelUpload` — DELETE `/modul/upload/{upload_id}`
- `CancelModulUpdateUpload` — DELETE `/modul/{id}/update/{upload_id}`

All TUS annotations include:
- `@Param Tus-Resumable header` for protocol version
- `@Header` annotations for TUS response headers (Upload-Offset, Upload-Length, Location, Tus-Resumable)
- `@Success 204` for PATCH (chunk) and DELETE (cancel) operations
- `@Accept application/offset+octet-stream` for chunk upload endpoints
- Proper `@Security BearerAuth` on all endpoints

### Task 2: User Controller Role Management Annotations (af68ae7)

Added Swagger annotations to 2 previously undocumented endpoints:

- `GetUsersForRole` — GET `/role/{id}/users` — Returns user list for a specific role
- `BulkAssignRole` — POST `/role/{id}/users/bulk` — Bulk assign role to multiple users

Both tagged as `Role Management` to match their route grouping under `/api/v1/role`.

## Verification Results

- `go build ./...` — PASSED
- `@Router` count in tus_controller.go: **12** ✅
- `@Router` count in tus_modul_controller.go: **11** ✅
- `@Router` count in user_controller.go: **10** (8 existing + 2 new) ✅
- Total `@Router` annotations across all controllers: **58** ✅

## Deviations from Plan

### Minor Deviation: Modul TUS endpoint count

The plan estimated 8 modul TUS endpoints, but the actual controller has 11 public handler methods (including `GetUploadInfo`, `GetModulUpdateUploadInfo`, `GetUploadStatus`, `GetModulUpdateUploadStatus` which are separate from their project counterparts). All 11 were annotated.

## Commits

| Hash | Message |
|------|---------|
| 44fc421 | feat(06-03): add Swagger annotations to TUS upload controllers |
| af68ae7 | feat(06-03): add Swagger annotations to role management endpoints |

## Self-Check: PASSED

- All 3 modified files exist on disk
- Both commits (44fc421, af68ae7) verified in git history
- @Router counts match expectations (12, 11, 10)
