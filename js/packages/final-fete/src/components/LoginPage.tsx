import { useAuth } from '@poltergeist/contexts';
import { Logister } from '@poltergeist/components';
import { useSearchParams } from 'react-router-dom';
import axios from 'axios';

export const LoginPage = () => {
  const { logister: authLogister, getVerificationCode, isWaitingForVerificationCode, isRegister, error } = useAuth();
  const [searchParams] = useSearchParams();
  const teamId = searchParams.get('teamId');

  // Wrapper to match Logister component's type signature
  // The Logister component type says it expects (phoneNumber, verificationCode, name, isRegister)
  // but actually calls it as (phoneNumber, code, isRegister) - there's a type mismatch in the component
  // We'll accept all 4 params to match the type, but ignore name since it's not used for final-fete
  const handleLogister = async (phoneNumber: string, verificationCode: string, _name: string, _isRegister: boolean) => {
    // If teamId is in URL, we need to make the API call with teamId as query parameter
    // Since the auth context doesn't support query params, we'll make the call directly
    if (teamId) {
      try {
        // Try login first
        const loginUrl = `${import.meta.env.VITE_API_URL}/final-fete/login?teamId=${encodeURIComponent(teamId)}`;
        try {
          const loginResponse = await axios.post(loginUrl, {
            phoneNumber,
            code: verificationCode
          });
          const { token } = loginResponse.data;
          localStorage.setItem('token', token);
          // Hard reload to trigger auth state update
          window.location.reload();
          return;
        } catch (loginError: any) {
          // If login fails, try register
          if (loginError.response?.status === 400 || loginError.response?.status === 401) {
            const registerUrl = `${import.meta.env.VITE_API_URL}/final-fete/register?teamId=${encodeURIComponent(teamId)}`;
            const registerResponse = await axios.post(registerUrl, {
              phoneNumber,
              code: verificationCode
            });
            const { token } = registerResponse.data;
            localStorage.setItem('token', token);
            // Hard reload to trigger auth state update
            window.location.reload();
            return;
          }
          throw loginError;
        }
      } catch (err) {
        // Fall back to normal auth flow if teamId handling fails
        await authLogister(phoneNumber, verificationCode, '');
        // Hard reload after auth
        window.location.reload();
      }
    } else {
      // No teamId, use normal auth flow
      await authLogister(phoneNumber, verificationCode, '');
      // Hard reload after auth
      window.location.reload();
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="bg-black/90 backdrop-blur-sm p-4 md:p-6 lg:p-8 rounded-lg border-2 border-[#00ff00] shadow-[0_0_20px_rgba(0,255,0,0.5)] w-full max-w-md mx-4 matrix-card">
        <h1 className="text-xl md:text-2xl font-bold mb-6 text-center text-[#00ff00]">BlauberTech BunkerKey</h1>
        <Logister
          logister={handleLogister}
          getVerificationCode={getVerificationCode}
          isWaitingOnVerificationCode={isWaitingForVerificationCode}
          isRegister={isRegister}
          error={error ? String(error) : undefined}
        />
      </div>
    </div>
  );
};

