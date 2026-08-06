import { useState } from 'react';
import { loginUser, registerUser } from '../platformApi';
import { ApiError } from '../api';
import { useUserAuth } from '../userAuth';
import { GoogleSignInButton } from './GoogleSignInButton';

// Shared email/password + Google sign-in-or-up form. Used both on the
// platform /signin page and inline wherever the GM console needs a signed-in
// account (e.g. landing on /e/:instanceId/gm signed out).
export const SignInForm = ({
  onSignedIn,
  title = 'Sign in',
}: {
  onSignedIn: () => void;
  title?: string;
}) => {
  const { login } = useUserAuth();
  const [mode, setMode] = useState<'signin' | 'signup'>('signin');
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    if (busy) return;
    setBusy(true);
    setError(null);
    try {
      const auth =
        mode === 'signin' ? await loginUser(email, password) : await registerUser(name, email, password);
      login(auth);
      onSignedIn();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Something went wrong. Try again.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="w-full max-w-sm rounded-lg border border-blood/50 bg-black/70 p-6">
      <h1 className="font-display text-2xl font-bold text-bone text-center mb-6">{title}</h1>

      <GoogleSignInButton onSignedIn={onSignedIn} onError={setError} />

      <div className="flex items-center gap-3 text-xs text-bone/40 my-5">
        <div className="h-px flex-1 bg-bone/10" />
        or
        <div className="h-px flex-1 bg-bone/10" />
      </div>

      <div className="space-y-3">
        {mode === 'signup' && (
          <div>
            <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">Name</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
            />
          </div>
        )}
        <div>
          <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && submit()}
            className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
          />
        </div>
        <div>
          <label className="block text-xs uppercase tracking-[0.2em] text-bone/60 mb-1">Password</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && submit()}
            className="w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
          />
        </div>
        {error && <p className="text-blood-bright text-sm">{error}</p>}
        <button
          onClick={submit}
          disabled={busy || !email || !password || (mode === 'signup' && !name)}
          className="w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
        >
          {busy ? 'Please wait…' : mode === 'signin' ? 'Sign in' : 'Create account'}
        </button>
      </div>

      <button
        onClick={() => {
          setMode((m) => (m === 'signin' ? 'signup' : 'signin'));
          setError(null);
        }}
        className="mt-4 w-full text-center text-bone/50 hover:text-bone text-xs uppercase tracking-[0.2em]"
      >
        {mode === 'signin' ? "Don't have an account? Sign up" : 'Already have an account? Sign in'}
      </button>
    </div>
  );
};
