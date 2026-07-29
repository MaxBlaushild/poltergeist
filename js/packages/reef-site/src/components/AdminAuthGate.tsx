import { useState, type ReactNode } from 'react';
import { ApiError } from '../api/client';
import { getAdminAuthHeader, setAdminPassword } from '../api/adminAuth';

// Wraps /operator pages. There's no session/user system in reef-site, just
// a shared password gating the whole surface (see server.go's gin.BasicAuth
// group) — this collects it once per tab and lets the wrapped page's own
// data fetch double as the credential check (a wrong password just surfaces
// as that fetch's 401).
export default function AdminAuthGate({ children }: { children: (onAuthError: () => void) => ReactNode }) {
  const [unlocked, setUnlocked] = useState(() => getAdminAuthHeader() !== null);
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setAdminPassword(password);
    setError(null);
    setUnlocked(true);
  };

  const handleAuthError = () => {
    setUnlocked(false);
    setError('Incorrect password.');
  };

  if (!unlocked) {
    return (
      <form onSubmit={handleSubmit} className="mx-auto max-w-xs space-y-3">
        <h1 className="font-display text-xl font-bold text-reef-ink">Operator login</h1>
        <input
          type="password"
          autoFocus
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Password"
          className="w-full rounded-lg border border-reef-teal/30 px-3 py-2"
        />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button type="submit" className="btn-primary w-full py-2">
          Log in
        </button>
      </form>
    );
  }

  return <>{children(handleAuthError)}</>;
}

export function isUnauthorized(err: unknown): boolean {
  return err instanceof ApiError && err.status === 401;
}
