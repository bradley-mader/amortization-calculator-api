# amortization-calculator-api

A Go service that computes loan amortization schedules. The API is defined in
Protobuf and exposed over [Connect](https://connectrpc.com/), so clients can
call it using Connect, gRPC, or gRPC-Web.

## API

One RPC, defined in [`proto/amortization/v1/amortization.proto`](proto/amortization/v1/amortization.proto):

- `AmortizationService.GetAmortizationSchedule`

### Request

| Field                 | Type        | Notes                                              |
| --------------------- | ----------- | -------------------------------------------------- |
| `principal`           | `double`    | Must be > 0                                        |
| `interest`            | `double`    | Annual rate as a percent, 0–100                    |
| `term`                | `int32`     | Number of monthly payments, 1–600                  |
| `start_date`          | `Timestamp` | Loan start date                                    |
| `payment_day_of_month`| `int32`     | 1–28                                               |
| `skip_first_month`    | `bool`      | If true, first payment is pushed an extra month    |
| `additional_payments` | `Payment[]` | Extra principal payments (date + amount)           |

### Response

- `monthly_payment` — fixed scheduled payment
- `schedule.payments[]` — per-row date, payment amount, principal/interest
  split, and running totals

The schedule terminates early once the remaining principal is paid off (e.g.
when additional payments accelerate the loan).

## Endpoints

- Connect/gRPC handler at the path generated for `AmortizationService`
- `GET /healthz` — liveness/readiness probe (returns 200)
- HTTP/2 cleartext (h2c) is enabled so gRPC clients work without TLS

CORS is allowlisted to `https://bradley-mader.github.io` and
`https://bradley-mader.com`.

## Configuration

| Variable | Default | Purpose          |
| -------- | ------- | ---------------- |
| `PORT`   | `8080`  | HTTP listen port |

## Running locally

```sh
go run ./cmd/server
```

Then call the service with any Connect/gRPC client pointed at
`http://localhost:8080`.

## Tests

```sh
go test ./...
```

## Regenerating Protobuf

Codegen is driven by [buf](https://buf.build/) (`buf.yaml`, `buf.gen.yaml`).

```sh
buf generate
```

Output lands in `gen/`.

## Docker

```sh
docker build -t amortization-api:local .
docker run --rm -p 8080:8080 amortization-api:local
```

## Kubernetes

Manifests live in `k8s/`, organized as a Kustomize base with overlays for
`local-dev`, `dev`, `stage`, and `prod`.

To deploy to a local minikube cluster:

```sh
./scripts/deploy-local.sh
kubectl port-forward svc/amortization-api 8080:8080
```

## Layout

```
cmd/server/            # main entrypoint
internal/amortization/ # calculator + Connect handler
proto/amortization/v1/ # Protobuf API definition
gen/                   # generated Go code (buf output)
k8s/                   # base + overlays
scripts/               # local deployment helpers
```
