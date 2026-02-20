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
    <div className="min-h-screen flex items-center justify-center px-4 py-10">
      <div className="login-shell w-full max-w-6xl">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1.1fr)_minmax(0,0.9fr)]">
          <section className="login-panel login-hero">
            <div className="login-glow" aria-hidden="true" />
            <div className="login-map" aria-hidden="true" />
            <div className="login-orb" aria-hidden="true" />
            <div className="login-spark login-spark--one" aria-hidden="true" />
            <div className="login-spark login-spark--two" aria-hidden="true" />

            <div className="relative z-10 flex flex-col gap-6">
              <div className="flex flex-wrap items-center gap-3 text-xs uppercase tracking-[0.35em] text-[#9ef9b0]">
                <span className="login-badge">Worldwalker Initiative</span>
                <span className="text-[#00cc00]">Expedition Cycle 07</span>
              </div>

              <header className="flex flex-col gap-4">
                <p className="text-sm uppercase tracking-[0.4em] text-[#00ff41]">BlauberTech Field Ops</p>
                <h1 className="text-3xl md:text-4xl lg:text-5xl font-semibold text-[#00ff00] leading-tight">
                  Step through the gate. The world is larger than you remember.
                </h1>
                <p className="text-base md:text-lg text-[#b9f8c6] max-w-xl">
                  Build your strength, map the unknown, and uncover both real and mystical locations. Each quest
                  unlocks new routes, deeper lore, and the next surge of power.
                </p>
              </header>

              <div className="grid gap-3 sm:grid-cols-2">
                <div className="login-quest-card">
                  <p className="text-xs uppercase tracking-[0.3em] text-[#f9d65c]">Quest Prime</p>
                  <p className="text-lg font-semibold text-[#e9fbe8]">Chart the wild corridors</p>
                  <p className="text-sm text-[#9ef9b0]">Reveal real-world landmarks and hidden paths.</p>
                </div>
                <div className="login-quest-card">
                  <p className="text-xs uppercase tracking-[0.3em] text-[#f9d65c]">Quest Prime</p>
                  <p className="text-lg font-semibold text-[#e9fbe8]">Awaken your strength</p>
                  <p className="text-sm text-[#9ef9b0]">Grow your abilities with every discovery.</p>
                </div>
                <div className="login-quest-card">
                  <p className="text-xs uppercase tracking-[0.3em] text-[#f9d65c]">Quest Prime</p>
                  <p className="text-lg font-semibold text-[#e9fbe8]">Trace the mythic sites</p>
                  <p className="text-sm text-[#9ef9b0]">Locate ruins, ley lines, and spectral waypoints.</p>
                </div>
                <div className="login-quest-card">
                  <p className="text-xs uppercase tracking-[0.3em] text-[#f9d65c]">Quest Prime</p>
                  <p className="text-lg font-semibold text-[#e9fbe8]">Complete the chain</p>
                  <p className="text-sm text-[#9ef9b0]">Finish quests to open the next realm.</p>
                </div>
              </div>

              <div className="login-steps">
                <div>
                  <p className="text-xs uppercase tracking-[0.3em] text-[#00ff41]">Your Path</p>
                  <p className="text-sm text-[#c9f9d3]">1. Receive your signal 2. Join your expedition 3. Launch the map</p>
                </div>
                <div className="login-signal">
                  <span className="login-signal__dot" />
                  <span className="text-xs uppercase tracking-[0.3em] text-[#f9d65c]">Signal Strong</span>
                </div>
              </div>
            </div>
          </section>

          <section className="login-panel login-auth">
            <div className="bg-black/90 backdrop-blur-sm p-6 md:p-8 rounded-lg border-2 border-[#00ff00] shadow-[0_0_20px_rgba(0,255,0,0.5)] matrix-card">
              <div className="flex flex-col gap-3">
                <p className="text-xs uppercase tracking-[0.3em] text-[#00cc00]">Access Gate</p>
                <h2 className="text-2xl font-bold text-[#00ff00]">BlauberTech BunkerKey</h2>
                <p className="text-sm text-[#9ef9b0]">
                  Enter your field number to receive the access pulse. Your expedition awaits.
                </p>
              </div>

              <div className="mt-6">
                <Logister
                  logister={handleLogister}
                  getVerificationCode={getVerificationCode}
                  isWaitingOnVerificationCode={isWaitingForVerificationCode}
                  isRegister={isRegister}
                  error={error ? String(error) : undefined}
                />
              </div>

              <div className="mt-6 text-xs text-[#9ef9b0]">
                By entering, you accept the expedition protocol and confirm you are ready for the next realm.
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
};
