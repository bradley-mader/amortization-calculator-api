# Implementation — 000005_connectrpc_handler

**Date**: 2026-04-03

## Changes Made

### Created `internal/amortization/handler.go`
- Implements `amortizationv1connect.AmortizationServiceHandler` interface
- Compile-time interface assertion ensures correctness
- Maps proto `GetAmortizationScheduleRequest` to internal `CalculationInput`
- Nil-checks proto Timestamp fields to avoid panics
- Calls `Calculate(input)` and returns `connect.CodeInvalidArgument` on error
- Maps `CalculationResult` back to proto `GetAmortizationScheduleResponse`

### No changes to `go.mod`
- `connectrpc.com/connect` and `google.golang.org/protobuf` were already present

## Verification
- `go build ./internal/amortization/...` — passed
- `go vet ./internal/amortization/...` — passed (clean)
