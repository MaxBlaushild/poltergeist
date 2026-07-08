import { useEffect, useRef, useState } from 'react';
import { getToken, submitMission, photoUrl } from '../api';
import type { MeResponse, Mission } from '../types';
import { TIER_LABEL } from '../theme';
import { fileToResizedDataURL } from '../photo';
import { VampireMark } from './VampireMark';

// A mission only offers photo upload when its prompt or answer format explicitly
// asks for a picture — most missions are text-only.
const wantsPhoto = (m: Mission) =>
  /\b(photo|picture|pic|pics|image|selfie|snapshot|snap)\b/i.test(`${m.prompt} ${m.answerFormat}`);

// "What you need to accomplish." Own tab now — the character's private missions,
// gated until the host opens the evening.
export const Missions = ({ me, reload }: { me: MeResponse; reload: () => void }) => {
  const token = getToken() || '';
  const { character, gameState } = me;

  if (!gameState.contentUnlocked) {
    return (
      <div className="pb-8">
        <Header />
        <div className="mt-2 rounded-lg border border-blood/40 bg-black/40 p-8 text-center">
          <VampireMark className="w-12 h-12 mx-auto mb-3 opacity-80" />
          <p className="font-heading text-gold uppercase tracking-[0.3em] text-xs mb-2">Sealed</p>
          <p className="text-bone/85">Your missions are revealed once the host opens the evening.</p>
        </div>
      </div>
    );
  }

  const missions = character?.missions || [];
  return (
    <div className="pb-8">
      <Header />
      {missions.length === 0 ? (
        <CenteredNote>You have no missions tonight.</CenteredNote>
      ) : (
        <div className="flex flex-col gap-4">
          {missions.map((m) => (
            <MissionCard key={m.id} mission={m} token={token} onSubmitted={reload} />
          ))}
        </div>
      )}
    </div>
  );
};

const Header = () => (
  <header className="text-center mb-6">
    <p className="text-xs uppercase tracking-[0.4em] text-gold">Your Missions</p>
    <h1 className="mt-3 font-display text-3xl font-bold text-bone">What You Must Accomplish</h1>
  </header>
);

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
  // Approved or redeemed both lock the mission — the answer is accepted and (being) paid.
  const done = sub?.status === 'approved' || sub?.status === 'redeemed';
  const photoMission = wantsPhoto(mission);

  const [answer, setAnswer] = useState(
    () => localStorage.getItem(draftKey) ?? sub?.playerAnswer ?? ''
  );
  const [newPhotos, setNewPhotos] = useState<string[]>([]);
  const [cleared, setCleared] = useState(false);
  const [adding, setAdding] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    localStorage.setItem(draftKey, answer);
  }, [answer, draftKey]);

  const existingPhotos = cleared ? [] : sub?.photoIds ?? [];
  const canSubmit = answer.trim().length > 0 || newPhotos.length > 0 || cleared;

  const addPhotos = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    e.target.value = '';
    if (files.length === 0) return;
    setAdding(true);
    const urls: string[] = [];
    for (const f of files) {
      try {
        urls.push(await fileToResizedDataURL(f));
      } catch {
        /* skip unreadable files */
      }
    }
    setNewPhotos((p) => [...p, ...urls].slice(0, 6));
    setAdding(false);
  };

  const submit = async () => {
    if (!canSubmit || submitting) return;
    setSubmitting(true);
    setErr(null);
    try {
      await submitMission(token, mission.id, answer.trim(), {
        photos: newPhotos.length ? newPhotos : undefined,
        clearPhotos: cleared || undefined,
      });
      setNewPhotos([]);
      setCleared(false);
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
        // The answer prompt: same white, same size, its own line with a paragraph
        // gap — but set in Lora italic, whose italics read more clearly than the
        // body face's.
        <p className="mt-4 text-bone leading-relaxed italic font-note">{mission.answerFormat}</p>
      )}

      {sub && <StatusBadge status={sub.status} awardedBt={sub.awardedBt} />}

      <textarea
        value={answer}
        onChange={(e) => setAnswer(e.target.value)}
        disabled={done}
        placeholder="Record your answer…"
        rows={3}
        className="mt-3 w-full rounded-md bg-black/60 border border-blood/40 p-3 text-bone placeholder:text-bone/30 focus:outline-none focus:border-blood-bright disabled:opacity-60"
      />

      {photoMission && (existingPhotos.length > 0 || newPhotos.length > 0) && (
        <div className="mt-3 flex flex-wrap gap-2">
          {existingPhotos.map((id) => (
            <img
              key={id}
              src={photoUrl(id)}
              alt=""
              className="w-16 h-16 object-cover rounded-md border border-blood/40"
            />
          ))}
          {newPhotos.map((d, i) => (
            <div key={i} className="relative">
              <img
                src={d}
                alt=""
                className="w-16 h-16 object-cover rounded-md border border-blood-bright/60"
              />
              {!done && (
                <button
                  onClick={() => setNewPhotos((p) => p.filter((_, j) => j !== i))}
                  className="absolute -top-1.5 -right-1.5 w-5 h-5 rounded-full bg-blood text-bone text-xs leading-none"
                >
                  ×
                </button>
              )}
            </div>
          ))}
        </div>
      )}

      {photoMission && !done && (
        <div className="mt-3 flex items-center gap-4">
          <button
            onClick={() => fileRef.current?.click()}
            className="text-xs text-gold uppercase tracking-[0.15em]"
          >
            {adding ? 'Adding…' : '+ Add photo'}
          </button>
          {existingPhotos.length > 0 && (
            <button
              onClick={() => setCleared(true)}
              className="text-xs text-bone/40 uppercase tracking-[0.15em]"
            >
              Remove photos
            </button>
          )}
          <input
            ref={fileRef}
            type="file"
            accept="image/*"
            multiple
            onChange={addPhotos}
            className="hidden"
          />
        </div>
      )}

      {err && <p className="mt-2 text-xs text-blood-bright">{err}</p>}
      {!done && (
        <button
          onClick={submit}
          disabled={submitting || !canSubmit}
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
    approved: {
      label: `Approved · ${awardedBt} BT — collect from Ivara at the Blood Bank`,
      cls: 'text-sky-300 border-sky-400/40',
    },
    redeemed: { label: `Redeemed · ${awardedBt} BT`, cls: 'text-green-300 border-green-400/40' },
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
  <div className="min-h-[40vh] flex items-center justify-center px-6 text-center">
    <div className="max-w-md text-bone/80">{children}</div>
  </div>
);
