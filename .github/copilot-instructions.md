# Dawnix Copilot Instructions

## Build, test, and lint commands

```bash
# Build all packages
go build ./...

# Run the full test suite
# Note: ./client tests require SMTP_TOKEN and will panic if it is missing.
go test ./...

# Run a single test
# Note: this test sends a real email via SMTP.
SMTP_TOKEN=<your_token> go test ./client -run Test_emailClient_Send -count=1
```

This repository does not define a dedicated lint command (no Makefile targets, no CI workflow, and no lint config in-repo). Use `go build ./...` and `go test ./...` as the standard checks.

## High-level architecture

Dawnix is a workflow/process engine with HTTP APIs, a runtime scheduler, and an async email worker.

1. **Bootstrap and manual wiring**
   - `cmd/server/main.go` loads config from `local.env` (via Viper), initializes zap logger, starts the email worker, and runs Gin server.
   - `cmd/server/manual.go` manually wires dependencies: GORM + Postgres, Redis, data repos, node registry, scheduler, services, and API handlers.
2. **Layered flow**
   - `api/`: Gin handlers bind request DTOs and call service layer.
   - `internal/service/`: request/business validation and orchestration.
   - `internal/biz/`: core engine logic (runtime graph, node handling, scheduler token movement, gateway routing).
   - `internal/data/`: repo and infra implementations (GORM repos, transaction manager, Redis MQ).
   - `worker/`: consumes `email_tasks` stream and sends emails through `client/email.go`.
3. **Core runtime path**
   - Definition creation stores process graph JSON (`structure`) + form schema (`form_definition`).
   - Instance start (`Scheduler.StartProcessInstance`) loads latest definition by `code`, snapshots structure into instance, creates first execution token, and advances token.
   - Task completion (`Scheduler.CompleteTask`) updates task status, merges form data back to instance, then advances token with gateway-specific behavior (fork/join/XOR/inclusive).
   - Email service nodes publish to Redis stream and auto-advance; worker consumes and sends.

## Key repository conventions

1. **Service/biz/data separation is strict**
   - Service layer owns business-facing orchestration and input validation.
   - Biz layer defines interfaces and engine behavior.
   - Data layer provides concrete GORM/Redis implementations.
2. **Transaction propagation uses context**
   - `internal/data/data.go` injects transaction handle into context (`contextTxKey`).
   - Repos must use `repo.db.DB(ctx).WithContext(ctx)` so scheduler transactions remain atomic across repo calls.
3. **Form data is list-based JSON, not a free-form map**
   - Use `biz.FormDataItem` (`key`, `type`, `value`) and helpers:
     - `DecodeFormDataItems`
     - `FormDataItemsToMap`
     - `MergeFormDataItems`
   - Gateway condition evaluation depends on this conversion path.
4. **Routing semantics are encoded in scheduler**
   - XOR gateway: exactly one matched edge (or one default edge).
   - Inclusive gateway: multiple matches allowed; default edge used only when none match.
   - Expressions are evaluated with `govaluate` against mapped form data.
5. **Node extensibility goes through `NodeRegistry`**
   - Add node type constant in `internal/domain/definition.go`.
   - Implement node behavior in `internal/biz/nodes.go`.
   - Register builder in `NewDefaultNodeRegistry` (`internal/biz/executor.go`).
6. **Domain ↔ persistence conversion is explicit**
   - Data models implement `ToDomain()`.
   - Repos use `xxxToPO()` helpers before persistence.
   - Keep conversion logic in repo/model layer, not service/biz.
7. **Error/logging style used in this repo**
   - Wrap propagated errors with `%w` (`fmt.Errorf("context: %w", err)`).
   - Follow the in-repo Go skill guidance: service-first zap logging; return wrapped errors upward from lower layers.

## Environment assumptions used by current code

- Postgres DSN is currently hardcoded in `cmd/server/manual.go` (`localhost:5432`, `user=root`, `password=123`, `dbname=root`).
- Redis address defaults to `127.0.0.1:16379` if not configured.
- CORS currently allows `http://localhost:5173`.
- `SMTP_TOKEN` is required for email worker/client paths.

## MCP servers (workspace)

Workspace MCP config is in `.vscode/mcp.json` with:

- `postgres` via `@modelcontextprotocol/server-postgres` (prompted `pg_url`)
- `redis` via `@modelcontextprotocol/server-redis` (prompted `redis_url`)

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