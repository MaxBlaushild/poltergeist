import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { acceptCoHostInvite } from '../platformApi';
import { ApiError } from '../api';
import { VampireMark } from './VampireMark';

// /accept-invite/:token — "Invite a Co-Host" lands here. Wrapped in
// <RequireUser> by App.tsx, so by the time this renders the visitor is
// signed in (sign-up first if they didn't have an account).
export const AcceptInvite = () => {
  const { token } = useParams();
  const navigate = useNavigate();
  const [status, setStatus] = useState<'pending' | 'error'>('pending');
  const [error, setError] = useState('');

  useEffect(() => {
    if (!token) {
      setStatus('error');
      setError('Missing invite.');
      return;
    }
    acceptCoHostInvite(token)
      .then((res) => navigate(`/e/${res.instanceId}/gm`, { replace: true }))
      .catch((err) => {
        setStatus('error');
        setError(err instanceof ApiError ? err.message : 'This invite could not be accepted.');
      });
  }, [token, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center px-4 text-center">
      <div className="max-w-sm">
        <VampireMark className="w-12 h-12 mx-auto mb-3" />
        {status === 'pending' ? (
          <p className="text-bone/70">Joining the Court…</p>
        ) : (
          <p className="text-blood-bright">{error}</p>
        )}
      </div>
    </div>
  );
};
