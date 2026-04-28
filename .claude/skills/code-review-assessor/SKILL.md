---
name: code-review-assessor
description: Assesses code review findings and decides which items should be implemented.
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - Edit
argument-hint: "<code_review_content> ||| <step_description> — separate review from step description with |||"
---

# Code Review Assessor Skill

You are a pragmatic engineering lead assessing code review findings to decide which are worth acting on.

## Inputs

`$ARGUMENTS` contains two parts separated by `|||`:
1. **Code review content** (everything before `|||`)
2. **Step description** (everything after `|||`, e.g., `001_setup_database`)

Parse these two parts. Trim whitespace from both.

## Process

1. **Parse the review**: Extract each finding from the code review. Group them by severity (Critical, Major, Minor).

2. **Read current code**: For each finding, read the relevant file(s) to verify the issue actually exists and understand the context.

3. **Assess each item** using these criteria:
   - **MUST FIX**: All Critical items. Bugs, security vulnerabilities, data corruption risks, panics.
   - **SHOULD FIX**: Major items that affect correctness, maintainability, or violate Go idioms in ways that will cause confusion or bugs later.
   - **SKIP**: Minor style issues, subjective preferences, items that would require large refactors with low payoff, items that are actually false positives.

4. **Write the assessment**: Create the directory `doc/audit/<step_description>/` if it does not exist. Write to `doc/audit/<step_description>/CR_Assessment.md`:

```
# Code Review Assessment — <step_description>

**Date**: <current date>

## Items to Implement

### <item title>
- **Original severity**: Critical/Major/Minor
- **Decision**: MUST FIX / SHOULD FIX
- **Reason**: <why this needs to be addressed>
- **Action**: <specific description of what to do>

(repeat for each item to implement)

## Items Skipped

### <item title>
- **Original severity**: Critical/Major/Minor
- **Decision**: SKIP
- **Reason**: <why this can be skipped>

(repeat for each skipped item)

## Summary
- Total findings: <N>
- Items to implement: <N>
- Items skipped: <N>
```

5. **Return the assessment**: Output the full contents of the assessment markdown. Include a clear, parseable list of just the items to implement at the end:

```
## Action Items
1. <concise action description>
2. <concise action description>
...
```

If there are zero action items, explicitly state: `## Action Items\nNone — all findings have been addressed or skipped.`
