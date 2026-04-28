---
name: plan-code-gen
description: Executes a full implementation plan by running step-code-gen for each step, with periodic architectural reviews.
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - Edit
  - Agent
argument-hint: "<full_plan> — the complete multi-step implementation plan to execute"
---

# Plan Code Gen Skill (Top-Level Orchestrator)

You orchestrate the execution of an entire implementation plan, step by step, with periodic architectural reviews.

## Inputs

`$ARGUMENTS` is the full implementation plan containing multiple steps.

## Process

### Phase 0: Parse the Plan

Split the plan into atomic steps that can be handed of to subagents to implement. Preserve the full text of each step including all details and sub-items.

Create a numbered list of steps for tracking progress.

### Phase 1: Execute Steps Sequentially

For each step, use the Agent tool to spawn a sub-agent:

```
Run the /step-code-gen skill with these arguments:
<text of the current step>
```

Wait for each step to complete before starting the next.

**Track the step count** — you need it for Phase 2.

### Phase 2: Architectural Review (every 3 steps)

After every 3rd completed step (steps 3, 6, 9, etc.), perform an architectural review:

1. Use the Agent tool to spawn a sub-agent with this prompt:

```
You are a senior Go architect performing an architectural review of the AmmortizationCalculatorAPI project.

Review the ENTIRE codebase (not just recent changes) for:

1. **Interface Design**: Are interfaces minimal and well-defined? Are they declared where they are used (consumer side), not where they are implemented? Do any interfaces have too many methods?

2. **Idiomatic Go**: Does the code follow Go idioms? Proper error handling with wrapping? Context propagation? Naming conventions (MixedCaps, short variable names, receiver naming)?

3. **Consistent Naming**: Are names consistent across packages? Are similar concepts named the same way everywhere? Do package names follow Go conventions (short, lowercase, no underscores)?

4. **Safety**: Are there race conditions? Is there proper input validation at boundaries? Are resources properly closed/deferred? Any potential nil pointer dereferences?

5. **Testability**: Can all business logic be tested without external dependencies? Are dependencies injected via interfaces? Is there separation between pure logic and I/O?

Read all Go files in the project. Write your findings to doc/audit/architectural_review_after_step_<N>.md where <N> is the plan step number just completed.

Format your review as:

# Architectural Review — After Step <N>

## Findings
### <category>
- **Issue**: <description>
- **Location**: <file(s)>
- **Recommendation**: <what to do>

## Overall Assessment
<1 paragraph summary>

Return the full review content.
```

2. Capture the architectural review output. Then use the Agent tool to spawn another sub-agent:

```
Run the /code-review-assessor skill with these arguments:
<architectural review findings> ||| architectural_review_step_<N>
```

3. If the assessor returns action items, for each action item, spawn a sub-agent:

```
Run the /step-code-gen skill with these arguments:
Architectural review fix: <action item description>
```

Run these sequentially.

### Phase 3: Continue

After any architectural review remediation completes, continue with the next step from the plan.

### Output

After all steps and reviews complete, output a final summary:

```
## Plan Code Gen Complete

### Steps Executed
1. <step description> — completed
2. <step description> — completed
...

### Architectural Reviews
- After step 3: <N> findings, <M> fixed
- After step 6: <N> findings, <M> fixed
...

### Artifacts
All audit documents are in doc/audit/
```
