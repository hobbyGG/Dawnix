---
name: go-dev-standard
description: "Go implementation standard. Use when implementing/refactoring Go service, biz, data, or API code with a plan-first workflow, balanced SOLID, idiomatic gopher style, selective zap logging (service-first; non-service fatal only), and wrapped errors via fmt.Errorf with %w. Never create/modify tests unless explicitly requested."
argument-hint: "Task context and target files"
---

# Go Development Standard

## When To Use
Use this skill when:
- Implementing new Go features
- Refactoring Go business logic or services
- Translating requirements into a concrete coding plan
- Applying team coding standards consistently

## Karpathy-inspired behavioral guidelines

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

### 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" -> "Write tests for invalid inputs, then make them pass"
- "Fix the bug" -> "Write a test that reproduces it, then make it pass"
- "Refactor X" -> "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] -> verify: [check]
2. [Step] -> verify: [check]
3. [Step] -> verify: [check]
```

## Workflow
1. Clarify implementation goal and boundaries.
2. Build a concise implementation plan before coding.
3. Apply SOLID in a balanced way: start simple, abstract only when clearly beneficial.
4. Implement with idiomatic Go style:
   - Keep interfaces small and behavior-focused.
   - Prefer composition over inheritance-like patterns.
   - Keep function responsibilities narrow.
   - Use clear package boundaries and naming.
5. Apply logging policy:
   - Service layer is the primary place for zap logging.
   - Outside service layer, only fatal-level operational logs are allowed.
   - Otherwise return errors upward without logging.
6. Apply error policy:
   - Wrap errors with fmt.Errorf and %w when returning context upward.
   - Prefer actionable context in error messages.
7. Do not create or modify test code unless the user explicitly asks.
8. Do not use any TrimpSpace() or similar functions that modify the input string in any way. 
9. Run targeted checks (build or lint) relevant to changed code.

## Decision Points
- If a change has no operational value, skip logging.
- Non-service logging is fatal-level only; otherwise return wrapped errors.
- If an error crosses layer boundaries, wrap with fmt.Errorf and %w.
- Avoid abstractions that do not improve substitution or decoupling.
- If test intent is ambiguous, do not change test code.

## Completion Checks
- A plan exists and was followed.
- Design choices reflect balanced SOLID usage.
- Code reads as idiomatic Go.
- Logging is service-first; non-service logging is fatal-level only.
- Errors are wrapped correctly when propagated.
- No test code changes were made unless explicitly requested.
