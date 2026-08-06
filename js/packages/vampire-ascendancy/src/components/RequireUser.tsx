import { Navigate, useLocation } from 'react-router-dom';
import { useUserAuth } from '../userAuth';

// Redirects to /signin (preserving where to come back to) unless a real
// account is signed in. For platform pages ("My Toasts", "Host a Toast",
// accept-invite) — the GM console uses its own inline prompt instead, since
// it needs to stay on the same Toast's URL after signing in.
export const RequireUser = ({ children }: { children: React.ReactNode }) => {
  const { auth } = useUserAuth();
  const location = useLocation();
  if (!auth) {
    const next = encodeURIComponent(location.pathname + location.search);
    return <Navigate to={`/signin?next=${next}`} replace />;
  }
  return <>{children}</>;
};
