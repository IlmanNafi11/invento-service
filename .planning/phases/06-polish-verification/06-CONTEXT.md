# Phase 6: Polish & Verification - Context

**Gathered:** 2026-02-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Remove dead code, verify memory stays within budget under realistic load, run full lint suite to zero warnings, update Swagger annotations for all endpoints, and update .env.example with all config options from Phases 1-5. Audit test coverage for untested critical paths and add missing tests. Dockerfile changes are deferred.

</domain>

<decisions>
## Implementation Decisions

### Dead code cleanup
- Remove ALL commented-out code blocks with no exceptions
- Trust the linter: delete any function/type/variable that golangci-lint reports as unused
- Audit for remaining hardcoded magic strings beyond Phase 1 work and replace with constants
- Remove ALL TODO/FIXME/HACK comments -- track future work in the roadmap, not in code

### Memory load testing
- Approach: Claude's discretion (scripted or manual -- whatever fits best)
- Simulate 5 concurrent uploads
- Pass/fail threshold: heap must stay under 350MB (matches GOMEMLIMIT=350MiB)
- Use small files only (1-10MB) to test concurrency without saturating disk

### Lint & quality gates
- Enable comprehensive linter set: deadcode, unused, errcheck, govet, staticcheck, gosimple, ineffassign, typecheck, gocritic, gofumpt
- Zero-warning policy: fix ALL warnings, not just errors
- One-time pass only -- no CI enforcement setup needed
- Run gofumpt across the entire codebase for final formatting consistency

### Swagger & docs
- Full audit of all endpoints: add missing annotations, regenerate with swag init
- Update .env.example with every config option added across Phases 1-5
- Dockerfile: skip for now (deferred per user request)
- Test coverage: audit untested critical paths and add missing tests

### Claude's Discretion
- Memory load testing approach (scripted vs manual checklist)
- Exact golangci-lint configuration details beyond the specified linter list
- How to structure the test coverage audit and which paths to prioritize

</decisions>

<specifics>
## Specific Ideas

No specific requirements -- open to standard approaches

</specifics>

<deferred>
## Deferred Ideas

- Dockerfile optimization (multi-stage build, Alpine base, image size audit) -- deferred per user decision

</deferred>

---

*Phase: 06-polish-verification*
*Context gathered: 2026-02-16*
