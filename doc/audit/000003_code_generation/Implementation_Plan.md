# Implementation Plan — 000003_code_generation

**Date**: 2026-04-03
**High-level step**: Generate Go protobuf and ConnectRPC code from proto definition

## Prerequisites
- `buf.yaml` (v2) exists with `buf.build/googleapis/googleapis` dependency
- `buf.gen.yaml` (v2) exists with `protocolbuffers/go` and `connectrpc/go` remote plugins, output to `gen/`
- `proto/amortization/v1/amortization.proto` exists with valid proto3 definitions
- `go.mod` exists with module `github.com/bradley-mader/amortization-calculator-api`
- `buf` CLI is installed and available on PATH

## Sub-tasks

### 1. Update buf dependencies
**Action**: Run command
**Details**:
- Run `buf dep update` from the project root
- This resolves and locks the `buf.build/googleapis/googleapis` dependency needed for `google/protobuf/timestamp.proto`
- Expect a `buf.lock` file to be created or updated in the `proto/` directory

### 2. Generate protobuf and ConnectRPC Go code
**Action**: Run command
**Details**:
- Run `buf generate` from the project root
- This invokes the two remote plugins configured in `buf.gen.yaml`:
  - `buf.build/protocolbuffers/go` generates `gen/amortization/v1/amortization.pb.go` containing Go structs for all proto messages and service descriptors
  - `buf.build/connectrpc/go` generates `gen/amortization/v1/amortizationv1connect/amortization.connect.go` containing the ConnectRPC client/handler interfaces and constructors
- Output directory `gen/` will be created automatically
- The `paths=source_relative` option ensures generated files mirror the proto directory structure under `gen/`

### 3. Tidy Go module dependencies
**Action**: Run command
**Details**:
- Run `go mod tidy` from the project root
- This adds required dependencies to `go.mod` and creates/updates `go.sum`:
  - `google.golang.org/protobuf` (protobuf runtime)
  - `connectrpc.com/connect` (ConnectRPC runtime)
  - Transitive dependencies

### 4. Verify generated files
**Action**: Run commands
**Details**:
- `ls gen/amortization/v1/` should show `amortization.pb.go`
- `ls gen/amortization/v1/amortizationv1connect/` should show `amortization.connect.go`
- `go build ./...` should succeed with no errors

## Verification
- `ls gen/amortization/v1/` shows `.pb.go` files
- `ls gen/amortization/v1/amortizationv1connect/` shows `.connect.go` files
- `go build ./...` succeeds with exit code 0

## Notes
- The generated files should NOT be manually edited; they are re-generated from the proto definition
- The `buf.lock` file should be committed to version control for reproducible builds
- Remote plugins are used (no local protoc/plugin installation required), which simplifies CI/CD
