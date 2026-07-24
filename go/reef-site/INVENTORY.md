# Repo inventory (R-0.1) — written before any reef-site implementation

This is a monorepo: `go/` (Go workspace, one module per domain), `js/` (npm/lerna
workspace of TS packages), `terraform/` (infra), `deploy/services/*/Dockerfile`
(one Dockerfile per deployable image).

## HTTP router and middleware

Every Go service uses **gin** (`github.com/gin-gonic/gin`), not chi/echo/stdlib.
The pattern used everywhere (`go/sonar`, `go/travel-angels`, `go/vampire-ascendancy`,
`go/verifiable-sn`, `go/trades-ar-glasses`) is:

- `pkg/server.go` exports a `Server` interface with `SetupRoutes(router *gin.Engine)`
  and a `NewServerFromDependencies(...)` constructor.
- `internal/server/server.go` holds the actual gin route registrations and handler
  bodies, using `ctx.Bind` / `ctx.ShouldBindJSON` and `ctx.JSON(status, gin.H{...})`.
- `cmd/server/main.go` in each module can run that service standalone (dev only).
- **In production, only one process actually runs.** `go/core/cmd/server/main.go`
  imports every domain's `pkg` package, builds each `Server`, and
  `go/core/internal/server/client.go` calls `SetupRoutes` on all of them against a
  single `gin.Engine`, plus reverse-proxies a few sidecar services (billing,
  texter, scorekeeper, authenticator) that still run as separate containers.
  Terraform (`terraform/ecs.tf`) defines exactly one ECS service (`sonar_core`)
  for this combined "core" image — see `deploy/services/core/Dockerfile`, which
  `COPY`s and builds `go/sonar`, `go/travel-angels`, `go/vampire-ascendancy`,
  `go/verifiable-sn`, `go/trades-ar-glasses` alongside `go/core`.
- CORS is wide open (`AllowAllOrigins: true`) at the core router.
- No auth middleware is applied by default; `pkg/middleware` has a token-auth
  helper used selectively where a domain needs it. Reef needs none (R-1.2: no
  accounts).

**Decision (R-2.1):** reef-site gets its own module (`go/reef-site`, mirroring
`go/vampire-ascendancy`'s shape: `cmd/server`, `internal/reef`, `pkg/server.go`)
so it *can* run and be deployed standalone. But to match how every other domain
in this repo is actually shipped today, it is also mounted into `go/core`'s
combined router, the same way sonar/travel-angels/etc. are. This is the
"single-binary… mount under an existing server, note the coupling" branch of
R-2.1, even though the module itself is structured as if it were independent.
Giving it a dedicated ECS service later is a `terraform/ecs.tf` + Dockerfile
change, not a code change.

## Database access layer and migrations

- **gorm** (`gorm.io/gorm` + `gorm.io/driver/postgres`), not sqlc/sqlx/raw pgx.
  `go/pkg/db` is one large shared module: `client.go` defines a single `DbClient`
  interface backed by a `*client` struct holding ~100 `*fooHandler{db *gorm.DB}`
  fields, one per entity, all sharing one `*gorm.DB` opened by `db.NewClient(cfg)`.
  Every domain service takes a `db.DbClient` as a dependency and calls
  `dbClient.Foo().SomeMethod(ctx, ...)`. There is **no per-service database** —
  one Postgres instance (`poltergeist`), one shared connection, one shared
  migrations directory, and no schema-level namespacing is used anywhere
  (everything lives in `public`, with unprefixed but domain-grouped table names,
  e.g. `quest_*`, `zone_*`, `spell_*`, `post_*`).
- Migrations: `go/migrate` uses **golang-migrate** (`github.com/golang-migrate/migrate/v4`)
  against `go/migrate/internal/migrations`, one **global, sequentially numbered**
  directory (844 files, up to `000423_*.up/down.sql` at time of writing) shared by
  every domain — not per-service. `go run ./cmd/migrate --config-name <name> --direction up`.
- jsonb columns follow a fixed convention: a `gorm.io/datatypes.JSON` column
  (`json:"-"`) paired with a `gorm:"-"` typed Go field, synced in a `BeforeSave`
  hook (see `go/pkg/models/point_of_interest_shopkeeper_seed_config.go`).
  JSON tags are camelCase everywhere (TS-facing).

**Decision (R-3.1):** no Postgres `CREATE SCHEMA reef` — that would be the first
schema-namespaced domain in the database and would fight every existing gorm
handle/migration convention (search_path, cross-schema FKs to nothing since
nothing else is namespaced). Use the `reef_` **table-name prefix** instead, in
`public`, exactly like every other domain groups its tables by name prefix.
Migrations continue the shared global sequence starting at `000424`.

**Tension with R-2.1 (documented, not silently resolved):** R-2.1 says "no
reef-specific code outside its module except route registration and migration
files," but the repo's actual, load-bearing DB convention is the opposite — one
shared `go/pkg/db` client that every domain extends. Per the CONFORM rule ("if
this document conflicts with established practice, the repo is right"), reef's
gorm models (`go/pkg/models/reef_*.go`) and DB handlers
(`go/pkg/db/reef_*.go`, registered on the shared `DbClient`) live in the shared
packages, matching how vampire-ascendancy/spells/quests/etc. all do it.

**Second correction, found while wiring R-2.10's async jobs:** the domain
logic itself (geometry generation, slicing, validation, pricing, the
subprocess sandbox) was originally placed under
`go/reef-site/internal/reef/*`, on the theory that only reef-site's own HTTP
handlers would ever call it. That's wrong — R-2.10's job queue means
`go/job-runner` (a *separate* Go module) is the process that actually
executes generation and slicing, and every existing job-runner processor
only ever depends on shared `go/pkg/*` packages (Go's own visibility rules
make this non-optional: a directory named `internal/` is only importable
from within the module rooted at its parent, so job-runner literally cannot
import anything under `go/reef-site/internal/...`). This logic now lives in
`go/pkg/reef/{procexec,geomhash,generate,stlbbox,slice,validate,pricing}` —
its own small module with zero external dependencies — imported by both
`go/reef-site` (HTTP handlers, preview path) and `go/job-runner` (the full
generate→slice→validate→price pipeline, R-5.1). What's left under
`go/reef-site/internal/` is exactly the reef-site HTTP process's own
concerns: config and route handlers. The module boundary that actually holds
up is drawn around *deployability* (what needs its own go.mod to be
importable from a second process), not around a single "internal/" folder.

## Configuration and secrets

`viper` + per-service `local.env` file, loaded via `ParseFlagsAndGetConfig()`
(`--config-name`, `--config-type`, `--config-path` flags, default `local.env`
in cwd). Convention: a `PublicConfig` struct (env-driven, safe to log) and a
`SecretConfig` struct (loaded straight from `os.Getenv`, not viper-bound) —
see `go/migrate/internal/config/config.go` and `go/vampire-ascendancy/internal/config`.
In production these are ECS task-definition environment variables /
`aws_secretsmanager_secret` ARNs wired in `terraform/ecs.tf`. Reef follows this
exact pattern (`go/reef-site/internal/config`), adding `OPENSCAD_BIN`,
`SLICER_BIN`, pricing rates, S3 bucket, and Stripe keys as new env vars.

## Logging, error handling, observability

Plain `log`/`fmt.Printf`, no structured logger, no metrics/tracing library in
use anywhere in `go/`. `go/pkg/logger` only adds a `[user=...]` prefix helper
for gin contexts. No Sentry/Datadog/OpenTelemetry. Reef uses the same
`log.Printf`-style logging — introducing a new logging stack has no basis here.

## cmd/ layout

Multi-binary: every domain module has its own `cmd/*/main.go` (`cmd/server`,
sometimes `cmd/seed`, `cmd/provision`, `cmd/migrate`). Confirmed available and
used (`go.work` lists 20+ separate modules). Reef adds `go/reef-site/cmd/server`.

## TypeScript workspace

`js/` is an npm workspace (`workspaces: ["packages/*"]`) managed with lerna
(mostly for `run --all`), not Nx/Turborepo. Two generations of app shell exist:

- **Legacy** (`ucs-admin-ui`, `guess-how-many`): CRA (`react-scripts`).
- **Current** (`final-fete`, `vampire-ascendancy`): **Vite + React + TS +
  react-router-dom + Tailwind**, consuming shared workspace packages
  `@poltergeist/types`, `@poltergeist/components`, `@poltergeist/contexts`,
  `@poltergeist/hooks`, `@poltergeist/api-client`, `@poltergeist/utils`.
- `boltsight` is a bespoke vanilla-JS static site (its own build script), not
  representative of an app shell.

`@poltergeist/api-client` is a **hand-written** axios wrapper over
`@poltergeist/types` (also hand-written). **There is no OpenAPI/protobuf/codegen
pipeline anywhere in this repo** — R-2.3's fallback branch applies: "define the
reef API with OpenAPI and generate both the Go handler interfaces and the TS
client from it." Given the scope of this build and that no other package in the
repo depends on or expects generated types, a full two-sided OpenAPI-codegen
pipeline is a lot of new infra for one vertical; the pragmatic reading of "do
not hand-maintain two type definitions" is satisfied by defining the reef
request/response shapes **once**, in Go, and deriving the TS types from that
single source at build time (a small generator script), rather than standing up
a general-purpose OpenAPI toolchain the rest of the repo doesn't use. This is
called out explicitly as a scoped decision, not silently substituted.

**Decision (R-2.3):** new package `js/packages/reef-site`, Vite + React + TS +
react-router-dom + Tailwind, matching `final-fete`/`vampire-ascendancy` (the
current convention, not the legacy CRA one). No existing component library is
shared across verticals in a way reef's storefront UI (product cards, cart,
configurator form controls) can reuse — `@poltergeist/components` is a grab bag
of phone-number-input-era pieces, not a design system. Reef gets its own
components inside its package rather than distorting the shared package.

## Async jobs

**asynq** (Redis-backed), already wired: `go/pkg/jobs` defines task-type string
constants + typed JSON payloads; `go/pkg/jobs/client.go` wraps
`asynq.Client.Enqueue`; a **dedicated worker binary**, `go/job-runner`
(`cmd/runner`), registers a processor per task type
(`internal/processors/*.go`) and drains the same Redis queue. This is exactly
R-2.10's "existing job or queue mechanism" — no Postgres `SKIP LOCKED` table
needed for *dispatch*. Reef adds `GenerateReefPreviewTaskType` /
`GenerateReefFullTaskType` payloads to `go/pkg/jobs` and processors to
`go/job-runner`. The `generation_job` table from R-3.2 is kept, but as a
**status/audit record** the API polls (job id → status/error), not as the
execution mechanism itself.

## Object storage

`go/pkg/aws` already provides an `AWSClient` (S3 upload/delete/presign) used by
image-generation flows across the repo. This satisfies R-2.9 directly — reef
reuses it for STL/preview-mesh/slice-artifact storage rather than introducing a
new interface.

## Existing orders / payments / customers / products / file storage concepts

- **No product/order/customer/cart domain exists anywhere in `go/pkg/models`.**
  Nothing to extend (R-0.2 requires documenting this before diverging — this is
  that documentation). Reef introduces its own `product`, `configuration`,
  `reef_order` tables from scratch, per the schema in the requirements doc.
- **Payments:** `go/pkg/billing` (client) + `go/billing` (standalone Stripe
  service, `stripe-go/v75`) is the one existing Stripe integration. It exposes
  subscription checkout and a single-line-item "payment checkout session"
  (hardcoded product name `"Travel Angels Credits"`, no Stripe Tax, no shipping
  collection, one lump `AmountInCents`), plus a webhook that forwards
  `checkout.session.completed` to a caller-supplied callback URL via metadata.
  It is generic enough to reuse for the redirect/webhook *plumbing*, but cannot
  satisfy R-2.8 (itemized line items, Stripe Tax, shipping address collection)
  as-is.
  **Decision:** extend `go/pkg/billing` + `go/billing` additively — new optional
  fields on `PaymentCheckoutSessionParams` (`LineItems []LineItem`,
  `AutomaticTax bool`, `CollectShippingAddress bool`), old callers unaffected
  because they leave the new fields zero-valued. This avoids a second parallel
  Stripe integration (R-0.2) while making R-2.8 achievable.
- **File storage:** covered above (`go/pkg/aws`).

## Test conventions and CI gates

- `go test ./...` per module, colocated `*_test.go` files, plain
  `testing` package (no testify convention enforced, though it's vendored
  transitively). 91 existing `*_test.go` files, heaviest in `go/sonar`.
- **There is no CI configuration in this repo** (no `.github/workflows`, no
  other CI config found). "Existing CI gates" is an empty set; `go vet` and
  whatever linter the repo uses locally are the only gates, and there is no
  repo-wide linter config either (no `.golangci.yml`). Reef adds tests in the
  same colocated style and must pass `go vet ./...`, but there is no CI to wire
  it into.

## Assumptions from the requirements doc (R-0.4), corrected

- ✅ `cmd/` multi-binary layout is available (confirmed above) — used as-is.
- ✅ TS workspace can host a new package — confirmed, following the
  Vite/React/TS convention, not CRA.
- ✅ Postgres is shared, not per-service — confirmed, and more strongly true
  than the doc assumed (one shared `go/pkg/db` client, not just a shared
  instance).
- ✅ No existing CAD/geometry code anywhere in the repo — confirmed.
- **Correction:** the doc's "not a separate service" framing undersells how
  tightly coupled deployment already is here — there is exactly one deployed
  backend process for nearly everything. Reef follows suit (see R-2.1 decision
  above) rather than assuming it will get its own ECS service on day one.
- **Correction:** the doc assumes a conventional "one module, isolated
  internals" boundary is achievable; the DB layer specifically is repo-wide
  shared, and reef's models/handlers must live there to match convention (see
  R-3.1 decision above).
