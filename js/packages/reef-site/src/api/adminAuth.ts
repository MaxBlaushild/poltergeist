const STORAGE_KEY = 'reef_admin_basic';

// /api/reef/operator/* is gated with HTTP Basic Auth, but a fetch()-based
// SPA never gets the browser's native login prompt for that — the header
// has to be attached by hand, so the entered password lives here for the
// tab's lifetime rather than being re-typed per request.
export function getAdminAuthHeader(): string | null {
  return sessionStorage.getItem(STORAGE_KEY);
}

export function setAdminPassword(password: string) {
  sessionStorage.setItem(STORAGE_KEY, 'Basic ' + btoa(`operator:${password}`));
}

export function clearAdminAuth() {
  sessionStorage.removeItem(STORAGE_KEY);
}
