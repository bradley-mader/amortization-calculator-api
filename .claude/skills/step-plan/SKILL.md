---
name: step-plan
description: Generates a detailed implementation plan from a high-level plan step.
allowed-tools:
  - Bash
  - Read
  - Glob
  - Grep
  - Write
argument-hint: "<step_to_plan> ||| <step_description> — separate the step from the step description with |||"
---

# Step Plan Skill

You are a senior Go architect creating a detailed, actionable implementation plan for a single step.

## Inputs

`$ARGUMENTS` contains two parts separated by `|||`:
1. **High-level step to plan** (everything before `|||`)
2. **Step description** (everything after `|||`, e.g., `001_setup_database`)

Parse these two parts. Trim whitespace from both.

## Process

1. **Understand the step**: Read the high-level step description carefully. Read `doc/spec.md` to understand the full project context and how this step fits in.

2. **Explore existing code**: Use Glob and Read to understand what already exists. Check:
   - `go.mod` for existing dependencies
   - Existing Go files for patterns, naming conventions, and imports
   - Existing proto files for message/service definitions
   - Existing test files for testing patterns
   - `doc/audit/` for previous implementation decisions

3. **Create the plan**: Break the step into atomic, ordered sub-tasks. Each sub-task should be:
   - Small enough to implement without ambiguity
   - Specific about which file to create or modify
   - Clear about what code to write (include function signatures, struct definitions, key logic)
   - Aware of dependencies on other sub-tasks

4. **Write the plan**: Create the directory `doc/audit/<step_description>/` if it does not exist. Write to `doc/audit/<step_description>/Implementation_Plan.md`:

```
# Implementation Plan — <step_description>

**Date**: <current date>
**High-level step**: <the original step text>

## Prerequisites
- <what must exist before this step can execute>

## Sub-tasks

### 1. <sub-task title>
**File**: <path to file to create or modify>
**Action**: Create / Modify
**Details**:
<Specific description of what to write. Include:
- Function/method signatures
- Struct definitions with field types
- Key algorithm logic or pseudocode
- Import paths needed
- Error handling approach>

### 2. <sub-task title>
...

## Verification
- <how to verify the implementation is correct — specific commands to run>

## Notes
- <any caveats, alternatives considered, or things to watch out for>
```

5. **Return the plan**: Output the full contents of the plan markdown.
