import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { MeResponse, Secret } from '../types';
import { accentFor, houseLabel } from '../theme';
import { VampireMark } from './VampireMark';

type Segment = 'bio' | 'postAct' | 'secrets';

// "Who you are." A fixed character header + segmented control; only the panel
// below changes. Post-Act and Secrets are gated until the host opens the evening
// (act one complete); until then they show a sealed panel.
export const Dossier = ({ me }: { me: MeResponse }) => {
  const { character, gameState } = me;
  const unlocked = gameState.contentUnlocked;
  const [seg, setSeg] = useState<Segment>('bio');

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
      <header className="text-center mb-5">
        <Portrait imageUrl={character.imageUrl} name={character.name} accent={accent} />
        <h1 className="mt-4 font-display text-3xl md:text-4xl font-bold text-bone leading-tight">
          {character.name}
        </h1>
        <p className="mt-2 text-bone/80 italic text-lg">{character.title}</p>
        {character.house && (
          <>
            <Link
              to={`/house/${character.house.id}`}
              className="inline-block mt-3 px-3 py-1 rounded-full text-xs uppercase tracking-[0.25em] border transition-colors hover:bg-white/5"
              style={{ color: accent, borderColor: accent }}
            >
              {houseLabel(character.house.name)}
            </Link>
            {character.house.tagline && (
              <p className="mt-2 text-xs uppercase tracking-[0.3em] text-bone/50 italic">
                {character.house.tagline}
              </p>
            )}
          </>
        )}
      </header>

      {/* Segmented control — fixed across all three panels. */}
      <div className="flex gap-1 p-1 rounded-lg bg-black/50 border border-blood/30 mb-5">
        <SegButton active={seg === 'bio'} onClick={() => setSeg('bio')}>
          Bio
        </SegButton>
        <SegButton active={seg === 'postAct'} onClick={() => setSeg('postAct')} locked={!unlocked}>
          The Night
        </SegButton>
        <SegButton active={seg === 'secrets'} onClick={() => setSeg('secrets')} locked={!unlocked}>
          Secrets
        </SegButton>
      </div>

      {seg === 'bio' && <Prose text={character.preEventInfo} />}

      {seg === 'postAct' &&
        (unlocked && character.postAct1Context ? (
          <Prose text={character.postAct1Context} />
        ) : (
          <SealedPanel body="As the night unfolds, your story will deepen. This chapter opens once the first act has passed." />
        ))}

      {seg === 'secrets' &&
        (unlocked ? (
          <SecretsView secrets={character.secrets || []} />
        ) : (
          <SealedPanel body="Your secrets are yours alone — sealed until the first act has passed." />
        ))}
    </div>
  );
};

const SegButton = ({
  active,
  locked,
  onClick,
  children,
}: {
  active: boolean;
  locked?: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) => (
  <button
    onClick={onClick}
    className={`flex-1 flex items-center justify-center gap-1.5 py-2 rounded-md uppercase tracking-[0.12em] text-xs sm:text-sm transition-colors ${
      active ? 'bg-blood text-bone' : 'text-bone/70 hover:text-bone'
    }`}
  >
    {children}
    {locked && <LockIcon className="w-3 h-3 opacity-70" />}
  </button>
);

const SealedPanel = ({ body }: { body: string }) => (
  <div className="rounded-lg border border-blood/40 bg-black/40 p-8 text-center">
    <LockIcon className="w-8 h-8 mx-auto mb-3 text-gold/80" />
    <p className="font-heading text-gold uppercase tracking-[0.3em] text-xs mb-2">Sealed for now</p>
    <p className="text-bone/85 leading-relaxed mb-4">{body}</p>
    <span className="inline-block px-3 py-1 rounded-full text-[11px] uppercase tracking-[0.2em] border border-gold/40 text-gold/90">
      Unlocks after Act One
    </span>
  </div>
);

const SecretsView = ({ secrets }: { secrets: Secret[] }) => {
  if (secrets.length === 0) return <CenteredNote>You carry no secrets tonight.</CenteredNote>;
  return (
    <div>
      <p className="text-center text-bone/60 text-sm italic mb-4">
        These are yours alone — reveal them, trade them, or guard them as you see fit.
      </p>
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
    </div>
  );
};

// Portraits are hosted in S3 (keeps the repo lean). The object key is the slug
// below + ".png" — e.g. "Ambrose Cipher" → ambrose-cipher.png. Must stay in sync
// with how those objects are named in the bucket.
const PORTRAIT_BASE = 'https://vampire-portraits.s3.us-east-1.amazonaws.com';
const portraitSlug = (name: string) =>
  name
    .toLowerCase()
    .replace(/['"’“”]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');

// Character portrait. Uses an explicit imageUrl if set, else the conventional
// S3 object <slug>.png, and falls back to a house-tinted crest if the image is
// missing — so the header never looks broken.
const Portrait = ({
  imageUrl,
  name,
  accent,
}: {
  imageUrl?: string;
  name: string;
  accent: string;
}) => {
  const [failed, setFailed] = useState(false);
  const [open, setOpen] = useState(false);
  const src = imageUrl || `${PORTRAIT_BASE}/${portraitSlug(name)}.png`;
  const hasImage = !!src && !failed;

  return (
    <>
      {hasImage ? (
        <button
          type="button"
          onClick={() => setOpen(true)}
          aria-label={`View ${name}'s portrait`}
          className="mt-4 mx-auto block w-56 max-w-full rounded-2xl overflow-hidden border-2 cursor-pointer hover:opacity-90 transition-opacity"
          style={{ borderColor: accent }}
        >
          {/* h-auto keeps the image's own aspect ratio — scales, never crops. */}
          <img src={src} alt={name} onError={() => setFailed(true)} className="block w-full h-auto" />
        </button>
      ) : (
        <div
          className="mt-4 mx-auto w-40 h-40 rounded-2xl overflow-hidden border-2 flex items-center justify-center bg-black/40"
          style={{ borderColor: accent }}
        >
          <VampireMark className="w-16 h-16 opacity-70" />
        </div>
      )}

      {open && hasImage && (
        <div
          className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/90 p-6"
          onClick={() => setOpen(false)}
          role="dialog"
          aria-label={`${name} portrait`}
        >
          <img
            src={src}
            alt={name}
            className="max-h-[82vh] max-w-[92vw] rounded-lg border-2 object-contain"
            style={{ borderColor: accent }}
          />
          <p className="mt-4 font-display text-2xl text-bone">{name}</p>
        </div>
      )}
    </>
  );
};

const LockIcon = ({ className = '' }: { className?: string }) => (
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className={className}>
    <rect x="4" y="11" width="16" height="10" rx="2" />
    <path d="M8 11V7a4 4 0 0 1 8 0v4" />
  </svg>
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
  <div className="min-h-[50vh] flex items-center justify-center px-6 text-center">
    <div className="max-w-md text-bone/80">{children}</div>
  </div>
);
