import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { reefApi, ApiError } from '../api/client';
import { useCustomerAuth } from '../hooks/useCustomerAuth';
import GoogleSignInButton from '../components/GoogleSignInButton';

export default function Signup() {
  const navigate = useNavigate();
  const { login } = useCustomerAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const auth = await reefApi.register(name, email, password);
      login(auth);
      navigate('/account');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Sign up failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-sm space-y-6">
      <h1 className="font-display text-2xl font-bold text-reef-ink">Sign up</h1>

      <GoogleSignInButton onError={setError} />

      <div className="flex items-center gap-3 text-xs text-reef-ink/40">
        <div className="h-px flex-1 bg-reef-ink/10" />
        or
        <div className="h-px flex-1 bg-reef-ink/10" />
      </div>

      <form onSubmit={handleSubmit} className="card space-y-4 p-6">
        <div>
          <label className="mb-1 block text-sm font-medium text-reef-ink">Name</label>
          <input
            type="text"
            required
            className="input-field"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-reef-ink">Email</label>
          <input
            type="email"
            required
            className="input-field"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-reef-ink">Password</label>
          <input
            type="password"
            required
            minLength={8}
            className="input-field"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button type="submit" disabled={loading} className="btn-primary w-full">
          {loading ? 'Creating account…' : 'Sign up'}
        </button>
      </form>
      <p className="text-center text-sm text-reef-ink/60">
        Already have an account?{' '}
        <Link to="/login" className="font-medium text-reef-teal underline underline-offset-2">
          Log in
        </Link>
      </p>
    </div>
  );
}
