# Platform findings (R-0.3) — what reef's build actually made reusable

Written after building bgi-site's vertical slice (board-game insert
organizers, Terraforming Mars only), the second product built on top of
reef-site's platform. This is the honest answer to R-0's real question:
does this repo have a reusable print-and-slice platform, or did reef-site
just happen to work once? Read this alongside `go/reef-site/INVENTORY.md`,
which answers the equivalent question for reef's own build.

## The short version

Reef's build produced a real, reusable platform for the parts of the
pipeline that don't know what they're printing: OpenSCAD rendering,
slicing, geometry hashing, pricing arithmetic, bounding-box math, and the
`Module`/`Analyze`/`ValidateParams` interface itself. bgi's `BgiCardTray`
module slotted into that interface with zero changes to any of those
packages — the strongest possible evidence the abstraction is real.

It did **not**, and by this repo's own conventions was never going to,
produce a reusable data model, frontend, or business-logic layer. Those are
one-off per domain everywhere in this codebase (`quest_*`, `zone_*`,
`spell_*` don't share models either), and bgi is no exception.

bgi's own contribution back to the platform is `go/pkg/reef/set` —
composing several independently generated parts into one validated, priced
product. Reef's single-part model never needed this. The next multi-part
microsite, if there is one, can reuse it the way bgi reused
`generate`/`validate`/`pricing`.

## Reusable as-is, zero changes

- `go/pkg/reef/generate` — the `Module` interface and registry. Already
  proven multi-product by FragRack/LidClip/ShelfRack before bgi existed;
  `BgiCardTray` is the fourth registration, in the same file, no different
  from adding a fourth item to any other registry.
- `go/pkg/reef/procexec` + `generate/render.go` — the sandboxed OpenSCAD
  subprocess runner.
- `go/pkg/reef/slice`, `go/pkg/reef/geomhash`, `go/pkg/reef/stlbbox`.
- `go/pkg/reef/pricing.Price`/`Shipping` — pure functions over
  weight/time/cents; bgi's job processor calls `pricing.Price` per resolved
  tray with zero changes to the package.
- `go/pkg/aws` (S3 abstraction) — bucket/key are plain strings; bgi uses its
  own bucket (`bgi-site-artifacts`) and key prefixes (`bgi/stl/...`,
  `bgi/preview/...`) with no code change.
- The `go/core` multi-domain router pattern (`Server` interface +
  `SetupRoutes`) — bgi-site mounts in with the same four touch points reef
  uses (`go/core/cmd/server/main.go`, `go/core/internal/server/client.go`).
- The flat-migrations-with-table-prefix convention — bgi's migrations
  continue the same global sequence (`000443`+) with a `bgi_` prefix, no new
  schema, no per-domain migrations directory.

## Reusable in spirit, blocked by a Go compiler boundary — not a design flaw

`go/reef-site/internal/paramschema` and `go/reef-site/internal/fulfillment`
were both fully generic already — genuinely product-agnostic JSON-Schema
validation and a fulfillment adapter interface with no reef-specific logic
inside them — but they lived under `reef-site/internal/`, invisible to any
second Go module by construction. This wasn't a design mistake, just an
untested assumption (nothing outside reef-site ever needed them before).
Fixing it was a pure relocation: moved to `go/pkg/reef/paramschema` and
`go/pkg/reef/fulfillment`, no logic changes beyond `ManualAdapter` gaining a
`SitePrefix` field so its S3 manifest key and operator-email copy stop
hardcoding "reef" — a real bug the move surfaced (every non-reef adapter
call before this would have silently mislabeled itself as a reef order).

`go/pkg/reef/validate`'s `checkSealedVoid` needed one similar
generalization: it was unconditionally wired into the six-rule pipeline
with aquarium-specific rejection copy ("buoyant and unstable underwater"),
which would have been actively wrong for a board-game tray. Fixed with a
`Thresholds.SealedVoidRuleEnabled bool` gate — reef sets it `true`
(magnetic frag rack's vented pockets are a real submerged-buoyancy
concern), bgi sets it `false` (open-top wells have no cavity story at all).

## Not reusable — built fresh for bgi, by this repo's own existing design

- **The entire product/config data model.** `ReefProduct`/
  `ReefConfiguration`/`ReefSliceResult` all have hardcoded `TableName()`s and
  are fixed methods on the shared `db.DbClient` interface. bgi gets its own
  full parallel set — `bgi_games`, `bgi_products`, `bgi_configurations`,
  `bgi_tray_slice_results`, `bgi_orders`, and so on — not because nothing
  generalized, but because this repo's own convention (every domain gets
  its own tables in `public`) never intended otherwise. This is consistent
  with the rest of the codebase, not a gap.
- **Analytics.** `reef_events` has no generic event-sink abstraction
  anywhere in the repo to fall back on. bgi got its own `bgi_events`
  table and event-type constants from scratch, including one genuinely new
  event type (`fit_indicator_shown`) reef never needed.
- **The entire frontend.** `SchemaForm.tsx`'s rendering *algorithm*
  (iterate schema properties, branch on type/enum/x-control) generalized
  perfectly — bgi's copy is structurally identical. But the actual file
  lives inside reef-site's own npm package, hardcoded to reef's Tailwind
  theme (`bg-reef-teal`, ...), with one reef-only control case
  (`tank-select`) baked into its if-chain. No shared design-system package
  exists in this repo for this kind of thing (confirmed:
  `@poltergeist/components` is described elsewhere as "a grab bag, not a
  design system"). bgi got its own themed copy (`colors.bgi.*`, a
  kraft-paper/wood-tone palette instead of reef's aquarium teal/coral) with
  two new control cases (`sleeve-select`, `box-select`) in place of
  `tank-select`.
- **The three.js viewer generalized least of all.** `StlViewer.tsx` is
  hardwired to exactly one mesh — reef never had a multi-part product to
  render. bgi's fit-preview requirement (assembling several trays into one
  box-shaped mental model, R-3.4) needed a materially different component
  (`StackedStlViewer.tsx`: N meshes, Y-stacked by cumulative height), not a
  themed copy. This is the one place "just copy and reskin" genuinely
  wasn't enough.
- **Operator dashboard.** Not built for bgi at all in this slice — reef's
  own dashboard (`Operator.tsx`/`operator.go`) is fully bespoke UI and
  bespoke metrics with no generic shell to plug into, so there was nothing
  to inherit. `bgiEventHandle` deliberately omits `CountByType`/
  `CountRejectionsByRule` for the same reason — added when that work
  actually starts, not before.

## The one genuinely new hard problem: `go/pkg/reef/set`

Reef's model is one product = one generated part. bgi's is one product = N
generated parts that have to collectively fit inside a box (R-3.3). Neither
`generate`, `validate`, nor `pricing` had any notion of "assemble a set" —
that didn't exist anywhere in this repo before bgi.

`set.Assemble` is a pure function (no DB, no subprocess) that, for each
component type in a game's manifest, searches for the fewest trays whose
*individual* height fits the target box — same-type trays sit side by side
within the box's footprint at one shared height; only *different* component
types stack on top of each other. Getting this height-accounting model
right took a real bug fix during testing: the first version summed each
component's height by its tray *quantity* (treating same-type trays as
stacked), which meant splitting a deck into more trays never actually
helped it fit a shallow box — the opposite of the intended effect, caught
by `TestAssemble_NeedsMoreTrays_WhenShallowerBoxForcesASplit` failing.

Two independent caches, deliberately layered rather than one:

- `config_hash` (`set.ConfigHash`) caches the pure selection inputs (game,
  expansions, sleeve, box, color) → "which trays, how many, do they fit,"
  skippable before any geometry work happens.
- `geometry_hash` (reef's existing `geomhash.Hash`, unchanged) caches each
  resolved tray's actual render/slice, independently of which config
  produced it — two different configs that happen to resolve to the same
  52-card standard-sleeve black tray share that tray's STL and price. This
  is *more* cache-effective than reef's single-level scheme, by
  construction, precisely because bgi's product decomposes into shared
  sub-parts reef's never did.

This package is the thing to reuse if a third multi-part microsite ever
gets built — the same way bgi reused `generate`/`validate`/`pricing`.

## Net assessment

The hard, product-agnostic core (subprocess orchestration, hashing, pricing
math, the render/slice/validate pipeline shape) is real, proven-twice-over
shared infrastructure. The data model, frontend, and business logic are
not, and were never going to be, by this repo's own standing conventions —
every domain gets its own. If a third site gets built, expect: near-zero
changes to `go/pkg/reef/{generate,geomhash,pricing,slice,procexec,stlbbox}`;
possibly one more small generalization to `validate` if it needs a rule
none of reef/bgi's products exercise; a full new set of `foo_*` tables and
handlers; a full new themed frontend package; and, if it's also a
multi-part product, reuse of `go/pkg/reef/set` with no changes at all.
