# Implementation — 000001_project_init_buf_config

**Date**: 2026-04-03

## Requirement
Initialize the Go module with the correct module path and configure Buf for protobuf code generation using ConnectRPC.

## Changes Made
- `go.mod`: Created via `go mod init github.com/bradley-mader/amortization-calculator-api` (Go 1.23.6)
- `buf.yaml`: Created with v2 format, proto module path, and googleapis dependency
- `buf.gen.yaml`: Created with v2 format, protocolbuffers/go and connectrpc/go remote plugins outputting to `gen/`
- `proto/amortization/v1/`: Created directory structure for future proto definitions

## Decisions
- Used Buf v2 format as specified in the plan
- Created `proto/amortization/v1/` proactively to support the next step (proto definition)
- Deferred `buf dep update` to Step 3 since there are no proto files to resolve yet

## Dependencies Added
- None (go mod tidy was a no-op since no Go source files exist yet)

## Verification
- `cat go.mod` — confirms module path `github.com/bradley-mader/amortization-calculator-api` and Go 1.23.6
- `cat buf.yaml` — confirms v2 config with proto module and googleapis dep
- `cat buf.gen.yaml` — confirms v2 config with both remote plugins
- `test -d proto/amortization/v1` — proto directory exists
