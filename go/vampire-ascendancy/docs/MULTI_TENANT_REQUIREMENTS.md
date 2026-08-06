# Vampire Ascendancy — Multi-Tenant Requirements

Status: implemented 2026-08-05 — schema, Go backend, and frontend all land
and build/test clean. See "Implementation notes" near the bottom for what
shipped, what was deliberately deferred, and one open inconsistency that
still needs a decision.

## Decisions already made

1. **Content stays one shared library.** The 46 characters, houses, items,
   and quiz questions remain a single global "Crimson Toast" catalog.
   Administrators toggle which subset is *included* in their event; they do
   not author new content or edit story text per instance.
2. **Administrators get real accounts.** Reuse `models.User`
   (email/password + Google OAuth, already in the repo) instead of the
   current shared `GM_PASSCODE`. Invites happen by email.
3. **The live event must be preserved.** The currently-running Crimson Toast
   event becomes "instance #1" via a data migration — no loss of players,
   game state, ledgers, or submissions.
4. **Any administrator can manage administrators** — invite and remove,
   not just the Owner (see "Admin management rights" below for the one
   safeguard kept on top of this).
5. **Instances are addressed by ID in the URL**, not a human slug —
   `/vampire-ascendancy/i/:instanceId/...` and `/e/:instanceId/...` on the
   frontend. No slug field.
6. **Players get real accounts eventually, not at launch.** Phase 1 ships
   with today's link + sigil flow unchanged. See "Player accounts" below
   for where that fits once it's built.
7. **Content-toggle guardrail: block, don't just warn.** An admin cannot
   un-include a character or item that already has an active
   assignment/player in that instance; they must reassign/unassign first.

## Current state (single-tenant), for reference

- `vampire_game_state` is a singleton row (`id = 1`) — one act, one unlock
  switch, one active notification, shared globally.
- `vampire_characters`, `vampire_houses`, `vampire_items`,
  `vampire_secrets`, `vampire_missions`, `vampire_quiz_questions` are global
  tables with no owner.
- Players authenticate with a per-character sigil (`character.password`,
  today a global field); GMs share one passcode (`GM_PASSCODE`) plus a
  free-text name for audit attribution.
- Everything is served under one flat `/vampire-ascendancy/*` namespace,
  folded into `core`, with one frontend build pointed at one API base URL.

## Terminology

- **Instance** — one group's run of the game (new top-level entity),
  addressed everywhere by its UUID.
- **Owner** — the user who created the instance. Exactly one per instance.
  Has one distinguishing protection over other admins: cannot be removed
  (only replaced via an explicit ownership transfer) — see "Admin
  management rights" below.
- **Administrator** — a user invited to manage an instance. Full
  game-management powers, identical to the Owner's, including inviting and
  removing other administrators.

## Naming: schema terms vs. what people actually read

The words above (`instance`, `owner`, `administrator`) are internal —
schema, API paths, this doc. They stay put; renaming them would just be
churn against everything already speced below. What changes is the label a
human sees, and it's worth being deliberate here rather than defaulting to
SaaS-speak, because the fiction already gives us a better word: the
in-story event is "The Crimson Toast," so a real group running one of
these is naturally **hosting a Toast**, not "creating an instance."

| Internal term (unchanged) | User-facing label |
|---|---|
| Instance | **Toast** — "My Toasts," "Sarah's 30th Birthday Toast" |
| Create instance | **"Host a Toast"** |
| Owner | **Host** |
| Administrator | **Co-Host** |
| Invite admin | **"Invite a Co-Host"** |
| Transfer ownership | **"Make [name] the Host"** |

Every UI-copy reference elsewhere in this doc (landing page CTA, dashboard,
admin tab labels) uses this table. Pure copy — cheap to change later if it
doesn't land.

## Admin management rights

Any administrator — Owner or invited admin — can invite new administrators
and remove existing ones, including removing each other. The one guardrail:
**the Owner cannot be removed by another administrator.** Without this, an
instance could end up with zero administrators (e.g. two admins remove each
other and then themselves), permanently locking everyone out with no
recovery path. The Owner can still leave/transfer ownership deliberately
(explicit "transfer ownership to another admin, then remove yourself" flow)
— they just can't be unilaterally removed by someone else.

## Data model changes

### New tables

- `vampire_instances` — id (UUID, this is what appears in every URL), name
  (display only, not unique, not URL-bearing), created_by (user_id), status
  (active/archived), created_at/updated_at.
- `vampire_instance_admins` — instance_id, user_id, role (`owner` |
  `admin`), added_at. Exactly one `owner` row per instance, enforced at the
  application layer (transfer-ownership flow swaps which row holds it).
- `vampire_instance_admin_invites` — id, instance_id, email, invited_by,
  token, expires_at, accepted_at (nullable). Supports inviting someone who
  doesn't have an account yet.

### Existing tables scoped to `instance_id`

Add `instance_id` (not null, FK to `vampire_instances`) and filter every
query by it:

`vampire_game_state` (drop the singleton `id=1` pattern — key by
instance_id instead), `vampire_players`, `vampire_games`,
`vampire_notifications`, `vampire_house_favor_ledger`,
`vampire_blood_token_log`, `vampire_mission_submissions`,
`vampire_submission_photos` (via submission → player → instance),
`vampire_quiz_submissions`, `vampire_gm_action_log`.

### New join tables for library selection + per-instance secrets

The catalog tables stay global, but two things are inherently
per-instance and can't live on the global row anymore: (a) whether a
character/item/question is included in a given event, and (b) a
character's sigil and portrait — two different instances running
concurrently must be able to give "Valen Drear" different PINs and
different player-submitted photos.

- `vampire_instance_characters` — instance_id, character_id, included
  (bool), sigil (moved off `vampire_characters.password`), image_url (moved
  off `vampire_characters.image_url`).
- `vampire_instance_items` — instance_id, item_id, included (bool).
- `vampire_instance_quiz_questions` — instance_id, question_id, included
  (bool).
- Houses are not separately toggled — a house is implicitly included if any
  of its characters are included. (Flagged as an assumption; call it out if
  you want whole-house toggles too.)

`vampire_characters.password` / `.image_url` are deprecated in place (kept
temporarily for the migration, dropped in a follow-up once verified).

## Auth changes

- **Player login** — unchanged in shape at launch (character + sigil →
  token), but the sigil now resolves through `vampire_instance_characters`
  for the instance in the URL, and login requires the instance context
  (`instanceId` + `characterId` + `sigil`). Player tokens remain globally
  unique; the resolved player's `instance_id` is double-checked against the
  URL's instance as defense in depth.
- **Admin login** — replaces the shared `X-GM-Passcode` / `X-GM-Name`
  headers with a real authenticated session (reusing whatever session
  mechanism the existing Google OAuth work established). A
  `withInstanceAdmin` middleware resolves `:instanceId` from the URL,
  checks the logged-in user has a row in `vampire_instance_admins` for that
  instance, and sets the instance + role on the request context. Audit log
  entries attribute to the real user instead of a typed-in name.

## Player accounts (future phase)

Phase 1 keeps today's flow exactly as-is: a player link plus a 4-digit
sigil, no account, nothing to sign up for. That's the right call at launch
— it's a walk-up party flow, and adding account creation to the door would
be friction with no immediate payoff.

The natural point to introduce real player accounts is once a player might
plausibly return across *more than one* instance or session — e.g. someone
who plays in multiple groups' events, or wants their character history /
submitted photos to persist instead of living and dying with one token.
Concretely, that means:

- Not required to build alongside this multi-tenant work.
- When built, it should be additive, not a replacement: after a normal
  sigil login, offer an optional "save this to your account" step (email or
  Google, same mechanism as admins) that links the existing
  `vampire_players` row to a `User`. The sigil-only flow keeps working for
  anyone who skips it.
- A logged-in player who has linked their account is where a "my events /
  my characters" history view would live — but that view has no reason to
  exist until this linking step does, so it's deferred with it.

## API surface changes

New instance-scoped namespace replaces the flat one, keyed by instance
UUID:

- Public/player: `/vampire-ascendancy/i/:instanceId/...` (login,
  characters, me, state, missions, quiz, games, inventory, broadcast feed).
- Admin: `/vampire-ascendancy/i/:instanceId/gm/...` (same sub-routes that
  exist today: state, act, submissions, awards, players, characters,
  houses, games, items, player-items, notifications, quiz).

New platform-level (not instance-scoped) routes:

- `POST /vampire-ascendancy/instances` — create an instance; caller becomes
  Owner.
- `GET /vampire-ascendancy/instances` — instances the current user
  administers.
- `POST /vampire-ascendancy/instances/:id/admins/invite` — invite an admin
  by email. Callable by any current administrator of that instance.
- `DELETE /vampire-ascendancy/instances/:id/admins/:userId` — remove an
  admin. Callable by any current administrator; rejected with 403 if the
  target is the Owner.
- `POST /vampire-ascendancy/instances/:id/transfer-ownership` — Owner-only;
  reassigns the `owner` role to another existing administrator.
- `POST /vampire-ascendancy/invites/:token/accept` — accept an invite
  (links/creates the User account, adds the `vampire_instance_admins` row).
- `GET/PUT /vampire-ascendancy/instances/:id/library/characters` (and the
  equivalent for items and quiz questions) — the include/exclude toggles.
  `PUT` rejects un-including a character/item that has an active
  assignment in that instance (409, with the blocking assignment named in
  the error) rather than silently allowing it.

Every existing "gm" handler keeps its current logic but reads/writes
scoped to `ctx.instance_id` instead of hitting singleton/global rows.

## Instance creation flow ("Hosting a Toast")

Today, an "instance" doesn't really exist as something created — there is
one event, and standing it up is two ops-only CLI steps
(`go run ./cmd/seed`, `go run ./cmd/provision`) run by whoever has shell +
DB access. Self-serve creation means both of those need an in-app
equivalent. UI copy below follows the naming table above — "Host a Toast,"
not "create instance."

1. **Who can create one** — any authenticated user, no approval step. This
   is the direct implication of "multiple groups of people should be able
   to start an instance," so it's treated as resolved rather than a build
   gate. (If you'd rather restrict who can spin up new instances, that's a
   small addition — an allowlist/flag on `User` — flag it and I'll add it.)
2. **`POST /vampire-ascendancy/instances { name }`** (behind the "Host a
   Toast" button) —
   - Creates the `vampire_instances` row.
   - Adds a `vampire_instance_admins` row for the creator with role
     `owner`.
   - Seeds `vampire_instance_characters` / `vampire_instance_items` /
     `vampire_instance_quiz_questions` from whatever is currently in the
     global library, with **`included = true` for everything**. A new
     instance starts fully populated — mirrors today's single-tenant
     default — and the admin trims down via the Content tab rather than
     opting in from an empty roster. (Instances created later do *not*
     retroactively pick up characters/items added to the library after
     their own creation — only instances created after that point see the
     addition. Flagged as a minor default, not a hard requirement.)
3. Response redirects the creator into `/e/:instanceId/gm`, landing on a
   setup-oriented path: **Content** (trim the roster to the actual guest
   count) → **Players** (provision seats) → **Admins** (invite co-GMs) →
   **Game** (once ready, start advancing acts).
4. **Player-seat provisioning moves in-app.** `cmd/provision`'s logic — for
   every included, non-optional, `roleType: player` character in this
   instance without a seat yet, generate a token and a fresh per-instance
   sigil, written to `vampire_instance_characters.sigil` — becomes a
   "Provision seats" action on the Players tab (`POST
   /vampire-ascendancy/i/:instanceId/gm/players/provision`). Same
   idempotency as today (only fills gaps, safe to re-run as the roster
   firms up), but the name·sigil·link handout renders/prints in the
   browser instead of a terminal, scoped to that instance.
5. From there, admin usage is unchanged from today: assign a guest's name
   to a character in the Players tab, hand them their character's
   link + sigil.

The global content library itself (adding brand-new characters, editing
story text, regenerating from the master PDF) stays exactly as it is today
— an ops-only step you run outside the app (`make extract` / `make seed`).
Nothing about multi-tenancy changes how the library's source content gets
authored, only how instances select from it.

## Landing page

New, public, unauthenticated — the app currently has no "front door" at
all (a player's `/c/:characterId` link or the GM passcode form are the only
entry points). This becomes the root route.

- **Content**: hero (title, tagline, reuse the existing house-sigil art
  from `public/houses/`), a short "how it works" (pick the story → invite
  your court → hand guests their characters → run the night live), one
  primary CTA.
- **CTA behavior, adapts to who's looking:**
  - Signed out → **"Host a Toast"** starts sign-in (email or Google), then
    drops straight into the create-Toast form — no dead-end back at the
    marketing page after auth.
  - Signed in, zero Toasts → same CTA, skips straight to the create form.
  - Signed in, ≥1 Toast → CTA becomes **"My Toasts"**, opening the
    dashboard (list of Toasts they Host or Co-Host, with "Host another
    Toast" as a secondary action from there).
- **Out of scope for the landing page itself**: player entry. Guests never
  see it — their character link (`/e/:instanceId/c/:characterId`) goes
  straight to sigil entry, same as today, completely bypassing sign-in.

## Frontend changes

- Instance context threaded through the URL: player links become
  `/e/:instanceId/c/:characterId`; the GM console becomes
  `/e/:instanceId/gm`.
- "My Toasts" dashboard (reached from the landing page once signed in):
  list of Toasts the user Hosts or Co-Hosts, "Host another Toast," and a
  per-Toast settings view (name, Co-Host roster, content library).
- `GMAdmin.tsx` gains two tabs: **Content** (toggle characters/items/quiz
  questions in or out of the roster, with the blocked-toggle error surfaced
  inline — "can't remove Valen Drear, assigned to player X") and
  **Co-Hosts** (invite/remove, "Make [name] the Host"). Existing tabs (Game,
  Submissions, Awards, Broadcast, Quiz, Players) keep their current
  behavior, just scoped by the instance in the URL.
- `Login.tsx` and the API client (`api.ts`, `gmApi.ts`) take the instance
  id as a parameter on every call instead of a fixed base path.
- Admin auth UI swaps the passcode/name form for real sign-in (email or
  Google), followed by an instance picker if the user administers more
  than one.

## Migration plan (preserving the live event)

1. Create `vampire_instances` + admin/invite tables. Insert one row for the
   live event ("The Crimson Toast"), owned by you.
2. Backfill `instance_id` onto every scoped table for all existing rows,
   pointing at instance #1.
3. For every existing character/item/quiz question, create an `included =
   true` row in the new instance-join tables for instance #1, copying
   `sigil`/`image_url` off the global character rows into
   `vampire_instance_characters`.
4. Ensure at least one real admin account is invited and accepted for
   instance #1 before the `GM_PASSCODE` path is retired, so admin access
   isn't lost mid-cutover.
5. Decide whether existing bookmarked player links for the live event need
   to keep working unchanged (redirect old flat URLs into
   `/e/<instance-1-id>/...`) or whether a hard cutover is acceptable —
   depends on whether the event is still upcoming/ongoing.

## Explicitly out of scope (this phase)

- Per-instance custom content authoring (new characters/items/story text).
- Billing or payment for creating instances.
- Public discovery/browsing of other instances.
- Real player accounts (see "Player accounts (future phase)" above).

## Implementation notes (2026-08-05)

Everything in this doc shipped: migrations `000454`–`000456`, the retrofitted
Go models/DB client/server (auth, instance/admin/invite routes, Content-tab
library toggles, in-app "Provision seats," `cmd/claim-owner` for the
migration-3 bootstrap step), and the frontend (landing page, sign-in/sign-up
with Google, "My Toasts," "Host a Toast," accept-invite, the GM console's new
Content and Co-Hosts tabs, and instance-scoped routing throughout). `go
build`/`go test` and `tsc --noEmit`/`vite build` all pass.

Deliberately deferred, not blocking:
- **Cutover checklist items 4 and 5** (bootstrap a real owner for the legacy
  instance; decide on redirecting old bookmarked player links) are still
  manual/undecided — do them before retiring `GM_PASSCODE` for real.
- **Automated Co-Host invite email** — `POST /gm/admins/invite` creates the
  invite and returns a link; nothing sends it yet, matching how player links
  already work (copy/hand out manually).
- Pre-existing `react-hooks/set-state-in-effect` ESLint findings across the
  frontend (present before this change, unrelated to it) were left alone.

**Resolved (2026-08-06): super users.** The inconsistency above is fixed.
Editing the shared content library (characters' bios/secrets/missions, house
taglines, the item catalog, quiz questions incl. the answer key) now
requires being a **super user** — a new, separate tier from Host/Co-Host,
granted via `vampire_super_users` (migration `000457`). It's checked by a
new `withSuperUser` middleware, gating a new `/vampire-ascendancy/admin/*`
route group that isn't scoped to any instance. Those editing endpoints were
removed from the per-instance `/gm/*` routes entirely — a Host/Co-Host's
only remaining character-level edit is that Toast's portrait for a
character (`PUT /gm/characters/:id/portrait`); sigils stay
provisioning-generated. The frontend gained a `/admin` Super Admin
dashboard (Characters / Houses / Items / Quiz / Super Users tabs), gated the
same way GMAdmin is (sign in, then checked by consequence). Bootstrapping
the first super user (nobody can grant one through the dashboard until one
exists) is `go run ./cmd/grant-super-user --email you@example.com` — same
shape as `cmd/claim-owner`. Any super user can grant/revoke another from
the dashboard after that.
