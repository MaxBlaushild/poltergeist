# Vampire Ascendancy — The Crimson Toast

Companion app for the live murder-mystery event. Go service (folded into `core`)
plus a React frontend in `js/packages/vampire-ascendancy`.

## Architecture

- **Backend**: `go/vampire-ascendancy`, folded into `core` — no separate ECS
  task. Served under `/vampire-ascendancy/*` on `api.unclaimedstreets.com`.
- **Frontend**: SPA in `js/packages/vampire-ascendancy` (Vite + React).
- **Data**: shared Postgres, all tables prefixed `vampire_`.
- **Auth**: players authenticate per-character — the `/c/<characterId>` link
  pre-selects a character and the per-character **sigil** (4-digit PIN) is the
  credential. GMs use a shared **passcode** (`GM_PASSCODE`) at `/gm`.

## Migrations (golang-migrate, in `go/migrate/internal/migrations`)

- `000407_create_vampire_ascendancy_tables`
- `000408_add_vampire_character_password`

Apply them against the target DB with the migrate tooling (same as every other
service).

## Local dev

```sh
# from repo root
make deps                      # postgres + redis
# apply migrations 407 + 408 to the local DB (uuid-ossp must be enabled)

cd go/vampire-ascendancy
export DB_PASSWORD=... GM_PASSCODE=...
make seed                      # loads packets + sigils + quiz
make provision                 # creates player seats, prints name · sigil · link
make dev                       # standalone server on :8090 (CORS-enabled for dev)

# frontend (separate shell)
npm run dev --prefix js/packages/vampire-ascendancy
# point VITE_API_URL at the backend (defaults to api.unclaimedstreets.com;
# for local use http://localhost:8090 via .env.local)
```

## Content (data-driven, re-importable)

- `seed/characters.json` — 46 packets. Regenerate from the master PDF:
  `make extract PDF="/path/to/Master File.pdf"`.
- `seed/quiz.json` — end-quiz questions (`multiple_choice` auto-grade and apply
  `hfEffect`; `open` answers are stored for GM review). Edit and re-run `make seed`.
- `make seed` is idempotent: characters upsert by name, **sigils are preserved**
  across re-runs, and the quiz is wholesale-replaced.

## Config / secrets

- `GM_PASSCODE` — GM admin passcode (env var / ECS task secret).
- `DB_HOST/USER/PORT/NAME` + `DB_PASSWORD` — as per other services.

## Deploy (production)

1. **Migrate**: apply `000407` + `000408` to the prod DB.
2. **Seed**: `make seed --config-name live` then `make provision --config-name live`
   (run from somewhere that can reach RDS). Keep the printed `name · sigil · link`
   list — that's the guest handout.
3. **Secret**: set `GM_PASSCODE` on the `core` ECS task.
4. **Backend**: `make core/ecr-push` (repo root) → `make deploy-all` to roll the
   core service. The vampire routes ship inside core.
5. **Frontend**: add the house sigils (below), then `npm run build` in
   `js/packages/vampire-ascendancy` and deploy `dist/` to its static host.
   > Hosting (S3 bucket + CloudFront/DNS) for the frontend still needs to be set
   > up — the module `Makefile` `deploy` target points at
   > `s3://vampire-ascendancy.blaubertech.com` as a placeholder.

## House sigils

Drop the five emblem PNGs into `js/packages/vampire-ascendancy/public/houses/`
(`spires.png`, `chains.png`, `cinders.png`, `ashglass.png`, `court.png`). The app
inverts black-on-white line art automatically. See that folder's README.

## GM console (`/gm`)

Passcode + name gate, then tabs: **Game** (unlock / act / playtest reset),
**Submissions** (verify/reject), **Awards** (House Favor + Blood Tokens),
**Broadcast** (full-screen takeovers), **Quiz** (open/close + review), **Players**
(assign characters, sigils, copy links). Every action is written to
`vampire_gm_action_log`.
