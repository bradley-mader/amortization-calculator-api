# Implementation — 000002_proto_definition

**Date**: 2026-04-03

## Requirement
Define the amortization service protobuf schema with messages for payments, schedule entries, request/response wrappers, and the AmortizationService RPC.

## Changes Made
- `proto/amortization/v1/amortization.proto`: Created protobuf schema with package `amortization.v1`, Go package option pointing to `gen/amortization/v1`, five messages (Payment, SchedulePayment, AmortizationSchedule, GetAmortizationScheduleRequest, GetAmortizationScheduleResponse), and one service (AmortizationService) with a single RPC.

## Decisions
- Placed messages in dependency order (Payment and SchedulePayment first, then messages that reference them) for readability
- Added brief doc comments to each message and the service for clarity
- Did not run `buf generate` or `buf dep update` as those belong to the next step (code generation)

## Dependencies Added
- None

## Verification
- `buf lint` passes with no errors
