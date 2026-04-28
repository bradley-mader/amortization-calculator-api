---
name: step-code-gen
description: Orchestrates the full code generation pipeline for a single implementation step — plan, implement, review, assess, and fix.
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - Edit
  - Agent
argument-hint: "<implementation_step> — a single step to implement end-to-end"
---

# Step Code Gen Skill (Orchestrator)

You are a pipeline orchestrator that drives a single implementation step through planning, implementation, review, and remediation.

## Inputs

`$ARGUMENTS` is the implementation step to execute (e.g., "Project Init & Proto Definition" or the full text of a plan step).

## Process

### Phase 0: Generate Step Description

Determine the step description in the format `<NNNNNN>_<short_description>`:

1. Run `ls doc/audit/ 2>/dev/null` to list existing step directories.
2. Find the highest numeric prefix among existing directories (parse the first 3 characters of each directory name as a number). If none exist or the directory is empty, start at 0.
3. Increment by 1 and zero-pad to 6 digits.
4. Create a short description (5 words or fewer) from the step content, using underscores for spaces, all lowercase. For example: `000001_project_init_proto`, `000002_code_generation`, `000003_business_logic`.
5. The step description is: `<NNNNNN>_<short_description>`

Store this as `STEP_DESC` for use in all subsequent sub-agent calls.

### Phase 1: Plan

Use the Agent tool to spawn a sub-agent with this prompt:

```
Run the /step-plan skill with these arguments:
<paste the full implementation step text> ||| <STEP_DESC>
```

Wait for the agent to complete. Capture the plan output.

### Phase 2: Implement

Use the Agent tool to spawn a sub-agent with this prompt:

```
Run the /code-implementer skill with these arguments:
<paste the full plan output from Phase 1> ||| <STEP_DESC>
```

Wait for the agent to complete.

### Phase 3: Review

Use the Agent tool to spawn a sub-agent with this prompt:

```
Run the /code-review skill with these arguments:
<STEP_DESC>
```

Wait for the agent to complete. Capture the review output.

### Phase 4: Assess

Use the Agent tool to spawn a sub-agent with this prompt:

```
Run the /code-review-assessor skill with these arguments:
<paste the full review output from Phase 3> ||| <STEP_DESC>
```

Wait for the agent to complete. Capture the assessment output.

### Phase 5: Remediate (if needed)

Parse the assessment output. Look for the `## Action Items` section.

- If it says "None" or there are no action items, the step is **complete**. Output a summary and stop.
- If there ARE action items, for EACH action item:
  - Use the Agent tool to spawn a sub-agent with this prompt:
    ```
    Run the /step-code-gen skill with these arguments:
    Fix from code review of <STEP_DESC>: <action item description>
    ```
  - Run these sequentially, not in parallel.

### Output

After all phases complete, output a summary:

```
## Step Code Gen Complete — <STEP_DESC>

### Pipeline Summary
1. Plan: doc/audit/<STEP_DESC>/Implementation_Plan.md
2. Implementation: doc/audit/<STEP_DESC>/Implementation.md
3. Review: doc/audit/<STEP_DESC>/CodeReview.md
4. Assessment: doc/audit/<STEP_DESC>/CR_Assessment.md
5. Remediation: <completed N fix(es) / not needed>
```
