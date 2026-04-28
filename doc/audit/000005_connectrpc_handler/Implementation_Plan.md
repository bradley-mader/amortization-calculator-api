# Implementation Plan — 000005_connectrpc_handler

**Date**: 2026-04-03
**High-level step**: Implement the ConnectRPC service handler mapping proto types to internal types

## Prerequisites
- `internal/amortization/calculator.go` exists with `CalculationInput`, `CalculationResult`, `ScheduleEntry`, `AdditionalPayment` types, and `Calculate` function
- Generated proto code exists in `gen/amortization/v1/` with `GetAmortizationScheduleRequest`, `GetAmortizationScheduleResponse`, `AmortizationSchedule`, `SchedulePayment`, `Payment` types
- Generated connect code exists in `gen/amortization/v1/amortizationv1connect/` with `AmortizationServiceHandler` interface
- `connectrpc.com/connect v1.19.1` already in `go.mod`
- `google.golang.org/protobuf` already in `go.mod` (provides `timestamppb`)

## Sub-tasks

### 1. Create `internal/amortization/handler.go`
**File**: `internal/amortization/handler.go`
**Action**: Create
**Details**:
- Package: `amortization`
- Imports:
  - `context`
  - `connectrpc.com/connect`
  - `google.golang.org/protobuf/types/known/timestamppb`
  - `github.com/bradley-mader/amortization-calculator-api/gen/amortization/v1` (aliased as `amortizationv1`)
- Define struct:
  ```go
  type AmortizationServiceHandler struct{}
  func NewAmortizationServiceHandler() *AmortizationServiceHandler {
      return &AmortizationServiceHandler{}
  }
  ```
- Implement `GetAmortizationSchedule` method matching the `amortizationv1connect.AmortizationServiceHandler` interface:
  ```go
  func (h *AmortizationServiceHandler) GetAmortizationSchedule(
      ctx context.Context,
      req *connect.Request[amortizationv1.GetAmortizationScheduleRequest],
  ) (*connect.Response[amortizationv1.GetAmortizationScheduleResponse], error)
  ```
- Request mapping logic:
  1. Extract `req.Msg` fields into `CalculationInput`
  2. `StartDate`: if `req.Msg.StartDate != nil`, use `req.Msg.StartDate.AsTime()`; otherwise leave as zero value (will fail validation)
  3. `AdditionalPayments`: iterate `req.Msg.AdditionalPayments`, map each `*amortizationv1.Payment` to `AdditionalPayment{Date: p.Date.AsTime(), Amount: p.Amount}` (nil-check `p.Date`)
  4. Direct field mappings: `Principal`, `Interest`, `Term`, `PaymentDayOfMonth`, `SkipFirstMonth`
- Call `Calculate(input)`:
  - On error: return `nil, connect.NewError(connect.CodeInvalidArgument, err)`
- Response mapping logic:
  1. Create `[]*amortizationv1.SchedulePayment` from `result.Schedule`
  2. Each `ScheduleEntry` maps to `&amortizationv1.SchedulePayment{Date: timestamppb.New(entry.Date), TotalPrincipalPaid, TotalInterestPaid, PaymentAmount, PaymentPrincipal, PaymentInterest}`
  3. Build response: `&amortizationv1.GetAmortizationScheduleResponse{MonthlyPayment: result.MonthlyPayment, Schedule: &amortizationv1.AmortizationSchedule{Payments: payments}}`
  4. Return `connect.NewResponse(resp), nil`

### 2. Run `go mod tidy`
**File**: `go.mod`, `go.sum`
**Action**: Modify (automatic)
**Details**:
- `connectrpc.com/connect` is already in `go.mod`, but `go mod tidy` ensures the dependency graph is clean after adding new imports that reference the connect and proto packages from within `internal/amortization/`.

## Verification
- `go build ./internal/amortization/...` succeeds
- `go vet ./internal/amortization/...` clean
- Confirm that `AmortizationServiceHandler` satisfies the `amortizationv1connect.AmortizationServiceHandler` interface (compile-time check via the build)

## Notes
- The handler struct has no dependencies (no database, no external services), so no constructor injection is needed at this stage.
- The `timestamppb` import is from `google.golang.org/protobuf/types/known/timestamppb`, already an indirect dependency via the generated proto code.
- Nil-checking proto Timestamp fields is important to avoid panics on malformed requests; zero `time.Time` will be caught by `Validate` inside `Calculate`.
- The `amortizationv1connect` package is NOT imported in handler.go itself; it is only needed by the server entrypoint (Step 5 in spec / next step). The handler just needs to satisfy the interface, which is checked at compile time when the server wires it up. However, we can add a compile-time interface assertion `var _ amortizationv1connect.AmortizationServiceHandler = (*AmortizationServiceHandler)(nil)` to catch mismatches early.
