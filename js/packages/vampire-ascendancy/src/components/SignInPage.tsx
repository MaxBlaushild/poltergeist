import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useUserAuth } from '../userAuth';
import { SignInForm } from './SignInForm';

// Platform sign-in, reached from the landing page's "Host a Toast" CTA (or
// directly). Redirects to ?next= (default /toasts) once signed in — and
// immediately if already signed in.
export const SignInPage = () => {
  const { auth } = useUserAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const next = params.get('next') || '/toasts';

  useEffect(() => {
    if (auth) navigate(next, { replace: true });
  }, [auth, next, navigate]);

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <SignInForm onSignedIn={() => navigate(next, { replace: true })} />
    </div>
  );
};
