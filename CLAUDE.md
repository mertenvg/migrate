# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`migrate` is a Go database migration library with built-in rollback support. It is structured around two pluggable interfaces so the core engine stays storage- and source-agnostic.

## Commands

- `make test` — run all tests with coverage (`go test ./... -coverprofile=test-with-coverage.out`)
- `make coverage` — open HTML coverage report
- `make start-postgres` — start a local Postgres container (`testuser`/`testpass`, db `test_db`, port 5432) used by `adapter/postgres` integration tests
- Run a single test: `go test ./adapter/postgres -run TestName -v`

## Architecture

The core engine lives in `migrate.go`. `Migrate.Migrate(ctx)` orchestrates everything via two interfaces defined in `interfaces.go`:

- **Provider** — yields `Migration` values one at a time via `Next()`. Implementation: `provider/files` reads migrations from a directory.
- **Adapter** — owns the target store, transactions, and the record of applied migrations (`Setup`, `List`, `Begin`, `Up`, `Down`, `Commit`, `Rollback`). Implementation: `adapter/postgres`.
- **Migration** — `Name()`, `Up()`/`Down()` as `io.Reader`s, and `Close()`.

The migration algorithm in `migrate.go`:
1. `Setup` the adapter, then collect every `Migration` from the provider into an ordered name list + map.
2. `List` the already-applied migrations from the adapter.
3. `Begin` a single transaction wrapping the whole run.
4. For any applied migration that no longer exists in the provider, call `adapter.Down(name)` — this is the rollback-on-removal behavior the library is built around.
5. For each provider migration not yet applied, call `adapter.Up(name, up, down)`. The adapter is responsible for persisting the down script alongside the applied record so future removals can roll it back.
6. `Commit`. Any error inside the transaction triggers `Rollback` and is wrapped into the returned error.

Names are the unique key for migrations across both interfaces — provider ordering determines apply order, and the adapter must round-trip the down script it was given at `Up` time.

## Supporting packages

- `pkg/reader` — `io.Reader` for SQL files, used by the files provider.
- `pkg/statements` — SQL statement splitter with dollar-quote (`$$ ... $$`) support; used by the postgres adapter to execute multi-statement migrations.
- `examples/files_to_progress` — end-to-end example wiring the files provider to the postgres adapter.

## Adding a new adapter or provider

Implement the relevant interface in `interfaces.go` and pass it via `migrate.WithAdapter` / `migrate.WithProvider` (see `options.go`). The engine itself does not need changes.
