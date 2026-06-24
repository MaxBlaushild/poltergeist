# BoltSight Go

Landing page for BoltSight Go, affordable AR note-taking safety eyewear for tradesworkers.

## Scripts

- `npm run dev --prefix js/packages/boltsight` starts the landing page dev server on port `4177`.
- `npm run build --prefix js/packages/boltsight` copies the static site to `dist/`.
- `npm test --prefix js/packages/boltsight` checks the static server scripts.
- `make -C js/packages/boltsight deploy` builds and syncs `dist/` to `s3://trades-ar-glasses`.

## Interest Endpoint

The landing page submits interest leads to the Go backend:

`POST /trades-ar-glasses/interest`

```json
{
  "email": "crew@example.com",
  "trade": "Electrical",
  "crewSize": "11-50"
}
```

Leads are persisted by `go/trades-ar-glasses` through the shared Postgres database. The landing page posts to `https://api.unclaimedstreets.com` by default, matching the Vampire Ascendancy webapp. Set `globalThis.BOLTSIGHT_API_URL` before `main.js` loads to point browser submissions elsewhere, or set `BOLTSIGHT_API_ORIGIN` to change the local dev server proxy target.
