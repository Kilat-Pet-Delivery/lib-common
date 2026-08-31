# Kilat Pet Delivery - lib-common

Shared Go building blocks every Kilat service imports: JWT auth, config loading, the Postgres connector and migration runner, the Kafka producer and consumer, structured logging, HTTP middleware, health checks, resilience helpers and the response envelope.
Jira project **KPD** - GitHub `Kilat-Pet-Delivery/lib-common` - stack **Go 1.24 - library**. Global rules live in `~/.claude/`;
this file only adds what is specific here.

## Orient here first

- `.claude/memory/project_state.md` - **resume here** (`/continue` reads it, `/recap` rewrites it).
- `README.md` - how to run it. `CHANGELOG.md` - what changed.
- The workspace map: `~/Documents/kilat-pet-delivery/CLAUDE.md`.

## Commands

| Task | Command |
|---|---|
| install | `go mod download` |
| test | `go test ./...` |
| integration tests | none in this repo |
| lint | `gofmt -l . && go vet ./...` |
| build | `go build ./...` |
| migrate | n/a - this repo owns no schema |

Needs no running stack: `go build ./...` and `go test ./...` work on a bare checkout.

## Conventions that differ from the global rules

- **Ticket branches and PRs** - company repo, never commit on `main` (`branch-guard` enforces it).
- This repo is imported by every service through a `replace` directive (`../lib-common`, `../lib-proto`), so CI has to check it out as a sibling - see the reusable workflow in lib-common.
- Protected paths (never edited in place, see `.claude/protected-paths.txt`): `migrations/*.sql`.

## Testing

0 on `main`. The `kafka` package gets its first tests with KPD-64.

## Where things are

- `auth/` JWT - `config/` the viper loader every service internal/config wraps - `database/` connect plus RunMigrations - `kafka/` producer and consumer - `middleware/` - `health/` - `logger/` - `resilience/` - `response/`

## Worth knowing

- Changing anything here changes every service. Nothing points out of this repo, so it builds alone.
