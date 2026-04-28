# Amortization Calculator API — Specification

## Context

Build a greenfield Go API service using ConnectRPC that calculates amortization schedules. The service will be containerized and deployable to Kubernetes via Kustomize overlays, with a single-script local-dev workflow targeting minikube.

## Project Structure

```
AmmortizationCalculatorAPI/
├── proto/amortization/v1/amortization.proto
├── gen/amortization/v1/                    # generated protobuf/connect code
├── cmd/server/main.go
├── internal/amortization/
│   ├── calculator.go
│   ├── calculator_test.go
│   └── handler.go
├── k8s/
│   ├── base/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── kustomization.yaml
│   └── overlays/
│       ├── local-dev/kustomization.yaml
│       ├── dev/kustomization.yaml
│       ├── stage/kustomization.yaml
│       └── prod/kustomization.yaml
├── scripts/deploy-local.sh
├── Dockerfile
├── buf.yaml
├── buf.gen.yaml
├── go.mod
└── go.sum
```

## Step 1: Project Init & Proto Definition

- `go mod init github.com/bradley-mader/amortization-calculator-api`
- Create `buf.yaml` (v2 format) with `buf.build/googleapis/googleapis` dep for Timestamp
- Create `buf.gen.yaml` with `protocolbuffers/go` and `connectrpc/go` remote plugins, output to `gen/`
- Create `proto/amortization/v1/amortization.proto`:
  - `Payment` message (date Timestamp, amount double) — for additional payments input
  - `SchedulePayment` message (date, totalPrincipalPaid, totalInterestPaid, paymentAmount, paymentPrincipal, paymentInterest)
  - `AmortizationSchedule` message (repeated SchedulePayment)
  - `GetAmortizationScheduleRequest` (principal, interest, term, startDate, paymentDayOfMonth, skipFirstMonth, repeated Payment additionalPayments)
  - `GetAmortizationScheduleResponse` with `monthlyPayment` (double) at root level + `AmortizationSchedule`
  - `AmortizationService` with single RPC `GetAmortizationSchedule`

## Step 2: Code Generation

- `buf dep update` → `buf generate` → `go mod tidy`
- Produces `gen/amortization/v1/*.pb.go` and `gen/amortization/v1/amortizationv1connect/*.connect.go`

## Step 3: Business Logic — `internal/amortization/calculator.go`

**Validation:**
- principal > 0, 0 < interest <= 100, term > 0, 1 <= paymentDayOfMonth <= 28, startDate non-zero
- Each additional payment: positive amount, non-zero date

**Monthly payment formula:**
- `r = annualRate / 100 / 12` (monthly rate)
- `M = P * r / (1 - (1+r)^(-n))` (standard amortization), or `P/n` if r=0

**First payment date:**
- Find next occurrence of paymentDayOfMonth on or after startDate
- If `skipFirstMonth == true` and candidate is < 1 month after startDate, advance one more month

**Schedule loop:**
- For each month: compute interest on remaining principal, compute principal portion, apply any additional payments in that period as extra principal
- Round all monetary values to 2 decimal places
- Cap final payment to remaining principal
- Terminate early if remaining principal <= 0.01

**Unit tests** in `calculator_test.go` covering: standard 30yr mortgage, short loan, skipFirstMonth, additional payments causing early payoff, validation errors

## Step 4: ConnectRPC Handler — `internal/amortization/handler.go`

- Implement generated `AmortizationServiceHandler` interface
- Map proto → internal types, call Validate then Calculate, map results back to proto
- Return `connect.CodeInvalidArgument` for validation errors

## Step 5: Server Entrypoint — `cmd/server/main.go`

- Register handler via `amortizationv1connect.NewAmortizationServiceHandler`
- Add `/healthz` endpoint returning 200
- Serve with `h2c` (HTTP/2 cleartext) on port 8080 (configurable via `PORT` env var)
- Dependencies: `connectrpc.com/connect`, `golang.org/x/net/http2`, `golang.org/x/net/http2/h2c`

## Step 6: Dockerfile

- Multi-stage: `golang:1.23-alpine` builder → `alpine:3.20` runtime
- `CGO_ENABLED=0` static binary
- Expose 8080

## Step 7: Kubernetes Configs

**Base** (`k8s/base/`): Deployment (1 replica, resource limits, health probes on `/healthz`, securityContext) + Service (port 8080)

**Overlays:**
- `local-dev`: namespace default, `imagePullPolicy: Never`, image tag `local`
- `dev`: namespace dev, registry image, tag `dev-latest`
- `stage`: namespace stage, 2 replicas
- `prod`: namespace prod, 3 replicas, higher resource limits

## Step 8: Deploy Script — `scripts/deploy-local.sh`

1. Start minikube if not running
2. `kubectl config use-context minikube`
3. `eval $(minikube docker-env)` + `docker build`
4. `kubectl apply -k k8s/overlays/local-dev`
5. `kubectl rollout status` + print port-forward instructions

## Step 9: Verification

1. `go test ./internal/amortization/... -v`
2. `go run ./cmd/server` then curl:
   ```
   curl -X POST http://localhost:8080/amortization.v1.AmortizationService/GetAmortizationSchedule \
     -H "Content-Type: application/json" \
     -d '{"principal":200000,"interest":6.5,"term":360,"startDate":"2026-05-01T00:00:00Z","paymentDayOfMonth":15}'
   ```
3. Verify: 360 payments, first date 2026-05-15, monthly ~$1,264.14, totalPrincipalPaid ≈ $200,000
4. Test validation: paymentDayOfMonth=29 → InvalidArgument error
5. Test skipFirstMonth + additional payments edge cases
6. `bash scripts/deploy-local.sh` + port-forward + repeat curl test

## Key Decisions

- **Interest**: `interest` field is annual percentage (e.g., 6.5 = 6.5%). No daily accrual for stub period — standard monthly amortization.
- **Additional payments**: Applied as extra principal in the period they fall. Fixed payment M does not change; loan pays off sooner.
- **Proto structure**: Separate request/response wrapper messages per RPC best practice.
- **h2c**: HTTP/2 cleartext for ConnectRPC gRPC protocol support without TLS.
- **minikube image loading**: `eval $(minikube docker-env)` + `imagePullPolicy: Never` — no registry needed for local-dev.
