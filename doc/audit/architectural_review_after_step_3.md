# Architectural Review — After Step 3

**Date:** 2026-04-03
**Scope:** Full codebase review at the end of Step 3 (code generation complete; business logic, handler, server, and infrastructure not yet implemented).

## Current State

The repository contains:
- `proto/amortization/v1/amortization.proto` — service and message definitions
- `gen/amortization/v1/` — generated protobuf Go code (`amortization.pb.go`)
- `gen/amortization/v1/amortizationv1connect/` — generated ConnectRPC code (`amortization.connect.go`)
- `buf.yaml`, `buf.gen.yaml` — Buf configuration
- `go.mod` / `go.sum` — Go module definition
- `doc/` — specification and audit trail

No hand-written Go code exists yet. All `.go` files are machine-generated. The review therefore focuses on proto design, project structure decisions, and forward-looking concerns for the upcoming implementation steps.

---

## Findings

### Interface Design

- **Issue**: The generated `AmortizationServiceHandler` interface (in `amortizationv1connect`) is a single-method interface, which is ideal. However, the spec plans for `internal/amortization/handler.go` to implement this interface directly, coupling business logic to the ConnectRPC transport layer. There is no planned domain-level interface (e.g., `Calculator` or `ScheduleGenerator`) that the handler would depend on.
- **Location**: `doc/spec.md` (Step 3 and Step 4 design), `gen/amortization/v1/amortizationv1connect/amortization.connect.go`
- **Recommendation**: Define a small interface at the handler (consumer) side for the calculator dependency, e.g.:
  ```go
  // in internal/amortization/handler.go
  type Calculator interface {
      Calculate(params CalculateParams) (*Schedule, error)
  }
  ```
  This keeps the handler testable in isolation (mock the calculator) and the calculator testable without proto types.

- **Issue**: The spec describes validation as part of the calculator (`calculator.go`). Validation is a boundary concern and should be separable from pure computation.
- **Location**: `doc/spec.md` (Step 3)
- **Recommendation**: Either expose validation as a separate exported function (`Validate(params) error`) or have the handler perform validation before calling the calculator. This improves testability: validation rules can be tested independently of schedule generation.

### Idiomatic Go

- **Issue**: The `interest` field in the proto is named ambiguously. In Go, the generated accessor is `GetInterest()` which does not convey that it is an annual percentage rate. The spec clarifies "6.5 = 6.5%" but the proto/code does not.
- **Location**: `proto/amortization/v1/amortization.proto` line 33
- **Recommendation**: Rename to `annual_interest_rate` or `interest_rate_percent` in the proto. This makes generated Go code self-documenting (`GetAnnualInterestRate()`). If renaming the proto field is too late, add a clear comment and use a well-named internal domain type.

- **Issue**: The spec uses `float64` (`double` in proto) for all monetary values. Floating-point arithmetic causes rounding errors in financial calculations (e.g., `0.1 + 0.2 != 0.3`).
- **Location**: `proto/amortization/v1/amortization.proto` (all `double` fields: `principal`, `amount`, `payment_amount`, etc.)
- **Recommendation**: For the proto wire format, `double` is acceptable since proto3 lacks a decimal type. Internally, the calculator should work in integer cents (or use a decimal library like `shopspring/decimal`) and convert at the boundary. The spec mentions "round all monetary values to 2 decimal places" which partially addresses this, but intermediate rounding can still accumulate errors over 360 iterations.

- **Issue**: No `context.Context` propagation is mentioned in the spec's calculator design. The calculator is a pure computation, so context is not strictly needed now, but if it ever needs to be cancellable (large schedules, streaming) the signature should accept context from the start.
- **Location**: `doc/spec.md` (Step 3)
- **Recommendation**: Accept `context.Context` as the first parameter of the `Calculate` function, even if it is not used initially. This is idiomatic Go and avoids a breaking API change later.

### Consistent Naming

- **Issue**: The repository directory is named `AmmortizationCalculatorAPI` (with double 'm'), while the Go module path uses `amortization-calculator-api` (correct spelling). The proto package is `amortization.v1` (correct). This inconsistency could cause confusion.
- **Location**: Repository root directory vs. `go.mod` line 1
- **Recommendation**: Rename the repository directory to match the module path (`amortization-calculator-api`), or at minimum document the discrepancy. The Go module path is correct and should not change.

- **Issue**: The spec places all business logic in a single package `internal/amortization/` with files `calculator.go`, `calculator_test.go`, and `handler.go`. Mixing the ConnectRPC handler (transport) with the calculator (domain) in the same package reduces separation of concerns.
- **Location**: `doc/spec.md` (Steps 3-4)
- **Recommendation**: Consider `internal/calculator/` for domain logic and `internal/handler/` (or `internal/server/`) for ConnectRPC glue. Alternatively, keep them in one package but ensure they communicate through an interface, not direct function calls. For a project this size, a single package is acceptable if the interface boundary is clean.

### Safety

- **Issue**: The `SchedulePayment` message lacks a `remaining_balance` field. Without it, clients cannot verify the schedule's correctness or display balance progression. The schedule loop in the spec tracks remaining principal internally but does not expose it.
- **Location**: `proto/amortization/v1/amortization.proto` lines 16-23
- **Recommendation**: Add `double remaining_balance = 7;` to `SchedulePayment`. This is a non-breaking proto change (additive field).

- **Issue**: The `term` field is `int32` with no documented unit. It is months per the spec, but the proto does not say so. A caller might pass years.
- **Location**: `proto/amortization/v1/amortization.proto` line 34
- **Recommendation**: Rename to `term_months` or add a proto comment clarifying the unit.

- **Issue**: The spec validates `paymentDayOfMonth` in range `[1, 28]` to avoid month-length issues, which is correct. However, no validation is specified for the `term` field upper bound. A request with `term = 2147483647` (max int32) would attempt to allocate a massive schedule slice.
- **Location**: `doc/spec.md` (Step 3 validation rules)
- **Recommendation**: Add an upper bound on `term` (e.g., 600 months / 50 years) to prevent resource exhaustion. Pre-allocate the schedule slice with `make([]SchedulePayment, 0, term)` only after validation.

- **Issue**: Additional payments with dates outside the loan period are not addressed in the spec. A payment dated before `start_date` or after the final scheduled payment would be silently ignored or cause unexpected behavior.
- **Location**: `doc/spec.md` (Step 3)
- **Recommendation**: Either validate that additional payment dates fall within the loan period, or document that out-of-range payments are ignored.

### Testability

- **Issue**: The spec's test plan covers the calculator but does not mention handler tests or integration tests. The handler (Step 4) will contain proto-to-domain mapping logic that can harbor bugs (e.g., nil `StartDate` causing a panic).
- **Location**: `doc/spec.md` (Steps 3-4)
- **Recommendation**: Plan handler unit tests that verify: (a) validation errors map to `connect.CodeInvalidArgument`, (b) nil/zero proto fields are handled gracefully, (c) the response proto is correctly populated. Use a mock `Calculator` interface.

- **Issue**: The generated `UnimplementedAmortizationServiceHandler` is provided for forward compatibility, which is good. The handler implementation should embed it to satisfy the interface even if new RPCs are added later.
- **Location**: `gen/amortization/v1/amortizationv1connect/amortization.connect.go` lines 106-110
- **Recommendation**: When implementing the handler in Step 4, embed `amortizationv1connect.UnimplementedAmortizationServiceHandler` as a safety net, though with a single-RPC service this is low risk.

---

## Overall Assessment

The project is well-structured for its early stage. The proto definition is clean, follows buf/ConnectRPC best practices (versioned package, single-purpose RPC, request/response wrappers), and the generated code is minimal and correct. The Go module path and buf configuration are properly set up. The most significant forward-looking concerns are: (1) the absence of a domain-level interface between the handler and calculator, which will limit testability if not introduced in Step 4; (2) the use of `double` for monetary values without an explicit internal decimal strategy, which risks accumulating rounding errors over long loan terms; and (3) missing safety bounds on `term` that could allow resource exhaustion. The repository directory name typo (`Ammortization` vs `Amortization`) is a minor but persistent source of confusion. Overall, the foundation is solid and the spec is thorough — addressing the items above during Steps 3-4 implementation will yield a clean, testable, and safe service.
