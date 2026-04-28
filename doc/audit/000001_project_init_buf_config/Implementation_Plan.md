# Implementation Plan — 000001_project_init_buf_config

**Date**: 2026-04-03
**High-level step**: Initialize Go module and configure Buf for proto code generation

## Prerequisites
- Go toolchain installed (1.23+)
- Buf CLI installed
- No existing `go.mod`, `buf.yaml`, or `buf.gen.yaml` in the project root

## Sub-tasks

### 1. Initialize Go module
**File**: `go.mod`
**Action**: Create
**Details**:
- Run `go mod init github.com/bradley-mader/amortization-calculator-api`
- This creates `go.mod` with the correct module path and Go version
- No dependencies are added yet; `go mod tidy` will be a no-op since there are no `.go` files

### 2. Create proto directory
**File**: `proto/` (directory)
**Action**: Create
**Details**:
- Create `proto/` directory as the module path referenced by `buf.yaml`
- Also create `proto/amortization/v1/` to prepare for the proto file in Step 2
- The directory must exist for Buf to recognize the module

### 3. Create buf.yaml
**File**: `buf.yaml`
**Action**: Create
**Details**:
- Use Buf v2 format
- Declare a single module at path `proto`
- Add `buf.build/googleapis/googleapis` as a dependency (needed for `google/protobuf/timestamp.proto`)
- Content:
  ```yaml
  version: v2
  modules:
    - path: proto
  deps:
    - buf.build/googleapis/googleapis
  ```

### 4. Create buf.gen.yaml
**File**: `buf.gen.yaml`
**Action**: Create
**Details**:
- Use Buf v2 format
- Two remote plugins:
  1. `buf.build/protocolbuffers/go` — generates `*.pb.go` files
  2. `buf.build/connectrpc/go` — generates `*connect/*.connect.go` files
- Both output to `gen/` with `paths=source_relative` option
- Content:
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

### 5. Run go mod tidy
**File**: `go.mod`
**Action**: Modify (no-op expected)
**Details**:
- Run `go mod tidy` to ensure the module file is clean
- Since there are no Go source files yet, this should be a no-op

## Verification
- `cat go.mod` — shows `module github.com/bradley-mader/amortization-calculator-api` and a Go version
- `cat buf.yaml` — shows v2 config with proto module path and googleapis dep
- `cat buf.gen.yaml` — shows v2 config with protocolbuffers/go and connectrpc/go plugins
- `test -d proto/amortization/v1` — proto directory exists

## Notes
- The `proto/` directory needs to exist before Buf can operate on the module
- We create `proto/amortization/v1/` proactively to support the next step (proto definition)
- `buf dep update` is deferred to Step 3 (Code Generation) since there are no proto files to resolve yet
- No `.go` files exist yet, so `go mod tidy` is effectively a no-op at this stage
