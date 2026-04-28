# Implementation Plan — 000004_business_logic_calculator

**Date**: 2026-04-03
**High-level step**: Implement amortization calculation engine with validation

## Prerequisites
- Go module initialized (`go.mod` exists with `github.com/bradley-mader/amortization-calculator-api`)
- Proto definitions and generated code exist in `gen/amortization/v1/`

## Sub-tasks

### 1. Create `internal/amortization/` directory
**File**: `internal/amortization/`
**Action**: Create directory
**Details**: Create the package directory for the business logic layer.

### 2. Create `internal/amortization/calculator.go` with types and Validate function
**File**: `internal/amortization/calculator.go`
**Action**: Create
**Details**:
- Package: `amortization`
- Imports: `errors`, `fmt`, `math`, `time`
- Define types:
  ```go
  type AdditionalPayment struct {
      Date   time.Time
      Amount float64
  }

  type ScheduleEntry struct {
      Date               time.Time
      TotalPrincipalPaid float64
      TotalInterestPaid  float64
      PaymentAmount      float64
      PaymentPrincipal   float64
      PaymentInterest    float64
  }

  type CalculationInput struct {
      Principal          float64
      Interest           float64
      Term               int32
      StartDate          time.Time
      PaymentDayOfMonth  int32
      SkipFirstMonth     bool
      AdditionalPayments []AdditionalPayment
  }

  type CalculationResult struct {
      MonthlyPayment float64
      Schedule       []ScheduleEntry
  }
  ```
- `func Validate(input CalculationInput) error`:
  - Check `Principal > 0` or return error
  - Check `0 <= Interest <= 100` (allow 0 for zero-interest loans) or return error
  - Check `Term > 0` or return error
  - Check `1 <= PaymentDayOfMonth <= 28` or return error
  - Check `StartDate` is non-zero or return error
  - For each `AdditionalPayment`: check `Amount > 0` and `Date` is non-zero
  - Return `nil` if all pass

### 3. Implement `Calculate` function
**File**: `internal/amortization/calculator.go`
**Action**: Modify (append)
**Details**:
- Helper: `func roundToTwoDecimals(val float64) float64` using `math.Round(val*100) / 100`
- Helper: `func firstPaymentDate(startDate time.Time, dayOfMonth int32, skipFirstMonth bool) time.Time`:
  - Candidate = startDate's year/month with day = dayOfMonth
  - If candidate is before startDate, advance one month
  - If skipFirstMonth and candidate is less than one month after startDate, advance one more month
  - Return candidate
- `func Calculate(input CalculationInput) (*CalculationResult, error)`:
  1. Call `Validate(input)`, return error if non-nil
  2. `r := input.Interest / 100.0 / 12.0`
  3. If `r == 0`: `M = float64(input.Principal) / float64(input.Term)` else `M = P * r / (1 - math.Pow(1+r, -float64(n)))`
  4. `M = roundToTwoDecimals(M)`
  5. Compute `paymentDate` via `firstPaymentDate`
  6. Sort additional payments by date
  7. Loop up to `Term` iterations:
     - `interest = roundToTwoDecimals(remainingPrincipal * r)`
     - `principalPortion = M - interest`
     - Compute next payment date for additional payment window
     - Sum additional payments where `ap.Date >= paymentDate && ap.Date < nextPaymentDate`
     - `totalPrincipalThisMonth = principalPortion + extraPrincipal`
     - Cap: if `totalPrincipalThisMonth > remainingPrincipal`, adjust down; also adjust paymentAmount accordingly
     - `remainingPrincipal -= totalPrincipalThisMonth`
     - Accumulate `totalPrincipalPaid += totalPrincipalThisMonth`, `totalInterestPaid += interest`
     - Append `ScheduleEntry` with date, running totals, this month's amounts
     - Advance `paymentDate` by one month
     - If `remainingPrincipal <= 0.01`, break
  8. Return `&CalculationResult{MonthlyPayment: M, Schedule: schedule}`

### 4. Create `internal/amortization/calculator_test.go`
**File**: `internal/amortization/calculator_test.go`
**Action**: Create
**Details**:
- Package: `amortization_test`
- Import: `testing`, `time`, `math`, `github.com/bradley-mader/amortization-calculator-api/internal/amortization`
- Test cases:
  - `TestStandard30YearMortgage`: P=200000, I=6.5, T=360, start=2026-05-01, day=15. Assert monthlyPayment ~= 1264.14 (within 0.01), len(schedule) == 360, schedule[0].Date == 2026-05-15, totalPrincipalPaid ~= 200000 (within 1.0).
  - `TestShortLoan`: P=10000, I=5.0, T=12, day=1, start=2026-01-01. Assert 12 payments, final entry totalPrincipalPaid ~= 10000 (within 0.01).
  - `TestSkipFirstMonth`: P=10000, I=5.0, T=12, start=2026-05-01, day=5, skip=true. Assert schedule[0].Date == 2026-06-05.
  - `TestAdditionalPaymentsEarlyPayoff`: P=10000, I=5.0, T=12, day=1, start=2026-01-01, extra 5000 on 2026-02-15. Assert len(schedule) < 12.
  - `TestValidationErrors`: Sub-tests for each invalid input (zero principal, negative interest, interest > 100, zero term, day < 1, day > 28, zero startDate, bad additional payment amount, bad additional payment date).
  - `TestZeroInterestRate`: P=12000, I=0, T=12, day=1, start=2026-01-01. Assert monthlyPayment == 1000.0, final totalInterestPaid == 0.

## Verification
- `go vet ./internal/amortization/...`
- `go test ./internal/amortization/... -v`

## Notes
- The Validate function allows Interest == 0 to support zero-interest loans (the spec says `0 < interest <= 100` but the test `TestZeroInterestRate` requires interest=0 to be valid, so we use `0 <= interest <= 100`).
- All monetary rounding uses `math.Round(val*100)/100` for consistent 2-decimal-place rounding.
- Additional payments are applied as extra principal only; they do not alter the fixed monthly payment amount M.
- Payment dates use `time.Date(year, month, day, 0, 0, 0, 0, time.UTC)` to avoid timezone issues.
