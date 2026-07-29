import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { reefApi, ApiError } from '../api/client';
import { useCustomerAuth } from '../hooks/useCustomerAuth';

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize(config: { client_id: string; callback: (resp: { credential: string }) => void }): void;
          renderButton(parent: HTMLElement, options: { theme?: string; size?: string; width?: number }): void;
        };
      };
    };
  }
}

// Google Identity Services' ID-token flow: the button below mints a
// credential entirely client-side (no redirect), which is POSTed to
// /api/reef/auth/google for server-side verification — the same
// {user, token} shape every other login path in this app returns.
export default function GoogleSignInButton({ onError }: { onError: (message: string) => void }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { login } = useCustomerAuth();
  const navigate = useNavigate();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID;
    if (!clientId || !containerRef.current) return;

    const tryInit = () => {
      if (!window.google || !containerRef.current) return false;
      window.google.accounts.id.initialize({
        client_id: clientId,
        callback: async (resp) => {
          try {
            const auth = await reefApi.loginWithGoogle(resp.credential);
            login(auth);
            navigate('/account');
          } catch (err) {
            onError(err instanceof ApiError ? err.message : 'Google sign-in failed');
          }
        },
      });
      window.google.accounts.id.renderButton(containerRef.current, { theme: 'outline', size: 'large', width: 320 });
      setReady(true);
      return true;
    };

    if (tryInit()) return;
    // The GSI script (loaded in index.html) may not have executed yet.
    const interval = setInterval(() => {
      if (tryInit()) clearInterval(interval);
    }, 200);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="flex justify-center">
      <div ref={containerRef} />
      {!ready && <p className="text-xs text-reef-ink/40">Loading Google sign-in…</p>}
    </div>
  );
}
