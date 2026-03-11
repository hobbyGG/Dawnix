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
8. Run targeted checks (build or lint) relevant to changed code.

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
