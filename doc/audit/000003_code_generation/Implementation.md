# Implementation — 000003_code_generation

**Date**: 2026-04-03

## Commands Executed

1. `buf dep update` — Updated buf dependency lock (warning: googleapis dep unused since `google/protobuf/timestamp.proto` is a well-known type bundled with protobuf)
2. `buf generate` — Generated Go protobuf and ConnectRPC code via remote plugins
3. `go mod tidy` — Resolved dependencies; Go toolchain auto-upgraded from 1.23.6 to 1.25.8 to satisfy `connectrpc.com/connect@v1.19.1` (requires go >= 1.24.0)

## Files Created
- `gen/amortization/v1/amortization.pb.go` — Go structs for all proto messages, service descriptors
- `gen/amortization/v1/amortizationv1connect/amortization.connect.go` — ConnectRPC client/handler interfaces and constructors

## Files Modified
- `go.mod` — Added dependencies: `connectrpc.com/connect v1.19.1`, `google.golang.org/protobuf v1.36.11`; upgraded Go toolchain to 1.25.8
- `go.sum` — Populated with dependency checksums

## Verification
- `ls gen/amortization/v1/` shows `amortization.pb.go`
- `ls gen/amortization/v1/amortizationv1connect/` shows `amortization.connect.go`
- `go build ./...` succeeds with exit code 0
