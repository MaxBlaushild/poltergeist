import { Navigate, useNavigate } from 'react-router-dom';
import { useUserAuth } from '../userAuth';
import { VampireMark } from './VampireMark';

const STEPS = [
  { title: 'Host a Toast', body: 'Name your event and it starts fully stocked from the shared story — trim the roster to fit your guest count.' },
  { title: 'Invite your Court', body: "Bring in Co-Hosts to help run the night, and text every guest their character's bio and an RSVP link." },
  { title: 'Run the night live', body: "Advance the acts, verify missions, and broadcast announcements — everyone's screen updates together." },
];

// Public landing page — the app's front door, but only for signed-out
// visitors. A guest's RSVP link (/rsvp/:token) bypasses this entirely, and
// a signed-in visitor skips straight past it to their dashboard — there's
// nothing on this page they need once they have an account.
export const Landing = () => {
  const { auth } = useUserAuth();

  if (auth) return <Navigate to="/toasts" replace />;

  return (
    <div className="min-h-screen px-4 py-16">
      <div className="mx-auto max-w-2xl text-center">
        <VampireMark className="w-16 h-16 mx-auto mb-5" />
        <p className="text-xs uppercase tracking-[0.4em] text-gold mb-3">Vampire Ascendancy</p>
        <h1 className="font-display text-4xl sm:text-5xl font-bold text-bone mb-4">The Crimson Toast</h1>
        <p className="text-bone/70 text-lg max-w-lg mx-auto mb-8">
          A live murder-mystery party in a box: houses, hidden secrets, missions, and a night that plays
          out in real time on everyone's phone. Host your own Court.
        </p>
        <CtaButton />
      </div>

      <div className="mx-auto max-w-3xl mt-16 grid gap-6 sm:grid-cols-3">
        {STEPS.map((s, i) => (
          <div key={s.title} className="rounded-lg border border-blood/30 bg-black/40 p-5">
            <p className="text-gold text-xs uppercase tracking-[0.3em] mb-2">Step {i + 1}</p>
            <h2 className="font-heading text-bone font-semibold mb-2">{s.title}</h2>
            <p className="text-bone/60 text-sm">{s.body}</p>
          </div>
        ))}
      </div>

      <div className="text-center mt-16">
        <p className="text-bone/40 text-xs uppercase tracking-[0.2em]">
          Got a text invite? Open the link to see your character and RSVP.
        </p>
      </div>
    </div>
  );
};

// Only ever rendered signed-out (Landing bails to /toasts otherwise), so
// this always starts the sign-in flow. The dashboard itself handles "no
// Toasts yet" with its own "Host another Toast" CTA, so sending everyone
// through /toasts on the way back works whether they're new or already
// have Toasts waiting (as a Host, Co-Host, or player).
const CtaButton = () => {
  const navigate = useNavigate();
  return (
    <button
      onClick={() => navigate('/signin?next=/toasts')}
      className="px-8 py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright"
    >
      Host a Toast
    </button>
  );
};
