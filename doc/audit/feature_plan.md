# Feature Implementation Plan

**Date**: 2026-04-03
**Source**: Amortization Calculator API — Specification

## Overview
Build a greenfield Go ConnectRPC API service that calculates amortization schedules, containerized and deployable to Kubernetes via Kustomize overlays with a minikube local-dev workflow.

## Steps

### Step 1: Project Init & Buf Configuration
**Goal**: Initialize Go module and configure Buf for proto code generation
**Creates**: `go.mod`, `buf.yaml`, `buf.gen.yaml`
**Modifies**: None
**Details**:
- Run `go mod init github.com/bradley-mader/amortization-calculator-api`
- Create `buf.yaml` (v2 format):
  ```yaml
  version: v2
  modules:
    - path: proto
  deps:
    - buf.build/googleapis/googleapis
  ```
- Create `buf.gen.yaml`:
  ```yaml
  version: v2
  plugins:
    - remote: buf.build/protocolbuffers/go
      out: gen
      opt: paths=source_relative
    - remote: buf.build/connectrpc/go
      out: gen
      opt: paths=source_relative
  ```
- Run `go mod tidy`
**Verification**: `cat go.mod` shows correct module path; `cat buf.yaml buf.gen.yaml` shows correct config

### Step 2: Proto Definition
**Goal**: Define the amortization service protobuf schema
**Creates**: `proto/amortization/v1/amortization.proto`
**Modifies**: None
**Details**:
- Package: `amortization.v1`
- Go package option: `github.com/bradley-mader/amortization-calculator-api/gen/amortization/v1;amortizationv1`
- Import `google/protobuf/timestamp.proto`
- Messages:
  - `Payment`: `google.protobuf.Timestamp date = 1; double amount = 2;`
  - `SchedulePayment`: `google.protobuf.Timestamp date = 1; double total_principal_paid = 2; double total_interest_paid = 3; double payment_amount = 4; double payment_principal = 5; double payment_interest = 6;`
  - `AmortizationSchedule`: `repeated SchedulePayment payments = 1;`
  - `GetAmortizationScheduleRequest`: `double principal = 1; double interest = 2; int32 term = 3; google.protobuf.Timestamp start_date = 4; int32 payment_day_of_month = 5; bool skip_first_month = 6; repeated Payment additional_payments = 7;`
  - `GetAmortizationScheduleResponse`: `double monthly_payment = 1; AmortizationSchedule schedule = 2;`
- Service `AmortizationService`:
  - `rpc GetAmortizationSchedule(GetAmortizationScheduleRequest) returns (GetAmortizationScheduleResponse);`
**Verification**: `buf lint` passes

### Step 3: Code Generation
**Goal**: Generate Go protobuf and ConnectRPC code from proto definition
**Creates**: `gen/amortization/v1/*.pb.go`, `gen/amortization/v1/amortizationv1connect/*.connect.go`
**Modifies**: `go.mod`, `go.sum`
**Details**:
- Run `buf dep update`
- Run `buf generate`
- Run `go mod tidy`
- Verify generated files exist in `gen/amortization/v1/` and `gen/amortization/v1/amortizationv1connect/`
**Verification**: `ls gen/amortization/v1/` shows `.pb.go` files; `ls gen/amortization/v1/amortizationv1connect/` shows `.connect.go` files; `go build ./...` succeeds

### Step 4: Business Logic — Calculator
**Goal**: Implement amortization calculation engine with validation
**Creates**: `internal/amortization/calculator.go`, `internal/amortization/calculator_test.go`
**Modifies**: None
**Details**:
- File: `internal/amortization/calculator.go`
- Package: `amortization`
- Types:
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
- Functions:
  - `func Validate(input CalculationInput) error` — validates: principal > 0, 0 < interest <= 100, term > 0, 1 <= paymentDayOfMonth <= 28, startDate non-zero; each additional payment: positive amount, non-zero date
  - `func Calculate(input CalculationInput) (*CalculationResult, error)` — calls Validate first, then:
    1. Compute monthly rate `r = interest / 100 / 12`
    2. Compute monthly payment `M = P * r / (1 - (1+r)^(-n))`, or `P/n` if r == 0
    3. Round M to 2 decimal places using `math.Round(val*100)/100`
    4. Find first payment date: next occurrence of paymentDayOfMonth on or after startDate. If skipFirstMonth is true and candidate is < 1 month after startDate, advance one more month.
    5. Loop: for each month, compute interest = remainingPrincipal * r, round to 2 decimals. Principal portion = M - interest. Apply additional payments falling in this period (between current and next payment date) as extra principal. Cap final payment so principal portion does not exceed remaining principal. Accumulate totalPrincipalPaid and totalInterestPaid. Terminate early if remainingPrincipal <= 0.01.
    6. Return CalculationResult with MonthlyPayment and Schedule.
- File: `internal/amortization/calculator_test.go`
- Package: `amortization_test`
- Test cases:
  - `TestStandard30YearMortgage`: principal=200000, interest=6.5, term=360, startDate=2026-05-01, paymentDayOfMonth=15. Assert monthlyPayment ~= 1264.14, len(schedule) == 360, first payment date == 2026-05-15, totalPrincipalPaid ~= 200000.
  - `TestShortLoan`: principal=10000, interest=5.0, term=12, paymentDayOfMonth=1. Assert 12 payments, final principal remaining ~0.
  - `TestSkipFirstMonth`: principal=10000, interest=5.0, term=12, startDate=2026-05-01, paymentDayOfMonth=5, skipFirstMonth=true. Assert first payment date is 2026-06-05 (not 2026-05-05).
  - `TestAdditionalPaymentsEarlyPayoff`: principal=10000, interest=5.0, term=12, paymentDayOfMonth=1, additionalPayment of 5000 in month 2. Assert fewer than 12 payments.
  - `TestValidationErrors`: test each validation rule returns an error with invalid input.
  - `TestZeroInterestRate`: principal=12000, interest=0, term=12. Assert monthlyPayment == 1000, no interest paid.
**Verification**: `go test ./internal/amortization/... -v` passes; `go vet ./internal/amortization/...` clean

### Step 5: ConnectRPC Handler
**Goal**: Implement the ConnectRPC service handler mapping proto types to internal types
**Creates**: `internal/amortization/handler.go`
**Modifies**: `go.mod` (new dependency on connectrpc.com/connect)
**Details**:
- File: `internal/amortization/handler.go`
- Package: `amortization`
- Import: `connectrpc.com/connect`, generated packages from `gen/amortization/v1` and `gen/amortization/v1/amortizationv1connect`
- Struct:
  ```go
  type AmortizationServiceHandler struct{}

  func NewAmortizationServiceHandler() *AmortizationServiceHandler {
      return &AmortizationServiceHandler{}
  }
  ```
- Method:
  ```go
  func (h *AmortizationServiceHandler) GetAmortizationSchedule(
      ctx context.Context,
      req *connect.Request[amortizationv1.GetAmortizationScheduleRequest],
  ) (*connect.Response[amortizationv1.GetAmortizationScheduleResponse], error)
  ```
  - Map `req.Msg` fields to `CalculationInput`:
    - `StartDate` from proto Timestamp via `req.Msg.StartDate.AsTime()`
    - `AdditionalPayments` mapped from proto `Payment` slice
  - Call `Calculate(input)` — on error, return `connect.NewError(connect.CodeInvalidArgument, err)`
  - Map `CalculationResult` back to proto response:
    - Each `ScheduleEntry` → `SchedulePayment` with date as `timestamppb.New(entry.Date)`
    - Set `MonthlyPayment` at root and `Schedule` with payments
  - Return `connect.NewResponse(resp)`
- Run `go mod tidy`
**Verification**: `go build ./internal/amortization/...` succeeds; `go vet ./internal/amortization/...` clean

### Step 6: Server Entrypoint
**Goal**: Create the main server binary with health check endpoint and h2c support
**Creates**: `cmd/server/main.go`
**Modifies**: `go.mod` (new dependencies on `golang.org/x/net/http2`, `golang.org/x/net/http2/h2c`)
**Details**:
- File: `cmd/server/main.go`
- Package: `main`
- Imports: `connectrpc.com/connect`, `golang.org/x/net/http2`, `golang.org/x/net/http2/h2c`, generated connect package, internal amortization package
- Logic:
  1. Read port from `os.Getenv("PORT")`, default "8080"
  2. Create `mux := http.NewServeMux()`
  3. Create handler: `svcHandler := amortization.NewAmortizationServiceHandler()`
  4. Register: `path, handler := amortizationv1connect.NewAmortizationServiceHandler(svcHandler)` then `mux.Handle(path, handler)`
  5. Add health check: `mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })`
  6. Create h2c server:
     ```go
     server := &http.Server{
         Addr:    ":" + port,
         Handler: h2c.NewHandler(mux, &http2.Server{}),
     }
     ```
  7. `log.Printf("server listening on :%s", port)` then `log.Fatal(server.ListenAndServe())`
- Run `go mod tidy`
**Verification**: `go build ./cmd/server/...` succeeds; `go vet ./cmd/server/...` clean

### Step 7: Dockerfile
**Goal**: Create multi-stage Docker build for the server binary
**Creates**: `Dockerfile`
**Modifies**: None
**Details**:
- Stage 1 — builder:
  ```dockerfile
  FROM golang:1.23-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server
  ```
- Stage 2 — runtime:
  ```dockerfile
  FROM alpine:3.20
  RUN apk --no-cache add ca-certificates
  COPY --from=builder /server /server
  EXPOSE 8080
  ENTRYPOINT ["/server"]
  ```
**Verification**: `cat Dockerfile` shows correct multi-stage build

### Step 8: Kubernetes Base Configs
**Goal**: Create base Kubernetes deployment and service manifests with Kustomize
**Creates**: `k8s/base/deployment.yaml`, `k8s/base/service.yaml`, `k8s/base/kustomization.yaml`
**Modifies**: None
**Details**:
- `k8s/base/deployment.yaml`:
  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: amortization-api
    labels:
      app: amortization-api
  spec:
    replicas: 1
    selector:
      matchLabels:
        app: amortization-api
    template:
      metadata:
        labels:
          app: amortization-api
      spec:
        securityContext:
          runAsNonRoot: true
          runAsUser: 65534
        containers:
          - name: server
            image: amortization-api:latest
            ports:
              - containerPort: 8080
            resources:
              requests:
                cpu: 100m
                memory: 64Mi
              limits:
                cpu: 250m
                memory: 128Mi
            livenessProbe:
              httpGet:
                path: /healthz
                port: 8080
              initialDelaySeconds: 5
              periodSeconds: 10
            readinessProbe:
              httpGet:
                path: /healthz
                port: 8080
              initialDelaySeconds: 2
              periodSeconds: 5
  ```
- `k8s/base/service.yaml`:
  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: amortization-api
  spec:
    selector:
      app: amortization-api
    ports:
      - port: 8080
        targetPort: 8080
        protocol: TCP
  ```
- `k8s/base/kustomization.yaml`:
  ```yaml
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  resources:
    - deployment.yaml
    - service.yaml
  ```
**Verification**: `kubectl kustomize k8s/base/` renders valid YAML (or visual inspection)

### Step 9: Kubernetes Overlays
**Goal**: Create environment-specific Kustomize overlays for local-dev, dev, stage, and prod
**Creates**: `k8s/overlays/local-dev/kustomization.yaml`, `k8s/overlays/dev/kustomization.yaml`, `k8s/overlays/stage/kustomization.yaml`, `k8s/overlays/prod/kustomization.yaml`
**Modifies**: None
**Details**:
- `k8s/overlays/local-dev/kustomization.yaml`:
  ```yaml
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  namespace: default
  resources:
    - ../../base
  images:
    - name: amortization-api
      newTag: local
  patches:
    - patch: |
        apiVersion: apps/v1
        kind: Deployment
        metadata:
          name: amortization-api
        spec:
          template:
            spec:
              containers:
                - name: server
                  imagePullPolicy: Never
      target:
        kind: Deployment
        name: amortization-api
  ```
- `k8s/overlays/dev/kustomization.yaml`:
  ```yaml
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  namespace: dev
  resources:
    - ../../base
  images:
    - name: amortization-api
      newName: <registry>/amortization-api
      newTag: dev-latest
  ```
- `k8s/overlays/stage/kustomization.yaml`:
  ```yaml
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  namespace: stage
  resources:
    - ../../base
  images:
    - name: amortization-api
      newName: <registry>/amortization-api
      newTag: stage-latest
  patches:
    - patch: |
        apiVersion: apps/v1
        kind: Deployment
        metadata:
          name: amortization-api
        spec:
          replicas: 2
      target:
        kind: Deployment
        name: amortization-api
  ```
- `k8s/overlays/prod/kustomization.yaml`:
  ```yaml
  apiVersion: kustomize.config.k8s.io/v1beta1
  kind: Kustomization
  namespace: prod
  resources:
    - ../../base
  images:
    - name: amortization-api
      newName: <registry>/amortization-api
      newTag: prod-latest
  patches:
    - patch: |
        apiVersion: apps/v1
        kind: Deployment
        metadata:
          name: amortization-api
        spec:
          replicas: 3
          template:
            spec:
              containers:
                - name: server
                  resources:
                    requests:
                      cpu: 250m
                      memory: 128Mi
                    limits:
                      cpu: 500m
                      memory: 256Mi
      target:
        kind: Deployment
        name: amortization-api
  ```
**Verification**: `kubectl kustomize k8s/overlays/local-dev/` renders valid YAML with namespace=default and imagePullPolicy=Never

### Step 10: Deploy Script
**Goal**: Create local deployment script for minikube
**Creates**: `scripts/deploy-local.sh`
**Modifies**: None
**Details**:
- File: `scripts/deploy-local.sh`
- Must be executable (`chmod +x`)
- Script content:
  ```bash
  #!/usr/bin/env bash
  set -euo pipefail

  # Start minikube if not running
  if ! minikube status | grep -q "Running"; then
      echo "Starting minikube..."
      minikube start
  fi

  # Use minikube context
  kubectl config use-context minikube

  # Build image in minikube's Docker daemon
  eval $(minikube docker-env)
  docker build -t amortization-api:local .

  # Apply local-dev overlay
  kubectl apply -k k8s/overlays/local-dev

  # Wait for rollout
  kubectl rollout status deployment/amortization-api --timeout=60s

  echo ""
  echo "Deployment successful!"
  echo "To access the service, run:"
  echo "  kubectl port-forward svc/amortization-api 8080:8080"
  ```
**Verification**: `bash -n scripts/deploy-local.sh` (syntax check passes)

## Dependency Graph
- Step 1: no dependencies
- Step 2: depends on Step 1
- Step 3: depends on Step 2
- Step 4: depends on Step 1 (uses go.mod but not generated code)
- Step 5: depends on Step 3 and Step 4
- Step 6: depends on Step 5
- Step 7: depends on Step 6
- Step 8: no dependencies (pure YAML)
- Step 9: depends on Step 8
- Step 10: depends on Step 7 and Step 9
