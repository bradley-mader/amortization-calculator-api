---
name: plan-feature-gen
description: Creates a step-by-step feature implementation plan from a project plan and executes it via plan-code-gen.
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - Agent
argument-hint: "<full_plan> — the high-level plan to decompose into atomic steps and execute"
---

# Plan Feature Gen Skill

You are a senior Go architect decomposing a project plan into atomic implementation steps and executing them.

## Inputs

`$ARGUMENTS` is the full high-level plan to decompose and execute.

## Process

### Phase 1: Analyze and Plan

1. **Understand the plan**: Read the provided plan carefully. Also read `doc/spec.md` for full project context.

2. **Explore existing code**: Use Glob and Read to understand the current project state:
   - Read `go.mod` if it exists
   - Read all existing Go source files
   - Read existing proto files
   - Check `doc/audit/` for previous implementation history
   - Read Dockerfile, k8s configs, and scripts if they exist

3. **Decompose into steps**: Break the plan into ordered steps where each step:
   - **Is atomic**: A single agent can execute it in one pass without needing to make judgment calls about scope
   - **Is specific**: Names exact files to create/modify, function signatures, struct definitions
   - **Has clear boundaries**: Start state and end state are unambiguous
   - **Respects dependencies**: Proto before codegen, codegen before handlers, handlers before server, etc.
   - **Is testable**: Each step produces something that can be verified (compiles, passes tests, runs)

   Follow these ordering principles:
   - Infrastructure and configuration first (go.mod, buf.yaml, proto files)
   - Code generation before hand-written code
   - Internal packages before cmd packages
   - Business logic before transport/handler layer
   - Server wiring after all components exist
   - Deployment configs (Dockerfile, k8s) last
   - Each step should produce a compilable/runnable state when possible

4. **Write the plan**: Write to `doc/audit/feature_plan.md`:

```
# Feature Implementation Plan

**Date**: <current date>
**Source**: <first line or title of the input plan>

## Overview
<1-2 sentence summary of what will be built>

## Steps

### Step 1: <title>
**Goal**: <what this step achieves>
**Creates**: <new files>
**Modifies**: <existing files, or "None">
**Details**:
<Detailed enough for an agent to implement without ambiguity:
- Exact file paths
- Package names and imports
- Function/method signatures with parameter and return types
- Struct definitions with field names and types
- Key logic or algorithm description
- Commands to run (go mod tidy, buf generate, etc.)>
**Verification**: <command to verify — go build, go test, go vet, etc.>

### Step 2: <title>
...

## Dependency Graph
- Step 1: no dependencies
- Step 2: depends on Step 1
- Step 3: depends on Step 2
...
```

### Phase 2: Execute the Plan

After writing the feature plan, use the Agent tool to spawn a sub-agent to execute it:

```
Run the /plan-code-gen skill with these arguments:
<paste the full feature plan content from Phase 1>
```

Wait for the agent to complete.

### Output

Output a summary of the feature plan execution results.
