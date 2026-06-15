import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getToken, submitMission } from '../api';
import type { MeResponse, Mission, Secret } from '../types';
import { TIER_LABEL, accentFor } from '../theme';
import { VampireMark } from './VampireMark';

export type DossierSection = 'dossier' | 'secrets' | 'missions';

// Presentational — the PlayerShell owns token capture, /me polling, the loading
// and error states, and the top navigation. Dossier renders the character header
// plus whichever section is selected.
export const Dossier = ({
  me,
  reload,
  section,
}: {
  me: MeResponse;
  reload: () => void;
  section: DossierSection;
}) => {
  const token = getToken() || '';
  const { character, gameState } = me;
  const unlocked = gameState.contentUnlocked;

  if (!character) {
    return (
      <CenteredNote>
        <VampireMark className="w-14 h-14 mx-auto mb-3 opacity-80" />
        <h1 className="font-display text-2xl font-bold text-bone mb-2">Awaiting your seat</h1>
        <p className="text-bone/80">
          You have entered the hall, but a role has not yet been bestowed upon you. The host will
          seat you shortly.
        </p>
      </CenteredNote>
    );
  }

  const accent = accentFor(character.house?.name);

  return (
    <div className="pb-8">
      <header className="text-center mb-6">
        <p className="text-xs uppercase tracking-[0.4em] text-gold">The Crimson Toast</p>
        <h1 className="mt-3 font-display text-3xl md:text-4xl font-bold text-bone leading-tight">
          {character.name}
        </h1>
        <p className="mt-2 text-bone/80 italic text-lg">{character.title}</p>
        {character.house && (
          <Link
            to={`/house/${character.house.id}`}
            className="inline-block mt-3 px-3 py-1 rounded-full text-xs uppercase tracking-[0.25em] border transition-colors hover:bg-white/5"
            style={{ color: accent, borderColor: accent }}
          >
            House of {character.house.name}
          </Link>
        )}
      </header>

      {section === 'dossier' && (
        <>
          <Section title="Pre-Event Briefing">
            <Prose text={character.preEventInfo} />
          </Section>

          {!unlocked && (
            <div className="mt-6 rounded-lg border border-blood/40 bg-black/40 p-6 text-center">
              <VampireMark className="w-12 h-12 mx-auto mb-3 opacity-80" />
              <p className="font-heading text-gold uppercase tracking-[0.3em] text-xs mb-2">
                Sealed
              </p>
              <p className="text-bone/85">
                The ceremony has not yet begun. Your full dossier will unlock when the host opens the
                evening.
              </p>
            </div>
          )}

          {unlocked && character.postAct1Context && (
            <Section title="As the Night Unfolds">
              <Prose text={character.postAct1Context} />
            </Section>
          )}
        </>
      )}

      {section === 'secrets' && <SecretsView secrets={character.secrets || []} />}
      {section === 'missions' && (
        <MissionsView missions={character.missions || []} token={token} onSubmitted={reload} />
      )}
    </div>
  );
};

const SecretsView = ({ secrets }: { secrets: Secret[] }) => {
  if (secrets.length === 0) return <CenteredNote>You carry no secrets tonight.</CenteredNote>;
  return (
    <div className="flex flex-col gap-4">
      {secrets.map((s) => (
        <div key={s.id} className="rounded-lg border border-blood/40 bg-black/40 p-5">
          <p className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-2">
            Secret {s.ordinal}
          </p>
          <Prose text={s.body} />
        </div>
      ))}
    </div>
  );
};

const MissionsView = ({
  missions,
  token,
  onSubmitted,
}: {
  missions: Mission[];
  token: string;
  onSubmitted: () => void;
}) => {
  if (missions.length === 0) return <CenteredNote>You have no missions tonight.</CenteredNote>;
  return (
    <div className="flex flex-col gap-4">
      {missions.map((m) => (
        <MissionCard key={m.id} mission={m} token={token} onSubmitted={onSubmitted} />
      ))}
    </div>
  );
};

const MissionCard = ({
  mission,
  token,
  onSubmitted,
}: {
  mission: Mission;
  token: string;
  onSubmitted: () => void;
}) => {
  const draftKey = `vampireDraft:${mission.id}`;
  const sub = mission.submission;
  const verified = sub?.status === 'verified';

  // Drafts persist locally so a flaky connection never loses a typed answer.
  const [answer, setAnswer] = useState(
    () => localStorage.getItem(draftKey) ?? sub?.playerAnswer ?? ''
  );
  const [submitting, setSubmitting] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    localStorage.setItem(draftKey, answer);
  }, [answer, draftKey]);

  const submit = async () => {
    if (!answer.trim() || submitting) return;
    setSubmitting(true);
    setErr(null);
    try {
      await submitMission(token, mission.id, answer.trim());
      onSubmitted();
    } catch {
      setErr('The court did not hear you — your answer is saved. Try again.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="rounded-lg border border-blood/40 bg-black/40 p-5">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs uppercase tracking-[0.25em] text-bone/60">
          {TIER_LABEL[mission.tier] || mission.tier}
        </span>
        <span className="text-blood-bright text-sm font-semibold">
          {mission.rewardBt} Blood Tokens
        </span>
      </div>
      <Prose text={mission.prompt} />
      {mission.answerFormat && (
        <p className="mt-3 text-xs text-bone/50 italic">{mission.answerFormat}</p>
      )}

      {sub && <StatusBadge status={sub.status} awardedBt={sub.awardedBt} />}

      <textarea
        value={answer}
        onChange={(e) => setAnswer(e.target.value)}
        disabled={verified}
        placeholder="Record your answer…"
        rows={3}
        className="mt-3 w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone placeholder:text-bone/30 focus:outline-none focus:border-blood-bright disabled:opacity-60"
      />
      {err && <p className="mt-2 text-xs text-blood-bright">{err}</p>}
      {!verified && (
        <button
          onClick={submit}
          disabled={submitting || !answer.trim()}
          className="mt-3 w-full py-2.5 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright transition-colors disabled:opacity-40"
        >
          {submitting ? 'Submitting…' : sub ? 'Resubmit' : 'Submit'}
        </button>
      )}
    </div>
  );
};

const StatusBadge = ({ status, awardedBt }: { status: string; awardedBt: number }) => {
  const styles: Record<string, { label: string; cls: string }> = {
    submitted: {
      label: 'Submitted — awaiting the court',
      cls: 'text-amber-300 border-amber-400/40',
    },
    verified: { label: `Verified · ${awardedBt} BT`, cls: 'text-green-300 border-green-400/40' },
    rejected: { label: 'Returned — try again', cls: 'text-blood-bright border-blood/50' },
  };
  const s = styles[status] || styles.submitted;
  return (
    <div
      className={`mt-3 inline-block px-3 py-1 rounded-full text-xs uppercase tracking-[0.18em] border ${s.cls}`}
    >
      {s.label}
    </div>
  );
};

const Section = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <section className="mt-6">
    <h2 className="font-heading text-gold text-sm uppercase tracking-[0.3em] mb-2">{title}</h2>
    {children}
  </section>
);

// Renders prose, splitting on blank lines into paragraphs.
const Prose = ({ text }: { text: string }) => (
  <div className="flex flex-col gap-3">
    {text
      .split(/\n\s*\n/)
      .map((p) => p.trim())
      .filter(Boolean)
      .map((p, i) => (
        <p key={i} className="text-bone leading-relaxed">
          {p}
        </p>
      ))}
  </div>
);

const CenteredNote = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-6 text-center">
    <div className="max-w-md text-bone/80">{children}</div>
  </div>
);
