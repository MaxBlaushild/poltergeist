import { useEffect, useState } from 'react';
import { listSelectableCharacters, chooseCharacter, ApiError } from '../api';
import type { SelectableCharacter } from '../api';
import { CharacterBrowser } from './gm/CharacterBrowser';
import { VampireMark } from './VampireMark';

// A player with no character yet — post-accept, pre-play — browses the
// Host's curated pool (tag filters and all, same UI the Host uses to
// curate it) and claims one. Takes over the whole player app until they've
// chosen; see PlayerShell.
export const CharacterPicker = ({ onChosen }: { onChosen: () => void }) => {
  const [characters, setCharacters] = useState<SelectableCharacter[] | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = () => {
    listSelectableCharacters()
      .then((d) => setCharacters(d.characters))
      .catch(() => setError('Could not load characters.'));
  };
  useEffect(() => {
    load();
  }, []);

  const confirm = async () => {
    if (!selectedId || busy) return;
    setBusy(true);
    setError(null);
    try {
      await chooseCharacter(selectedId);
      onChosen();
    } catch (e) {
      // Most likely someone else just claimed it — refresh the list so it
      // drops off, rather than leaving a stale, now-unpickable row selected.
      setError(e instanceof ApiError ? e.message : 'Could not choose that character.');
      setSelectedId(null);
      load();
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="min-h-screen max-w-2xl mx-auto px-4 py-10">
      <header className="text-center mb-6">
        <VampireMark className="w-14 h-14 mx-auto mb-3" />
        <h1 className="font-display text-2xl font-bold text-bone mb-2">Choose your character</h1>
        <p className="text-bone/80">Browse the Court and pick who you'll play tonight — this is permanent once confirmed.</p>
      </header>

      {characters === null && <p className="text-bone/50 text-center">Gathering the Court…</p>}
      {characters !== null && characters.length === 0 && (
        <p className="text-gold/80 text-center">
          No characters are available to choose yet — ask your host to open the character pool.
        </p>
      )}
      {characters !== null && characters.length > 0 && (
        <>
          <CharacterBrowser
            characters={characters}
            selectedId={selectedId ?? undefined}
            onSelect={(c) => setSelectedId(c.id)}
          />
          {error && <p className="text-blood-bright text-sm mt-3 text-center">{error}</p>}
          <button
            onClick={confirm}
            disabled={!selectedId || busy}
            className="mt-4 w-full py-3 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm hover:bg-blood-bright disabled:opacity-40"
          >
            {busy ? 'Entering the Court…' : 'Confirm this character'}
          </button>
        </>
      )}
    </div>
  );
};
