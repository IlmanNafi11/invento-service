# Deferred Items - Phase 05

## Pre-existing Issues Discovered During 05-05 Execution

1. **Missing `dto` import in test files** - Multiple test files reference `dto.` types without importing `"invento-service/internal/dto"`. Affected files:
   - `internal/controller/http/auth_controller_test.go`
   - `internal/controller/http/health_controller_test.go`
   - `internal/controller/http/statistic_controller_test.go`
   - `internal/usecase/statistic_usecase_test.go`
   - `internal/httputil/validator_test.go`
   - `internal/middleware/validation_test.go`

   **Origin:** Phase 05-02 (DTO migration) moved types to `dto` package but some test files were not updated with the new import.

2. **Auth mock missing context.Context** - `internal/usecase/auth_usecase_test.go` mock types don't implement updated interfaces with `context.Context` from Phase 05-03/05-04.

3. **Modul repository coverage test missing context.Context** - `internal/usecase/repo/modul_repository_coverage_test.go` calls repo methods without `ctx` parameter. Origin: Phase 05-04 context propagation.

4. **User controller test left uncommitted** - `internal/controller/http/user_controller_test.go` has pending changes from prior plan context propagation.

5. **Auth integration test left uncommitted** - `internal/usecase/auth_integration_test.go` has pending changes from prior plan context propagation.
