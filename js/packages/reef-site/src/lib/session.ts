const SESSION_KEY = 'reef_session_id';

// A per-browser session id, not tied to any account (R-1.2: no login).
// Used to correlate analytics events (R-9.1) and to rate-limit preview
// requests per session (R-2.6).
export function getSessionId(): string {
  let id = localStorage.getItem(SESSION_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(SESSION_KEY, id);
  }
  return id;
}
