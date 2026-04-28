# Implementation Plan — 000002_proto_definition

**Date**: 2026-04-03
**High-level step**: Define the amortization service protobuf schema

## Prerequisites
- `go.mod` initialized with module `github.com/bradley-mader/amortization-calculator-api` (done in step 000001)
- `buf.yaml` configured with v2 format and `buf.build/googleapis/googleapis` dependency (done in step 000001)
- `buf.gen.yaml` configured with protocolbuffers/go and connectrpc/go plugins (done in step 000001)
- Directory `proto/amortization/v1/` exists (created in step 000001)

## Sub-tasks

### 1. Create the protobuf schema file
**File**: `proto/amortization/v1/amortization.proto`
**Action**: Create
**Details**:
- Set `syntax = "proto3";`
- Set `package amortization.v1;`
- Set `option go_package = "github.com/bradley-mader/amortization-calculator-api/gen/amortization/v1;amortizationv1";`
- Import `google/protobuf/timestamp.proto`
- Define the following messages in order:
  1. `Payment` — used for additional payments input
     - `google.protobuf.Timestamp date = 1;`
     - `double amount = 2;`
  2. `SchedulePayment` — a single row in the amortization schedule
     - `google.protobuf.Timestamp date = 1;`
     - `double total_principal_paid = 2;`
     - `double total_interest_paid = 3;`
     - `double payment_amount = 4;`
     - `double payment_principal = 5;`
     - `double payment_interest = 6;`
  3. `AmortizationSchedule` — wrapper for the schedule
     - `repeated SchedulePayment payments = 1;`
  4. `GetAmortizationScheduleRequest` — RPC request
     - `double principal = 1;`
     - `double interest = 2;`
     - `int32 term = 3;`
     - `google.protobuf.Timestamp start_date = 4;`
     - `int32 payment_day_of_month = 5;`
     - `bool skip_first_month = 6;`
     - `repeated Payment additional_payments = 7;`
  5. `GetAmortizationScheduleResponse` — RPC response
     - `double monthly_payment = 1;`
     - `AmortizationSchedule schedule = 2;`
- Define `service AmortizationService` with:
  - `rpc GetAmortizationSchedule(GetAmortizationScheduleRequest) returns (GetAmortizationScheduleResponse);`

## Verification
- `buf lint` passes with no errors (validates naming conventions, field numbering, package structure)

## Notes
- Message ordering places `Payment` and `SchedulePayment` before the messages that reference them for readability
- All field names use snake_case per protobuf convention; buf lint enforces this
- The `AmortizationSchedule` wrapper message is used instead of inlining `repeated SchedulePayment` directly in the response, keeping the schema extensible
- `buf dep update` is not needed here — it was deferred to the code generation step and googleapis deps were already configured in step 000001
