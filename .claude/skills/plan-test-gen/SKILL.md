---
name: plan-test-gen
description: Creates a step-by-step test development plan from a project plan and executes it via plan-code-gen.
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - Agent
argument-hint: "<full_plan> — the implementation plan to create and execute a test plan for"
---

# Plan Test Gen Skill

You are a senior Go test engineer creating and executing a comprehensive test development plan.

## Inputs

`$ARGUMENTS` is the full implementation plan to create tests for.

## Process

### Phase 1: Analyze and Plan

1. **Understand the plan**: Read the provided plan carefully. Also read `doc/spec.md` for full project context.

2. **Explore existing code**: Use Glob and Read to understand what code already exists:
   - Read all `*_test.go` files to understand existing test patterns
   - Read all Go source files to understand what needs testing
   - Check `go.mod` for test dependencies (e.g., testify, gomock)
   - Read proto files to understand the API contract

3. **Design the test plan**: Create a step-by-step plan where each step is:
   - **Atomic**: A single agent can execute it in one pass without further breakdown
   - **Self-contained**: Includes all context needed (file paths, function signatures, expected behavior)
   - **Ordered**: Dependencies flow correctly (setup before tests that need it)

   Cover these test categories:
   - **Unit tests**: Pure function tests for business logic (calculator, validation)
   - **Handler tests**: ConnectRPC handler tests using `httptest` and connect clients
   - **Integration tests**: End-to-end tests that start the server and make real requests
   - **Edge cases**: Boundary values, error paths, zero values, negative inputs
   - **Table-driven tests**: Use Go table-driven test pattern for parameterized cases

4. **Write the plan**: Write to `doc/audit/test_plan.md`:

```
# Test Development Plan

**Date**: <current date>
**Source plan**: <first line or title of the input plan>

## Test Strategy
<brief overview of the testing approach and coverage goals>

## Prerequisites
- <any test infrastructure needed — test helpers, fixtures, mocks>

## Steps

### Step 1: <title>
**File**: <path to test file>
**Type**: Unit / Handler / Integration
**Tests**:
- `TestXxx`: <what it validates>
- `TestYyy`: <what it validates>
**Details**:
<Specific enough for an agent to implement:
- Which functions/methods to test
- Input values and expected outputs
- Error conditions to verify
- Table-driven test cases to include>

### Step 2: <title>
...

## Coverage Goals
- All exported functions have tests
- All validation rules have positive and negative tests
- All ConnectRPC error codes are tested
- Edge cases: zero interest rate, single payment term, max payment day
```

### Phase 2: Execute the Plan

After writing the test plan, use the Agent tool to spawn a sub-agent to execute it:

```
Run the /plan-code-gen skill with these arguments:
<paste the full test plan content from Phase 1>
```

Wait for the agent to complete.

### Output

Output a summary of the test plan execution results.
