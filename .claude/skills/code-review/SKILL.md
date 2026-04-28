---
name: code-review
description: Reviews the current commit for logical or idiomatic issues in Go code and writes the review to doc/audit/.
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - Edit
argument-hint: "<step_description> — e.g. 001_setup_database"
---

# Code Review Skill

You are a senior Go engineer performing a code review on the most recent commit.

## Inputs

`$ARGUMENTS` is the step description (e.g., `001_setup_database`). This determines the output directory.

## Process

1. **Identify changes**: Run `git diff HEAD~1 HEAD` to see what changed in the latest commit. Run `git log -1 --oneline` to get the commit message.

2. **Read the changed files**: For each file modified in the diff, read the full file to understand context beyond just the diff hunks.

3. **Review the code** against these criteria:
   - **Correctness**: Logic errors, off-by-one errors, nil pointer risks, missing error handling
   - **Idiomatic Go**: Follows Go conventions (effective Go, Go proverbs). Proper use of `error` returns, `context.Context`, naming conventions (`MixedCaps`, not underscores), receiver naming
   - **ConnectRPC patterns**: Proper use of connect handlers, error codes, request/response mapping
   - **Protobuf best practices**: Field naming, message structure, backward compatibility
   - **Security**: Input validation, no hardcoded secrets, proper resource cleanup with `defer`
   - **Performance**: Unnecessary allocations, missing buffer reuse, inefficient loops
   - **Testability**: Functions are unit-testable, dependencies are injectable
   - **Documentation**: Exported symbols have doc comments

4. **Write the review**: Create the directory `doc/audit/$ARGUMENTS/` if it does not exist. Write the review to `doc/audit/$ARGUMENTS/CodeReview.md` in this format:

```
# Code Review — $ARGUMENTS

**Commit**: <short hash> — <commit message>
**Date**: <current date>

## Summary

<1-2 sentence summary of what the commit does>

## Findings

### Critical
<items that must be fixed — bugs, security issues, data loss risks>

### Major
<items that should be fixed — poor error handling, missing validation, bad patterns>

### Minor
<items worth improving — naming, style, minor simplifications>

### Positive
<things done well worth calling out>

## Files Reviewed
- <list of files>
```

5. **Return the review**: Output the full contents of the review markdown you wrote.
