# Implementation — 000004_business_logic_calculator

**Date**: 2026-04-03

## Files Created
- `internal/amortization/calculator.go` — Package `amortization` with types, validation, and calculation logic
- `internal/amortization/calculator_test.go` — Package `amortization_test` with 6 test functions (9 sub-tests for validation)

## Summary

### calculator.go
- Four exported types: `AdditionalPayment`, `ScheduleEntry`, `CalculationInput`, `CalculationResult`
- `Validate()` checks all input constraints (principal > 0, 0 <= interest <= 100, term > 0, 1 <= paymentDayOfMonth <= 28, non-zero startDate, valid additional payments)
- `Calculate()` implements the standard amortization formula with:
  - Monthly rate computation and monthly payment formula (with zero-interest fallback)
  - First payment date calculation with skipFirstMonth support
  - Schedule loop with additional payment application, final payment adjustment, and early termination

### calculator_test.go
- `TestStandard30YearMortgage` — 360 payments, correct monthly payment and total principal
- `TestShortLoan` — 12 payments, full principal repayment
- `TestSkipFirstMonth` — Verifies first payment date advances when skipFirstMonth is true
- `TestAdditionalPaymentsEarlyPayoff` — Verifies early termination with extra payment
- `TestValidationErrors` — 9 sub-tests covering each validation rule
- `TestZeroInterestRate` — Zero interest with exact monthly payment

## Verification
- `go vet ./internal/amortization/...` — clean
- `go test ./internal/amortization/... -v` — all PASS

## Issues Encountered
- Final payment rounding: The standard amortization formula with 2-decimal rounding produces a slight cumulative shortfall. Fixed by setting `principalPortion = remainingPrincipal` on the last scheduled payment (i == Term-1).
