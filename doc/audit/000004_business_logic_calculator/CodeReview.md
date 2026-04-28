# Code Review — 000004_business_logic_calculator

**Commit**: a61e949 — feat: implement amortization calculation engine with validation
**Date**: 2026-04-03

## Summary

Adds the `internal/amortization` package containing types, validation, and a calculation engine for amortization schedules, along with comprehensive unit tests covering standard loans, zero-interest, additional payments, skip-first-month, and validation error paths.

## Findings

### Critical

None.

### Major

1. **Additional payments scanned on every iteration (O(n*m))**: The inner loop on line 149-152 iterates over all additional payments for every month. For typical use (small number of additional payments) this is fine, but the pattern could be improved by consuming matched extras or using a pointer/index into the sorted slice. Low urgency since m is expected to be small.

2. **`firstPaymentDate` edge case with `skipFirstMonth` when `startDate` day equals `dayOfMonth`**: When `startDate` is e.g. 2026-05-05 and `dayOfMonth` is 5, the candidate is exactly `startDate`. With `skipFirstMonth=true`, `oneMonthLater` is 2026-06-05 and `candidate` (2026-05-05) is before it, so it advances to 2026-06-05. This is correct behavior. However, when `startDate` day is equal to `dayOfMonth` and `skipFirstMonth` is false, the candidate equals startDate exactly (payment on the same day loan starts). This is arguably correct but worth documenting as an intentional decision.

### Minor

1. **`Validate` returns on first error only**: The function returns the first validation error encountered rather than collecting all errors. This is fine for the current use case (ConnectRPC handler will map to `InvalidArgument`), but for a richer UX, consider returning multiple validation errors. Low priority.

2. **`ScheduleEntry.PaymentPrincipal` field name mismatch with spec**: The spec defines the field as `PaymentPrincipal` but in the `Calculate` function it is populated with `totalPrincipalThisMonth` which includes extra principal from additional payments. This is semantically correct (it is the total principal paid in that payment), but the field name `PaymentPrincipal` could be confused with only the standard principal portion. The naming is consistent with the proto `payment_principal` field so this is acceptable.

3. **No pre-allocation of `schedule` slice**: On line 127, `var schedule []ScheduleEntry` grows via `append`. Pre-allocating with `make([]ScheduleEntry, 0, input.Term)` would avoid re-allocations for the common case. Minor performance improvement.

4. **Unused `sort` import possibility**: The `sort` package is used correctly for additional payments, but if `AdditionalPayments` is empty, the sort is a no-op. No issue, just noting.

5. **`fmt.Errorf` without error wrapping**: Validation errors use `fmt.Errorf` with `%f` and `%d` verbs. These are descriptive but do not use `%w` for error wrapping. Since these are leaf errors (not wrapping an underlying cause), this is acceptable and idiomatic.

### Positive

1. **Clean separation of concerns**: `Validate` is exported separately from `Calculate`, enabling the future handler to call validation independently if needed.

2. **Defensive copy of additional payments**: Line 117-118 copies the input slice before sorting, avoiding mutation of the caller's data. Good practice.

3. **Final payment adjustment**: The `i == input.Term-1` check on line 139 correctly handles cumulative rounding drift by forcing the last payment to clear the remaining balance. Well-thought-out fix.

4. **Comprehensive test coverage**: Six test functions with 9 validation sub-tests cover the key scenarios including edge cases (zero interest, skip first month, early payoff). The use of table-driven tests for validation is idiomatic Go.

5. **Consistent rounding**: All monetary values pass through `roundToTwoDecimals` before being stored, preventing floating-point drift in the schedule output.

6. **External test package**: Using `amortization_test` (not `amortization`) tests the public API surface only, which is the correct approach for this package.

## Files Reviewed
- `internal/amortization/calculator.go`
- `internal/amortization/calculator_test.go`
- `doc/audit/000004_business_logic_calculator/Implementation.md`
- `doc/audit/000004_business_logic_calculator/Implementation_Plan.md`
