# Code Review Assessment — 000004_business_logic_calculator

**Date**: 2026-04-03

## Items to Implement

### Pre-allocate schedule slice
- **Original severity**: Minor
- **Decision**: SHOULD FIX
- **Reason**: Simple one-line change that avoids unnecessary re-allocations for the common case (360 payments for a 30-year mortgage means multiple grow-and-copy cycles). Low effort, measurable benefit.
- **Action**: Change `var schedule []ScheduleEntry` to `schedule := make([]ScheduleEntry, 0, input.Term)` on line 127.

## Items Skipped

### Additional payments scanned on every iteration (O(n*m))
- **Original severity**: Major
- **Decision**: SKIP
- **Reason**: The review itself notes "low urgency since m is expected to be small." In practice, additional payments will number in the single digits while n is at most 360. The O(n*m) cost is negligible. Adding index tracking would add complexity for no real-world gain.

### `firstPaymentDate` edge case documentation
- **Original severity**: Major
- **Decision**: SKIP
- **Reason**: The review confirms the behavior is correct. The edge case (payment on the same day the loan starts) is reasonable default behavior. Adding a code comment would be fine but is not a correctness or maintainability issue. The existing doc comment on the function is sufficient.

### `Validate` returns on first error only
- **Original severity**: Minor
- **Decision**: SKIP
- **Reason**: Single-error-return is the standard Go idiom for validation functions. The ConnectRPC handler will map this to a single `InvalidArgument` error code. Multi-error collection would be premature complexity; it can be added later if UX demands it.

### `ScheduleEntry.PaymentPrincipal` field name
- **Original severity**: Minor
- **Decision**: SKIP
- **Reason**: The review acknowledges it is consistent with the proto definition (`payment_principal`). The semantics (total principal applied in that payment including extras) are correct and documented by the field's position in the schedule entry.

### Unused sort import possibility
- **Original severity**: Minor
- **Decision**: SKIP
- **Reason**: Not an issue. The sort is correctly used and handles the empty case as a no-op. No action needed.

### `fmt.Errorf` without error wrapping
- **Original severity**: Minor
- **Decision**: SKIP
- **Reason**: These are leaf errors with no underlying cause to wrap. Using `%w` would be incorrect here since there is no sentinel error. The current approach is idiomatic.

## Summary
- Total findings: 7
- Items to implement: 1
- Items skipped: 6

## Action Items
1. Pre-allocate the `schedule` slice with `make([]ScheduleEntry, 0, input.Term)` in the `Calculate` function.
