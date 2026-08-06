import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useUserAuth } from '../userAuth';
import { listMyToasts } from '../platformApi';
import { VampireMark } from './VampireMark';

const STEPS = [
  { title: 'Host a Toast', body: 'Name your event and it starts fully stocked from the shared story — trim the roster to fit your guest count.' },
  { title: 'Invite your Court', body: "Bring in Co-Hosts to help run the night, and text every guest their character's bio and an RSVP link." },
  { title: 'Run the night live', body: "Advance the acts, verify missions, and broadcast announcements — everyone's screen updates together." },
];

// Public landing page — the app's front door. A guest's RSVP link
// (/rsvp/:token) bypasses this entirely.
export const Landing = () => {
  const { auth } = useUserAuth();
  const navigate = useNavigate();
  const [toastCount, setToastCount] = useState<number | null>(null);

  useEffect(() => {
    if (!auth) {
      setToastCount(null);
      return;
    }
    listMyToasts()
      .then((d) => setToastCount(d.instances.length))
      .catch(() => setToastCount(null));
  }, [auth]);

  const primaryCta = () => {
    if (!auth) {
      navigate('/signin?next=/toasts/new');
      return;
    }
    if (toastCount && toastCount > 0) {
      navigate('/toasts');
    } else {
      navigate('/toasts/new');
    }
  };

  const ctaLabel = auth && toastCount && toastCount > 0 ? 'My Toasts' : 'Host a Toast';

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
        <button
          onClick={primaryCta}
          className="px-8 py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright"
        >
          {ctaLabel}
        </button>
        {auth && (
          <p className="mt-3 text-xs text-bone/40">
            Signed in as {auth.user.name}
          </p>
        )}
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
          Have a character link? Just open it — no account needed.
        </p>
      </div>
    </div>
  );
};
