import { useEffect, useState } from 'react';
import { gmGetCharacter } from '../../gmApi';
import type { GMCharacterFull } from '../../gmApi';
import type { House, MeResponse } from '../../types';
import { Dossier } from '../Dossier';

// Build the MeResponse shape the player Dossier consumes from GM character data.
// `unlocked` toggles the post-Act / secrets gate so the GM can preview both states.
const buildMe = (c: GMCharacterFull, houses: House[], unlocked: boolean): MeResponse => {
  const house = houses.find((h) => h.id === c.houseId);
  return {
    player: { id: 'preview' },
    gameState: {
      currentAct: unlocked ? 'act2' : 'pre_event',
      contentUnlocked: unlocked,
      quizPart1Open: false,
      quizPart2Open: false,
      quizPart1OpenedAt: null,
      activeNotificationId: null,
    },
    character: {
      id: c.id,
      name: c.name,
      title: c.title,
      roleType: c.roleType,
      preEventInfo: c.preEventInfo,
      imageUrl: c.imageUrl || undefined,
      house: house ? { id: house.id, name: house.name, tagline: house.tagline } : undefined,
      postAct1Context: c.postAct1Context,
      secrets: c.secrets.map((s, i) => ({ id: `s${i}`, ordinal: s.ordinal || i + 1, body: s.body })),
      missions: [],
    },
    notification: null,
  };
};

// Full-screen modal that renders the actual player Dossier for one character, so
// a GM can confirm every field is populated and reads correctly.
export const DossierPreview = ({
  character,
  houses,
  onClose,
}: {
  character: GMCharacterFull;
  houses: House[];
  onClose: () => void;
}) => {
  const [unlocked, setUnlocked] = useState(false);
  const me = buildMe(character, houses, unlocked);

  // Flag empty fields so the GM doesn't have to eyeball it.
  const gaps: string[] = [];
  if (!character.title.trim()) gaps.push('title');
  if (!character.houseId) gaps.push('house');
  if (!character.preEventInfo.trim()) gaps.push('pre-event bio');
  if (!character.postAct1Context.trim()) gaps.push('post-act bio');
  if (character.secrets.length === 0) gaps.push('secrets');
  if (character.secrets.some((s) => !s.body.trim())) gaps.push('an empty secret');

  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-black/80" role="dialog" aria-label="Dossier preview">
      {/* Toolbar */}
      <div className="shrink-0 flex items-center gap-3 flex-wrap px-4 py-3 bg-blood-ink border-b border-blood/40">
        <span className="font-heading text-xs uppercase tracking-[0.3em] text-gold">Dossier preview</span>
        <label className="flex items-center gap-2 text-sm text-bone/80">
          <input type="checkbox" checked={unlocked} onChange={(e) => setUnlocked(e.target.checked)} />
          Show unlocked (Act One passed)
        </label>
        {gaps.length > 0 && (
          <span className="text-xs text-blood-bright">Missing: {gaps.join(', ')}</span>
        )}
        <button
          onClick={onClose}
          className="ml-auto text-xs text-bone/70 uppercase tracking-[0.2em] border border-blood/40 rounded-md px-3 py-1.5 hover:text-bone"
        >
          Close
        </button>
      </div>

      {/* The player view, in a phone-width column against the app background. */}
      <div className="flex-1 overflow-y-auto bg-blood-ink">
        <div className="max-w-md mx-auto px-4 py-6">
          <Dossier me={me} />
        </div>
      </div>
    </div>
  );
};

// Convenience wrapper for previewing a saved character straight from a roster row
// (fetches the full character, then shows the modal).
export const DossierPreviewLoader = ({
  characterId,
  houses,
  onClose,
}: {
  characterId: string;
  houses: House[];
  onClose: () => void;
}) => {
  const [c, setC] = useState<GMCharacterFull | null>(null);
  const [err, setErr] = useState(false);

  useEffect(() => {
    gmGetCharacter(characterId).then(setC).catch(() => setErr(true));
  }, [characterId]);

  if (err)
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80" onClick={onClose}>
        <p className="text-bone/80">Could not load character.</p>
      </div>
    );
  if (!c)
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80">
        <p className="text-bone/60">Loading preview…</p>
      </div>
    );
  return <DossierPreview character={c} houses={houses} onClose={onClose} />;
};
