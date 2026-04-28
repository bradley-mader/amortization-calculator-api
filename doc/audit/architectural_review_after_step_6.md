# Architectural Review — After Step 6

## Findings

### Interface Design

- **Issue**: The handler type `AmortizationServiceHandler` in `internal/amortization/handler.go` directly implements the generated `amortizationv1connect.AmortizationServiceHandler` interface, which is correct. However, the business logic (`Calculate`, `Validate`) is consumed as package-level functions rather than through an interface. This means the handler cannot be tested with a stubbed calculator, and the handler is tightly coupled to the concrete implementation.
- **Location**: `internal/amortization/handler.go`, `internal/amortization/calculator.go`
- **Recommendation**: Define a `Calculator` interface on the handler (consumer) side with a single `Calculate(CalculationInput) (*CalculationResult, error)` method. Inject it into `AmortizationServiceHandler` via a constructor parameter. This follows the Go idiom of declaring interfaces where they are used and enables handler-level unit tests with a mock calculator.

---

- **Issue**: The generated ConnectRPC interface (`AmortizationServiceHandler`) is a single-method interface, which is ideal. The project correctly avoids defining any unnecessary custom interfaces. This is good.
- **Location**: `gen/amortization/v1/amortizationv1connect/amortization.connect.go`
- **Recommendation**: No action needed. Keep interfaces minimal as the service grows.

### Idiomatic Go

- **Issue**: Error wrapping is not used. The `Validate` function returns errors via `fmt.Errorf` with `%f` / `%d` formatting but never wraps sentinel errors or uses `%w`. While the current codebase is small enough that this is acceptable, it prevents callers from programmatically distinguishing error types (e.g., distinguishing a principal validation failure from a term validation failure).
- **Location**: `internal/amortization/calculator.go` (lines 72-97)
- **Recommendation**: For now this is acceptable since all validation errors are mapped to `connect.CodeInvalidArgument` uniformly. If finer-grained error handling is ever needed, introduce sentinel errors (e.g., `var ErrInvalidPrincipal = errors.New(...)`) and wrap them with `fmt.Errorf("...: %w", ErrInvalidPrincipal)`.

---

- **Issue**: The `context.Context` parameter received by `GetAmortizationSchedule` is not propagated to `Calculate`. While `Calculate` is currently a pure computation that does not need a context, accepting a context would future-proof it (e.g., for cancellation on very large term values or adding tracing spans).
- **Location**: `internal/amortization/handler.go` (line 29), `internal/amortization/calculator.go` (line 100)
- **Recommendation**: Low priority. Consider adding `context.Context` as the first parameter to `Calculate` if observability or cancellation becomes relevant.

---

- **Issue**: The `AmortizationServiceHandler` struct has no fields and its constructor `NewAmortizationServiceHandler()` takes no parameters. This is a missed opportunity for dependency injection.
- **Location**: `internal/amortization/handler.go` (lines 17-21)
- **Recommendation**: Add a calculator dependency to the struct (see Interface Design finding above). Even without an interface today, structuring the constructor to accept dependencies establishes the right pattern.

### Consistent Naming

- **Issue**: The struct name `AmortizationServiceHandler` collides with the generated interface name `amortizationv1connect.AmortizationServiceHandler`. While Go's package-based namespacing prevents compilation errors, it creates confusion when reading code — a reader must check imports to know which `AmortizationServiceHandler` is being referenced.
- **Location**: `internal/amortization/handler.go` (line 17) vs. `gen/amortization/v1/amortizationv1connect/amortization.connect.go` (line 78)
- **Recommendation**: Rename the concrete type to something distinct, such as `ServiceHandler` or `Handler`. Since it lives in the `amortization` package, `amortization.Handler` is clear and avoids the collision.

---

- **Issue**: Package naming is good. `amortization`, `main`, and the generated packages all follow Go conventions (short, lowercase, no underscores). Function and type names use MixedCaps correctly. Variable names are appropriately short (`r`, `n`, `M`, `ap`, `i`).
- **Location**: All files.
- **Recommendation**: No action needed.

---

- **Issue**: The proto field `Payment` is named generically, while the Go domain type is `AdditionalPayment`. The naming mismatch between proto and Go is intentional (proto uses shorter names, Go is more descriptive), but it is worth documenting this mapping.
- **Location**: `proto/amortization/v1/amortization.proto` (line 10), `internal/amortization/calculator.go` (line 11)
- **Recommendation**: No action needed. The handler cleanly maps between the two, and the proto naming is appropriate for a wire format.

### Safety

- **Issue**: The code uses `float64` for monetary calculations. This is inherently imprecise due to IEEE 754 floating-point representation. The `roundToTwoDecimals` helper mitigates this for individual values, but accumulated rounding errors across 360 payments could produce a final `TotalPrincipalPaid` that differs from the original principal by several cents.
- **Location**: `internal/amortization/calculator.go` (entire file)
- **Recommendation**: For a calculator/estimation tool this is acceptable. If this were used for actual billing or accounting, switch to integer-based cent arithmetic or a decimal library (e.g., `shopspring/decimal`). Document the precision limitation.

---

- **Issue**: No race conditions exist. All functions are pure (no shared mutable state), and the handler struct has no fields. The `Calculate` function copies the additional payments slice before sorting, which is correct defensive behavior.
- **Location**: `internal/amortization/calculator.go` (lines 117-121)
- **Recommendation**: No action needed. Good practice.

---

- **Issue**: The `main` function does not configure HTTP server timeouts (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`). A production server without timeouts is vulnerable to slowloris attacks and resource exhaustion.
- **Location**: `cmd/server/main.go` (lines 31-34)
- **Recommendation**: Add timeouts to the `http.Server` configuration. Example: `ReadTimeout: 10 * time.Second`, `WriteTimeout: 30 * time.Second`, `IdleTimeout: 120 * time.Second`.

---

- **Issue**: The server does not handle graceful shutdown. `log.Fatal(server.ListenAndServe())` will terminate the process immediately on error, and there is no signal handling for SIGTERM/SIGINT to drain in-flight requests.
- **Location**: `cmd/server/main.go` (line 38)
- **Recommendation**: Add `os/signal` handling to listen for `SIGTERM`/`SIGINT` and call `server.Shutdown(ctx)` with a timeout context. This is important for container deployments (Kubernetes, Docker).

---

- **Issue**: Input validation correctly caps `PaymentDayOfMonth` at 28 to avoid month-length edge cases. The `StartDate` zero-value check is present. Additional payments are validated for positive amounts and non-zero dates. This is thorough.
- **Location**: `internal/amortization/calculator.go` (lines 72-97)
- **Recommendation**: No action needed. Validation is solid.

### Testability

- **Issue**: The business logic in `calculator.go` is a pure function with no external dependencies, making it highly testable. The existing tests cover the key scenarios: standard mortgage, short loan, skip-first-month, additional payments with early payoff, zero interest, and validation errors. This is well-structured.
- **Location**: `internal/amortization/calculator_test.go`
- **Recommendation**: No action needed for calculator tests.

---

- **Issue**: There are no tests for the handler (`handler.go`). The handler contains non-trivial mapping logic (proto to domain types and back) that could have bugs — for example, if a new field is added to the proto but not mapped. Without handler tests, this mapping is only validated through manual/integration testing.
- **Location**: `internal/amortization/handler.go`
- **Recommendation**: Add unit tests for `GetAmortizationSchedule` that verify the proto-to-domain and domain-to-proto mapping. With the current package-level function design, these tests would exercise the full stack (handler + calculator), which is acceptable. If the calculator were injected via interface (see Interface Design finding), the handler could be tested in isolation with a stub.

---

- **Issue**: There are no tests for `main.go`. While `main` is thin, there is no test that the server wires up correctly (routes are registered, health endpoint works).
- **Location**: `cmd/server/main.go`
- **Recommendation**: Low priority. Consider extracting server setup into a `newServer()` function that returns an `*http.Server`, and writing a test that starts it on a random port and hits `/healthz`.

---

- **Issue**: The `firstPaymentDate` helper function is unexported and only tested indirectly through the schedule tests. It contains date arithmetic that is a common source of bugs.
- **Location**: `internal/amortization/calculator.go` (lines 50-69)
- **Recommendation**: Add dedicated table-driven tests for `firstPaymentDate` via an exported test helper or by testing it through carefully chosen `Calculate` inputs that exercise edge cases (e.g., start date on the 31st, start date equal to payment day, start date one day before payment day).

## Overall Assessment

The codebase is clean, well-organized, and demonstrates good Go fundamentals for an early-stage project. The separation between pure business logic (`calculator.go`) and transport concerns (`handler.go`) is sound, and the use of ConnectRPC with protobuf provides a solid API foundation. The most impactful improvements would be: (1) injecting the calculator as an interface dependency into the handler to enable isolated handler testing, (2) adding HTTP server timeouts and graceful shutdown for production readiness, and (3) adding handler-level tests to catch proto mapping regressions. The floating-point arithmetic is acceptable for an estimation tool but should be documented as a known limitation.
