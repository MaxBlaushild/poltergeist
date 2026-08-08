# Vampire Ascendancy — Mystery Requirements

Implementation plan only — nothing in this doc is built yet. Mirrors how
[MULTI_TENANT_REQUIREMENTS.md](./MULTI_TENANT_REQUIREMENTS.md) was written before
that work started: decisions first, then data model, then a phased build.

## What "Mystery" means

Today the shared content library (characters, houses, items, quiz questions) is a
single undifferentiated pool — there is one story ("The Crimson Toast"), so nothing
distinguishes "this content belongs to this story." A **Mystery** makes the story
itself a first-class, authorable, reusable piece of content, separate from the
characters who might appear in it:

- **Summary** — a short logline (used in the Host's mystery picker).
- **Full lore** — the detailed "what actually happened" writeup, for the Host/GM's
  own reference (never shown to players directly).
- **Beats** — an ordered list of discoverable facts about the mystery ("Valen Drear
  was poisoned, not stabbed", "the poison was meant for Thorne Virell", …). Beats are
  the mystery's atomic units of revealed truth.
- **Quiz** — Part 1 (open-end) and Part 2 (multiple choice) questions, scoped to this
  mystery instead of shared globally.

A character's **secrets** become mystery-scoped too: the same character (say, Serel
Nox) can be cast in multiple mysteries over time, and has a different set of secrets
in each one — because the secrets are what that character knows about *that*
mystery's beats. Each secret is tied to exactly one beat (which beat it reveals).
A character's **post-Act-1 context** is mystery-scoped for the same reason (a
follow-up decision, made after this doc's original phase shipped — see the
`vampire_character_mystery_contexts` migration). Pre-event bio and missions stay
global to the character, not per-mystery.

Each **instance** ("Toast") has exactly one mystery, chosen once at creation and
fixed for the life of that instance (see "Decisions already made").

## Decisions already made

Resolved via clarifying questions before writing this plan:

1. **Scope of "mystery-ness"** — only secrets are mystery-scoped. A character's
   pre-event bio, post-act bio, and missions stay global to the character across
   every mystery they appear in. This is the smaller, more contained version of the
   two options considered; if it turns out a character needs a meaningfully
   different bio per mystery later, that's a follow-up, not blocking this phase.
2. **Mystery lock-in** — a Host picks the mystery once, at "Host a Toast" time. It
   cannot be changed afterward. This sidesteps the real mess a later change would
   create (already-invited players and quiz answers pointing at secrets/questions
   that no longer apply) at the cost of "start a new Toast" being the fix if a Host
   picks wrong — acceptable given Toasts are cheap to create.
3. **Character eligibility gating** — "configured secrets for the mystery" means at
   least one `vampire_secrets` row for that `(character, mystery)` pair. A character
   with zero secrets authored for a given mystery can't be invited to an instance
   running that mystery; the Invites tab's character picker filters them out the
   same way it already filters out already-cast characters.
4. **Beats ↔ secrets cardinality** — many-to-one, not one-to-one. Multiple
   characters' secrets can point at the same beat (several people know the same
   fact, or know different angles on it); a beat has no requirement to be covered
   by any specific number of secrets.

## Current state, for reference

- `vampire_characters` — global, shared across every instance. `PreEventInfo`,
  `PostAct1Context` are single strings on the character row.
- `vampire_secrets` — `{ id, character_id, ordinal, body }`. Flat list per
  character, no story/mystery concept.
- `vampire_quiz_questions` — global, shared across every instance. Today's Content
  tab used to let a Host trim which questions applied per instance
  (`vampire_instance_quiz_questions`); that toggle was removed from the UI this
  session (quiz treated as fixed) but the join table and its backing code are still
  there, unused.
- `vampire_instances` — no story/content reference at all; every instance draws
  from the same single global pool.
- Super Admin dashboard (`/admin`) has Characters / Houses / Items / Quiz / Super
  Users tabs. The Quiz tab edits **the** quiz — one global Part 1 prompt + Part 2
  set, wholesale-replaced on save (`adminUpdateQuizQuestions`).
- Player-facing `GET /me` (`me.go`) returns `character.secrets` straight from
  `GetCharacterByID`'s preload — all of a character's secrets, unfiltered, because
  there's only one set today.

## Data model changes

### New tables

```sql
CREATE TABLE vampire_mysteries (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  name TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  full_lore TEXT NOT NULL DEFAULT '',
  active BOOLEAN NOT NULL DEFAULT TRUE  -- hide from the "Host a Toast" picker without deleting
);

CREATE TABLE vampire_mystery_beats (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  mystery_id UUID NOT NULL REFERENCES vampire_mysteries(id) ON DELETE CASCADE,
  ordinal INT NOT NULL DEFAULT 0,
  body TEXT NOT NULL DEFAULT ''
);
```

### Existing tables changed

```sql
-- Secrets become (character, mystery) scoped, and point at the beat they reveal.
ALTER TABLE vampire_secrets ADD COLUMN mystery_id UUID REFERENCES vampire_mysteries(id);
ALTER TABLE vampire_secrets ADD COLUMN beat_id UUID REFERENCES vampire_mystery_beats(id);
-- ordinal becomes "ordinal within this character's secrets for this mystery"
-- (unique index moves from (character_id) implicit-order to
-- (character_id, mystery_id, ordinal) if we want to enforce it, but the
-- existing code never enforced uniqueness on ordinal — no change needed there).

-- Quiz questions belong to exactly one mystery.
ALTER TABLE vampire_quiz_questions ADD COLUMN mystery_id UUID REFERENCES vampire_mysteries(id);

-- An instance is one mystery, for its whole life.
ALTER TABLE vampire_instances ADD COLUMN mystery_id UUID REFERENCES vampire_mysteries(id);
```

`mystery_id`/`beat_id` start nullable to make the backfill (below) straightforward,
then get backfilled and — for `vampire_quiz_questions.mystery_id` and
`vampire_instances.mystery_id` — set `NOT NULL` in a follow-up migration once the
backfill's confirmed. `vampire_secrets.mystery_id` stays nullable indefinitely: a
secret with no mystery is simply not usable by any instance (it fails the "has a
secret for this mystery" eligibility check), which is a fine terminal state for,
e.g., a secret imported from old data that nobody's gotten around to assigning yet.
`beat_id` also stays nullable — forcing every secret to reference a beat at write
time is a product call the Super Admin editor can enforce in its own validation
(required field in the form) without a hard DB constraint blocking edge cases.

### Retired

`vampire_instance_quiz_questions` (and the already-unused-by-UI
`gmListLibraryQuizQuestions`/`gmSetQuizQuestionIncluded`/`SetQuizQuestionIncluded`/
`ListLibraryQuizQuestions` code paths) go away entirely. An instance's quiz is now
simply "every quiz question whose `mystery_id` matches this instance's
`mystery_id`" — no per-instance toggle layer needed, which is the natural
conclusion of this session's earlier "quiz isn't customizable per instance" change.
`SeedInstanceLibrary` drops its quiz-question-seeding step accordingly.

`vampire_instance_characters` is unaffected by this change (it already only still
carries `image_url`; `included`/`sigil` were deprecated earlier this session) and
keeps existing as-is.

## Migration plan (preserving live data)

1. Create `vampire_mysteries` + `vampire_mystery_beats`. Insert one row for the
   live story: name "The Crimson Toast", summary/full_lore backfilled from
   whatever prose currently exists in the docs/seed source (a human writes this —
   not mechanically derivable from existing DB columns, since there's no
   "mystery-level summary" text anywhere today).
2. Add the nullable `mystery_id`/`beat_id` columns above.
3. Backfill: every existing row in `vampire_quiz_questions`, `vampire_secrets`, and
   `vampire_instances` gets `mystery_id` set to the one row from step 1. No beats
   exist yet to backfill `vampire_secrets.beat_id` against — that's a follow-up
   authoring pass in the Super Admin dashboard after this ships (write the Crimson
   Toast's beats, then go through and tag each existing secret with the beat it
   reveals). Ship with `beat_id` null on all existing secrets rather than block the
   migration on writing that content first.
4. Once confirmed, `NOT NULL` on `vampire_quiz_questions.mystery_id` and
   `vampire_instances.mystery_id` in a follow-up migration (same two-step pattern
   used earlier this session for other backfilled columns).

## API surface changes

**New — Super Admin (`/admin`, super-user only):**
- `GET/POST /admin/mysteries` — list / create.
- `GET/PUT /admin/mysteries/:id` — full editor payload: name, summary, full_lore,
  active, beats (ordered list, replaced wholesale on save — same pattern
  `adminUpdateCharacter` already uses for secrets/missions).
- `GET/PUT /admin/mysteries/:id/quiz` — same shape as today's
  `adminGetQuizQuestions`/`adminUpdateQuizQuestions`, scoped to one mystery instead
  of being the single global quiz.
- `GET/PUT /admin/mysteries/:id/characters/:characterId/secrets` — a character's
  secrets *for this mystery*: list, each with `body` + `beatId`. This is where
  secret-authoring actually happens — from the mystery's side, not the character's
  (see "Super Admin UI" below for why).

**Changed:**
- `POST /vampire-ascendancy/instances` (`createInstance`) — body gains a required
  `mysteryId`. Validates the mystery exists and is `active`.
- `gmListCharacters` (Invites tab picker) — filters to characters with at least one
  secret for the instance's mystery, same place the existing "already taken"
  filter lives (`InvitesSection.tsx` cross-references `takenCharacterIds`; this
  becomes a second exclusion set, `noSecretsForMysteryIds`, filtered server-side so
  the frontend doesn't need the full secrets table).
- `CreatePlayerInvite` (`vampire_player_invite.go`) — defense-in-depth check
  mirroring the frontend filter: reject with a `*ConflictError` if the target
  character has no secret for the instance's mystery, same "don't trust the
  client for a rule that matters" posture `gmRecordGameResult`'s
  `GetActivePlayerByCharacterID` check established last session.
- `getMe` (`me.go`) — `character.secrets` now needs to be resolved scoped to the
  player's instance's mystery instead of "all of this character's secrets." New DB
  method `ListSecretsForCharacterAndMystery(ctx, characterID, mysteryID)`, used
  here instead of the existing all-secrets preload on `GetCharacterByID`.
  `GetCharacterByID` itself is unchanged (still preloads everything) — it's used
  by the Super Admin editor, which legitimately wants to see a character across
  every mystery they're in.
- Quiz serving (`quiz.go`, `gm_quiz.go`) — `GetIncludedPart1Question`/
  `ListIncludedQuizQuestionsByPart` (instance-scoped via the join table) replaced
  with `GetPart1QuestionForMystery`/`ListQuizQuestionsByMysteryAndPart`
  (mystery-scoped via the new column), resolving `instanceID → mysteryID` once at
  the top of each handler via the instance row already in context.

## Frontend changes

**Super Admin dashboard** gains a **Mysteries** tab (`SuperAdminMysteries.tsx`),
following the same list-then-expand-to-edit pattern `SuperAdminCharacters.tsx`
already uses:

- List: name, summary, active toggle, beat count, quiz question count.
- Editor, per mystery:
  - Name / Summary / Full lore fields.
  - Beats: an ordered add/remove/reorder list (same `ListEditor` pattern secrets
    and missions already use in the character editor).
  - Quiz: reuses the existing `SuperAdminQuiz.tsx` editor UI, parameterized by
    mystery instead of hitting the global endpoint.
  - **Character secrets**: a picker over playable characters (reusing
    `CharacterBrowser` from the Invites tab), and for the selected character, a
    secrets list editor where each secret is a body + a beat dropdown (populated
    from this mystery's beats). This is deliberately mystery-first, not
    character-first — authoring a mystery means walking its cast and deciding what
    each of them knows, which is a mystery-centric task even though the data ends
    up attached to characters. `SuperAdminCharacters.tsx` drops its Secrets
    `ListEditor` entirely; that editing surface moves here.

**"Host a Toast" flow** (`CreateToast` in `MyToasts.tsx`) gains a mystery picker —
a `CharacterBrowser`-style card list (name + summary), required before the
"Host this Toast" button enables. Once submitted, immutable (see "Decisions
already made" #2) — no mystery field appears anywhere in the GM console's Setup
tab afterward, just a read-only "Mystery: The Crimson Toast" line for reference.

**Invites tab** (`InvitesSection.tsx`) — no structural change; the existing
"characters already taken" filtering gains a second reason a character might be
missing from the picker ("no secrets written for this mystery yet"), surfaced the
same way — filtered out, with the existing empty-state copy adjusted to mention
both reasons.

## Explicitly out of scope (this phase)

- Editing a mystery's beats/secrets from anywhere other than Super Admin (no
  per-instance override of secrets, no Host-level "add a bonus secret" tool).
- Multiple mysteries per instance, or an instance switching mysteries after
  creation (see decision #2).
- Per-mystery character pre-event bios/missions (see decision #1) — post-Act-1
  context was carved out of this and made mystery-scoped in a follow-up (see
  `vampire_character_mystery_contexts`); pre-event bio and missions remain global.
- Enforcing beat coverage (e.g. "every beat must have at least one secret pointing
  at it" as a publish-blocking rule for a mystery) — the Super Admin dashboard
  shows the data but doesn't gate on completeness.
- Migrating/backfilling `vampire_secrets.beat_id` for the live Crimson Toast data —
  ships null, filled in by hand afterward through the new editor.
