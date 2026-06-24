# BoltSight

Fake landing page for BoltSight Go, affordable AR note-taking safety glasses for tradesworkers.

## Scripts

- `npm run dev --prefix js/packages/boltsight` starts the landing page dev server on port `4177` and proxies interest submissions to the Go backend.
- `npm run build --prefix js/packages/boltsight` copies the static site to `dist/`.
- `npm test --prefix js/packages/boltsight` checks the static server scripts.

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

Leads are persisted by `go/trades-ar-glasses` through the shared Postgres database. The local static dev server proxies submissions to `http://127.0.0.1:8080` by default; set `BOLTSIGHT_API_ORIGIN` to point it elsewhere.
