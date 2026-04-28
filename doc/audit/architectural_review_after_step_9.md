# Architectural Review — After Step 9

## Findings

### Interface Design

- **Issue**: The handler calls `Calculate` and `Validate` as package-level functions rather than through an injected interface. The `AmortizationServiceHandler` struct has zero fields, making it impossible to substitute the calculator in tests. This was identified in the step-6 review and remains unresolved.
- **Location**: `internal/amortization/handler.go` (lines 17-21, 56)
- **Recommendation**: Define a consumer-side `Calculator` interface with a single `Calculate(CalculationInput) (*CalculationResult, error)` method. Inject it into the handler struct via the constructor. This enables handler-level unit tests with a stub and follows the Go idiom of declaring interfaces at the call site.

---

- **Issue**: The generated ConnectRPC interface `AmortizationServiceHandler` is a single-method interface, which is ideal and follows the Go proverb of small interfaces. No unnecessary custom interfaces exist.
- **Location**: `gen/amortization/v1/amortizationv1connect/amortization.connect.go`
- **Recommendation**: No action needed.

### Idiomatic Go

- **Issue**: Errors returned by `Validate` use `fmt.Errorf` with verb formatting (`%f`, `%d`) but never wrap sentinel errors with `%w`. Callers cannot programmatically distinguish between validation failure categories.
- **Location**: `internal/amortization/calculator.go` (lines 72-97)
- **Recommendation**: Acceptable for now since all validation errors are uniformly mapped to `connect.CodeInvalidArgument`. If finer-grained error handling is ever needed, introduce sentinel errors and wrap with `%w`.

---

- **Issue**: The `context.Context` received by `GetAmortizationSchedule` is not propagated to `Calculate`. The calculation is currently pure and does not need a context, but this blocks future addition of tracing spans or cancellation for large term values.
- **Location**: `internal/amortization/handler.go` (line 29), `internal/amortization/calculator.go` (line 100)
- **Recommendation**: Low priority. Add `context.Context` as the first parameter to `Calculate` when observability is introduced.

---

- **Issue**: The `Validate` function is exported and called independently from `Calculate`, yet `Calculate` also calls `Validate` internally (line 101). This dual-path design is slightly unusual. External callers might call `Validate` then `Calculate`, resulting in double validation.
- **Location**: `internal/amortization/calculator.go` (lines 72, 100-103)
- **Recommendation**: This is a minor inefficiency, not a bug. Consider making `Validate` unexported if it is not needed outside the package, or document that `Calculate` validates internally so callers need not call `Validate` separately. The test file does call `Validate` directly (line 209), which is fine for testing but would work equally well through `Calculate`.

---

- **Issue**: The variable `M` (monthly payment, line 108) uses a single uppercase letter, which in Go convention suggests an exported name. While single-letter variables are idiomatic for loop indices and short-lived values, `M` as uppercase is slightly misleading.
- **Location**: `internal/amortization/calculator.go` (line 108)
- **Recommendation**: Rename to `monthlyPayment` or `mp` for clarity. Very low priority.

### Consistent Naming

- **Issue**: The concrete type `AmortizationServiceHandler` in the `amortization` package collides with the generated interface name `amortizationv1connect.AmortizationServiceHandler`. This was identified in step-6 review and remains unresolved. A reader must inspect imports to determine which type is referenced.
- **Location**: `internal/amortization/handler.go` (line 17) vs. `gen/amortization/v1/amortizationv1connect/amortization.connect.go` (line 78)
- **Recommendation**: Rename the concrete type to `Handler` or `ServiceHandler`. Since it lives in the `amortization` package, `amortization.Handler` is unambiguous.

---

- **Issue**: The constructor `NewAmortizationServiceHandler()` returns `*AmortizationServiceHandler`, and the generated connect function is also called `NewAmortizationServiceHandler`. In `main.go`, both are called in adjacent lines (23-24) from different packages, which is confusing: `amortization.NewAmortizationServiceHandler()` vs. `amortizationv1connect.NewAmortizationServiceHandler(...)`.
- **Location**: `cmd/server/main.go` (lines 23-24)
- **Recommendation**: Renaming the concrete type (see above) would also fix the constructor name collision. `amortization.NewHandler()` and `amortizationv1connect.NewAmortizationServiceHandler(...)` are clearly distinct.

---

- **Issue**: The proto message `Payment` maps to Go domain type `AdditionalPayment`. The naming difference is intentional and the handler maps cleanly between them. Package naming throughout (`amortization`, `main`, generated packages) follows Go conventions.
- **Location**: All files
- **Recommendation**: No action needed.

### Safety

- **Issue**: The `http.Server` in `main.go` has no timeouts configured (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`). This was identified in step-6 review and remains unresolved. A production server without timeouts is vulnerable to slowloris attacks and resource exhaustion from slow clients.
- **Location**: `cmd/server/main.go` (lines 31-34)
- **Recommendation**: Add timeouts: `ReadTimeout: 10 * time.Second`, `WriteTimeout: 30 * time.Second`, `IdleTimeout: 120 * time.Second`.

---

- **Issue**: The server does not handle graceful shutdown. `log.Fatal(server.ListenAndServe())` terminates the process immediately, and there is no signal handling for SIGTERM/SIGINT to drain in-flight requests. This is important for Kubernetes deployments where the pod receives SIGTERM during rolling updates.
- **Location**: `cmd/server/main.go` (line 37)
- **Recommendation**: Add `os/signal.NotifyContext` for SIGTERM/SIGINT, run `ListenAndServe` in a goroutine, and call `server.Shutdown(ctx)` with a deadline on signal receipt.

---

- **Issue**: The `float64` type is used for all monetary values. IEEE 754 floating-point arithmetic introduces rounding errors that accumulate over 360+ payments. The `roundToTwoDecimals` helper mitigates this per-value, but cumulative drift is possible.
- **Location**: `internal/amortization/calculator.go` (entire file)
- **Recommendation**: Acceptable for an estimation/calculator tool. Not suitable for billing or accounting without switching to integer-cent arithmetic or a decimal library. Document the precision limitation.

---

- **Issue**: No race conditions. All functions are pure with no shared mutable state. The handler struct is stateless. The `Calculate` function defensively copies the additional payments slice before sorting (line 117-118), preventing mutation of the caller's data.
- **Location**: `internal/amortization/calculator.go`
- **Recommendation**: No action needed.

---

- **Issue**: Input validation is thorough. `PaymentDayOfMonth` is capped at 28 to avoid month-length edge cases. `StartDate` zero-value is checked. Additional payments are validated for positive amounts and non-zero dates. The `Term` upper bound of 600 (50 years) is reasonable.
- **Location**: `internal/amortization/calculator.go` (lines 72-97)
- **Recommendation**: No action needed.

---

- **Issue**: The Dockerfile uses `golang:1.23-alpine` as the builder image, but `go.mod` declares `go 1.25.0`. This version mismatch means the Dockerfile will fail to build since Go 1.23 cannot compile a module requiring Go 1.25.
- **Location**: `Dockerfile` (line 1), `go.mod` (line 3)
- **Recommendation**: Update the Dockerfile to use `golang:1.25-alpine` (or the latest available image that satisfies `go 1.25.0`). This is a build-breaking issue.

---

- **Issue**: The Dockerfile's runtime stage uses `alpine:3.20` but does not pin a specific digest. For reproducible production builds, pinning the exact digest is a best practice.
- **Location**: `Dockerfile` (line 8)
- **Recommendation**: Low priority. Consider pinning the digest or using a specific patch version for reproducibility.

### Testability

- **Issue**: The business logic in `calculator.go` is a pure function with no I/O, making it highly testable. The existing test suite covers six scenarios: standard 30-year mortgage, short loan, skip-first-month, additional payments with early payoff, zero interest, and a comprehensive validation error table. This is solid coverage.
- **Location**: `internal/amortization/calculator_test.go`
- **Recommendation**: No action needed for calculator tests.

---

- **Issue**: There are still no handler-level tests. The handler contains mapping logic between proto types and domain types that could regress when new fields are added to the proto. This was identified in step-6 review and remains unresolved.
- **Location**: `internal/amortization/handler.go`
- **Recommendation**: Add unit tests for `GetAmortizationSchedule` using `connect.NewRequest` to verify the proto-to-domain and domain-to-proto mapping. With calculator injection (see Interface Design), the handler can be tested in isolation.

---

- **Issue**: There are no tests for the health endpoint or server wiring in `main.go`.
- **Location**: `cmd/server/main.go`
- **Recommendation**: Low priority. Consider extracting server setup into a testable `newServer()` function.

---

- **Issue**: The `firstPaymentDate` helper is unexported and only tested indirectly through schedule calculations. It contains date arithmetic (month advancement, skip logic) that is a common source of off-by-one bugs.
- **Location**: `internal/amortization/calculator.go` (lines 50-69)
- **Recommendation**: Add table-driven tests that exercise edge cases: start date equals payment day, start date is one day before payment day, start date on the last day of a short month, skip-first-month with various day combinations.

### Kubernetes Configuration

- **Issue**: The k8s overlays for `dev`, `stage`, and `prod` use placeholder `<registry>` in the image `newName` field. While this is expected for a template, it will fail if applied directly without substitution.
- **Location**: `k8s/overlays/dev/kustomization.yaml`, `k8s/overlays/stage/kustomization.yaml`, `k8s/overlays/prod/kustomization.yaml`
- **Recommendation**: Document the expected substitution mechanism (CI/CD pipeline variable, kustomize edit, etc.) or use a real registry path.

---

- **Issue**: The deployment does not define a `PodDisruptionBudget` for `stage` or `prod` overlays. With 2-3 replicas, a PDB would prevent simultaneous eviction of all pods during node maintenance.
- **Location**: `k8s/overlays/stage/kustomization.yaml`, `k8s/overlays/prod/kustomization.yaml`
- **Recommendation**: Add a `PodDisruptionBudget` with `minAvailable: 1` for stage and prod.

---

- **Issue**: The base deployment correctly includes `securityContext` with `runAsNonRoot: true` and `runAsUser: 65534` (nobody). Liveness and readiness probes are configured against `/healthz`. Resource requests and limits are defined. This is well-structured for a Kubernetes deployment.
- **Location**: `k8s/base/deployment.yaml`
- **Recommendation**: No action needed.

## Overall Assessment

The codebase is clean, well-organized, and demonstrates strong Go fundamentals. The separation between pure business logic (`calculator.go`) and transport concerns (`handler.go`) is sound, the protobuf/ConnectRPC setup is correctly configured, and the Kubernetes manifests follow best practices with security contexts, resource limits, and health probes. The test suite provides good coverage of the core calculation logic. Three issues from the step-6 review remain unresolved and should be prioritized: (1) the Dockerfile Go version mismatch with `go.mod` is a build-breaking bug that needs immediate attention, (2) injecting the calculator as an interface dependency would unlock handler-level testing and follow idiomatic Go dependency injection patterns, and (3) adding HTTP server timeouts and graceful shutdown is essential for production readiness in Kubernetes where SIGTERM handling ensures zero-downtime deployments.
