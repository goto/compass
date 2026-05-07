# Compass Code Review Guidelines

This document is a reference for human reviewers and review agents. It captures the architectural contracts, coding conventions, and quality gates enforced in this repository. Any contribution should satisfy every section before approval.

> **Reviewer scope:** Raise comments only for **Critical** and **High** severity findings (defined below). Do **not** flag cosmetic issues — formatting, import ordering, naming style, and similar surface-level concerns are fully enforced by the automated toolchain (`gofumpt`, `gci`, `golangci-lint`). Cosmetic nits clutter the review and duplicate what CI already catches.

---

## Severity Classification

Every finding must be assigned one of the following levels. Only **Critical** and **High** findings warrant a blocking review comment.

| Severity | Definition | Examples |
|---|---|---|
| **Critical** | Correctness or security defect that will cause data loss, a security breach, or a production outage | SQL injection risk, architecture boundary violation, missing error check on a write path, data race, goroutine leak |
| **High** | Logic error or contract violation with significant operational impact | Incorrect gRPC status code mapping, missing changelog storage, transaction used outside `RunWithinTx`, `context.Background()` in a request-scoped function |
| **Medium** | Code smell or sub-optimal pattern that degrades maintainability but does not break behaviour | Unexported helper duplicated across files, complex function approaching (but not exceeding) the linter threshold |
| **Low / Cosmetic** | Style, formatting, naming preference, or minor readability nit | Import order, blank lines, variable name length, comment phrasing |

**Do not raise Medium, Low, or Cosmetic findings as review comments.** If a pattern bothers you, add a `golangci-lint` rule or `gofumpt` config instead so it is enforced automatically for everyone.

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Architecture and Layer Rules](#2-architecture-and-layer-rules)
3. [Pre-Review Checklist](#3-pre-review-checklist)

4. [Go Code Standards](#4-go-code-standards)
5. [Error Handling](#5-error-handling)
6. [Testing Standards](#6-testing-standards)
7. [Database Patterns](#7-database-patterns)
8. [gRPC Handler Patterns](#8-grpc-handler-patterns)
9. [Concurrency Guidelines](#9-concurrency-guidelines)
10. [Observability and Logging](#10-observability-and-logging)
11. [Configuration Changes](#11-configuration-changes)
12. [Protobuf and Mock Generation](#12-protobuf-and-mock-generation)
13. [Common Anti-Patterns](#13-common-anti-patterns)
14. [PR Checklist Template](#14-pr-checklist-template)

---

## 1. Project Overview

**Compass** is a metadata search and discovery service written in Go (`github.com/goto/compass`, Go 1.23).

| Concern | Implementation |
|---|---|
| API protocol | gRPC + REST (grpc-gateway), defined in `goto/proton` |
| Primary store | PostgreSQL via `sqlx` + `squirrel` query builder |
| Search | Elasticsearch 7.x (`olivere/elastic/v7`) |
| Async indexing | Worker queue via `pgq` (postgres-backed job queue) |
| Auth / identity | `goto/salt`, user identity from gRPC metadata headers |
| Observability | OpenTelemetry (traces + metrics) + New Relic |
| Asset versioning | Semantic versioning (`Masterminds/semver/v3`) + changelogs (`r3labs/diff/v2`) |
| Formatting | `gofumpt` + `gci` |
| Linting | `golangci-lint` with `.golangci-prod.toml` (40+ linters) |
| Mocks | `mockery v2` (auto-generated, never hand-written) |
| Protobuf | `buf` toolchain, source at `goto/proton` |

---

## 2. Architecture and Layer Rules `[Critical]`

The codebase is strictly layered. **Violating these boundaries is a Critical-severity finding.**

```
proto/          ← Generated gRPC types (do not edit manually)
internal/server/v1beta1/  ← gRPC handlers — translate proto ↔ domain types
core/           ← Domain logic (asset, user, lineage, discussion, tag, star)
internal/store/ ← Infrastructure (postgres, elasticsearch)
pkg/            ← Shared utilities (no domain knowledge)
cli/            ← CLI entry points (wires everything together)
```

### Mandatory rules

| Rule | Rationale |
|---|---|
| `core/` packages must not import `internal/store/` | Domain must not know about infrastructure |
| `internal/store/` must not import `internal/server/` | Infrastructure must not know about HTTP/gRPC |
| `internal/store/postgres/` must not contain HTTP clients | Repository code must only talk to PostgreSQL |
| `core/` interfaces define the contract; `internal/store/` implements them | Dependency inversion — domain drives infrastructure |
| `cli/` is the only layer that wires concrete implementations to interfaces | CLI is the composition root |
| External I/O (HTTP calls, message queues) belongs in a dedicated adapter, not in a repository | Keeps repositories testable and single-purpose |

**How to check:** If a file inside `internal/store/postgres/` imports `net/http`, `net/url`, or any HTTP client library, that is an architecture violation.

---

## 3. Pre-Review Checklist

Before reading a single line of code, verify these pass locally and in CI.

```bash
# Format (gofumpt + gci import ordering)
make fmt

# Lint (only new issues since parent commit — mirrors CI)
make lint

# Unit + integration tests with race detector
make test

# If mocks were added or removed
make generate

# If protobuf files were changed
make proto
```

CI gate: `lint.yml` runs `golangci-lint --config=".golangci-prod.toml" --new-from-rev=HEAD~1`. The `--new-from-rev` flag means only lines touched by the PR are linted — a PR cannot introduce new lint violations.

---

## 4. Go Code Standards

> These rules are enforced automatically by the linter and formatter. **Do not raise manual review comments for violations in this section** — CI will catch them. They are documented here so authors understand what the toolchain enforces.

### 4.1 Formatting

- All code must pass `gofumpt` (stricter than `gofmt`): no blank lines inside struct definitions, no unnecessary blank lines between chained calls.
- Imports must be grouped by `gci`: **standard library** first, then **everything else**, separated by a blank line. No mixing.

```go
// CORRECT
import (
    "context"
    "errors"

    "github.com/goto/compass/core/asset"
    "github.com/goto/salt/log"
)

// WRONG — gci will fail
import (
    "github.com/goto/compass/core/asset"
    "context"
    "errors"
)
```

### 4.2 No `init()` functions

The `gochecknoinits` linter is enabled. Adding an `init()` anywhere other than blank-import side-effect packages will fail lint.

### 4.3 No package-level variables (except errors and compiled regexps)

`gochecknoglobals` is enabled. Package-level `var` blocks are allowed only for:
- Sentinel errors (`var ErrFoo = errors.New(...)`)
- Compiled regular expressions (`var fooRegexp = regexp.MustCompile(...)`)
- Constructor helpers that pre-build immutable values used only in that file

Mutable global state is never acceptable.

### 4.4 Line length

`lll` is enabled. Lines should stay within the configured limit. Long `//go:generate` lines are exempt.

### 4.5 Complexity

| Linter | Threshold |
|---|---|
| `gocognit` | 20 (cognitive complexity) |
| `cyclop` | cyclomatic complexity enabled |
| `nestif` | deeply nested `if` blocks flagged |

If a function exceeds the threshold, extract a helper or refactor. Do not add a `//nolint` comment without a documented reason.

### 4.6 Naming

- Error types: `FooError` (struct types that implement `error`), sentinel errors: `errFoo` (unexported) or `ErrFoo` (exported) — enforced by `errname`.
- Avoid re-declaring predeclared identifiers (`len`, `cap`, `new`, `true`, etc.) — `predeclared` linter.
- Use `http.StatusOK` not `200`, use standard library constants — `usestdlibvars`.
- Misspellings are caught by `misspell`.

### 4.7 Struct tags

All structs that are serialized (JSON, DB, protobuf) must have complete struct tags. The `musttag` linter enforces this. The `exhaustruct` linter enforces that `internal/server/v1beta1.APIServerDeps` is always fully initialized.

### 4.8 No `TODO`/`FIXME` comments in merged code

`godox` is enabled and will flag `TODO`, `FIXME`, `HACK` comments. All such markers must be resolved or converted to GitHub issues before merge.

### 4.9 Unused/dead code

`unused`, `ineffassign`, `wastedassign`, and `unconvert` are all enabled. Remove dead assignments and unnecessary type conversions.

---

## 5. Error Handling `[Critical / High]`

### 5.1 Always check errors

`errcheck` is enabled with `check-type-assertions = true`. Every error return must be handled. Every type assertion involving `ok` must check the boolean.

```go
// WRONG
val, _ := someMap[key]
result, _ := someFunc()

// CORRECT
val, ok := someMap[key]
if !ok { ... }
result, err := someFunc()
if err != nil { ... }
```

### 5.2 Use `errors.Is` / `errors.As` — never direct comparison or type assertion

`errorlint` is enabled. Wrapped errors must be inspected with the standard library functions.

```go
// WRONG
if err == asset.ErrEmptyID { ... }
if postgresErr, ok := err.(*pgconn.PgError); ok { ... }

// CORRECT
if errors.Is(err, asset.ErrEmptyID) { ... }
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) { ... }
```

### 5.3 Wrap errors at each layer boundary

Add context without hiding the original error. Use `fmt.Errorf("doing X: %w", err)` when crossing a layer boundary. Avoid bare `return err` if the caller will lose context.

```go
// In repository — add context
if err := r.client.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
    return "", fmt.Errorf("inserting asset: %w", err)
}
```

### 5.4 Domain error types

Domain packages define their error types in `errors.go`. Use value-carrying struct errors (e.g., `asset.NotFoundError{AssetID: id}`) for errors that handlers need to map to gRPC status codes.

### 5.5 `nilerr` — never compare `err != nil` when `err` is an `error` interface that wraps nil

Use `errors.Is(err, nil)` or rely on direct assignment. The `nilerr` linter catches this.

---

## 6. Testing Standards

### 6.1 Test package naming

`testpackage` is enabled. Tests **must** live in the `_test` package:

```go
// CORRECT: file core/asset/service_test.go
package asset_test

// WRONG
package asset
```

Exception: when a test needs unexported symbols, place it explicitly in the same package and add a `//nolint:testpackage` with justification.

### 6.2 Table-driven tests

All new test functions must use a table-driven structure that lists test cases in a slice. Each test case must have a `Description` string.

```go
testCases := []struct {
    Description string
    // inputs and expected outputs
}{
    {
        Description: "should return error when repository fails",
        // ...
    },
}
for _, tc := range testCases {
    t.Run(tc.Description, func(t *testing.T) {
        // ...
    })
}
```

### 6.3 Mocks

- All mocks are generated by `mockery v2`. Never write a mock by hand.
- Each `//go:generate mockery ...` directive must use `--with-expecter` to enable type-safe `.EXPECT()` chains.
- Use `.EXPECT()` syntax, not `.On(...)` / `.Return(...)` directly, as the former gives compile-time safety.

```go
// CORRECT
ar.EXPECT().GetAll(ctx, asset.Filter{}).Return([]asset.Asset{}, nil)

// AVOID unless EXPECT() is unavailable
ar.On("GetAll", ctx, mock.Anything).Return([]asset.Asset{}, nil)
```

### 6.4 Goroutine-heavy tests

When testing code that spawns goroutines, use a short but non-zero timeout with `time.After` or a context deadline, and assert on a channel result to avoid flaky sleeps.

```go
// WRONG — waits full 2 seconds even on success
time.Sleep(2 * time.Second)

// CORRECT
select {
case result := <-resultCh:
    assert.Equal(t, expected, result)
case <-time.After(500 * time.Millisecond):
    t.Fatal("timed out waiting for goroutine")
}
```

### 6.5 Test helpers

Mark test helper functions with `t.Helper()` — `thelper` linter enforces this. Helper functions must have the signature `func helperName(t *testing.T, ...)`.

### 6.6 What to test

- Every service method must have at least one `_test.go` counterpart.
- Test both the happy path and error paths (especially repository errors bubbling up).
- Repository integration tests live in `internal/store/postgres/*_test.go` and rely on `dockertest` for a real Postgres instance.
- Elasticsearch integration tests use the `ES_TEST_SERVER_URL` env var.
- `internal/testutils/` is excluded from linting — do not put non-test code there.

### 6.7 Excluded test rules

In `_test.go` files the following linters are relaxed: `dupl`, `gosec`, `lll`, `gocognit`, `goconst`, `exhaustruct`, and `errcheck` for error return values. Still check all other rules.

---

## 7. Database Patterns `[Critical / High]`

### 7.1 Query building

All SQL is built with `github.com/Masterminds/squirrel`. Raw string interpolation in SQL is forbidden unless the string is a pre-validated constant (e.g., column name from a `const` block).

```go
// CORRECT
builder := sq.Select("*").From("assets").Where(sq.Eq{"id": id})
query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
if err != nil { ... }

// WRONG — SQL injection risk
query := fmt.Sprintf("SELECT * FROM assets WHERE id = '%s'", id)
```

### 7.2 Transactions

Use `Client.RunWithinTx` for multi-step operations that must be atomic:

```go
err := r.client.RunWithinTx(ctx, func(tx *sqlx.Tx) error {
    // step 1
    // step 2
    return nil
})
```

Never open a transaction and manually commit/rollback outside of this helper.

### 7.3 Resource cleanup

`sqlclosecheck` and `rowserrcheck` are enabled. Every `*sql.Rows` must be:
- deferred-closed immediately after `Query`
- checked for `rows.Err()` after iteration

```go
rows, err := r.client.db.QueryContext(ctx, query, args...)
if err != nil { return err }
defer rows.Close()

for rows.Next() {
    if err := rows.Scan(&val); err != nil { return err }
}
return rows.Err()
```

### 7.4 Migrations

SQL migrations live in `internal/store/postgres/migrations/` and are embedded via `//go:embed`. Migration files follow `golang-migrate` naming: `NNNN_description.up.sql` / `NNNN_description.down.sql`. Always provide a down migration.

### 7.5 Context propagation

`noctx` is enabled. Every database call must receive the request context, not `context.Background()`. This ensures traces and timeouts flow correctly.

### 7.6 Postgres error handling

Use `pgconn.PgError` (via `errors.As`) to inspect PostgreSQL error codes. Reference `pgerrcode` constants — never compare raw string codes.

```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) {
    if pgErr.Code == pgerrcode.UniqueViolation {
        return errDuplicateKey
    }
}
```

---

## 8. gRPC Handler Patterns `[High]`

### 8.1 File structure

Handlers live in `internal/server/v1beta1/`. Each domain has its own file (e.g., `asset.go`, `lineage.go`, `user.go`). Service interfaces are declared at the top of each file with `//go:generate mockery` directives.

### 8.2 User validation

Every mutating (and most read) endpoints must call `server.ValidateUserInCtx(ctx)` first:

```go
userID, err := server.ValidateUserInCtx(ctx)
if err != nil {
    return nil, err
}
```

Do not access `user.FromContext(ctx)` directly in handlers — always go through `ValidateUserInCtx`.

### 8.3 Error to gRPC status mapping

Map domain errors to gRPC status codes at the handler layer only. Standard mappings:

| Domain error | gRPC code |
|---|---|
| `user.ErrNoUserInformation` | `codes.InvalidArgument` |
| `user.DuplicateRecordError` | `codes.AlreadyExists` |
| `asset.NotFoundError` | `codes.NotFound` |
| Unknown / infrastructure errors | `codes.Internal` with no detail leaked to caller |

Never return raw `err.Error()` for internal errors — this leaks implementation details.

### 8.4 Request validation

Use protobuf-generated `req.ValidateAll()` for field validation before calling the service layer. Return `codes.InvalidArgument` on failure.

### 8.5 `exhaustruct` for `APIServerDeps`

The `exhaustruct` linter requires `APIServerDeps` to always be fully initialized. Any new service dependency added to `APIServerDeps` must be set in every call site that constructs it.

---

## 9. Concurrency Guidelines `[Critical]`

### 9.1 Goroutines must be bounded and cancellable

Never fire a goroutine with `context.Background()` inside a request handler or service method. Use the request context or a context derived from the service's shutdown context.

If a goroutine outlives the request (e.g., async column lineage dispatch), it must:
1. Be tracked in a `sync.Map` (keyed by a unique ID) or a wait group.
2. Have a cancellation function stored that can be called during service shutdown.
3. Be cleaned up in the service's `cancel` function returned by `NewService`.

### 9.2 HTTP clients must have timeouts

Any `http.Client` created in the codebase (including adapter code) must set an explicit `Timeout`. Never use `http.DefaultClient` for outbound I/O — it has no timeout and will leak goroutines.

```go
// CORRECT
client := &http.Client{Timeout: 30 * time.Second}

// WRONG
resp, err := http.DefaultClient.Do(req)
```

### 9.3 `sync.Map` usage

When storing cancellation functions in `cancelFnMap` (a `*sync.Map`), always use a stable, unique key (e.g., asset URN + operation type). Avoid keys built from external inputs without sanitisation.

### 9.4 Race detector

All tests run with `-race` (`make test`). New code must not introduce data races. Pay attention to:
- Goroutines closing over loop variables (use explicit parameter passing in Go < 1.22, or range-over-int in Go 1.22+).
- Shared mutable state accessed from multiple goroutines without synchronisation.

---

## 10. Observability and Logging

### 10.1 OpenTelemetry metrics

Service-layer operations (Upsert, Delete, etc.) should record metrics via `otel.Meter(...)`. Counters are created once in the constructor and stored on the service struct. Never create a meter per request.

```go
// In NewService / constructor
counter, err := otel.Meter("github.com/goto/compass/core/asset").
    Int64Counter("compass.asset.operation")
if err != nil {
    otel.Handle(err) // do not panic, just handle
}
```

### 10.2 Logging

Use `goto/salt/log.Logger` — never `fmt.Println`, `log.Println`, or `os.Stderr` directly in production code. The `log.Logger` interface is injected as a dependency and must be passed down, not stored globally.

### 10.3 Sensitive data

Do not log request/response bodies that may contain PII (email, names, labels). Log IDs and URNs only.

---

## 11. Configuration Changes

Configuration structs are defined in `internal/*/config.go` files and unmarshalled from `compass.yaml` via `mapstructure`. Reference `compass.yaml.example` for the full configuration schema.

When adding a new configuration field:
1. Add the field to the relevant `Config` struct with a `mapstructure:"field_name"` tag. Include a `default:"..."` tag if a sensible default exists.
2. Update `compass.yaml.example` with the new field and a comment.
3. If the field is required, add validation in the `Validate()` method of the config struct.
4. Update the documentation in `docs/docs/configuration.md`.

---

## 12. Protobuf and Mock Generation

### 12.1 Protobuf

Proto source lives in the external `goto/proton` repository. The pinned commit is stored in `Makefile` as `PROTON_COMMIT`. To regenerate:

```bash
make proto
```

Do **not** hand-edit files in `proto/` — they are generated output. If an API change is needed, update the proto in `goto/proton` and bump `PROTON_COMMIT`.

### 12.2 Mocks

Mocks are generated by `mockery`. The generation directive appears at the top of the interface file:

```go
//go:generate mockery --name=Repository -r --case underscore --with-expecter \
    --structname AssetRepository --filename asset_repository.go --output=./mocks
```

To regenerate all mocks:

```bash
make generate
```

Never edit files in `mocks/` by hand. If a mock is out of date with its interface, regenerate it — do not patch the generated file.

---

## 13. Common Anti-Patterns `[Critical / High]`

These patterns have appeared in this codebase before. Reject PRs that introduce them. All are **Critical** or **High** severity — do not raise findings below that threshold.

### 13.1 HTTP client inside a repository

**Problem:** A `*http.Client` instantiated or used inside `internal/store/postgres/` to call an external service.  
**Impact:** Architecture violation; cannot be mocked; bypasses context propagation; often has no timeout.  
**Fix:** Move the HTTP call to a dedicated adapter (e.g., `internal/lineage/client.go`) that implements a `core/` interface. Inject that interface into the service, not the repository.

### 13.2 `http.DefaultClient` with no timeout

**Problem:** Using `http.DefaultClient.Do(req)` or `http.Get(url)` anywhere.  
**Impact:** Goroutine and connection leak if the server is slow or down.  
**Fix:** Always construct `&http.Client{Timeout: N}`.

### 13.3 `context.Background()` in service methods

**Problem:** Using `context.Background()` inside a request-scoped function.  
**Impact:** Bypasses deadlines, cancellations, and trace propagation.  
**Fix:** Pass the incoming `ctx` through. For fire-and-forget goroutines, use a service-scoped context stored at construction time.

### 13.4 Duplicate goroutine blocks

**Problem:** Copy-pasting a goroutine dispatch block across two methods.  
**Impact:** Duplication causes divergent behaviour when one copy is patched; `dupl` will flag it.  
**Fix:** Extract the goroutine logic into a private method and call it from both places.

### 13.5 Hardcoded service or infrastructure names

**Problem:** String literals like `"maxcompute"`, `"optimus"`, or `/api/v1/lineage/columns` scattered inside repository or service code.  
**Impact:** Breaks other services silently; difficult to configure.  
**Fix:** Move such constants to the relevant `Config` struct and inject them via configuration.

### 13.6 Unsafe SQL path construction

**Problem:** Using `fmt.Sprintf` to build SQL path strings (e.g., `JSONB` path operators).  
**Impact:** Potential SQL injection; `gosec` (G201/G202) can flag it.  
**Fix:** Use parameterised queries or `squirrel`'s safe expression builders. For JSONB paths from changelogs, validate the path against a regex before embedding.

### 13.7 Edge dedup key collisions

**Problem:** Building map keys by concatenating strings with `.` or `->` without escaping.  
**Impact:** Two distinct keys can produce the same concatenated string.  
**Fix:** Use a separator that cannot appear in the component values, or use a composite struct as the map key.

### 13.8 Leftover placeholder comments in tests

**Problem:** Test files that contain `// ...existing code...`, `// ...existing test...`, or similar placeholder comments.  
**Impact:** Misleads reviewers; tests may not cover the intended scenario.  
**Fix:** Remove or replace with real test cases.

### 13.9 Silent changelog discard

**Problem:** An `Upsert` path that computes a changelog but never stores it, while the `UpsertPatch` path stores it — creating invisible inconsistency.  
**Impact:** Version history is incomplete for assets upserted via the full-replace path.  
**Fix:** Either store the changelog consistently in both paths, or document explicitly (in code and docs) that the full-replace path intentionally omits changelog.

---

## 14. PR Checklist Template

Copy this into your PR description:

```markdown
## Review Checklist

### Code Quality
- [ ] `make fmt` passes with no diff
- [ ] `make lint` passes (no new issues from `--new-from-rev=HEAD~1`)
- [ ] `make test` passes with `-race` flag
- [ ] No `TODO`/`FIXME` comments left in code

### Architecture
- [ ] No HTTP clients in `internal/store/postgres/`
- [ ] Layer boundaries respected (`core/` ↔ `internal/store/` dependency direction)
- [ ] New external I/O has a timeout
- [ ] New goroutines are tracked and cancellable

### Go Standards
- [ ] Imports ordered: standard library then everything else (gci)
- [ ] No new `init()` functions
- [ ] No new package-level mutable variables
- [ ] Errors use `errors.Is` / `errors.As` (not `==` or type assertions)
- [ ] Errors are wrapped with `%w` at layer boundaries

### Testing
- [ ] Tests are in `package foo_test` (not `package foo`)
- [ ] Table-driven tests with `Description` per case
- [ ] Mocks regenerated with `make generate` if interface changed
- [ ] No `time.Sleep` in tests; goroutine tests use channels with timeout

### Database
- [ ] All SQL built with `squirrel` — no `fmt.Sprintf` in queries
- [ ] `rows.Close()` deferred; `rows.Err()` checked after loop
- [ ] Transactions use `RunWithinTx`
- [ ] DB calls use request context, not `context.Background()`

### Configuration
- [ ] New config fields have `mapstructure` tags
- [ ] `compass.yaml.example` updated
- [ ] Validation added in `Validate()` if field is required

### Protobuf/Mocks
- [ ] Proto files not hand-edited (generated via `make proto`)
- [ ] Mock files not hand-edited (generated via `make generate`)

### Documentation
- [ ] If public API changed, `docs/` updated
- [ ] If new config, `docs/docs/configuration.md` updated
```

---

*Last updated: derived from codebase exploration at compass@main, Go 1.23, golangci-lint v1.64.8.*
