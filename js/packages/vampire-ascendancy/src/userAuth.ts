// Real-account auth for Hosts/Co-Hosts (email/password or Google) — separate
// from the player's per-character sigil token (api.ts) and unrelated to it.
// A signed-in user administers zero or more Toasts; nothing here is scoped
// to any one instance.
import { useCallback, useEffect, useState } from 'react';

export interface AuthUser {
  id: string;
  name: string;
  email?: string;
  phoneNumber?: string;
}

export interface UserAuth {
  token: string;
  user: AuthUser;
}

const AUTH_KEY = 'vampireUserAuth';
const AUTH_CHANGED_EVENT = 'vampire-user-auth-changed';

function readAuth(): UserAuth | null {
  try {
    const raw = localStorage.getItem(AUTH_KEY);
    return raw ? (JSON.parse(raw) as UserAuth) : null;
  } catch {
    return null;
  }
}

function writeAuth(auth: UserAuth | null) {
  if (auth) {
    localStorage.setItem(AUTH_KEY, JSON.stringify(auth));
  } else {
    localStorage.removeItem(AUTH_KEY);
  }
  window.dispatchEvent(new Event(AUTH_CHANGED_EVENT));
}

// Also used directly by platformApi.ts / gmApi.ts (outside React) to attach
// the token to requests without importing the hook into a non-component module.
export function getUserAuth(): UserAuth | null {
  return readAuth();
}

export function clearUserAuth() {
  writeAuth(null);
}

export function useUserAuth() {
  const [auth, setAuth] = useState<UserAuth | null>(() => readAuth());

  useEffect(() => {
    const onChange = () => setAuth(readAuth());
    window.addEventListener(AUTH_CHANGED_EVENT, onChange);
    window.addEventListener('storage', onChange);
    return () => {
      window.removeEventListener(AUTH_CHANGED_EVENT, onChange);
      window.removeEventListener('storage', onChange);
    };
  }, []);

  const login = useCallback((next: UserAuth) => writeAuth(next), []);
  const logout = useCallback(() => writeAuth(null), []);

  return { auth, login, logout };
}
