import { useCallback, useEffect, useState } from 'react';
import type { CustomerAuth } from '../api/types';

const AUTH_KEY = 'reef_customer_auth';
const AUTH_CHANGED_EVENT = 'reef-auth-changed';

function readAuth(): CustomerAuth | null {
  try {
    const raw = localStorage.getItem(AUTH_KEY);
    return raw ? (JSON.parse(raw) as CustomerAuth) : null;
  } catch {
    return null;
  }
}

function writeAuth(auth: CustomerAuth | null) {
  if (auth) {
    localStorage.setItem(AUTH_KEY, JSON.stringify(auth));
  } else {
    localStorage.removeItem(AUTH_KEY);
  }
  window.dispatchEvent(new Event(AUTH_CHANGED_EVENT));
}

// Also used directly by api/client.ts (outside React) to attach the token
// to requests without importing the hook into a non-component module.
export function getStoredAuth(): CustomerAuth | null {
  return readAuth();
}

export function clearStoredAuth() {
  writeAuth(null);
}

export function useCustomerAuth() {
  const [auth, setAuth] = useState<CustomerAuth | null>(() => readAuth());

  useEffect(() => {
    const onChange = () => setAuth(readAuth());
    window.addEventListener(AUTH_CHANGED_EVENT, onChange);
    window.addEventListener('storage', onChange);
    return () => {
      window.removeEventListener(AUTH_CHANGED_EVENT, onChange);
      window.removeEventListener('storage', onChange);
    };
  }, []);

  const login = useCallback((next: CustomerAuth) => writeAuth(next), []);
  const logout = useCallback(() => writeAuth(null), []);

  return { auth, login, logout };
}
